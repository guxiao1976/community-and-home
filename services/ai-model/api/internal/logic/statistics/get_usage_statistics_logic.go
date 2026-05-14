// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package statistics

import (
	"context"
	"fmt"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/model"

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
	// 构建查询条件
	query := "SELECT * FROM am_usage_statistics WHERE 1=1"
	args := []interface{}{}

	if req.ModelId > 0 {
		query += " AND model_id = ?"
		args = append(args, req.ModelId)
	}

	if req.StartDate != "" {
		query += " AND stat_date >= ?"
		args = append(args, req.StartDate)
	}

	if req.EndDate != "" {
		query += " AND stat_date <= ?"
		args = append(args, req.EndDate)
	}

	query += " ORDER BY stat_date DESC"

	// 计算分页
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query += " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	// 查询数据
	var statistics []*model.AmUsageStatistics
	err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &statistics, query, args...)
	if err != nil {
		l.Errorf("query usage statistics failed: %v", err)
		return &types.UsageStatisticsResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: fmt.Sprintf("query failed: %v", err),
			},
		}, nil
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM am_usage_statistics WHERE 1=1"
	countArgs := []interface{}{}

	if req.ModelId > 0 {
		countQuery += " AND model_id = ?"
		countArgs = append(countArgs, req.ModelId)
	}

	if req.StartDate != "" {
		countQuery += " AND stat_date >= ?"
		countArgs = append(countArgs, req.StartDate)
	}

	if req.EndDate != "" {
		countQuery += " AND stat_date <= ?"
		countArgs = append(countArgs, req.EndDate)
	}

	var total int32
	err = l.svcCtx.DB.QueryRowCtx(l.ctx, &total, countQuery, countArgs...)
	if err != nil {
		l.Errorf("query total count failed: %v", err)
		total = 0
	}

	// 转换为响应格式
	statisticsInfo := make([]types.UsageStatisticsInfo, 0, len(statistics))
	for _, stat := range statistics {
		statisticsInfo = append(statisticsInfo, types.UsageStatisticsInfo{
			Id:           stat.Id,
			ModelId:      stat.ModelId,
			Date:         stat.StatDate.Format("2006-01-02"),
			TotalCalls:   stat.TotalCalls,
			SuccessCalls: stat.SuccessCalls,
			FailedCalls:  stat.FailedCalls,
			TotalTokens:  stat.TotalTokens,
			InputTokens:  stat.InputTokens,
			OutputTokens: stat.OutputTokens,
			TotalCost:    stat.TotalCost,
			AvgLatency:   float64(stat.AvgLatencyMs),
			CreatedAt:    stat.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedAt:    stat.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.UsageStatisticsResponse{
		BaseResponse: types.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Data: types.UsageStatisticsData{
			Statistics: statisticsInfo,
			Total:      total,
		},
	}, nil
}
