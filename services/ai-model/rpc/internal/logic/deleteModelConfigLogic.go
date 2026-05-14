package logic

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteModelConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteModelConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteModelConfigLogic {
	return &DeleteModelConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteModelConfigLogic) DeleteModelConfig(in *pb.DeleteModelConfigReq) (*pb.DeleteModelConfigResp, error) {
	// 删除模型配置
	err := l.svcCtx.ModelManager.DeleteModelConfig(l.ctx, in.Id)
	if err != nil {
		l.Errorf("delete model config failed: %v", err)
		return nil, err
	}

	return &pb.DeleteModelConfigResp{
		Success: true,
	}, nil
}
