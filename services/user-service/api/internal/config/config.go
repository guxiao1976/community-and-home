package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config user-service REST API 配置
type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	UserRpc        zrpc.RpcClientConf
	PermissionRpc  zrpc.RpcClientConf
	SysConfigRedis redis.RedisConf // 系统参数配置 Redis（用于 sysconfig 客户端）
}
