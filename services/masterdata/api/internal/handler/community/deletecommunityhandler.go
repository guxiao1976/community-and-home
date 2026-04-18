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

// Soft delete community
func DeleteCommunityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteCommunityReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := community.NewDeleteCommunityLogic(r.Context(), svcCtx)
		resp, err := l.DeleteCommunity(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
