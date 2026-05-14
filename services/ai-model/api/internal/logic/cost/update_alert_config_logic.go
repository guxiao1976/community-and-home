// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"context"
	"time"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAlertConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新预警配置
func NewUpdateAlertConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAlertConfigLogic {
	return &UpdateAlertConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAlertConfigLogic) UpdateAlertConfig(req *types.UpdateAlertConfigRequest) (resp *types.AlertConfigResponse, err error) {
	// 查询现有配置
	config, err := l.svcCtx.CostAlertConfigModel.FindOne(l.ctx, 1)
	if err != nil {
		return &types.AlertConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    404,
				Message: "预警配置不存在",
			},
		}, nil
	}

	// 更新字段
	if req.DailyLimit > 0 {
		config.DailyLimit = req.DailyLimit
	}
	if req.MonthlyLimit > 0 {
		config.MonthlyLimit = req.MonthlyLimit
	}
	if req.AlertThreshold > 0 {
		config.AlertThreshold = req.AlertThreshold
	}
	config.UpdatedTime = time.Now()

	// 更新数据库
	err = l.svcCtx.CostAlertConfigModel.Update(l.ctx, config)
	if err != nil {
		return &types.AlertConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "更新预警配置失败: " + err.Error(),
			},
		}, nil
	}

	return &types.AlertConfigResponse{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		Data: types.AlertConfigData{
			DailyLimit:     config.DailyLimit,
			MonthlyLimit:   config.MonthlyLimit,
			AlertThreshold: config.AlertThreshold,
		},
	}, nil
}
