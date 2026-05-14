package logic

import (
	"context"
	"fmt"
	"sync"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallModelBatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCallModelBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelBatchLogic {
	return &CallModelBatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通用模型调用（批量）
func (l *CallModelBatchLogic) CallModelBatch(in *pb.ModelBatchRequest) (*pb.ModelBatchResponse, error) {
	if in.ModelName == "" {
		return nil, fmt.Errorf("model_name is required")
	}
	if len(in.Prompts) == 0 {
		return nil, fmt.Errorf("prompts cannot be empty")
	}

	results := make([]*pb.ModelCallResponse, 0, len(in.Prompts))
	var totalCost float64
	var successCount, failedCount int32

	// 并发调用模型
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, prompt := range in.Prompts {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			// 构造单次调用请求
			req := &pb.ModelCallRequest{
				ModelName:     in.ModelName,
				Prompt:        p,
				Parameters:    in.Parameters,
				CallerService: in.CallerService,
			}

			// 调用单次模型接口
			callLogic := NewCallModelLogic(l.ctx, l.svcCtx)
			resp, err := callLogic.CallModel(req)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				l.Errorf("batch call failed for prompt: %v", err)
				failedCount++
				// 记录失败的响应
				results = append(results, &pb.ModelCallResponse{
					Content:   fmt.Sprintf("Error: %v", err),
					ModelUsed: in.ModelName,
				})
			} else {
				successCount++
				totalCost += resp.Cost
				results = append(results, resp)
			}
		}(prompt)
	}

	wg.Wait()

	return &pb.ModelBatchResponse{
		Results:   results,
		Total:     int32(len(in.Prompts)),
		Success:   successCount,
		Failed:    failedCount,
		TotalCost: totalCost,
	}, nil
}
