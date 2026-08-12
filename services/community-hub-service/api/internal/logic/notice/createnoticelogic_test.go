package notice

import (
	"context"
	"encoding/json"
	"strconv"
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

// fakeNoticeServiceRpc 捕获 CreateNotice 请求与调用上下文
type fakeNoticeServiceRpc struct {
	communityv1.NoticeServiceClient
	got    *communityv1.CreateNoticeRequest
	gotCtx context.Context
}

func (f *fakeNoticeServiceRpc) CreateNotice(ctx context.Context, in *communityv1.CreateNoticeRequest, opts ...grpc.CallOption) (*communityv1.CreateNoticeResponse, error) {
	f.got = in
	f.gotCtx = ctx
	return &communityv1.CreateNoticeResponse{Base: responsex.NewBaseResp(), Id: 4542136688377323520}, nil
}

func noticeJSONNum(n int64) string { return strconv.FormatInt(n, 10) }

// SEE: [[verify-api-before-calling]] — publisher_id 一律取 JWT，忽略客户端 body 值
func TestCreateNotice_PublisherId_TakenFromJWT(t *testing.T) {
	tests := []struct {
		name    string
		ctxUID  int64
		bodyUID int64 // 客户端伪造的 publisher_id
		wantUID int64
	}{
		{
			name:    "伪造 publisher_id → 落库为 JWT 身份",
			ctxUID:  2001,
			bodyUID: 888888,
			wantUID: 2001,
		},
		{
			name:    "合法发布者身份正确落库",
			ctxUID:  2002,
			bodyUID: 2002,
			wantUID: 2002,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqCtx := context.WithValue(context.Background(), "user_id", json.Number(noticeJSONNum(tc.ctxUID)))
			fake := &fakeNoticeServiceRpc{}
			svcCtx := &svc.ServiceContext{NoticeServiceRpc: fake}

			l := NewCreateNoticeLogic(reqCtx, svcCtx)
			resp, err := l.CreateNotice(&types.CreateNoticeReq{
				CommunityId: 300,
				Title:       "停水通知",
				Content:     "本周六停水",
				Role:        1,
				Publisher:   "物业",
				PublisherId: tc.bodyUID, // 伪造值，必须被忽略
			})
			require.NoError(t, err)
			require.Equal(t, int64(4542136688377323520), resp.Id)

			require.NotNil(t, fake.got, "RPC 必须被调用")
			assert.Equal(t, tc.wantUID, fake.got.PublisherId, "publisher_id 必须取 JWT 身份，忽略客户端 body 值")

			md, ok := metadata.FromOutgoingContext(fake.gotCtx)
			assert.True(t, ok, "出站 metadata 必须存在")
			assert.Contains(t, md.Get("user_id"), noticeJSONNum(tc.ctxUID))
		})
	}
}
