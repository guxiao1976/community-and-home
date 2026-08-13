package permission

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoleModel mocks SysRoleModel interface
type MockRoleModel struct {
	mock.Mock
}

func (m *MockRoleModel) Insert(ctx context.Context, data *model.SysRole) (int64, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleModel) FindOne(ctx context.Context, id int64) (*model.SysRole, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SysRole), args.Error(1)
}

func (m *MockRoleModel) Update(ctx context.Context, data *model.SysRole) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockRoleModel) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleModel) FindAll(ctx context.Context) ([]*model.SysRole, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysRole), args.Error(1)
}

func (m *MockRoleModel) FindByCode(ctx context.Context, roleCode string) (*model.SysRole, error) {
	args := m.Called(ctx, roleCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SysRole), args.Error(1)
}

func (m *MockRoleModel) FindByIds(ctx context.Context, ids []int64) ([]*model.SysRole, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.SysRole), args.Error(1)
}

func (m *MockRoleModel) FindList(ctx context.Context, status *int64, page, pageSize int64) ([]*model.SysRole, int64, error) {
	args := m.Called(ctx, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.SysRole), args.Get(1).(int64), args.Error(2)
}

func (m *MockRoleModel) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestAssignRole_Success 测试成功分配角色
func TestAssignRole_Success(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(1)).
		Return(&model.SysRole{Id: 1, RoleName: "owner"}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("InsertIgnore", mock.Anything, mock.MatchedBy(func(ur *model.RelUserRole) bool {
		return ur.UserId == 1001 && ur.RoleId == 1 && ur.ScopeType == "community" && ur.ScopeId == 100
	})).Return(nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:     mockRole,
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewAssignRoleLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.AssignRole(&permissionv1.AssignRoleRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: "community",
		ScopeId:   100,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)

	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestAssignRole_RoleNotFound 测试角色不存在
func TestAssignRole_RoleNotFound(t *testing.T) {
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

	logic := NewAssignRoleLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.AssignRole(&permissionv1.AssignRoleRequest{
		UserId:    1001,
		RoleId:    999,
		ScopeType: "community",
		ScopeId:   100,
	})

	// Assert - 代码返回自定义错误码（不是 Go error）
	assert.NoError(t, err) // Go error 是 nil
	assert.NotNil(t, resp)
	assert.Equal(t, int32(60001), resp.Base.Code) // 业务错误码
	assert.Contains(t, resp.Base.Msg, "角色不存在")

	mockRole.AssertExpectations(t)
}

// TestAssignRole_Idempotent 测试幂等性（重复分配只一条：INSERT IGNORE 唯一键冲突静默成功）
func TestAssignRole_Idempotent(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(1)).
		Return(&model.SysRole{Id: 1, RoleName: "owner"}, nil)

	mockUserRole := new(MockUserRoleModel)
	// INSERT IGNORE：即使已存在（uk_user_role_scope 冲突）也不报错（重复 Assign 只一条）
	mockUserRole.On("InsertIgnore", mock.Anything, mock.Anything).
		Return(nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:     mockRole,
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewAssignRoleLogic(context.Background(), svcCtx)

	// Execute（连续两次重复分配）
	for i := 0; i < 2; i++ {
		resp, err := logic.AssignRole(&permissionv1.AssignRoleRequest{
			UserId:    1001,
			RoleId:    1,
			ScopeType: "community",
			ScopeId:   100,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.Base.Code)
	}

	// Assert - 幂等：InsertIgnore 恰好被调用两次且均成功（底层 INSERT IGNORE 保证只落一条）
	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
	mockUserRole.AssertNumberOfCalls(t, "InsertIgnore", 2)
}

// TestAssignRole_LifecycleParams 覆盖个体生命周期参数分支：Status/VerifiedAt/ExpiresAt 非 nil 时透传
func TestAssignRole_LifecycleParams(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(1)).
		Return(&model.SysRole{Id: 1, RoleName: "owner"}, nil)

	mockUserRole := new(MockUserRoleModel)
	status := int64(2)
	st32 := int32(status)
	verifiedAt := time.Now().Add(-24 * time.Hour).Unix()
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	mockUserRole.On("InsertIgnore", mock.Anything, mock.MatchedBy(func(ur *model.RelUserRole) bool {
		return ur.UserId == 1001 && ur.RoleId == 1 &&
			ur.Status == status &&
			ur.VerifiedAt.Valid && ur.VerifiedAt.Time.Unix() == verifiedAt &&
			ur.ExpiresAt.Valid && ur.ExpiresAt.Time.Unix() == expiresAt
	})).Return(nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:     mockRole,
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewAssignRoleLogic(context.Background(), svcCtx)
	resp, err := logic.AssignRole(&permissionv1.AssignRoleRequest{
		UserId:     1001,
		RoleId:     1,
		ScopeType:  "community",
		ScopeId:    100,
		Status:     &st32,
		VerifiedAt: &verifiedAt,
		ExpiresAt:  &expiresAt,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestAssignRole_LifecycleParams_ZeroTimestamps 覆盖 VerifiedAt/ExpiresAt =0（未设置）→ NullTime 不 Valid
func TestAssignRole_LifecycleParams_ZeroTimestamps(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(1)).
		Return(&model.SysRole{Id: 1, RoleName: "owner"}, nil)

	mockUserRole := new(MockUserRoleModel)
	zero := int64(0)
	mockUserRole.On("InsertIgnore", mock.Anything, mock.MatchedBy(func(ur *model.RelUserRole) bool {
		return ur.UserId == 1001 && ur.RoleId == 1 &&
			!ur.VerifiedAt.Valid && !ur.ExpiresAt.Valid
	})).Return(nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:     mockRole,
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewAssignRoleLogic(context.Background(), svcCtx)
	resp, err := logic.AssignRole(&permissionv1.AssignRoleRequest{
		UserId:     1001,
		RoleId:     1,
		ScopeType:  "community",
		ScopeId:    100,
		VerifiedAt: &zero,
		ExpiresAt:  &zero,
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestAssignRole_InsertFailed 覆盖 InsertIgnore 返回 error → 直接透传 nil, err
func TestAssignRole_InsertFailed(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(1)).
		Return(&model.SysRole{Id: 1, RoleName: "owner"}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("InsertIgnore", mock.Anything, mock.Anything).
		Return(assert.AnError)

	svcCtx := &svc.ServiceContext{
		RoleModel:     mockRole,
		UserRoleModel: mockUserRole,
		RedisClient:   redis.NewClient(&redis.Options{Addr: mr.Addr()}),
	}

	logic := NewAssignRoleLogic(context.Background(), svcCtx)
	resp, err := logic.AssignRole(&permissionv1.AssignRoleRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: "community",
		ScopeId:   100,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}

// TestAssignRole_CacheInvalidated 测试缓存失效
func TestAssignRole_CacheInvalidated(t *testing.T) {
	// Setup
	mr := miniredis.RunT(t)
	defer mr.Close()

	// 预先设置缓存
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	redisClient.Set(ctx, "perm:user:1001", "cached_value", 0)
	redisClient.Set(ctx, "perm:scopes:1001:community", "cached_scopes", 0)

	mockRole := new(MockRoleModel)
	mockRole.On("FindOne", mock.Anything, int64(1)).
		Return(&model.SysRole{Id: 1, RoleName: "owner"}, nil)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("InsertIgnore", mock.Anything, mock.Anything).
		Return(nil)

	svcCtx := &svc.ServiceContext{
		RoleModel:     mockRole,
		UserRoleModel: mockUserRole,
		RedisClient:   redisClient,
	}

	logic := NewAssignRoleLogic(ctx, svcCtx)

	// Execute
	_, err := logic.AssignRole(&permissionv1.AssignRoleRequest{
		UserId:    1001,
		RoleId:    1,
		ScopeType: "community",
		ScopeId:   100,
	})

	// Assert
	assert.NoError(t, err)

	// 验证缓存已被删除
	val, err := redisClient.Get(ctx, "perm:user:1001").Result()
	assert.Error(t, err) // 应该是 redis.Nil
	assert.Empty(t, val)

	val, err = redisClient.Get(ctx, "perm:scopes:1001:community").Result()
	assert.Error(t, err)
	assert.Empty(t, val)

	mockRole.AssertExpectations(t)
	mockUserRole.AssertExpectations(t)
}
