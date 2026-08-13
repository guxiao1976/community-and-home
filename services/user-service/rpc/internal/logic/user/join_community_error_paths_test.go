package user

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// JoinCommunity 错误路径 + 分支补齐（变异测试补测）
// =============================================================================

func TestJoinCommunity_UserNotFound_Returns10001(t *testing.T) {
	// J-E-01: 用户不存在 → 10001
	svc, _ := certTestSvc(t)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 9999, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10001), resp.Base.Code)
}

func TestJoinCommunity_FindUserError(t *testing.T) {
	// J-E-02: FindOne 用户返回非 ErrNotFound 错误 → 透传
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	ub.findErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_CountActiveError(t *testing.T) {
	// J-E-03: CountActiveByUserId 错误 → 透传
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	mm.countActiveErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_FindByUserCommError(t *testing.T) {
	// J-E-04: FindByUserAndCommunity 返回非 ErrNotFound 错误 → 透传
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	mm.byUserCommErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_YearlyCountError(t *testing.T) {
	// J-E-05: CountDistinctCommunitiesThisYear 错误 → 透传
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	expectNoVerifiedRoles(permMock)
	mm.countYearErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_LifetimeCountError(t *testing.T) {
	// J-E-06: CountDistinctCommunities 错误 → 透传
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	expectNoVerifiedRoles(permMock)
	mm.countDistinctErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_AddressCountError(t *testing.T) {
	// J-E-07: CountActiveByAddress 错误 → 透传
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	expectNoVerifiedRoles(permMock)
	mm.countAddrErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_ReactivateUpdateStatusError(t *testing.T) {
	// J-E-08: 重新激活时 UpdateBindStatus 错误 → 透传
	svc, _ := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	ms := createTestMembership(t, mm, 5001, 1001, 2001)
	ms.BindStatus = model.MembershipBindStatusLeft
	mm.updateStatusErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_InsertError(t *testing.T) {
	// J-E-09: Insert membership 错误 → 透传
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	expectNoVerifiedRoles(permMock)
	mm.insertErr = errors.New("db down")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
}

func TestJoinCommunity_ReactivateAssignFailure_Compensates(t *testing.T) {
	// J-E-10: 重新激活后自动授权失败 → 补偿回 left 并返回错误
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	ms := createTestMembership(t, mm, 5001, 1001, 2001)
	ms.BindStatus = model.MembershipBindStatusLeft

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("permission down")).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
	// 补偿：membership 恢复为 left
	assert.Equal(t, int64(model.MembershipBindStatusLeft), ms.BindStatus, "授权失败后应补偿回 left")
}

func TestJoinCommunity_VerifiedOwnerInTarget_SkipsYearlyLimit(t *testing.T) {
	// J-E-11: 目标小区已认证 owner → 豁免年度频次限制
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	// 今年已首次加入 3 个不同小区（若未认证将触发 10012）；置 left 避免触发活跃数上限
	for cid := int64(4000); cid <= 4002; cid++ {
		ms := createTestMembership(t, mm, 6000+cid, 1001, cid)
		ms.BindStatus = model.MembershipBindStatusLeft
	}

	// 目标小区 2001 的 owner 已认证（status=2, scope=2001）
	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 2, 2001)}, nil).Times(1)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "已认证 owner 应豁免年度频次限制")
}

func TestJoinCommunity_VerifiedNonOwnerRole_YearlyApplies(t *testing.T) {
	// J-E-12: 已认证但角色非 owner/tenant（grid_worker）→ 不豁免年度限制
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	for cid := int64(4000); cid <= 4002; cid++ {
		ms := createTestMembership(t, mm, 6000+cid, 1001, cid)
		ms.BindStatus = model.MembershipBindStatusLeft
	}

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(4, "grid_worker", 2, 2001)}, nil).AnyTimes()

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10012), resp.Base.Code, "grid_worker 认证不豁免年度限制")
}

func TestJoinCommunity_GetUserRolesError_YearlyApplies(t *testing.T) {
	// J-E-13: GetUserRoles 错误 → 视为未认证 → 年度限制生效
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	for cid := int64(4000); cid <= 4002; cid++ {
		ms := createTestMembership(t, mm, 6000+cid, 1001, cid)
		ms.BindStatus = model.MembershipBindStatusLeft
	}

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("permission down")).AnyTimes()

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10012), resp.Base.Code)
}

func TestJoinCommunity_PermissionClientNil_YearlyApplies(t *testing.T) {
	// J-E-14: PermissionClient 为 nil → 视为未认证 → 年度限制生效
	svc := testSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	for cid := int64(4000); cid <= 4002; cid++ {
		ms := createTestMembership(t, mm, 6000+cid, 1001, cid)
		ms.BindStatus = model.MembershipBindStatusLeft
	}

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10012), resp.Base.Code)
}

func TestJoinCommunity_RoleCodeNotFound_AssignFails(t *testing.T) {
	// J-E-15: 角色表无 owner → assignCommunityRole 失败 → 补偿 left 并返回错误
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1001, "phone_1001")
	expectNoVerifiedRoles(permMock)
	// 角色表只有 tenant/registered_user，没有 owner
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(&permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{{Id: 5, Code: "tenant"}, {Id: 9, Code: "registered_user"}},
	}, nil).AnyTimes()

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.Error(t, err)
	require.Nil(t, resp)
	ms, ferr := mm.FindByUserAndCommunity(context.Background(), 1001, 2001)
	require.NoError(t, ferr)
	assert.Equal(t, int64(model.MembershipBindStatusLeft), ms.BindStatus)
}

func TestJoinCommunity_UpdateDefault_PrefsValidNoDefault(t *testing.T) {
	// J-E-16: preferences 为合法 JSON 但不含 default_community_id → 应设置默认小区
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	u := createTestUser(t, ub, 1001, "phone_1001")
	u.Preferences.String = `{"other":"x"}`
	u.Preferences.Valid = true

	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	updated, _ := ub.FindOne(context.Background(), 1001)
	assert.Contains(t, updated.Preferences.String, "default_community_id", "应写入 default_community_id")
}

func TestJoinCommunity_UpdateDefault_PrefsInvalidJSON(t *testing.T) {
	// J-E-17: preferences 为非法 JSON → 应覆盖设置默认小区
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	u := createTestUser(t, ub, 1001, "phone_1001")
	u.Preferences.String = "not-json"
	u.Preferences.Valid = true

	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	updated, _ := ub.FindOne(context.Background(), 1001)
	assert.Contains(t, updated.Preferences.String, "default_community_id", "非法 JSON 应被覆盖")
}
