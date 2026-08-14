package permission

import (
	"context"
	"database/sql"
	"testing"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestUpdateRole_Success 测试成功更新角色
func TestUpdateRole_Success(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{
		Id:          1,
		RoleCode:    "owner",
		RoleName:    "业主",
		Description: sql.NullString{String: "小区业主", Valid: true},
		IsSystem:    0,
		Status:      1,
	}
	mockRole.On("FindOne", mock.Anything, int64(1)).
		Return(existingRole, nil)
	mockRole.On("Update", mock.Anything, mock.MatchedBy(func(role *model.SysRole) bool {
		return role.Id == 1 && role.RoleName == "高级业主"
	})).Return(nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("DeleteByRoleId", mock.Anything, int64(1)).Return(nil)
	mockRolePerm.On("BatchInsert", mock.Anything, mock.Anything).Return(nil)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).
		Return([]*model.RelRolePermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindByRoleId", mock.Anything, int64(1)).
		Return([]*model.RelUserRole{
			{UserId: 1001, RoleId: 1},
			{UserId: 1002, RoleId: 1},
		}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
		UserRoleModel:       mockUserRole,
		RedisClient:         redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	// Execute
	newName := "高级业主"
	newDesc := "高级小区业主"
	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{
		Id:            1,
		Name:          &newName,
		Description:   &newDesc,
		PermissionIds: []int64{10, 20, 30},
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)

	mockRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestUpdateRole_RoleNotFound 测试角色不存在
func TestUpdateRole_RoleNotFound(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(999)).
		Return(nil, sql.ErrNoRows)

	svcCtx := &svc.ServiceContext{
		RoleModel:   mockRole,
		RedisClient: redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	// Execute
	newName := "不存在的角色"
	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{
		Id:   999,
		Name: &newName,
	})

	// Assert - 代码返回业务错误码，不是 Go error
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(60001), resp.Base.Code)
	assert.Contains(t, resp.Base.Msg, "角色不存在")

	mockRole.AssertExpectations(t)
}

// TestUpdateRole_SystemRoleFieldLevel — 系统角色字段级策略（REQ-UPDATE-4）：name/platforms 放行、status 保持原值
// RED: 当前实现整单拒绝系统角色（60004）→ Base.Code 应为 0 却得 60004 → FAIL
// SEE: [[is-system-no-permission-shortcut]] — 字段级编辑放行不改变权限模型
func TestUpdateRole_SystemRoleFieldLevel(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{
		Id: 1, RoleCode: "owner", RoleName: "业主", IsSystem: 1, Status: 1, Platforms: "pc",
	}
	mockRole.On("FindOne", mock.Anything, int64(1)).Return(existingRole, nil)
	var captured *model.SysRole
	mockRole.On("Update", mock.Anything, mock.MatchedBy(func(r *model.SysRole) bool {
		captured = r
		return r.Id == 1
	})).Return(nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelUserRole{}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
		UserRoleModel:       mockUserRole,
		RedisClient:         redisClient,
	}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	newName := "改名"
	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{
		Id:        1,
		Name:      &newName,
		Platforms: []string{"mobile"},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "系统角色 name/platforms 编辑应放行")
	require.NotNil(t, captured, "Update 应被调用")
	assert.Equal(t, "改名", captured.RoleName)
	assert.Equal(t, "mobile", captured.Platforms)
	assert.Equal(t, int64(1), captured.Status, "系统角色 status 应保持原值")
	mockRole.AssertExpectations(t)
}

// TestUpdateRole_SystemRoleStatusAtomic — 系统角色 + status + name → 60004 原子拒绝，Update 不被调用
// RED: 当前实现整单拒绝，断言 msg 含「状态不可修改」失败（现 msg 为「系统角色不可修改」）
func TestUpdateRole_SystemRoleStatusAtomic(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{Id: 1, RoleCode: "owner", RoleName: "业主", IsSystem: 1, Status: 1, Platforms: "pc"}
	mockRole.On("FindOne", mock.Anything, int64(1)).Return(existingRole, nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole, RedisClient: redis.NewClient(&redis.Options{Addr: mr.Addr()})}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	newName := "改名"
	status := int32(0)
	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{Id: 1, Name: &newName, Status: &status})
	assert.NoError(t, err)
	assert.Equal(t, int32(60004), resp.Base.Code, "系统角色 status 修改应 60004")
	assert.Contains(t, resp.Base.Msg, "状态不可修改", "60004 语义应收窄为「系统角色状态不可修改」")
	mockRole.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestUpdateRole_SystemRoleStatusGatePriority — 系统角色 + status + 非法 platforms → 60004 优先（status 门禁先于 platforms 校验）
func TestUpdateRole_SystemRoleStatusGatePriority(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{Id: 1, RoleCode: "owner", IsSystem: 1, Status: 1, Platforms: "pc"}
	mockRole.On("FindOne", mock.Anything, int64(1)).Return(existingRole, nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole, RedisClient: redis.NewClient(&redis.Options{Addr: mr.Addr()})}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	status := int32(0)
	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{Id: 1, Status: &status, Platforms: []string{"web"}})
	assert.NoError(t, err)
	assert.Equal(t, int32(60004), resp.Base.Code, "系统角色 status 门禁优先于 60008 platforms 校验")
	mockRole.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestUpdateRole_SystemRoleInvalidPlatform — 系统角色 + platforms=["web"]（无 status）→ 60008
// RED: 当前实现整单拒绝返回 60004 → 期望 60008 → FAIL
func TestUpdateRole_SystemRoleInvalidPlatform(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{Id: 1, RoleCode: "owner", IsSystem: 1, Status: 1, Platforms: "pc"}
	mockRole.On("FindOne", mock.Anything, int64(1)).Return(existingRole, nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole, RedisClient: redis.NewClient(&redis.Options{Addr: mr.Addr()})}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{Id: 1, Platforms: []string{"web"}})
	assert.NoError(t, err)
	assert.Equal(t, int32(60008), resp.Base.Code, "系统角色非法端（无 status）应 60008")
	mockRole.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestUpdateRole_NonSystemStatus — 非系统角色 + status=0 → 成功（状态落库）
func TestUpdateRole_NonSystemStatus(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{Id: 2, RoleCode: "custom", RoleName: "自定义", IsSystem: 0, Status: 1}
	mockRole.On("FindOne", mock.Anything, int64(2)).Return(existingRole, nil)
	var captured *model.SysRole
	mockRole.On("Update", mock.Anything, mock.MatchedBy(func(r *model.SysRole) bool {
		captured = r
		return r.Id == 2
	})).Return(nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(2)).Return([]*model.RelRolePermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindByRoleId", mock.Anything, int64(2)).Return([]*model.RelUserRole{}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel: mockRole, RolePermissionModel: mockRolePerm, UserRoleModel: mockUserRole,
		RedisClient: redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	status := int32(0)
	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{Id: 2, Status: &status})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "非系统角色 status 修改应成功")
	require.NotNil(t, captured)
	assert.Equal(t, int64(0), captured.Status, "非系统角色状态应落库")
	mockRole.AssertExpectations(t)
}

// TestUpdateRole_SortOrderPersisted — 传 sort_order=5 → Update 捕获 SortOrder==5（D6 落库修复）
func TestUpdateRole_SortOrderPersisted(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{Id: 3, RoleCode: "custom", RoleName: "自定义", IsSystem: 0, Status: 1, SortOrder: 1}
	mockRole.On("FindOne", mock.Anything, int64(3)).Return(existingRole, nil)
	var captured *model.SysRole
	mockRole.On("Update", mock.Anything, mock.MatchedBy(func(r *model.SysRole) bool {
		captured = r
		return r.Id == 3 && r.SortOrder == 5
	})).Return(nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(3)).Return([]*model.RelRolePermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindByRoleId", mock.Anything, int64(3)).Return([]*model.RelUserRole{}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel: mockRole, RolePermissionModel: mockRolePerm, UserRoleModel: mockUserRole,
		RedisClient: redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	sortOrder := int32(5)
	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{Id: 3, SortOrder: &sortOrder})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	require.NotNil(t, captured)
	assert.Equal(t, int64(5), captured.SortOrder)
	mockRole.AssertExpectations(t)
}

// TestUpdateRole_PlatformsEmptyClears — platforms=[] → Update 捕获 Platforms==""（无条件覆盖 = fail-open 清空，REQ-PLAT-5）
// RED: 当前实现不覆盖 Platforms → 捕获到既有 "pc" ≠ "" → FAIL
func TestUpdateRole_PlatformsEmptyClears(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{Id: 4, RoleCode: "custom", RoleName: "自定义", IsSystem: 0, Status: 1, Platforms: "pc"}
	mockRole.On("FindOne", mock.Anything, int64(4)).Return(existingRole, nil)
	var captured *model.SysRole
	mockRole.On("Update", mock.Anything, mock.MatchedBy(func(r *model.SysRole) bool {
		captured = r
		return r.Id == 4
	})).Return(nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(4)).Return([]*model.RelRolePermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindByRoleId", mock.Anything, int64(4)).Return([]*model.RelUserRole{}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel: mockRole, RolePermissionModel: mockRolePerm, UserRoleModel: mockUserRole,
		RedisClient: redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}
	logic := NewUpdateRoleLogic(context.Background(), svcCtx)

	resp, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{Id: 4, Platforms: []string{}})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	require.NotNil(t, captured)
	assert.Equal(t, "", captured.Platforms, "空列表应显式清空 platforms（fail-open）")
	mockRole.AssertExpectations(t)
}

