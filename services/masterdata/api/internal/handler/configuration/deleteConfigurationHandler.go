// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package configuration

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/configuration"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Soft delete configuration
func DeleteConfigurationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteConfigurationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := configuration.NewDeleteConfigurationLogic(r.Context(), svcCtx)
		resp, err := l.DeleteConfiguration(&req)
		responsex.Response(w, resp, err)
	}
}
