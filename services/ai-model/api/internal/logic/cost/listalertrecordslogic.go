// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package cost

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAlertRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取预警记录
func NewListAlertRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAlertRecordsLogic {
	return &ListAlertRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAlertRecordsLogic) ListAlertRecords(req *types.ListAlertRecordsRequest) (resp *types.AlertRecordsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
