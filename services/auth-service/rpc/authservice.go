package main

import (
	"flag"
	"fmt"
	"os"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-auth/rpc/internal/config"
	"github.com/guxiao1976/community-auth/rpc/internal/server"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/authservice.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	// 初始化加密模块
	if c.AesKey != "" {
		if err := crypto.InitAES(c.AesKey); err != nil {
			panic(fmt.Sprintf("InitAES failed: %v", err))
		}
	}
	if c.RsaPublicKey != "" {
		privateKey := c.RsaPrivateKey
		if c.RsaPrivateKeyPath != "" {
			keyBytes, err := os.ReadFile(c.RsaPrivateKeyPath)
			if err != nil {
				panic(fmt.Sprintf("failed to read RSA private key from %s: %v", c.RsaPrivateKeyPath, err))
			}
			privateKey = string(keyBytes)
		}
		if privateKey != "" {
			if err := crypto.InitRSA(c.RsaPublicKey, privateKey); err != nil {
				panic(fmt.Sprintf("InitRSA failed: %v", err))
			}
		}
	}

	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		authv1.RegisterAuthServiceServer(grpcServer, server.NewAuthServiceServer(ctx))
	})

	fmt.Printf("Starting Auth Service gRPC server at %s...\n", c.ListenOn)
	s.Start()
	defer s.Stop()
}
