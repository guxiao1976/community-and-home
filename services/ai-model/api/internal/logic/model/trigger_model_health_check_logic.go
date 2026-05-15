package model

import (
	"context"
	"fmt"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/aimodel"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerModelHealthCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTriggerModelHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerModelHealthCheckLogic {
	return &TriggerModelHealthCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TriggerModelHealthCheckLogic) TriggerModelHealthCheck(req *types.TriggerHealthCheckRequest) (resp *types.HealthCheckRecordResponse, err error) {
	// 调用 RPC 服务触发健康检查
	rpcResp, err := l.svcCtx.AiModelRpc.CheckModelHealth(l.ctx, &aimodel.ModelHealthCheckReq{
		ModelId: req.Id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to trigger health check: %w", err)
	}

	// 转换状态字符串为数字: healthy=1, unhealthy=0
	var status int64 = 0
	if rpcResp.Status == "healthy" {
		status = 1
	}

	// 转换响应
	resp = &types.HealthCheckRecordResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: types.HealthCheckRecord{
			ModelId:      req.Id,
			Status:       status,
			ResponseTime: rpcResp.ResponseTime,
			ErrorMessage: rpcResp.Message,
		},
	}

	return resp, nil
}
