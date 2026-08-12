package notice

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

type fakeGetNoticeServiceRpc struct {
	communityv1.NoticeServiceClient
	gotCtx context.Context
	resp   *communityv1.GetNoticeResponse
}

func (f *fakeGetNoticeServiceRpc) GetNotice(ctx context.Context, in *communityv1.GetNoticeRequest, opts ...grpc.CallOption) (*communityv1.GetNoticeResponse, error) {
	f.gotCtx = ctx
	return f.resp, nil
}

// API Get handler 补 CallCtx 注入（评审 CRITICAL：此前与 List 不一致，未带身份 metadata，
// RPC 层越权 Get 会漏网）——RPC 已补读过滤，此处必须让身份可达 RPC。
//
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
// SEE: [[testing-discipline]] — API 层与 List 同路径注入身份（T4.6 一致）
func TestGetNotice_InjectsIdentity(t *testing.T) {
	fake := &fakeGetNoticeServiceRpc{
		resp: &communityv1.GetNoticeResponse{Base: responsex.NewBaseResp(), Notice: &communityv1.Notice{Id: 1, CommunityId: 100}},
	}
	svcCtx := &svc.ServiceContext{NoticeServiceRpc: fake}
	reqCtx := context.WithValue(context.Background(), "user_id", json.Number("42"))

	l := NewGetNoticeLogic(reqCtx, svcCtx)
	resp, err := l.GetNotice(&types.GetNoticeReq{Id: 1})
	require.NoError(t, err)
	require.NotNil(t, resp.Notice, "允许读取时返回内容")

	md, ok := metadata.FromOutgoingContext(fake.gotCtx)
	require.True(t, ok, "出站 metadata 必须存在")
	assert.Contains(t, md.Get("user_id"), "42", "get handler 必须注入 JWT 身份（与 List 一致）")
}

// RPC 080006 数据权限拒绝（越权 Get）→ API 透出错误，不静默返回空。
func TestGetNotice_SurfacesScopeDenied(t *testing.T) {
	fake := &fakeGetNoticeServiceRpc{
		resp: &communityv1.GetNoticeResponse{Base: responsex.NewBaseRespWithError(80006, "目标小区超出发布者数据范围")},
	}
	svcCtx := &svc.ServiceContext{NoticeServiceRpc: fake}
	reqCtx := context.WithValue(context.Background(), "user_id", json.Number("42"))

	l := NewGetNoticeLogic(reqCtx, svcCtx)
	_, err := l.GetNotice(&types.GetNoticeReq{Id: 1})
	require.Error(t, err)
	ce := errx.FromError(err)
	require.NotNil(t, ce)
	assert.Equal(t, 80006, ce.Code, "越权 Get → API 透出 080006")
}
