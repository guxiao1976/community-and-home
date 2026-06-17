package contact

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/logic/contact"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListContactsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListContactsReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := contact.NewListContactsLogic(r.Context(), svcCtx)
		resp, err := l.ListContacts(&req)
		responsex.Response(w, resp, err)
	}
}

func UpsertContactsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpsertContactsReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := contact.NewUpsertContactsLogic(r.Context(), svcCtx)
		err := l.UpsertContacts(&req)
		responsex.Response(w, nil, err)
	}
}
