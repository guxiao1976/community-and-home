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

// TestCreateRole_PlatformsPassThrough — req.Platforms → grpcReq.Platforms 透传（REQ-PLAT-2）
// RED: API 层未构造 grpcReq.Platforms → gotReq.Platforms 为空 → assert.Equal FAIL
func TestCreateRole_PlatformsPassThrough(t *testing.T) {
	var gotReq *permissionv1.CreateRoleRequest
	mockRpc := &mockPermRpc{
		createRoleFn: func(ctx context.Context, in *permissionv1.CreateRoleRequest) (*permissionv1.CreateRoleResponse, error) {
			gotReq = in
			return &permissionv1.CreateRoleResponse{
				Base: responsex.NewBaseResp(),
				Role: &permissionv1.Role{Id: 5, Code: in.Code, Name: in.Name, Platforms: in.Platforms},
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewCreateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.CreateRole(&types.CreateRoleReq{
		Code:      "owner",
		Name:      "业主",
		Platforms: []string{"pc", "mobile"},
	})
	assert.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, gotReq, "gRPC 请求应被调用")
	assert.Equal(t, []string{"pc", "mobile"}, gotReq.Platforms, "platforms 应透传到 gRPC 请求")
	assert.Equal(t, []string{"pc", "mobile"}, resp.Role.Platforms, "响应 Role.platforms 应回显")
}

// TestCreateRole_BaseErrorNoPanic — Base 60006 + Role=nil → 返回 Go error，不 panic（REQ-UPDATE-3）
// RED: API 层未检查 grpcResp.Base → 返回 nil err + 零值 Role → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]] — 业务错误不可被静默吞掉
func TestCreateRole_BaseErrorNoPanic(t *testing.T) {
	mockRpc := &mockPermRpc{
		createRoleFn: func(ctx context.Context, in *permissionv1.CreateRoleRequest) (*permissionv1.CreateRoleResponse, error) {
			return &permissionv1.CreateRoleResponse{
				Base: responsex.NewBaseRespWithError(60006, "角色编码已存在"),
			}, nil // Role=nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewCreateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.CreateRole(&types.CreateRoleReq{Code: "owner", Name: "业主"})
	assert.Error(t, err, "Base.Code=60006 应转换为 Go error")
	assert.Nil(t, resp)
}
