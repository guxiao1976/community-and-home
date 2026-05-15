// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/aimodel"

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
	// 调用 RPC 服务获取模型列表
	rpcResp, err := l.svcCtx.AiModelRpc.GetAvailableModels(l.ctx, &aimodel.GetModelsRequest{
		Provider:    req.Provider,
		OnlyHealthy: req.Status == 1, // status=1 表示只返回启用的模型
	})
	if err != nil {
		l.Logger.Errorf("Failed to get models from RPC: %v", err)
		return nil, err
	}

	// 转换为 API 响应格式
	var models []types.ModelInfo
	for _, m := range rpcResp.Models {
		// 将 capabilities 数组转换为逗号分隔的字符串
		supportedFeatures := ""
		if len(m.Capabilities) > 0 {
			for i, cap := range m.Capabilities {
				if i > 0 {
					supportedFeatures += ","
				}
				supportedFeatures += cap
			}
		}

		models = append(models, types.ModelInfo{
			Id:                    m.Id,
			Name:                  m.Name,
			Type:                  m.Type,
			Provider:              m.Provider,
			DisplayName:           m.DisplayName,
			SupportedFeatures:     supportedFeatures,
			CostPer1KInputTokens:  m.CostPer_1KInputTokens,
			CostPer1KOutputTokens: m.CostPer_1KOutputTokens,
			Status:                func() int64 { if m.Enabled { return 1 }; return 0 }(),
		})
	}

	return &types.GetModelsResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: types.ModelsData{
			Models: models,
			Total:  int32(len(models)),
		},
	}, nil
}
