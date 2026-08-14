package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeUserRoleLogic {
	return &RevokeUserRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// RevokeUserRole 管理员撤销用户的角色
func (l *RevokeUserRoleLogic) RevokeUserRole(req *types.RevokeUserRoleReq) error {
	grpcReq := &permissionv1.RevokeRoleRequest{
		UserId:    req.UserId,
		RoleId:    req.RoleId,
		ScopeType: &req.ScopeType,
		ScopeId:   &req.ScopeId,
	}
	// If scope fields are empty, pass nil to revoke all scopes
	if req.ScopeType == "" {
		grpcReq.ScopeType = nil
		grpcReq.ScopeId = nil
	}

	grpcResp, err := l.svcCtx.PermissionRpc.RevokeRole(l.ctx, grpcReq)
	if err != nil {
		return err
	}

	// Base 检查：业务错误写入 grpcResp.Base，需转为 Go error，不再静默成功
	// SEE: [[rpc-callback-must-check-response-base]]
	return responsex.ToError(grpcResp.Base)
}
