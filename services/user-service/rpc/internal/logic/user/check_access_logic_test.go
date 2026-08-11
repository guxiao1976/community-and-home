package user

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// §3.3 权限校验 (CheckAccess) 测试
// 鉴权从 permission-service 获取已认证角色（status=2）
// =============================================================================

func TestCheckAccess_OwnerAllowed(t *testing.T) {
	// U-C-01: owner 已认证角色校验通过
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.GetUserRolesResponse{
		Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 2},
		},
	}, nil)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6001, RoleCodes: []string{"owner"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.True(t, resp.Allowed)
	assert.Equal(t, "owner", resp.MatchedRole)
}

func TestCheckAccess_WrongRole(t *testing.T) {
	// U-C-02: 用户有 owner，但请求 grid_worker → 拒绝
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.GetUserRolesResponse{
		Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 2},
		},
	}, nil)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6001, RoleCodes: []string{"grid_worker"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckAccess_WrongCommunity(t *testing.T) {
	// U-C-03: owner 但在不同小区 → 拒绝
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.GetUserRolesResponse{
		Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 2},
		},
	}, nil)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6001, RoleCodes: []string{"owner"}, CommunityId: 9999,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckAccess_UnverifiedRole(t *testing.T) {
	// U-C-04: 角色未认证（status=0）→ 拒绝
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.GetUserRolesResponse{
		Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 0},
		},
	}, nil)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6001, RoleCodes: []string{"owner"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckAccess_MerchantGlobal(t *testing.T) {
	// U-C-05: merchant 全局角色，任何 community 都放行
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.GetUserRolesResponse{
		Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 7, Code: "merchant"}, ScopeType: "global", ScopeId: 0, Status: 2},
		},
	}, nil)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6001, RoleCodes: []string{"merchant"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.True(t, resp.Allowed)
}

func TestCheckAccess_ExpiredRole(t *testing.T) {
	// U-C-06: 角色已过期（status=4）→ 拒绝
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.GetUserRolesResponse{
		Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 4},
		},
	}, nil)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6001, RoleCodes: []string{"owner"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed)
}
