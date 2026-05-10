// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/division"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Batch submit divisions
func BatchSubmitDivisionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchSubmitReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := division.NewBatchSubmitDivisionsLogic(r.Context(), svcCtx)
		resp, err := l.BatchSubmitDivisions(&req)
		if err != nil {
			responsex.Response(w, nil, err)
		} else {
			responsex.Response(w, resp, nil)
		}
	}
}
