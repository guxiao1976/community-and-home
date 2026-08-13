package permission

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRevokeRole_Success 测试成功撤销角色
func TestRevokeRole_Success(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("DeleteByUserIdAndRoleId", mock.Anything, int64(1001), int64(1), "community", int64(100)).
		Return(nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewRevokeRoleLogic(context.Background(), svcCtx)

	// Execute
	scopeType := "community"
	scopeId := int64(100)
	resp, err := logic.RevokeRole(&permissionv1.RevokeRoleRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: &scopeType,
		ScopeId:   &scopeId,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)

	mockUserRole.AssertExpectations(t)
}

// TestRevokeRole_DefaultScope 测试使用默认 scope（不传参数）
func TestRevokeRole_DefaultScope(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockUserRole := new(MockUserRoleModel)
	// 默认 scopeType="community", scopeId=0
	mockUserRole.On("DeleteByUserIdAndRoleId", mock.Anything, int64(1001), int64(1), "community", int64(0)).
		Return(nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewRevokeRoleLogic(context.Background(), svcCtx)

	// Execute - 不传 ScopeType 和 ScopeId
	resp, err := logic.RevokeRole(&permissionv1.RevokeRoleRequest{
		UserId: 1001,
		RoleId: 1,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)

	mockUserRole.AssertExpectations(t)
}

// TestRevokeRole_CacheInvalidated 测试缓存失效
func TestRevokeRole_CacheInvalidated(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	// 预先设置缓存
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	redisClient.Set(ctx, "perm:user:1001", "cached_permissions", 0)
	redisClient.Set(ctx, "perm:scopes:1001:community", "cached_scopes_community", 0)
	redisClient.Set(ctx, "perm:scopes:1001:building", "cached_scopes_building", 0)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("DeleteByUserIdAndRoleId", mock.Anything, int64(1001), int64(1), "community", int64(100)).
		Return(nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redisClient,
	}

	logic := NewRevokeRoleLogic(ctx, svcCtx)

	// Execute
	scopeType := "community"
	scopeId := int64(100)
	_, err := logic.RevokeRole(&permissionv1.RevokeRoleRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: &scopeType,
		ScopeId:   &scopeId,
	})

	// Assert
	assert.NoError(t, err)

	// 验证所有相关缓存已被删除
	val, err := redisClient.Get(ctx, "perm:user:1001").Result()
	assert.Error(t, err) // redis.Nil
	assert.Empty(t, val)

	val, err = redisClient.Get(ctx, "perm:scopes:1001:community").Result()
	assert.Error(t, err)
	assert.Empty(t, val)

	val, err = redisClient.Get(ctx, "perm:scopes:1001:building").Result()
	assert.Error(t, err)
	assert.Empty(t, val)

	mockUserRole.AssertExpectations(t)
}

// TestRevokeRole_ScopesCacheInvalidated_GetDataScopesEmpty
// T1.6: Revoke 后 perm:scopes 缓存 DEL 生效 → GetDataScopes 立即重算为 EMPTY
// SEE: [[redis-cache-soft-delete]] — 失效收敛到 grant 变更处理器
func TestRevokeRole_ScopesCacheInvalidated_GetDataScopesEmpty(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	// 预先写入「旧」缓存（LIMITED），验证 Revoke 后会被清掉而非命中旧值
	redisClient.Set(ctx, "perm:scopes:1001:community", `{"state":"limited","ids":[100]}`, 0)
	redisClient.Set(ctx, "perm:user:1001", "stale", 0)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("DeleteByUserIdAndRoleId", mock.Anything, int64(1001), int64(1), "community", int64(100)).
		Return(nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redisClient,
	}

	// Revoke 触发缓存失效
	logic := NewRevokeRoleLogic(ctx, svcCtx)
	scopeType := "community"
	scopeId := int64(100)
	_, err := logic.RevokeRole(&permissionv1.RevokeRoleRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: &scopeType,
		ScopeId:   &scopeId,
	})
	assert.NoError(t, err)

	// 缓存已 DEL → GetDataScopes MISS → 重算
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1001)).Return([]*model.UserRoleWithInfo{}, nil)
	gd := NewGetDataScopesLogic(ctx, svcCtx)
	resp, err := gd.GetDataScopes(&permissionv1.GetDataScopesRequest{UserId: 1001, ScopeType: "community"})
	assert.NoError(t, err)
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, resp.State, "Revoke 后应立即 EMPTY（缓存 DEL 生效）")
	assert.Empty(t, resp.ScopeIds)

	mockUserRole.AssertExpectations(t)
}

