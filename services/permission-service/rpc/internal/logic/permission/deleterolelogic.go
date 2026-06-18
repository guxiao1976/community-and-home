package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
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
//   校验角色存在 → 系统角色不可删除 → 检查是否被用户引用 → 软删除
func (l *DeleteRoleLogic) DeleteRole(in *permissionv1.DeleteRoleRequest) (*permissionv1.DeleteRoleResponse, error) {
	// 校验角色存在
	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &permissionv1.DeleteRoleResponse{
			Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
		}, nil
	}

	// 系统角色不可删除
	if role.IsSystem == 1 {
		return &permissionv1.DeleteRoleResponse{
			Base: responsex.NewBaseRespWithError(60004, "系统角色不可删除"),
		}, nil
	}

	// 检查是否有用户引用了该角色
	count, err := l.svcCtx.UserRoleModel.CountByRoleId(l.ctx, in.Id)
	if err != nil {
		l.Errorf("DeleteRole: count users failed: %v", err)
		return nil, err
	}
	if count > 0 {
		return &permissionv1.DeleteRoleResponse{
			Base: responsex.NewBaseRespWithError(60004, "角色已被分配，无法删除"),
		}, nil
	}

	// 软删除
	if err := l.svcCtx.RoleModel.SoftDelete(l.ctx, in.Id); err != nil {
		l.Errorf("DeleteRole: soft delete failed: %v", err)
		return nil, err
	}

	// 清理角色-权限关联
	if err := l.svcCtx.RolePermissionModel.DeleteByRoleId(l.ctx, in.Id); err != nil {
		l.Errorf("DeleteRole: clear permissions failed: %v", err)
	}

	l.Infof("DeleteRole success: id=%d", in.Id)

	return &permissionv1.DeleteRoleResponse{Base: responsex.NewBaseResp()}, nil
}
