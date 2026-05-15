// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/guxiao/community-and-home/services/identity/api/internal/config"
	"github.com/guxiao/community-and-home/services/identity/api/internal/handler"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/identity-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	// Add CORS middleware
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(w, r)
		}
	})

	ctx := svc.NewServiceContext(c)

	// Note: PermissionMiddleware is not registered globally to avoid blocking public endpoints like login.
	// Permission checks are handled by route-level JWT middleware (rest.WithJwt) and Casbin policies.

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
