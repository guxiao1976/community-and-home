package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshTokenLogic) RefreshToken(in *pb.RefreshTokenReq) (*pb.RefreshTokenResp, error) {
	// todo: add your logic here and delete this line

	return &pb.RefreshTokenResp{}, nil
}
