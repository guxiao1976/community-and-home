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
	// 仅 080006 精确匹配：低于/高于该码的 CodeError 均不算数据权限拒绝
	assert.False(t, IsPublishScopeDenied(errx.NewCodeError(CodePublishScopeDenied-1, "below")))
	assert.False(t, IsPublishScopeDenied(errx.NewCodeError(CodePublishScopeDenied+1, "above")))
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

	// 负数 userID（非法身份）→ 不因 userID<=0 短路拒绝，仍走 checkScope（身份解析失败由调用方决定）
	negCtx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(UserIDMetadataKey, "-1"))
	var gotNegUserID int64
	neg := &fakePermClient{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			gotNegUserID = in.GetUserId()
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		},
	}
	resp, err = CheckPublishScope(negCtx, neg, 100)
	require.NoError(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, int64(-1), gotNegUserID)
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

	// gRPC 传输错误（非 080006）→ checkScope 原样传播 (nil, err)
	transportErr := &fakePermClient{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return nil, errors.New("permission rpc unavailable")
		},
	}
	resp, err = CheckSystemPublishScope(context.Background(), transportErr, 100)
	require.Error(t, err)
	assert.Nil(t, resp, "传输错误路径不构造 denyResp，直接原样传播")
}
