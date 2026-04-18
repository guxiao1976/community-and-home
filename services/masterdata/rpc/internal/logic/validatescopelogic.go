package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateScopeLogic {
	return &ValidateScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ValidateScopeLogic) ValidateScope(in *pb.ValidateScopeReq) (*pb.ValidateScopeResp, error) {
	d, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, in.DivisionId)
	if err != nil || d.DeleteTime.Valid {
		return &pb.ValidateScopeResp{Allowed: false}, nil
	}
	// TODO: 校验用户是否有该区域的权限 (基于 auth_user 的 scope_id)
	return &pb.ValidateScopeResp{Allowed: true}, nil
}
