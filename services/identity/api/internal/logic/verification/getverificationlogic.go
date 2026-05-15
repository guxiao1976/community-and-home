// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get verification details
func NewGetVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVerificationLogic {
	return &GetVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVerificationLogic) GetVerification(req *types.GetVerificationReq) (resp *types.GetVerificationResp, err error) {
	// todo: add your logic here and delete this line

	return
}
