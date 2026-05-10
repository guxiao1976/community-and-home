// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package property

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/identity/api/internal/logic/property"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Unbind property from user
func UnbindPropertyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UnbindPropertyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := property.NewUnbindPropertyLogic(r.Context(), svcCtx)
		resp, err := l.UnbindProperty(&req)
		responsex.Response(w, resp, err)
	}
}
