package notice

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/logic/notice"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateNoticeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateNoticeReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewCreateNoticeLogic(r.Context(), svcCtx)
		resp, err := l.CreateNotice(&req)
		responsex.Response(w, resp, err)
	}
}

func ListNoticesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListNoticesReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewListNoticesLogic(r.Context(), svcCtx)
		resp, err := l.ListNotices(&req)
		responsex.Response(w, resp, err)
	}
}

func GetNoticeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetNoticeReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewGetNoticeLogic(r.Context(), svcCtx)
		resp, err := l.GetNotice(&req)
		responsex.Response(w, resp, err)
	}
}

func UpdateNoticeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateNoticeReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewUpdateNoticeLogic(r.Context(), svcCtx)
		err := l.UpdateNotice(&req)
		responsex.Response(w, nil, err)
	}
}

func DeleteNoticeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteNoticeReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewDeleteNoticeLogic(r.Context(), svcCtx)
		err := l.DeleteNotice(&req)
		responsex.Response(w, nil, err)
	}
}
