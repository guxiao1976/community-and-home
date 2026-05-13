package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type TextModerationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTextModerationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TextModerationLogic {
	return &TextModerationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Python 引擎请求/响应结构
type pythonTextRequest struct {
	Content         string            `json:"content"`
	CheckCategories []string          `json:"check_categories,omitempty"`
	Context         string            `json:"context,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type pythonTextResponse struct {
	IsSafe       bool     `json:"is_safe"`
	RiskLevel    string   `json:"risk_level"`
	Categories   []string `json:"categories"`
	Reason       string   `json:"reason"`
	Confidence   float64  `json:"confidence"`
	LatencyMs    int64    `json:"latency_ms"`
	ModelVersion string   `json:"model_version"`
}

func (l *TextModerationLogic) ModerateText(in *pb.TextModerationRequest) (*pb.TextModerationResponse, error) {
	startTime := time.Now()

	// 构建请求
	reqBody := pythonTextRequest{
		Content:         in.Content,
		CheckCategories: in.CheckCategories,
		Context:         in.Context,
		Metadata:        in.Metadata,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		l.Errorf("Failed to marshal request: %v", err)
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 调用 Python 引擎
	url := fmt.Sprintf("%s/api/moderate/text", l.svcCtx.PythonURL)
	req, err := http.NewRequestWithContext(l.ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		l.Errorf("Failed to create request: %v", err)
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求（带重试）
	var resp *http.Response
	maxRetries := l.svcCtx.Config.PythonEngine.MaxRetries
	for i := 0; i <= maxRetries; i++ {
		resp, err = l.svcCtx.PythonClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if i < maxRetries {
			l.Infof("Request failed (attempt %d/%d): %v", i+1, maxRetries+1, err)
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}

	if err != nil {
		l.Errorf("Failed to call Python engine after %d retries: %v", maxRetries+1, err)
		return nil, fmt.Errorf("python engine unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		l.Errorf("Python engine returned error: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("python engine error: status=%d", resp.StatusCode)
	}

	// 解析响应
	var pythonResp pythonTextResponse
	if err := json.NewDecoder(resp.Body).Decode(&pythonResp); err != nil {
		l.Errorf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	// 转换为 gRPC 响应
	totalLatency := time.Since(startTime).Milliseconds()

	result := &pb.TextModerationResponse{
		IsSafe:       pythonResp.IsSafe,
		RiskLevel:    pythonResp.RiskLevel,
		Categories:   pythonResp.Categories,
		Reason:       pythonResp.Reason,
		Confidence:   pythonResp.Confidence,
		LatencyMs:    totalLatency,
		ModelVersion: pythonResp.ModelVersion,
	}

	l.Infof("Text moderation completed: is_safe=%v, risk_level=%s, latency=%dms",
		result.IsSafe, result.RiskLevel, result.LatencyMs)

	return result, nil
}
