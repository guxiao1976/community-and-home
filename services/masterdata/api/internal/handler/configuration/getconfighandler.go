package configuration

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/configuration"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
)

func GetConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetConfigReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := configuration.NewGetConfigLogic(r.Context(), svcCtx)
		resp, err := l.GetConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
