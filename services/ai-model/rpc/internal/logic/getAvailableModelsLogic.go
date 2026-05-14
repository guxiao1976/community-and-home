package logic

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

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
	configs, err := l.svcCtx.ModelManager.GetAvailableModels(l.ctx, in.Provider)
	if err != nil {
		return nil, err
	}

	models := make([]*pb.ModelInfo, 0, len(configs))
	for _, config := range configs {
		// Filter by health status if requested
		if in.OnlyHealthy && config.HealthStatus != "healthy" {
			continue
		}

		// Parse capabilities from JSON if needed
		var capabilities []string
		// For now, just return empty array - would need to parse JSON in production

		displayName := config.ModelName
		if config.DisplayName.Valid {
			displayName = config.DisplayName.String
		}

		models = append(models, &pb.ModelInfo{
			Id:                      config.Id,
			Name:                    config.ModelName,
			Type:                    config.ModelType,
			Provider:                config.Provider,
			DisplayName:             displayName,
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
