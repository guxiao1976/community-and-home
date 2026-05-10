package approval

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchReviewItemsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchReviewItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchReviewItemsLogic {
	return &BatchReviewItemsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchReviewItemsLogic) BatchReviewItems(req *types.BatchReviewReq) (resp *types.BatchReviewResp, err error) {
	reviewLogic := NewReviewItemLogic(l.ctx, l.svcCtx)

	var successCount, failCount int64
	for _, id := range req.Ids {
		reviewReq := &types.ReviewItemReq{
			Id:          id,
			EntityType:  req.EntityType,
			Action:      req.Action,
			ReviewNotes: req.ReviewNotes,
		}
		_, reviewErr := reviewLogic.ReviewItem(reviewReq)
		if reviewErr != nil {
			failCount++
			l.Logger.Errorf("batch review failed for id=%d entity=%s: %v", id, req.EntityType, reviewErr)
		} else {
			successCount++
		}
	}

	return &types.BatchReviewResp{
		SuccessCount: successCount,
		FailCount:    failCount,
	}, nil
}
