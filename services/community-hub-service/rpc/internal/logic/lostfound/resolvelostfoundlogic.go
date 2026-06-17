package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveLostFoundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResolveLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveLostFoundLogic {
	return &ResolveLostFoundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResolveLostFoundLogic) ResolveLostFound(in *communityv1.ResolveLostFoundRequest) (*communityv1.ResolveLostFoundResponse, error) {
	// 校验存在
	_, err := l.svcCtx.LostFoundItemModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("ResolveLostFound: not found id=%d", in.Id)
		return &communityv1.ResolveLostFoundResponse{
			Base: responsex.NewBaseRespWithError(80004, "寻失记录不存在"),
		}, nil
	}

	if err := l.svcCtx.LostFoundItemModel.UpdateStatus(l.ctx, in.Id, "resolved"); err != nil {
		l.Errorf("ResolveLostFound: update status failed: %v", err)
		return nil, err
	}

	l.Infof("ResolveLostFound success: id=%d", in.Id)
	return &communityv1.ResolveLostFoundResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
