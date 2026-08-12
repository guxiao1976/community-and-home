package contact

import (
	"context"
	"fmt"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type fakePerm struct {
	permissionv1.PermissionServiceClient
	assertFn func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error)
}

func (f *fakePerm) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	return f.assertFn(ctx, in, opts...)
}

type fakeContactModel struct {
	model.CommunityContactModel
	deleteCalled bool
	findByCalled bool
	inserted     []*model.CommunityContact
}

func (f *fakeContactModel) DeleteByCommunityId(ctx context.Context, communityId int64) error {
	f.deleteCalled = true
	return nil
}

func (f *fakeContactModel) Insert(ctx context.Context, c *model.CommunityContact) (int64, error) {
	f.inserted = append(f.inserted, c)
	return c.Id, nil
}

func (f *fakeContactModel) FindByCommunityId(ctx context.Context, communityId int64) ([]*model.CommunityContact, error) {
	f.findByCalled = true
	return nil, nil
}

func contactCtxWithUserID(t *testing.T, uid int64) context.Context {
	t.Helper()
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("user_id", fmt.Sprintf("%d", uid)))
}

func contactDenyAll() *fakePerm {
	return &fakePerm{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{
				Base:    responsex.NewBaseRespWithError(60007, "目标小区超出发布者数据范围"),
				Allowed: false,
			}, nil
		},
	}
}

func contactAllowAll() *fakePerm {
	return &fakePerm{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		},
	}
}

// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
func TestUpsertContacts_ScopeDenied_OutOfScope(t *testing.T) {
	mdl := &fakeContactModel{}
	sc := &svc.ServiceContext{CommunityContactModel: mdl, PermissionClient: contactDenyAll()}

	l := NewUpsertContactsLogic(contactCtxWithUserID(t, 100), sc)
	resp, err := l.UpsertContacts(&communityv1.UpsertContactsRequest{
		CommunityId: 200, // 目标小区超出发布者数据范围
		Contacts: []*communityv1.ContactEntry{
			{Category: communityv1.ContactCategory_CONTACT_CATEGORY_WATER, Name: "自来水", Phone: "123"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "目标小区超范围 → 080006")
	assert.False(t, mdl.deleteCalled, "数据权限拒绝后不得删除旧数据")
	assert.Len(t, mdl.inserted, 0, "数据权限拒绝后不得插入")
}

func TestUpsertContacts_ScopeAllowed(t *testing.T) {
	mdl := &fakeContactModel{}
	sc := &svc.ServiceContext{CommunityContactModel: mdl, PermissionClient: contactAllowAll()}

	l := NewUpsertContactsLogic(contactCtxWithUserID(t, 100), sc)
	resp, err := l.UpsertContacts(&communityv1.UpsertContactsRequest{
		CommunityId: 100,
		Contacts: []*communityv1.ContactEntry{
			{Category: communityv1.ContactCategory_CONTACT_CATEGORY_WATER, Name: "自来水", Phone: "123"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, mdl.deleteCalled, "数据权限允许后先删旧数据")
	assert.Len(t, mdl.inserted, 1, "数据权限允许后插入新数据")
}
