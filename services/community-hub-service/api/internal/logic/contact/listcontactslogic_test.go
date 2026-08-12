package contact

import (
	"context"
	"encoding/json"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type fakeListContactsServiceRpc struct {
	communityv1.ContactServiceClient
	gotCtx context.Context
	resp   *communityv1.ListContactsResponse
}

func (f *fakeListContactsServiceRpc) ListContacts(ctx context.Context, in *communityv1.ListContactsRequest, opts ...grpc.CallOption) (*communityv1.ListContactsResponse, error) {
	f.gotCtx = ctx
	return f.resp, nil
}

// API List handler 注入 JWT 身份 metadata（T4.6），供 RPC 层 GetDataScopes 读过滤。
// 多视角评审修复：ListContacts 的 CallCtx 注入块无任何测试引用。
//
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
// SEE: [[testing-discipline]] — API 层与 UpsertContacts 同路径注入身份（T4.6 一致）
func TestListContacts_InjectsIdentity(t *testing.T) {
	fake := &fakeListContactsServiceRpc{
		resp: &communityv1.ListContactsResponse{Base: responsex.NewBaseResp()},
	}
	svcCtx := &svc.ServiceContext{ContactServiceRpc: fake}
	reqCtx := context.WithValue(context.Background(), "user_id", json.Number("42"))

	l := NewListContactsLogic(reqCtx, svcCtx)
	resp, err := l.ListContacts(&types.ListContactsReq{CommunityId: 100})
	require.NoError(t, err)
	assert.Empty(t, resp.Contacts, "RPC 返回透传")

	md, ok := metadata.FromOutgoingContext(fake.gotCtx)
	require.True(t, ok, "出站 metadata 必须存在")
	assert.Contains(t, md.Get("user_id"), "42", "list handler 必须注入 JWT 身份（与 UpsertContacts 一致）")
}
