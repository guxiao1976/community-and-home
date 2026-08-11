package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRolePermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRolePermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolePermissionsLogic {
	return &AssignRolePermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AssignRolePermissions 更新角色的权限列表（替换模式：先删后插）
func (l *AssignRolePermissionsLogic) AssignRolePermissions(req *types.AssignRolePermissionsReq, roleId int64) error {
	_, err := l.svcCtx.PermissionRpc.UpdateRole(l.ctx, &permissionv1.UpdateRoleRequest{
		Id:            roleId,
		PermissionIds: req.PermissionIds,
	})
	return err
}
