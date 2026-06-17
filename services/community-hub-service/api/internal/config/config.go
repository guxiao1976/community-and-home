package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 社区枢纽 REST API 配置
type Config struct {
	rest.RestConf
	CommunityHubRpc zrpc.RpcClientConf
}
