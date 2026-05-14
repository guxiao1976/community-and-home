package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"community-and-home/services/ai-model/rpc/internal/adapter"
	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/model"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCallModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelLogic {
	return &CallModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通用模型调用（单次）
func (l *CallModelLogic) CallModel(in *pb.ModelCallRequest) (*pb.ModelCallResponse, error) {
	// Get model config by name
	modelConfig, err := l.getModelByName(in.ModelName)
	if err != nil {
		return nil, fmt.Errorf("get model config: %w", err)
	}

	// Get adapter for the model
	adp, err := l.svcCtx.ModelManager.GetAdapter(l.ctx, modelConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("get adapter: %w", err)
	}

	// Prepare prompt (use template if provided)
	prompt := in.Prompt
	if in.TemplateId != "" {
		renderedPrompt, err := l.svcCtx.TemplateManager.RenderTemplateByName(l.ctx, in.TemplateId, in.Variables)
		if err != nil {
			return nil, fmt.Errorf("render template: %w", err)
		}
		prompt = renderedPrompt
	}

	// Convert parameters
	parameters := make(map[string]interface{})
	for k, v := range in.Parameters {
		parameters[k] = v
	}

	// Call the model
	startTime := time.Now()
	req := &adapter.CallRequest{
		Prompt:     prompt,
		Parameters: parameters,
	}

	resp, err := adp.Call(l.ctx, req)
	latencyMs := time.Since(startTime).Milliseconds()

	// Record call log
	callLog := &model.AmCallLog{
		ModelId:       modelConfig.Id,
		ModelName:     modelConfig.ModelName,
		CallerService: sql.NullString{String: in.CallerService, Valid: in.CallerService != ""},
		Method:        sql.NullString{String: "CallModel", Valid: true},
		PromptLength:  sql.NullInt64{Int64: int64(len(prompt)), Valid: true},
		InputTokens:   0,
		OutputTokens:  0,
		TotalTokens:   0,
		Cost:          0,
		LatencyMs:     sql.NullInt64{Int64: latencyMs, Valid: true},
		Success:       0,
		CreatedTime:   time.Now(),
	}

	if err != nil {
		callLog.ErrorMsg = sql.NullString{String: err.Error(), Valid: true}
		l.svcCtx.CostManager.RecordCall(l.ctx, callLog)
		return nil, fmt.Errorf("call model: %w", err)
	}

	// Update call log with success data
	callLog.Success = 1
	callLog.InputTokens = int64(resp.InputTokens)
	callLog.OutputTokens = int64(resp.OutputTokens)
	callLog.TotalTokens = int64(resp.InputTokens + resp.OutputTokens)
	callLog.Cost = resp.Cost

	if err := l.svcCtx.CostManager.RecordCall(l.ctx, callLog); err != nil {
		l.Errorf("failed to record call log: %v", err)
	}

	return &pb.ModelCallResponse{
		Content:      resp.Content,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
		Cost:         resp.Cost,
		LatencyMs:    latencyMs,
		ModelUsed:    modelConfig.ModelName,
		RequestId:    fmt.Sprintf("%d", time.Now().UnixNano()),
	}, nil
}

func (l *CallModelLogic) getModelByName(modelName string) (*model.AmModelConfig, error) {
	query := "SELECT * FROM am_model_config WHERE model_name = ? AND status = 1 LIMIT 1"

	var config model.AmModelConfig
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &config, query, modelName)
	if err != nil {
		return nil, fmt.Errorf("query model config: %w", err)
	}

	return &config, nil
}
