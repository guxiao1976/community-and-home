package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNoticeLogic {
	return &GetNoticeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetNoticeLogic) GetNotice(req *types.GetNoticeReq) (*types.GetNoticeResp, error) {
	resp, err := l.svcCtx.NoticeServiceRpc.GetNotice(l.ctx, &communityv1.GetNoticeRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.GetNoticeResp{Notice: toNoticeInfo(resp.Notice)}, nil
}
