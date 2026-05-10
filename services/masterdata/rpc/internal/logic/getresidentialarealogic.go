package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResidentialAreaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetResidentialAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResidentialAreaLogic {
	return &GetResidentialAreaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetResidentialAreaLogic) GetResidentialArea(in *pb.GetResidentialAreaReq) (*pb.GetResidentialAreaResp, error) {
	c, err := l.svcCtx.MdResidentialAreaModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &pb.GetResidentialAreaResp{}, nil
	}
	if c.DeleteTime.Valid {
		return &pb.GetResidentialAreaResp{}, nil
	}
	return &pb.GetResidentialAreaResp{ResidentialArea: modelResidentialAreaToPb(c)}, nil
}

func modelResidentialAreaToPb(c *model.MdResidentialArea) *pb.ResidentialArea {
	result := &pb.ResidentialArea{
		Id:               c.Id,
		Name:             c.Name,
		Address:          c.Address,
		CommunityType:    int32(c.CommunityType),
		SubmissionStatus: int32(c.SubmissionStatus),
		SubmitterId:      c.SubmitterId,
		CreatedTime:      c.CreatedTime.Format("2006-01-02 15:04:05"),
		UpdatedTime:      c.UpdatedTime.Format("2006-01-02 15:04:05"),
	}
	if c.CountyId.Valid {
		result.CountyId = c.CountyId.Int64
	}
	if c.StreetId.Valid {
		result.StreetId = c.StreetId.Int64
	}
	if c.CommunityDivId.Valid {
		result.CommunityDivId = c.CommunityDivId.Int64
	}
	if c.Code.Valid {
		result.Code = c.Code.String
	}
	if c.Area.Valid {
		result.Area = c.Area.Float64
	}
	if c.Population.Valid {
		result.Population = int32(c.Population.Int64)
	}
	if c.SubmitTime.Valid {
		result.SubmitTime = c.SubmitTime.Time.Format("2006-01-02 15:04:05")
	}
	if c.ReviewerId.Valid {
		result.ReviewerId = c.ReviewerId.Int64
	}
	if c.ReviewTime.Valid {
		result.ReviewTime = c.ReviewTime.Time.Format("2006-01-02 15:04:05")
	}
	if c.ReviewNotes.Valid {
		result.ReviewNotes = c.ReviewNotes.String
	}
	return result
}
