package handler

import (
	"net/http"

	"github.com/guxiao1976/community-hub/api/internal/handler/contact"
	"github.com/guxiao1976/community-hub/api/internal/handler/lostfound"
	"github.com/guxiao1976/community-hub/api/internal/handler/notice"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册所有 HTTP 路由
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.PermMiddleware.Handle},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/api/community/notices",
					Handler: notice.CreateNoticeHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/notices",
					Handler: notice.ListNoticesHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/notices/:id",
					Handler: notice.GetNoticeHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/api/community/notices/:id",
					Handler: notice.UpdateNoticeHandler(serverCtx),
				},
				{
					Method:  http.MethodDelete,
					Path:    "/api/community/notices/:id",
					Handler: notice.DeleteNoticeHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/contacts",
					Handler: contact.ListContactsHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/api/community/contacts",
					Handler: contact.UpsertContactsHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/api/community/lostfound",
					Handler: lostfound.CreateLostFoundHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/lostfound",
					Handler: lostfound.ListLostFoundHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/lostfound/:id",
					Handler: lostfound.GetLostFoundHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/api/community/lostfound/:id/resolve",
					Handler: lostfound.ResolveLostFoundHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)
}
