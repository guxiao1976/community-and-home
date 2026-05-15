package logic

import (
	"context"
	"encoding/json"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAvailableModelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAvailableModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailableModelsLogic {
	return &GetAvailableModelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取可用模型列表
func (l *GetAvailableModelsLogic) GetAvailableModels(in *pb.GetModelsRequest) (*pb.GetModelsResponse, error) {
	l.Logger.Infof("GetAvailableModels called with provider=%s, only_healthy=%v", in.Provider, in.OnlyHealthy)

	// 构建状态过滤条件
	var status int64 = 0
	if in.OnlyHealthy {
		status = 1 // 只返回启用的模型
	}

	// 查询列表（不分页，返回所有）
	configs, err := l.svcCtx.ModelConfigModel.FindList(l.ctx, in.Provider, status, 1, 1000)
	if err != nil {
		l.Logger.Errorf("Failed to query models: %v", err)
		return nil, err
	}

	l.Logger.Infof("Found %d models", len(configs))

	// 转换为响应格式
	var models []*pb.ModelInfo
	for _, config := range configs {
		// 解析 capabilities JSON 数组
		var capabilities []string
		if config.Capabilities.Valid && config.Capabilities.String != "" {
			if err := json.Unmarshal([]byte(config.Capabilities.String), &capabilities); err != nil {
				l.Logger.Errorf("Failed to parse capabilities for model %s: %v", config.ModelName, err)
				capabilities = []string{}
			}
		}

		models = append(models, &pb.ModelInfo{
			Id:                      config.Id,
			Name:                    config.ModelName,
			Type:                    config.ModelType,
			Provider:                config.Provider,
			DisplayName:             config.DisplayName.String,
			Capabilities:            capabilities,
			CostPer_1KInputTokens:   config.CostPer1KInputTokens,
			CostPer_1KOutputTokens:  config.CostPer1KOutputTokens,
			HealthStatus:            config.HealthStatus,
			Priority:                int32(config.Priority),
			Enabled:                 config.Status == 1,
		})
	}

	return &pb.GetModelsResponse{
		Models: models,
	}, nil
}
