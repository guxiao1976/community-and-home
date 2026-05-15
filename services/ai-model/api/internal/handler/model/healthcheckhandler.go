// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/logic/model"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 健康检查
func HealthCheckHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := model.NewHealthCheckLogic(r.Context(), svcCtx)
		resp, err := l.HealthCheck()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
