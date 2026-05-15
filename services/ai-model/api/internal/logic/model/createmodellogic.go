// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/api/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建模型配置
func NewCreateModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateModelLogic {
	return &CreateModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateModelLogic) CreateModel(req *types.CreateModelRequest) (resp *types.ModelResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
