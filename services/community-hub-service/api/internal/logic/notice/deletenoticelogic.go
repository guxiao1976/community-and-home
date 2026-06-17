package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteNoticeLogic {
	return &DeleteNoticeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteNoticeLogic) DeleteNotice(req *types.DeleteNoticeReq) error {
	_, err := l.svcCtx.NoticeServiceRpc.DeleteNotice(l.ctx, &communityv1.DeleteNoticeRequest{
		Id: req.Id,
	})
	return err
}
