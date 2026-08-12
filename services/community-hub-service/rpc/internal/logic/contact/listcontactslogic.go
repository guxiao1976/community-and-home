package contact

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListContactsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListContactsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContactsLogic {
	return &ListContactsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListContactsLogic) ListContacts(in *communityv1.ListContactsRequest) (*communityv1.ListContactsResponse, error) {
	// 读过滤（T4.6 / REQ-1.6）：GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表。
	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, in.GetCommunityId())
	if err != nil {
		l.Errorf("ListContacts: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		return &communityv1.ListContactsResponse{
			Base:     responsex.NewBaseResp(),
			Contacts: []*communityv1.Contact{},
		}, nil
	}

	list, err := l.svcCtx.CommunityContactModel.FindByCommunityId(l.ctx, in.CommunityId)
	if err != nil {
		l.Errorf("ListContacts: query failed: %v", err)
		return nil, err
	}

	contacts := make([]*communityv1.Contact, 0, len(list))
	for _, c := range list {
		contacts = append(contacts, &communityv1.Contact{
			Id:          c.Id,
			CommunityId: c.CommunityId,
			Category:    stringToCategory(c.Category),
			Name:        c.Name,
			Phone:       c.Phone,
			SortOrder:   c.SortOrder,
		})
	}

	return &communityv1.ListContactsResponse{
		Base:     responsex.NewBaseResp(),
		Contacts: contacts,
	}, nil
}
