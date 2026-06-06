package health

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

func checkServices(cfg config.MonitoringConfig) []types.ServiceHealth {
	results := make([]types.ServiceHealth, 0, len(cfg.Services))
	for _, svc := range cfg.Services {
		apiStatus, apiErr := checkPort(svc.ApiPort)
		rpcStatus, rpcErr := checkPort(svc.RpcPort)
		healthStatus, healthErr := checkHealth(svc.HealthEndpoint, svc.HealthTimeout)

		healthEndpoint := ""
		if svc.HealthEndpoint != "" {
			healthEndpoint = svc.HealthEndpoint
		}

		results = append(results, types.ServiceHealth{
			Name:          svc.Name,
			DisplayName:   svc.DisplayName,
			ApiPort:       svc.ApiPort,
			ApiStatus:     apiStatus,
			ApiError:      apiErr,
			RpcPort:       svc.RpcPort,
			RpcStatus:     rpcStatus,
			RpcError:      rpcErr,
			HealthStatus:  healthStatus,
			HealthError:   healthErr,
			HealthEndpoint: healthEndpoint,
		})
	}
	return results
}

func checkPort(port int) (string, string) {
	if port == 0 {
		return "unknown", ""
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return "unhealthy", err.Error()
	}
	conn.Close()
	return "healthy", ""
}

func checkHealth(endpoint string, timeout time.Duration) (string, string) {
	if endpoint == "" {
		return "unknown", ""
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(endpoint)
	if err != nil {
		return "unhealthy", err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "healthy", ""
	}
	return "unhealthy", fmt.Sprintf("HTTP %d", resp.StatusCode)
}
