package division

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionChildCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDivisionChildCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionChildCountLogic {
	return &GetDivisionChildCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDivisionChildCountLogic) GetDivisionChildCount(req *types.DeleteDivisionReq) (resp *types.DivisionChildCountResp, err error) {
	div, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("行政区划不存在")
		}
		return nil, errorx.NewDefaultError("查询行政区划失败")
	}

	childCount, _ := l.svcCtx.MdAdministrativeDivisionModel.CountByParentId(l.ctx, req.Id)
	hasChildDivisions := childCount > 0

	var hasResidentialAreas bool
	switch div.Level {
	case 3:
		count, _ := l.svcCtx.MdResidentialAreaModel.Count(l.ctx, &req.Id, nil, nil, nil, nil, nil, nil)
		hasResidentialAreas = count > 0
	case 4:
		count, _ := l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, &req.Id, nil, nil, nil, nil, nil)
		hasResidentialAreas = count > 0
	case 5:
		count, _ := l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, &req.Id, nil, nil, nil, nil)
		hasResidentialAreas = count > 0
	}

	return &types.DivisionChildCountResp{
		HasChildDivisions:   hasChildDivisions,
		HasResidentialAreas: hasResidentialAreas,
		HasData:             hasChildDivisions || hasResidentialAreas,
	}, nil
}
