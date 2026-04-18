package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommunitiesByDivisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommunitiesByDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommunitiesByDivisionLogic {
	return &GetCommunitiesByDivisionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCommunitiesByDivisionLogic) GetCommunitiesByDivision(in *pb.GetCommunitiesByDivisionReq) (*pb.GetCommunitiesByDivisionResp, error) {
	communities, err := l.svcCtx.MdCommunityModel.FindByDivisionId(l.ctx, in.DivisionId)
	if err != nil {
		return nil, err
	}

	var result []*pb.Community
	for _, c := range communities {
		if c.DeleteTime.Valid {
			continue
		}
		if in.Status == 1 && c.SubmissionStatus != 2 {
			continue // 只返回审核通过的
		}
		result = append(result, modelCommunityToPb(c))
	}
	return &pb.GetCommunitiesByDivisionResp{Communities: result}, nil
}
