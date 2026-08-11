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
)

// SEE: [[testing-discipline]] — CheckPermission 核心场景测试：系统角色走 rel_role_permission（无短路）、普通用户权限匹配、拒绝、缓存

// MockUserRoleModel mocks RelUserRoleModel interface
type MockUserRoleModel struct {
	mock.Mock
}

func (m *MockUserRoleModel) Insert(ctx context.Context, data *model.RelUserRole) (int64, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*model.RelUserRole, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelUserRole), args.Error(1)
}

func (m *MockUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserRoleWithInfo), args.Error(1)
}

func (m *MockUserRoleModel) FindScopesByUserId(ctx context.Context, userId int64, scopeType string) ([]int64, error) {
	args := m.Called(ctx, userId, scopeType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockUserRoleModel) DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64, scopeType string, scopeId int64) error {
	args := m.Called(ctx, userId, roleId, scopeType, scopeId)
	return args.Error(0)
}

func (m *MockUserRoleModel) BatchInsertUserRoles(ctx context.Context, records []*model.RelUserRole) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

func (m *MockUserRoleModel) CountByRoleId(ctx context.Context, roleId int64) (int64, error) {
	args := m.Called(ctx, roleId)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRoleModel) FindAllByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserRoleWithInfo), args.Error(1)
}

func (m *MockUserRoleModel) UpdateRoleStatus(ctx context.Context, userId, roleId int64, scopeType string, scopeId, status int64, verifiedAt, expiresAt sql.NullTime) error {
	args := m.Called(ctx, userId, roleId, scopeType, scopeId, status, verifiedAt, expiresAt)
	return args.Error(0)
}

func (m *MockUserRoleModel) FindByRoleId(ctx context.Context, roleId int64) ([]*model.RelUserRole, error) {
	args := m.Called(ctx, roleId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelUserRole), args.Error(1)
}

// MockRolePermissionModel mocks RelRolePermissionModel interface
type MockRolePermissionModel struct {
	mock.Mock
}

func (m *MockRolePermissionModel) Insert(ctx context.Context, data *model.RelRolePermission) (int64, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRolePermissionModel) FindByRoleId(ctx context.Context, roleId int64) ([]*model.RelRolePermission, error) {
	args := m.Called(ctx, roleId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelRolePermission), args.Error(1)
}

func (m *MockRolePermissionModel) DeleteByRoleId(ctx context.Context, roleId int64) error {
	args := m.Called(ctx, roleId)
	return args.Error(0)
}

func (m *MockRolePermissionModel) BatchInsert(ctx context.Context, records []*model.RelRolePermission) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

// MockPermissionModel mocks SysPermissionModel interface
type MockPermissionModel struct {
	mock.Mock
}

func (m *MockPermissionModel) FindAll(ctx context.Context) ([]*model.SysPermission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysPermission), args.Error(1)
}

func (m *MockPermissionModel) FindByIds(ctx context.Context, ids []int64) ([]*model.SysPermission, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysPermission), args.Error(1)
}

func (m *MockPermissionModel) FindByCode(ctx context.Context, code string) (*model.SysPermission, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SysPermission), args.Error(1)
}

func (m *MockPermissionModel) FindWithFilter(ctx context.Context, typeFilter, statusFilter *int64) ([]*model.SysPermission, error) {
	args := m.Called(ctx, typeFilter, statusFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysPermission), args.Error(1)
}

func setupMiniRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client, mr
}

func TestCheckPermission_SystemRole_RespectsPermissions(t *testing.T) {
	// SEE: [[testing-discipline]] — 系统角色（is_system=1）走 rel_role_permission 匹配，无特权短路
	mockUserRole := new(MockUserRoleModel)
	mockRolePerm := new(MockRolePermissionModel)
	mockPerm := new(MockPermissionModel)
	redisClient, _ := setupMiniRedis(t)

	// Mock 用户拥有系统角色（有对应权限 — 应放行）
	mockUserRole.On("FindActiveByUserId", mock.Anything, int64(100)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, RoleCode: "owner", IsSystem: 1},
	}, nil)

	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 111},
	}, nil)

	mockPerm.On("FindByIds", mock.Anything, []int64{111}).Return([]*model.SysPermission{
		{Id: 111, Code: "user:read:list-api", Path: sql.NullString{String: "GET:/api/users", Valid: true}},
	}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
		RedisClient:         redisClient,
	}

	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  100,
		Action:  "GET",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "系统角色拥有匹配权限应放行（不走短路，走rel_role_permission）")
	mockUserRole.AssertExpectations(t)

	// 第二个测试：系统角色无匹配权限 — 应拒绝
	mockUserRole2 := new(MockUserRoleModel)
	mockRolePerm2 := new(MockRolePermissionModel)
	mockPerm2 := new(MockPermissionModel)

	mockUserRole2.On("FindActiveByUserId", mock.Anything, int64(200)).Return([]*model.UserRoleWithInfo{
		{RoleId: 4, RoleCode: "grid_worker", IsSystem: 1},
	}, nil)

	mockRolePerm2.On("FindByRoleId", mock.Anything, int64(4)).Return([]*model.RelRolePermission{
		{RoleId: 4, PermissionId: 411},
	}, nil)

	mockPerm2.On("FindByIds", mock.Anything, []int64{411}).Return([]*model.SysPermission{
		{Id: 411, Code: "community:read:list-api", Path: sql.NullString{String: "GET:/api/community/communities", Valid: true}},
	}, nil)

	svcCtx2 := &svc.ServiceContext{
		UserRoleModel:       mockUserRole2,
		RolePermissionModel: mockRolePerm2,
		PermissionModel:     mockPerm2,
		RedisClient:         redisClient,
	}

	logic2 := NewCheckPermissionLogic(context.Background(), svcCtx2)
	resp2, err2 := logic2.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  200,
		Action:  "DELETE",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err2)
	assert.False(t, resp2.Allowed, "系统角色无匹配权限应拒绝（证明不再短路）")
}

