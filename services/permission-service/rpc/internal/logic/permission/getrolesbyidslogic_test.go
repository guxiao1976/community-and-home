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

// TestGetRolesByIds_Success 测试批量查询角色成功
func TestGetRolesByIds_Success(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByIds", mock.Anything, []int64{1, 2}).Return([]*model.SysRole{
		{Id: 1, RoleCode: "owner", RoleName: "业主", Status: 1},
		{Id: 2, RoleCode: "property_admin", RoleName: "物业管理员", Status: 1},
	}, nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewGetRolesByIdsLogic(context.Background(), svcCtx)

	resp, err := logic.GetRolesByIds(&permissionv1.GetRolesByIdsRequest{Ids: []int64{1, 2}})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.Roles, 2)
	assert.Equal(t, "owner", resp.Roles[0].Code)
	assert.Equal(t, "property_admin", resp.Roles[1].Code)
	mockRole.AssertExpectations(t)
}

// TestGetRolesByIds_EmptyIds 测试空 ID 列表 → Roles 为 nil（不查 DB）
func TestGetRolesByIds_EmptyIds(t *testing.T) {
	mockRole := new(MockRoleModel) // 不设 FindByIds expectation：空列表不应查 DB

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewGetRolesByIdsLogic(context.Background(), svcCtx)

	resp, err := logic.GetRolesByIds(&permissionv1.GetRolesByIdsRequest{})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Nil(t, resp.Roles)
	mockRole.AssertExpectations(t)
}

// TestGetRolesByIds_FindByIdsError 测试查询失败 → 透传 error
func TestGetRolesByIds_FindByIdsError(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByIds", mock.Anything, []int64{1, 2}).Return(nil, assert.AnError)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewGetRolesByIdsLogic(context.Background(), svcCtx)

	resp, err := logic.GetRolesByIds(&permissionv1.GetRolesByIdsRequest{Ids: []int64{1, 2}})
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRole.AssertExpectations(t)
}

// TestGetRolesByIds_EmptyResult 覆盖 DB 返回空列表 → Roles 为空（非 nil）
func TestGetRolesByIds_EmptyResult(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindByIds", mock.Anything, []int64{1, 2}).Return([]*model.SysRole{}, nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewGetRolesByIdsLogic(context.Background(), svcCtx)

	resp, err := logic.GetRolesByIds(&permissionv1.GetRolesByIdsRequest{Ids: []int64{1, 2}})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Empty(t, resp.Roles)
	mockRole.AssertExpectations(t)
}
