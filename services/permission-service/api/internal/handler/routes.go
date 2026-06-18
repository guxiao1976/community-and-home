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
	server.AddRoutes(
		[]rest.Route{
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
				// List permissions (tree structure)
				Method:  http.MethodGet,
				Path:    "/permissions",
				Handler: listPermissionsHandler(serverCtx),
			},
		},
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
