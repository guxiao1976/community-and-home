package permission

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// SEE: [[testing-discipline]] — CheckPermission 核心场景测试：系统角色走 rel_role_permission（无短路）、普通用户权限匹配、拒绝、缓存

// MockUserRoleModel mocks RelUserRoleModel interface
type MockUserRoleModel struct {
	mock.Mock
}

func (m *MockUserRoleModel) Insert(ctx context.Context, data *model.RelUserRole) (int64, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRoleModel) InsertIgnore(ctx context.Context, data *model.RelUserRole) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*model.RelUserRole, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelUserRole), args.Error(1)
}

func (m *MockUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserRoleWithInfo), args.Error(1)
}

func (m *MockUserRoleModel) FindActiveRolesByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserRoleWithInfo), args.Error(1)
}

func (m *MockUserRoleModel) FindScopesByUserId(ctx context.Context, userId int64, scopeType string) ([]int64, error) {
	args := m.Called(ctx, userId, scopeType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockUserRoleModel) DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64, scopeType string, scopeId int64) error {
	args := m.Called(ctx, userId, roleId, scopeType, scopeId)
	return args.Error(0)
}

func (m *MockUserRoleModel) BatchInsertUserRoles(ctx context.Context, records []*model.RelUserRole) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

func (m *MockUserRoleModel) CountByRoleId(ctx context.Context, roleId int64) (int64, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRoleModel) FindAllByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserRoleWithInfo), args.Error(1)
}

func (m *MockUserRoleModel) UpdateRoleStatus(ctx context.Context, userId, roleId int64, scopeType string, scopeId, status int64, verifiedAt, expiresAt sql.NullTime) error {
	args := m.Called(ctx, userId, roleId, scopeType, scopeId, status, verifiedAt, expiresAt)
	return args.Error(0)
}

func (m *MockUserRoleModel) CountActiveByRoleAndScope(ctx context.Context, roleId int64, scopeType string, scopeId, excludeUserId int64) (int64, error) {
	args := m.Called(ctx, roleId, scopeType, scopeId, excludeUserId)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRoleModel) FindByRoleId(ctx context.Context, roleId int64) ([]*model.RelUserRole, error) {
	args := m.Called(ctx, roleId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelUserRole), args.Error(1)
}

// MockRolePermissionModel mocks RelRolePermissionModel interface
type MockRolePermissionModel struct {
	mock.Mock
}

func (m *MockRolePermissionModel) Insert(ctx context.Context, data *model.RelRolePermission) (int64, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRolePermissionModel) FindByRoleId(ctx context.Context, roleId int64) ([]*model.RelRolePermission, error) {
	args := m.Called(ctx, roleId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelRolePermission), args.Error(1)
}

func (m *MockRolePermissionModel) DeleteByRoleId(ctx context.Context, roleId int64) error {
	args := m.Called(ctx, roleId)
	return args.Error(0)
}

func (m *MockRolePermissionModel) BatchInsert(ctx context.Context, records []*model.RelRolePermission) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

// MockPermissionModel mocks SysPermissionModel interface
type MockPermissionModel struct {
	mock.Mock
}

func (m *MockPermissionModel) FindAll(ctx context.Context) ([]*model.SysPermission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysPermission), args.Error(1)
}

func (m *MockPermissionModel) FindByIds(ctx context.Context, ids []int64) ([]*model.SysPermission, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysPermission), args.Error(1)
}

func (m *MockPermissionModel) FindByCode(ctx context.Context, code string) (*model.SysPermission, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SysPermission), args.Error(1)
}

func (m *MockPermissionModel) FindByPath(ctx context.Context, path string) (*model.SysPermission, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SysPermission), args.Error(1)
}

func (m *MockPermissionModel) FindWithFilter(ctx context.Context, typeFilter, statusFilter *int64) ([]*model.SysPermission, error) {
	args := m.Called(ctx, typeFilter, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysPermission), args.Error(1)
}

func setupMiniRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client, mr
}

// TestCheckPermission_CapabilityLayering — T1.5 能力分层聚合规则（对照 access-control-design §5.1.1 矩阵）
// SEE: [[is-system-no-permission-shortcut]] — registered_user 权限经 rel_role_permission 配置，禁止角色名短路
// SEE: [[permission-seed-api-path-must-match-routes]] — needle 为 path 含 METHOD 前缀

