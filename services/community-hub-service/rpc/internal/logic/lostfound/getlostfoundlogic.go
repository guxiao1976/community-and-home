package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLostFoundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLostFoundLogic {
	return &GetLostFoundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLostFoundLogic) GetLostFound(in *communityv1.GetLostFoundRequest) (*communityv1.GetLostFoundResponse, error) {
	it, err := l.svcCtx.LostFoundItemModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("GetLostFound: not found id=%d", in.Id)
		return &communityv1.GetLostFoundResponse{
			Base: responsex.NewBaseRespWithError(80004, "寻失记录不存在"),
		}, nil
	}

	return &communityv1.GetLostFoundResponse{
		Base: responsex.NewBaseResp(),
		Item: toProtoLostFoundItem(it),
	}, nil
}
