// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package template

import (
	"net/http"

	"community-and-home/services/ai-model/api/internal/logic/template"
	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 创建提示词模板
func CreateTemplateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateTemplateRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := template.NewCreateTemplateLogic(r.Context(), svcCtx)
		resp, err := l.CreateTemplate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
