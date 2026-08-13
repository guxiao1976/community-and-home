package permission

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGetUserRoles_Success 测试成功获取用户角色
func TestGetUserRoles_Success(t *testing.T) {
	// Setup
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).
		Return([]*model.UserRoleWithInfo{
			{
				RoleId:      1,
				RoleCode:    "owner",
				RoleName:    "业主",
				Description: "小区业主",
				IsSystem:    0,
				Status:      1,
				Platforms:   "pc,mobile",
				ScopeType:   "community",
				ScopeId:     100,
				URStatus:    2,
			},
			{
				RoleId:      2,
				RoleCode:    "property_admin",
				RoleName:    "物业管理员",
				Description: "物业管理人员",
				IsSystem:    0,
				Status:      1,
				Platforms:   "pc",
				ScopeType:   "building",
				ScopeId:     200,
				URStatus:    2,
			},
		}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
	}

	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{
		UserId: 1001,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.Roles, 2)

	// 验证第一个角色
	assert.Equal(t, int64(1), resp.Roles[0].Role.Id)
	assert.Equal(t, "owner", resp.Roles[0].Role.Code)
	assert.Equal(t, "业主", resp.Roles[0].Role.Name)
	assert.Equal(t, "community", resp.Roles[0].ScopeType)
	assert.Equal(t, int64(100), resp.Roles[0].ScopeId)
	// Platforms 透出断言：owner 双端 → ["pc","mobile"]
	// SEE: [[is-system-no-permission-shortcut]] — platforms 为配置属性，仅透出供 auth 端准入判定
	assert.Equal(t, []string{"pc", "mobile"}, resp.Roles[0].Role.Platforms)

	// 验证第二个角色
	assert.Equal(t, int64(2), resp.Roles[1].Role.Id)
	assert.Equal(t, "property_admin", resp.Roles[1].Role.Code)
	assert.Equal(t, "building", resp.Roles[1].ScopeType)
	// Platforms 透出断言：property_admin 单端 → ["pc"]
	assert.Equal(t, []string{"pc"}, resp.Roles[1].Role.Platforms)

	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_NoRoles 测试用户无角色
func TestGetUserRoles_NoRoles(t *testing.T) {
	// Setup
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).
		Return([]*model.UserRoleWithInfo{}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
	}

	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{
		UserId: 1001,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Nil(t, resp.Roles)

	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_SystemRole 测试系统角色
func TestGetUserRoles_SystemRole(t *testing.T) {
	// Setup
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).
		Return([]*model.UserRoleWithInfo{
			{
				RoleId:      999,
				RoleCode:    "super_admin",
				RoleName:    "超级管理员",
				Description: "系统超级管理员",
				IsSystem:    1, // 系统角色
				Status:      1,
				ScopeType:   "system",
				ScopeId:     0,
			},
		}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
	}

	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{
		UserId: 1001,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.Roles, 1)
	assert.True(t, resp.Roles[0].Role.IsSystem) // 系统角色标记

	mockUserRole.AssertExpectations(t)
}
