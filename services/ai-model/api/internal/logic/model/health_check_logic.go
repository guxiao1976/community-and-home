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

type HealthCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 健康检查
func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthCheckLogic) HealthCheck() (resp *types.HealthCheckResponse, err error) {
	// 调用 RPC 服务进行健康检查
	rpcResp, err := l.svcCtx.AiModelRpc.HealthCheck(l.ctx, &pb.HealthCheckRequest{})
	if err != nil {
		return &types.HealthCheckResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换 RPC 响应为 API 响应
	modelHealths := make(map[string]types.ModelHealth)
	for modelName, m := range rpcResp.Models {
		modelHealths[modelName] = types.ModelHealth{
			Status:       m.Status,
			LastCheck:    "", // proto 中没有此字段，设置为空
			ResponseTime: m.AvgLatencyMs,
			ErrorMessage: "", // proto 中没有此字段，设置为空
		}
	}

	return &types.HealthCheckResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.HealthCheckData{
			Status: rpcResp.Status,
			Models: modelHealths,
		},
	}, nil
}
