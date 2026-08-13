package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GetUserMemberships 查询用户小区成员关系
// =============================================================================

func TestGetUserMemberships_Success(t *testing.T) {
	// G-M-01: 正常返回用户所有活跃 membership
	svc := testSvc(t)
	mm := membershipModel(svc)

	createTestMembership(t, mm, 5001, 1001, 2001)
	createTestMembership(t, mm, 5002, 1001, 2002)
	// 非活跃 membership 不应返回（FindByUserId mock 只返回 active）
	left := createTestMembership(t, mm, 5003, 1001, 2003)
	left.BindStatus = model.MembershipBindStatusLeft

	logic := NewGetUserMembershipsLogic(context.Background(), svc)
	resp, err := logic.GetUserMemberships(&userv1.GetUserMembershipsRequest{UserId: 1001})

	require.NoError(t, err)
	require.Len(t, resp.Memberships, 2)
	ids := map[int64]bool{resp.Memberships[0].Id: true, resp.Memberships[1].Id: true}
	assert.True(t, ids[5001])
	assert.True(t, ids[5002])
	assert.Equal(t, int32(0), resp.Base.Code)
}

func TestGetUserMemberships_LeaveTimePreserved(t *testing.T) {
	// G-M-02: LeaveTime 有效时应回填 proto（重新激活场景）
	svc := testSvc(t)
	mm := membershipModel(svc)

	ms := createTestMembership(t, mm, 5001, 1001, 2001)
	ms.LeaveTime = sql.NullTime{Time: time.Now(), Valid: true}

	logic := NewGetUserMembershipsLogic(context.Background(), svc)
	resp, err := logic.GetUserMemberships(&userv1.GetUserMembershipsRequest{UserId: 1001})

	require.NoError(t, err)
	require.Len(t, resp.Memberships, 1)
	assert.NotZero(t, resp.Memberships[0].LeaveTime, "LeaveTime 应被回填")
}

func TestGetUserMemberships_Empty(t *testing.T) {
	// G-M-03: 无 membership → 返回空列表（非 nil）
	svc := testSvc(t)

	logic := NewGetUserMembershipsLogic(context.Background(), svc)
	resp, err := logic.GetUserMemberships(&userv1.GetUserMembershipsRequest{UserId: 9999})

	require.NoError(t, err)
	require.NotNil(t, resp.Memberships)
	assert.Empty(t, resp.Memberships)
}

func TestGetUserMemberships_FindError(t *testing.T) {
	// G-M-04: FindByUserId 返回错误 → 透传错误
	svc := testSvc(t)
	mm := membershipModel(svc)
	mm.findErr = errors.New("db down")

	logic := NewGetUserMembershipsLogic(context.Background(), svc)
	resp, err := logic.GetUserMemberships(&userv1.GetUserMembershipsRequest{UserId: 1001})

	require.Error(t, err)
	require.Nil(t, resp)
}
