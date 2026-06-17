package svc

import (
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-user/api/internal/config"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext REST API 服务上下文，持有 gRPC 客户端
type ServiceContext struct {
	Config    config.Config
	UserRpc   userv1.UserServiceClient
	SysConfig *sysconfig.Client // 系统参数配置读取器
}

// NewServiceContext 创建服务上下文，初始化 gRPC 客户端
func NewServiceContext(c config.Config) *ServiceContext {
	userCli := zrpc.MustNewClient(c.UserRpc)

	// 初始化系统参数配置客户端
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", nil)

	return &ServiceContext{
		Config:    c,
		UserRpc:   userv1.NewUserServiceClient(userCli.Conn()),
		SysConfig: sysCfg,
	}
}
