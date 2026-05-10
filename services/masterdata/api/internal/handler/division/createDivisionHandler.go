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

// Create new administrative division
func CreateDivisionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateDivisionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := division.NewCreateDivisionLogic(r.Context(), svcCtx)
		resp, err := l.CreateDivision(&req)
		responsex.Response(w, resp, err)
	}
}
