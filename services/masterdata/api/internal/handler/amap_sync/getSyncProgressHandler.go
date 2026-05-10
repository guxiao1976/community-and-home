package amap_sync

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/amap_sync"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetSyncProgressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSyncProgressReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		l := amap_sync.NewGetSyncProgressLogic(r.Context(), svcCtx)
		resp, err := l.GetSyncProgress(&req)
		responsex.Response(w, resp, err)
	}
}
