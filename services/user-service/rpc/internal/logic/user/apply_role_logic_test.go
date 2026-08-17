package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// §3.2 申请角色 (ApplyRole) 测试
// 角色授予已迁移到 permission-service，通过 mock PermissionClient 验证
// =============================================================================

func TestApplyRole_Owner(t *testing.T) {
	// U-A-01: 申请业主角色（绑定小区）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	createTestMembership(t, mm, 5001, 1001, 2001)

	// roleMapper 拉取角色表（owner → role_id=1）
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{{Id: 1, Code: "owner"}, {Id: 5, Code: "tenant"}, {Id: 7, Code: "merchant"}},
	}, nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "owner",
		Building: "3", Unit: "2", Room: "1501",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "owner", resp.Role.RoleCode)
	assert.Equal(t, int32(model.RoleVerfStatusUnverified), resp.Role.VerfStatus)
	assert.Equal(t, int64(2001), resp.Role.CommunityId)
}

func TestApplyRole_Merchant(t *testing.T) {
	// U-A-02: 申请商家角色（不绑小区，global 作用域）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)

	createTestUser(t, ub, 1002, "phone_1002")

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{{Id: 1, Code: "owner"}, {Id: 5, Code: "tenant"}, {Id: 7, Code: "merchant"}},
	}, nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.AssignRoleRequest, _ ...interface{}) (*permissionv1.AssignRoleResponse, error) {
			assert.Equal(t, "global", req.ScopeType)
			assert.Equal(t, int64(0), req.ScopeId)
			return &permissionv1.AssignRoleResponse{}, nil
		})

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1002, RoleCode: "merchant",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "merchant", resp.Role.RoleCode)
	assert.Equal(t, int64(0), resp.Role.CommunityId)
}

func TestApplyRole_NoMembership(t *testing.T) {
	// U-A-03: 非商家角色但无小区成员关系 → 拒绝
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	createTestUser(t, ub, 1003, "phone_1003")

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1003, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10005), resp.Base.Code) // 小区成员关系不存在
}

func TestApplyRole_UserNotFound(t *testing.T) {
	// U-A-04: 用户不存在 → 拒绝
	svc, _ := certTestSvc(t)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 9999, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10001), resp.Base.Code) // 用户不存在
}

func TestApplyRole_GridWorker(t *testing.T) {
	// U-A-05: 申请网格员角色
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1005, "phone_1005")
	createTestMembership(t, mm, 5005, 1005, 2001)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{{Id: 1, Code: "owner"}, {Id: 4, Code: "grid_worker"}},
	}, nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1005, CommunityId: 2001, RoleCode: "grid_worker",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "grid_worker", resp.Role.RoleCode)
	assert.Equal(t, int64(2001), resp.Role.CommunityId, "非 merchant 角色应绑定小区作用域")
}

func TestApplyRole_FindOneError(t *testing.T) {
	// U-A-06: FindOne 返回非 ErrNotFound 错误 → 透传错误（不走用户不存在分支）
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	ub.findErr = errors.New("db down")

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "owner",
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestApplyRole_FindMembershipError(t *testing.T) {
	// U-A-07: FindByUserAndCommunity 返回非 ErrNotFound 错误 → 透传错误（非「成员不存在」分支）
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	mm.byUserCommErr = errors.New("db down")

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "owner",
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestApplyRole_MembershipNotActive(t *testing.T) {
	// U-A-08: membership 存在但非 active（已退出）→ 10005
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	ms := createTestMembership(t, mm, 5001, 1001, 2001)
	ms.BindStatus = model.MembershipBindStatusLeft

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10005), resp.Base.Code)
}

func TestApplyRole_AssignRoleError(t *testing.T) {
	// U-A-09: permission AssignRole 失败 → 透传错误
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	createTestMembership(t, mm, 5001, 1001, 2001)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{{Id: 1, Code: "owner"}, {Id: 5, Code: "tenant"}, {Id: 7, Code: "merchant"}},
	}, nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("permission service down")).Times(1)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "owner",
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestApplyRole_RoleCodeNotFound(t *testing.T) {
	// U-A-10: role_code 在 permission-service 不存在 → 10008
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	createTestMembership(t, mm, 5001, 1001, 2001)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{{Id: 1, Code: "owner"}},
	}, nil).AnyTimes()

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "unknown_role",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10008), resp.Base.Code)
}

func TestApplyRole_PermissionClientNil(t *testing.T) {
	// U-A-11: PermissionClient 为 nil（缓存已预热）→ 50000 系统繁忙
	resetRoleMapper()
	mapper.mu.Lock()
	mapper.codeToID = map[string]int64{"owner": 1}
	mapper.idToCode = map[int64]string{1: "owner"}
	mapper.loadedAt = time.Now()
	mapper.mu.Unlock()

	svc := testSvc(t) // PermissionClient 为 nil
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	createTestMembership(t, mm, 5001, 1001, 2001)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(50000), resp.Base.Code)
}

