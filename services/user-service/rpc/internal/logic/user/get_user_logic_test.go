package user

import (
	"context"
	"testing"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 3.6: GetUser 同屋互见
// =============================================================================
func TestGetUser_NoViewer_Masked(t *testing.T) {
	// 无 viewer → 脱敏 + 无房屋号（默认安全）
	svc := testSvc(t)
	ub := userBaseModel(svc)
	createTestUser(t, ub, 1001, "13812341234")

	logic := NewGetUserLogic(context.Background(), svc)
	resp, err := logic.GetUser(&userv1.GetUserRequest{Id: 1001})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "138****1234", resp.User.Phone)
	assert.Nil(t, resp.SameHouse)
}

func TestGetUser_Self_PlaintextAndOwnHouse(t *testing.T) {
	// 查看自身 → 明文 + 自身房屋号
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1002, "13812341234")
	createTestMembershipAt(t, mm, 1, 1002, 2001, 3, 2, 1501)

	logic := NewGetUserLogic(context.Background(), svc)
	resp, err := logic.GetUser(&userv1.GetUserRequest{Id: 1002, ViewerId: int64Ptr(1002)})
	require.NoError(t, err)
	assert.Equal(t, "13812341234", resp.User.Phone)
	require.NotNil(t, resp.SameHouse)
	assert.True(t, resp.SameHouse.SameHouse)
	assert.Equal(t, int32(3), resp.SameHouse.Building)
	assert.Equal(t, int32(2), resp.SameHouse.Unit)
	assert.Equal(t, int32(1501), resp.SameHouse.Room)
}

func TestGetUser_SameHouse_PlaintextAndHouse(t *testing.T) {
	// 同屋 → 明文 + 房屋号
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1003, "13812341234")
	createTestUser(t, ub, 1004, "13812349999")
	createTestMembershipAt(t, mm, 1, 1003, 2001, 1, 1, 301)
	createTestMembershipAt(t, mm, 2, 1004, 2001, 1, 1, 301)

	logic := NewGetUserLogic(context.Background(), svc)
	resp, err := logic.GetUser(&userv1.GetUserRequest{Id: 1003, ViewerId: int64Ptr(1004)})
	require.NoError(t, err)
	assert.Equal(t, "13812341234", resp.User.Phone)
	require.NotNil(t, resp.SameHouse)
	assert.True(t, resp.SameHouse.SameHouse)
	assert.Equal(t, int32(1), resp.SameHouse.Building)
}

func TestGetUser_NotSameHouse_Masked(t *testing.T) {
	// 非同屋 → 脱敏 + 无房屋号
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1005, "13812341234")
	createTestUser(t, ub, 1006, "13812349999")
	createTestMembershipAt(t, mm, 1, 1005, 2001, 1, 1, 301)
	createTestMembershipAt(t, mm, 2, 1006, 2001, 1, 1, 302)

	logic := NewGetUserLogic(context.Background(), svc)
	resp, err := logic.GetUser(&userv1.GetUserRequest{Id: 1005, ViewerId: int64Ptr(1006)})
	require.NoError(t, err)
	assert.Equal(t, "138****1234", resp.User.Phone)
	require.NotNil(t, resp.SameHouse)
	assert.False(t, resp.SameHouse.SameHouse)
	assert.Equal(t, int32(0), resp.SameHouse.Building)
}

func TestGetUser_DecryptFailFallback_ReturnsOriginal(t *testing.T) {
	// 解密失败兜底返回原值（非手机号不脱敏成脏数据）
	svc := testSvc(t)
	ub := userBaseModel(svc)
	createTestUser(t, ub, 1007, "ENCRYPTED_BLOB_NOT_PHONE")

	logic := NewGetUserLogic(context.Background(), svc)
	resp, err := logic.GetUser(&userv1.GetUserRequest{Id: 1007})
	require.NoError(t, err)
	assert.Equal(t, "ENCRYPTED_BLOB_NOT_PHONE", resp.User.Phone)
}
