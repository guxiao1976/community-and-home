package user

import (
	"context"
	"database/sql"
	"fmt"
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
// §3.2 申请角色 (ApplyRole) 测试
// =============================================================================

func TestApplyRole_Owner(t *testing.T) {
	// U-A-01: 申请业主角色
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	createTestMembership(t, mm, 5001, 1001, 2001)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1001, CommunityId: 2001, RoleCode: "owner",
		Building: "3", Unit: "2", Room: "1501",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "owner", resp.Role.RoleCode)
	assert.Equal(t, int32(model.RoleVerfStatusUnverified), resp.Role.VerfStatus)
	assert.Equal(t, int64(2001), resp.Role.CommunityId)
}

func TestApplyRole_Merchant(t *testing.T) {
	// U-A-03: 申请商家角色（不绑小区）
	svc := testSvc(t)
	ub := userBaseModel(svc)

	createTestUser(t, ub, 1002, "phone_1002")

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1002, RoleCode: "merchant",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "merchant", resp.Role.RoleCode)
	assert.Equal(t, int64(0), resp.Role.MembershipId)
	assert.Equal(t, int64(0), resp.Role.CommunityId)
}

func TestApplyRole_NoMembership(t *testing.T) {
	// U-A-04: 未加入小区直接申请角色
	svc := testSvc(t)
	ub := userBaseModel(svc)

	createTestUser(t, ub, 1003, "phone_1003")
	// 不创建 membership

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1003, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10005), resp.Base.Code)
}

func TestApplyRole_UserNotFound(t *testing.T) {
	// Q-05: 用户不存在时拒绝申请角色
	svc := testSvc(t)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1099, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10001), resp.Base.Code, "用户不存在")
}

func TestApplyRole_MembershipLeft(t *testing.T) {
	// U-A-05: 已退出小区后申请角色
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1004, "phone_1004")

	ms := createTestMembership(t, mm, 5002, 1004, 2001)
	ms.BindStatus = model.MembershipBindStatusLeft // 模拟已退出

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1004, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10005), resp.Base.Code)
}

func TestApplyRole_Duplicate(t *testing.T) {
	// U-A-06: 同一角色重复申请
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1005, "phone_1005")
	createTestMembership(t, mm, 5003, 1005, 2001)
	createTestRole(t, roleModel(svc), 1, 1005, 5003, 2001, "owner", model.RoleVerfStatusUnverified)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1005, CommunityId: 2001, RoleCode: "owner",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10008), resp.Base.Code, "该角色已存在")
}

func TestApplyRole_DifferentRolesInCommunity(t *testing.T) {
	// U-A-07: 同一小区申请不同角色
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	rm := roleModel(svc)

	createTestUser(t, ub, 1006, "phone_1006")
	createTestMembership(t, mm, 5004, 1006, 2001)
	createTestRole(t, rm, 1, 1006, 5004, 2001, "owner", model.RoleVerfStatusUnverified)

	// 再申请 committee
	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1006, CommunityId: 2001, RoleCode: "committee",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "不同 role_code 应允许")
}

func TestApplyRole_GridWorker(t *testing.T) {
	// U-A-08: 申请网格员角色
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1007, "phone_1007")
	createTestMembership(t, mm, 5005, 1007, 2001)

	logic := NewApplyRoleLogic(context.Background(), svc)
	resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
		UserId: 1007, CommunityId: 2001, RoleCode: "grid_worker",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, "grid_worker", resp.Role.RoleCode)
}

