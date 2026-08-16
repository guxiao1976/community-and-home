package notice

import (
	"context"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// MarqueeWindowDays 跑马灯时间窗口：15 天（REUSE:notice-D32，含端点）。
const MarqueeWindowDays = 15

// MarqueeLimit 跑马灯封顶条数。
const MarqueeLimit = 10

type GetMarqueeNoticesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMarqueeNoticesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMarqueeNoticesLogic {
	return &GetMarqueeNoticesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetMarqueeNotices 跑马灯（≤10、置顶优先、15 天含端点、仅审核通过且完整）。
// 板块固定 notice（无 section_code 入参，评审 INFO 1——跑马灯仅通知场景）。
func (l *GetMarqueeNoticesLogic) GetMarqueeNotices(in *communityv1.GetMarqueeNoticesRequest) (*communityv1.GetMarqueeNoticesResponse, error) {
	if in.CommunityId <= 0 {
		return l.err(scope.CodeInvalidParam, "缺少小区上下文"), nil
	}

	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, in.CommunityId)
	if err != nil {
		l.Errorf("GetMarqueeNotices: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		return &communityv1.GetMarqueeNoticesResponse{Base: responsex.NewBaseResp(), Items: []*communityv1.ContentPostMarqueeItem{}}, nil
	}

	since := time.Now().Add(-MarqueeWindowDays * 24 * time.Hour)
	list, err := l.svcCtx.ContentPostModel.FindMarquee(l.ctx, in.CommunityId, since, MarqueeLimit)
	if err != nil {
		l.Errorf("GetMarqueeNotices: query failed: %v", err)
		return nil, err
	}

	items := make([]*communityv1.ContentPostMarqueeItem, 0, len(list))
	for _, n := range list {
		items = append(items, &communityv1.ContentPostMarqueeItem{
			Id:    n.Id,
			Title: n.Title,
		})
	}
	return &communityv1.GetMarqueeNoticesResponse{
		Base:  responsex.NewBaseResp(),
		Items: items,
	}, nil
}

func (l *GetMarqueeNoticesLogic) err(code int32, msg string) *communityv1.GetMarqueeNoticesResponse {
	return &communityv1.GetMarqueeNoticesResponse{Base: responsex.NewBaseRespWithError(code, msg)}
}
