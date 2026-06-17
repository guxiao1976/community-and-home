package contact

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpsertContactsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpsertContactsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertContactsLogic {
	return &UpsertContactsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpsertContactsLogic) UpsertContacts(req *types.UpsertContactsReq) error {
	entries := make([]*communityv1.ContactEntry, 0, len(req.Contacts))
	for _, c := range req.Contacts {
		entries = append(entries, &communityv1.ContactEntry{
			Category: communityv1.ContactCategory(c.Category),
			Name:     c.Name,
			Phone:    c.Phone,
		})
	}

	_, err := l.svcCtx.ContactServiceRpc.UpsertContacts(l.ctx, &communityv1.UpsertContactsRequest{
		CommunityId: req.CommunityId,
		Contacts:    entries,
	})
	return err
}
