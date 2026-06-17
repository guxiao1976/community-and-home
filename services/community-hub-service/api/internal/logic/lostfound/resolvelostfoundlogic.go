package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ResolveLostFoundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResolveLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveLostFoundLogic {
	return &ResolveLostFoundLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ResolveLostFoundLogic) ResolveLostFound(req *types.ResolveLostFoundReq) error {
	_, err := l.svcCtx.LostFoundServiceRpc.ResolveLostFound(l.ctx, &communityv1.ResolveLostFoundRequest{
		Id: req.Id,
	})
	return err
}
