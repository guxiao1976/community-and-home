package residentialarea

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResidentialAreaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetResidentialAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResidentialAreaLogic {
	return &GetResidentialAreaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetResidentialAreaLogic) GetResidentialArea(req *types.GetResidentialAreaReq) (resp *types.GetResidentialAreaResp, err error) {
	area, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("住宅小区不存在")
	}
	if area.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("住宅小区不存在")
	}
	return &types.GetResidentialAreaResp{
		ResidentialArea: modelToResidentialArea(area),
	}, nil
}
