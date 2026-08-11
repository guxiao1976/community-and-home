package user

import (
	"context"
	"testing"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestBatchUpdateUsers 批量更新用户状态
func TestBatchUpdateUsers(t *testing.T) {
	mockModel := new(MockUserBaseModel)
	mockModel.On("UpdateStatus", mock.Anything, int64(100), int64(2)).Return(nil)
	mockModel.On("UpdateStatus", mock.Anything, int64(101), int64(2)).Return(nil)

	svcCtx := &svc.ServiceContext{UserBaseModel: mockModel}
	l := NewBatchUpdateUsersLogic(context.Background(), svcCtx)

	req := &userv1.BatchUpdateUsersRequest{
		UserIds: []int64{100, 101},
		Status:  2,
	}

	resp, err := l.BatchUpdateUsers(req)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), resp.UpdatedCount)
	mockModel.AssertExpectations(t)
}

// TestBatchUpdateUsers_Empty 空用户列表
func TestBatchUpdateUsers_Empty(t *testing.T) {
	mockModel := new(MockUserBaseModel)
	svcCtx := &svc.ServiceContext{UserBaseModel: mockModel}
	l := NewBatchUpdateUsersLogic(context.Background(), svcCtx)

	req := &userv1.BatchUpdateUsersRequest{
		UserIds: []int64{},
		Status:  2,
	}

	resp, err := l.BatchUpdateUsers(req)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.UpdatedCount)
}
