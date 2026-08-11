package svc

import (
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	permMiddleware "github.com/guxiao1976/community-common/v2/pkg/middleware"
	"github.com/guxiao1976/community-permission/api/internal/config"
	"github.com/guxiao1976/community-permission/rpc/permission"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 权限中心 API 服务上下文
type ServiceContext struct {
	Config         config.Config
	PermissionRpc  permission.PermissionService         // gRPC 客户端
	PermMiddleware *permMiddleware.PermissionMiddleware // 权限校验中间件
	RouteRegistry  *RouteRegistry                       // 路由注册器（auto-discover）
	DB             sqlx.SqlConn                         // MySQL 连接
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c config.Config) *ServiceContext {
	permCli := zrpc.MustNewClient(c.PermissionRpc)
	permClient := permissionv1.NewPermissionServiceClient(permCli.Conn())

	return &ServiceContext{
		Config:         c,
		PermissionRpc:  permission.NewPermissionService(c.PermissionRpc),
		PermMiddleware: permMiddleware.NewPermissionMiddleware(permClient),
		RouteRegistry:  NewRouteRegistry(),
		DB:             sqlx.NewMysql(c.DataSource),
	}
}
