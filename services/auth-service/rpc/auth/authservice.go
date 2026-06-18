package auth

import (
	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/zeromicro/go-zero/zrpc"
)

// AuthService 认证中心 gRPC 客户端接口
type AuthService interface {
	authv1.AuthServiceClient
}

// defaultAuthService 默认实现
type defaultAuthService struct {
	authv1.AuthServiceClient
}

// NewAuthService 创建认证中心 gRPC 客户端
func NewAuthService(c zrpc.RpcClientConf) AuthService {
	cli := zrpc.MustNewClient(c)
	return &defaultAuthService{
		AuthServiceClient: authv1.NewAuthServiceClient(cli.Conn()),
	}
}
