package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 用户中心 RPC 服务配置
type Config struct {
	zrpc.RpcServerConf
	DataSource     string          // MySQL 数据源
	Cache          redis.RedisConf // Redis 缓存配置
	SysConfigRedis redis.RedisConf // 系统参数配置 Redis（用于 sysconfig 客户端）
	AesKey         string          // AES 密钥（用于手机号加密存储）
	MasterDataRpc  zrpc.RpcClientConf // Master Data Service gRPC 客户端配置（用于 sysconfig fallback）
}