func TestCheckPermission_CapabilityLayering(t *testing.T) {
	publishPerm := &model.SysPermission{
		Id: 435, Code: "community:lostfound:create-api", Type: 3,
		Path: sql.NullString{String: "POST:/api/community/lostfound", Valid: true}, MinVerfLevel: 0,
	}
	electionPerm := &model.SysPermission{
		Id: 600, Code: "committee:election:vote", Type: 2,
		Path: sql.NullString{}, MinVerfLevel: 2,
	}
	browsePerm := &model.SysPermission{
		Id: 422, Code: "community:notice:read-list-api", Type: 3,
		Path: sql.NullString{String: "GET:/api/community/notices", Valid: true}, MinVerfLevel: 0,
	}
	verifiedAt := sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true}
	expiredAt := sql.NullTime{Time: time.Now().Add(-24 * time.Hour), Valid: true}

	tests := []struct {
		name        string
		grants      []*model.UserRoleWithInfo
		rolePerms   map[int64][]*model.RelRolePermission
		perms       []*model.SysPermission
		req         *permissionv1.CheckPermissionRequest
		wantAllowed bool
	}{
		{
			name:   "未认证业主发布✅（status=0, min_verf_level=0）",
			grants: []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0}},
			rolePerms: map[int64][]*model.RelRolePermission{
				1: {{RoleId: 1, PermissionId: 435}},
			},
			perms:       []*model.SysPermission{publishPerm},
			req:         &permissionv1.CheckPermissionRequest{UserId: 1001, Action: "POST", ApiPath: "/api/community/lostfound"},
			wantAllowed: true,
		},
		{
			name:   "未认证业主选举❌（status=0, min_verf_level=2）",
			grants: []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0}},
			rolePerms: map[int64][]*model.RelRolePermission{
				1: {{RoleId: 1, PermissionId: 600}},
			},
			perms:       []*model.SysPermission{electionPerm},
			req:         &permissionv1.CheckPermissionRequest{UserId: 1002, ApiPath: "committee:election:vote"},
			wantAllowed: false,
		},
		{
			name:   "认证业主选举✅（status=2 + verified_at NOT NULL）",
			grants: []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 2, VerifiedAt: verifiedAt}},
			rolePerms: map[int64][]*model.RelRolePermission{
				1: {{RoleId: 1, PermissionId: 600}},
			},
			perms:       []*model.SysPermission{electionPerm},
			req:         &permissionv1.CheckPermissionRequest{UserId: 1003, ApiPath: "committee:election:vote"},
			wantAllowed: true,
		},
		{
			name:   "待审发布✅（status=1）",
			grants: []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 1}},
			rolePerms: map[int64][]*model.RelRolePermission{
				1: {{RoleId: 1, PermissionId: 435}},
			},
			perms:       []*model.SysPermission{publishPerm},
			req:         &permissionv1.CheckPermissionRequest{UserId: 1004, Action: "POST", ApiPath: "/api/community/lostfound"},
			wantAllowed: true,
		},
		{
			name:   "已过期❌（expires_at < NOW()）",
			grants: []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 2, VerifiedAt: verifiedAt, ExpiresAt: expiredAt}},
			rolePerms: map[int64][]*model.RelRolePermission{
				1: {{RoleId: 1, PermissionId: 435}},
			},
			perms:       []*model.SysPermission{publishPerm},
			req:         &permissionv1.CheckPermissionRequest{UserId: 1005, Action: "POST", ApiPath: "/api/community/lostfound"},
			wantAllowed: false,
		},
		{
			name: "多角色叠加取最高✅（owner status=0 + committee status=2 verified）",
			grants: []*model.UserRoleWithInfo{
				{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
				{RoleId: 6, ScopeType: "community", ScopeId: 100, URStatus: 2, VerifiedAt: verifiedAt},
			},
			rolePerms: map[int64][]*model.RelRolePermission{
				1: {{RoleId: 1, PermissionId: 600}},
				6: {{RoleId: 6, PermissionId: 600}},
			},
			perms:       []*model.SysPermission{electionPerm},
			req:         &permissionv1.CheckPermissionRequest{UserId: 1006, ApiPath: "committee:election:vote"},
			wantAllowed: true,
		},
		{
			name: "registered_user 仅 browse✅且不满足 level-2✅（status=2, verified_at NULL）",
			grants: []*model.UserRoleWithInfo{
				{RoleId: 9, ScopeType: "", ScopeId: 0, URStatus: 2}, // registered_user: verified_at NULL
			},
			rolePerms: map[int64][]*model.RelRolePermission{
				9: {{RoleId: 9, PermissionId: 422}, {RoleId: 9, PermissionId: 600}},
			},
			perms:       []*model.SysPermission{browsePerm, electionPerm},
			req:         &permissionv1.CheckPermissionRequest{UserId: 1007, Action: "GET", ApiPath: "/api/community/notices"},
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构建 needle
			needle := tt.req.ApiPath
			if tt.req.Action != "" && !strings.HasPrefix(tt.req.ApiPath, tt.req.Action+":") {
				needle = tt.req.Action + ":" + tt.req.ApiPath
			}

			mockUserRole := new(MockUserRoleModel)
			mockRolePerm := new(MockRolePermissionModel)
			mockPerm := new(MockPermissionModel)
			redisClient, _ := setupMiniRedis(t)

			mockUserRole.On("FindActiveRolesByUserId", mock.Anything, tt.req.UserId).Return(tt.grants, nil)
			for roleID, rps := range tt.rolePerms {
				mockRolePerm.On("FindByRoleId", mock.Anything, roleID).Return(rps, nil)
			}

			// 权限定义查找（perm:def 缓存回源）：先 FindByPath，再 FindByCode 兜底
			var matched *model.SysPermission
			for _, p := range tt.perms {
				if p.Path.Valid && p.Path.String == needle {
					matched = p
				} else if !p.Path.Valid && p.Code == needle {
					matched = p
				}
			}
			if matched != nil {
				if matched.Path.Valid {
					mockPerm.On("FindByPath", mock.Anything, needle).Return(matched, nil)
				} else {
					mockPerm.On("FindByPath", mock.Anything, needle).Return(nil, sql.ErrNoRows)
					mockPerm.On("FindByCode", mock.Anything, needle).Return(matched, nil)
				}
			}

			// 权限集合（用户聚合）— 顺序无关（实现端 map 迭代）
			var wantPermIds []int64
			for _, p := range tt.perms {
				wantPermIds = append(wantPermIds, p.Id)
			}
			mockPerm.On("FindByIds", mock.Anything, mock.MatchedBy(func(ids []int64) bool {
				if len(ids) != len(wantPermIds) {
					return false
				}
				got := make(map[int64]struct{}, len(ids))
				for _, id := range ids {
					got[id] = struct{}{}
				}
				for _, id := range wantPermIds {
					if _, ok := got[id]; !ok {
						return false
					}
				}
				return true
			})).Return(tt.perms, nil)

			svcCtx := &svc.ServiceContext{
				UserRoleModel:       mockUserRole,
				RolePermissionModel: mockRolePerm,
				PermissionModel:     mockPerm,
				RedisClient:         redisClient,
			}

			logic := NewCheckPermissionLogic(context.Background(), svcCtx)
			resp, err := logic.CheckPermission(tt.req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, resp.Allowed, "CheckPermission 期望 allowed=%v（needle=%s）", tt.wantAllowed, needle)
			mockUserRole.AssertExpectations(t)
			mockRolePerm.AssertExpectations(t)
			mockPerm.AssertExpectations(t)
		})
	}
}

