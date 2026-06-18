package main

import (
	"flag"
	"fmt"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/rpc/internal/config"
	"github.com/guxiao1976/community-permission/rpc/internal/server"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/permissionservice.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		permissionv1.RegisterPermissionServiceServer(grpcServer, server.NewPermissionServiceServer(ctx))
	})

	fmt.Printf("Starting Permission Service gRPC server at %s...\n", c.ListenOn)
	s.Start()
	defer s.Stop()
}
