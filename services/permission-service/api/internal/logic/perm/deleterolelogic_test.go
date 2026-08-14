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

// TestDeleteRole_Base60004 — DeleteRole Base 60004（系统角色不可删除）→ Go error，不再静默成功（REQ-UPDATE-3 静默类）
// RED: API 层忽略 grpcResp.Base → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestDeleteRole_Base60004(t *testing.T) {
	mockRpc := &mockPermRpc{
		deleteRoleFn: func(ctx context.Context, in *permissionv1.DeleteRoleRequest) (*permissionv1.DeleteRoleResponse, error) {
			return &permissionv1.DeleteRoleResponse{
				Base: responsex.NewBaseRespWithError(60004, "系统角色不可删除"),
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewDeleteRoleLogic(context.Background(), svcCtx)

	err := logic.DeleteRole(&types.DeleteRoleReq{Id: 1})
	assert.Error(t, err, "Base.Code=60004 应转换为 Go error，不再静默成功")
}
