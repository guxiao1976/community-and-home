// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Remove role from user
func NewRemoveUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveUserRoleLogic {
	return &RemoveUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveUserRoleLogic) RemoveUserRole(req *types.RemoveUserRoleReq) (resp *types.RemoveUserRoleResp, err error) {
	// todo: add your logic here and delete this line

	return
}
