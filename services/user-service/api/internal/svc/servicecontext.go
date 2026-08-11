package svc

import (
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	permMiddleware "github.com/guxiao1976/community-common/v2/pkg/middleware"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-user/api/internal/config"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext REST API 服务上下文，持有 gRPC 客户端
type ServiceContext struct {
	Config         config.Config
	UserRpc        userv1.UserServiceClient
	PermClient     permissionv1.PermissionServiceClient
	PermMiddleware *permMiddleware.PermissionMiddleware
	SysConfig      *sysconfig.Client // 系统参数配置读取器
}

// NewServiceContext 创建服务上下文，初始化 gRPC 客户端
func NewServiceContext(c config.Config) *ServiceContext {
	userCli := zrpc.MustNewClient(c.UserRpc)
	permCli := zrpc.MustNewClient(c.PermissionRpc)
	permClient := permissionv1.NewPermissionServiceClient(permCli.Conn())

	// 初始化系统参数配置客户端
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", nil)

	return &ServiceContext{
		Config:         c,
		UserRpc:        userv1.NewUserServiceClient(userCli.Conn()),
		PermClient:     permClient,
		PermMiddleware: permMiddleware.NewPermissionMiddleware(permClient),
		SysConfig:      sysCfg,
	}
}
