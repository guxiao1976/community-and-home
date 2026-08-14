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

// TestGetUserPermissions_BaseError — GetUserPermissions Base 60001 → Go error（REQ-UPDATE-3 panic 类）
// RED: API 层未检查 grpcResp.Base → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestGetUserPermissions_BaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		getUserPermissionsFn: func(ctx context.Context, in *permissionv1.GetUserPermissionsRequest) (*permissionv1.GetUserPermissionsResponse, error) {
			return &permissionv1.GetUserPermissionsResponse{
				Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewGetUserPermissionsLogic(context.Background(), svcCtx)

	resp, err := logic.GetUserPermissions(&types.GetUserPermissionsReq{UserId: 1001})
	assert.Error(t, err, "Base.Code=60001 应转换为 Go error")
	assert.Nil(t, resp)
}
