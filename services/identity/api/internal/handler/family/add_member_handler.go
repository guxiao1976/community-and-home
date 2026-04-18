package family

import (
	"net/http"

	"github.com/guxiao/community-and-home/services/identity/api/internal/logic/family"
	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AddFamilyMemberHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddFamilyMemberReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := family.NewAddFamilyMemberLogic(r.Context(), svcCtx)
		resp, err := l.AddFamilyMember(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
