package dataquery

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryResidentialAreasLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryResidentialAreasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryResidentialAreasLogic {
	return &QueryResidentialAreasLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryResidentialAreasLogic) QueryResidentialAreas(req *types.QueryResidentialAreasReq) (resp *types.QueryResidentialAreasResp, err error) {
	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 50 {
		pageSize = 50
	}

	// Default: only approved (submission_status=2), exclude deleted (4)
	approved := int32(2)
	excludeDeleted := int32(4)

	countyId := req.CountyId
	streetId := req.StreetId
	communityDivId := req.CommunityDivId
	keyword := req.Keyword

	var countyIds []int64
	excludeArg := []int32{excludeDeleted}

	if req.CityId != nil && countyId == nil && streetId == nil && communityDivId == nil {
		divisions, err := l.svcCtx.MdAdministrativeDivisionModel.FindChildren(l.ctx, *req.CityId)
		if err != nil {
			return nil, err
		}
		for _, d := range divisions {
			countyIds = append(countyIds, d.Id)
		}

		if len(countyIds) == 0 {
			return &types.QueryResidentialAreasResp{List: []types.QueryResidentialAreaItem{}, Total: 0}, nil
		}
	}

	var areas []*model.MdResidentialArea
	var total int64

	var keywordPtr *string
	if keyword != "" {
		keywordPtr = &keyword
	}

	if keyword != "" {
		areas, err = l.svcCtx.MdResidentialAreaModel.SearchByName(l.ctx, keyword, countyId, streetId, communityDivId, countyIds, &approved, req.CommunityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, &approved, countyIds, keywordPtr, req.CommunityType, excludeArg...)
	} else if communityDivId != nil {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCommunityDivId(l.ctx, *communityDivId, &approved, req.CommunityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, &approved, countyIds, nil, req.CommunityType, excludeArg...)
	} else if streetId != nil {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByStreetId(l.ctx, *streetId, &approved, req.CommunityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, &approved, countyIds, nil, req.CommunityType, excludeArg...)
	} else if countyId != nil {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCountyId(l.ctx, *countyId, &approved, req.CommunityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, &approved, countyIds, nil, req.CommunityType, excludeArg...)
	} else if len(countyIds) > 0 {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCountyIds(l.ctx, countyIds, &approved, req.CommunityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, nil, &approved, countyIds, nil, req.CommunityType, excludeArg...)
	} else {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindAll(l.ctx, &approved, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, nil, &approved, nil, nil, req.CommunityType, excludeArg...)
	}
	if err != nil {
		return nil, err
	}

	// Resolve division names with caching
	nameCache := make(map[int64]string)
	getDivName := func(id int64) string {
		if id == 0 {
			return ""
		}
		if name, ok := nameCache[id]; ok {
			return name
		}
		div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, id)
		if err != nil || div == nil {
			return ""
		}
		nameCache[id] = div.Name
		return div.Name
	}

	list := make([]types.QueryResidentialAreaItem, 0, len(areas))
	for _, a := range areas {
		item := types.QueryResidentialAreaItem{
			Id:            a.Id,
			Name:          a.Name,
			Address:       a.Address,
			CommunityType: int32(a.CommunityType),
		}
		if a.Code.Valid {
			item.Code = a.Code.String
		}
		if a.CityId.Valid && a.CityId.Int64 > 0 {
			item.CityId = &a.CityId.Int64
			item.CityName = getDivName(a.CityId.Int64)
		}
		if a.CountyId.Valid && a.CountyId.Int64 > 0 {
			item.CountyId = &a.CountyId.Int64
			item.CountyName = getDivName(a.CountyId.Int64)
		}
		if a.StreetId.Valid && a.StreetId.Int64 > 0 {
			item.StreetId = &a.StreetId.Int64
			item.StreetName = getDivName(a.StreetId.Int64)
		}
		if a.CommunityDivId.Valid && a.CommunityDivId.Int64 > 0 {
			item.CommunityDivId = &a.CommunityDivId.Int64
			item.CommunityName = getDivName(a.CommunityDivId.Int64)
		}
		list = append(list, item)
	}

	return &types.QueryResidentialAreasResp{
		List:  list,
		Total: total,
	}, nil
}
