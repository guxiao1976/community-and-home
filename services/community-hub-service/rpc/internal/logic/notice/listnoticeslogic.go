package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListNoticesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNoticesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNoticesLogic {
	return &ListNoticesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNoticesLogic) ListNotices(in *communityv1.ListNoticesRequest) (*communityv1.ListNoticesResponse, error) {
	// 读过滤（T4.6 / REQ-1.6）：GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表。
	// 空列表在逻辑层返回，SQL 不拼空 IN 子句。
	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, in.GetCommunityId())
	if err != nil {
		l.Errorf("ListNotices: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		return &communityv1.ListNoticesResponse{
			Base:    responsex.NewBaseResp(),
			Notices: []*communityv1.Notice{},
			Total:   0,
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
	roleStr := roleToString(in.Role)

	list, total, err := l.svcCtx.NoticeModel.FindList(l.ctx, in.CommunityId, roleStr, offset, limit)
	if err != nil {
		l.Errorf("ListNotices: query failed: %v", err)
		return nil, err
	}

	notices := make([]*communityv1.Notice, 0, len(list))
	for _, n := range list {
		notices = append(notices, toProtoNotice(n))
	}

	return &communityv1.ListNoticesResponse{
		Base:    responsex.NewBaseResp(),
		Notices: notices,
		Total:   total,
	}, nil
}
