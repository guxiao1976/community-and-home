package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/pb"

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
	// 从数据库查询模型配置
	modelConfig, err := l.svcCtx.ModelConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("find model config failed: %v", err)
		return nil, err
	}

	// 处理可空字段
	endpoint := ""
	if modelConfig.Endpoint.Valid {
		endpoint = modelConfig.Endpoint.String
	}

	description := ""
	if modelConfig.Description.Valid {
		description = modelConfig.Description.String
	}

	displayName := ""
	if modelConfig.DisplayName.Valid {
		displayName = modelConfig.DisplayName.String
	}

	capabilities := ""
	if modelConfig.Capabilities.Valid {
		capabilities = modelConfig.Capabilities.String
	}

	return &pb.ModelConfigResp{
		Id:                     modelConfig.Id,
		Name:                   modelConfig.ModelName,
		DisplayName:            displayName,
		Provider:               modelConfig.Provider,
		Type:                   modelConfig.ModelType,
		Endpoint:               endpoint,
		MaxTokens:              0, // 数据库中没有此字段，返回0
		SupportedFeatures:      capabilities,
		CostPer_1KInputTokens:  modelConfig.CostPer1KInputTokens,
		CostPer_1KOutputTokens: modelConfig.CostPer1KOutputTokens,
		Status:                 modelConfig.Status,
		Description:            description,
	}, nil
}
