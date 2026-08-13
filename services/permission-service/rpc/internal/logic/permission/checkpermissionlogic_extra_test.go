package permission

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	zeroredis "github.com/zeromicro/go-zero/core/stores/redis"
)

// TestCheckPermission_PermDefAndUserCacheHit — perm:def + perm:user 双 HIT，不触 DB
func TestCheckPermission_PermDefAndUserCacheHit(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)
	ctx := context.Background()

	// 预置缓存：perm:def 命中 minLevel=2，perm:user 命中 maxLevel=2 → 放行
	redisClient.Set(ctx, "perm:def:GET:/api/vote", "2", 0)
	redisClient.HSet(ctx, "perm:user:777", "GET:/api/vote", "2")

	// 无任何 mock：若触 DB（nil 接口）会 panic/失败
	svcCtx := &svc.ServiceContext{RedisClient: redisClient}
	logic := NewCheckPermissionLogic(ctx, svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  777,
		Action:  "GET",
		ApiPath: "/api/vote",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "双缓存 HIT 应放行")
}

// TestCheckPermission_UserLevelAboveMin — maxLevel > minLevel（能力超出所需）→ 放行
// 守护 `maxLevel < minLevel` 判定（若改成 `!=`，等级更高会被误拒）
func TestCheckPermission_UserLevelAboveMin(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)
	ctx := context.Background()
	redisClient.Set(ctx, "perm:def:GET:/api/read", "0", 0) // minLevel=0
	redisClient.HSet(ctx, "perm:user:789", "GET:/api/read", "2") // maxLevel=2

	svcCtx := &svc.ServiceContext{RedisClient: redisClient}
	logic := NewCheckPermissionLogic(ctx, svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  789,
		Action:  "GET",
		ApiPath: "/api/read",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "maxLevel=2 高于 minLevel=0 应放行")
}

// TestCheckPermission_ReverseOrderGrantLevels — 同 path 多角色，低层级 grant 后处理时不被覆盖
// 守护 `lv > cur` 聚合（若改成 `!=`，低层级会覆盖高层级导致误拒）
func TestCheckPermission_ReverseOrderGrantLevels(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	verifiedAt := time.Now().Add(-24 * time.Hour)
	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "committee:election:vote").Return(&model.SysPermission{
		Id: 600, Code: "committee:election:vote", Path: sql.NullString{}, MinVerfLevel: 2,
	}, nil)
	// 两个角色都授予 permission 600（key=code），但顺序：高层级在前、低层级在后
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(790)).Return([]*model.UserRoleWithInfo{
		{RoleId: 6, ScopeType: "community", ScopeId: 100, URStatus: 2, VerifiedAt: sql.NullTime{Time: verifiedAt, Valid: true}}, // lv=2
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},                          // lv=0
	}, nil)
	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(6)).Return([]*model.RelRolePermission{
		{RoleId: 6, PermissionId: 600},
	}, nil)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 600},
	}, nil)
	mockPerm.On("FindByIds", mock.Anything, []int64{600}).Return([]*model.SysPermission{
		{Id: 600, Code: "committee:election:vote"},
	}, nil)

	svcCtx := &svc.ServiceContext{
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  790,
		ApiPath: "committee:election:vote",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "高层级 grant 应先聚合且不被低层级覆盖")
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}

// TestCheckPermission_GrantsWithError — FindActiveRolesByUserId 返回 (grants, err)：err 优先 → 拒绝
// 守护 `err != nil || len(grants) == 0` 短路（若整条件被替换为 false，会错误地继续聚合）
func TestCheckPermission_GrantsWithError(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(791)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
	}, assert.AnError)

	// 正常路径 err 优先 → 早退，以下聚合调用不会发生；但若 `err != nil || len==0` 被变异为 false，
	// 会继续聚合 → 用 .Maybe() 让这些 expectation 在变异下生效、正常下不强制。
	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil).Maybe()

	mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{
		{Id: 435, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}},
	}, nil).Maybe()

	svcCtx := &svc.ServiceContext{
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  791,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "聚合查询报错应拒绝（err 优先）")
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}

