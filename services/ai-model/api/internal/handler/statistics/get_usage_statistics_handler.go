package statistics

import (
	"net/http"

	"community-and-home/services/ai-model/api/internal/logic/statistics"
	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取使用统计
func GetUsageStatisticsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUsageStatisticsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := statistics.NewGetUsageStatisticsLogic(r.Context(), svcCtx)
		resp, err := l.GetUsageStatistics(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
