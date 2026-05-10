// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sensitiveword

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/sensitiveword"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Create new sensitive word
func CreateSensitiveWordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateSensitiveWordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := sensitiveword.NewCreateSensitiveWordLogic(r.Context(), svcCtx)
		resp, err := l.CreateSensitiveWord(&req)
		responsex.Response(w, resp, err)
	}
}
