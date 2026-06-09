package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	SysConfigRedis redis.RedisConf
	MasterDataRpc  zrpc.RpcClientConf // Master Data Service gRPC 客户端配置（用于 sysconfig fallback）
	DataSource struct {
		Driver string
		Source string
	}
	Minio struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
		UseSSL    bool
	}
}
