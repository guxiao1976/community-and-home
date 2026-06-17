package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNoticeLogic {
	return &CreateNoticeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateNoticeLogic) CreateNotice(req *types.CreateNoticeReq) (*types.CreateNoticeResp, error) {
	resp, err := l.svcCtx.NoticeServiceRpc.CreateNotice(l.ctx, &communityv1.CreateNoticeRequest{
		CommunityId: req.CommunityId,
		Title:       req.Title,
		Content:     req.Content,
		Role:        communityv1.NoticeRole(req.Role),
		Publisher:   req.Publisher,
		PublisherId: req.PublisherId,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateNoticeResp{Id: resp.Id}, nil
}
