package community

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/api/internal/logic/community"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func JoinCommunityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.JoinCommunityReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := community.NewJoinCommunityLogic(r.Context(), svcCtx)
		resp, err := l.JoinCommunity(&req)
		responsex.Response(w, resp, err)
	}
}

func LeaveCommunityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LeaveCommunityReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := community.NewLeaveCommunityLogic(r.Context(), svcCtx)
		resp, err := l.LeaveCommunity(&req)
		responsex.Response(w, resp, err)
	}
}

func GetMembershipsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := community.NewGetMembershipsLogic(r.Context(), svcCtx)
		resp, err := l.GetMemberships()
		responsex.Response(w, resp, err)
	}
}

func BindResidenceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindResidenceReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := community.NewBindResidenceLogic(r.Context(), svcCtx)
		resp, err := l.BindResidence(&req)
		responsex.Response(w, resp, err)
	}
}

func ApplyRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApplyRoleReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := community.NewApplyRoleLogic(r.Context(), svcCtx)
		resp, err := l.ApplyRole(&req)
		responsex.Response(w, resp, err)
	}
}

func GetUserRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := community.NewGetUserRolesLogic(r.Context(), svcCtx)
		resp, err := l.GetUserRoles()
		responsex.Response(w, resp, err)
	}
}
