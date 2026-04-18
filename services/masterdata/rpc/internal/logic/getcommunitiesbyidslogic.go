package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommunitiesByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommunitiesByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommunitiesByIdsLogic {
	return &GetCommunitiesByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCommunitiesByIdsLogic) GetCommunitiesByIds(in *pb.GetCommunitiesByIdsReq) (*pb.GetCommunitiesByIdsResp, error) {
	var communities []*pb.Community
	for _, id := range in.Ids {
		c, err := l.svcCtx.MdCommunityModel.FindOne(l.ctx, id)
		if err != nil || c.DeleteTime.Valid {
			continue
		}
		communities = append(communities, modelCommunityToPb(c))
	}
	return &pb.GetCommunitiesByIdsResp{Communities: communities}, nil
}
