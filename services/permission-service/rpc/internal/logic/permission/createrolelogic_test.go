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

// TestCreateRole_Success 测试创建角色成功（含权限关联 + SortOrder）
func TestCreateRole_Success(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByCode", mock.Anything, "owner").Return(nil, sql.ErrNoRows)
	mockRole.On("Insert", mock.Anything, mock.MatchedBy(func(r *model.SysRole) bool {
		return r.RoleCode == "owner" && r.RoleName == "业主" && r.Status == 1 && r.SortOrder == 10
	})).Return(int64(5), nil)
	mockRole.On("FindOne", mock.Anything, int64(5)).Return(&model.SysRole{
		Id: 5, RoleCode: "owner", RoleName: "业主", Status: 1, SortOrder: 10,
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("DeleteByRoleId", mock.Anything, int64(5)).Return(nil)
	mockRolePerm.On("BatchInsert", mock.Anything, mock.MatchedBy(func(rs []*model.RelRolePermission) bool {
		return len(rs) == 2 && rs[0].RoleId == 5 && rs[0].PermissionId == 435 && rs[1].PermissionId == 600
	})).Return(nil)

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

	logic := NewCreateRoleLogic(context.Background(), svcCtx)
	resp, err := logic.CreateRole(&permissionv1.CreateRoleRequest{
		Code:          "owner",
		Name:          "业主",
		Description:   "小区业主",
		SortOrder:     10,
		PermissionIds: []int64{435, 600},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(5), resp.Role.Id)
	assert.Equal(t, "owner", resp.Role.Code)
	assert.Len(t, resp.Role.Permissions, 2, "应返回权限列表")
	mockRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}

// TestCreateRole_Success_NoPermissions 测试创建角色成功（无权限关联，SortOrder 缺省）
func TestCreateRole_Success_NoPermissions(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByCode", mock.Anything, "grid_worker").Return(nil, sql.ErrNoRows)
	mockRole.On("Insert", mock.Anything, mock.MatchedBy(func(r *model.SysRole) bool {
		return r.RoleCode == "grid_worker" && r.Status == 1 && r.SortOrder == 0
	})).Return(int64(6), nil)
	mockRole.On("FindOne", mock.Anything, int64(6)).Return(&model.SysRole{
		Id: 6, RoleCode: "grid_worker", RoleName: "网格员", Status: 1,
	}, nil)

	// 无权限关联：DeleteByRoleId / BatchInsert 均不应被调用
	mockRolePerm := new(MockRolePermissionModel)
	mockPerm := new(MockPermissionModel) // FindByIds 不应被调用（permIds 为空）

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
	}

	logic := NewCreateRoleLogic(context.Background(), svcCtx)
	resp, err := logic.CreateRole(&permissionv1.CreateRoleRequest{
		Code: "grid_worker",
		Name: "网格员",
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(6), resp.Role.Id)
	assert.Empty(t, resp.Role.Permissions)
	mockRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}

// TestCreateRole_DuplicateCode 测试角色编码已存在 → 60006
func TestCreateRole_DuplicateCode(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByCode", mock.Anything, "owner").Return(&model.SysRole{Id: 1, RoleCode: "owner"}, nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewCreateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.CreateRole(&permissionv1.CreateRoleRequest{Code: "owner", Name: "业主"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(60006), resp.Base.Code, "角色编码已存在应返回 60006")
	assert.Contains(t, resp.Base.Msg, "已存在")
	mockRole.AssertExpectations(t)
}

// TestCreateRole_InsertFailed 测试插入角色失败 → 透传 error
func TestCreateRole_InsertFailed(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByCode", mock.Anything, "owner").Return(nil, sql.ErrNoRows)
	mockRole.On("Insert", mock.Anything, mock.Anything).Return(int64(0), assert.AnError)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewCreateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.CreateRole(&permissionv1.CreateRoleRequest{Code: "owner", Name: "业主"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRole.AssertExpectations(t)
}

// TestCreateRole_BatchInsertFailed 测试批量插入权限失败 → 透传 error
func TestCreateRole_BatchInsertFailed(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByCode", mock.Anything, "owner").Return(nil, sql.ErrNoRows)
	mockRole.On("Insert", mock.Anything, mock.Anything).Return(int64(5), nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("DeleteByRoleId", mock.Anything, int64(5)).Return(nil)
	mockRolePerm.On("BatchInsert", mock.Anything, mock.Anything).Return(assert.AnError)

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
	}

	logic := NewCreateRoleLogic(context.Background(), svcCtx)
	resp, err := logic.CreateRole(&permissionv1.CreateRoleRequest{
		Code:          "owner",
		Name:          "业主",
		PermissionIds: []int64{435},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}

// TestCreateRole_FindOneFailed 测试插入后查询角色失败 → 透传 error
func TestCreateRole_FindOneFailed(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByCode", mock.Anything, "owner").Return(nil, sql.ErrNoRows)
	mockRole.On("Insert", mock.Anything, mock.Anything).Return(int64(5), nil)
	mockRole.On("FindOne", mock.Anything, int64(5)).Return(nil, assert.AnError)

	mockRolePerm := new(MockRolePermissionModel) // 无权限关联

	svcCtx := &svc.ServiceContext{RoleModel: mockRole, RolePermissionModel: mockRolePerm}
	logic := NewCreateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.CreateRole(&permissionv1.CreateRoleRequest{Code: "owner", Name: "业主"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRole.AssertExpectations(t)
}

// TestCreateRole_FindByCodeError 覆盖 FindByCode 返回非 ErrNotFound 错误仍继续（编码唯一性校验只拦截 err==nil 命中）
func TestCreateRole_FindByCodeError(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByCode", mock.Anything, "owner").Return(nil, assert.AnError)
	mockRole.On("Insert", mock.Anything, mock.Anything).Return(int64(7), nil)
	mockRole.On("FindOne", mock.Anything, int64(7)).Return(&model.SysRole{Id: 7, RoleCode: "owner"}, nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewCreateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.CreateRole(&permissionv1.CreateRoleRequest{Code: "owner", Name: "业主"})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	mockRole.AssertExpectations(t)
}
