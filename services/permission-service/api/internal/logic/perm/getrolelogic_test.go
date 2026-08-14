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

// TestGetRole_BaseError — GetRole Base 60001 + Role=nil → 返回 Go error，不 deref（REQ-UPDATE-3 panic 类）
// RED: API 层未检查 grpcResp.Base → 返回 nil err + 零值 Role → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestGetRole_BaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		getRoleFn: func(ctx context.Context, in *permissionv1.GetRoleRequest) (*permissionv1.GetRoleResponse, error) {
			return &permissionv1.GetRoleResponse{
				Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
			}, nil // Role=nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewGetRoleLogic(context.Background(), svcCtx)

	resp, err := logic.GetRole(&types.GetRoleReq{Id: 999})
	assert.Error(t, err, "Base.Code=60001 应转换为 Go error")
	assert.Nil(t, resp)
}
