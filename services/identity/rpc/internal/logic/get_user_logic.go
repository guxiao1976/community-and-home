package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *pb.GetUserReq) (*pb.GetUserResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetUserResp{}, nil
}
