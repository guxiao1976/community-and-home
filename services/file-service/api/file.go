package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/guxiao1976/community-file/api/internal/config"
	"github.com/guxiao1976/community-file/api/internal/handler"
	"github.com/guxiao1976/community-file/api/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"

	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/file-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	configx.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	// 注册 responsex 中间件（将 ResponseWriter 注入 context，支持 CtxSuccess/CtxError）
	server.Use(responsex.ResponseInterceptor)

	// CORS 中间件
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(w, r)
		}
	})

	ctx := svc.NewServiceContext(c)

	// 注册路由
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting File Service API server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
