package perm

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/stretchr/testify/assert"
)

// TestListPermissions_BaseError — ListPermissions 业务错误（99400）→ Go error，不 deref Permissions（REQ-UPDATE-3）
// RED: API 层未检查 grpcResp.Base → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestListPermissions_BaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		listPermissionsFn: func(ctx context.Context, in *permissionv1.ListPermissionsRequest) (*permissionv1.ListPermissionsResponse, error) {
			return &permissionv1.ListPermissionsResponse{
				Base: responsex.NewBaseRespWithError(int32(errx.CodeInvalidParam), "非法参数"),
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewListPermissionsLogic(context.Background(), svcCtx)

	resp, err := logic.ListPermissions(&types.ListPermissionsReq{})
	assert.Error(t, err, "业务错误应转换为 Go error")
	assert.Nil(t, resp)
}
