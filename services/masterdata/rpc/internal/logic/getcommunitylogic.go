package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommunityLogic {
	return &GetCommunityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCommunityLogic) GetCommunity(in *pb.GetCommunityReq) (*pb.GetCommunityResp, error) {
	c, err := l.svcCtx.MdCommunityModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &pb.GetCommunityResp{}, nil
	}
	if c.DeleteTime.Valid {
		return &pb.GetCommunityResp{}, nil
	}
	return &pb.GetCommunityResp{Community: modelCommunityToPb(c)}, nil
}
