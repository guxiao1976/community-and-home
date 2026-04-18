package sensitiveword

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSensitiveWordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSensitiveWordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSensitiveWordsLogic {
	return &GetSensitiveWordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSensitiveWordsLogic) GetSensitiveWords(req *types.GetSensitiveWordsReq) (resp *types.GetSensitiveWordsResp, err error) {
	offset := (req.Page - 1) * req.PageSize
	words, total, err := l.svcCtx.MdSensitiveWordModel.FindByCategory(l.ctx, req.Category, int(req.PageSize), int(offset))
	if err != nil {
		return nil, err
	}

	list := make([]types.SensitiveWord, 0, len(words))
	for _, word := range words {
		list = append(list, types.SensitiveWord{
			Id:          word.Id,
			Word:        word.Word,
			Category:    word.Category,
			Severity:    word.Severity,
			Action:      word.Action,
			Status:      word.Status,
			CreatedTime: word.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: word.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetSensitiveWordsResp{
		List:  list,
		Total: total,
	}, nil
}
