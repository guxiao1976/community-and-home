// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package configuration

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

type SubmitConfigurationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Submit configuration
func NewSubmitConfigurationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitConfigurationLogic {
	return &SubmitConfigurationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitConfigurationLogic) SubmitConfiguration(req *types.SubmitReq) (resp *types.SubmitResp, err error) {
	config, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("系统配置不存在")
		}
		return nil, errorx.NewDefaultError("查询系统配置失败")
	}
	if config.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("系统配置已删除")
	}

	if config.SubmissionStatus != 0 && config.SubmissionStatus != 3 {
		return nil, errorx.NewDefaultError("当前状态不允许提交")
	}

	config.SubmissionStatus = 1
	var submitterId int64
	if uid := l.ctx.Value("userId"); uid != nil {
		submitterId = uid.(int64)
	}
	config.SubmitterId = sql.NullInt64{Int64: submitterId, Valid: true}
	config.SubmitTime = sql.NullTime{Time: time.Now(), Valid: true}
	if err := l.svcCtx.MdConfigurationModel.Update(l.ctx, config); err != nil {
		return nil, errorx.NewDefaultError("提交失败: " + err.Error())
	}

	subType := int64(1)
	if config.SubmissionType.Int64 > 0 {
		subType = config.SubmissionType.Int64
	}
	_, _ = l.svcCtx.SubmissionRecordModel.Insert(l.ctx, &model.SubmissionRecord{
		EntityType:     "configuration",
		EntityId:       config.Id,
		EntityName:     sql.NullString{String: config.ConfigKey, Valid: true},
		SubmissionType: subType,
		SubmitterId:    submitterId,
		SubmitTime:     time.Now(),
	})

	return &types.SubmitResp{Success: true}, nil
}
