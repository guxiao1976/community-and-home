package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMarqueeNoticesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMarqueeNoticesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarqueeNoticesLogic {
	return &GetMarqueeNoticesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetMarqueeNotices 跑马灯代理（透传 community_id，响应 ContentPostMarqueeItemInfo）。
func (l *GetMarqueeNoticesLogic) GetMarqueeNotices(req *types.GetMarqueeNoticesReq) (*types.GetMarqueeNoticesResp, error) {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.ContentPostServiceRpc.GetMarqueeNotices(callCtx, &communityv1.GetMarqueeNoticesRequest{
		CommunityId: req.CommunityId,
	})
	if err != nil {
		return nil, err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return nil, responsex.ToError(resp.GetBase())
	}

	items := make([]types.ContentPostMarqueeItemInfo, 0, len(resp.Items))
	for _, it := range resp.Items {
		items = append(items, types.ContentPostMarqueeItemInfo{Id: it.Id, Title: it.Title})
	}
	return &types.GetMarqueeNoticesResp{Items: items}, nil
}
