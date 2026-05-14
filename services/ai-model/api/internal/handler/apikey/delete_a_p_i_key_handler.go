// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"net/http"

	"community-and-home/services/ai-model/api/internal/logic/apikey"
	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除 API Key
func DeleteAPIKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteAPIKeyRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := apikey.NewDeleteAPIKeyLogic(r.Context(), svcCtx)
		resp, err := l.DeleteAPIKey(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
