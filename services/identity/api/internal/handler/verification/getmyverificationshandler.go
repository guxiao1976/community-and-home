// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"net/http"

	"community-and-home/services/identity/api/internal/logic/verification"
	"community-and-home/services/identity/api/internal/svc"
	"community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get my verification requests
func GetMyVerificationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyVerificationsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := verification.NewGetMyVerificationsLogic(r.Context(), svcCtx)
		resp, err := l.GetMyVerifications(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
