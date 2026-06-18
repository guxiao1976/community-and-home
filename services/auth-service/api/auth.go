package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/guxiao1976/community-auth/api/internal/config"
	"github.com/guxiao1976/community-auth/api/internal/handler"
	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/auth-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	// 初始化 RSA 密钥对（供 /api/auth/public-key 端点使用）
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

	// 创建服务上下文（gRPC 客户端 + Redis 客户端）
	ctx := svc.NewServiceContext(c)

	// 创建 REST Server
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 注册路由
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting auth-api server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
