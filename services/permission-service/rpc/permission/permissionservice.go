package permission

import (
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/zeromicro/go-zero/zrpc"
)

// PermissionService 权限中心 gRPC 客户端接口
type PermissionService interface {
	permissionv1.PermissionServiceClient
}

type defaultPermissionService struct {
	permissionv1.PermissionServiceClient
}

// NewPermissionService 创建权限中心 gRPC 客户端
func NewPermissionService(c zrpc.RpcClientConf) PermissionService {
	cli := zrpc.MustNewClient(c)
	return &defaultPermissionService{
		PermissionServiceClient: permissionv1.NewPermissionServiceClient(cli.Conn()),
	}
}
