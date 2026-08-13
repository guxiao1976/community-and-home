package user

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/api/internal/logic/user"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetAppStateHandler 读取当前用户的应用状态（当前小区）
func GetAppStateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetAppStateLogic(r.Context(), svcCtx)
		resp, err := l.GetAppState()
		responsex.Response(w, resp, err)
	}
}

// SetCurrentCommunityHandler 切换当前用户的小区
func SetCurrentCommunityHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetCurrentCommunityReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := user.NewSetCurrentCommunityLogic(r.Context(), svcCtx)
		resp, err := l.SetCurrentCommunity(&req)
		responsex.Response(w, resp, err)
	}
}
