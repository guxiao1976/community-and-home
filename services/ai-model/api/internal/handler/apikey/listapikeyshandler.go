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

// 获取 API Key 列表
func ListAPIKeysHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAPIKeysRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := apikey.NewListAPIKeysLogic(r.Context(), svcCtx)
		resp, err := l.ListAPIKeys(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
