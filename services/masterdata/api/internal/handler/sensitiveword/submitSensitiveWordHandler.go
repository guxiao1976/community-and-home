// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sensitiveword

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/sensitiveword"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Submit sensitive word
func SubmitSensitiveWordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sensitiveword.NewSubmitSensitiveWordLogic(r.Context(), svcCtx)
		resp, err := l.SubmitSensitiveWord(&req)
		if err != nil {
			responsex.Response(w, nil, err)
		} else {
			responsex.Response(w, resp, nil)
		}
	}
}
