package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResidentialAreasByDivisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetResidentialAreasByDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResidentialAreasByDivisionLogic {
	return &GetResidentialAreasByDivisionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetResidentialAreasByDivisionLogic) GetResidentialAreasByDivision(in *pb.GetResidentialAreasByDivisionReq) (*pb.GetResidentialAreasByDivisionResp, error) {
	var areas []*model.MdResidentialArea
	var err error

	switch {
	case in.CommunityDivId > 0:
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCommunityDivId(l.ctx, in.CommunityDivId, nil, 1, 1000, 4)
	case in.StreetId > 0:
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByStreetId(l.ctx, in.StreetId, nil, 1, 1000, 4)
	case in.CountyId > 0:
		areas, err = l.svcCtx.MdResidentialAreaModel.FindByCountyId(l.ctx, in.CountyId, nil, 1, 1000, 4)
	default:
		areas, err = l.svcCtx.MdResidentialAreaModel.FindAll(l.ctx, nil, 1, 1000, 4)
	}

	if err != nil {
		return nil, err
	}

	var result []*pb.ResidentialArea
	for _, c := range areas {
		if c.DeleteTime.Valid {
			continue
		}
		if in.Status == 1 && c.SubmissionStatus != 2 {
			continue
		}
		result = append(result, modelResidentialAreaToPb(c))
	}
	return &pb.GetResidentialAreasByDivisionResp{ResidentialAreas: result}, nil
}
