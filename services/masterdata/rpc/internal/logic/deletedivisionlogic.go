package logic

import (
	"context"
	"fmt"

	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDivisionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDivisionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDivisionLogic {
	return &DeleteDivisionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDivisionLogic) DeleteDivision(in *pb.DeleteDivisionReq) (*pb.DeleteDivisionResp, error) {
	existing, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("division not found")
		}
		return nil, err
	}
	if existing.DeleteTime.Valid {
		return nil, fmt.Errorf("division already deleted")
	}

	children, err := l.svcCtx.MdAdministrativeDivisionModel.FindChildren(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	activeChildren := 0
	for _, child := range children {
		if !child.DeleteTime.Valid {
			activeChildren++
		}
	}
	if activeChildren > 0 {
		return nil, fmt.Errorf("cannot delete division with active children")
	}

	err = l.svcCtx.MdAdministrativeDivisionModel.SoftDelete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteDivisionResp{}, nil
}
