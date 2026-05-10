package approval

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/approval"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
)

func GetPendingCountsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := approval.NewGetPendingCountsLogic(r.Context(), svcCtx)
		resp, err := l.GetPendingCounts()
		responsex.Response(w, resp, err)
	}
}
