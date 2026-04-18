package community

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewCommunityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewCommunityLogic {
	return &ReviewCommunityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewCommunityLogic) ReviewCommunity(req *types.ReviewCommunityReq) (resp *types.ReviewCommunityResp, err error) {
	// 1. Find existing
	existing, err := l.svcCtx.MdCommunityModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewNotFoundError("社区不存在")
	}
	if existing.DeleteTime.Valid {
		return nil, errorx.NewNotFoundError("社区已删除")
	}

	// 2. Validate: current status must be Submitted
	if existing.SubmissionStatus != 1 {
		return nil, errorx.NewInvalidParamError("只有待审核状态的社区可以进行审核")
	}

	// 3. Apply review action
	now := time.Now()
	reviewerId := int64(0) // TODO: Get from JWT
	existing.ReviewerId = sql.NullInt64{Int64: reviewerId, Valid: true}
	existing.ReviewTime = sql.NullTime{Time: now, Valid: true}
	existing.UpdatedTime = now

	switch req.Action {
	case "approve":
		existing.SubmissionStatus = 2 // Approved
		if req.ReviewNotes != "" {
			existing.ReviewNotes = sql.NullString{String: req.ReviewNotes, Valid: true}
		}
	case "reject":
		if req.ReviewNotes == "" {
			return nil, errorx.NewInvalidParamError("退回时必须填写审核意见")
		}
		existing.SubmissionStatus = 3 // Rejected
		existing.ReviewNotes = sql.NullString{String: req.ReviewNotes, Valid: true}
	default:
		return nil, errorx.NewInvalidParamError("审核操作必须是 approve 或 reject")
	}

	// 4. Save
	if err := l.svcCtx.MdCommunityModel.Update(l.ctx, existing); err != nil {
		return nil, errorx.NewDefaultError("审核操作失败")
	}

	// 5. Create audit log
	auditLog := &model.MdAuditLog{
		UserId:     reviewerId,
		EntityType: "md_community",
		EntityId:   existing.Id,
		Action:     "REVIEW_" + req.Action,
		NewValue:   sql.NullString{String: fmt.Sprintf(`{"submission_status": %d, "review_notes": "%s"}`, existing.SubmissionStatus, req.ReviewNotes), Valid: true},
		IpAddress:  "0.0.0.0", // TODO: Get from request
		CreatedTime: now,
	}
	_, _ = l.svcCtx.MdAuditLogModel.Insert(l.ctx, auditLog)

	return &types.ReviewCommunityResp{Success: true}, nil
}