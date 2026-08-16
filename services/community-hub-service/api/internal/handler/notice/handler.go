package notice

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/logic/notice"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateContentPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateContentPostReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewCreateContentPostLogic(r.Context(), svcCtx)
		resp, err := l.CreateContentPost(&req)
		responsex.Response(w, resp, err)
	}
}

func ListContentPostsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListContentPostsReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewListContentPostsLogic(r.Context(), svcCtx)
		resp, err := l.ListContentPosts(&req)
		responsex.Response(w, resp, err)
	}
}

func GetContentPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetContentPostReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewGetContentPostLogic(r.Context(), svcCtx)
		resp, err := l.GetContentPost(&req)
		responsex.Response(w, resp, err)
	}
}

func UpdateContentPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateContentPostReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewUpdateContentPostLogic(r.Context(), svcCtx)
		err := l.UpdateContentPost(&req)
		responsex.Response(w, nil, err)
	}
}

func DeleteContentPostHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteContentPostReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewDeleteContentPostLogic(r.Context(), svcCtx)
		err := l.DeleteContentPost(&req)
		responsex.Response(w, nil, err)
	}
}

func GetMarqueeNoticesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMarqueeNoticesReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := notice.NewGetMarqueeNoticesLogic(r.Context(), svcCtx)
		resp, err := l.GetMarqueeNotices(&req)
		responsex.Response(w, resp, err)
	}
}

func GetPublishPermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPublishPermissionReq
		l := notice.NewGetPublishPermissionLogic(r.Context(), svcCtx)
		resp, err := l.GetPublishPermission(&req)
		responsex.Response(w, resp, err)
	}
}
