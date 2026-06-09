package svc

import (
	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-auth/api/internal/config"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext REST API 服务上下文
type ServiceContext struct {
	Config      config.Config
	AuthRpc     authv1.AuthServiceClient // Auth Service gRPC 客户端（调用自身 RPC 层）
	RedisClient *redis.Client            // Redis 客户端（短信验证码存储和限流）
	SysConfig   *sysconfig.Client        // 系统参数配置读取器
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	// 创建 Auth Service gRPC 客户端（通过 etcd 发现 auth.rpc）
	authRpcCli := zrpc.MustNewClient(c.AuthRpc)

	// 创建 Redis 客户端（用于短信验证码存储和限流）
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.RedisAddr,
		Password: "",
		DB:       0,
	})

	// 初始化系统参数配置客户端
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", nil)

	return &ServiceContext{
		Config:      c,
		AuthRpc:     authv1.NewAuthServiceClient(authRpcCli.Conn()),
		RedisClient: redisClient,
		SysConfig:   sysCfg,
	}
}
