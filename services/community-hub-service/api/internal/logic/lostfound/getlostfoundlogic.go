package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLostFoundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLostFoundLogic {
	return &GetLostFoundLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetLostFoundLogic) GetLostFound(req *types.GetLostFoundReq) (*types.GetLostFoundResp, error) {
	resp, err := l.svcCtx.LostFoundServiceRpc.GetLostFound(l.ctx, &communityv1.GetLostFoundRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetLostFoundResp{Item: toLostFoundItemInfo(resp.Item)}, nil
}
