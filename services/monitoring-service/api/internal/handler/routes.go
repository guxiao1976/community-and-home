package handler

import (
	"net/http"

	"github.com/guxiao1976/community-monitoring/api/internal/handler/health"
	"github.com/guxiao1976/community-monitoring/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/health",
				Handler: health.HealthHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/monitoring"),
	)
}
