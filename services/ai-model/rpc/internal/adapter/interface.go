package adapter

import (
	"context"
	"time"
)

type ModelAdapter interface {
	Call(ctx context.Context, req *CallRequest) (*CallResponse, error)
	HealthCheck(ctx context.Context) error
	EstimateCost(req *CallRequest) (float64, error)
	GetModelInfo() *ModelInfo
}

type CallRequest struct {
	Prompt       string
	SystemPrompt string
	Parameters   map[string]interface{}
	Timeout      time.Duration
	MaxTokens    int32
	Temperature  float32
}

type CallResponse struct {
	Content      string
	InputTokens  int32
	OutputTokens int32
	Cost         float64
	Latency      time.Duration
	ModelVersion string
	FinishReason string
}

type ModelInfo struct {
	Provider     string
	ModelName    string
	Version      string
	MaxTokens    int32
	SupportedFeatures []string
}

type ModelConfig struct {
	Provider     string
	ModelName    string
	APIEndpoint  string
	APIKey       string
	MaxTokens    int32
	Timeout      time.Duration
	RetryCount   int32
	Parameters   map[string]interface{}
}
