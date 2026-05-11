package content_review

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/logic/content_review"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/svc"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SubmitReviewHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubmitReviewReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := content_review.NewSubmitReviewLogic(r.Context(), svcCtx)
		resp, err := l.SubmitReview(&req)
		responsex.Response(w, resp, err)
	}
}
