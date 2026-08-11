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
	routes := rest.WithMiddleware(serverCtx.PermMiddleware.Handle, []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/users/profile",
			Handler: user.GetProfileHandler(serverCtx),
		},
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
	}...)
	server.AddRoutes(routes,
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

	// 房产绑定（需 JWT 认证）
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/users/residences/bind",
				Handler: community.BindResidenceHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)

	// 角色管理（需 JWT 认证）
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/users/roles/apply",
				Handler: community.ApplyRoleHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/users/roles",
				Handler: community.GetUserRolesHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/users/certifications",
				Handler: community.SubmitCertificationHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/users/certifications",
				Handler: community.GetMyCertificationsHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)

	// 管理员认证审核（审核中心用）
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/verifications",
				Handler: community.ListCertificationsHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/verifications/:id/review",
				Handler: community.ReviewCertificationHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)
}
