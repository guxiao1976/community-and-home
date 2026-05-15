// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
