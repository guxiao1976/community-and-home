// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package submissionrecord

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/logic/submissionrecord"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get reviewed submission records
func GetReviewedSubmissionRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSubmissionRecordsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := submissionrecord.NewGetReviewedSubmissionRecordsLogic(r.Context(), svcCtx)
		resp, err := l.GetReviewedSubmissionRecords(&req)
		responsex.Response(w, resp, err)
	}
}
