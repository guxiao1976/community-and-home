package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginSmsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginSmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginSmsLogic {
	return &LoginSmsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginSmsLogic) LoginSms(in *pb.LoginSmsReq) (*pb.LoginSmsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.LoginSmsResp{}, nil
}
