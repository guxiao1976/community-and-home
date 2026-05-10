// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/division"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Request delete for approved division
func RequestDeleteDivisionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := division.NewRequestDeleteDivisionLogic(r.Context(), svcCtx)
		resp, err := l.RequestDeleteDivision(&req)
		responsex.Response(w, resp, err)
	}
}
