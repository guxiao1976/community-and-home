package permission

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInvalidateUserCache_NonPositiveUserId 覆盖边界：userId<=0 提前返回，不碰 Redis。
func TestInvalidateUserCache_NonPositiveUserId(t *testing.T) {
	// RedisClient 为 nil：若误触 invalidateUserCaches 会 panic/报错，此用例守护提前返回分支。
	sc := &svc.ServiceContext{}
	l := NewInvalidateUserCacheLogic(context.Background(), sc)
	resp, err := l.InvalidateUserCache(&permissionv1.InvalidateUserCacheRequest{UserId: 0})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
}

// TestInvalidateUserCache_NegativeUserId 覆盖边界：userId<0 同样提前返回，不碰 Redis。
func TestInvalidateUserCache_NegativeUserId(t *testing.T) {
	sc := &svc.ServiceContext{}
	l := NewInvalidateUserCacheLogic(context.Background(), sc)
	resp, err := l.InvalidateUserCache(&permissionv1.InvalidateUserCacheRequest{UserId: -1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
}

// TestInvalidateUserCache_ZeroUserIdKeepsCache 覆盖 userId=0 提前返回：真实 Redis 下缓存不被删除
// 守护 `in.UserId <= 0` 边界（若被改成 `<`，userId=0 会误入失效逻辑删掉缓存）
func TestInvalidateUserCache_ZeroUserIdKeepsCache(t *testing.T) {
	client, _ := setupMiniRedis(t)
	ctx := context.Background()
	client.Set(ctx, "perm:user:0", "stale", 0)
	client.Set(ctx, "perm:scopes:0:community", "stale", 0)

	sc := &svc.ServiceContext{RedisClient: client}
	l := NewInvalidateUserCacheLogic(ctx, sc)
	resp, err := l.InvalidateUserCache(&permissionv1.InvalidateUserCacheRequest{UserId: 0})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())

	// userId<=0 提前返回 → 缓存应原样保留
	v, err := client.Get(ctx, "perm:user:0").Result()
	assert.NoError(t, err)
	assert.Equal(t, "stale", v, "userId=0 不应触发缓存失效")
	v, err = client.Get(ctx, "perm:scopes:0:community").Result()
	assert.NoError(t, err)
	assert.Equal(t, "stale", v)
}

// TestInvalidateUserCache_PositiveUserId 覆盖正常路径：userId>0 触发缓存失效（miniredis）。
func TestInvalidateUserCache_PositiveUserId(t *testing.T) {
	client, _ := setupMiniRedis(t)
	ctx := context.Background()
	// 预置缓存，验证正例确实删除（杀 in.UserId <= 0 → true 变异）
	client.Set(ctx, "perm:user:42", "stale", 0)
	client.Set(ctx, "perm:scopes:42:community", "stale", 0)

	sc := &svc.ServiceContext{RedisClient: client}
	l := NewInvalidateUserCacheLogic(ctx, sc)
	resp, err := l.InvalidateUserCache(&permissionv1.InvalidateUserCacheRequest{UserId: 42})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())

	// 正例应删除缓存（若 userId<=0 恒 true 提前返回，则缓存仍在 → 断言失败）
	_, err = client.Get(ctx, "perm:user:42").Result()
	assert.Error(t, err, "正例应删除 perm:user:42 缓存")
	_, err = client.Get(ctx, "perm:scopes:42:community").Result()
	assert.Error(t, err, "正例应删除 perm:scopes:42:community 缓存")
}
