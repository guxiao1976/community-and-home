package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLostFoundModerationStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateLostFoundModerationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLostFoundModerationStatusLogic {
	return &UpdateLostFoundModerationStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateLostFoundModerationStatusLogic) UpdateLostFoundModerationStatus(in *communityv1.UpdateModerationStatusRequest) (*communityv1.UpdateModerationStatusResponse, error) {
	if err := l.svcCtx.LostFoundItemModel.UpdateModerationStatus(l.ctx, in.Id, int64(in.ModerationStatus)); err != nil {
		l.Errorf("UpdateLostFoundModerationStatus failed: %v", err)
		return nil, err
	}
	l.Infof("UpdateLostFoundModerationStatus: id=%d, status=%d", in.Id, in.ModerationStatus)
	return &communityv1.UpdateModerationStatusResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
