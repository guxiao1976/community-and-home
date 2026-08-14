package perm

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/stretchr/testify/assert"
)

// TestGetRolePermissions_BaseError — GetRole Base 60001 + Role=nil → Go error，不 deref Role.Permissions（REQ-UPDATE-3）
// RED: API 层未检查 grpcResp.Base → 直接 range grpcResp.Role.Permissions → nil panic 或 nil err
// SEE: [[rpc-callback-must-check-response-base]]
func TestGetRolePermissions_BaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		getRoleFn: func(ctx context.Context, in *permissionv1.GetRoleRequest) (*permissionv1.GetRoleResponse, error) {
			return &permissionv1.GetRoleResponse{
				Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
			}, nil // Role=nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewGetRolePermissionsLogic(context.Background(), svcCtx)

	resp, err := logic.GetRolePermissions(&types.GetRoleReq{Id: 999})
	assert.Error(t, err, "Base.Code=60001 应转换为 Go error，禁止 deref Role.Permissions")
	assert.Nil(t, resp)
}
