package logic

import (
	"context"
	"time"

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

// 健康检查
func (l *HealthCheckLogic) HealthCheck(in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	// Get all available models
	configs, err := l.svcCtx.ModelManager.GetAvailableModels(l.ctx, "")
	if err != nil {
		return &pb.HealthCheckResponse{
			Status:        "unhealthy",
			Models:        make(map[string]*pb.ModelHealthStatus),
			UptimeSeconds: 0,
			Version:       "1.0.0",
		}, nil
	}

	// Build model health status map
	modelStatuses := make(map[string]*pb.ModelHealthStatus)
	overallHealthy := true

	for _, config := range configs {
		status := &pb.ModelHealthStatus{
			Name:   config.ModelName,
			Status: config.HealthStatus,
		}

		// Get recent statistics for this model
		endDate := time.Now().Format("2006-01-02")
		startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		stats, err := l.svcCtx.CostManager.GetStatistics(l.ctx, config.Id, startDate, endDate)
		if err == nil && len(stats) > 0 {
			var totalCalls, successCalls int64
			var totalLatency int64

			for _, stat := range stats {
				totalCalls += stat.TotalCalls
				successCalls += stat.SuccessCalls
				totalLatency += stat.AvgLatencyMs
			}

			if totalCalls > 0 {
				status.TotalRequests = totalCalls
				status.SuccessRate = float64(successCalls) / float64(totalCalls)
				status.AvgLatencyMs = totalLatency / int64(len(stats))
			}
		}

		modelStatuses[config.ModelName] = status

		if config.HealthStatus != "healthy" {
			overallHealthy = false
		}
	}

	overallStatus := "healthy"
	if !overallHealthy {
		overallStatus = "degraded"
	}
	if len(modelStatuses) == 0 {
		overallStatus = "unhealthy"
	}

	return &pb.HealthCheckResponse{
		Status:        overallStatus,
		Models:        modelStatuses,
		UptimeSeconds: 0, // Would need to track service start time
		Version:       "1.0.0",
	}, nil
}
