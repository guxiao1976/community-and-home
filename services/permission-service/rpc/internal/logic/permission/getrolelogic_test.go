package permission

import (
	"context"
	"database/sql"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGetRole_Success 测试获取单个角色详情（含权限）
func TestGetRole_Success(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(5)).Return(&model.SysRole{
		Id: 5, RoleCode: "owner", RoleName: "业主", Status: 1, IsSystem: 1,
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(5)).Return([]*model.RelRolePermission{
		{RoleId: 5, PermissionId: 435},
		{RoleId: 5, PermissionId: 600},
	}, nil)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByIds", mock.Anything, []int64{435, 600}).Return([]*model.SysPermission{
		{Id: 435, Code: "community:lostfound:create-api", Path: sql.NullString{String: "POST:/api", Valid: true}},
		{Id: 600, Code: "committee:election:vote"},
	}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
	}

	logic := NewGetRoleLogic(context.Background(), svcCtx)
	resp, err := logic.GetRole(&permissionv1.GetRoleRequest{Id: 5})

	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(5), resp.Role.Id)
	assert.Equal(t, "owner", resp.Role.Code)
	assert.Len(t, resp.Role.Permissions, 2)
	assert.True(t, resp.Role.IsSystem)
	mockRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}

// TestGetRole_NotFound 测试角色不存在 → 60001
func TestGetRole_NotFound(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewGetRoleLogic(context.Background(), svcCtx)

	resp, err := logic.GetRole(&permissionv1.GetRoleRequest{Id: 999})
	assert.NoError(t, err)
	assert.Equal(t, int32(60001), resp.Base.Code)
	assert.Contains(t, resp.Base.Msg, "角色不存在")
	mockRole.AssertExpectations(t)
}

// TestGetRole_FindByRoleIdError 覆盖 FindByRoleId 返回 error（错误被忽略 → permIds 为空 → 无权限）
func TestGetRole_FindByRoleIdError(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(5)).Return(&model.SysRole{Id: 5, RoleCode: "owner"}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(5)).Return(nil, assert.AnError)

	// FindByIds 不应被调用（permIds 为空）
	mockPerm := new(MockPermissionModel)

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
	}

	logic := NewGetRoleLogic(context.Background(), svcCtx)
	resp, err := logic.GetRole(&permissionv1.GetRoleRequest{Id: 5})

	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "owner", resp.Role.Code)
	assert.Empty(t, resp.Role.Permissions)
	mockRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}

// TestGetRole_NoPerms 覆盖角色存在但无权限关联 → permIds 为空
func TestGetRole_NoPerms(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(5)).Return(&model.SysRole{Id: 5, RoleCode: "owner"}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(5)).Return([]*model.RelRolePermission{}, nil)

	mockPerm := new(MockPermissionModel)

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
	}

	logic := NewGetRoleLogic(context.Background(), svcCtx)
	resp, err := logic.GetRole(&permissionv1.GetRoleRequest{Id: 5})

	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Empty(t, resp.Role.Permissions)
	mockRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}