// TestUpdateRole_PlatformsChangeInvalidatesCaches — platforms 变更 → 持有者 perm:user / perm:scopes 被 DEL（REQ-PLAT-7）
// 角色变更必须失效持有者缓存（既有 invalidateRoleCache 模式复用）
func TestUpdateRole_PlatformsChangeInvalidatesCaches(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	redisClient.Set(ctx, "perm:user:1001", "cached_1001", 0)
	redisClient.Set(ctx, "perm:user:1002", "cached_1002", 0)
	redisClient.Set(ctx, "perm:scopes:1001:community", "cached_scopes_1001", 0)

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{Id: 1, RoleCode: "owner", RoleName: "业主", IsSystem: 0, Status: 1, Platforms: "pc"}
	mockRole.On("FindOne", mock.Anything, int64(1)).Return(existingRole, nil)
	mockRole.On("Update", mock.Anything, mock.Anything).Return(nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelUserRole{
		{UserId: 1001, RoleId: 1},
		{UserId: 1002, RoleId: 1},
	}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel: mockRole, RolePermissionModel: mockRolePerm, UserRoleModel: mockUserRole,
		RedisClient: redisClient,
	}
	logic := NewUpdateRoleLogic(ctx, svcCtx)

	_, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{Id: 1, Platforms: []string{"mobile"}})
	assert.NoError(t, err)

	for _, key := range []string{"perm:user:1001", "perm:user:1002", "perm:scopes:1001:community"} {
		val, err := redisClient.Get(ctx, key).Result()
		assert.Error(t, err, "key %s 应已被 DEL", key)
		assert.Empty(t, val)
	}
	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestUpdateRole_CacheInvalidatedForMultipleUsers 测试批量失效缓存
