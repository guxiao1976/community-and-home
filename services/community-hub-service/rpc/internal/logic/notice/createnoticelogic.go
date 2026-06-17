package notice

import (
	"context"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNoticeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNoticeLogic {
	return &CreateNoticeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateNoticeLogic) CreateNotice(in *communityv1.CreateNoticeRequest) (*communityv1.CreateNoticeResponse, error) {
	if in.Title == "" || in.Content == "" {
		return &communityv1.CreateNoticeResponse{
			Base: responsex.NewBaseRespWithError(80005, "标题和内容不能为空"),
		}, nil
	}

	noticeRole := roleToString(in.Role)
	id := snowflake.NextID()

	n := &model.Notice{
		Id:          id,
		CommunityId: in.CommunityId,
		Title:       in.Title,
		Content:     in.Content,
		Role:        noticeRole,
		Publisher:   in.Publisher,
		PublisherId: &in.PublisherId,
		IsPinned:    0,
		PublishedAt: time.Now(),
	}

	if _, err := l.svcCtx.NoticeModel.Insert(l.ctx, n); err != nil {
		l.Errorf("CreateNotice: insert failed: %v", err)
		return nil, err
	}

	l.Infof("CreateNotice success: id=%d, communityId=%d", id, in.CommunityId)
	return &communityv1.CreateNoticeResponse{
		Base: responsex.NewBaseResp(),
		Id:   id,
	}, nil
}
