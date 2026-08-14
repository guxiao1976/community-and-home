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

// TestUpdateRole_PlatformsAlwaysPassed — req.Platforms=[] 恒透传（D3，空列表也设置，与「未传」区分）
// RED: API 层未构造 grpcReq.Platforms → gotReq.Platforms 为 nil → assert.NotNil FAIL
func TestUpdateRole_PlatformsAlwaysPassed(t *testing.T) {
	var gotReq *permissionv1.UpdateRoleRequest
	mockRpc := &mockPermRpc{
		updateRoleFn: func(ctx context.Context, in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
			gotReq = in
			return &permissionv1.UpdateRoleResponse{
				Base: responsex.NewBaseResp(),
				Role: &permissionv1.Role{Id: in.Id, Code: "owner", Name: "业主", Platforms: in.Platforms},
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	req := &types.UpdateRoleReq{Id: 1, Platforms: []string{}}
	resp, err := logic.UpdateRole(req)
	assert.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, gotReq, "gRPC 请求应被调用")
	assert.NotNil(t, gotReq.Platforms, "空 platforms 也应恒透传（空列表，非 nil）")
	assert.Equal(t, 0, len(gotReq.Platforms))
}

// TestUpdateRole_Base60004NoPanic — Base 60004 + Role=nil → 返回 Go error，不 panic（REQ-UPDATE-1）
// RED: API 层未检查 grpcResp.Base → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestUpdateRole_Base60004NoPanic(t *testing.T) {
	mockRpc := &mockPermRpc{
		updateRoleFn: func(ctx context.Context, in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
			return &permissionv1.UpdateRoleResponse{
				Base: responsex.NewBaseRespWithError(60004, "系统角色状态不可修改"),
			}, nil // Role=nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	name := "改名"
	resp, err := logic.UpdateRole(&types.UpdateRoleReq{Id: 1, Name: &name})
	assert.Error(t, err, "Base.Code=60004 应转换为 Go error（不再 500）")
	assert.Nil(t, resp)
}

// TestUpdateRole_Success — 正常路径：Base=0 + Role → 返回 UpdateRoleResp 转换成功（避免 ToError 过度拦截）
func TestUpdateRole_Success(t *testing.T) {
	mockRpc := &mockPermRpc{
		updateRoleFn: func(ctx context.Context, in *permissionv1.UpdateRoleRequest) (*permissionv1.UpdateRoleResponse, error) {
			return &permissionv1.UpdateRoleResponse{
				Base: responsex.NewBaseResp(),
				Role: &permissionv1.Role{Id: in.Id, Code: "owner", Name: "业主", Platforms: []string{"mobile"}},
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	name := "改名"
	resp, err := logic.UpdateRole(&types.UpdateRoleReq{Id: 1, Name: &name, Platforms: []string{"mobile"}})
	assert.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Role.Id)
	assert.Equal(t, []string{"mobile"}, resp.Role.Platforms)
}
