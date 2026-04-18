// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/identity/api/internal/logic/property"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get my properties
func GetMyPropertiesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyPropertiesReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := property.NewGetMyPropertiesLogic(r.Context(), svcCtx)
		resp, err := l.GetMyProperties(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
