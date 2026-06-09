package svc

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	"github.com/guxiao1976/community-auth/model"
	"github.com/guxiao1976/community-auth/rpc/internal/config"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 认证中心 RPC 服务上下文
type ServiceContext struct {
	Config         config.Config
	SysConfig      *sysconfig.Client
	CredentialModel model.AuthCredentialModel  // 登录凭证 DB 模型
	RedisClient    *redis.Client               // Redis（RT 存储、AT 黑名单）
	UserServiceRpc userv1.UserServiceClient    // User Service gRPC 客户端
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	// 创建 Redis 客户端（用于 RT 和 AT 黑名单，不使用 go-zero 的 Cache）
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // docker-compose 中 Redis 未设置密码
		DB:       0,
	})

	// 创建 User Service gRPC 客户端
	userRpcCli := zrpc.MustNewClient(c.UserServiceRpc)

	// 初始化系统参数配置客户端
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", func(ctx context.Context, key string) (*sysconfig.ConfigValue, error) {
		// gRPC fallback to master-data GetConfig
		conn, err := zrpc.NewClient(c.MasterDataRpc)
		if err != nil {
			return nil, err
		}
		client := masterdatav1.NewMasterdataServiceClient(conn.Conn())
		resp, err := client.GetConfig(ctx, &masterdatav1.GetConfigReq{ConfigKey: key})
		if err != nil {
			return nil, err
		}
		return &sysconfig.ConfigValue{
			Value: resp.ConfigValue,
			Type:  resp.ValueType,
			Desc:  resp.Description,
		}, nil
	})

	return &ServiceContext{
		Config:          c,
		SysConfig:       sysCfg,
		CredentialModel: model.NewAuthCredentialModel(conn, c.Cache),
		RedisClient:     redisClient,
		UserServiceRpc:  userv1.NewUserServiceClient(userRpcCli.Conn()),
	}
}
