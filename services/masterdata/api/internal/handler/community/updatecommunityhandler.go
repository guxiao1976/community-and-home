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

// Update community
func UpdateCommunityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateCommunityReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := community.NewUpdateCommunityLogic(r.Context(), svcCtx)
		resp, err := l.UpdateCommunity(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
