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

// 获取 API Key 详情
func GetAPIKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAPIKeyRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := apikey.NewGetAPIKeyLogic(r.Context(), svcCtx)
		resp, err := l.GetAPIKey(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
