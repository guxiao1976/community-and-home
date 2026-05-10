// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List or tree view of administrative divisions
func NewGetDivisionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionsLogic {
	return &GetDivisionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDivisionsLogic) GetDivisions(req *types.GetDivisionsReq) (resp *types.GetDivisionsResp, err error) {
	var divisions []*model.MdAdministrativeDivision
	var total int64

	if req.ParentId != nil {
		var level *int64
		var submissionStatus *int64
		if req.Level != nil {
			lv := int64(*req.Level)
			level = &lv
		}
		if req.SubmissionStatus != nil {
			ss := int64(*req.SubmissionStatus)
			submissionStatus = &ss
		}
		divisions, err = l.svcCtx.MdAdministrativeDivisionModel.FindChildrenWithFilter(l.ctx, *req.ParentId, level, submissionStatus)
		if err == nil {
			total, _ = l.svcCtx.MdAdministrativeDivisionModel.CountChildren(l.ctx, *req.ParentId, level, submissionStatus)
		}
	} else if req.SubmissionStatus != nil || req.Level != nil || req.MinLevel != nil {
		var level *int64
		var minLevel *int64
		var submissionStatus *int64
		if req.Level != nil {
			lv := int64(*req.Level)
			level = &lv
		}
		if req.MinLevel != nil {
			ml := int64(*req.MinLevel)
			minLevel = &ml
		}
		if req.SubmissionStatus != nil {
			ss := int64(*req.SubmissionStatus)
			submissionStatus = &ss
		}
		divisions, err = l.svcCtx.MdAdministrativeDivisionModel.FindAllWithFilter(l.ctx, level, minLevel, submissionStatus, int64(req.Page), int64(req.PageSize))
		if err == nil {
			total, _ = l.svcCtx.MdAdministrativeDivisionModel.CountAllWithFilter(l.ctx, level, minLevel, submissionStatus)
		}
	} else if req.Level != nil {
		divisions, err = l.svcCtx.MdAdministrativeDivisionModel.FindByLevel(l.ctx, int64(*req.Level))
	} else {
		divisions, err = l.svcCtx.MdAdministrativeDivisionModel.FindAll(l.ctx, int64(req.Page), int64(req.PageSize))
	}

	if err != nil {
		return nil, err
	}

	list := make([]types.Division, 0, len(divisions))
	for _, div := range divisions {
		var parentId *int64
		if div.ParentId.Valid {
			parentId = &div.ParentId.Int64
		}

		list = append(list, types.Division{
			Id:               div.Id,
			ParentId:         parentId,
			Code:             div.Code,
			Name:             div.Name,
			Level:            int32(div.Level),
			Path:             div.Path,
			SortOrder:        int32(div.SortOrder),
			Status:           int32(div.Status),
			SubmissionStatus: int32(div.SubmissionStatus),
			SubmissionType: func() *int64 { if div.SubmissionType.Valid { v := div.SubmissionType.Int64; return &v }; return nil }(),
			CreatedBy:        div.CreatedBy,
			CreatedTime:      div.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime:      div.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetDivisionsResp{
		List:  list,
		Total: total,
	}, nil
}
