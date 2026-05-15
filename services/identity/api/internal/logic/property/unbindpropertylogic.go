// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbindPropertyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Unbind property from user
func NewUnbindPropertyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindPropertyLogic {
	return &UnbindPropertyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnbindPropertyLogic) UnbindProperty(req *types.UnbindPropertyReq) (resp *types.UnbindPropertyResp, err error) {
	// todo: add your logic here and delete this line

	return
}
