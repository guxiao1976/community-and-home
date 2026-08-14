package perm

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssignRolePermissions_PreservesPlatforms — 先 GetRole 读当前 platforms，UpdateRole 显式保留（REQ-PLAT-8）
// RED: 当前实现直接 UpdateRole 不传 Platforms → 捕获请求 Platforms 为空 ≠ ["mobile"] → FAIL
// SEE: [[verify-api-before-calling]] — 先 GetRole 验证角色存在，再 UpdateRole
func TestAssignRolePermissions_PreservesPlatforms(t *testing.T) {
	var updateReq *permissionv1.UpdateRoleRequest
	var getRoleCalled bool
	mockRpc := &mockPermRpc{
		getRoleFn: func(ctx context.Context, in *permissionv1.GetRoleRequest) (*permissionv1.GetRoleResponse, error) {
			getRoleCalled = true
			assert.Equal(t, int64(7), in.Id)
			return &permissionv1.GetRoleResponse{
				Base: responsex.NewBaseResp(),
				Role: &permissionv1.Role{Id: 7, Code: "owner", Name: "业主", Platforms: []string{"mobile"}},
			}, nil
		},
		updateRoleFn: func(ctx context.Context, in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
			updateReq = in
			return &permissionv1.UpdateRoleResponse{Base: responsex.NewBaseResp()}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewAssignRolePermissionsLogic(context.Background(), svcCtx)

	err := logic.AssignRolePermissions(&types.AssignRolePermissionsReq{PermissionIds: []int64{10, 20}}, 7)
	assert.NoError(t, err)
	assert.True(t, getRoleCalled, "应先调用 GetRole 读取当前 platforms")
	require.NotNil(t, updateReq, "UpdateRole 应被调用")
	assert.Equal(t, []string{"mobile"}, updateReq.Platforms, "UpdateRole 应显式保留现有 platforms（防 D3 无条件覆盖清空端限制）")
	assert.Equal(t, []int64{10, 20}, updateReq.PermissionIds)
}

// TestAssignRolePermissions_GetRoleBaseAbort — GetRole Base 60001 → abort 返回 Go error，UpdateRole 不被调用（REQ-PLAT-8）
// RED: 当前实现不调用 GetRole → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestAssignRolePermissions_GetRoleBaseAbort(t *testing.T) {
	var updateCalled bool
	mockRpc := &mockPermRpc{
		getRoleFn: func(ctx context.Context, in *permissionv1.GetRoleRequest) (*permissionv1.GetRoleResponse, error) {
			return &permissionv1.GetRoleResponse{
				Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
			}, nil // Role=nil
		},
		updateRoleFn: func(ctx context.Context, in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
			updateCalled = true
			return &permissionv1.UpdateRoleResponse{Base: responsex.NewBaseResp()}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewAssignRolePermissionsLogic(context.Background(), svcCtx)

	err := logic.AssignRolePermissions(&types.AssignRolePermissionsReq{PermissionIds: []int64{10}}, 999)
	assert.Error(t, err, "GetRole Base=60001 应 abort 返回 Go error")
	assert.False(t, updateCalled, "GetRole 失败时不得调用 UpdateRole（否则带空 platforms 清空端限制）")
}

// TestAssignRolePermissions_UpdateRoleBaseError — UpdateRole 响应 Base 非零 → 转 Go error（REQ-UPDATE-3）
func TestAssignRolePermissions_UpdateRoleBaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		getRoleFn: func(ctx context.Context, in *permissionv1.GetRoleRequest) (*permissionv1.GetRoleResponse, error) {
			return &permissionv1.GetRoleResponse{
				Base: responsex.NewBaseResp(),
				Role: &permissionv1.Role{Id: 7, Code: "owner", Name: "业主", Platforms: []string{"pc"}},
			}, nil
		},
		updateRoleFn: func(ctx context.Context, in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
			return &permissionv1.UpdateRoleResponse{
				Base: responsex.NewBaseRespWithError(60004, "系统角色状态不可修改"),
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewAssignRolePermissionsLogic(context.Background(), svcCtx)

	err := logic.AssignRolePermissions(&types.AssignRolePermissionsReq{PermissionIds: []int64{10}}, 7)
	assert.Error(t, err, "UpdateRole Base 非零应转换为 Go error")
}
