package svc

import (
	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/config"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext REST API 服务上下文，持有 gRPC 客户端
type ServiceContext struct {
	Config             config.Config
	NoticeServiceRpc    communityv1.NoticeServiceClient
	ContactServiceRpc   communityv1.ContactServiceClient
	LostFoundServiceRpc communityv1.LostFoundServiceClient
}

// NewServiceContext 创建服务上下文，初始化 gRPC 客户端
func NewServiceContext(c config.Config) *ServiceContext {
	cli := zrpc.MustNewClient(c.CommunityHubRpc)
	conn := cli.Conn()

	return &ServiceContext{
		Config:              c,
		NoticeServiceRpc:    communityv1.NewNoticeServiceClient(conn),
		ContactServiceRpc:   communityv1.NewContactServiceClient(conn),
		LostFoundServiceRpc: communityv1.NewLostFoundServiceClient(conn),
	}
}
