package permission

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// SEE: [[redis-cache-soft-delete]] — GetDataScopes 读穿缓存：HIT 直接返回，MISS 计算后写 JSON + EXPIRE

func TestGetDataScopes_Limited(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1001)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 0},
		{RoleId: 1, ScopeType: model.ScopeTypeCommunity, ScopeId: 200, URStatus: 1},
	}, nil)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole, RedisClient: redisClient}
	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	resp, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{UserId: 1001, ScopeType: model.ScopeTypeCommunity})
	assert.NoError(t, err)
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, resp.State)
	assert.ElementsMatch(t, []int64{100, 200}, resp.ScopeIds)
	mockUserRole.AssertExpectations(t)

	// 缓存已写入 JSON
	raw, err := redisClient.Get(context.Background(), "perm:scopes:1001:community").Result()
	assert.NoError(t, err)
	var cached struct {
		State string  `json:"state"`
		Ids   []int64 `json:"ids"`
	}
	assert.NoError(t, json.Unmarshal([]byte(raw), &cached))
	assert.Equal(t, "limited", cached.State)
	assert.ElementsMatch(t, []int64{100, 200}, cached.Ids)
}

func TestGetDataScopes_Empty(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1002)).Return([]*model.UserRoleWithInfo{
		{RoleId: 9, ScopeType: model.ScopeTypeEmpty, ScopeId: 0, URStatus: 2},
	}, nil)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole, RedisClient: redisClient}
	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	resp, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{UserId: 1002, ScopeType: model.ScopeTypeCommunity})
	assert.NoError(t, err)
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, resp.State)
	assert.Empty(t, resp.ScopeIds)
	mockUserRole.AssertExpectations(t)
}

func TestGetDataScopes_Global(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1003)).Return([]*model.UserRoleWithInfo{
		{RoleId: 8, ScopeType: model.ScopeTypeGlobal, ScopeId: 0, URStatus: 2},
	}, nil)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole, RedisClient: redisClient}
	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	resp, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{UserId: 1003, ScopeType: model.ScopeTypeCommunity})
	assert.NoError(t, err)
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL, resp.State)
	assert.Empty(t, resp.ScopeIds)
	mockUserRole.AssertExpectations(t)
}

func TestGetDataScopes_CacheHitNoDB(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// 预先写入缓存（JSON）
	cached := `{"state":"limited","ids":[100,200]}`
	assert.NoError(t, redisClient.Set(context.Background(), "perm:scopes:1004:community", cached, 0).Err())

	mockUserRole := new(MockUserRoleModel) // 不设任何 expectation：若命中缓存不应查 DB

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole, RedisClient: redisClient}
	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	resp, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{UserId: 1004, ScopeType: model.ScopeTypeCommunity})
	assert.NoError(t, err)
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, resp.State)
	assert.ElementsMatch(t, []int64{100, 200}, resp.ScopeIds)
	mockUserRole.AssertExpectations(t) // 无 expectation → 任何对 mock 的调用都会失败
}

func TestGetDataScopes_CacheMissWritesCache(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1005)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: model.ScopeTypeCommunity, ScopeId: 500, URStatus: 0},
	}, nil)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole, RedisClient: redisClient}
	logic := NewGetDataScopesLogic(context.Background(), svcCtx)

	_, err := logic.GetDataScopes(&permissionv1.GetDataScopesRequest{UserId: 1005, ScopeType: model.ScopeTypeCommunity})
	assert.NoError(t, err)

	// 缓存存在且带 TTL
	ttl, err := redisClient.TTL(context.Background(), "perm:scopes:1005:community").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl.Seconds(), float64(0))
	mockUserRole.AssertExpectations(t)
}
