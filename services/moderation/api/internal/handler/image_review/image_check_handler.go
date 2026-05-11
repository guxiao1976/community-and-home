package image_review

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/logic/image_review"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/svc"
)

func ImageCheckHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := image_review.NewImageCheckLogic(r.Context(), svcCtx)
		resp, err := l.ImageCheck(r)
		responsex.Response(w, resp, err)
	}
}
