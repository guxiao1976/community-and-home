package handler

import (
	"net/http"

	"github.com/guxiao1976/community-user/api/internal/handler/community"
	"github.com/guxiao1976/community-user/api/internal/handler/user"
	"github.com/guxiao1976/community-user/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册所有 HTTP 路由
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/users",
				Handler: user.ListUsersHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/users",
				Handler: user.CreateUserHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/users/:id",
				Handler: user.GetUserHandler(serverCtx),
			},
			{
				Method:  http.MethodPut,
				Path:    "/api/users/:id",
				Handler: user.UpdateUserHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/api/users/:id",
				Handler: user.DeleteUserHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)

	// 社区成员关系（需 JWT 认证）
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/users/communities/join",
				Handler: community.JoinCommunityHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/users/communities/memberships",
				Handler: community.GetMembershipsHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/users/communities/leave",
				Handler: community.LeaveCommunityHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)
}
