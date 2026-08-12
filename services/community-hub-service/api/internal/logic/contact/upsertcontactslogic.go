package contact

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
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

// UpsertContacts 批量维护便民联络。
//
// 注入 JWT 身份 metadata 供 rpc 层 AssertPublishScope 数据权限校验（T4.4），
// 并将 rpc 业务错误（如 080006）透出给客户端。
func (l *UpsertContactsLogic) UpsertContacts(req *types.UpsertContactsReq) error {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return err
	}

	entries := make([]*communityv1.ContactEntry, 0, len(req.Contacts))
	for _, c := range req.Contacts {
		entries = append(entries, &communityv1.ContactEntry{
			Category: communityv1.ContactCategory(c.Category),
			Name:     c.Name,
			Phone:    c.Phone,
		})
	}

	resp, err := l.svcCtx.ContactServiceRpc.UpsertContacts(callCtx, &communityv1.UpsertContactsRequest{
		CommunityId: req.CommunityId,
		Contacts:    entries,
	})
	if err != nil {
		return err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return responsex.ToError(resp.GetBase())
	}
	return nil
}
