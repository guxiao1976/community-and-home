// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyVerificationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get my verification requests
func NewGetMyVerificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyVerificationsLogic {
	return &GetMyVerificationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyVerificationsLogic) GetMyVerifications(req *types.GetMyVerificationsReq) (resp *types.GetMyVerificationsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
