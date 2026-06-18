package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// DeleteRole 删除角色
func (l *DeleteRoleLogic) DeleteRole(req *types.DeleteRoleReq) error {
	grpcReq := &permissionv1.DeleteRoleRequest{
		Id: req.Id,
	}

	_, err := l.svcCtx.PermissionRpc.DeleteRole(l.ctx, grpcReq)
	if err != nil {
		l.Errorf("DeleteRole gRPC call failed: %v", err)
		return err
	}

	return nil
}
