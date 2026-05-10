package deleteditems

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeletedCountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeletedCountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeletedCountsLogic {
	return &GetDeletedCountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeletedCountsLogic) GetDeletedCounts() (resp *types.DeletedCountsResp, err error) {
	var residentialCount, divisionCount, configCount, sensitiveCount int64

	if c, e := l.svcCtx.MdResidentialAreaModel.CountDeleted(l.ctx); e == nil {
		residentialCount = c
	}
	if c, e := l.svcCtx.MdAdministrativeDivisionModel.CountDeleted(l.ctx); e == nil {
		divisionCount = c
	}
	if c, e := l.svcCtx.MdConfigurationModel.CountDeleted(l.ctx); e == nil {
		configCount = c
	}
	if c, e := l.svcCtx.MdSensitiveWordModel.CountDeleted(l.ctx); e == nil {
		sensitiveCount = c
	}

	return &types.DeletedCountsResp{
		ResidentialArea:        residentialCount,
		AdministrativeDivision: divisionCount,
		Configuration:          configCount,
		SensitiveWord:          sensitiveCount,
		Total:                  residentialCount + divisionCount + configCount + sensitiveCount,
	}, nil
}
