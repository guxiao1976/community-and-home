package statistics

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionCountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDivisionCountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionCountsLogic {
	return &GetDivisionCountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDivisionCountsLogic) GetDivisionCounts(req *types.DivisionCountsReq) (resp *types.DivisionCountsResp, err error) {
	statDate, err := l.svcCtx.MdDivisionStatisticsModel.FindLatestDate(l.ctx)
	if err != nil {
		return &types.DivisionCountsResp{List: []types.DivisionCountItem{}}, nil
	}

	rows, err := l.svcCtx.MdDivisionStatisticsModel.FindCountsByParentId(l.ctx, req.ParentId, statDate)
	if err != nil {
		return nil, err
	}

	list := make([]types.DivisionCountItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.DivisionCountItem{
			Id:             r.Id,
			Name:           r.Name,
			Level:          r.Level,
			CommunityCount: r.CommunityCount,
			VillageCount:   r.VillageCount,
			TotalCount:     r.TotalCount,
		})
	}

	return &types.DivisionCountsResp{List: list}, nil
}
