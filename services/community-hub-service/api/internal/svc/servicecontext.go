package svc

import (
	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	permMiddleware "github.com/guxiao1976/community-common/v2/pkg/middleware"
	"github.com/guxiao1976/community-hub/api/internal/config"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext REST API 服务上下文，持有 gRPC 客户端
type ServiceContext struct {
	Config              config.Config
	NoticeServiceRpc    communityv1.NoticeServiceClient
	ContactServiceRpc   communityv1.ContactServiceClient
	LostFoundServiceRpc communityv1.LostFoundServiceClient
	PermClient          permissionv1.PermissionServiceClient
	PermMiddleware      *permMiddleware.PermissionMiddleware
}

// NewServiceContext 创建服务上下文，初始化 gRPC 客户端
func NewServiceContext(c config.Config) *ServiceContext {
	cli := zrpc.MustNewClient(c.CommunityHubRpc)
	conn := cli.Conn()

	permCli := zrpc.MustNewClient(c.PermissionRpc)
	permClient := permissionv1.NewPermissionServiceClient(permCli.Conn())
	permMW := permMiddleware.NewPermissionMiddleware(permClient)

	return &ServiceContext{
		Config:              c,
		NoticeServiceRpc:    communityv1.NewNoticeServiceClient(conn),
		ContactServiceRpc:   communityv1.NewContactServiceClient(conn),
		LostFoundServiceRpc: communityv1.NewLostFoundServiceClient(conn),
		PermClient:          permClient,
		PermMiddleware:      permMW,
	}
}
