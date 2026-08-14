package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserRoleLogic {
	return &AssignUserRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AssignUserRole 管理员为用户分配角色（含数据范围 scope）
func (l *AssignUserRoleLogic) AssignUserRole(req *types.AssignUserRoleReq) error {
	grpcResp, err := l.svcCtx.PermissionRpc.AssignRole(l.ctx, &permissionv1.AssignRoleRequest{
		UserId:    req.UserId,
		RoleId:    req.RoleId,
		ScopeType: req.ScopeType,
		ScopeId:   req.ScopeId,
	})
	if err != nil {
		return err
	}

	// Base 检查：业务错误写入 grpcResp.Base，需转为 Go error，不再静默成功
	// SEE: [[rpc-callback-must-check-response-base]]
	return responsex.ToError(grpcResp.Base)
}
