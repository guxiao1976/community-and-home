// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/aimodel"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建模型配置
func NewCreateModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateModelLogic {
	return &CreateModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateModelLogic) CreateModel(req *types.CreateModelRequest) (resp *types.ModelResponse, err error) {
	// 调用 RPC 服务创建模型配置
	rpcReq := &aimodel.CreateModelConfigReq{
		Name:                   req.Name,
		DisplayName:            req.DisplayName,
		Provider:               req.Provider,
		Type:                   req.Type,
		Endpoint:               req.Endpoint,
		MaxTokens:              req.MaxTokens,
		SupportedFeatures:      req.SupportedFeatures,
		CostPer_1KInputTokens:  req.CostPer1KInputTokens,
		CostPer_1KOutputTokens: req.CostPer1KOutputTokens,
		Description:            req.Description,
	}

	rpcResp, err := l.svcCtx.AiModelRpc.CreateModelConfig(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("CreateModelConfig failed: %v", err)
		return nil, err
	}

	// 构造响应
	resp = &types.ModelResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: types.ModelInfo{
			Id:                    rpcResp.Id,
			Name:                  rpcResp.Name,
			DisplayName:           rpcResp.DisplayName,
			Provider:              rpcResp.Provider,
			Type:                  rpcResp.Type,
			Endpoint:              rpcResp.Endpoint,
			MaxTokens:             rpcResp.MaxTokens,
			SupportedFeatures:     rpcResp.SupportedFeatures,
			CostPer1KInputTokens:  rpcResp.CostPer_1KInputTokens,
			CostPer1KOutputTokens: rpcResp.CostPer_1KOutputTokens,
			Status:                rpcResp.Status,
			Description:           rpcResp.Description,
		},
	}

	return resp, nil
}
