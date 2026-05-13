package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthCheckLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type pythonHealthResponse struct {
	Status                string  `json:"status"`
	ModelLoaded           bool    `json:"model_loaded"`
	ModelName             string  `json:"model_name"`
	ModelVersion          string  `json:"model_version"`
	UptimeSeconds         int64   `json:"uptime_seconds"`
	TotalRequests         int64   `json:"total_requests"`
	AvgLatencyMs          float64 `json:"avg_latency_ms"`
	GpuAvailable          bool    `json:"gpu_available"`
	GpuMemoryAllocatedMb  float64 `json:"gpu_memory_allocated_mb"`
}

func (l *HealthCheckLogic) HealthCheck(in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	// 调用 Python 引擎健康检查
	url := fmt.Sprintf("%s/health", l.svcCtx.PythonURL)
	req, err := http.NewRequestWithContext(l.ctx, "GET", url, nil)
	if err != nil {
		return &pb.HealthCheckResponse{
			Status: "error",
			Models: map[string]*pb.ModelStatus{
				"text": {
					Name:    "Qwen2.5-7B-Instruct",
					Version: "v1.0",
					Loaded:  false,
				},
			},
		}, nil
	}

	resp, err := l.svcCtx.PythonClient.Do(req)
	if err != nil {
		l.Errorf("Python engine health check failed: %v", err)
		return &pb.HealthCheckResponse{
			Status: "degraded",
			Models: map[string]*pb.ModelStatus{
				"text": {
					Name:    "Qwen2.5-7B-Instruct",
					Version: "v1.0",
					Loaded:  false,
				},
			},
		}, nil
	}
	defer resp.Body.Close()

	var healthResp pythonHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		l.Errorf("Failed to decode health response: %v", err)
		return &pb.HealthCheckResponse{
			Status: "error",
		}, nil
	}

	// 构建响应
	status := "ok"
	if !healthResp.ModelLoaded {
		status = "degraded"
	}

	return &pb.HealthCheckResponse{
		Status:        status,
		UptimeSeconds: healthResp.UptimeSeconds,
		Models: map[string]*pb.ModelStatus{
			"text": {
				Name:           healthResp.ModelName,
				Version:        healthResp.ModelVersion,
				Loaded:         healthResp.ModelLoaded,
				MemoryMb:       int64(healthResp.GpuMemoryAllocatedMb),
				AvgLatencyMs:   healthResp.AvgLatencyMs,
				TotalRequests:  healthResp.TotalRequests,
			},
		},
	}, nil
}
