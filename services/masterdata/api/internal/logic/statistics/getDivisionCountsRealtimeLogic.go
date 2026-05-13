package statistics

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionCountsRealtimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDivisionCountsRealtimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionCountsRealtimeLogic {
	return &GetDivisionCountsRealtimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDivisionCountsRealtimeLogic) GetDivisionCountsRealtime(req *types.DivisionCountsReq) (resp *types.DivisionCountsResp, err error) {
	rows, err := l.svcCtx.MdDivisionStatisticsModel.FindRealtimeCountsByParentId(l.ctx, req.ParentId)
	if err != nil {
		return &types.DivisionCountsResp{List: []types.DivisionCountItem{}}, nil
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
