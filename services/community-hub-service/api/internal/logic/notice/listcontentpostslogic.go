package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListContentPostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListContentPostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContentPostsLogic {
	return &ListContentPostsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListContentPosts 列表代理（R2 wire 兼容：响应 JSON 键保持 notices）。
//
// since_days 透传（REVISION r2-2，必须贯通）：缺省 0 → RPC 缺省不过滤，PC 管理列表行为不变；
// 漏传会导致移动端 30 天窗口静默失效。RPC 侧 080005 以 Base 错误返回（非 gRPC err），
// 经 responsex.ToError 上抛（禁止静默吞错，与 getcontentpostlogic.go 等既有模式一致）。
func (l *ListContentPostsLogic) ListContentPosts(req *types.ListContentPostsReq) (*types.ListContentPostsResp, error) {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.ContentPostServiceRpc.ListContentPosts(callCtx, &communityv1.ListContentPostsRequest{
		CommunityId: req.CommunityId,
		Role:        communityv1.ContentPostRole(req.Role),
		SectionCode: req.SectionCode,
		Page:        req.Page,
		PageSize:    req.PageSize,
		SinceDays:   req.SinceDays, // r2-2：form since_days → RPC since_days
	})
	if err != nil {
		return nil, err
	}
	if err := responsex.ToError(resp.GetBase()); err != nil {
		return nil, err
	}

	posts := make([]types.ContentPostInfo, 0, len(resp.Posts))
	for _, p := range resp.Posts {
		posts = append(posts, toContentPostInfo(p))
	}

	return &types.ListContentPostsResp{
		Notices: posts, // R2 wire 键保持 notices
		Total:   resp.Total,
	}, nil
}
