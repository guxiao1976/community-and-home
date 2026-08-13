package scope

import (
	"context"
	"errors"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TestUserIDFromCtx 覆盖 userctx.go 的身份解析（此前仅经 CheckPublishScope 间接覆盖）。
func TestUserIDFromCtx(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantUID int64
	}{
		{"无 metadata → 0", context.Background(), 0},
		{"user_id=42 → 42", metadata.NewIncomingContext(context.Background(), metadata.Pairs(UserIDMetadataKey, "42")), 42},
		{"非法 user_id=abc → 0", metadata.NewIncomingContext(context.Background(), metadata.Pairs(UserIDMetadataKey, "abc")), 0},
		{"user_id 缺失（仅其他 key）→ 0", metadata.NewIncomingContext(context.Background(), metadata.Pairs("x", "1")), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantUID, UserIDFromCtx(tc.ctx))
		})
	}
}

func TestIsPublishScopeDenied(t *testing.T) {
	assert.False(t, IsPublishScopeDenied(nil))
	assert.True(t, IsPublishScopeDenied(errx.NewCodeError(CodePublishScopeDenied, "denied")))
	assert.False(t, IsPublishScopeDenied(errors.New("generic error")))
}

func TestDenyBase(t *testing.T) {
	b := DenyBase()
	require.NotNil(t, b)
	assert.Equal(t, int32(CodePublishScopeDenied), b.GetCode())
}

// TestCheckPublishScope 覆盖：无身份 fail-closed、有身份走 AssertCommunityScope（允许/拒绝）。
func TestCheckPublishScope(t *testing.T) {
	// 无身份 → fail-closed 拒绝
	resp, err := CheckPublishScope(context.Background(), &fakePermClient{}, 100)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(CodePublishScopeDenied), resp.GetCode())

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(UserIDMetadataKey, "42"))

	// 有身份 + 允许 → (nil, nil)
	allow := &fakePermClient{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			assert.Equal(t, int64(42), in.GetUserId())
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		},
	}
	resp, err = CheckPublishScope(ctx, allow, 100)
	require.NoError(t, err)
	assert.Nil(t, resp)

	// 有身份 + 拒绝 → (denyResp, nil)
	deny := &fakePermClient{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseRespWithError(60007, "denied"), Allowed: false}, nil
		},
	}
	resp, err = CheckPublishScope(ctx, deny, 100)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(CodePublishScopeDenied), resp.GetCode())
}

// TestCheckSystemPublishScope 系统身份（SystemUserID=0）不经 fail-closed 分支，直接走 AssertCommunityScope。
func TestCheckSystemPublishScope(t *testing.T) {
	allow := &fakePermClient{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			assert.Equal(t, SystemUserID, in.GetUserId(), "系统身份应为 0")
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		},
	}
	resp, err := CheckSystemPublishScope(context.Background(), allow, 100)
	require.NoError(t, err)
	assert.Nil(t, resp)
}
