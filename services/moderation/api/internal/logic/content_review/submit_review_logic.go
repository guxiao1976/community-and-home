package content_review

import (
	"context"

	"github.com/guxiao/community-and-home/services/moderation/api/internal/svc"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitReviewLogic {
	return &SubmitReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitReviewLogic) SubmitReview(req *types.SubmitReviewReq) (resp *types.SubmitReviewResp, err error) {
	return &types.SubmitReviewResp{Message: "not implemented"}, nil
}
