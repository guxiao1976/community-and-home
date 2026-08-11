// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/logic/perm"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	routes := []rest.Route{
		{
			// List roles (paginated)
			Method:  http.MethodGet,
			Path:    "/roles",
			Handler: listRolesHandler(serverCtx),
		},
		{
			// Create role (with permission_ids)
			Method:  http.MethodPost,
			Path:    "/roles",
			Handler: createRoleHandler(serverCtx),
		},
		{
			// Get role details (with permissions)
			Method:  http.MethodGet,
			Path:    "/roles/:id",
			Handler: getRoleHandler(serverCtx),
		},
		{
			// Update role
			Method:  http.MethodPut,
			Path:    "/roles/:id",
			Handler: updateRoleHandler(serverCtx),
		},
		{
			// Delete role
			Method:  http.MethodDelete,
			Path:    "/roles/:id",
			Handler: deleteRoleHandler(serverCtx),
		},
		{
			// List users by role
			Method:  http.MethodGet,
			Path:    "/roles/:id/users",
			Handler: listRoleUsersHandler(serverCtx),
		},
		{
			// List permissions (tree structure)
			Method:  http.MethodGet,
			Path:    "/permissions",
			Handler: listPermissionsHandler(serverCtx),
		},
		{
			// Get role permissions
			Method:  http.MethodGet,
			Path:    "/roles/:id/permissions",
			Handler: getRolePermissionsHandler(serverCtx),
		},
		{
			// Assign permissions to role
			Method:  http.MethodPost,
			Path:    "/roles/:id/permissions",
			Handler: assignRolePermissionsHandler(serverCtx),
		},
		{
			// Assign role to user
			Method:  http.MethodPost,
			Path:    "/user-roles",
			Handler: assignUserRoleHandler(serverCtx),
		},
		{
			// Revoke role from user
			Method:  http.MethodDelete,
			Path:    "/user-roles",
			Handler: revokeUserRoleHandler(serverCtx),
		},
		{
			// Get user roles
			Method:  http.MethodGet,
			Path:    "/users/:userId/roles",
			Handler: getUserRolesHandler(serverCtx),
		},
		{
			// Get user permissions
			Method:  http.MethodGet,
			Path:    "/users/:userId/permissions",
			Handler: getUserPermissionsHandler(serverCtx),
		},
		{
			// Auto-discover permissions
			Method:  http.MethodPost,
			Path:    "/permissions/auto-discover",
			Handler: autoDiscoverPermissionsHandler(serverCtx),
		},
	}

	// 注册路由到 RouteRegistry，供 auto-discover 使用
	prefix := "/api/perm"
	serverCtx.RouteRegistry.Register("GET", prefix+"/roles", "List roles")
	serverCtx.RouteRegistry.Register("POST", prefix+"/roles", "Create role")
	serverCtx.RouteRegistry.Register("GET", prefix+"/roles/:id", "Get role detail")
	serverCtx.RouteRegistry.Register("PUT", prefix+"/roles/:id", "Update role")
	serverCtx.RouteRegistry.Register("DELETE", prefix+"/roles/:id", "Delete role")
	serverCtx.RouteRegistry.Register("GET", prefix+"/roles/:id/users", "List users by role")
	serverCtx.RouteRegistry.Register("GET", prefix+"/permissions", "List permissions")
	serverCtx.RouteRegistry.Register("GET", prefix+"/roles/:id/permissions", "Get role permissions")
	serverCtx.RouteRegistry.Register("POST", prefix+"/roles/:id/permissions", "Assign permissions to role")
	serverCtx.RouteRegistry.Register("POST", prefix+"/user-roles", "Assign role to user")
	serverCtx.RouteRegistry.Register("DELETE", prefix+"/user-roles", "Revoke role from user")
	serverCtx.RouteRegistry.Register("GET", prefix+"/users/:userId/roles", "Get user roles")
	serverCtx.RouteRegistry.Register("GET", prefix+"/users/:userId/permissions", "Get user permissions")
	serverCtx.RouteRegistry.Register("POST", prefix+"/permissions/auto-discover", "Auto-discover API permissions")

	// 应用权限校验中间件
	routes = rest.WithMiddleware(serverCtx.PermMiddleware.Handle, routes...)

	server.AddRoutes(routes,
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/perm"),
	)
}

// ==================== Role Handlers ====================

// listRolesHandler GET /api/perm/roles
func listRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListRolesReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		l := perm.NewListRolesLogic(r.Context(), svcCtx)
		resp, err := l.ListRoles(&req)
		responsex.Response(w, resp, err)
	}
}

// createRoleHandler POST /api/perm/roles
func createRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		l := perm.NewCreateRoleLogic(r.Context(), svcCtx)
		resp, err := l.CreateRole(&req)
		responsex.Response(w, resp, err)
	}
}

// getRoleHandler GET /api/perm/roles/:id
func getRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		l := perm.NewGetRoleLogic(r.Context(), svcCtx)
		resp, err := l.GetRole(&req)
		responsex.Response(w, resp, err)
	}
}

// updateRoleHandler PUT /api/perm/roles/:id
func updateRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		l := perm.NewUpdateRoleLogic(r.Context(), svcCtx)
		resp, err := l.UpdateRole(&req)
		responsex.Response(w, resp, err)
	}
}

// deleteRoleHandler DELETE /api/perm/roles/:id
func deleteRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		l := perm.NewDeleteRoleLogic(r.Context(), svcCtx)
		err := l.DeleteRole(&req)
		responsex.Response(w, nil, err)
	}
}

// ==================== Permission Handlers ====================

// listPermissionsHandler GET /api/perm/permissions
func listPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListPermissionsReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		l := perm.NewListPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.ListPermissions(&req)
		responsex.Response(w, resp, err)
	}
}
