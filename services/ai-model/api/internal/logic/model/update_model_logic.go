// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新模型配置
func NewUpdateModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateModelLogic {
	return &UpdateModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateModelLogic) UpdateModel(req *types.UpdateModelRequest) (resp *types.ModelResponse, err error) {
	// 调用 RPC 服务更新模型配置
	rpcResp, err := l.svcCtx.AiModelRpc.UpdateModelConfig(l.ctx, &pb.UpdateModelConfigReq{
		Id:                     req.Id,
		DisplayName:            req.DisplayName,
		Endpoint:               req.Endpoint,
		MaxTokens:              req.MaxTokens,
		SupportedFeatures:      req.SupportedFeatures,
		CostPer_1KInputTokens:  req.CostPer1KInputTokens,
		CostPer_1KOutputTokens: req.CostPer1KOutputTokens,
		Status:                 req.Status,
		Description:            req.Description,
	})
	if err != nil {
		l.Errorf("update model config failed: %v", err)
		return nil, err
	}

	// 转换响应
	return &types.ModelResponse{
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
	}, nil
}
