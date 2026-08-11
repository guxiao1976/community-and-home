package user

import (
	"context"
	"testing"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// §3.1 加入小区 (JoinCommunity) 测试
// =============================================================================

func TestJoinCommunity_Success_FirstTime(t *testing.T) {
	// U-J-01: 正常首次加入小区
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "encrypted_phone_1001")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.NotZero(t, resp.Membership.Id)
	assert.Equal(t, int64(1001), resp.Membership.UserId)
	assert.Equal(t, int64(2001), resp.Membership.CommunityId)
	assert.Equal(t, int32(1), resp.Membership.BindStatus)

	// 验证 default_community_id 已设置
	u, _ := ub.FindOne(context.Background(), 1001)
	assert.Contains(t, u.Preferences.String, "default_community_id")

	// 验证 membership 记录
	count, _ := mm.CountActiveByUserId(context.Background(), 1001)
	assert.Equal(t, int64(1), count)
}

func TestJoinCommunity_Idempotent(t *testing.T) {
	// U-J-02: 已加入同一小区 — 幂等
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1002, "phone_1002")
	createTestMembership(t, mm, 5001, 1002, 2001)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1002, CommunityId: 2001,
	})

	require.NoError(t, err)
	// 已活跃加入，不能重复，返回 10007
	assert.Equal(t, int32(10007), resp.Base.Code)

	// 不应创建重复记录
	count, _ := mm.CountActiveByUserId(context.Background(), 1002)
	assert.Equal(t, int64(1), count)
}

func TestJoinCommunity_Rejoin(t *testing.T) {
	// U-J-03: 退出后重新加入
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1003, "phone_1003")
	ms := createTestMembership(t, mm, 5002, 1003, 2001)

	// 先退出
	leaveLogic := NewLeaveCommunityLogic(context.Background(), svc)
	lResp, _ := leaveLogic.LeaveCommunity(&userv1.LeaveCommunityRequest{
		UserId: 1003, CommunityId: 2001,
	})
	assert.Equal(t, int32(0), lResp.Base.Code)
	assert.Equal(t, int64(0), ms.BindStatus)

	// 再重新加入
	joinLogic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := joinLogic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1003, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(model.MembershipBindStatusActive), ms.BindStatus)
}

func TestJoinCommunity_MaxFive(t *testing.T) {
	// U-J-04: 已达 5 个小区上限
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1004, "phone_1004")
	// 加入 5 个小区
	for cid := int64(2001); cid <= 2005; cid++ {
		createTestMembership(t, mm, 5000+cid, 1004, cid)
	}

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1004, CommunityId: 2006,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10006), resp.Base.Code)
}

func TestJoinCommunity_DefaultCommunityNotOverwritten(t *testing.T) {
	// U-J-06: 已有 default_community_id 不覆盖
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	u := createTestUser(t, ub, 1005, "phone_1005")
	u.Preferences.String = `{"default_community_id":1001}`
	u.Preferences.Valid = true

	createTestMembership(t, mm, 5010, 1005, 2001)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	_, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1005, CommunityId: 2002,
	})
	require.NoError(t, err)

	// default_community_id 应不变
	updated, _ := ub.FindOne(context.Background(), 1005)
	assert.Contains(t, updated.Preferences.String, "1001")
	assert.NotContains(t, updated.Preferences.String, "2002")
}

// =============================================================================
