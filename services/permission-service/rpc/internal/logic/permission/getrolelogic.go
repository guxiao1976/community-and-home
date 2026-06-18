package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetRole 获取单个角色详情（含权限列表）
func (l *GetRoleLogic) GetRole(in *permissionv1.GetRoleRequest) (*permissionv1.GetRoleResponse, error) {
	role, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &permissionv1.GetRoleResponse{
			Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
		}, nil
	}

	// 查询角色关联的权限 ID
	rps, _ := l.svcCtx.RolePermissionModel.FindByRoleId(l.ctx, in.Id)
	var permIds []int64
	for _, rp := range rps {
		permIds = append(permIds, rp.PermissionId)
	}

	return &permissionv1.GetRoleResponse{
		Base: responsex.NewBaseResp(),
		Role: roleToPbWithPermissions(l.ctx, role, permIds, l.svcCtx.PermissionModel),
	}, nil
}
