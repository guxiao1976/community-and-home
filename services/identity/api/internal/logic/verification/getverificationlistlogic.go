// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVerificationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List all verification requests (admin)
func NewGetVerificationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVerificationListLogic {
	return &GetVerificationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVerificationListLogic) GetVerificationList(req *types.GetVerificationListReq) (resp *types.GetVerificationListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