func TestApplyRole_ServiceRoles_NoMembership_Allowed(t *testing.T) {
	// 用户拍板（有意反转 08-16 security-arch 回滚）：非 community_admin 的服务角色
	// （网格员/物业管理员）可自助申请，无需本小区带房号 membership。安全由「盖章文件 +
	// 人工审核 + 敏感权限 min_verf_level=2」保证（permission-service 并行加固：未认证
	// status=0 grant 不能行使破坏性操作）。数据权限来自角色 grant：scope_type=community,
	// scope_id=communityId, status=0 未认证待审。
	// 注意：community_admin 不在此白名单（security-arch CRITICAL 修复——其驱动 division
	// 子树发布范围展开，数据范围必须绑定 membership，见 TestApplyRole_CommunityAdmin_NoMembership_Rejected）。
	// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]
	cases := []struct {
		name     string
		roleCode string
		roleID   int64
	}{
		{"grid_worker", model.RoleCodeGridWorker, 4},
		{"property_admin", model.RoleCodePropertyAdmin, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, permMock := certTestSvc(t)
			ub := userBaseModel(svc)
			// 只建用户，不建 membership —— 免 membership 白名单服务角色可自助申请
			createTestUser(t, ub, 1005, "phone_1005_"+tc.name)

			permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
				Roles: []*permissionv1.Role{
					{Id: 4, Code: "grid_worker"},
					{Id: 5, Code: "community_admin"},
					{Id: 6, Code: "property_admin"},
				},
			}, nil).AnyTimes()
			permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, req *permissionv1.AssignRoleRequest, _ ...interface{}) (*permissionv1.AssignRoleResponse, error) {
					assert.Equal(t, "community", req.ScopeType, "服务角色作用域应为 community")
					assert.Equal(t, int64(2001), req.ScopeId)
					assert.Equal(t, tc.roleID, req.RoleId)
					require.NotNil(t, req.Status)
					assert.Equal(t, int32(0), *req.Status, "服务角色申请 status=0 未认证待审")
					return &permissionv1.AssignRoleResponse{}, nil
				})

			logic := NewApplyRoleLogic(context.Background(), svc)
			resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
				UserId: 1005, CommunityId: 2001, RoleCode: tc.roleCode,
			})

			require.NoError(t, err)
			assert.Equal(t, int32(0), resp.Base.Code, "免 membership 白名单服务角色无 membership 应允许自助申请")
			require.NotNil(t, resp.Role)
			assert.Equal(t, tc.roleCode, resp.Role.RoleCode)
			assert.Equal(t, int64(2001), resp.Role.CommunityId)
		})
	}
}

func TestApplyRole_CommunityAdmin_NoMembership_Rejected(t *testing.T) {
	// security-arch CRITICAL 回归：community_admin 驱动 resolvePublishScope 的 division
	// 子树展开（未认证 status=0 grant 亦被 grantActive 视为活跃），免 membership 自助申请
	// 会让任意注册用户对任意小区Id 获得发布范围放大。修复后 community_admin 必须绑定
	// 目标小区有效 membership —— 无则 10005 且绝不触达 permission-service AssignRole。
	// 安全断言：不注册 AssignRole 期望，若逻辑误放行则 gomock 报 unexpected call。
	// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	// 只建用户，不建 membership —— 无目标小区成员关系
	createTestUser(t, ub, 1005, "phone_1005_commadmin")

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{
			{Id: 4, Code: "grid_worker"},
			{Id: 5, Code: "community_admin"},
			{Id: 6, Code: "property_admin"},
		},
	}, nil).AnyTimes()
	// 注意：不 EXPECT AssignRole —— community_admin 无 membership 必须被拦在 AssignRole 之前

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1005, CommunityId: 2001, RoleCode: model.RoleCodeCommunityAdmin,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10005), resp.Base.Code, "community_admin 无 membership 应返回 10005")
	require.Nil(t, resp.Role, "拒绝时不返回角色")
}

