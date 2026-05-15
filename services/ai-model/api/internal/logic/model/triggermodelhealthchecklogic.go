// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerModelHealthCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 触发模型健康检查
func NewTriggerModelHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerModelHealthCheckLogic {
	return &TriggerModelHealthCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TriggerModelHealthCheckLogic) TriggerModelHealthCheck(req *types.TriggerHealthCheckRequest) (resp *types.HealthCheckRecordResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
