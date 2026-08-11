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

func SubmitCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitCertificationReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := community.NewSubmitCertificationLogic(r.Context(), svcCtx)
		resp, err := l.SubmitCertification(&req)
		responsex.Response(w, resp, err)
	}
}

func ReviewCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Id          int64  `path:"id"`
			Result      int32  `json:"result"`
			ReviewNotes string `json:"review_notes,optional"`
			ExpiresAt   string `json:"expires_at,optional"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := community.NewReviewCertificationLogic(r.Context(), svcCtx)
		resp, err := l.ReviewCertification(req.Id, &types.ReviewCertificationReq{
			Result:      req.Result,
			ReviewNotes: req.ReviewNotes,
			ExpiresAt:   req.ExpiresAt,
		})
		responsex.Response(w, resp, err)
	}
}

func ListCertificationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListCertificationsReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := community.NewListCertificationsLogic(r.Context(), svcCtx)
		resp, err := l.ListCertifications(&req)
		responsex.Response(w, resp, err)
	}
}

func GetMyCertificationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := community.NewGetMyCertificationsLogic(r.Context(), svcCtx)
		resp, err := l.GetMyCertifications()
		responsex.Response(w, resp, err)
	}
}
