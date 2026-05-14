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

type GetModelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取可用模型列表
func NewGetModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetModelsLogic {
	return &GetModelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetModelsLogic) GetModels(req *types.GetModelsRequest) (resp *types.GetModelsResponse, err error) {
	// 调用 RPC 服务获取可用模型列表
	rpcResp, err := l.svcCtx.AiModelRpc.GetAvailableModels(l.ctx, &pb.GetModelsRequest{
		Provider:    req.Provider,
		OnlyHealthy: req.Status == 1, // Status=1 表示只返回启用的模型
	})
	if err != nil {
		return &types.GetModelsResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换 RPC 响应为 API 响应
	models := make([]types.ModelInfo, 0, len(rpcResp.Models))
	for _, m := range rpcResp.Models {
		// 如果指定了 Type 过滤，则跳过不匹配的模型
		if req.Type != "" && m.Type != req.Type {
			continue
		}

		// 将 capabilities (repeated string) 转换为 SupportedFeatures (string)
		supportedFeatures := ""
		if len(m.Capabilities) > 0 {
			supportedFeatures = m.Capabilities[0]
			for i := 1; i < len(m.Capabilities); i++ {
				supportedFeatures += "," + m.Capabilities[i]
			}
		}

		// 将 enabled (bool) 转换为 Status (int64)
		var status int64 = 0
		if m.Enabled {
			status = 1
		}

		models = append(models, types.ModelInfo{
			Id:                    m.Id,
			Name:                  m.Name,
			DisplayName:           m.DisplayName,
			Provider:              m.Provider,
			Type:                  m.Type,
			SupportedFeatures:     supportedFeatures,
			CostPer1KInputTokens:  m.CostPer_1KInputTokens,
			CostPer1KOutputTokens: m.CostPer_1KOutputTokens,
			Status:                status,
		})
	}

	return &types.GetModelsResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.ModelsData{
			Models: models,
			Total:  int32(len(models)),
		},
	}, nil
}
