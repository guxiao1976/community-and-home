package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 用户中心 RPC 服务配置
type Config struct {
	zrpc.RpcServerConf
	DataSource string          // MySQL 数据源
	Cache      redis.RedisConf // Redis 缓存配置
	AesKey     string          // AES 密钥（用于手机号加密存储）
}
