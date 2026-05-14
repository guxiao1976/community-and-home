package cost

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

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
	// 查询总体统计
	var totalCost float64
	var totalRequests int64
	var totalTokens int64

	query := `
		SELECT
			COALESCE(SUM(cost), 0) as total_cost,
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM am_call_log
		WHERE created_time >= ? AND created_time < DATE_ADD(?, INTERVAL 1 DAY)
	`

	conn, _ := l.svcCtx.DB.RawDB()
	err = conn.QueryRowContext(l.ctx, query, req.StartDate, req.EndDate).Scan(
		&totalCost, &totalRequests, &totalTokens,
	)
	if err != nil {
		return &types.CostStatsResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询成本统计失败: " + err.Error(),
			},
		}, nil
	}

	// 查询按模型分组的统计
	modelQuery := `
		SELECT
			model_name,
			COALESCE(SUM(cost), 0) as total_cost
		FROM am_call_log
		WHERE created_time >= ? AND created_time < DATE_ADD(?, INTERVAL 1 DAY)
		GROUP BY model_name
		ORDER BY total_cost DESC
	`

	rows, err := conn.QueryContext(l.ctx, modelQuery, req.StartDate, req.EndDate)
	if err != nil {
		return &types.CostStatsResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询模型成本统计失败: " + err.Error(),
			},
		}, nil
	}
	defer rows.Close()

	modelCosts := make(map[string]float64)
	for rows.Next() {
		var modelName string
		var cost float64

		err = rows.Scan(&modelName, &cost)
		if err != nil {
			continue
		}

		modelCosts[modelName] = cost
	}

	return &types.CostStatsResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.CostStatsData{
			TotalCost:     totalCost,
			DailyCost:     totalCost, // 简化处理，实际应该按日期分组
			MonthlyCost:   totalCost,
			ModelCosts:    modelCosts,
			TotalRequests: totalRequests,
			TotalTokens:   totalTokens,
		},
	}, nil
}
