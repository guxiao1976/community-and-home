// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package verification

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/identity/api/internal/logic/verification"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Submit homeowner verification
func SubmitVerificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitVerificationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := verification.NewSubmitVerificationLogic(r.Context(), svcCtx)
		resp, err := l.SubmitVerification(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
