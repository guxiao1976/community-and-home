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

type GetSensitiveWordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List sensitive words with filters
func NewGetSensitiveWordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSensitiveWordsLogic {
	return &GetSensitiveWordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSensitiveWordsLogic) GetSensitiveWords(req *types.GetSensitiveWordsReq) (resp *types.GetSensitiveWordsResp, err error) {
	// Calculate pagination
	limit := int(req.PageSize)
	offset := (int(req.Page) - 1) * limit

	// Convert filters to model types
	var severity *int64
	var status *int64

	if req.Severity != nil {
		sev := int64(*req.Severity)
		severity = &sev
	}
	if req.Status != nil {
		st := int64(*req.Status)
		status = &st
	}

	// Query with filters
	words, total, err := l.svcCtx.MdSensitiveWordModel.FindWithFilters(l.ctx, req.Category, severity, status, limit, offset)
	if err != nil {
		return nil, errorx.NewDefaultError("查询敏感词列表失败")
	}

	// Convert to response types
	var list []types.SensitiveWord
	for _, w := range words {
		list = append(list, types.SensitiveWord{
			Id:          w.Id,
			Word:        w.Word,
			Category:    w.Category,
			Severity:    int32(w.Severity),
			Action:      int32(w.Action),
			Status:      int32(w.Status),
			CreatedBy:   w.CreatedBy,
			CreatedTime: w.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: w.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetSensitiveWordsResp{
		List:  list,
		Total: total,
	}, nil
}
