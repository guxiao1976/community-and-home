package user

import (
	"context"
	"fmt"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserRolesLogic {
	return &AssignUserRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignUserRolesLogic) AssignUserRoles(req *types.AssignUserRolesReq) (resp *types.AssignUserRolesResp, err error) {
	if len(req.RoleIds) == 0 {
		return &types.AssignUserRolesResp{Success: true}, nil
	}

	roles, err := l.svcCtx.AuthRoleModel.FindByIds(l.ctx, req.RoleIds)
	if err != nil {
		return nil, err
	}
	if len(roles) != len(req.RoleIds) {
		return nil, fmt.Errorf("one or more role IDs do not exist")
	}
	for _, role := range roles {
		if role.Status != 1 {
			return nil, fmt.Errorf("role %d (%s) is not active", role.Id, role.Code)
		}
	}

	err = l.svcCtx.AuthUserRoleModel.BatchInsertUserRoles(l.ctx, req.Id, req.RoleIds)
	if err != nil {
		logx.Errorf("Failed to assign roles to user %d: %v", req.Id, err)
		return nil, err
	}

	return &types.AssignUserRolesResp{Success: true}, nil
}
