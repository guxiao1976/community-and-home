package residentialarea

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResidentialAreasLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetResidentialAreasLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResidentialAreasLogic {
	return &GetResidentialAreasLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetResidentialAreasLogic) GetResidentialAreas(req *types.GetResidentialAreasReq) (resp *types.GetResidentialAreasResp, err error) {
	var areas []*model.MdResidentialArea
	var total int64

	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	countyId := req.CountyId
	streetId := req.StreetId
	communityDivId := req.CommunityDivId
	communityType := req.CommunityType
	keyword := req.Keyword
	submissionStatus := req.SubmissionStatus
	excludeStatus := (*int32)(nil)
	// 默认排除待删除(4)的记录
	if submissionStatus == nil {
		s := int32(4)
		excludeStatus = &s
	}

	// When city_id is set but no finer division filter, resolve to county IDs
	var countyIds []int64
	if req.CityId != nil && countyId == nil && streetId == nil && communityDivId == nil {
		divisions, err := l.svcCtx.MdAdministrativeDivisionModel.FindChildren(l.ctx, *req.CityId)
		if err != nil {
			return nil, err
		}
		for _, d := range divisions {
			countyIds = append(countyIds, d.Id)
		}
		if len(countyIds) == 0 {
			return &types.GetResidentialAreasResp{List: []types.ResidentialArea{}, Total: 0}, nil
		}
	}

	// Build variadic exclude arg
	var excludeArg []int32
	if excludeStatus != nil {
		excludeArg = []int32{*excludeStatus}
	}

	// Build keyword pointer for Count calls
	var keywordPtr *string
	if keyword != "" {
		keywordPtr = &keyword
	}

	if keyword != "" {
		areas, err = l.svcCtx.MdResidentialAreaModel.SearchByName(l.ctx, keyword, countyId, streetId, communityDivId, countyIds, submissionStatus, communityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, submissionStatus, countyIds, keywordPtr, communityType, excludeArg...)
	} else if communityDivId != nil {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCommunityDivId(l.ctx, *communityDivId, submissionStatus, communityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, submissionStatus, countyIds, nil, communityType, excludeArg...)
	} else if streetId != nil {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByStreetId(l.ctx, *streetId, submissionStatus, communityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, submissionStatus, countyIds, nil, communityType, excludeArg...)
	} else if countyId != nil {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCountyId(l.ctx, *countyId, submissionStatus, communityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, countyId, streetId, communityDivId, submissionStatus, countyIds, nil, communityType, excludeArg...)
	} else if len(countyIds) > 0 {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCountyIds(l.ctx, countyIds, submissionStatus, communityType, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, nil, submissionStatus, countyIds, nil, communityType, excludeArg...)
	} else if submissionStatus != nil {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindBySubmissionStatus(l.ctx, int64(*submissionStatus), page, pageSize)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, nil, submissionStatus, nil, nil, communityType)
	} else {
		areas, err = l.svcCtx.MdResidentialAreaModel.FindAll(l.ctx, nil, page, pageSize, excludeArg...)
		total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, nil, nil, nil, nil, communityType, excludeArg...)
	}

	if err != nil {
		return nil, err
	}

	list := make([]types.ResidentialArea, 0, len(areas))
	for _, a := range areas {
		list = append(list, modelToResidentialArea(a))
	}

	return &types.GetResidentialAreasResp{
		List:  list,
		Total: total,
	}, nil
}

func modelToResidentialArea(a *model.MdResidentialArea) types.ResidentialArea {
	ra := types.ResidentialArea{
		Id:               a.Id,
		Name:             a.Name,
		Address:          a.Address,
		CommunityType:    int32(a.CommunityType),
		SubmissionStatus: int32(a.SubmissionStatus),
		SubmitterId:      a.SubmitterId,
		CreatedTime:      a.CreatedTime.Format("2006-01-02 15:04:05"),
		UpdatedTime:      a.UpdatedTime.Format("2006-01-02 15:04:05"),
	}
	if a.CountyId.Valid {
		ra.CountyId = &a.CountyId.Int64
	}
	if a.CityId.Valid {
		ra.CityId = &a.CityId.Int64
	}
	if a.StreetId.Valid {
		ra.StreetId = &a.StreetId.Int64
	}
	if a.CommunityDivId.Valid {
		CommunityDivId := a.CommunityDivId.Int64
		ra.CommunityDivId = &CommunityDivId
	}
	if a.Code.Valid {
		ra.Code = a.Code.String
	}
	if a.Area.Valid {
		ra.Area = a.Area.Float64
	}
	if a.Population.Valid {
		ra.Population = int32(a.Population.Int64)
	}
	if a.SubmitTime.Valid {
		ra.SubmitTime = a.SubmitTime.Time.Format("2006-01-02 15:04:05")
	}
	if a.ReviewerId.Valid {
		reviewerId := a.ReviewerId.Int64
		ra.ReviewerId = &reviewerId
	}
	if a.ReviewTime.Valid {
		ra.ReviewTime = a.ReviewTime.Time.Format("2006-01-02 15:04:05")
	}
	if a.ReviewNotes.Valid {
		ra.ReviewNotes = a.ReviewNotes.String
	}
	return ra
}
