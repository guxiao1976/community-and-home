// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"
	"fmt"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 调用模型
func NewCallModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelLogic {
	return &CallModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallModelLogic) CallModel(req *types.ModelCallRequest) (resp *types.ModelCallResponse, err error) {
	// 构建参数
	parameters := make(map[string]string)
	if req.SystemPrompt != "" {
		parameters["system_prompt"] = req.SystemPrompt
	}
	if req.Temperature > 0 {
		parameters["temperature"] = fmt.Sprintf("%f", req.Temperature)
	}
	if req.MaxTokens > 0 {
		parameters["max_tokens"] = fmt.Sprintf("%d", req.MaxTokens)
	}

	// 调用 RPC 服务
	rpcReq := &pb.ModelCallRequest{
		ModelName:     req.ModelName,
		Prompt:        req.Prompt,
		Parameters:    parameters,
		CallerService: "ai-model-api",
	}

	rpcResp, err := l.svcCtx.AiModelRpc.CallModel(l.ctx, rpcReq)
	if err != nil {
		return &types.ModelCallResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &types.ModelCallResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.ModelCallData{
			Content:      rpcResp.Content,
			InputTokens:  rpcResp.InputTokens,
			OutputTokens: rpcResp.OutputTokens,
			Cost:         rpcResp.Cost,
			Latency:      rpcResp.LatencyMs,
			ModelVersion: rpcResp.ModelUsed,
		},
	}, nil
}
