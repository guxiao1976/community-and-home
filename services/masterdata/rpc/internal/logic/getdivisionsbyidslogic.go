package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionsByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDivisionsByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionsByIdsLogic {
	return &GetDivisionsByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDivisionsByIdsLogic) GetDivisionsByIds(in *pb.GetDivisionsByIdsReq) (*pb.GetDivisionsByIdsResp, error) {
	var divisions []*pb.Division
	for _, id := range in.Ids {
		d, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, id)
		if err != nil || d.DeleteTime.Valid {
			continue
		}
		divisions = append(divisions, modelDivisionToPb(d))
	}
	return &pb.GetDivisionsByIdsResp{Divisions: divisions}, nil
}