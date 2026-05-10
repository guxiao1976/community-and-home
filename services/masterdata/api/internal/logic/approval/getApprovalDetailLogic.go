package approval

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetApprovalDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetApprovalDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApprovalDetailLogic {
	return &GetApprovalDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetApprovalDetailLogic) GetApprovalDetail(req *types.GetApprovalDetailReq) (resp *types.ApprovalDetailResp, err error) {
	switch req.EntityType {
	case "residential_area":
		return l.getResidentialAreaDetail(req.Id)
	case "administrative_division":
		return l.getDivisionDetail(req.Id)
	case "configuration":
		return l.getConfigurationDetail(req.Id)
	case "sensitive_word":
		return l.getSensitiveWordDetail(req.Id)
	default:
		return nil, errorx.NewDefaultError("不支持的实体类型: " + req.EntityType)
	}
}

func (l *GetApprovalDetailLogic) getResidentialAreaDetail(id int64) (*types.ApprovalDetailResp, error) {
	area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("住宅小区不存在")
		}
		return nil, errorx.NewDefaultError("查询住宅小区失败")
	}

	currentData := map[string]interface{}{
		"name":           area.Name,
		"address":        area.Address,
		"community_type": area.CommunityType,
	}
	if area.Code.Valid {
		currentData["code"] = area.Code.String
	}
	if area.CountyId.Valid {
		currentData["county_id"] = area.CountyId.Int64
	}
	if area.StreetId.Valid {
		currentData["street_id"] = area.StreetId.Int64
	}
	if area.CommunityDivId.Valid {
		currentData["community_div_id"] = area.CommunityDivId.Int64
	}
	if area.Area.Valid {
		currentData["area"] = area.Area.Float64
	}
	if area.Population.Valid {
		currentData["population"] = area.Population.Int64
	}

	currentJson, _ := json.Marshal(currentData)

	return &types.ApprovalDetailResp{
		Id:             id,
		EntityType:     "residential_area",
		SubmissionType: int32(area.SubmissionType.Int64),
		CurrentData:    string(currentJson),
		SnapshotData:   area.ChangeSnapshot.String,
		SubmitterId:    area.SubmitterId,
		SubmitTime:     formatNullTime(area.SubmitTime),
	}, nil
}

func (l *GetApprovalDetailLogic) getDivisionDetail(id int64) (*types.ApprovalDetailResp, error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("行政区划不存在")
		}
		return nil, errorx.NewDefaultError("查询行政区划失败")
	}

	currentData := map[string]interface{}{
		"name":       div.Name,
		"code":       div.Code,
		"sort_order": div.SortOrder,
		"status":     div.Status,
		"level":      div.Level,
	}
	if div.ParentId.Valid {
		currentData["parent_id"] = div.ParentId.Int64
	}

	currentJson, _ := json.Marshal(currentData)

	resp := &types.ApprovalDetailResp{
		Id:             id,
		EntityType:     "administrative_division",
		SubmissionType: int32(div.SubmissionType.Int64),
		CurrentData:    string(currentJson),
		SnapshotData:   div.ChangeSnapshot.String,
		SubmitTime:     formatNullTime(div.SubmitTime),
	}
	if div.SubmitterId.Valid {
		resp.SubmitterId = div.SubmitterId.Int64
	}
	if div.ReviewerId.Valid {
		reviewerId := div.ReviewerId.Int64
		resp.ReviewerId = &reviewerId
	}
	if div.ReviewNotes.Valid {
		resp.ReviewNotes = div.ReviewNotes.String
	}
	resp.ReviewTime = formatNullTime(div.ReviewTime)

	return resp, nil
}

func (l *GetApprovalDetailLogic) getConfigurationDetail(id int64) (*types.ApprovalDetailResp, error) {
	cfg, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("系统配置不存在")
		}
		return nil, errorx.NewDefaultError("查询系统配置失败")
	}

	currentData := map[string]interface{}{
		"module":      cfg.Module,
		"config_key":  cfg.ConfigKey,
		"value":       cfg.ConfigValue,
		"value_type":  cfg.ValueType,
		"is_public":   cfg.IsPublic,
	}
	if cfg.Description.Valid {
		currentData["description"] = cfg.Description.String
	}

	currentJson, _ := json.Marshal(currentData)

	resp := &types.ApprovalDetailResp{
		Id:             id,
		EntityType:     "configuration",
		SubmissionType: int32(cfg.SubmissionType.Int64),
		CurrentData:    string(currentJson),
		SnapshotData:   cfg.ChangeSnapshot.String,
		SubmitTime:     formatNullTime(cfg.SubmitTime),
	}
	if cfg.SubmitterId.Valid {
		resp.SubmitterId = cfg.SubmitterId.Int64
	}
	if cfg.ReviewerId.Valid {
		reviewerId := cfg.ReviewerId.Int64
		resp.ReviewerId = &reviewerId
	}
	if cfg.ReviewNotes.Valid {
		resp.ReviewNotes = cfg.ReviewNotes.String
	}
	resp.ReviewTime = formatNullTime(cfg.ReviewTime)

	return resp, nil
}

func (l *GetApprovalDetailLogic) getSensitiveWordDetail(id int64) (*types.ApprovalDetailResp, error) {
	word, err := l.svcCtx.MdSensitiveWordModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("敏感词不存在")
		}
		return nil, errorx.NewDefaultError("查询敏感词失败")
	}

	currentData := map[string]interface{}{
		"word":     word.Word,
		"category": word.Category,
		"severity": word.Severity,
		"action":   word.Action,
		"status":   word.Status,
	}

	currentJson, _ := json.Marshal(currentData)

	resp := &types.ApprovalDetailResp{
		Id:             id,
		EntityType:     "sensitive_word",
		SubmissionType: int32(word.SubmissionType.Int64),
		CurrentData:    string(currentJson),
		SnapshotData:   word.ChangeSnapshot.String,
		SubmitTime:     formatNullTime(word.SubmitTime),
	}
	if word.SubmitterId.Valid {
		resp.SubmitterId = word.SubmitterId.Int64
	}
	if word.ReviewerId.Valid {
		reviewerId := word.ReviewerId.Int64
		resp.ReviewerId = &reviewerId
	}
	if word.ReviewNotes.Valid {
		resp.ReviewNotes = word.ReviewNotes.String
	}
	resp.ReviewTime = formatNullTime(word.ReviewTime)

	return resp, nil
}

func formatNullTime(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}
