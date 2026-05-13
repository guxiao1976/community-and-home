package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RemoteLLMClient is an HTTP-based LLM moderation client that talks to a
// remote moderation service (Python inference service).
type RemoteLLMClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewRemoteLLMClient creates a new RemoteLLMClient.
func NewRemoteLLMClient(endpoint string) *RemoteLLMClient {
	if endpoint == "" {
		endpoint = "http://localhost:8001"
	}
	return &RemoteLLMClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type moderateRequest struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type moderateResponse struct {
	IsSafe     bool     `json:"is_safe"`
	RiskLevel  string   `json:"risk_level"`
	Categories []string `json:"categories"`
	Confidence float64  `json:"confidence"`
	Reason     string   `json:"reason"`
}

// CheckText sends text content to a remote LLM service for compliance checking.
func (c *RemoteLLMClient) CheckText(ctx context.Context, content, contentType string) (*CheckResult, error) {
	reqBody := moderateRequest{
		Content:     content,
		ContentType: contentType,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/moderate/text", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var modResp moderateResponse
	if err := json.NewDecoder(resp.Body).Decode(&modResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Map response to CheckResult
	result := &CheckResult{
		Compliant:  modResp.IsSafe,
		Confidence: modResp.Confidence,
		Reason:     modResp.Reason,
	}

	// Use first risk category if available
	if len(modResp.Categories) > 0 {
		result.Category = strings.Join(modResp.Categories, ", ")
	}

	return result, nil
}

// CheckImage sends image data to a remote LLM service for compliance checking.
func (c *RemoteLLMClient) CheckImage(_ context.Context, _ []byte, _ string) (*CheckResult, error) {
	return nil, ErrNotImplemented
}
