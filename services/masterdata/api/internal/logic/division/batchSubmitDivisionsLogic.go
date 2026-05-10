// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package division

import (
	"context"
	"database/sql"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSubmitDivisionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Batch submit divisions
func NewBatchSubmitDivisionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSubmitDivisionsLogic {
	return &BatchSubmitDivisionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchSubmitDivisionsLogic) BatchSubmitDivisions(req *types.BatchSubmitReq) (resp *types.BatchSubmitResp, err error) {
	if len(req.Ids) == 0 {
		return nil, errorx.NewDefaultError("请选择要提交的数据")
	}

	var submitterId int64
	if uid := l.ctx.Value("userId"); uid != nil {
		submitterId = uid.(int64)
	}

	successCount := 0
	for _, id := range req.Ids {
		div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, id)
		if err != nil || div.DeleteTime.Valid {
			continue
		}
		if div.SubmissionStatus != 0 && div.SubmissionStatus != 3 {
			continue
		}
		div.SubmissionStatus = 1
		if err := l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, div); err != nil {
			continue
		}

		subType := int64(1)
		if div.SubmissionType.Valid {
			subType = div.SubmissionType.Int64
		}
		_, _ = l.svcCtx.SubmissionRecordModel.Insert(l.ctx, &model.SubmissionRecord{
			EntityType:     "administrative_division",
			EntityId:       div.Id,
			EntityName:     sql.NullString{String: div.Name, Valid: true},
			EntityCode:     sql.NullString{String: div.Code, Valid: true},
			SubmissionType: subType,
			SubmitterId:    submitterId,
			SubmitTime:     time.Now(),
		})

		successCount++
	}

	return &types.BatchSubmitResp{Success: successCount > 0}, nil
}
