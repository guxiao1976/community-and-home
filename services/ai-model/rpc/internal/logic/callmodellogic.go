package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCallModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelLogic {
	return &CallModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通用模型调用（单次）
func (l *CallModelLogic) CallModel(in *pb.ModelCallRequest) (*pb.ModelCallResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ModelCallResponse{}, nil
}