// TestCheckPermission_PermDefCacheMissUserHit — perm:def MISS 回源 DB，perm:user HIT
func TestCheckPermission_PermDefCacheMissUserHit(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)
	ctx := context.Background()
	redisClient.HSet(ctx, "perm:user:778", "GET:/api/users", "0")

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)

	// userMaxLevel HIT → FindActiveRolesByUserId 不应被调用
	mockUserRole := new(MockUserRoleModel)

	svcCtx := &svc.ServiceContext{
		PermissionModel: mockPerm,
		UserRoleModel:   mockUserRole,
		RedisClient:     redisClient,
	}
	logic := NewCheckPermissionLogic(ctx, svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  778,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "perm:def MISS 回源 + perm:user HIT 应放行")

	// perm:def 回填缓存 TTL ≈ 30min（守护 30*time.Minute 算术变异）
	ttl, err := redisClient.TTL(ctx, "perm:def:GET:/api/users").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl, 29*time.Minute, "perm:def 缓存 TTL 应约 30min")
	assert.Less(t, ttl, 31*time.Minute, "perm:def 缓存 TTL 应约 30min")
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestCheckPermission_PermDefSentinel — perm:def 缓存 "-1"（not-found sentinel）→ 拒绝，不触 DB
func TestCheckPermission_PermDefSentinel(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)
	ctx := context.Background()
	redisClient.Set(ctx, "perm:def:GET:/api/unknown", "-1", 0)

	svcCtx := &svc.ServiceContext{RedisClient: redisClient}
	logic := NewCheckPermissionLogic(ctx, svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  779,
		Action:  "GET",
		ApiPath: "/api/unknown",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "perm:def sentinel 应拒绝")
}

// TestCheckPermission_PermDefMissingBoth — 权限定义两路回源（path+code）均 MISS → 拒绝 + 写入 "-1" sentinel
func TestCheckPermission_PermDefMissingBoth(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)
	ctx := context.Background()

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/ghost").Return(nil, sql.ErrNoRows)
	mockPerm.On("FindByCode", mock.Anything, "GET:/api/ghost").Return(nil, sql.ErrNoRows)

	svcCtx := &svc.ServiceContext{PermissionModel: mockPerm, RedisClient: redisClient}
	logic := NewCheckPermissionLogic(ctx, svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  780,
		Action:  "GET",
		ApiPath: "/api/ghost",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed)

	// sentinel 已回填缓存（TTL ≈ 30min，守护 30*time.Minute 算术变异）
	v, err := redisClient.Get(ctx, "perm:def:GET:/api/ghost").Result()
	assert.NoError(t, err)
	assert.Equal(t, "-1", v)
	ttl, err := redisClient.TTL(ctx, "perm:def:GET:/api/ghost").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl, 29*time.Minute, "sentinel 缓存 TTL 应约 30min")
	assert.Less(t, ttl, 31*time.Minute, "sentinel 缓存 TTL 应约 30min")
	mockPerm.AssertExpectations(t)
}

// TestCheckPermission_UserMaxLevelHit — perm:def HIT（minLevel=0），perm:user HIT（level=0）→ 放行
func TestCheckPermission_UserMaxLevelHit(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)
	ctx := context.Background()
	redisClient.Set(ctx, "perm:def:GET:/api/read", "0", 0)
	redisClient.HSet(ctx, "perm:user:781", "GET:/api/read", "0")

	svcCtx := &svc.ServiceContext{RedisClient: redisClient}
	logic := NewCheckPermissionLogic(ctx, svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  781,
		Action:  "GET",
		ApiPath: "/api/read",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed)
}

