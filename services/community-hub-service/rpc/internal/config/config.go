package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 社区枢纽 RPC 服务配置
type Config struct {
	zrpc.RpcServerConf
	DataSource     string          // MySQL DSN（数据库 community_hub_db）
	SysConfigRedis redis.RedisConf // 系统参数配置 Redis
	MasterDataRpc  zrpc.RpcClientConf // Master Data Service gRPC 客户端配置（用于 sysconfig fallback）
}
