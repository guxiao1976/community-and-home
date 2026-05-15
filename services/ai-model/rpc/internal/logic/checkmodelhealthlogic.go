package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/adapter"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/model"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckModelHealthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckModelHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckModelHealthLogic {
	return &CheckModelHealthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 单个模型健康检查
func (l *CheckModelHealthLogic) CheckModelHealth(in *pb.ModelHealthCheckReq) (*pb.ModelHealthCheckResp, error) {
	// 1. 获取模型配置
	modelConfig, err := l.svcCtx.ModelConfigModel.FindOne(l.ctx, in.ModelId)
	if err != nil {
		return &pb.ModelHealthCheckResp{
			Status:  "unhealthy",
			Message: fmt.Sprintf("模型配置不存在: %v", err),
		}, nil
	}

	// 2. 获取 API Key
	apiKey := in.ApiKey
	if apiKey == "" {
		// 如果请求中没有提供 API Key，尝试从数据库获取
		apiKeyModel, err := l.svcCtx.ApiKeyModel.FindOneByModelId(l.ctx, in.ModelId)
		if err != nil {
			return &pb.ModelHealthCheckResp{
				Status:  "unhealthy",
				Message: fmt.Sprintf("未找到可用的 API Key: %v", err),
			}, nil
		}
		apiKey = apiKeyModel.ApiKey
	}

	// 3. 根据 provider 调用对应的 API 进行测试
	startTime := time.Now()

	var checkErr error
	switch modelConfig.Provider {
	case "openai":
		checkErr = l.checkOpenAI(modelConfig, apiKey)
	case "anthropic", "claude":
		checkErr = l.checkAnthropic(modelConfig, apiKey)
	default:
		checkErr = fmt.Errorf("不支持的 provider: %s", modelConfig.Provider)
	}

	responseTime := time.Since(startTime).Milliseconds()

	// 4. 返回结果
	if checkErr != nil {
		return &pb.ModelHealthCheckResp{
			Status:       "unhealthy",
			ResponseTime: responseTime,
			Message:      checkErr.Error(),
		}, nil
	}

	return &pb.ModelHealthCheckResp{
		Status:       "healthy",
		ResponseTime: responseTime,
		Message:      "模型响应正常",
	}, nil
}

// checkOpenAI 检查 OpenAI 模型健康状态
func (l *CheckModelHealthLogic) checkOpenAI(modelConfig *model.AmModelConfig, apiKey string) error {
	// 获取 endpoint，如果为空则使用默认值
	endpoint := "https://api.openai.com/v1/chat/completions"
	if modelConfig.Endpoint.Valid && modelConfig.Endpoint.String != "" {
		endpoint = modelConfig.Endpoint.String
	}

	// 创建 OpenAI Adapter
	adapterConfig := &adapter.ModelConfig{
		Provider:    modelConfig.Provider,
		ModelName:   modelConfig.ModelName,
		APIEndpoint: endpoint,
		APIKey:      apiKey,
		MaxTokens:   100,
		Timeout:     10 * time.Second,
	}

	openaiAdapter := adapter.NewOpenAIAdapter(adapterConfig)
	return openaiAdapter.HealthCheck(l.ctx)
}

// checkAnthropic 检查 Anthropic/Claude 模型健康状态
func (l *CheckModelHealthLogic) checkAnthropic(modelConfig *model.AmModelConfig, apiKey string) error {
	// 获取 endpoint，如果为空则使用默认值
	endpoint := "https://api.anthropic.com/v1/messages"
	if modelConfig.Endpoint.Valid && modelConfig.Endpoint.String != "" {
		endpoint = modelConfig.Endpoint.String
	}

	// 创建 Claude Adapter
	adapterConfig := &adapter.ModelConfig{
		Provider:    modelConfig.Provider,
		ModelName:   modelConfig.ModelName,
		APIEndpoint: endpoint,
		APIKey:      apiKey,
		MaxTokens:   100,
		Timeout:     10 * time.Second,
	}

	claudeAdapter := adapter.NewClaudeAdapter(adapterConfig)
	return claudeAdapter.HealthCheck(l.ctx)
}
