package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListLostFoundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLostFoundLogic {
	return &ListLostFoundLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListLostFoundLogic) ListLostFound(req *types.ListLostFoundReq) (*types.ListLostFoundResp, error) {
	// 注入 JWT 身份 metadata，供 rpc 层 GetDataScopes 读过滤（T4.6）
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.LostFoundServiceRpc.ListLostFound(callCtx, &communityv1.ListLostFoundRequest{
		CommunityId: req.CommunityId,
		Type:        communityv1.LostFoundType(req.Type),
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.LostFoundItemInfo, 0, len(resp.Items))
	for _, it := range resp.Items {
		items = append(items, toLostFoundItemInfo(it))
	}

	return &types.ListLostFoundResp{
		Items: items,
		Total: resp.Total,
	}, nil
}