// TestCheckPermission_RegisteredUser_NotLevel2 — registered_user 对 level-2 权限必须拒绝（S2 数据不变量）
func TestCheckPermission_RegisteredUser_NotLevel2(t *testing.T) {
	electionPerm := &model.SysPermission{
		Id: 600, Code: "committee:election:vote", Type: 2,
		Path: sql.NullString{}, MinVerfLevel: 2,
	}

	mockUserRole := new(MockUserRoleModel)
	mockRolePerm := new(MockRolePermissionModel)
	mockPerm := new(MockPermissionModel)
	redisClient, _ := setupMiniRedis(t)

	// registered_user status=2, verified_at NULL
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1008)).Return([]*model.UserRoleWithInfo{
		{RoleId: 9, ScopeType: "", ScopeId: 0, URStatus: 2},
	}, nil)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(9)).Return([]*model.RelRolePermission{
		{RoleId: 9, PermissionId: 600},
	}, nil)
	mockPerm.On("FindByPath", mock.Anything, "committee:election:vote").Return(nil, sql.ErrNoRows)
	mockPerm.On("FindByCode", mock.Anything, "committee:election:vote").Return(electionPerm, nil)
	mockPerm.On("FindByIds", mock.Anything, []int64{600}).Return([]*model.SysPermission{electionPerm}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
		RedisClient:         redisClient,
	}

	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  1008,
		ApiPath: "committee:election:vote",
	})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "registered_user 即使被误配 level-2 权限也不应放行（verified_at=NULL → 恒 level-0）")
}

// TestCheckPermission_NoRoles_Denied — 无任何活跃角色拒绝
func TestCheckPermission_NoRoles_Denied(t *testing.T) {
	mockUserRole := new(MockUserRoleModel)
	mockPerm := new(MockPermissionModel)
	redisClient, _ := setupMiniRedis(t)

	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(500)).Return([]*model.UserRoleWithInfo{}, nil)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read:list-api", Type: 3,
		Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel:   mockUserRole,
		PermissionModel: mockPerm,
		RedisClient:     redisClient,
	}

	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  500,
		Action:  "GET",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "无角色应该拒绝")
}

// TestCheckPermission_UserDisabled — 用户禁用标记命中拒绝
func TestCheckPermission_UserDisabled(t *testing.T) {
	redisClient, mr := setupMiniRedis(t)
	mr.Set("user:disabled:600", "1")

	svcCtx := &svc.ServiceContext{RedisClient: redisClient}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  600,
		Action:  "GET",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "禁用用户应拒绝")
}
