package contact

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListContactsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListContactsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContactsLogic {
	return &ListContactsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListContactsLogic) ListContacts(req *types.ListContactsReq) (*types.ListContactsResp, error) {
	resp, err := l.svcCtx.ContactServiceRpc.ListContacts(l.ctx, &communityv1.ListContactsRequest{
		CommunityId: req.CommunityId,
	})
	if err != nil {
		return nil, err
	}

	contacts := make([]types.ContactInfo, 0, len(resp.Contacts))
	for _, c := range resp.Contacts {
		contacts = append(contacts, types.ContactInfo{
			Id:          c.Id,
			CommunityId: c.CommunityId,
			Category:    int32(c.Category),
			Name:        c.Name,
			Phone:       c.Phone,
			SortOrder:   c.SortOrder,
		})
	}

	return &types.ListContactsResp{Contacts: contacts}, nil
}
