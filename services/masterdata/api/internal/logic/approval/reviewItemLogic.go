package approval

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewItemLogic {
	return &ReviewItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewItemLogic) ReviewItem(req *types.ReviewItemReq) (resp *types.ReviewItemResp, err error) {
	if req.Action != "approve" && req.Action != "reject" {
		return nil, errorx.NewDefaultError("无效的操作，只能 approve 或 reject")
	}
	if req.Action == "reject" && req.ReviewNotes == "" {
		return nil, errorx.NewDefaultError("拒绝时必须填写审核备注")
	}

	reviewerId := l.getReviewerId()

	switch req.EntityType {
	case "residential_area":
		return l.reviewResidentialArea(req.Id, req.Action, req.ReviewNotes, reviewerId)
	case "administrative_division":
		return l.reviewDivision(req.Id, req.Action, req.ReviewNotes, reviewerId)
	case "configuration":
		return l.reviewConfiguration(req.Id, req.Action, req.ReviewNotes, reviewerId)
	case "sensitive_word":
		return l.reviewSensitiveWord(req.Id, req.Action, req.ReviewNotes, reviewerId)
	default:
		return nil, errorx.NewDefaultError("不支持的实体类型: " + req.EntityType)
	}
}

func (l *ReviewItemLogic) getReviewerId() int64 {
	if uid := l.ctx.Value("userId"); uid != nil {
		return uid.(int64)
	}
	return 0
}

func (l *ReviewItemLogic) reviewResidentialArea(id int64, action, notes string, reviewerId int64) (*types.ReviewItemResp, error) {
	area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("住宅小区不存在")
		}
		return nil, errorx.NewDefaultError("查询住宅小区失败")
	}
	if area.SubmissionStatus != 1 {
		return nil, errorx.NewDefaultError("该记录不在待审核状态")
	}

	now := time.Now()
	area.ReviewerId = sql.NullInt64{Int64: reviewerId, Valid: true}
	area.ReviewTime = sql.NullTime{Time: now, Valid: true}
	area.ReviewNotes = sql.NullString{String: notes, Valid: notes != ""}

	if action == "approve" {
		area.SubmissionStatus = 2
		if area.SubmissionType.Int64 == 3 {
			area.DeleteTime = sql.NullTime{Time: now, Valid: true}
		}
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "residential_area", id, reviewerId, 1, "")
	} else {
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "residential_area", id, reviewerId, 2, notes)
		area.SubmissionStatus = 3
		if area.SubmissionType.Int64 == 2 && area.ChangeSnapshot.Valid && area.ChangeSnapshot.String != "" {
			l.restoreResidentialAreaFromSnapshot(area)
		}
		if area.SubmissionType.Int64 == 3 {
			area.SubmissionStatus = 0
			area.SubmissionType = sql.NullInt64{Valid: false}
		}
	}

	if err := l.svcCtx.MdResidentialAreaModel.Update(l.ctx, area); err != nil {
		return nil, errorx.NewDefaultError("审核操作失败: " + err.Error())
	}
	return &types.ReviewItemResp{Success: true}, nil
}

func (l *ReviewItemLogic) restoreResidentialAreaFromSnapshot(area *model.MdResidentialArea) {
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(area.ChangeSnapshot.String), &snapshot); err != nil {
		return
	}
	if v, ok := snapshot["name"].(string); ok {
		area.Name = v
	}
	if v, ok := snapshot["address"].(string); ok {
		area.Address = v
	}
	if v, ok := snapshot["community_type"].(float64); ok {
		area.CommunityType = int64(v)
	}
	if v, ok := snapshot["code"].(string); ok {
		area.Code = sql.NullString{String: v, Valid: true}
	}
	if v, ok := snapshot["county_id"].(float64); ok {
		area.CountyId = sql.NullInt64{Int64: int64(v), Valid: true}
	}
	if v, ok := snapshot["street_id"].(float64); ok {
		area.StreetId = sql.NullInt64{Int64: int64(v), Valid: true}
	}
	if v, ok := snapshot["community_div_id"].(float64); ok {
		area.CommunityDivId = sql.NullInt64{Int64: int64(v), Valid: true}
	}
	if v, ok := snapshot["area"].(float64); ok {
		area.Area = sql.NullFloat64{Float64: v, Valid: true}
	}
	if v, ok := snapshot["population"].(float64); ok {
		area.Population = sql.NullInt64{Int64: int64(v), Valid: true}
	}
	area.ChangeSnapshot = sql.NullString{Valid: false}
}

