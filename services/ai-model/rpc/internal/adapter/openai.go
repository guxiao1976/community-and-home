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

type OpenAIAdapter struct {
	config *ModelConfig
	client *http.Client
}

func NewOpenAIAdapter(config *ModelConfig) *OpenAIAdapter {
	return &OpenAIAdapter{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (a *OpenAIAdapter) Call(ctx context.Context, req *CallRequest) (*CallResponse, error) {
	startTime := time.Now()

	messages := []map[string]interface{}{}

	if req.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": req.SystemPrompt,
		})
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": req.Prompt,
	})

	requestBody := map[string]interface{}{
		"model":      a.config.ModelName,
		"messages":   messages,
		"max_tokens": req.MaxTokens,
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
	httpReq.Header.Set("Authorization", "Bearer "+a.config.APIKey)

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

	var openaiResp struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index        int32 `json:"index"`
			Message      struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int32 `json:"prompt_tokens"`
			CompletionTokens int32 `json:"completion_tokens"`
			TotalTokens      int32 `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	var content string
	var finishReason string
	if len(openaiResp.Choices) > 0 {
		content = openaiResp.Choices[0].Message.Content
		finishReason = openaiResp.Choices[0].FinishReason
	}

	cost := a.calculateCost(openaiResp.Usage.PromptTokens, openaiResp.Usage.CompletionTokens)

	return &CallResponse{
		Content:      content,
		InputTokens:  openaiResp.Usage.PromptTokens,
		OutputTokens: openaiResp.Usage.CompletionTokens,
		Cost:         cost,
		Latency:      time.Since(startTime),
		ModelVersion: openaiResp.Model,
		FinishReason: finishReason,
	}, nil
}

func (a *OpenAIAdapter) HealthCheck(ctx context.Context) error {
	req := &CallRequest{
		Prompt:    "Hello",
		MaxTokens: 10,
		Timeout:   5 * time.Second,
	}

	_, err := a.Call(ctx, req)
	return err
}

func (a *OpenAIAdapter) EstimateCost(req *CallRequest) (float64, error) {
	estimatedInputTokens := int32(len(req.Prompt) / 4)
	estimatedOutputTokens := req.MaxTokens
	return a.calculateCost(estimatedInputTokens, estimatedOutputTokens), nil
}

func (a *OpenAIAdapter) GetModelInfo() *ModelInfo {
	return &ModelInfo{
		Provider:  "openai",
		ModelName: a.config.ModelName,
		Version:   "v1",
		MaxTokens: a.config.MaxTokens,
		SupportedFeatures: []string{
			"text-generation",
			"system-prompt",
			"temperature",
			"function-calling",
		},
	}
}

func (a *OpenAIAdapter) calculateCost(inputTokens, outputTokens int32) float64 {
	var inputCostPer1M, outputCostPer1M float64

	switch {
	case strings.Contains(a.config.ModelName, "gpt-4o"):
		inputCostPer1M = 2.5
		outputCostPer1M = 10.0
	case strings.Contains(a.config.ModelName, "gpt-4-turbo"):
		inputCostPer1M = 10.0
		outputCostPer1M = 30.0
	case strings.Contains(a.config.ModelName, "gpt-4"):
		inputCostPer1M = 30.0
		outputCostPer1M = 60.0
	case strings.Contains(a.config.ModelName, "gpt-3.5-turbo"):
		inputCostPer1M = 0.5
		outputCostPer1M = 1.5
	default:
		inputCostPer1M = 2.5
		outputCostPer1M = 10.0
	}

	inputCost := float64(inputTokens) / 1000000.0 * inputCostPer1M
	outputCost := float64(outputTokens) / 1000000.0 * outputCostPer1M

	return inputCost + outputCost
}
