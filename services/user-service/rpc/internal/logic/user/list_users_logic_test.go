package user

import (
	"context"
	"database/sql"
	"testing"
	"time"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserBaseModel mocks UserBaseModel interface
type MockUserBaseModel struct {
	mock.Mock
}

func (m *MockUserBaseModel) Insert(ctx context.Context, data *model.UserBase) (sql.Result, error) {
	args := m.Called(ctx, data)
	return nil, args.Error(1)
}

func (m *MockUserBaseModel) FindOne(ctx context.Context, id int64) (*model.UserBase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserBase), args.Error(1)
}

func (m *MockUserBaseModel) FindOneByPhone(ctx context.Context, encryptedPhone string) (*model.UserBase, error) {
	args := m.Called(ctx, encryptedPhone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserBase), args.Error(1)
}

func (m *MockUserBaseModel) FindByIds(ctx context.Context, ids []int64) ([]*model.UserBase, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.UserBase), args.Error(1)
}

func (m *MockUserBaseModel) FindPage(ctx context.Context, keyword string, encryptedPhone string, status *int64, page, pageSize int32) ([]*model.UserBase, int64, error) {
	args := m.Called(ctx, keyword, encryptedPhone, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.UserBase), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserBaseModel) Update(ctx context.Context, data *model.UserBase) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockUserBaseModel) SoftDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserBaseModel) UpdateStatus(ctx context.Context, id int64, status int64) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockUserBaseModel) UpdateModerationStatus(ctx context.Context, userId int64, status int64) error {
	args := m.Called(ctx, userId, status)
	return args.Error(0)
}

func (m *MockUserBaseModel) UpdateNickname(ctx context.Context, userId int64, nickname string) error {
	args := m.Called(ctx, userId, nickname)
	return args.Error(0)
}

func (m *MockUserBaseModel) UpdateNicknameModerationStatus(ctx context.Context, userId int64, status int64) error {
	args := m.Called(ctx, userId, status)
	return args.Error(0)
}

func (m *MockUserBaseModel) UpdateRealNameAndIdCard(ctx context.Context, userId int64, realName, idCardNumber string) error {
	args := m.Called(ctx, userId, realName, idCardNumber)
	return args.Error(0)
}

