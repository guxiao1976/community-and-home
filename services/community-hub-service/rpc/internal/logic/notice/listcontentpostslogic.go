package notice

import (
	"context"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListContentPostsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListContentPostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContentPostsLogic {
	return &ListContentPostsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListContentPosts 列表读路径：scope 过滤（GLOBAL/LIMITED/EMPTY）+ section_code/role 筛选 + 完整性谓词
// + 可选时间窗口（since_days）。
//
// since_days 校验（REQ-NTW-2 场景 3，r2-5）：<0 或 >365 → 080005（以 Base 错误返回，非 gRPC err）；
// ==0（缺省）→ 不过滤，PC 管理列表行为不变（场景 2）；>0 → WithTimeWindow(now-since_days) 窗口谓词（D10）。
// FilterAllowed false → 空列表（不泄露）；FindListByCommunity（JOIN scope + IsReviewComplete 谓词）；
// order by is_pinned desc, published_at desc（NULLS LAST）；分页。
// 响应每条 community_id=请求小区（scope 派生，不读弃用列）。
func (l *ListContentPostsLogic) ListContentPosts(in *communityv1.ListContentPostsRequest) (*communityv1.ListContentPostsResponse, error) {
	// 参数校验先于业务逻辑（fail-fast）：非法 since_days 无论数据范围如何一律 080005。
	if sinceDays := in.GetSinceDays(); sinceDays < 0 || sinceDays > 365 {
		return &communityv1.ListContentPostsResponse{
			Base:  responsex.NewBaseRespWithError(scope.CodeInvalidParam, "since_days 超出有效范围 1..365"),
			Posts: []*communityv1.ContentPost{},
			Total: 0,
		}, nil
	}

	// 读过滤（T4.6 / REQ-CPR-1）：GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表
	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, in.GetCommunityId())
	if err != nil {
		l.Errorf("ListContentPosts: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		return &communityv1.ListContentPostsResponse{
			Base:  responsex.NewBaseResp(),
			Posts: []*communityv1.ContentPost{},
			Total: 0,
		}, nil
	}

	page := in.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := in.GetPageSize()
	maxPageSize := int32(100)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "community.list.max_page_size"); err == nil {
			maxPageSize = int32(v)
		}
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = 10
	}

	offset := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	// role 筛选：枚举→DB 列映射收敛 helper.go 单源（评审 data-model v4 I2，与写侧 PublishRoleToString 同字符串集合）
	roleStr := ContentPostRoleToString(in.GetRole())

	// 时间窗口传参（D10）：since_days>0 → 下界 now-since_days（含端点，上界 now 由模型层填充）；==0 不传选项。
	var opts []model.ContentPostListOption
	if sd := in.GetSinceDays(); sd > 0 {
		opts = append(opts, model.WithTimeWindow(time.Now().AddDate(0, 0, -int(sd))))
	}

	list, total, err := l.svcCtx.ContentPostModel.FindListByCommunity(l.ctx, in.GetCommunityId(), in.GetSectionCode(), roleStr, offset, limit, opts...)
	if err != nil {
		l.Errorf("ListContentPosts: query failed: %v", err)
		return nil, err
	}

	posts := make([]*communityv1.ContentPost, 0, len(list))
	for _, n := range list {
		posts = append(posts, toProtoContentPost(n, in.GetCommunityId()))
	}

	return &communityv1.ListContentPostsResponse{
		Base:  responsex.NewBaseResp(),
		Posts: posts,
		Total: total,
	}, nil
}
