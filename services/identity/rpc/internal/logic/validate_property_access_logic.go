package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidatePropertyAccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidatePropertyAccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidatePropertyAccessLogic {
	return &ValidatePropertyAccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidatePropertyAccessLogic) ValidatePropertyAccess(in *pb.ValidatePropertyAccessReq) (*pb.ValidatePropertyAccessResp, error) {
	binding, err := l.svcCtx.AuthPropertyBindingModel.FindOneByUserIdPropertyUnitId(l.ctx, in.UserId, in.PropertyUnitId)
	if err != nil {
		return &pb.ValidatePropertyAccessResp{
			HasAccess: false,
		}, nil
	}

	bindingType := "tenant"
	if binding.IsPrimary == 1 {
		bindingType = "homeowner"
	}

	return &pb.ValidatePropertyAccessResp{
		HasAccess:   true,
		BindingType: bindingType,
		IsPrimary:   int32(binding.IsPrimary),
	}, nil
}