func TestUpdateRole_CacheInvalidatedForMultipleUsers(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	// 预先设置多个用户的缓存
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	redisClient.Set(ctx, "perm:user:1001", "cached_1001", 0)
	redisClient.Set(ctx, "perm:user:1002", "cached_1002", 0)
	redisClient.Set(ctx, "perm:scopes:1001:community", "cached_scopes_1001", 0)

	mockRole := new(MockRoleModel)
	existingRole := &model.SysRole{
		Id:       1,
		RoleCode: "owner",
		RoleName: "业主",
		IsSystem: 0,
		Status:   1,
	}
	mockRole.On("FindOne", mock.Anything, int64(1)).Return(existingRole, nil)
	mockRole.On("Update", mock.Anything, mock.Anything).Return(nil)

	mockRolePerm := new(MockRolePermissionModel)
	mockRolePerm.On("DeleteByRoleId", mock.Anything, int64(1)).Return(nil)
	mockRolePerm.On("BatchInsert", mock.Anything, mock.Anything).Return(nil)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).
		Return([]*model.RelRolePermission{}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindByRoleId", mock.Anything, int64(1)).
		Return([]*model.RelUserRole{
			{UserId: 1001, RoleId: 1},
			{UserId: 1002, RoleId: 1},
		}, nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:           mockRole,
		RolePermissionModel: mockRolePerm,
		UserRoleModel:       mockUserRole,
		RedisClient:         redisClient,
	}

	logic := NewUpdateRoleLogic(ctx, svcCtx)

	// Execute
	newName := "新业主"
	_, err := logic.UpdateRole(&permissionv1.UpdateRoleRequest{
		Id:            1,
		Name:          &newName,
		PermissionIds: []int64{10},
	})

	// Assert
	assert.NoError(t, err)

	// 验证所有用户的缓存都被删除
	val, err := redisClient.Get(ctx, "perm:user:1001").Result()
	assert.Error(t, err)
	assert.Empty(t, val)

	val, err = redisClient.Get(ctx, "perm:user:1002").Result()
	assert.Error(t, err)
	assert.Empty(t, val)

	val, err = redisClient.Get(ctx, "perm:scopes:1001:community").Result()
	assert.Error(t, err)
	assert.Empty(t, val)

	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}
