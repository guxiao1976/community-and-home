package residentialarea

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewResidentialAreaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewResidentialAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewResidentialAreaLogic {
	return &ReviewResidentialAreaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewResidentialAreaLogic) ReviewResidentialArea(req *types.ReviewResidentialAreaReq) (resp *types.ReviewResidentialAreaResp, err error) {
	area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("住宅小区不存在")
	}
	if area.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("住宅小区已删除")
	}

	// 只有已提交状态才能审核
	if area.SubmissionStatus != 1 {
		return nil, errorx.NewDefaultError("当前状态不允许审核")
	}

	// 获取审核人 ID
	reviewerId := int64(1)
	if uid := l.ctx.Value("userId"); uid != nil {
		reviewerId = uid.(int64)
	}

	now := time.Now()
	area.ReviewerId = sql.NullInt64{Int64: reviewerId, Valid: true}
	area.ReviewTime = sql.NullTime{Time: now, Valid: true}
	area.ReviewNotes = sql.NullString{String: req.ReviewNotes, Valid: req.ReviewNotes != ""}

	switch req.Action {
	case "approve":
		if area.SubmissionStatus == 4 {
			// 待删除审批通过 → 执行软删除
			area.DeleteTime = sql.NullTime{Time: now, Valid: true}
		}
		area.SubmissionStatus = 2 // 已通过
	case "reject":
		if req.ReviewNotes == "" {
			return nil, errorx.NewDefaultError("拒绝时必须填写审核备注")
		}
		if area.SubmissionStatus == 4 {
			// 待删除审批拒绝 → 恢复为已拒绝
			area.SubmissionStatus = 3
		} else {
			area.SubmissionStatus = 3 // 已拒绝
		}
	default:
		return nil, errorx.NewDefaultError("无效的审核操作")
	}

	if err := l.svcCtx.MdResidentialAreaModel.Update(l.ctx, area); err != nil {
		return nil, errorx.NewDefaultError("审核失败: " + err.Error())
	}

	return &types.ReviewResidentialAreaResp{Success: true}, nil
}
