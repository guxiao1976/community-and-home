package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/guxiao/community-and-home/services/moderation/api/internal/config"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/handler"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/moderation-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

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

	// Load word store on startup
	if err := ctx.WordStore.Load(context.Background()); err != nil {
		logx.Errorf("failed to load word store: %v", err)
		fmt.Printf("WARNING: word store load failed: %v\n", err)
	} else {
		logx.Info("word store loaded successfully")
	}

	// Start background sync
	ctx.WordStore.StartSync(context.Background())

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
