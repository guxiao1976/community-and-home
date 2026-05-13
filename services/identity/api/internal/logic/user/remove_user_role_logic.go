package user

import (
	"context"
	"fmt"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveUserRoleLogic {
	return &RemoveUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveUserRoleLogic) RemoveUserRole(req *types.RemoveUserRoleReq) (resp *types.RemoveUserRoleResp, err error) {
	_, err = l.svcCtx.AuthUserRoleModel.FindOneByUserIdRoleId(l.ctx, req.Id, req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("user %d does not have role %d assigned", req.Id, req.RoleId)
	}

	role, err := l.svcCtx.AuthRoleModel.FindOne(l.ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	if role.IsSystem == 1 {
		return nil, fmt.Errorf("cannot remove system role %s", role.Code)
	}

	err = l.svcCtx.AuthUserRoleModel.DeleteByUserIdAndRoleId(l.ctx, req.Id, req.RoleId)
	if err != nil {
		logx.Errorf("Failed to remove role %d from user %d: %v", req.RoleId, req.Id, err)
		return nil, err
	}

	return &types.RemoveUserRoleResp{Success: true}, nil
}
