package main

import (
	"flag"
	"fmt"

	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/guxiao1976/community-user/rpc/internal/config"
	"github.com/guxiao1976/community-user/rpc/internal/server"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/userservice.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	// AES 加密在 ServiceContext 中初始化
	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		userv1.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))
	})

	fmt.Printf("Starting User Service gRPC server at %s...\n", c.ListenOn)
	s.Start()
	defer s.Stop()
}
