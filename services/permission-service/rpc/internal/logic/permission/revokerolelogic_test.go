package permission

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
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
