package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/pb"

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
	l.Logger.Infof("UpdateModelConfig called with: id=%d", in.Id)

	// 查询现有模型配置
	existingConfig, err := l.svcCtx.ModelConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("Failed to find model config: %v", err)
		return nil, err
	}

	// 准备更新数据 - 保留现有值，只更新传入的非空字段
	updatedConfig := *existingConfig
	updatedConfig.UpdatedTime = time.Now()

	// 更新字段（如果提供了新值）
	if in.Name != "" {
		updatedConfig.ModelName = in.Name
	}
	if in.DisplayName != "" {
		updatedConfig.DisplayName = sql.NullString{String: in.DisplayName, Valid: true}
	}
	if in.Provider != "" {
		updatedConfig.Provider = in.Provider
	}
	if in.Type != "" {
		updatedConfig.ModelType = in.Type
	}
	if in.Endpoint != "" {
		updatedConfig.Endpoint = sql.NullString{String: in.Endpoint, Valid: true}
	}
	if in.Description != "" {
		updatedConfig.Description = sql.NullString{String: in.Description, Valid: true}
	}

	// 处理 supported_features - 转换为 JSON 数组
	if in.SupportedFeatures != "" {
		features := strings.Split(in.SupportedFeatures, ",")
		for i := range features {
			features[i] = strings.TrimSpace(features[i])
		}
		jsonBytes, err := json.Marshal(features)
		if err != nil {
			l.Logger.Errorf("Failed to marshal capabilities: %v", err)
			return nil, err
		}
		updatedConfig.Capabilities = sql.NullString{String: string(jsonBytes), Valid: true}
	}

	// 更新数值字段
	if in.MaxTokens > 0 {
		// Note: max_tokens 不在数据库表中，这里跳过
	}
	if in.CostPer_1KInputTokens >= 0 {
		updatedConfig.CostPer1KInputTokens = in.CostPer_1KInputTokens
	}
	if in.CostPer_1KOutputTokens >= 0 {
		updatedConfig.CostPer1KOutputTokens = in.CostPer_1KOutputTokens
	}
	if in.Status >= 0 {
		updatedConfig.Status = in.Status
	}

	// 执行更新
	err = l.svcCtx.ModelConfigModel.Update(l.ctx, &updatedConfig)
	if err != nil {
		l.Logger.Errorf("Failed to update model config: %v", err)
		return nil, err
	}

	l.Logger.Infof("Model config updated successfully: id=%d", in.Id)

	// 处理返回的可空字段
	endpoint := ""
	if updatedConfig.Endpoint.Valid {
		endpoint = updatedConfig.Endpoint.String
	}

	description := ""
	if updatedConfig.Description.Valid {
		description = updatedConfig.Description.String
	}

	displayName := ""
	if updatedConfig.DisplayName.Valid {
		displayName = updatedConfig.DisplayName.String
	}

	capabilities := ""
	if updatedConfig.Capabilities.Valid {
		capabilities = updatedConfig.Capabilities.String
	}

	// 返回更新后的模型配置
	return &pb.ModelConfigResp{
		Id:                     updatedConfig.Id,
		Name:                   updatedConfig.ModelName,
		DisplayName:            displayName,
		Provider:               updatedConfig.Provider,
		Type:                   updatedConfig.ModelType,
		Endpoint:               endpoint,
		MaxTokens:              in.MaxTokens,
		SupportedFeatures:      capabilities,
		CostPer_1KInputTokens:  updatedConfig.CostPer1KInputTokens,
		CostPer_1KOutputTokens: updatedConfig.CostPer1KOutputTokens,
		Status:                 updatedConfig.Status,
		Description:            description,
	}, nil
}
