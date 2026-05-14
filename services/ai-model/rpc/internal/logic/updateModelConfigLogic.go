package logic

import (
	"context"
	"database/sql"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateModelConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateModelConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateModelConfigLogic {
	return &UpdateModelConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateModelConfigLogic) UpdateModelConfig(in *pb.UpdateModelConfigReq) (*pb.ModelConfigResp, error) {
	// 先查询现有配置
	existingConfig, err := l.svcCtx.ModelManager.GetModelConfig(l.ctx, in.Id)
	if err != nil {
		l.Errorf("get model config failed: %v", err)
		return nil, err
	}

	// 更新字段（只更新提供的字段）
	if in.DisplayName != "" {
		existingConfig.DisplayName = sql.NullString{String: in.DisplayName, Valid: true}
	}
	if in.Endpoint != "" {
		existingConfig.Endpoint = sql.NullString{String: in.Endpoint, Valid: true}
	}
	if in.MaxTokens > 0 {
		// MaxTokens 在数据库中没有对应字段，跳过
	}
	if in.SupportedFeatures != "" {
		existingConfig.Capabilities = sql.NullString{String: in.SupportedFeatures, Valid: true}
	}
	if in.CostPer_1KInputTokens > 0 {
		existingConfig.CostPer1KInputTokens = in.CostPer_1KInputTokens
	}
	if in.CostPer_1KOutputTokens > 0 {
		existingConfig.CostPer1KOutputTokens = in.CostPer_1KOutputTokens
	}
	if in.Status > 0 {
		existingConfig.Status = in.Status
	}
	if in.Description != "" {
		existingConfig.Description = sql.NullString{String: in.Description, Valid: true}
	}

	// 更新数据库
	err = l.svcCtx.ModelManager.UpdateModelConfig(l.ctx, existingConfig)
	if err != nil {
		l.Errorf("update model config failed: %v", err)
		return nil, err
	}

	// 返回更新后的配置
	return &pb.ModelConfigResp{
		Id:                     existingConfig.Id,
		Name:                   existingConfig.ModelName,
		DisplayName:            existingConfig.DisplayName.String,
		Provider:               existingConfig.Provider,
		Type:                   existingConfig.ModelType,
		Endpoint:               existingConfig.Endpoint.String,
		MaxTokens:              0, // 数据库中没有此字段
		SupportedFeatures:      existingConfig.Capabilities.String,
		CostPer_1KInputTokens:  existingConfig.CostPer1KInputTokens,
		CostPer_1KOutputTokens: existingConfig.CostPer1KOutputTokens,
		Status:                 existingConfig.Status,
		Description:            existingConfig.Description.String,
	}, nil
}