func TestCheckPermission_NormalUser_PermissionMatched(t *testing.T) {
	// SEE: [[testing-discipline]] — 普通用户权限匹配通过
	mockUserRole := new(MockUserRoleModel)
	mockRolePerm := new(MockRolePermissionModel)
	mockPerm := new(MockPermissionModel)
	redisClient, _ := setupMiniRedis(t)

	// Mock 用户拥有普通角色
	mockUserRole.On("FindActiveByUserId", mock.Anything, int64(200)).Return([]*model.UserRoleWithInfo{
		{RoleId: 2, RoleCode: "property_admin", IsSystem: 0},
	}, nil)

	// Mock 角色权限关联
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(2)).Return([]*model.RelRolePermission{
		{RoleId: 2, PermissionId: 101},
	}, nil)

	// Mock 权限定义（匹配请求） - path 含 METHOD 前缀
	mockPerm.On("FindByIds", mock.Anything, []int64{101}).Return([]*model.SysPermission{
		{Id: 101, Code: "user:read", Path: sql.NullString{String: "GET:/api/users", Valid: true}},
	}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
		RedisClient:         redisClient,
	}

	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  200,
		Action:  "GET",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "权限匹配应该通过")
	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}

func TestCheckPermission_NormalUser_PermissionDenied(t *testing.T) {
	// SEE: [[testing-discipline]] — 普通用户无权限，拒绝访问
	mockUserRole := new(MockUserRoleModel)
	mockRolePerm := new(MockRolePermissionModel)
	mockPerm := new(MockPermissionModel)
	redisClient, _ := setupMiniRedis(t)

	// Mock 用户拥有普通角色
	mockUserRole.On("FindActiveByUserId", mock.Anything, int64(300)).Return([]*model.UserRoleWithInfo{
		{RoleId: 3, RoleCode: "viewer", IsSystem: 0},
	}, nil)

	// Mock 角色权限关联
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(3)).Return([]*model.RelRolePermission{
		{RoleId: 3, PermissionId: 101},
	}, nil)

	// Mock 权限定义（路径完全不同，应该拒绝）
	mockPerm.On("FindByIds", mock.Anything, []int64{101}).Return([]*model.SysPermission{
		{Id: 101, Code: "post:read", Path: sql.NullString{String: "/api/posts", Valid: true}},
	}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
		RedisClient:         redisClient,
	}

	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  300,
		Action:  "DELETE",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "无权限应该拒绝")
}

func TestCheckPermission_CacheHit(t *testing.T) {
	// SEE: [[testing-discipline]] — Redis 缓存命中，不查询数据库
	redisClient, mr := setupMiniRedis(t)

	// 预先设置缓存
	mr.SetAdd("perm:user:400", "GET:/api/users")

	svcCtx := &svc.ServiceContext{
		RedisClient: redisClient,
	}

	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  400,
		Action:  "GET",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err)
	assert.True(t, resp.Allowed, "缓存命中应该直接返回")
}

func TestCheckPermission_NoRoles_Denied(t *testing.T) {
	// SEE: [[testing-discipline]] — 用户无任何角色，拒绝访问
	mockUserRole := new(MockUserRoleModel)
	redisClient, _ := setupMiniRedis(t)

	mockUserRole.On("FindActiveByUserId", mock.Anything, int64(500)).Return([]*model.UserRoleWithInfo{}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
		RedisClient:   redisClient,
	}

	logic := NewCheckPermissionLogic(context.Background(), svcCtx)
	resp, err := logic.CheckPermission(&permissionv1.CheckPermissionRequest{
		UserId:  500,
		Action:  "GET",
		ApiPath: "/api/users",
	})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed, "无角色应该拒绝")
}
