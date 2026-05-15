// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package statistics

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUsageStatisticsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取使用统计
func NewGetUsageStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsageStatisticsLogic {
	return &GetUsageStatisticsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUsageStatisticsLogic) GetUsageStatistics(req *types.GetUsageStatisticsRequest) (resp *types.UsageStatisticsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
