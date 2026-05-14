// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"net/http"

	"community-and-home/services/ai-model/api/internal/logic/model"
	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取模型详情
func GetModelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetModelRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := model.NewGetModelLogic(r.Context(), svcCtx)
		resp, err := l.GetModel(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
