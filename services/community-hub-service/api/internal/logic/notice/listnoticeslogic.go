package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListNoticesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListNoticesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNoticesLogic {
	return &ListNoticesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListNoticesLogic) ListNotices(req *types.ListNoticesReq) (*types.ListNoticesResp, error) {
	// 注入 JWT 身份 metadata，供 rpc 层 GetDataScopes 读过滤（T4.6）
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.NoticeServiceRpc.ListNotices(callCtx, &communityv1.ListNoticesRequest{
		CommunityId: req.CommunityId,
		Role:        communityv1.NoticeRole(req.Role),
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	notices := make([]types.NoticeInfo, 0, len(resp.Notices))
	for _, n := range resp.Notices {
		notices = append(notices, toNoticeInfo(n))
	}

	return &types.ListNoticesResp{
		Notices: notices,
		Total:   resp.Total,
	}, nil
}
