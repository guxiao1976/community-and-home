package svc

import (
	"context"
	"fmt"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/config"
	"github.com/guxiao1976/community-hub/rpc/internal/kafkapush"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 社区枢纽 RPC 服务上下文
type ServiceContext struct {
	Config                     config.Config
	SysConfig                  *sysconfig.Client
	Conn                       sqlx.SqlConn
	ContentPostModel           model.ContentPostModel
	ContentPostScopeModel      model.ContentPostScopeModel
	ContentPostAttachmentModel model.ContentPostAttachmentModel
	CommunityContactModel      model.CommunityContactModel
	LostFoundItemModel         model.LostFoundItemModel
	ModerationClient           moderationv1.ModerationServiceClient
	RedisClient                *redis.Redis
	PermissionClient           permissionv1.PermissionServiceClient // 数据权限（AssertPublishScope/GetDataScopes/GetUserRoles）
	MasterDataClient           masterdatav1.MasterdataServiceClient // 板块配额（GetSectionQuota）/ division 展开（GetResidentialArea/GetResidentialAreasByDivision）
	UserClient                 userv1.UserServiceClient             // user-service（GetUsersByIds 档案查询，R5）
	FileClient                 filev1.FileServiceClient             // file-service（GetFileUrl 附件绑定/重生，REQ-CPB-6）
	KafkaProducer              kafkapush.Pusher                     // content-review 生产者（D20，接口供测试注入 mock）
	KafkaRescanner             *kafkapush.Rescanner                 // 定时重推扫描器（D20）
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

	// Initialize master-data client (section quota: GetSectionQuota / division: GetResidentialArea/GetResidentialAreasByDivision)
	mdConn, err := zrpc.NewClient(c.MasterDataRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create master-data client: %v", err))
	}
	mdClient := masterdatav1.NewMasterdataServiceClient(mdConn.Conn())

	// Initialize user-service client (publisher display name: GetUsersByIds, R5)
	userConn, err := zrpc.NewClient(c.UserRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create user client: %v", err))
	}
	userClient := userv1.NewUserServiceClient(userConn.Conn())

	// Initialize file-service client (attachment binding/regeneration: GetFileUrl, REQ-CPB-6)
	fileConn, err := zrpc.NewClient(c.FileRpc)
	if err != nil {
		panic(fmt.Sprintf("failed to create file client: %v", err))
	}
	fileClient := filev1.NewFileServiceClient(fileConn.Conn())

	// Initialize content-post models + Kafka producer/rescanner（D20：at-least-once 待推标记 + 定时重推）
	contentPostModel := model.NewContentPostModel(conn)
	contentPostAttachmentModel := model.NewContentPostAttachmentModel(conn)
	var kafkaProducer *kafkapush.Producer
	var kafkaRescanner *kafkapush.Rescanner
	if len(c.Kafka.Brokers) > 0 {
		kafkaProducer = kafkapush.NewProducer(c.Kafka, contentPostModel, fileClient)
		retryInterval := time.Duration(c.Kafka.RetryIntervalSeconds) * time.Second
		if retryInterval <= 0 {
			retryInterval = 60 * time.Second
		}
		kafkaRescanner = kafkapush.NewRescanner(kafkaProducer, contentPostModel, contentPostAttachmentModel, retryInterval)
		kafkaRescanner.Start(context.Background())
	}

	return &ServiceContext{
		Config:                     c,
		SysConfig:                  sysCfg,
		Conn:                       conn,
		ContentPostModel:           contentPostModel,
		ContentPostScopeModel:      model.NewContentPostScopeModel(conn),
		ContentPostAttachmentModel: contentPostAttachmentModel,
		CommunityContactModel:      model.NewCommunityContactModel(conn),
		LostFoundItemModel:         model.NewLostFoundItemModel(conn),
		ModerationClient:           modClient,
		RedisClient:                redisClient,
		PermissionClient:           permClient,
		MasterDataClient:           mdClient,
		UserClient:                 userClient,
		FileClient:                 fileClient,
		KafkaProducer:              kafkaProducer,
		KafkaRescanner:             kafkaRescanner,
	}
}
