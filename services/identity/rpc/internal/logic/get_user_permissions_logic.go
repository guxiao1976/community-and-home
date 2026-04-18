package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionsLogic {
	return &GetUserPermissionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserPermissionsLogic) GetUserPermissions(in *pb.GetUserPermissionsReq) (*pb.GetUserPermissionsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetUserPermissionsResp{}, nil
}
