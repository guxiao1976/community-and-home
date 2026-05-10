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

// Restore a deleted item
func RestoreDeletedItemHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RestoreDeletedItemReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := deleteditems.NewRestoreDeletedItemLogic(r.Context(), svcCtx)
		resp, err := l.RestoreDeletedItem(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
