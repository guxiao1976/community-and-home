package health

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/logic/health"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/svc"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := health.NewHealthLogic(r.Context(), svcCtx)
		resp, err := l.Health()
		responsex.Response(w, resp, err)
	}
}
