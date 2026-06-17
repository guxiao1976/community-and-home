package main

import (
	"flag"
	"fmt"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/guxiao1976/community-hub/rpc/internal/config"
	"github.com/guxiao1976/community-hub/rpc/internal/server"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/communityhub.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		communityv1.RegisterNoticeServiceServer(grpcServer, server.NewNoticeServiceServer(ctx))
		communityv1.RegisterContactServiceServer(grpcServer, server.NewContactServiceServer(ctx))
		communityv1.RegisterLostFoundServiceServer(grpcServer, server.NewLostFoundServiceServer(ctx))
	})

	fmt.Printf("Starting Community Hub Service gRPC server at %s...\n", c.ListenOn)
	s.Start()
	defer s.Stop()
}
