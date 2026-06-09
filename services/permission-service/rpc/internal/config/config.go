package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 权限中心 RPC 服务配置
type Config struct {
	zrpc.RpcServerConf
	DataSource     string
	Cache          cache.CacheConf
	SysConfigRedis redis.RedisConf
}
