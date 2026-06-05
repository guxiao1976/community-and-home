package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

type aiModelResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Models []aiModelItem `json:"models"`
	} `json:"data"`
}

type aiModelItem struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Provider    string `json:"provider"`
	Status      int    `json:"status"`
}

func checkAiModels(cfg config.MonitoringConfig) []types.AiModelHealth {
	client := &http.Client{
		Timeout: cfg.AiModelCheck.Timeout,
	}
	if client.Timeout == 0 {
		client.Timeout = 5 * time.Second
	}

	resp, err := client.Get(cfg.AiModelCheck.Endpoint)
	if err != nil {
		return []types.AiModelHealth{{
			Id:          "",
			Name:        "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:    "",
			Status:      "unhealthy",
			Error:       fmt.Sprintf("无法连接 AI 模型服务: %v", err),
		}}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []types.AiModelHealth{{
			Id:          "",
			Name:        "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:    "",
			Status:      "unhealthy",
			Error:       fmt.Sprintf("AI 模型服务返回 %d", resp.StatusCode),
		}}
	}

	var result aiModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []types.AiModelHealth{{
			Id:          "",
			Name:        "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:    "",
			Status:      "unhealthy",
			Error:       fmt.Sprintf("解析模型列表失败: %v", err),
		}}
	}

	if result.Code != 0 {
		return []types.AiModelHealth{{
			Id:          "",
			Name:        "ai-model-service",
			DisplayName: "AI 模型服务",
			Provider:    "",
			Status:      "unhealthy",
			Error:       fmt.Sprintf("AI 模型服务错误: %s", result.Msg),
		}}
	}

	models := make([]types.AiModelHealth, 0, len(result.Data.Models))
	for _, m := range result.Data.Models {
		status := "unhealthy"
		if m.Status == 0 {
			status = "healthy"
		}
		models = append(models, types.AiModelHealth{
			Id:          fmt.Sprintf("%d", m.Id),
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Provider:    m.Provider,
			Status:      status,
			Error:       "",
		})
	}
	return models
}
