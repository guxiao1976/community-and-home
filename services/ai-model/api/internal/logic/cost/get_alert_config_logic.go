// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAlertConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取预警配置
func NewGetAlertConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAlertConfigLogic {
	return &GetAlertConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAlertConfigLogic) GetAlertConfig() (resp *types.AlertConfigResponse, err error) {
	// 查询预警配置（全局配置，ID=1）
	config, err := l.svcCtx.CostAlertConfigModel.FindOne(l.ctx, 1)
	if err != nil {
		return &types.AlertConfigResponse{
			BaseResponse: types.BaseResponse{
				Code:    500,
				Message: "查询预警配置失败: " + err.Error(),
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
