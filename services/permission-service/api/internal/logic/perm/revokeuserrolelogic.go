package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
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

	_, err := l.svcCtx.PermissionRpc.RevokeRole(l.ctx, grpcReq)
	return err
}
