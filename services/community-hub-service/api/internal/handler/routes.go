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
					Handler: notice.CreateContentPostHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/notices",
					Handler: notice.ListContentPostsHandler(serverCtx),
				},
				// 静态路径先于 :id 注册（防被 :id 通配吞掉）
				{
					Method:  http.MethodGet,
					Path:    "/api/community/notices/marquee",
					Handler: notice.GetMarqueeNoticesHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/notices/publish-permission",
					Handler: notice.GetPublishPermissionHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/api/community/notices/:id",
					Handler: notice.GetContentPostHandler(serverCtx),
				},
				{
					Method:  http.MethodPut,
					Path:    "/api/community/notices/:id",
					Handler: notice.UpdateContentPostHandler(serverCtx),
				},
				{
					Method:  http.MethodDelete,
					Path:    "/api/community/notices/:id",
					Handler: notice.DeleteContentPostHandler(serverCtx),
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
