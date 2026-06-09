package svc

import (
	"context"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/config"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 权限中心 RPC 服务上下文
type ServiceContext struct {
	Config    config.Config
	SysConfig *sysconfig.Client
	// RBAC 模型层
	RoleModel           model.SysRoleModel
	PermissionModel     model.SysPermissionModel
	RolePermissionModel model.RelRolePermissionModel
	UserRoleModel       model.RelUserRoleModel
	// Redis（权限缓存）
	RedisClient *redis.Client
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

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
		Config:              c,
		SysConfig:           sysCfg,
		RoleModel:           model.NewSysRoleModel(conn, c.Cache),
		PermissionModel:     model.NewSysPermissionModel(conn, c.Cache),
		RolePermissionModel: model.NewRelRolePermissionModel(conn, c.Cache),
		UserRoleModel:       model.NewRelUserRoleModel(conn, c.Cache),
		RedisClient:         redisClient,
	}
}
