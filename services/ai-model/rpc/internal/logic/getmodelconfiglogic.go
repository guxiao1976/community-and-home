package logic

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetModelConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetModelConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetModelConfigLogic {
	return &GetModelConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetModelConfigLogic) GetModelConfig(in *pb.GetModelConfigReq) (*pb.ModelConfigResp, error) {
	// todo: add your logic here and delete this line

	return &pb.ModelConfigResp{}, nil
}
