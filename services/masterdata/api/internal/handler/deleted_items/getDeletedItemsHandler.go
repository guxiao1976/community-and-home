// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package deleteditems

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/deleted_items"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

)

// Get deleted items list with optional entity type filter
func GetDeletedItemsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetDeletedItemsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := deleteditems.NewGetDeletedItemsLogic(r.Context(), svcCtx)
		resp, err := l.GetDeletedItems(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
