package svc

import (
	"context"
	"fmt"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 社区枢纽 RPC 服务上下文
type ServiceContext struct {
	Config                config.Config
	SysConfig             *sysconfig.Client
	NoticeModel           model.NoticeModel
	NoticeAttachmentModel model.NoticeAttachmentModel
	CommunityContactModel model.CommunityContactModel
	LostFoundItemModel    model.LostFoundItemModel
	ModerationClient      moderationv1.ModerationServiceClient
	RedisClient           *redis.Redis
	PermissionClient      permissionv1.PermissionServiceClient // 数据权限（AssertPublishScope/GetDataScopes）
	MasterDataClient      masterdatav1.MasterdataServiceClient // 板块配额（GetSectionQuota）
}

// NewServiceContext 创建服务上下文，初始化所有数据库模型
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

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

	// Initialize moderation gRPC client
	modConn, err := zrpc.NewClient(c.ModerationRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create moderation client: %v", err))
	}
	modClient := moderationv1.NewModerationServiceClient(modConn.Conn())

	// Initialize Redis client for task queue
	redisClient := redis.New(c.ModerationRedis.Host, func(r *redis.Redis) {
		r.Pass = c.ModerationRedis.Pass
	})

	// Initialize permission-service client (data scope: AssertPublishScope / GetDataScopes)
	permConn, err := zrpc.NewClient(c.PermissionRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create permission client: %v", err))
	}
	permClient := permissionv1.NewPermissionServiceClient(permConn.Conn())

	// Initialize master-data client (section quota: GetSectionQuota)
	mdConn, err := zrpc.NewClient(c.MasterDataRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create master-data client: %v", err))
	}
	mdClient := masterdatav1.NewMasterdataServiceClient(mdConn.Conn())

	return &ServiceContext{
		Config:                c,
		SysConfig:             sysCfg,
		NoticeModel:           model.NewNoticeModel(conn),
		NoticeAttachmentModel: model.NewNoticeAttachmentModel(conn),
		CommunityContactModel: model.NewCommunityContactModel(conn),
		LostFoundItemModel:    model.NewLostFoundItemModel(conn),
		ModerationClient:      modClient,
		RedisClient:           redisClient,
		PermissionClient:      permClient,
		MasterDataClient:      mdClient,
	}
}
