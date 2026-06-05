package health

import (
	"fmt"
	"net"
	"time"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/types"
)

func checkServices(cfg config.MonitoringConfig) []types.ServiceHealth {
	results := make([]types.ServiceHealth, 0, len(cfg.Services))
	for _, svc := range cfg.Services {
		apiStatus, apiErr := checkPort(svc.ApiPort)
		rpcStatus, rpcErr := checkPort(svc.RpcPort)
		results = append(results, types.ServiceHealth{
			Name:        svc.Name,
			DisplayName: svc.DisplayName,
			ApiPort:     svc.ApiPort,
			ApiStatus:   apiStatus,
			ApiError:    apiErr,
			RpcPort:     svc.RpcPort,
			RpcStatus:   rpcStatus,
			RpcError:    rpcErr,
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
