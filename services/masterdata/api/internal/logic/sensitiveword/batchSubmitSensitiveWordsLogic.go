// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sensitiveword

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSubmitSensitiveWordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Batch submit sensitive words
func NewBatchSubmitSensitiveWordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSubmitSensitiveWordsLogic {
	return &BatchSubmitSensitiveWordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchSubmitSensitiveWordsLogic) BatchSubmitSensitiveWords(req *types.BatchSubmitReq) (resp *types.BatchSubmitResp, err error) {
	if len(req.Ids) == 0 {
		return nil, errorx.NewDefaultError("请选择要提交的数据")
	}

	successCount := 0
	for _, id := range req.Ids {
		word, err := l.svcCtx.MdSensitiveWordModel.FindOne(l.ctx, id)
		if err != nil || word.DeleteTime.Valid {
			continue
		}
		if word.SubmissionStatus != 0 && word.SubmissionStatus != 3 {
			continue
		}
		word.SubmissionStatus = 1
		if err := l.svcCtx.MdSensitiveWordModel.Update(l.ctx, word); err != nil {
			continue
		}
		successCount++
	}

	return &types.BatchSubmitResp{Success: successCount > 0}, nil
}
