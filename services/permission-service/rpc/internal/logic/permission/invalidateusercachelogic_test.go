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

// TestInvalidateUserCache_PositiveUserId 覆盖正常路径：userId>0 触发缓存失效（miniredis）。
func TestInvalidateUserCache_PositiveUserId(t *testing.T) {
	client, _ := setupMiniRedis(t)
	sc := &svc.ServiceContext{RedisClient: client}
	l := NewInvalidateUserCacheLogic(context.Background(), sc)
	resp, err := l.InvalidateUserCache(&permissionv1.InvalidateUserCacheRequest{UserId: 42})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
}
