package user

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	permissionmocks "github.com/guxiao1976/api-proto/gen/go/permission/v1/mocks"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getRolesTestSvc 创建带 permission mock 的 ServiceContext（复用 certTestSvc）
func getRolesTestSvc(t *testing.T) (*svc.ServiceContext, *permissionmocks.MockPermissionServiceClient) {
	return certTestSvc(t)
}

func TestGetUserRoles_ReturnsCertified(t *testing.T) {
	// U-G-01: 返回用户已认证角色（status=2）
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 2, 2001)}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	req := &userv1.GetUserRolesRequest{
		UserId:     7001,
		VerfStatus: int32Ptr(2), // 只取已认证
	}
	resp, err := logic.GetUserRoles(req)

	require.NoError(t, err)
	require.Len(t, resp.Roles, 1)
	assert.Equal(t, "owner", resp.Roles[0].RoleCode)
	assert.Equal(t, int32(2), resp.Roles[0].VerfStatus)
}

func TestGetUserRoles_AllStatuses(t *testing.T) {
	// U-G-02: 不传 verf_status 返回所有状态
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{
			{Role: &permissionv1.Role{Id: 1, Code: "owner"}, ScopeType: "community", ScopeId: 2001, Status: 0},
			{Role: &permissionv1.Role{Id: 2, Code: "property_admin"}, ScopeType: "community", ScopeId: 2001, Status: 2},
		}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 7001})

	require.NoError(t, err)
	assert.Len(t, resp.Roles, 2)
}

func TestGetUserRoles_NoRoles(t *testing.T) {
	// U-G-03: 无角色返回空列表
	svc, permMock := getRolesTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{}}, nil)

	logic := NewGetUserRolesLogic(context.Background(), svc)
	resp, err := logic.GetUserRoles(&userv1.GetUserRolesRequest{UserId: 9999})

	require.NoError(t, err)
	assert.Empty(t, resp.Roles)
}
