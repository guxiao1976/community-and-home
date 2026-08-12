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

type fakeDeleteNoticeServiceRpc struct {
	communityv1.NoticeServiceClient
	gotCtx context.Context
	resp   *communityv1.DeleteNoticeResponse
}

func (f *fakeDeleteNoticeServiceRpc) DeleteNotice(ctx context.Context, in *communityv1.DeleteNoticeRequest, opts ...grpc.CallOption) (*communityv1.DeleteNoticeResponse, error) {
	f.gotCtx = ctx
	return f.resp, nil
}

// API Delete handler 注入 JWT 身份 metadata（T4.3，超范围不删），
// 并将 RPC 业务错误（080006）透出。多视角评审修复：此前无任何 DeleteNotice 测试。
//
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
// SEE: [[testing-discipline]] — API 写路径与 UpsertContacts 同模板（InjectJWT + SurfaceBaseError）
func TestDeleteNotice_InjectJWTAndSurfaceBaseError(t *testing.T) {
	t.Run("注入 JWT 身份 metadata + 成功路径", func(t *testing.T) {
		fake := &fakeDeleteNoticeServiceRpc{resp: &communityv1.DeleteNoticeResponse{Base: responsex.NewBaseResp()}}
		svcCtx := &svc.ServiceContext{NoticeServiceRpc: fake}
		reqCtx := context.WithValue(context.Background(), "user_id", json.Number("42"))

		l := NewDeleteNoticeLogic(reqCtx, svcCtx)
		err := l.DeleteNotice(&types.DeleteNoticeReq{Id: 1})
		require.NoError(t, err, "成功 Base → 无错误")

		md, ok := metadata.FromOutgoingContext(fake.gotCtx)
		require.True(t, ok, "出站 metadata 必须存在")
		assert.Contains(t, md.Get("user_id"), "42", "delete handler 必须注入 JWT 身份")
	})

	t.Run("rpc 数据权限拒绝 080006 → 透出给客户端", func(t *testing.T) {
		fake := &fakeDeleteNoticeServiceRpc{
			resp: &communityv1.DeleteNoticeResponse{Base: responsex.NewBaseRespWithError(80006, "目标小区超出发布者数据范围")},
		}
		svcCtx := &svc.ServiceContext{NoticeServiceRpc: fake}
		reqCtx := context.WithValue(context.Background(), "user_id", json.Number("42"))

		l := NewDeleteNoticeLogic(reqCtx, svcCtx)
		err := l.DeleteNotice(&types.DeleteNoticeReq{Id: 1})
		require.Error(t, err)
		ce := errx.FromError(err)
		require.NotNil(t, ce)
		assert.Equal(t, 80006, ce.Code, "越权 Delete → API 透出 080006")
	})
}
