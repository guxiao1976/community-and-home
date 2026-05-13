package user

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserRolesLogic) GetUserRoles(req *types.GetUserRolesReq) (resp *types.GetUserRolesResp, err error) {
	userRoles, err := l.svcCtx.AuthUserRoleModel.FindActiveByUserId(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	roles := make([]types.Role, 0, len(userRoles))
	for _, ur := range userRoles {
		roles = append(roles, types.Role{
			Id:          ur.RoleId,
			Name:        ur.RoleName,
			Code:        ur.RoleCode,
			Description: ur.Description,
			IsSystem:    int32(ur.IsSystem),
			Status:      int32(ur.RoleStatus),
		})
	}

	return &types.GetUserRolesResp{Roles: roles}, nil
}
