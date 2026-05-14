package logic

import (
	"context"
	"fmt"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetModelConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetModelConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetModelConfigLogic {
	return &GetModelConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetModelConfigLogic) GetModelConfig(in *pb.GetModelConfigReq) (*pb.ModelConfigResp, error) {
	config, err := l.svcCtx.ModelManager.GetModelConfig(l.ctx, in.Id)
	if err != nil {
		l.Errorf("failed to get model config: %v", err)
		return nil, err
	}

	// Check if deleted
	if config.DeleteTime.Valid {
		return nil, fmt.Errorf("model config not found")
	}

	displayName := config.ModelName
	if config.DisplayName.Valid {
		displayName = config.DisplayName.String
	}

	description := ""
	if config.Description.Valid {
		description = config.Description.String
	}

	endpoint := ""
	if config.Endpoint.Valid {
		endpoint = config.Endpoint.String
	}

	capabilities := ""
	if config.Capabilities.Valid {
		capabilities = config.Capabilities.String
	}

	return &pb.ModelConfigResp{
		Id:                      config.Id,
		Name:                    config.ModelName,
		DisplayName:             displayName,
		Provider:                config.Provider,
		Type:                    config.ModelType,
		Endpoint:                endpoint,
		MaxTokens:               0, // Not in database model
		SupportedFeatures:       capabilities,
		CostPer_1KInputTokens:   config.CostPer1KInputTokens,
		CostPer_1KOutputTokens:  config.CostPer1KOutputTokens,
		Status:                  config.Status,
		Description:             description,
	}, nil
}
