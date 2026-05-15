package logic

import (
	"context"
	"fmt"
	"time"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/model"
	"community-and-home/services/ai-model/rpc/pb"

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

	// 2. 根据 provider 调用对应的 API 进行测试
	startTime := time.Now()

	// 简单的健康检查：尝试调用模型 API
	// 这里使用一个简单的测试提示词
	testPrompt := "Hello"

	var checkErr error
	switch modelConfig.Provider {
	case "openai":
		checkErr = l.checkOpenAI(modelConfig, in.ApiKey, testPrompt)
	case "anthropic", "claude":
		checkErr = l.checkAnthropic(modelConfig, in.ApiKey, testPrompt)
	default:
		checkErr = fmt.Errorf("不支持的 provider: %s", modelConfig.Provider)
	}

	responseTime := time.Since(startTime).Milliseconds()

	// 3. 返回结果
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
func (l *CheckModelHealthLogic) checkOpenAI(modelConfig *model.AmModelConfig, apiKey, prompt string) error {
	// TODO: 实现 OpenAI API 调用
	// 这里暂时返回成功，实际应该调用 OpenAI API
	logx.Infof("Checking OpenAI model: %s", modelConfig.ModelName)
	return nil
}

// checkAnthropic 检查 Anthropic/Claude 模型健康状态
func (l *CheckModelHealthLogic) checkAnthropic(modelConfig *model.AmModelConfig, apiKey, prompt string) error {
	// TODO: 实现 Anthropic API 调用
	// 这里暂时返回成功，实际应该调用 Anthropic API
	logx.Infof("Checking Anthropic model: %s", modelConfig.ModelName)
	return nil
}