func (l *ReviewItemLogic) reviewDivision(id int64, action, notes string, reviewerId int64) (*types.ReviewItemResp, error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("行政区划不存在")
		}
		return nil, errorx.NewDefaultError("查询行政区划失败")
	}
	if div.SubmissionStatus != 1 {
		return nil, errorx.NewDefaultError("该记录不在待审核状态")
	}

	now := time.Now()
	div.ReviewerId = sql.NullInt64{Int64: reviewerId, Valid: true}
	div.ReviewTime = sql.NullTime{Time: now, Valid: true}
	div.ReviewNotes = sql.NullString{String: notes, Valid: notes != ""}

	if action == "approve" {
		div.SubmissionStatus = 2
		if div.SubmissionType.Valid && div.SubmissionType.Int64 == 3 {
			div.DeleteTime = sql.NullTime{Time: now, Valid: true}
		}
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "administrative_division", id, reviewerId, 1, "")
	} else {
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "administrative_division", id, reviewerId, 2, notes)
		if div.SubmissionType.Valid && div.SubmissionType.Int64 == 1 {
			if err := l.svcCtx.MdAdministrativeDivisionModel.Delete(l.ctx, id); err != nil {
				return nil, errorx.NewDefaultError("删除被拒绝的新增数据失败: " + err.Error())
			}
			return &types.ReviewItemResp{Success: true}, nil
		}
		if div.SubmissionType.Valid && div.SubmissionType.Int64 == 3 {
			div.SubmissionStatus = 2
			div.SubmissionType = sql.NullInt64{Valid: false}
		} else {
			div.SubmissionStatus = 3
			if div.SubmissionType.Valid && div.SubmissionType.Int64 == 2 && div.ChangeSnapshot.Valid && div.ChangeSnapshot.String != "" {
				l.restoreDivisionFromSnapshot(div)
			}
		}
	}

	if err := l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, div); err != nil {
		return nil, errorx.NewDefaultError("审核操作失败: " + err.Error())
	}
	return &types.ReviewItemResp{Success: true}, nil
}

func (l *ReviewItemLogic) restoreDivisionFromSnapshot(div *model.MdAdministrativeDivision) {
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(div.ChangeSnapshot.String), &snapshot); err != nil {
		return
	}
	if v, ok := snapshot["name"].(string); ok {
		div.Name = v
	}
	if v, ok := snapshot["code"].(string); ok {
		div.Code = v
	}
	if v, ok := snapshot["sort_order"].(float64); ok {
		div.SortOrder = int64(v)
	}
	if v, ok := snapshot["status"].(float64); ok {
		div.Status = int64(v)
	}
	if v, ok := snapshot["parent_id"].(float64); ok {
		div.ParentId = sql.NullInt64{Int64: int64(v), Valid: true}
	}
	div.ChangeSnapshot = sql.NullString{Valid: false}
}

