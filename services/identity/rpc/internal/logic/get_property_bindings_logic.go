package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/types/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPropertyBindingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPropertyBindingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPropertyBindingsLogic {
	return &GetPropertyBindingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPropertyBindingsLogic) GetPropertyBindings(in *pb.GetPropertyBindingsReq) (*pb.GetPropertyBindingsResp, error) {
	bindings, err := l.svcCtx.AuthPropertyBindingModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	var pbBindings []*pb.PropertyBinding
	for _, binding := range bindings {
		bindingType := "tenant"
		if binding.IsPrimary == 1 {
			bindingType = "homeowner"
		}

		pbBinding := &pb.PropertyBinding{
			Id:             binding.Id,
			UserId:         binding.UserId,
			PropertyUnitId: binding.PropertyUnitId,
			BindingType:    bindingType,
			IsPrimary:      int32(binding.IsPrimary),
			BoundTime:      binding.BindTime.Format("2006-01-02 15:04:05"),
		}
		pbBindings = append(pbBindings, pbBinding)
	}

	return &pb.GetPropertyBindingsResp{
		Bindings: pbBindings,
	}, nil
}
