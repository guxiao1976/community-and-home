// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/logic/cost"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取预警记录
func ListAlertRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListAlertRecordsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := cost.NewListAlertRecordsLogic(r.Context(), svcCtx)
		resp, err := l.ListAlertRecords(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
