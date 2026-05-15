// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 调用模型
func NewCallModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelLogic {
	return &CallModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallModelLogic) CallModel(req *types.ModelCallRequest) (resp *types.ModelCallResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
