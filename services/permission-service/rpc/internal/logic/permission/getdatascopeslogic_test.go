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

// TestGetDataScopes_Success 测试成功获取数据范围
func TestGetDataScopes_Success(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindScopesByUserId", mock.Anything, int64(1001), "community").
		Return([]int64{100, 200, 300}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{
		UserId:    1001,
		ScopeType: "community",
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.ScopeIds, 3)
	assert.Contains(t, resp.ScopeIds, int64(100))
	assert.Contains(t, resp.ScopeIds, int64(200))
	assert.Contains(t, resp.ScopeIds, int64(300))

	mockUserRole.AssertExpectations(t)
}

// TestGetDataScopes_EmptyResult 测试用户无数据范围
func TestGetDataScopes_EmptyResult(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindScopesByUserId", mock.Anything, int64(1001), "building").
		Return([]int64{}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{
		UserId:    1001,
		ScopeType: "building",
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Empty(t, resp.ScopeIds)

	mockUserRole.AssertExpectations(t)
}

// TestGetDataScopes_MultipleScopeTypes 测试不同的 scope_type
func TestGetDataScopes_MultipleScopeTypes(t *testing.T) {
	testCases := []struct {
		name      string
		userId    int64
		scopeType string
		scopeIds  []int64
	}{
		{
			name:      "Community scope",
			userId:    1001,
			scopeType: "community",
			scopeIds:  []int64{100},
		},
		{
			name:      "Building scope",
			userId:    1002,
			scopeType: "building",
			scopeIds:  []int64{201, 202},
		},
		{
			name:      "Unit scope",
			userId:    1003,
			scopeType: "unit",
			scopeIds:  []int64{301, 302, 303},
		},
		{
			name:      "Grid scope",
			userId:    1004,
			scopeType: "grid",
			scopeIds:  []int64{401},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mr := miniredis.RunT(t)
			defer mr.Close()

			mockUserRole := new(MockUserRoleModel)
			mockUserRole.On("FindScopesByUserId", mock.Anything, tc.userId, tc.scopeType).
				Return(tc.scopeIds, nil)

			svcCtx := &svc.ServiceContext{
				UserRoleModel: mockUserRole,
				RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
			}

			logic := NewGetDataScopesLogic(context.Background(), svcCtx)

			// Execute
			resp, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{
				UserId:    tc.userId,
				ScopeType: tc.scopeType,
			})

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, int32(0), resp.Base.Code)
			assert.Equal(t, tc.scopeIds, resp.ScopeIds)

			mockUserRole.AssertExpectations(t)
		})
	}
}

// TestGetDataScopes_CacheWritten 测试缓存是否正确写入 Redis
func TestGetDataScopes_CacheWritten(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindScopesByUserId", mock.Anything, int64(1001), "community").
		Return([]int64{100, 200}, nil)

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redisClient,
	}

	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	// Execute
	_, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{
		UserId:    1001,
		ScopeType: "community",
	})

	// Assert
	assert.NoError(t, err)

	// 验证 Redis 缓存
	cacheKey := "perm:scopes:1001:community"
	members, err := redisClient.SMembers(context.Background(), cacheKey).Result()
	assert.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Contains(t, members, "100")
	assert.Contains(t, members, "200")

	// 验证 TTL 设置
	ttl, err := redisClient.TTL(context.Background(), cacheKey).Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl.Seconds(), float64(0))

	mockUserRole.AssertExpectations(t)
}
