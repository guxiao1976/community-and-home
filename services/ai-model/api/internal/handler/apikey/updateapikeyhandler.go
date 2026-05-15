// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package apikey

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/logic/apikey"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新 API Key
func UpdateAPIKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateAPIKeyRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := apikey.NewUpdateAPIKeyLogic(r.Context(), svcCtx)
		resp, err := l.UpdateAPIKey(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
