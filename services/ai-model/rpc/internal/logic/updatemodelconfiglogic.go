package logic

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateModelConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateModelConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateModelConfigLogic {
	return &UpdateModelConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateModelConfigLogic) UpdateModelConfig(in *pb.UpdateModelConfigReq) (*pb.ModelConfigResp, error) {
	// todo: add your logic here and delete this line

	return &pb.ModelConfigResp{}, nil
}