func TestApplyRole_ServiceRoles_WithMembership(t *testing.T) {
	// 服务角色 + 有效 membership → 申请成功（grid_worker/property_admin 免 membership；
	// community_admin 需 membership，此处已绑定 → 正常申请）
	cases := []struct {
		name     string
		roleCode string
		roleID   int64
	}{
		{"grid_worker", model.RoleCodeGridWorker, 4},
		{"community_admin", model.RoleCodeCommunityAdmin, 5},
		{"property_admin", model.RoleCodePropertyAdmin, 6},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, permMock := certTestSvc(t)
			ub := userBaseModel(svc)
			mm := membershipModel(svc)

			createTestUser(t, ub, 1005, "phone_1005_"+tc.name)
			createTestMembership(t, mm, 5005, 1005, 2001)

			permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
				Roles: []*permissionv1.Role{
					{Id: 4, Code: "grid_worker"},
					{Id: 5, Code: "community_admin"},
					{Id: 6, Code: "property_admin"},
				},
			}, nil).AnyTimes()
			permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, req *permissionv1.AssignRoleRequest, _ ...interface{}) (*permissionv1.AssignRoleResponse, error) {
					assert.Equal(t, "community", req.ScopeType, "服务角色作用域应为 community")
					assert.Equal(t, int64(2001), req.ScopeId)
					assert.Equal(t, tc.roleID, req.RoleId)
					return &permissionv1.AssignRoleResponse{}, nil
				})

			logic := NewApplyRoleLogic(context.Background(), svc)
			resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
				UserId: 1005, CommunityId: 2001, RoleCode: tc.roleCode,
			})

			require.NoError(t, err)
			assert.Equal(t, int32(0), resp.Base.Code, "服务角色 + membership 应申请成功")
			assert.Equal(t, tc.roleCode, resp.Role.RoleCode)
			assert.Equal(t, int64(2001), resp.Role.CommunityId)
		})
	}
}

func TestApplyRole_ServiceRoles_MembershipInactive(t *testing.T) {
	// 服务角色 + membership 已退出（bind_status=0）→ 仍允许申请（服务角色不查 membership，
	// 与 WithMembership 一起覆盖「服务角色申请不依赖 membership 状态」）。
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1005, "phone_1005")
	ms := createTestMembership(t, mm, 5005, 1005, 2001)
	ms.BindStatus = model.MembershipBindStatusLeft

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{{Id: 4, Code: "grid_worker"}},
	}, nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.AssignRoleRequest, _ ...interface{}) (*permissionv1.AssignRoleResponse, error) {
			assert.Equal(t, "community", req.ScopeType)
			assert.Equal(t, int64(2001), req.ScopeId)
			return &permissionv1.AssignRoleResponse{}, nil
		})

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1005, CommunityId: 2001, RoleCode: model.RoleCodeGridWorker,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "服务角色免 membership，membership 已退出也允许申请")
	assert.Equal(t, int64(2001), resp.Role.CommunityId)
}

func TestApplyRole_ResidentRoles_NoMembership(t *testing.T) {
	// 居民角色（owner/tenant/committee）无带房号 membership → 仍 10005
	cases := []struct {
		name     string
		roleCode string
	}{
		{"owner", model.RoleCodeOwner},
		{"tenant", model.RoleCodeTenant},
		{"committee", model.RoleCodeCommittee},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := certTestSvc(t)
			ub := userBaseModel(svc)
			createTestUser(t, ub, 1006, "phone_1006_"+tc.name)

			logic := NewApplyRoleLogic(context.Background(), svc)
			resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
				UserId: 1006, CommunityId: 2001, RoleCode: tc.roleCode,
			})

			require.NoError(t, err)
			assert.Equal(t, int32(10005), resp.Base.Code, "居民角色无 membership 应返回 10005")
		})
	}
}

func TestNeedMembership(t *testing.T) {
	// 安全模型（security-arch CRITICAL 修复后）：
	//   居民角色（owner/tenant/committee）数据范围绑定带房号 membership → true；
	//   community_admin 特权角色驱动 division 子树发布范围展开（resolvePublishScope），
	//     未认证(status=0) grant 亦被 grantActive 视为活跃，故其数据范围必须绑定 membership
	//     ——禁止免 membership 对任意小区自助申请 → true（见安全评审 CRITICAL）；
	//   仅显式白名单（grid_worker/property_admin/merchant）可免 membership 自助申请 → false；
	//   未知/未来角色 fail-closed（default → true，需 membership），防潜在提权地雷。
	cases := []struct {
		name     string
		roleCode string
		want     bool
	}{
		{"owner", model.RoleCodeOwner, true},
		{"tenant", model.RoleCodeTenant, true},
		{"committee", model.RoleCodeCommittee, true},
		{"community_admin", model.RoleCodeCommunityAdmin, true},
		{"grid_worker", model.RoleCodeGridWorker, false},
		{"property_admin", model.RoleCodePropertyAdmin, false},
		{"merchant", model.RoleCodeMerchant, false},
		// fail-closed：未来在 permission-service 新增的特权角色（如 super_admin/moderator）
		// 若未同步白名单，将默认要求 membership，不得免 membership 自助申请（防提权地雷）。
		{"unknown_role_fail_closed", "super_admin", true},
		{"empty_role_fail_closed", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, needMembership(tc.roleCode), "needMembership(%s)", tc.roleCode)
		})
	}
}

// =============================================================================
