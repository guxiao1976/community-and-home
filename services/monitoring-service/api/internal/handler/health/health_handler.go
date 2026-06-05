package health

import (
	"net/http"

	"github.com/guxiao1976/community-monitoring/api/internal/logic/health"
	"github.com/guxiao1976/community-monitoring/api/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := health.NewHealthLogic(svcCtx.Config.Monitoring)
		resp, err := l.CheckHealth(r.Context())
		if err != nil {
			responsex.Error(w, err)
			return
		}
		responsex.Success(w, resp)
	}
}
