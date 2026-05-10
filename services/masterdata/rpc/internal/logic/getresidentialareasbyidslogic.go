package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResidentialAreasByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetResidentialAreasByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResidentialAreasByIdsLogic {
	return &GetResidentialAreasByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetResidentialAreasByIdsLogic) GetResidentialAreasByIds(in *pb.GetResidentialAreasByIdsReq) (*pb.GetResidentialAreasByIdsResp, error) {
	var areas []*pb.ResidentialArea
	for _, id := range in.Ids {
		c, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, id)
		if err != nil || c.DeleteTime.Valid {
			continue
		}
		areas = append(areas, modelResidentialAreaToPb(c))
	}
	return &pb.GetResidentialAreasByIdsResp{ResidentialAreas: areas}, nil
}
