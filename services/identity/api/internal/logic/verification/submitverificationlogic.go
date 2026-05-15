// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Submit homeowner verification
func NewSubmitVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitVerificationLogic {
	return &SubmitVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitVerificationLogic) SubmitVerification(req *types.SubmitVerificationReq) (resp *types.SubmitVerificationResp, err error) {
	// todo: add your logic here and delete this line

	return
}
