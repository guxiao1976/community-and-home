// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user permissions
func NewGetUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionsLogic {
	return &GetUserPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserPermissionsLogic) GetUserPermissions(req *types.GetUserPermissionsReq) (resp *types.GetUserPermissionsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
