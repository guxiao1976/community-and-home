package residentialarea

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

type SubmitResidentialAreaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitResidentialAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitResidentialAreaLogic {
	return &SubmitResidentialAreaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitResidentialAreaLogic) SubmitResidentialArea(req *types.SubmitResidentialAreaReq) (resp *types.SubmitResidentialAreaResp, err error) {
	area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("住宅小区不存在")
	}
	if area.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("住宅小区已删除")
	}

	// 只有草稿或已拒绝状态才能提交
	if area.SubmissionStatus != 0 && area.SubmissionStatus != 3 {
		return nil, errorx.NewDefaultError("当前状态不允许提交")
	}

	area.SubmissionStatus = 1
	if err := l.svcCtx.MdResidentialAreaModel.Update(l.ctx, area); err != nil {
		return nil, errorx.NewDefaultError("提交失败: " + err.Error())
	}

	var submitterId int64
	if uid := l.ctx.Value("userId"); uid != nil {
		submitterId = uid.(int64)
	}
	subType := int64(1)
	if area.SubmissionType.Valid {
		subType = area.SubmissionType.Int64
	}
	_, _ = l.svcCtx.SubmissionRecordModel.Insert(l.ctx, &model.SubmissionRecord{
		EntityType:     "residential_area",
		EntityId:       area.Id,
		EntityName:     sql.NullString{String: area.Name, Valid: true},
		EntityCode:     area.Code,
		SubmissionType: subType,
		SubmitterId:    submitterId,
		SubmitTime:     time.Now(),
	})

	return &types.SubmitResidentialAreaResp{Success: true}, nil
}
