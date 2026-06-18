package svc

import (
	"github.com/guxiao1976/community-permission/api/internal/config"
	"github.com/guxiao1976/community-permission/rpc/permission"
)

// ServiceContext 权限中心 API 服务上下文
type ServiceContext struct {
	Config        config.Config
	PermissionRpc permission.PermissionService // gRPC 客户端（通过 etcd 发现 permission.rpc）
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		PermissionRpc: permission.NewPermissionService(c.PermissionRpc),
	}
}
