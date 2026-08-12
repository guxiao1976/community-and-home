package contact

import (
	"context"
	"encoding/json"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type fakeContactRpc struct {
	communityv1.ContactServiceClient
	got    *communityv1.UpsertContactsRequest
	gotCtx context.Context
}

func (f *fakeContactRpc) UpsertContacts(ctx context.Context, in *communityv1.UpsertContactsRequest, opts ...grpc.CallOption) (*communityv1.UpsertContactsResponse, error) {
	f.got = in
	f.gotCtx = ctx
	return &communityv1.UpsertContactsResponse{Base: responsex.NewBaseResp()}, nil
}

// SEE: [[testing-discipline]] — API 层注入 JWT metadata + rpc 业务错误透出
func TestUpsertContacts_InjectJWTAndSurfaceBaseError(t *testing.T) {
	t.Run("注入 JWT 身份 metadata", func(t *testing.T) {
		reqCtx := context.WithValue(context.Background(), "user_id", json.Number("4542136688377323520"))
		fake := &fakeContactRpc{}
		sc := &svc.ServiceContext{ContactServiceRpc: fake}

		l := NewUpsertContactsLogic(reqCtx, sc)
		err := l.UpsertContacts(&types.UpsertContactsReq{
			CommunityId: 200,
			Contacts:    []types.ContactEntry{{Category: 1, Name: "自来水", Phone: "123"}},
		})
		require.NoError(t, err)
		require.NotNil(t, fake.got)
		assert.Equal(t, int64(200), fake.got.CommunityId)

		md, ok := metadata.FromOutgoingContext(fake.gotCtx)
		assert.True(t, ok, "出站 metadata 必须存在")
		assert.Contains(t, md.Get("user_id"), "4542136688377323520")
	})

	t.Run("rpc 数据权限拒绝 080006 → 透出给客户端", func(t *testing.T) {
		reqCtx := context.WithValue(context.Background(), "user_id", json.Number("1001"))
		fake := &fakeContactRpcDenied{}
		sc := &svc.ServiceContext{ContactServiceRpc: fake}

		l := NewUpsertContactsLogic(reqCtx, sc)
		err := l.UpsertContacts(&types.UpsertContactsReq{CommunityId: 200})
		require.Error(t, err)
		ce, ok := err.(*errx.CodeError)
		require.True(t, ok, "应为 CodeError，实际 %T", err)
		assert.Equal(t, 80006, ce.Code, "数据权限拒绝必须透出 080006")
	})
}

// fakeContactRpcDenied 模拟 rpc 层返回 080006 拒绝
type fakeContactRpcDenied struct {
	communityv1.ContactServiceClient
}

func (f *fakeContactRpcDenied) UpsertContacts(ctx context.Context, in *communityv1.UpsertContactsRequest, opts ...grpc.CallOption) (*communityv1.UpsertContactsResponse, error) {
	return &communityv1.UpsertContactsResponse{
		Base: responsex.NewBaseRespWithError(80006, "目标小区超出发布者数据范围"),
	}, nil
}
