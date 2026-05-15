// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"context"

	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Review verification request (admin)
func NewReviewVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewVerificationLogic {
	return &ReviewVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewVerificationLogic) ReviewVerification(req *types.ReviewVerificationReq) (resp *types.ReviewVerificationResp, err error) {
	// todo: add your logic here and delete this line

	return
}