func (l *ReviewItemLogic) reviewConfiguration(id int64, action, notes string, reviewerId int64) (*types.ReviewItemResp, error) {
	cfg, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("系统配置不存在")
		}
		return nil, errorx.NewDefaultError("查询系统配置失败")
	}
	if cfg.SubmissionStatus != 1 {
		return nil, errorx.NewDefaultError("该记录不在待审核状态")
	}

	now := time.Now()
	cfg.ReviewerId = sql.NullInt64{Int64: reviewerId, Valid: true}
	cfg.ReviewTime = sql.NullTime{Time: now, Valid: true}
	cfg.ReviewNotes = sql.NullString{String: notes, Valid: notes != ""}

	if action == "approve" {
		cfg.SubmissionStatus = 2
		if cfg.SubmissionType.Int64 == 3 {
			cfg.DeleteTime = sql.NullTime{Time: now, Valid: true}
		}
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "configuration", id, reviewerId, 1, "")
	} else {
		cfg.SubmissionStatus = 3
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "configuration", id, reviewerId, 2, notes)
		if cfg.SubmissionType.Int64 == 2 && cfg.ChangeSnapshot.Valid && cfg.ChangeSnapshot.String != "" {
			l.restoreConfigurationFromSnapshot(cfg)
		}
		if cfg.SubmissionType.Int64 == 3 {
			cfg.SubmissionStatus = 0
			cfg.SubmissionType = sql.NullInt64{Valid: false}
		}
	}

	if err := l.svcCtx.MdConfigurationModel.Update(l.ctx, cfg); err != nil {
		return nil, errorx.NewDefaultError("审核操作失败: " + err.Error())
	}
	return &types.ReviewItemResp{Success: true}, nil
}

func (l *ReviewItemLogic) restoreConfigurationFromSnapshot(cfg *model.MdConfiguration) {
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(cfg.ChangeSnapshot.String), &snapshot); err != nil {
		return
	}
	if v, ok := snapshot["value"].(string); ok {
		cfg.ConfigValue = v
	}
	if v, ok := snapshot["is_public"].(float64); ok {
		cfg.IsPublic = int64(v)
	}
	if v, ok := snapshot["description"].(string); ok {
		cfg.Description = sql.NullString{String: v, Valid: true}
	}
	cfg.ChangeSnapshot = sql.NullString{Valid: false}
}

func (l *ReviewItemLogic) reviewSensitiveWord(id int64, action, notes string, reviewerId int64) (*types.ReviewItemResp, error) {
	word, err := l.svcCtx.MdSensitiveWordModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("敏感词不存在")
		}
		return nil, errorx.NewDefaultError("查询敏感词失败")
	}
	if word.SubmissionStatus != 1 {
		return nil, errorx.NewDefaultError("该记录不在待审核状态")
	}

	now := time.Now()
	word.ReviewerId = sql.NullInt64{Int64: reviewerId, Valid: true}
	word.ReviewTime = sql.NullTime{Time: now, Valid: true}
	word.ReviewNotes = sql.NullString{String: notes, Valid: notes != ""}

	if action == "approve" {
		word.SubmissionStatus = 2
		if word.SubmissionType.Int64 == 3 {
			word.DeleteTime = sql.NullTime{Time: now, Valid: true}
		}
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "sensitive_word", id, reviewerId, 1, "")
	} else {
		word.SubmissionStatus = 3
		_ = l.svcCtx.SubmissionRecordModel.UpdateResultByEntity(l.ctx, "sensitive_word", id, reviewerId, 2, notes)
		if word.SubmissionType.Int64 == 2 && word.ChangeSnapshot.Valid && word.ChangeSnapshot.String != "" {
			l.restoreSensitiveWordFromSnapshot(word)
		}
		if word.SubmissionType.Int64 == 3 {
			word.SubmissionStatus = 0
			word.SubmissionType = sql.NullInt64{Valid: false}
		}
	}

	if err := l.svcCtx.MdSensitiveWordModel.Update(l.ctx, word); err != nil {
		return nil, errorx.NewDefaultError("审核操作失败: " + err.Error())
	}
	return &types.ReviewItemResp{Success: true}, nil
}

func (l *ReviewItemLogic) restoreSensitiveWordFromSnapshot(word *model.MdSensitiveWord) {
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(word.ChangeSnapshot.String), &snapshot); err != nil {
		return
	}
	if v, ok := snapshot["word"].(string); ok {
		word.Word = v
	}
	if v, ok := snapshot["category"].(string); ok {
		word.Category = v
	}
	if v, ok := snapshot["severity"].(float64); ok {
		word.Severity = int64(v)
	}
	if v, ok := snapshot["action"].(float64); ok {
		word.Action = int64(v)
	}
	if v, ok := snapshot["status"].(float64); ok {
		word.Status = int64(v)
	}
	word.ChangeSnapshot = sql.NullString{Valid: false}
}
