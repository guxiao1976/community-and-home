package user

import (
	"context"
	"database/sql"
	"testing"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestUpdateUser_StatusChange 验证状态变更能更新用户状态
func TestUpdateUser_StatusChange(t *testing.T) {
	existing := &model.UserBase{
		Id:       4542136688377323520,
		Phone:    "encrypted-phone",
		Nickname: sql.NullString{String: "测试用户", Valid: true},
		Status:   1,
	}

	mockModel := new(MockUserBaseModel)
	mockModel.On("FindOne", mock.Anything, int64(4542136688377323520)).Return(existing, nil)
	mockModel.On("Update", mock.Anything, mock.Anything).Return(nil)

	svcCtx := &svc.ServiceContext{UserBaseModel: mockModel}
	l := NewUpdateUserLogic(context.Background(), svcCtx)

	status := int32(2) // 禁用
	req := &userv1.UpdateUserRequest{
		Id:     4542136688377323520,
		Status: &status,
	}

	resp, err := l.UpdateUser(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// 更新后状态应为 2
	assert.Equal(t, int64(2), existing.Status)
	mockModel.AssertCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestUpdateUser_NotFound 用户不存在时返回业务错误
func TestUpdateUser_NotFound(t *testing.T) {
	mockModel := new(MockUserBaseModel)
	mockModel.On("FindOne", mock.Anything, int64(999)).Return(nil, model.ErrNotFound)

	svcCtx := &svc.ServiceContext{UserBaseModel: mockModel}
	l := NewUpdateUserLogic(context.Background(), svcCtx)

	req := &userv1.UpdateUserRequest{Id: 999}
	resp, err := l.UpdateUser(req)
	assert.NoError(t, err)
	assert.Equal(t, int32(10001), resp.Base.GetCode()) // 用户不存在
}
