package permission

import (
	"context"
	"database/sql"
	"testing"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestUpdateUserRoleStatus_Success_InvalidatesCaches — 状态更新成功后，perm:user 与 perm:scopes 缓存必须被失效
// T1.6: UpdateUserRoleStatus 收敛调用 invalidateUserCaches（DEL perm:user + SCAN-DEL perm:scopes:{userId}:*）
// 断言：预置缓存 + VerifiedAt unix 时间戳解析 → 调用后缓存均被删除（Get 返回 redis.Nil）
// SEE: [[redis-cache-soft-delete]] — 失效与 grant 变更联动
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — QA 补测，RED 摘录见 CHANGELOG 2026-08-12 补测节
func TestUpdateUserRoleStatus_Success_InvalidatesCaches(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	// 预置缓存（调用前存在，调用后必须被删）
	redisClient.Set(ctx, "perm:user:1001", "cached_permissions", 0)
	redisClient.Set(ctx, "perm:scopes:1001:community", `{"state":"limited","ids":[100]}`, 0)

	verifiedAtUnix := int64(1723420800)

	mockUserRole := new(MockUserRoleModel)
	// 验证 unix 时间戳 → sql.NullTime 解析（认证通过时间）
	mockUserRole.On("UpdateRoleStatus", mock.Anything, int64(1001), int64(1), "community", int64(100), int64(2),
		mock.MatchedBy(func(vt sql.NullTime) bool {
			return vt.Valid && vt.Time.Unix() == verifiedAtUnix
		}),
		sql.NullTime{},
	).Return(nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redisClient,
	}

	logic := NewUpdateUserRoleStatusLogic(ctx, svcCtx)

	resp, err := logic.UpdateUserRoleStatus(&permissionv1.UpdateUserRoleStatusRequest{
		UserId:     1001,
		RoleId:     1,
		ScopeType:  "community",
		ScopeId:    100,
		Status:     2,
		VerifiedAt: &verifiedAtUnix,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)

	// GREEN：断言缓存已被 invalidateUserCaches 删除（Get 返回 redis.Nil）
	_, err = redisClient.Get(ctx, "perm:user:1001").Result()
	assert.ErrorIs(t, err, redis.Nil, "perm:user:1001 应被 DEL")
	_, err = redisClient.Get(ctx, "perm:scopes:1001:community").Result()
	assert.ErrorIs(t, err, redis.Nil, "perm:scopes:1001:community 应被 SCAN-DEL")

	mockUserRole.AssertExpectations(t)
}

// TestUpdateUserRoleStatus_ModelError_Propagated — UpdateRoleStatus 返回 error 时传播，不调用缓存失效
func TestUpdateUserRoleStatus_ModelError_Propagated(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("UpdateRoleStatus", mock.Anything, int64(1001), int64(1), "community", int64(100), int64(3),
		sql.NullTime{}, sql.NullTime{}).
		Return(assert.AnError)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redisClient,
	}

	logic := NewUpdateUserRoleStatusLogic(context.Background(), svcCtx)

	resp, err := logic.UpdateUserRoleStatus(&permissionv1.UpdateUserRoleStatusRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: "community",
		ScopeId:   100,
		Status:    3, // 已驳回
	})

	assert.Error(t, err)
	assert.Nil(t, resp)

	mockUserRole.AssertExpectations(t)
}
