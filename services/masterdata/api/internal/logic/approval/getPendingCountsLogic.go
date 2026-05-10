package approval

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPendingCountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPendingCountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPendingCountsLogic {
	return &GetPendingCountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPendingCountsLogic) GetPendingCounts() (resp *types.PendingCountsResp, err error) {
	var residentialAreaCount, administrativeDivisionCount, configurationCount, sensitiveWordCount int64

	if c, e := l.svcCtx.MdResidentialAreaModel.CountBySubmissionStatus(l.ctx, 1); e == nil {
		residentialAreaCount = c
	}
	if c, e := l.svcCtx.MdAdministrativeDivisionModel.CountBySubmissionStatus(l.ctx, 1); e == nil {
		administrativeDivisionCount = c
	}
	if c, e := l.svcCtx.MdConfigurationModel.CountBySubmissionStatus(l.ctx, 1); e == nil {
		configurationCount = c
	}
	if c, e := l.svcCtx.MdSensitiveWordModel.CountBySubmissionStatus(l.ctx, 1); e == nil {
		sensitiveWordCount = c
	}

	resp = &types.PendingCountsResp{
		ResidentialArea:         residentialAreaCount,
		AdministrativeDivision: administrativeDivisionCount,
		Configuration:          configurationCount,
		SensitiveWord:          sensitiveWordCount,
	}
	resp.Total = resp.ResidentialArea + resp.AdministrativeDivision + resp.Configuration + resp.SensitiveWord

	return resp, nil
}
