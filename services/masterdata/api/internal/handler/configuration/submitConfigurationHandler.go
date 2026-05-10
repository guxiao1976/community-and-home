// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package configuration

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/configuration"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Submit configuration
func SubmitConfigurationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := configuration.NewSubmitConfigurationLogic(r.Context(), svcCtx)
		resp, err := l.SubmitConfiguration(&req)
		if err != nil {
			responsex.Response(w, nil, err)
		} else {
			responsex.Response(w, resp, nil)
		}
	}
}
