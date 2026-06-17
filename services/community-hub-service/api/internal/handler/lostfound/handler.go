package lostfound

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/logic/lostfound"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateLostFoundHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateLostFoundReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := lostfound.NewCreateLostFoundLogic(r.Context(), svcCtx)
		resp, err := l.CreateLostFound(&req)
		responsex.Response(w, resp, err)
	}
}

func ListLostFoundHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListLostFoundReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := lostfound.NewListLostFoundLogic(r.Context(), svcCtx)
		resp, err := l.ListLostFound(&req)
		responsex.Response(w, resp, err)
	}
}

func GetLostFoundHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetLostFoundReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := lostfound.NewGetLostFoundLogic(r.Context(), svcCtx)
		resp, err := l.GetLostFound(&req)
		responsex.Response(w, resp, err)
	}
}

func ResolveLostFoundHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ResolveLostFoundReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := lostfound.NewResolveLostFoundLogic(r.Context(), svcCtx)
		err := l.ResolveLostFound(&req)
		responsex.Response(w, nil, err)
	}
}
