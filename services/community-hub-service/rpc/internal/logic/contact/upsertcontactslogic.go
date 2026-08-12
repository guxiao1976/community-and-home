package contact

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpsertContactsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpsertContactsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpsertContactsLogic {
	return &UpsertContactsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpsertContactsLogic) UpsertContacts(in *communityv1.UpsertContactsRequest) (*communityv1.UpsertContactsResponse, error) {
	// 数据权限（T4.4）：落库前 AssertPublishScope(目标 community_id)。
	denyResp, err := scope.CheckPublishScope(l.ctx, l.svcCtx.PermissionClient, in.GetCommunityId())
	if err != nil {
		l.Errorf("UpsertContacts: assert publish scope failed: %v", err)
		return nil, err
	}
	if denyResp != nil {
		return &communityv1.UpsertContactsResponse{Base: denyResp}, nil
	}

	// 先删除旧数据，再插入新数据
	if err := l.svcCtx.CommunityContactModel.DeleteByCommunityId(l.ctx, in.CommunityId); err != nil {
		l.Errorf("UpsertContacts: delete old failed: %v", err)
		return nil, err
	}

	for i, entry := range in.Contacts {
		c := &model.CommunityContact{
			Id:          snowflake.NextID(),
			CommunityId: in.CommunityId,
			Category:    categoryToString(entry.Category),
			Name:        entry.Name,
			Phone:       entry.Phone,
			SortOrder:   int32(i),
		}
		if _, err := l.svcCtx.CommunityContactModel.Insert(l.ctx, c); err != nil {
			l.Errorf("UpsertContacts: insert failed: %v", err)
			return nil, err
		}
	}

	l.Infof("UpsertContacts success: communityId=%d, count=%d", in.CommunityId, len(in.Contacts))
	return &communityv1.UpsertContactsResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
