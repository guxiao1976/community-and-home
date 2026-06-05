package main

import (
	"flag"
	"fmt"

	"github.com/guxiao1976/community-monitoring/api/internal/config"
	"github.com/guxiao1976/community-monitoring/api/internal/handler"
	"github.com/guxiao1976/community-monitoring/api/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/configx"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/monitoring-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting monitoring server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
