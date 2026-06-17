package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNoticeModerationStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNoticeModerationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNoticeModerationStatusLogic {
	return &UpdateNoticeModerationStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNoticeModerationStatusLogic) UpdateNoticeModerationStatus(in *communityv1.UpdateModerationStatusRequest) (*communityv1.UpdateModerationStatusResponse, error) {
	if err := l.svcCtx.NoticeModel.UpdateModerationStatus(l.ctx, in.Id, int64(in.ModerationStatus)); err != nil {
		l.Errorf("UpdateNoticeModerationStatus failed: %v", err)
		return nil, err
	}
	l.Infof("UpdateNoticeModerationStatus: id=%d, status=%d", in.Id, in.ModerationStatus)
	return &communityv1.UpdateModerationStatusResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
