// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BindPropertyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Bind property to user
func NewBindPropertyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindPropertyLogic {
	return &BindPropertyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BindPropertyLogic) BindProperty(req *types.BindPropertyReq) (resp *types.BindPropertyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
