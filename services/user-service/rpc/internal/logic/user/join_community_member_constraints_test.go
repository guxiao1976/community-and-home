package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestMembershipAt 创建带指定楼/单元/房号的 membership（测试同屋/每户人数约束用）
func createTestMembershipAt(t *testing.T, m *mockMembershipModel, id, uid, cid int64, building, unit, room int) *model.UserCommunityMembership {
	t.Helper()
	ms := &model.UserCommunityMembership{
		Id: id, UserId: uid, CommunityId: cid, BindStatus: model.MembershipBindStatusActive,
		JoinTime: time.Now(), CreatedTime: time.Now(), UpdatedTime: time.Now(),
		Building: building, Unit: unit, Room: room,
	}
	m.data[ms.Id] = ms
	m.byUserCommIdx[fmt.Sprintf("%d_%d", uid, cid)] = ms.Id
	return ms
}

// =============================================================================
// Task 3.4: JoinCommunity 房屋必填 + 每户 ≤6
// =============================================================================

func TestJoinCommunity_MissingBuilding_Returns10040(t *testing.T) {
	// 缺楼号 → 10040（楼/单元/房号必填）
	svc, _ := certTestSvc(t)
	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 101, CommunityId: 201, Building: 0, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10040), resp.Base.Code)
}

func TestJoinCommunity_MissingUnit_Returns10040(t *testing.T) {
	// 缺单元号 → 10040
	svc, _ := certTestSvc(t)
	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 102, CommunityId: 201, Building: 1, Unit: 0, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10040), resp.Base.Code)
}

func TestJoinCommunity_MissingRoom_Returns10040(t *testing.T) {
	// 缺房号 → 10040
	svc, _ := certTestSvc(t)
	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 103, CommunityId: 201, Building: 1, Unit: 1, Room: 0,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10040), resp.Base.Code)
}

func TestJoinCommunity_House5Members_NewUser_Allowed(t *testing.T) {
	// 房屋 5 人 + 新用户 → 放行（5 < 6）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 104, "phone_104")

	for i := int64(0); i < 5; i++ {
		createTestMembershipAt(t, mm, 1000+i, 2000+i, 2001, 1, 1, 101)
	}

	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 104, CommunityId: 2001, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

func TestJoinCommunity_House6Members_NewUser_Returns10014(t *testing.T) {
	// 房屋 6 人 + 新用户 → 10014（该房屋已满员）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 105, "phone_105")

	for i := int64(0); i < 6; i++ {
		createTestMembershipAt(t, mm, 1100+i, 2100+i, 2001, 1, 1, 102)
	}

	expectNoVerifiedRoles(permMock)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 105, CommunityId: 2001, Building: 1, Unit: 1, Room: 102,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10014), resp.Base.Code)
}

func TestJoinCommunity_HouseLeftMemberNotCounted(t *testing.T) {
	// 退出者不计：5 活跃 + 1 已退出 → 放行
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 106, "phone_106")

	for i := int64(0); i < 5; i++ {
		createTestMembershipAt(t, mm, 1200+i, 2200+i, 2001, 1, 1, 103)
	}
	left := createTestMembershipAt(t, mm, 1205, 2205, 2001, 1, 1, 103)
	left.BindStatus = model.MembershipBindStatusLeft

	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 106, CommunityId: 2001, Building: 1, Unit: 1, Room: 103,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

func TestJoinCommunity_Reactivate_ExcludesSelf(t *testing.T) {
	// 重新激活场景：用户自身旧 membership（已退出）不计入房屋人数
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 107, "phone_107")

	old := createTestMembershipAt(t, mm, 1300, 107, 2001, 1, 1, 104)
	old.BindStatus = model.MembershipBindStatusLeft

	for i := int64(0); i < 5; i++ {
		createTestMembershipAt(t, mm, 1400+i, 2300+i, 2001, 1, 1, 104)
	}

	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 107, CommunityId: 2001, Building: 1, Unit: 1, Room: 104,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

// =============================================================================
// Task 3.5: 终身限制对齐（对全部用户生效）+ 认证粒度 per-community
// =============================================================================

func TestJoinCommunity_VerifiedUser_Lifetime12_Returns10013(t *testing.T) {
	// 认证用户终身达 12 → 10013（终身限制对全部用户生效，不因认证豁免）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 108, "phone_108")

	for i := int64(0); i < 12; i++ {
		ms := createTestMembershipAt(t, mm, 1500+i, 108, 3000+i, 1, 1, 100+int(i))
		ms.BindStatus = model.MembershipBindStatusLeft
	}

	// 目标小区 2001 的 owner 已认证（status=2）
	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 2, 2001)}, nil).AnyTimes()

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 108, CommunityId: 2001, Building: 1, Unit: 1, Room: 201,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10013), resp.Base.Code)
}

func TestJoinCommunity_VerifiedInA_JoiningB_SubjectToYearlyLimit(t *testing.T) {
	// A 小区认证、加入 B 小区 → 受 B 每年限制（per-community 粒度，认证不跨小区豁免）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 109, "phone_109")

	// 今年已首次加入 3 个不同小区：1 个活跃（A 小区）+ 2 个已退出
	createTestMembershipAt(t, mm, 1600, 109, 4000, 1, 1, 301)
	l1 := createTestMembershipAt(t, mm, 1601, 109, 4001, 1, 1, 302)
	l1.BindStatus = model.MembershipBindStatusLeft
	l2 := createTestMembershipAt(t, mm, 1602, 109, 4002, 1, 1, 303)
	l2.BindStatus = model.MembershipBindStatusLeft

	// 用户 109 仅在 A 小区(4000) 认证 owner，目标小区为 B(5000)
	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(1, "owner", 2, 4000)}, nil).AnyTimes()

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 109, CommunityId: 5000, Building: 2, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10012), resp.Base.Code)
}
