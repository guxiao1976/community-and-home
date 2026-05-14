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

type CallModelBatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量调用模型
func NewCallModelBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelBatchLogic {
	return &CallModelBatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallModelBatchLogic) CallModelBatch(req *types.ModelBatchRequest) (resp *types.ModelBatchResponse, err error) {
	if len(req.Requests) == 0 {
		return &types.ModelBatchResponse{
			BaseResponse: types.BaseResponse{
				Code:    400,
				Message: "requests cannot be empty",
			},
		}, nil
	}

	// 按 model_name 分组请求
	groupedRequests := make(map[string][]types.ModelCallRequest)
	for _, r := range req.Requests {
		groupedRequests[r.ModelName] = append(groupedRequests[r.ModelName], r)
	}

	// 收集所有结果
	var allResults []types.ModelCallData
	var totalSuccess, totalFailed int32

	// 对每个模型分别调用
	for modelName, requests := range groupedRequests {
		// 提取 prompts
		prompts := make([]string, 0, len(requests))
		for _, r := range requests {
			prompts = append(prompts, r.Prompt)
		}

		// 使用第一个请求的参数（假设同一模型的请求使用相同参数）
		parameters := make(map[string]string)
		firstReq := requests[0]
		if firstReq.SystemPrompt != "" {
			parameters["system_prompt"] = firstReq.SystemPrompt
		}
		if firstReq.Temperature > 0 {
			parameters["temperature"] = fmt.Sprintf("%f", firstReq.Temperature)
		}
		if firstReq.MaxTokens > 0 {
			parameters["max_tokens"] = fmt.Sprintf("%d", firstReq.MaxTokens)
		}

		// 调用 RPC 批量接口
		rpcResp, err := l.svcCtx.AiModelRpc.CallModelBatch(l.ctx, &pb.ModelBatchRequest{
			ModelName:     modelName,
			Prompts:       prompts,
			Parameters:    parameters,
			CallerService: "ai-model-api",
		})
		if err != nil {
			// 如果某个模型调用失败，记录错误但继续处理其他模型
			logx.Errorf("CallModelBatch failed for model %s: %v", modelName, err)
			totalFailed += int32(len(requests))
			continue
		}

		// 转换响应
		for _, r := range rpcResp.Results {
			allResults = append(allResults, types.ModelCallData{
				Content:      r.Content,
				InputTokens:  r.InputTokens,
				OutputTokens: r.OutputTokens,
				Cost:         r.Cost,
				Latency:      r.LatencyMs,
				ModelVersion: r.ModelUsed,
			})
		}

		totalSuccess += rpcResp.Success
		totalFailed += rpcResp.Failed
	}

	return &types.ModelBatchResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.ModelBatchData{
			Results:       allResults,
			TotalRequests: int32(len(req.Requests)),
			SuccessCount:  totalSuccess,
		},
	}, nil
}
