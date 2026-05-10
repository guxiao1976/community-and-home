package residentialarea

import (
	"net/http"
	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/residentialarea"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateResidentialAreaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateResidentialAreaReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := residentialarea.NewCreateResidentialAreaLogic(r.Context(), svcCtx)
		resp, err := l.CreateResidentialArea(&req)
		responsex.Response(w, resp, err)
	}
}
