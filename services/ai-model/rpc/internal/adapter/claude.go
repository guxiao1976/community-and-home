package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ClaudeAdapter struct {
	config *ModelConfig
	client *http.Client
}

func NewClaudeAdapter(config *ModelConfig) *ClaudeAdapter {
	return &ClaudeAdapter{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (a *ClaudeAdapter) Call(ctx context.Context, req *CallRequest) (*CallResponse, error) {
	startTime := time.Now()

	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": req.Prompt,
		},
	}

	requestBody := map[string]interface{}{
		"model":      a.config.ModelName,
		"messages":   messages,
		"max_tokens": req.MaxTokens,
	}

	if req.SystemPrompt != "" {
		requestBody["system"] = req.SystemPrompt
	}

	if req.Temperature > 0 {
		requestBody["temperature"] = req.Temperature
	}

	for k, v := range req.Parameters {
		if _, exists := requestBody[k]; !exists {
			requestBody[k] = v
		}
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.config.APIEndpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var claudeResp struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model        string `json:"model"`
		StopReason   string `json:"stop_reason"`
		Usage        struct {
			InputTokens  int32 `json:"input_tokens"`
			OutputTokens int32 `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	var content string
	if len(claudeResp.Content) > 0 {
		content = claudeResp.Content[0].Text
	}

	cost := a.calculateCost(claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens)

	return &CallResponse{
		Content:      content,
		InputTokens:  claudeResp.Usage.InputTokens,
		OutputTokens: claudeResp.Usage.OutputTokens,
		Cost:         cost,
		Latency:      time.Since(startTime),
		ModelVersion: claudeResp.Model,
		FinishReason: claudeResp.StopReason,
	}, nil
}

func (a *ClaudeAdapter) HealthCheck(ctx context.Context) error {
	req := &CallRequest{
		Prompt:    "Hello",
		MaxTokens: 10,
		Timeout:   5 * time.Second,
	}

	_, err := a.Call(ctx, req)
	return err
}

func (a *ClaudeAdapter) EstimateCost(req *CallRequest) (float64, error) {
	estimatedInputTokens := int32(len(req.Prompt) / 4)
	estimatedOutputTokens := req.MaxTokens
	return a.calculateCost(estimatedInputTokens, estimatedOutputTokens), nil
}

func (a *ClaudeAdapter) GetModelInfo() *ModelInfo {
	return &ModelInfo{
		Provider:  "claude",
		ModelName: a.config.ModelName,
		Version:   "2023-06-01",
		MaxTokens: a.config.MaxTokens,
		SupportedFeatures: []string{
			"text-generation",
			"system-prompt",
			"temperature",
		},
	}
}

func (a *ClaudeAdapter) calculateCost(inputTokens, outputTokens int32) float64 {
	var inputCostPer1M, outputCostPer1M float64

	switch {
	case strings.Contains(a.config.ModelName, "opus"):
		inputCostPer1M = 15.0
		outputCostPer1M = 75.0
	case strings.Contains(a.config.ModelName, "sonnet"):
		inputCostPer1M = 3.0
		outputCostPer1M = 15.0
	case strings.Contains(a.config.ModelName, "haiku"):
		inputCostPer1M = 0.25
		outputCostPer1M = 1.25
	default:
		inputCostPer1M = 3.0
		outputCostPer1M = 15.0
	}

	inputCost := float64(inputTokens) / 1000000.0 * inputCostPer1M
	outputCost := float64(outputTokens) / 1000000.0 * outputCostPer1M

	return inputCost + outputCost
}
