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

type SubmitDivisionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Submit division
func NewSubmitDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitDivisionLogic {
	return &SubmitDivisionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitDivisionLogic) SubmitDivision(req *types.SubmitReq) (resp *types.SubmitResp, err error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("行政区划不存在")
		}
		return nil, errorx.NewDefaultError("查询行政区划失败")
	}
	if div.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("行政区划已删除")
	}

	if div.SubmissionStatus != 0 && div.SubmissionStatus != 3 {
		return nil, errorx.NewDefaultError("当前状态不允许提交")
	}

	div.SubmissionStatus = 1
	var submitterId int64
	if uid := l.ctx.Value("userId"); uid != nil {
		submitterId = uid.(int64)
	}
	div.SubmitterId = sql.NullInt64{Int64: submitterId, Valid: true}
	div.SubmitTime = sql.NullTime{Time: time.Now(), Valid: true}
	if err := l.svcCtx.MdAdministrativeDivisionModel.Update(l.ctx, div); err != nil {
		return nil, errorx.NewDefaultError("提交失败: " + err.Error())
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

	return &types.SubmitResp{Success: true}, nil
}
