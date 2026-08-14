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

// TestGetUserRoles_BaseError — GetUserRoles Base 60001 → Go error，不 deref Roles（REQ-UPDATE-3 panic 类）
// RED: API 层未检查 grpcResp.Base → 返回 nil err + 空 Roles → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestGetUserRoles_BaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		getUserRolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest) (*permissionv1.GetUserRolesResponse, error) {
			return &permissionv1.GetUserRolesResponse{
				Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
			}, nil // Roles=nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	resp, err := logic.GetUserRoles(&types.GetUserRolesReq{UserId: 1001})
	assert.Error(t, err, "Base.Code=60001 应转换为 Go error")
	assert.Nil(t, resp)
}
