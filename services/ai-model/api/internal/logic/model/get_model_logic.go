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

type GetModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取模型详情
func NewGetModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetModelLogic {
	return &GetModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetModelLogic) GetModel(req *types.GetModelRequest) (resp *types.ModelResponse, err error) {
	// 调用 RPC 服务获取模型配置详情
	rpcResp, err := l.svcCtx.AiModelRpc.GetModelConfig(l.ctx, &pb.GetModelConfigReq{
		Id: req.Id,
	})
	if err != nil {
		return &types.ModelResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

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
