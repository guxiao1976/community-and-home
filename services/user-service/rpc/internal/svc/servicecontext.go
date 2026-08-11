package svc

import (
	"context"
	"fmt"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 用户中心 RPC 服务上下文
type ServiceContext struct {
	Config config.Config

	Redis     *redis.Redis      // Redis 缓存客户端
	SysConfig *sysconfig.Client // 系统参数配置读取器

	UserBaseModel                model.UserBaseModel
	UserCommunityMembershipModel model.UserCommunityMembershipModel
	UserCertificationModel       model.UserCertificationModel
	UserResidenceModel           model.UserResidenceModel

	ModerationClient moderationv1.ModerationServiceClient
	RedisClient      *redis.Redis
	PermissionClient permissionv1.PermissionServiceClient // permission-service gRPC 客户端（失效权限缓存）
}

// NewServiceContext 创建服务上下文，初始化所有数据库模型、Redis 和加密组件
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	// 初始化全局 AES 加密
	if c.AesKey != "" {
		if err := crypto.InitAES(c.AesKey); err != nil {
			panic("init AES failed: " + err.Error())
		}
	}

	// 初始化 Redis
	var rds *redis.Redis
	if c.Cache.Host != "" {
		rds = c.Cache.NewRedis()
	}

	// 初始化系统参数配置客户端（无 gRPC fallback，需单独配置）
	// 若要启用 Redis 未命中时回退到 master-data-service，可传入带 GetConfig RPC 的 FallbackFunc
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

	// Initialize moderation gRPC client
	modConn, err := zrpc.NewClient(c.ModerationRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create moderation client: %v", err))
	}
	modClient := moderationv1.NewModerationServiceClient(modConn.Conn())

	// Initialize Redis client for moderation task queue
	redisClient := redis.New(c.ModerationRedis.Host, func(r *redis.Redis) {
		r.Pass = c.ModerationRedis.Pass
	})

	// Initialize permission-service gRPC client
	permConn, err := zrpc.NewClient(c.PermissionRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create permission client: %v", err))
	}
	permClient := permissionv1.NewPermissionServiceClient(permConn.Conn())

	return &ServiceContext{
		Config:                       c,
		Redis:                        rds,
		SysConfig:                    sysCfg,
		UserBaseModel:                model.NewUserBaseModel(conn),
		UserCommunityMembershipModel: model.NewUserCommunityMembershipModel(conn),
		UserCertificationModel:       model.NewUserCertificationModel(conn),
		UserResidenceModel:           model.NewUserResidenceModel(conn),
		ModerationClient:             modClient,
		RedisClient:                  redisClient,
		PermissionClient:             permClient,
	}
}
