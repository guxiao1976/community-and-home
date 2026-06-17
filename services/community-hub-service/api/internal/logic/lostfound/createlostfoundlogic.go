package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLostFoundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLostFoundLogic {
	return &CreateLostFoundLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateLostFoundLogic) CreateLostFound(req *types.CreateLostFoundReq) (*types.CreateLostFoundResp, error) {
	resp, err := l.svcCtx.LostFoundServiceRpc.CreateLostFound(l.ctx, &communityv1.CreateLostFoundRequest{
		CommunityId:  req.CommunityId,
		Type:         communityv1.LostFoundType(req.Type),
		Title:        req.Title,
		Description:  req.Description,
		ImageUrls:    req.ImageUrls,
		ContactPhone: req.ContactPhone,
		PublisherId:  req.PublisherId,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateLostFoundResp{Id: resp.Id}, nil
}