func TestApplyRole_RoleCodes(t *testing.T) {
	// 覆盖所有角色编码
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	tests := []struct {
		roleCode string
	}{
		{"owner"}, {"tenant"}, {"grid_worker"},
		{"community_admin"}, {"property_admin"}, {"committee"},
	}

	for _, tt := range tests {
		uid := int64(2000 + len(membershipModel(svc).data))
		cid := int64(3000 + len(membershipModel(svc).data))
		createTestUser(t, ub, uid, fmt.Sprintf("phone_%d", uid))
		createTestMembership(t, mm, uid, uid, cid)

		logic := NewApplyRoleLogic(context.Background(), svc)
		resp, err := logic.ApplyRole(&userv1.ApplyRoleRequest{
			UserId: uid, CommunityId: cid, RoleCode: tt.roleCode,
		})
		require.NoError(t, err)
		assert.Equal(t, int32(0), resp.Base.Code, "role=%s", tt.roleCode)
		assert.Equal(t, tt.roleCode, resp.Role.RoleCode)
	}
}

// =============================================================================
// §3.6 权限校验 (CheckAccess) 测试
// =============================================================================

func TestCheckAccess_OwnerAllowed(t *testing.T) {
	// U-C-01: owner 角色校验通过
	svc := testSvc(t)
	mm := membershipModel(svc)
	rm := roleModel(svc)

	createTestMembership(t, mm, 5001, 6001, 2001)
	createTestRole(t, rm, 1, 6001, 5001, 2001, "owner", model.RoleVerfStatusApproved)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6001, RoleCodes: []string{"owner"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.True(t, resp.Allowed)
	assert.Equal(t, "owner", resp.MatchedRole)
}

func TestCheckAccess_WrongRole(t *testing.T) {
	// U-C-02: 角色不匹配
	svc := testSvc(t)
	mm := membershipModel(svc)
	rm := roleModel(svc)

	createTestMembership(t, mm, 5002, 6002, 2001)
	createTestRole(t, rm, 1, 6002, 5002, 2001, "owner", model.RoleVerfStatusApproved)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6002, RoleCodes: []string{"community_admin"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckAccess_WrongCommunity(t *testing.T) {
	// U-C-03: 小区不匹配
	svc := testSvc(t)
	mm := membershipModel(svc)
	rm := roleModel(svc)

	createTestMembership(t, mm, 5003, 6003, 2001)
	createTestRole(t, rm, 1, 6003, 5003, 2001, "owner", model.RoleVerfStatusApproved)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6003, RoleCodes: []string{"owner"}, CommunityId: 9999,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed, "社区范围不对应拒绝")
}

func TestCheckAccess_UnverifiedRole(t *testing.T) {
	// U-C-04: 未认证角色不通过
	svc := testSvc(t)
	mm := membershipModel(svc)
	rm := roleModel(svc)

	createTestMembership(t, mm, 5004, 6004, 2001)
	createTestRole(t, rm, 1, 6004, 5004, 2001, "owner", model.RoleVerfStatusPending) // 待审核

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6004, RoleCodes: []string{"owner"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed, "未认证角色不应通过权限校验")
}

func TestCheckAccess_MerchantGlobal(t *testing.T) {
	// U-C-05: merchant 全局角色 (community_id=0)
	svc := testSvc(t)
	rm := roleModel(svc)

	// merchant 角色 membership_id=NULL, community_id=0
	createTestRole(t, rm, 1, 6005, 0, 0, "merchant", model.RoleVerfStatusApproved)
	rm.data[1].MembershipId = sql.NullInt64{} // 确保 NULL

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6005, RoleCodes: []string{"merchant"}, CommunityId: 0,
	})

	require.NoError(t, err)
	assert.True(t, resp.Allowed)
	assert.Equal(t, "merchant", resp.MatchedRole)
}

func TestCheckAccess_ExpiredRole(t *testing.T) {
	// U-C-07: 角色已过期
	svc := testSvc(t)
	mm := membershipModel(svc)
	rm := roleModel(svc)

	createTestMembership(t, mm, 5007, 6007, 2001)
	createTestRole(t, rm, 1, 6007, 5007, 2001, "owner", model.RoleVerfStatusExpired)

	logic := NewCheckAccessLogic(context.Background(), svc)
	resp, err := logic.CheckAccess(&userv1.CheckAccessRequest{
		UserId: 6007, RoleCodes: []string{"owner"}, CommunityId: 2001,
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed, "已过期角色不应通过权限校验")
}
