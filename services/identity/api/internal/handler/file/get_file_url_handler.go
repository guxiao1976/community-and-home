// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package file

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/identity/api/internal/logic/file"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get file URL
func GetFileUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetFileUrlReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := file.NewGetFileUrlLogic(r.Context(), svcCtx)
		resp, err := l.GetFileUrl(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
