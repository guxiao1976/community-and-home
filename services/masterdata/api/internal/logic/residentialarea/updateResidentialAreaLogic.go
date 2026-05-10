package residentialarea

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateResidentialAreaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateResidentialAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateResidentialAreaLogic {
	return &UpdateResidentialAreaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateResidentialAreaLogic) UpdateResidentialArea(req *types.UpdateResidentialAreaReq) (resp *types.UpdateResidentialAreaResp, err error) {
	area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("住宅小区不存在")
	}
	if area.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("住宅小区已删除")
	}
	if area.SubmissionStatus == 1 || area.SubmissionStatus == 4 {
		return nil, errorx.NewDefaultError("仅待提交或已拒绝状态的小区可以编辑")
	}

	// Capture change snapshot of current values before modification
	snapshot := map[string]interface{}{
		"name": area.Name,
		"address": area.Address,
		"community_type": area.CommunityType,
	}
	if area.Code.Valid {
		snapshot["code"] = area.Code.String
	}
	if area.CountyId.Valid {
		snapshot["county_id"] = area.CountyId.Int64
	}
	if area.StreetId.Valid {
		snapshot["street_id"] = area.StreetId.Int64
	}
	if area.CommunityDivId.Valid {
		snapshot["community_div_id"] = area.CommunityDivId.Int64
	}
	if area.Area.Valid {
		snapshot["area"] = area.Area.Float64
	}
	if area.Population.Valid {
		snapshot["population"] = area.Population.Int64
	}
	snapshotJson, _ := json.Marshal(snapshot)
	area.ChangeSnapshot = sql.NullString{String: string(snapshotJson), Valid: true}
	area.SubmissionType = sql.NullInt64{Int64: 2, Valid: true}

	if req.Code != "" && req.Code != area.Code.String {
		existing, err := l.svcCtx.MdResidentialAreaModel.FindByCode(l.ctx, req.Code)
		if err == nil && existing != nil {
			return nil, errorx.NewDefaultError("小区编码已存在")
		}
		area.Code = sql.NullString{String: req.Code, Valid: true}
	}

	if req.Name != "" {
		area.Name = req.Name
	}
	if req.Address != "" {
		area.Address = req.Address
	}
	if req.Area > 0 {
		area.Area = sql.NullFloat64{Float64: req.Area, Valid: true}
	}
	if req.Population > 0 {
		area.Population = sql.NullInt64{Int64: int64(req.Population), Valid: true}
	}
	if req.CommunityType > 0 {
		area.CommunityType = int64(req.CommunityType)
	}
	if req.StreetId != nil {
		area.StreetId = toNullInt64(req.StreetId)
	}
	if req.CommunityDivId != nil {
		area.CommunityDivId = toNullInt64(req.CommunityDivId)
	}

	area.SubmissionStatus = 0
	area.SubmitterId = 0
	area.SubmitTime = sql.NullTime{}
	if err := l.svcCtx.MdResidentialAreaModel.Update(l.ctx, area); err != nil {
		return nil, errorx.NewDefaultError("更新住宅小区失败: " + err.Error())
	}

	return &types.UpdateResidentialAreaResp{Success: true}, nil
}
