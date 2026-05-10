// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package configuration

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSubmitConfigurationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Batch submit configurations
func NewBatchSubmitConfigurationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSubmitConfigurationsLogic {
	return &BatchSubmitConfigurationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchSubmitConfigurationsLogic) BatchSubmitConfigurations(req *types.BatchSubmitReq) (resp *types.BatchSubmitResp, err error) {
	if len(req.Ids) == 0 {
		return nil, errorx.NewDefaultError("请选择要提交的数据")
	}

	successCount := 0
	for _, id := range req.Ids {
		config, err := l.svcCtx.MdConfigurationModel.FindOne(l.ctx, id)
		if err != nil || config.DeleteTime.Valid {
			continue
		}
		if config.SubmissionStatus != 0 && config.SubmissionStatus != 3 {
			continue
		}
		config.SubmissionStatus = 1
		if err := l.svcCtx.MdConfigurationModel.Update(l.ctx, config); err != nil {
			continue
		}
		successCount++
	}

	return &types.BatchSubmitResp{Success: successCount > 0}, nil
}
