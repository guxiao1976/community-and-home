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

// TestAssignUserRole_BaseError — AssignRole Base 业务错误 → Go error，不再静默成功（REQ-UPDATE-3 静默类）
// RED: API 层忽略 grpcResp.Base → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]]
func TestAssignUserRole_BaseError(t *testing.T) {
	mockRpc := &mockPermRpc{
		assignRoleFn: func(ctx context.Context, in *permissionv1.AssignRoleRequest) (*permissionv1.AssignRoleResponse, error) {
			return &permissionv1.AssignRoleResponse{
				Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewAssignUserRoleLogic(context.Background(), svcCtx)

	err := logic.AssignUserRole(&types.AssignUserRoleReq{UserId: 1001, RoleId: 999, ScopeType: "global"})
	assert.Error(t, err, "Base.Code=60001 应转换为 Go error，不再静默成功")
}