// TestInvalidateUserCaches_ScanDelete — invalidateUserCaches 共享 helper：DEL perm:user + SCAN-DEL perm:scopes:{userId}:*
func TestInvalidateUserCaches_ScanDelete(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	// 预置多 scopeType 缓存 + 用户权限 Hash
	redisClient.Set(ctx, "perm:user:777", "stale", 0)
	redisClient.Set(ctx, "perm:scopes:777:community", `{"state":"limited","ids":[100]}`, 0)
	redisClient.Set(ctx, "perm:scopes:777:building", `{"state":"limited","ids":[200]}`, 0)
	redisClient.Set(ctx, "perm:scopes:778:community", "other-user-should-survive", 0) // 其他用户不删

	invalidateUserCaches(ctx, redisClient, 777)

	// perm:user 已删
	_, err := redisClient.Get(ctx, "perm:user:777").Result()
	assert.ErrorIs(t, err, redis.Nil)

	// 本用户所有 perm:scopes 已删（SCAN-DEL）
	_, err = redisClient.Get(ctx, "perm:scopes:777:community").Result()
	assert.ErrorIs(t, err, redis.Nil)
	_, err = redisClient.Get(ctx, "perm:scopes:777:building").Result()
	assert.ErrorIs(t, err, redis.Nil)

	// 其他用户缓存不受影响
	v, err := redisClient.Get(ctx, "perm:scopes:778:community").Result()
	assert.NoError(t, err)
	assert.Equal(t, "other-user-should-survive", v)
}

// TestInvalidateUserCaches_NilRedis — rdb==nil 防御：直接返回不 panic
func TestInvalidateUserCaches_NilRedis(t *testing.T) {
	// RedisClient 为 nil，invalidateUserCaches 应安全返回
	invalidateUserCaches(context.Background(), nil, 42)
}

// TestInvalidateUserCaches_NoScopesKeys — 用户仅有 perm:user（无 perm:scopes 键）→ Scan 返回空 → 跳过 Del
func TestInvalidateUserCaches_NoScopesKeys(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	redisClient.Set(ctx, "perm:user:999", "stale", 0)
	// 无 perm:scopes:999:* 键

	invalidateUserCaches(ctx, redisClient, 999)

	_, err := redisClient.Get(ctx, "perm:user:999").Result()
	assert.ErrorIs(t, err, redis.Nil)
}

// TestInvalidateUserCaches_MultiPageScan — 大量 perm:scopes 键触发多轮 SCAN（cursor 翻页）
func TestInvalidateUserCaches_MultiPageScan(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	// 预置 150 个 scopes 键（超过单页 100）→ 触发多轮 SCAN
	for i := 0; i < 150; i++ {
		redisClient.Set(ctx, fmt.Sprintf("perm:scopes:555:scope%d", i), `{"state":"limited","ids":[100]}`, 0)
	}
	// 其他用户键不应被删除
	redisClient.Set(ctx, "perm:scopes:556:community", "survive", 0)

	invalidateUserCaches(ctx, redisClient, 555)

	// 全部 150 个键应被删
	for i := 0; i < 150; i++ {
		_, err := redisClient.Get(ctx, fmt.Sprintf("perm:scopes:555:scope%d", i)).Result()
		assert.ErrorIs(t, err, redis.Nil, "key scope%d 应被 SCAN-DEL", i)
	}
	// 其他用户不受影响
	v, err := redisClient.Get(ctx, "perm:scopes:556:community").Result()
	assert.NoError(t, err)
	assert.Equal(t, "survive", v)
}

// TestRevokeRole_DifferentScopes 测试不同 scope 类型的撤销
func TestRevokeRole_DifferentScopes(t *testing.T) {
	testCases := []struct {
		name      string
		scopeType string
		scopeId   int64
	}{
		{
			name:      "Revoke building scope",
			scopeType: "building",
			scopeId:   200,
		},
		{
			name:      "Revoke unit scope",
			scopeType: "unit",
			scopeId:   300,
		},
		{
			name:      "Revoke grid scope",
			scopeType: "grid",
			scopeId:   400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mr := miniredis.RunT(t)
			defer mr.Close()

			mockUserRole := new(MockUserRoleModel)
			mockUserRole.On("DeleteByUserIdAndRoleId", mock.Anything, int64(1001), int64(1), tc.scopeType, tc.scopeId).
				Return(nil)

			svcCtx := &svc.ServiceContext{
				UserRoleModel: mockUserRole,
				RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
			}

			logic := NewRevokeRoleLogic(context.Background(), svcCtx)

			// Execute
			resp, err := logic.RevokeRole(&permissionv1.RevokeRoleRequest{
				UserId:    1001,
				RoleId:    1,
				ScopeType: &tc.scopeType,
				ScopeId:   &tc.scopeId,
			})

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, int32(0), resp.Base.Code)

			mockUserRole.AssertExpectations(t)
		})
	}
}

// TestRevokeRole_DeleteFailed 测试删除失败的情况
func TestRevokeRole_DeleteFailed(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("DeleteByUserIdAndRoleId", mock.Anything, int64(1001), int64(1), "community", int64(100)).
		Return(assert.AnError)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewRevokeRoleLogic(context.Background(), svcCtx)

	// Execute
	scopeType := "community"
	scopeId := int64(100)
	resp, err := logic.RevokeRole(&permissionv1.RevokeRoleRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: &scopeType,
		ScopeId:   &scopeId,
	})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	mockUserRole.AssertExpectations(t)
}
