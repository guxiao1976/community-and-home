package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// =============================================================================
// Redis 角色缓存 Cache-Aside 测试
// 设计文档参考: auth-design.md §五
// =============================================================================

func TestRoleCache_CacheHit(t *testing.T) {
	// R-C-02: 缓存命中 → 不调用 gRPC
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(4001)

	cachedRoles := []roleEntry{{R: "owner", C: 1001}}
	data, _ := json.Marshal(cachedRoles)
	mr.Set(fmt.Sprintf("auth:roles:%d", userId), string(data))

	gRPCCalls := 0
	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			gRPCCalls++
			return &userv1.GetUserRolesResponse{Base: okResp()}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	// 两次调用都应命中缓存
	for i := 0; i < 2; i++ {
		roles, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
		require.NoError(t, err)
		assert.Len(t, roles, 1)
		assert.Equal(t, "owner", roles[0].R)
		assert.Equal(t, int64(1001), roles[0].C)
	}
	assert.Equal(t, 0, gRPCCalls, "缓存命中时不应调用 gRPC")
}

func TestRoleCache_CacheMiss(t *testing.T) {
	// R-C-01: 缓存 MISS → 穿透 gRPC → 回填 Redis
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(4002)

	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base: okResp(),
				Roles: []*userv1.MembershipRole{
					{RoleCode: "grid_worker", CommunityId: 1001, VerfStatus: 2},
					{RoleCode: "grid_worker", CommunityId: 1002, VerfStatus: 2},
				},
			}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	roles, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	require.NoError(t, err)
	assert.Len(t, roles, 2)

	cacheKey := fmt.Sprintf("auth:roles:%d", userId)
	assert.True(t, mr.Exists(cacheKey), "缓存应回填到 Redis")
}

func TestRoleCache_CorruptJSON(t *testing.T) {
	// R-C-03: 缓存 JSON 损坏 → 穿透兜底
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(4003)

	mr.Set(fmt.Sprintf("auth:roles:%d", userId), "{corrupt_json{{{")

	gRPCCalled := false
	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			gRPCCalled = true
			return &userv1.GetUserRolesResponse{
				Base:  okResp(),
				Roles: []*userv1.MembershipRole{{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2}},
			}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	roles, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	require.NoError(t, err)
	assert.True(t, gRPCCalled, "缓存 JSON 损坏时应穿透 gRPC")
	assert.Len(t, roles, 1)
}

func TestRoleCache_TTLExpiry(t *testing.T) {
	// R-C-04: 缓存 TTL 到期自然失效
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(4004)

	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base:  okResp(),
				Roles: []*userv1.MembershipRole{{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2}},
			}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	// 首次调用 → 回填缓存
	_, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	require.NoError(t, err)
	cacheKey := fmt.Sprintf("auth:roles:%d", userId)
	assert.True(t, mr.Exists(cacheKey))

	// 快进 301 秒（超过 300s TTL）
	mr.FastForward(301 * time.Second)
	assert.False(t, mr.Exists(cacheKey), "缓存 TTL 到期应自动删除")

	// 再次调用 → MISS → 穿透并回填
	roles, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.True(t, mr.Exists(cacheKey))
}

func TestRoleCache_EmptyRoles(t *testing.T) {
	// 用户无角色 → roles=[]，缓存空数组
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(4005)

	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{Base: okResp(), Roles: []*userv1.MembershipRole{}}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	roles, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	require.NoError(t, err)
	assert.Empty(t, roles)

	cacheKey := fmt.Sprintf("auth:roles:%d", userId)
	assert.True(t, mr.Exists(cacheKey))
	cachedStr, _ := mr.Get(cacheKey)
	assert.Equal(t, "[]", cachedStr)
}

func TestRoleCache_gRPCDown(t *testing.T) {
	// gRPC 不可用时返回 error
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(4006)

	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return nil, fmt.Errorf("gRPC connection refused")
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	_, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GetUserRoles")
}

func TestRoleCache_MerchantRole(t *testing.T) {
	// 商家角色 c=0
	mr, rdb := setupRedis(t)
	setupTestCrypto(t)
	mr.FlushAll()
	userId := int64(4007)

	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			return &userv1.GetUserRolesResponse{
				Base:  okResp(),
				Roles: []*userv1.MembershipRole{{RoleCode: "merchant", CommunityId: 0, VerfStatus: 2}},
			}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	roles, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, "merchant", roles[0].R)
	assert.Equal(t, int64(0), roles[0].C)
}

func TestRoleCache_OnlyVerifiedRoles(t *testing.T) {
	// 只查询 verf_status=2 的角色
	_, rdb := setupRedis(t)
	setupTestCrypto(t)
	userId := int64(4008)

	verified := false
	userRpc := &mockUserServiceClient{
		GetUserRolesFn: func(ctx context.Context, in *userv1.GetUserRolesRequest, opts ...grpc.CallOption) (*userv1.GetUserRolesResponse, error) {
			if in.VerfStatus != nil && *in.VerfStatus == 2 {
				verified = true
			}
			return &userv1.GetUserRolesResponse{
				Base:  okResp(),
				Roles: []*userv1.MembershipRole{{RoleCode: "owner", CommunityId: 1001, VerfStatus: 2}},
			}, nil
		},
	}
	svcCtx := newTestServiceContext(t, rdb, userRpc, defaultMockCredentialModel(userId))

	_, err := getUserRolesWithCache(context.Background(), svcCtx, userId)
	require.NoError(t, err)
	assert.True(t, verified, "应传 verf_status=2 查询已认证角色")
}
