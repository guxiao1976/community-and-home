package lostfound

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListLostFoundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLostFoundLogic {
	return &ListLostFoundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLostFoundLogic) ListLostFound(in *communityv1.ListLostFoundRequest) (*communityv1.ListLostFoundResponse, error) {
	// 读过滤（T4.6 / REQ-1.6）：GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表。
	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, in.GetCommunityId())
	if err != nil {
		l.Errorf("ListLostFound: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		return &communityv1.ListLostFoundResponse{
			Base:  responsex.NewBaseResp(),
			Items: []*communityv1.LostFoundItem{},
			Total: 0,
		}, nil
	}

	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
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
	typStr := typeToString(in.Type)

	list, total, err := l.svcCtx.LostFoundItemModel.FindList(l.ctx, in.CommunityId, typStr, offset, limit)
	if err != nil {
		l.Errorf("ListLostFound: query failed: %v", err)
		return nil, err
	}

	items := make([]*communityv1.LostFoundItem, 0, len(list))
	for _, it := range list {
		items = append(items, toProtoLostFoundItem(it))
	}

	return &communityv1.ListLostFoundResponse{
		Base:  responsex.NewBaseResp(),
		Items: items,
		Total: total,
	}, nil
}
