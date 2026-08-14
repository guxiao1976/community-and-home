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

// TestRevokeUserRole_BaseError — RevokeRole Base 业务错误 → Go error，不再静默成功（REQ-UPDATE-3 静默类）
// RED: API 层忽略 grpcResp.Base → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestRevokeUserRole_BaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		revokeRoleFn: func(ctx context.Context, in *permissionv1.RevokeRoleRequest) (*permissionv1.RevokeRoleResponse, error) {
			return &permissionv1.RevokeRoleResponse{
				Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewRevokeUserRoleLogic(context.Background(), svcCtx)

	err := logic.RevokeUserRole(&types.RevokeUserRoleReq{UserId: 1001, RoleId: 999})
	assert.Error(t, err, "Base.Code=60001 应转换为 Go error，不再静默成功")
}
