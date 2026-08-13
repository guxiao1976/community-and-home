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

// =============================================================================
