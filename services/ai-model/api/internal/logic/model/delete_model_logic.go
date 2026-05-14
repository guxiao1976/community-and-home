// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package model

import (
	"context"

	"community-and-home/services/ai-model/api/internal/svc"
	"community-and-home/services/ai-model/api/internal/types"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除模型配置
func NewDeleteModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteModelLogic {
	return &DeleteModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteModelLogic) DeleteModel(req *types.DeleteModelRequest) (resp *types.BaseResponse, err error) {
	// 调用 RPC 服务删除模型配置
	_, err = l.svcCtx.AiModelRpc.DeleteModelConfig(l.ctx, &pb.DeleteModelConfigReq{
		Id: req.Id,
	})
	if err != nil {
		l.Errorf("delete model config failed: %v", err)
		return nil, err
	}

	// 返回响应
	return &types.BaseResponse{
		Code:    0,
		Message: "success",
	}, nil
}
