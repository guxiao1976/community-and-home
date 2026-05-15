package logic

import (
	"context"

	"community-and-home/services/ai-model/rpc/internal/svc"
	"community-and-home/services/ai-model/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallModelBatchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCallModelBatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallModelBatchLogic {
	return &CallModelBatchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 通用模型调用（批量）
func (l *CallModelBatchLogic) CallModelBatch(in *pb.ModelBatchRequest) (*pb.ModelBatchResponse, error) {
	// todo: add your logic here and delete this line

	return &pb.ModelBatchResponse{}, nil
}