// TestListUsers_Success 测试成功获取用户列表
func TestListUsers_Success(t *testing.T) {
	// Setup
	mockUserBase := new(MockUserBaseModel)
	now := time.Now()

	mockUserBase.On("FindPage", mock.Anything, "", "", (*int64)(nil), int32(1), int32(20)).
		Return([]*model.UserBase{
			{
				Id:          1001,
				Phone:       "encrypted_13800138001",
				Nickname:    sql.NullString{String: "张三", Valid: true},
				AvatarUrl:   sql.NullString{String: "http://avatar.com/1.jpg", Valid: true},
				Status:      1,
				CreditScore: 100,
				CreatedTime: now,
			},
			{
				Id:          1002,
				Phone:       "encrypted_13800138002",
				Nickname:    sql.NullString{String: "李四", Valid: true},
				AvatarUrl:   sql.NullString{},
				Status:      1,
				CreditScore: 95,
				CreatedTime: now,
			},
		}, int64(2), nil)

	svcCtx := &svc.ServiceContext{
		UserBaseModel: mockUserBase,
	}

	logic := NewListUsersLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.ListUsers(&userv1.ListUsersRequest{
		Page: &commonv1.PageRequest{
			Page:     1,
			PageSize: 20,
		},
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.Users, 2)
	assert.Equal(t, int64(2), resp.Page.Total)
	assert.Equal(t, int32(1), resp.Page.TotalPages)

	// 验证第一个用户
	assert.Equal(t, int64(1001), resp.Users[0].Id)
	assert.Equal(t, "encrypted_13800138001", resp.Users[0].Phone)
	assert.Equal(t, "张三", resp.Users[0].Nickname)

	mockUserBase.AssertExpectations(t)
}

// TestListUsers_WithKeyword 测试带关键词搜索
func TestListUsers_WithKeyword(t *testing.T) {
	// Setup
	mockUserBase := new(MockUserBaseModel)
	keyword := "张三"

	mockUserBase.On("FindPage", mock.Anything, keyword, "", (*int64)(nil), int32(1), int32(20)).
		Return([]*model.UserBase{
			{
				Id:       1001,
				Phone:    "encrypted_13800138001",
				Nickname: sql.NullString{String: "张三", Valid: true},
				Status:   1,
			},
		}, int64(1), nil)

	svcCtx := &svc.ServiceContext{
		UserBaseModel: mockUserBase,
	}

	logic := NewListUsersLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.ListUsers(&userv1.ListUsersRequest{
		Page: &commonv1.PageRequest{
			Page:     1,
			PageSize: 20,
		},
		Keyword: &keyword,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, "张三", resp.Users[0].Nickname)

	mockUserBase.AssertExpectations(t)
}

// TestListUsers_WithStatus 测试状态筛选
func TestListUsers_WithStatus(t *testing.T) {
	// Setup
	mockUserBase := new(MockUserBaseModel)
	status := int32(1) // 正常状态
	statusInt64 := int64(1)

	mockUserBase.On("FindPage", mock.Anything, "", "", &statusInt64, int32(1), int32(20)).
		Return([]*model.UserBase{
			{
				Id:       1001,
				Phone:    "encrypted_13800138001",
				Nickname: sql.NullString{String: "张三", Valid: true},
				Status:   1,
			},
		}, int64(1), nil)

	svcCtx := &svc.ServiceContext{
		UserBaseModel: mockUserBase,
	}

	logic := NewListUsersLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.ListUsers(&userv1.ListUsersRequest{
		Page: &commonv1.PageRequest{
			Page:     1,
			PageSize: 20,
		},
		Status: &status,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, int32(1), resp.Users[0].Status) // Status 是 int32

	mockUserBase.AssertExpectations(t)
}

// TestListUsers_EmptyResult 测试空结果
func TestListUsers_EmptyResult(t *testing.T) {
	// Setup
	mockUserBase := new(MockUserBaseModel)

	mockUserBase.On("FindPage", mock.Anything, "", "", (*int64)(nil), int32(1), int32(20)).
		Return([]*model.UserBase{}, int64(0), nil)

	svcCtx := &svc.ServiceContext{
		UserBaseModel: mockUserBase,
	}

	logic := NewListUsersLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.ListUsers(&userv1.ListUsersRequest{
		Page: &commonv1.PageRequest{
			Page:     1,
			PageSize: 20,
		},
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Users)
	assert.Equal(t, int64(0), resp.Page.Total)

	mockUserBase.AssertExpectations(t)
}

// TestListUsers_Pagination 测试分页
func TestListUsers_Pagination(t *testing.T) {
	// Setup
	mockUserBase := new(MockUserBaseModel)

	// 第二页，每页10条
	mockUserBase.On("FindPage", mock.Anything, "", "", (*int64)(nil), int32(2), int32(10)).
		Return([]*model.UserBase{
			{Id: 1011, Phone: "phone11", Status: 1},
			{Id: 1012, Phone: "phone12", Status: 1},
		}, int64(25), nil) // 总共25条

	svcCtx := &svc.ServiceContext{
		UserBaseModel: mockUserBase,
	}

	logic := NewListUsersLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.ListUsers(&userv1.ListUsersRequest{
		Page: &commonv1.PageRequest{
			Page:     2,
			PageSize: 10,
		},
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(2), resp.Page.Page)
	assert.Equal(t, int32(10), resp.Page.PageSize)
	assert.Equal(t, int64(25), resp.Page.Total)
	assert.Equal(t, int32(3), resp.Page.TotalPages) // 25 / 10 = 3 页

	mockUserBase.AssertExpectations(t)
}
