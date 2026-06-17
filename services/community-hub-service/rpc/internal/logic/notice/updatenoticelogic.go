package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNoticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNoticeLogic {
	return &UpdateNoticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNoticeLogic) UpdateNotice(in *communityv1.UpdateNoticeRequest) (*communityv1.UpdateNoticeResponse, error) {
	// 校验存在
	_, err := l.svcCtx.NoticeModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Infof("UpdateNotice: not found id=%d", in.Id)
		return &communityv1.UpdateNoticeResponse{
			Base: responsex.NewBaseRespWithError(80001, "通知不存在"),
		}, nil
	}

	isPinned := int32(0)
	if in.IsPinned {
		isPinned = 1
	}

	if err := l.svcCtx.NoticeModel.Update(l.ctx, in.Id, in.Title, in.Content, isPinned); err != nil {
		l.Errorf("UpdateNotice: update failed: %v", err)
		return nil, err
	}

	l.Infof("UpdateNotice success: id=%d", in.Id)
	return &communityv1.UpdateNoticeResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
