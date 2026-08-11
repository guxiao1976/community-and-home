package handler

import (
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/logic/perm"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func listRoleUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListRoleUsersReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := perm.NewListRoleUsersLogic(r.Context(), svcCtx)
		resp, err := l.ListRoleUsers(&req)
		responsex.Response(w, resp, err)
	}
}

func getRolePermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := perm.NewGetRolePermissionsLogic(r.Context(), svcCtx)
		resp, err := l.GetRolePermissions(&req)
		responsex.Response(w, resp, err)
	}
}

func assignRolePermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Id            int64            `path:"id"`
			PermissionIds types.Int64Array `json:"permissionIds"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := perm.NewAssignRolePermissionsLogic(r.Context(), svcCtx)
		err := l.AssignRolePermissions(&types.AssignRolePermissionsReq{PermissionIds: req.PermissionIds}, req.Id)
		responsex.Response(w, nil, err)
	}
}

func assignUserRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssignUserRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := perm.NewAssignUserRoleLogic(r.Context(), svcCtx)
		err := l.AssignUserRole(&req)
		responsex.Response(w, nil, err)
	}
}

func revokeUserRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RevokeUserRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := perm.NewRevokeUserRoleLogic(r.Context(), svcCtx)
		err := l.RevokeUserRole(&req)
		responsex.Response(w, nil, err)
	}
}

func getUserRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserRolesReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := perm.NewGetUserRolesLogic(r.Context(), svcCtx)
		resp, err := l.GetUserRoles(&req)
		responsex.Response(w, resp, err)
	}
}

func getUserPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserPermissionsReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := perm.NewGetUserPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.GetUserPermissions(&req)
		responsex.Response(w, resp, err)
	}
}

func autoDiscoverPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := perm.NewAutoDiscoverPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.AutoDiscover()
		responsex.Response(w, resp, err)
	}
}
