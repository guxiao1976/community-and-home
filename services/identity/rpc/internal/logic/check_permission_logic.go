package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermissionLogic {
	return &CheckPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckPermissionLogic) CheckPermission(in *pb.CheckPermissionReq) (*pb.CheckPermissionResp, error) {
	// todo: add your logic here and delete this line

	return &pb.CheckPermissionResp{}, nil
}
