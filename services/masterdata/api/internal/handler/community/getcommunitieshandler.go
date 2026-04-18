// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package community

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/community"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// List communities with filters
func GetCommunitiesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetCommunitiesReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := community.NewGetCommunitiesLogic(r.Context(), svcCtx)
		resp, err := l.GetCommunities(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
