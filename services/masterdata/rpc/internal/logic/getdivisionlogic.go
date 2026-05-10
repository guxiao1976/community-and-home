package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"


	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionLogic {
	return &GetDivisionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDivisionLogic) GetDivision(in *pb.GetDivisionReq) (*pb.GetDivisionResp, error) {
	d, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return &pb.GetDivisionResp{}, nil
		}
		return nil, err
	}
	if d.DeleteTime.Valid {
		return &pb.GetDivisionResp{}, nil
	}

	return &pb.GetDivisionResp{
		Division: modelDivisionToPb(d),
	}, nil
}

func modelDivisionToPb(d *model.MdAdministrativeDivision) *pb.Division {
	result := &pb.Division{
		Id:          d.Id,
		Level:       int32(d.Level),
		Name:        d.Name,
		Code:        d.Code,
		Path:        d.Path,
		SortOrder:   int32(d.SortOrder),
		Status:      int32(d.Status),
		CreatedBy:   d.CreatedBy,
		CreatedTime: d.CreatedTime.Format("2006-01-02 15:04:05"),
		UpdatedTime: d.UpdatedTime.Format("2006-01-02 15:04:05"),
	}
	if d.ParentId.Valid {
		result.ParentId = d.ParentId.Int64
	}
	return result
}
