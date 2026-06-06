package health

import (
	"context"
	"time"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

type HealthLogic struct {
	cfg config.MonitoringConfig
}

func NewHealthLogic(cfg config.MonitoringConfig) *HealthLogic {
	return &HealthLogic{cfg: cfg}
}

func (l *HealthLogic) CheckHealth(ctx context.Context) (*types.HealthResponse, error) {
	resp := &types.HealthResponse{
		Timestamp:     time.Now().Format(time.RFC3339),
		OverallStatus: "healthy",
	}

	resp.Services = checkServices(l.cfg)
	resp.Containers = checkContainers(l.cfg)
	resp.AiModels = checkAiModels(l.cfg)

	for _, s := range resp.Services {
		if s.ApiStatus == "unhealthy" || s.RpcStatus == "unhealthy" || s.HealthStatus == "unhealthy" {
			resp.OverallStatus = "unhealthy"
			break
		}
	}
	if resp.OverallStatus == "healthy" {
		for _, c := range resp.Containers {
			if c.Status == "unhealthy" {
				resp.OverallStatus = "unhealthy"
				break
			}
		}
	}
	if resp.OverallStatus == "healthy" {
		for _, m := range resp.AiModels {
			if m.Status == "unhealthy" {
				resp.OverallStatus = "unhealthy"
				break
			}
		}
	}

	return resp, nil
}
