package deleteditems

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeletedItemsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeletedItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeletedItemsLogic {
	return &GetDeletedItemsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeletedItemsLogic) GetDeletedItems(req *types.GetDeletedItemsReq) (resp *types.GetDeletedItemsResp, err error) {
	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	et := req.EntityType

	var allItems []types.DeletedItem
	var totalCount int64

	if et == "" || et == "residential_area" {
		areas, total, e := l.svcCtx.MdResidentialAreaModel.FindDeleted(l.ctx, page, pageSize)
		if e == nil {
			totalCount += total
			for _, a := range areas {
				code, dt := "", ""
				if a.Code.Valid { code = a.Code.String }
				if a.DeleteTime.Valid { dt = a.DeleteTime.Time.Format("2006-01-02 15:04:05") }
				allItems = append(allItems, types.DeletedItem{Id: a.Id, EntityType: "residential_area", Name: a.Name, Code: code, DeleteTime: dt})
			}
		}
	}

	if et == "" || et == "administrative_division" {
		divs, total, e := l.svcCtx.MdAdministrativeDivisionModel.FindDeleted(l.ctx, page, pageSize)
		if e == nil {
			totalCount += total
			for _, d := range divs {
				dt := ""
				if d.DeleteTime.Valid { dt = d.DeleteTime.Time.Format("2006-01-02 15:04:05") }
				allItems = append(allItems, types.DeletedItem{Id: d.Id, EntityType: "administrative_division", Name: d.Name, Code: d.Code, DeleteTime: dt})
			}
		}
	}

	if et == "" || et == "configuration" {
		cfgs, total, e := l.svcCtx.MdConfigurationModel.FindDeleted(l.ctx, page, pageSize)
		if e == nil {
			totalCount += total
			for _, c := range cfgs {
				dt := ""
				if c.DeleteTime.Valid { dt = c.DeleteTime.Time.Format("2006-01-02 15:04:05") }
				allItems = append(allItems, types.DeletedItem{Id: c.Id, EntityType: "configuration", Name: c.ConfigKey, Code: c.Module, DeleteTime: dt})
			}
		}
	}

	if et == "" || et == "sensitive_word" {
		words, total, e := l.svcCtx.MdSensitiveWordModel.FindDeleted(l.ctx, page, pageSize)
		if e == nil {
			totalCount += total
			for _, w := range words {
				dt := ""
				if w.DeleteTime.Valid { dt = w.DeleteTime.Time.Format("2006-01-02 15:04:05") }
				allItems = append(allItems, types.DeletedItem{Id: w.Id, EntityType: "sensitive_word", Name: w.Word, Code: w.Category, DeleteTime: dt})
			}
		}
	}

	return &types.GetDeletedItemsResp{List: allItems, Total: totalCount}, nil
}
