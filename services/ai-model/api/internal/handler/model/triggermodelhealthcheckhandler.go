// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/logic/model"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 触发模型健康检查
func TriggerModelHealthCheckHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TriggerHealthCheckRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := model.NewTriggerModelHealthCheckLogic(r.Context(), svcCtx)
		resp, err := l.TriggerModelHealthCheck(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
