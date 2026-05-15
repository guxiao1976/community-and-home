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
	// todo: add your logic here and delete this line

	return &pb.DeleteModelConfigResp{}, nil
}
