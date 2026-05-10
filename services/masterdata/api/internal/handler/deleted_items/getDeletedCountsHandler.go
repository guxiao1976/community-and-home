// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package deleteditems

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/deleted_items"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
)

// Get deleted items counts per entity type
func GetDeletedCountsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := deleteditems.NewGetDeletedCountsLogic(r.Context(), svcCtx)
		resp, err := l.GetDeletedCounts()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
