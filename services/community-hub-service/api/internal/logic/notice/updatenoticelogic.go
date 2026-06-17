package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNoticeLogic {
	return &UpdateNoticeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateNoticeLogic) UpdateNotice(req *types.UpdateNoticeReq) error {
	_, err := l.svcCtx.NoticeServiceRpc.UpdateNotice(l.ctx, &communityv1.UpdateNoticeRequest{
		Id:       req.Id,
		Title:    req.Title,
		Content:  req.Content,
		IsPinned: req.IsPinned,
	})
	return err
}
