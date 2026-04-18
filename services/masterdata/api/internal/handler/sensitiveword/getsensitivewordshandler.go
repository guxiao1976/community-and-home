package sensitiveword

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/sensitiveword"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
)

func GetSensitiveWordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSensitiveWordsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sensitiveword.NewGetSensitiveWordsLogic(r.Context(), svcCtx)
		resp, err := l.GetSensitiveWords(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
