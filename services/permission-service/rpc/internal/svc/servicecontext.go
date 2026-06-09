package svc

import (
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/config"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", nil)

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
