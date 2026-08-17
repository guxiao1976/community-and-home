package permission

// privileged_roles_min_verf_level_test.go — 服务角色敏感写/管理权限 min_verf_level=2 判定验证
//
// 背景（security-arch 评审 CRITICAL）：服务角色（网格员 grid_worker / 社区管理员 community_admin /
// 物业管理员 property_admin）可免 membership 自助申请（user-service 并行改造），但其未认证(status=0)
// grant 不得行使破坏性操作。min_verf_level=2（需已认证）为数据驱动约束：CheckPermission 放行
// ⟺ maxLevel(满足层级) >= minLevel；未认证 grant 恒 level-0（grantSatisfiedLevel）→ 无法放行。
//
// 本测试验证对 6 个已加固权限码：status=0 未认证 grant 一律拒绝，status=2+verified_at 已认证 grant 放行。
//
// SEE: [[auto-grant-unverified-grant-confers-scope-level0]] — 未认证 grant 立即生效的既有语义，
// 现经 min_verf_level=2 数据驱动收窄
// SEE: [[is-system-no-permission-shortcut]] — 权限经 rel_role_permission 配置，认证要求经 min_verf_level 数据驱动

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckPermission_PrivilegedRoleDestructiveOps_NeedVerf(t *testing.T) {
	verifiedAt := sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true}

	// 6 个已加固权限：path 含 METHOD 前缀（API 型）或 code 无 path（按钮型 role:read）
	perms := []*model.SysPermission{
		{Id: 427, Code: "community:notice:delete-api", Type: 3,
			Path: sql.NullString{String: "DELETE:/api/community/notices/:id", Valid: true}, MinVerfLevel: 2},
		{Id: 428, Code: "community:notice:update-api", Type: 3,
			Path: sql.NullString{String: "PUT:/api/community/notices/:id", Valid: true}, MinVerfLevel: 2},
		{Id: 432, Code: "community:activity:create-api", Type: 3,
			Path: sql.NullString{String: "POST:/api/community/activities", Valid: true}, MinVerfLevel: 2},
		{Id: 210, Code: "role:read", Type: 2,
			Path: sql.NullString{}, MinVerfLevel: 2},
		{Id: 211, Code: "role:read:list-api", Type: 3,
			Path: sql.NullString{String: "GET:/api/perm/roles", Valid: true}, MinVerfLevel: 2},
		{Id: 212, Code: "role:read:detail-api", Type: 3,
			Path: sql.NullString{String: "GET:/api/perm/roles/:id", Valid: true}, MinVerfLevel: 2},
	}
	permById := make(map[int64]*model.SysPermission, len(perms))
	for _, p := range perms {
		permById[p.Id] = p
	}

	// 服务角色：grid_worker(4) 持 427/428；community_admin(3) 持 432/210/211/212
	type caseRow struct {
		name   string
		roleID int64
		permID int64
		needle string // 请求 needle（path 或 code）
	}
	cases := []caseRow{
		{"网格员删公告 427", 4, 427, "DELETE:/api/community/notices/:id"},
		{"网格员改公告 428", 4, 428, "PUT:/api/community/notices/:id"},
		{"社区管理员建活动 432", 3, 432, "POST:/api/community/activities"},
		{"社区管理员查角色 210", 3, 210, "role:read"},
		{"社区管理员查角色列表 211", 3, 211, "GET:/api/perm/roles"},
		{"社区管理员查角色详情 212", 3, 212, "GET:/api/perm/roles/:id"},
	}

	// 构造请求：ApiPath 与 Action 分离的形态（中间件传 Method 前缀 path）
	newReq := func(needle string) *permissionv1.CheckPermissionRequest {
		if strings.Contains(needle, ":") && !strings.Contains(needle, ":/api") {
			return &permissionv1.CheckPermissionRequest{UserId: 100, ApiPath: needle}
		}
		parts := strings.SplitN(needle, ":", 2)
		return &permissionv1.CheckPermissionRequest{UserId: 100, Action: parts[0], ApiPath: parts[1]}
	}

	build := func(t *testing.T, c caseRow, grant *model.UserRoleWithInfo, want bool) {
		t.Helper()
		p := permById[c.permID]
		mockUserRole := new(MockUserRoleModel)
		mockRolePerm := new(MockRolePermissionModel)
		mockPerm := new(MockPermissionModel)
		redisClient, _ := setupMiniRedis(t)

		mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(100)).Return(
			[]*model.UserRoleWithInfo{grant}, nil)
		mockRolePerm.On("FindByRoleId", mock.Anything, c.roleID).Return(
			[]*model.RelRolePermission{{RoleId: c.roleID, PermissionId: c.permID}}, nil)
		needle := c.needle
		if p.Path.Valid {
			mockPerm.On("FindByPath", mock.Anything, needle).Return(p, nil)
		} else {
			mockPerm.On("FindByPath", mock.Anything, needle).Return(nil, sql.ErrNoRows)
			mockPerm.On("FindByCode", mock.Anything, needle).Return(p, nil)
		}
		mockPerm.On("FindByIds", mock.Anything, mock.MatchedBy(func(ids []int64) bool {
			return len(ids) == 1 && ids[0] == c.permID
		})).Return([]*model.SysPermission{p}, nil)

		svcCtx := &svc.ServiceContext{
			UserRoleModel:       mockUserRole,
			RolePermissionModel: mockRolePerm,
			PermissionModel:     mockPerm,
			RedisClient:         redisClient,
		}
		resp, err := NewCheckPermissionLogic(context.Background(), svcCtx).CheckPermission(newReq(needle))
		assert.NoError(t, err)
		assert.Equalf(t, want, resp.Allowed,
			"[%s] status=%d min_verf_level=%d → allowed 应=%v", c.name, grant.URStatus, p.MinVerfLevel, want)
		mockUserRole.AssertExpectations(t)
		mockRolePerm.AssertExpectations(t)
		mockPerm.AssertExpectations(t)
	}

	for _, c := range cases {
		t.Run(c.name+"/未认证status0_拒绝", func(t *testing.T) {
			build(t, c, &model.UserRoleWithInfo{RoleId: c.roleID, ScopeType: "community", ScopeId: 100, URStatus: 0}, false)
		})
		t.Run(c.name+"/已认证status2_放行", func(t *testing.T) {
			build(t, c, &model.UserRoleWithInfo{RoleId: c.roleID, ScopeType: "community", ScopeId: 100, URStatus: 2, VerifiedAt: verifiedAt}, true)
		})
	}
}
