// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallModelBatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 批量调用模型
func NewCallModelBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelBatchLogic {
	return &CallModelBatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallModelBatchLogic) CallModelBatch(req *types.ModelBatchRequest) (resp *types.ModelBatchResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
