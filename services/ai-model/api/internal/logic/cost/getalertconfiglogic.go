// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
