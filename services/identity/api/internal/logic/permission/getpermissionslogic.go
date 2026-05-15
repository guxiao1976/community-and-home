// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package permission

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get permission tree
func NewGetPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionsLogic {
	return &GetPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPermissionsLogic) GetPermissions(req *types.GetPermissionsReq) (resp *types.GetPermissionsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
