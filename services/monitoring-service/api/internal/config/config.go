package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Monitoring MonitoringConfig
}

type MonitoringConfig struct {
	Services     []ServiceMonitorConfig
	Containers   []ContainerMonitorConfig
	AiModelCheck AiModelCheckConfig
}

type ServiceMonitorConfig struct {
	Name            string
	DisplayName     string
	ApiPort         int
	RpcPort         int
	HealthEndpoint  string
	HealthTimeout   time.Duration
}

type ContainerMonitorConfig struct {
	Name        string
	DisplayName string
}

type AiModelCheckConfig struct {
	Endpoint string
	Timeout  time.Duration
}
