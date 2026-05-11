// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package statistics

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/statistics"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get division counts realtime statistics
func GetDivisionCountsRealtimeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DivisionCountsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := statistics.NewGetDivisionCountsRealtimeLogic(r.Context(), svcCtx)
		resp, err := l.GetDivisionCountsRealtime(&req)
		responsex.Response(w, resp, err)
	}
}
