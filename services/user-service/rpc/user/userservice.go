package user

import (
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/zeromicro/go-zero/zrpc"
)

// UserService 用户中心 gRPC 客户端接口
type UserService interface {
	userv1.UserServiceClient
}

// defaultUserService 默认实现，封装 zrpc.Client
type defaultUserService struct {
	userv1.UserServiceClient
}

// NewUserService 创建用户中心 gRPC 客户端
func NewUserService(c zrpc.RpcClientConf) UserService {
	cli := zrpc.MustNewClient(c)
	return &defaultUserService{
		UserServiceClient: userv1.NewUserServiceClient(cli.Conn()),
	}
}
