package submissionrecord

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReviewedSubmissionRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetReviewedSubmissionRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReviewedSubmissionRecordsLogic {
	return &GetReviewedSubmissionRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReviewedSubmissionRecordsLogic) GetReviewedSubmissionRecords(req *types.GetSubmissionRecordsReq) (resp *types.GetSubmissionRecordsResp, err error) {
	userId := int64(0)
	if uid := l.ctx.Value("userId"); uid != nil {
		userId = uid.(int64)
	}

	var entityType *string
	if req.EntityType != "" {
		et := req.EntityType
		entityType = &et
	}
	var reviewResult *int64
	if req.ReviewResult != nil {
		rr := int64(*req.ReviewResult)
		reviewResult = &rr
	}

	records, total, err := l.svcCtx.SubmissionRecordModel.FindByReviewer(l.ctx, userId, entityType, reviewResult, int64(req.Page), int64(req.PageSize))
	if err != nil {
		return nil, err
	}

	var list []types.SubmissionRecordItem
	for _, r := range records {
		item := types.SubmissionRecordItem{
			Id:             r.Id,
			EntityType:     r.EntityType,
			EntityId:       r.EntityId,
			SubmissionType: int32(r.SubmissionType),
			SubmitterId:    r.SubmitterId,
			SubmitTime:     r.SubmitTime.Format("2006-01-02 15:04:05"),
			ReviewResult:   int32(r.ReviewResult),
			ReviewNotes:    "",
		}
		if r.EntityName.Valid {
			item.EntityName = r.EntityName.String
		}
		if r.EntityCode.Valid {
			item.EntityCode = r.EntityCode.String
		}
		if r.ReviewerId.Valid {
			rid := r.ReviewerId.Int64
			item.ReviewerId = &rid
		}
		if r.ReviewTime.Valid {
			item.ReviewTime = r.ReviewTime.Time.Format("2006-01-02 15:04:05")
		}
		if r.ReviewNotes.Valid {
			item.ReviewNotes = r.ReviewNotes.String
		}
		list = append(list, item)
	}

	if list == nil {
		list = []types.SubmissionRecordItem{}
	}

	return &types.GetSubmissionRecordsResp{List: list, Total: total}, nil
}
