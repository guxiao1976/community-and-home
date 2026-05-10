package approval

import (
	"context"
	"database/sql"
	"sort"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPendingItemsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPendingItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPendingItemsLogic {
	return &GetPendingItemsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPendingItemsLogic) GetPendingItems(req *types.GetPendingItemsReq) (resp *types.GetPendingItemsResp, err error) {
	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	var submissionType *int64
	if req.SubmissionType != nil {
		st := int64(*req.SubmissionType)
		submissionType = &st
	}

	var items []types.ApprovalPendingItem

	switch req.EntityType {
	case "residential_area":
		areas, _ := l.svcCtx.MdResidentialAreaModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, page, pageSize)
		for _, a := range areas {
			items = append(items, types.ApprovalPendingItem{
				Id:             a.Id,
				EntityType:     "residential_area",
				SubmissionType: int32(a.SubmissionType.Int64),
				Name:           a.Name,
				ChangeSummary:  buildChangeSummary(a.ChangeSnapshot.String, a.SubmissionType.Int64),
				SubmitterId:    a.SubmitterId,
				SubmitTime:     formatTime(a.SubmitTime),
				SubmissionStatus: int32(a.SubmissionStatus),
			})
		}
		count, _ := l.svcCtx.MdResidentialAreaModel.CountBySubmissionStatus(l.ctx, 1)
		return &types.GetPendingItemsResp{List: items, Total: count}, nil

	case "administrative_division":
		divs, _ := l.svcCtx.MdAdministrativeDivisionModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, page, pageSize)
		for _, d := range divs {
			items = append(items, types.ApprovalPendingItem{
				Id:             d.Id,
				EntityType:     "administrative_division",
				SubmissionType: int32(d.SubmissionType.Int64),
				Name:           d.Name,
				ChangeSummary:  buildChangeSummary(d.ChangeSnapshot.String, d.SubmissionType.Int64),
				SubmitterId:    d.SubmitterId.Int64,
				SubmitTime:     formatTime(d.SubmitTime),
				SubmissionStatus: int32(d.SubmissionStatus),
			})
		}
		count, _ := l.svcCtx.MdAdministrativeDivisionModel.CountBySubmissionStatus(l.ctx, 1)
		return &types.GetPendingItemsResp{List: items, Total: count}, nil

	case "configuration":
		configs, _ := l.svcCtx.MdConfigurationModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, page, pageSize)
		for _, c := range configs {
			items = append(items, types.ApprovalPendingItem{
				Id:             c.Id,
				EntityType:     "configuration",
				SubmissionType: int32(c.SubmissionType.Int64),
				Name:           c.Module + "." + c.ConfigKey,
				ChangeSummary:  buildChangeSummary(c.ChangeSnapshot.String, c.SubmissionType.Int64),
				SubmitterId:    c.SubmitterId.Int64,
				SubmitTime:     formatTime(c.SubmitTime),
				SubmissionStatus: int32(c.SubmissionStatus),
			})
		}
		count, _ := l.svcCtx.MdConfigurationModel.CountBySubmissionStatus(l.ctx, 1)
		return &types.GetPendingItemsResp{List: items, Total: count}, nil

	case "sensitive_word":
		words, _ := l.svcCtx.MdSensitiveWordModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, page, pageSize)
		for _, w := range words {
			items = append(items, types.ApprovalPendingItem{
				Id:             w.Id,
				EntityType:     "sensitive_word",
				SubmissionType: int32(w.SubmissionType.Int64),
				Name:           w.Word,
				ChangeSummary:  buildChangeSummary(w.ChangeSnapshot.String, w.SubmissionType.Int64),
				SubmitterId:    w.SubmitterId.Int64,
				SubmitTime:     formatTime(w.SubmitTime),
				SubmissionStatus: int32(w.SubmissionStatus),
			})
		}
		count, _ := l.svcCtx.MdSensitiveWordModel.CountBySubmissionStatus(l.ctx, 1)
		return &types.GetPendingItemsResp{List: items, Total: count}, nil

	default:
		return l.getMergedPendingItems(submissionType, page, pageSize)
	}
}

func (l *GetPendingItemsLogic) getMergedPendingItems(submissionType *int64, page, pageSize int64) (*types.GetPendingItemsResp, error) {
	type submitEntry struct {
		timeStr string
		index   int
	}

	largePageSize := int64(10000)
	raList, _ := l.svcCtx.MdResidentialAreaModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, 1, largePageSize)
	divList, _ := l.svcCtx.MdAdministrativeDivisionModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, 1, largePageSize)
	cfgList, _ := l.svcCtx.MdConfigurationModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, 1, largePageSize)
	swList, _ := l.svcCtx.MdSensitiveWordModel.FindPendingBySubmissionStatus(l.ctx, 1, submissionType, 1, largePageSize)

	var allItems []types.ApprovalPendingItem
	for _, a := range raList {
		allItems = append(allItems, types.ApprovalPendingItem{
			Id:               a.Id,
			EntityType:       "residential_area",
			SubmissionType:   int32(a.SubmissionType.Int64),
			Name:             a.Name,
			ChangeSummary:    buildChangeSummary(a.ChangeSnapshot.String, a.SubmissionType.Int64),
			SubmitterId:      a.SubmitterId,
			SubmitTime:       formatTime(a.SubmitTime),
			SubmissionStatus: int32(a.SubmissionStatus),
		})
	}
	for _, d := range divList {
		allItems = append(allItems, types.ApprovalPendingItem{
			Id:               d.Id,
			EntityType:       "administrative_division",
			SubmissionType:   int32(d.SubmissionType.Int64),
			Name:             d.Name,
			ChangeSummary:    buildChangeSummary(d.ChangeSnapshot.String, d.SubmissionType.Int64),
			SubmitterId:      d.SubmitterId.Int64,
			SubmitTime:       formatTime(d.SubmitTime),
			SubmissionStatus: int32(d.SubmissionStatus),
		})
	}
	for _, c := range cfgList {
		allItems = append(allItems, types.ApprovalPendingItem{
			Id:               c.Id,
			EntityType:       "configuration",
			SubmissionType:   int32(c.SubmissionType.Int64),
			Name:             c.Module + "." + c.ConfigKey,
			ChangeSummary:    buildChangeSummary(c.ChangeSnapshot.String, c.SubmissionType.Int64),
			SubmitterId:      c.SubmitterId.Int64,
			SubmitTime:       formatTime(c.SubmitTime),
			SubmissionStatus: int32(c.SubmissionStatus),
		})
	}
	for _, w := range swList {
		allItems = append(allItems, types.ApprovalPendingItem{
			Id:               w.Id,
			EntityType:       "sensitive_word",
			SubmissionType:   int32(w.SubmissionType.Int64),
			Name:             w.Word,
			ChangeSummary:    buildChangeSummary(w.ChangeSnapshot.String, w.SubmissionType.Int64),
			SubmitterId:      w.SubmitterId.Int64,
			SubmitTime:       formatTime(w.SubmitTime),
			SubmissionStatus: int32(w.SubmissionStatus),
		})
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].SubmitTime > allItems[j].SubmitTime
	})

	total := int64(len(allItems))
	offset := (page - 1) * pageSize
	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}

	return &types.GetPendingItemsResp{List: allItems[offset:end], Total: total}, nil
}

func buildChangeSummary(snapshot string, submissionType int64) string {
	if submissionType == 1 {
		return "新增记录"
	}
	if submissionType == 3 {
		return "申请删除"
	}
	if snapshot != "" {
		return "修改了部分字段"
	}
	return ""
}

func formatTime(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}
