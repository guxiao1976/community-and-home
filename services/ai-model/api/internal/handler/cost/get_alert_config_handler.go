// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"net/http"

	"community-and-home/services/ai-model/api/internal/logic/cost"
	"community-and-home/services/ai-model/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取预警配置
func GetAlertConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := cost.NewGetAlertConfigLogic(r.Context(), svcCtx)
		resp, err := l.GetAlertConfig()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
