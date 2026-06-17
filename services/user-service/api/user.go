package main

import (
	"flag"
	"fmt"

	"github.com/guxiao1976/community-user/api/internal/config"
	"github.com/guxiao1976/community-user/api/internal/handler"
	"github.com/guxiao1976/community-user/api/internal/svc"

	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/user-api.yaml", "配置文件路径")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting User Service REST API at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
