package lostfound

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

// fakeLostFoundServiceRpc 捕获 CreateLostFound 请求与调用上下文
type fakeLostFoundServiceRpc struct {
	communityv1.LostFoundServiceClient
	got    *communityv1.CreateLostFoundRequest
	gotCtx context.Context
}

func (f *fakeLostFoundServiceRpc) CreateLostFound(ctx context.Context, in *communityv1.CreateLostFoundRequest, opts ...grpc.CallOption) (*communityv1.CreateLostFoundResponse, error) {
	f.got = in
	f.gotCtx = ctx
	return &communityv1.CreateLostFoundResponse{Base: responsex.NewBaseResp(), Id: 4542136688377323520}, nil
}

func jsonNum(n int64) string { return strconv.FormatInt(n, 10) }

// SEE: [[testing-discipline]] — 身份规范化行为，伪造 body 值必须被 JWT 覆盖
// SEE: [[verify-api-before-calling]] — publisher_id 一律取 JWT，忽略客户端 body 值
func TestCreateLostFound_PublisherId_TakenFromJWT(t *testing.T) {
	tests := []struct {
		name      string
		ctxUID    int64
		bodyUID   int64 // 客户端伪造的 publisher_id
		wantUID   int64
		wantComID int64
	}{
		{
			name:      "伪造 publisher_id → 落库为 JWT 身份",
			ctxUID:    1001,
			bodyUID:   999999, // 攻击者伪造的管理员 id
			wantUID:   1001,
			wantComID: 200,
		},
		{
			name:      "合法发布者身份正确落库",
			ctxUID:    1002,
			bodyUID:   1002, // 客户端传了与 JWT 一致的身份
			wantUID:   1002,
			wantComID: 201,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reqCtx := context.WithValue(context.Background(), "user_id", json.Number(jsonNum(tc.ctxUID)))
			fake := &fakeLostFoundServiceRpc{}
			svcCtx := &svc.ServiceContext{LostFoundServiceRpc: fake}

			l := NewCreateLostFoundLogic(reqCtx, svcCtx)
			resp, err := l.CreateLostFound(&types.CreateLostFoundReq{
				CommunityId:  tc.wantComID,
				Type:         1,
				Title:        "寻物",
				Description:  "黑色钱包",
				ContactPhone: "13800000000",
				PublisherId:  tc.bodyUID, // 伪造值，必须被忽略
			})
			require.NoError(t, err)
			require.Equal(t, int64(4542136688377323520), resp.Id)

			require.NotNil(t, fake.got, "RPC 必须被调用")
			assert.Equal(t, tc.wantUID, fake.got.PublisherId, "publisher_id 必须取 JWT 身份，忽略客户端 body 值")
			assert.Equal(t, tc.wantComID, fake.got.CommunityId)

			// JWT 身份必须注入出站 gRPC metadata（供 rpc 层 AssertPublishScope 使用）
			md, ok := metadata.FromOutgoingContext(fake.gotCtx)
			assert.True(t, ok, "出站 metadata 必须存在")
			assert.Contains(t, md.Get("user_id"), jsonNum(tc.ctxUID))
		})
	}
}
