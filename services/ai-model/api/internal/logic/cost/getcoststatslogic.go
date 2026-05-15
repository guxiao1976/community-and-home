// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCostStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取成本统计
func NewGetCostStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCostStatsLogic {
	return &GetCostStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCostStatsLogic) GetCostStats(req *types.GetCostStatsRequest) (resp *types.CostStatsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
