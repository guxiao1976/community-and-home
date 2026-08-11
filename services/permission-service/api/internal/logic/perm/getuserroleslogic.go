package perm

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetUserRoles 查询指定用户的角色列表（管理员操作）
func (l *GetUserRolesLogic) GetUserRoles(req *types.GetUserRolesReq) (*types.GetUserRolesResp, error) {
	grpcResp, err := l.svcCtx.PermissionRpc.GetUserRoles(l.ctx, &permissionv1.GetUserRolesRequest{UserId: req.UserId})
	if err != nil {
		return nil, err
	}

	roles := make([]types.UserRoleInfo, 0, len(grpcResp.Roles))
	for _, r := range grpcResp.Roles {
		roles = append(roles, types.UserRoleInfo{
			Role:      toRoleInfo(r.Role),
			ScopeType: r.ScopeType,
			ScopeId:   r.ScopeId,
		})
	}

	return &types.GetUserRolesResp{Roles: roles}, nil
}
