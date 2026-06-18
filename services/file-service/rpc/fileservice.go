package main

import (
	"flag"
	"fmt"

	"github.com/guxiao1976/community-file/rpc/internal/config"
	"github.com/guxiao1976/community-file/rpc/internal/server"
	"github.com/guxiao1976/community-file/rpc/internal/svc"
	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/fileservice.yaml", "config file")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		filev1.RegisterFileServiceServer(grpcServer, server.NewFileServiceServer(ctx))
	})

	fmt.Printf("Starting File Service gRPC server at %s...\n", c.ListenOn)
	s.Start()
	defer s.Stop()
}