// TestCheckPermission_FindActiveRolesError — 聚合查询报错 → 拒绝
func TestCheckPermission_FindActiveRolesError(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(782)).Return(nil, assert.AnError)

	svcCtx := &svc.ServiceContext{
		PermissionModel: mockPerm,
		UserRoleModel:   mockUserRole,
		RedisClient:     redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  782,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestCheckPermission_NoRolePerms — 角色无任何权限关联 → 拒绝（permIdSet 空）
func TestCheckPermission_NoRolePerms(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(783)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{}, nil)

	svcCtx := &svc.ServiceContext{
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  783,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}

// TestCheckPermission_FindByIdsError — 权限 ID 集合查询报错 → 拒绝
func TestCheckPermission_FindByIdsError(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)
	mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return(nil, assert.AnError)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(784)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil)

	svcCtx := &svc.ServiceContext{
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  784,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}

// TestCheckPermission_OrphanPermissionId — 角色关联了权限 ID 但权限集合查询返回空（孤儿引用）→ 跳过 → 拒绝
func TestCheckPermission_OrphanPermissionId(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)
	// FindByIds 返回空列表（孤儿权限 ID 435 不在权限表）
	mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(785)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil)

	svcCtx := &svc.ServiceContext{
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  785,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "孤儿权限 ID 应被跳过（fail-closed）")
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}

// TestCheckPermission_DuplicateRoleId — 同 RoleId 多条 grant：rolePermIds 只填充一次（seen 去重）
func TestCheckPermission_DuplicateRoleId(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)
	// 若 seen 去重失效，FindByRoleId 会被调用两次（本测试用 Once 断言恰好一次）
	mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{
		{Id: 435, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}},
	}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(786)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
		{RoleId: 1, ScopeType: "community", ScopeId: 200, URStatus: 1},
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil).Once()

	svcCtx := &svc.ServiceContext{
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  786,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed)
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockRolePerm.AssertNumberOfCalls(t, "FindByRoleId", 1)
}

// TestCheckPermission_SysConfigTTL — SysConfig 非 nil 且配置了 ttl → 缓存 Expire 使用配置值
func TestCheckPermission_SysConfigTTL(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	// 预置 sys_config hash（go-zero redis 客户端读取）
	redisClient.HSet(ctx, "sys_config", "permission.cache.ttl_seconds",
		`{"value":"100","type":"number","desc":"ttl"}`)

	// 构造 sysconfig.Client（go-zero redis → miniredis）
	sc := sysconfig.MustInit(zeroredis.RedisConf{Host: mr.Addr(), Type: "node"}, "", nil)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(787)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil)

	mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{
		{Id: 435, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}},
	}, nil)

	svcCtx := &svc.ServiceContext{
		SysConfig:           sc,
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(ctx, svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  787,
		Action:  "GET",
		ApiPath: "/api/users",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed)

	ttl, err := redisClient.TTL(ctx, "perm:user:787").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl.Seconds(), float64(90), "应使用 sys_config 配置的 TTL（100s）")
	assert.Less(t, ttl.Seconds(), float64(110))
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}

// TestCheckPermission_ActionPrefix — Action + ApiPath 已含 "Action:" 前缀 → needle 保持 ApiPath 不重复拼接
func TestCheckPermission_ActionPrefix(t *testing.T) {
	redisClient, _ := setupMiniRedis(t)

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindByPath", mock.Anything, "GET:/api/users").Return(&model.SysPermission{
		Id: 111, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}, MinVerfLevel: 0,
	}, nil)
	mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{
		{Id: 435, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}},
	}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(788)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, ScopeType: "community", ScopeId: 100, URStatus: 0},
	}, nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil)

	svcCtx := &svc.ServiceContext{
		PermissionModel:     mockPerm,
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		RedisClient:         redisClient,
	}
	logic := NewCheckPermissionLogic(context.Background(), svcCtx)

	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  788,
		Action:  "GET",
		ApiPath: "GET:/api/users",
	})
	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "ApiPath 已含 METHOD 前缀时不应重复拼接")
	mockPerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
}
