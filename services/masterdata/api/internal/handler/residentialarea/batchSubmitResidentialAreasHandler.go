// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package residentialarea

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/residentialarea"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Batch submit residential areas
func BatchSubmitResidentialAreasHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchSubmitReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := residentialarea.NewBatchSubmitResidentialAreasLogic(r.Context(), svcCtx)
		resp, err := l.BatchSubmitResidentialAreas(&req)
		if err != nil {
			responsex.Response(w, nil, err)
		} else {
			responsex.Response(w, resp, nil)
		}
	}
}
