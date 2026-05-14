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

type OllamaAdapter struct {
	config *ModelConfig
	client *http.Client
}

func NewOllamaAdapter(config *ModelConfig) *OllamaAdapter {
	return &OllamaAdapter{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (a *OllamaAdapter) Call(ctx context.Context, req *CallRequest) (*CallResponse, error) {
	startTime := time.Now()

	prompt := req.Prompt
	if req.SystemPrompt != "" {
		prompt = req.SystemPrompt + "\n\n" + req.Prompt
	}

	requestBody := map[string]interface{}{
		"model":  a.config.ModelName,
		"prompt": prompt,
		"stream": false,
	}

	if req.Temperature > 0 {
		requestBody["options"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
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

	endpoint := strings.TrimSuffix(a.config.APIEndpoint, "/") + "/api/generate"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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

	var ollamaResp struct {
		Model     string `json:"model"`
		CreatedAt string `json:"created_at"`
		Response  string `json:"response"`
		Done      bool   `json:"done"`
		Context   []int  `json:"context"`
		TotalDuration     int64 `json:"total_duration"`
		LoadDuration      int64 `json:"load_duration"`
		PromptEvalCount   int32 `json:"prompt_eval_count"`
		PromptEvalDuration int64 `json:"prompt_eval_duration"`
		EvalCount         int32 `json:"eval_count"`
		EvalDuration      int64 `json:"eval_duration"`
	}

	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	cost := a.calculateCost(ollamaResp.PromptEvalCount, ollamaResp.EvalCount)

	return &CallResponse{
		Content:      ollamaResp.Response,
		InputTokens:  ollamaResp.PromptEvalCount,
		OutputTokens: ollamaResp.EvalCount,
		Cost:         cost,
		Latency:      time.Since(startTime),
		ModelVersion: ollamaResp.Model,
		FinishReason: "stop",
	}, nil
}

func (a *OllamaAdapter) HealthCheck(ctx context.Context) error {
	endpoint := strings.TrimSuffix(a.config.APIEndpoint, "/") + "/api/tags"
	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

func (a *OllamaAdapter) EstimateCost(req *CallRequest) (float64, error) {
	estimatedInputTokens := int32(len(req.Prompt) / 4)
	estimatedOutputTokens := req.MaxTokens
	return a.calculateCost(estimatedInputTokens, estimatedOutputTokens), nil
}

func (a *OllamaAdapter) GetModelInfo() *ModelInfo {
	return &ModelInfo{
		Provider:  "ollama",
		ModelName: a.config.ModelName,
		Version:   "local",
		MaxTokens: a.config.MaxTokens,
		SupportedFeatures: []string{
			"text-generation",
			"temperature",
			"local-inference",
		},
	}
}

func (a *OllamaAdapter) calculateCost(inputTokens, outputTokens int32) float64 {
	gpuCostPerHour := 0.5
	tokensPerSecond := 50.0

	totalTokens := float64(inputTokens + outputTokens)
	estimatedSeconds := totalTokens / tokensPerSecond
	estimatedHours := estimatedSeconds / 3600.0

	return estimatedHours * gpuCostPerHour
}
