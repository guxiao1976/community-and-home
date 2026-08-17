package user

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 用户拍板：joinCommunity 房号/权属改为可选
// - 全部提供（ownership∈{OWNED,RENTED} 且 building/unit/room>0）→ 现状：建带房号 membership + 自动授权 owner/tenant
// - 全不提供（无房号+无权属）→ 仅建 membership（building/unit/room=0, bind_status=active），不自动授权
//   （网格员/物业管理员等角色后续 applyRole）
// - 部分提供（只给了房号没给权属，或反之）→ 10040 参数错误
// =============================================================================

func TestJoinResidenceProvided(t *testing.T) {
	cases := []struct {
		name             string
		ownership        userv1.CommunityOwnership
		building, unit   int64
		room             int64
		wantHasResidence bool
		wantRoleCode     string
		wantOK           bool
	}{
		{name: "全部提供-自有", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED, building: 1, unit: 2, room: 301, wantHasResidence: true, wantRoleCode: model.RoleCodeOwner, wantOK: true},
		{name: "全部提供-租住", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_RENTED, building: 3, unit: 1, room: 502, wantHasResidence: true, wantRoleCode: model.RoleCodeTenant, wantOK: true},
		{name: "全不提供", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_UNSPECIFIED, building: 0, unit: 0, room: 0, wantHasResidence: false, wantRoleCode: "", wantOK: true},
		{name: "部分-仅权属无房号", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED, building: 0, unit: 0, room: 0, wantHasResidence: false, wantRoleCode: "", wantOK: false},
		{name: "部分-仅楼号", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_UNSPECIFIED, building: 1, unit: 0, room: 0, wantHasResidence: false, wantRoleCode: "", wantOK: false},
		{name: "部分-仅单元号", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_UNSPECIFIED, building: 0, unit: 2, room: 0, wantHasResidence: false, wantRoleCode: "", wantOK: false},
		{name: "部分-仅房号", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_UNSPECIFIED, building: 0, unit: 0, room: 301, wantHasResidence: false, wantRoleCode: "", wantOK: false},
		{name: "部分-有房号无权属", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_UNSPECIFIED, building: 1, unit: 2, room: 301, wantHasResidence: false, wantRoleCode: "", wantOK: false},
		{name: "部分-缺房号", ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED, building: 1, unit: 2, room: 0, wantHasResidence: false, wantRoleCode: "", wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hasResidence, roleCode, ok := joinResidenceProvided(tc.ownership, tc.building, tc.unit, tc.room)
			assert.Equal(t, tc.wantOK, ok, "ok 判定错误")
			assert.Equal(t, tc.wantHasResidence, hasResidence, "hasResidence 判定错误")
			assert.Equal(t, tc.wantRoleCode, roleCode, "自动授权角色映射错误")
		})
	}
}

func TestJoinCommunity_NoResidence_CreatesMembershipNoAutoGrant(t *testing.T) {
	// 无房号 join（无房号+无权属）→ 仅建 membership（building/unit/room=0, bind_status=active），不调 assignCommunityRole
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1101, "phone_1101")
	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).
		Return(&permissionv1.AssignRoleResponse{}, nil).Times(0)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1101, CommunityId: 2001, // 无房号、无权属
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "无房号 join 应成功")

	// membership 已建且 active，地址为 0
	count, _ := mm.CountActiveByUserId(context.Background(), 1101)
	assert.Equal(t, int64(1), count, "无房号 join 应建 membership")
	ms, ferr := mm.FindByUserAndCommunity(context.Background(), 1101, 2001)
	require.NoError(t, ferr)
	assert.Equal(t, int64(model.MembershipBindStatusActive), ms.BindStatus)
	assert.Equal(t, 0, ms.Building, "无房号 membership building=0")
	assert.Equal(t, 0, ms.Unit, "无房号 membership unit=0")
	assert.Equal(t, 0, ms.Room, "无房号 membership room=0")
}

func TestJoinCommunity_NoResidence_NoAutoGrant_OnReactivate(t *testing.T) {
	// 无房号重新激活（曾退出）→ 成功、地址清 0、不自动授权
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1102, "phone_1102")
	ms := createTestMembershipAt(t, mm, 7100, 1102, 2002, 1, 2, 301)
	ms.BindStatus = model.MembershipBindStatusLeft

	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).
		Return(&permissionv1.AssignRoleResponse{}, nil).Times(0)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1102, CommunityId: 2002, // 无房号、无权属
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "无房号重新激活应成功")
	assert.Equal(t, int64(model.MembershipBindStatusActive), ms.BindStatus)
	assert.Equal(t, 0, ms.Building, "重新激活后地址应清空为 0")
	assert.Equal(t, 0, ms.Unit, "重新激活后地址应清空为 0")
	assert.Equal(t, 0, ms.Room, "重新激活后地址应清空为 0")
}

func TestJoinCommunity_NoResidence_RespectsMaxCommunities(t *testing.T) {
	// 无房号 join 仍受「最多加入小区数」上限约束（校验不因房号可选而跳过）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)
	createTestUser(t, ub, 1103, "phone_1103")
	for cid := int64(2001); cid <= 2003; cid++ {
		createTestMembership(t, mm, 7200+cid, 1103, cid)
	}
	expectNoVerifiedRoles(permMock)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1103, CommunityId: 2004, // 无房号
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10006), resp.Base.Code, "无房号 join 也应受小区数上限约束")
}

func TestJoinCommunity_Partial_RoomOnly_NoOwnership_Returns10040(t *testing.T) {
	// 部分提供（有房号没权属）→ 10040，不建 membership
	svc, _ := certTestSvc(t)
	mm := membershipModel(svc)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1104, CommunityId: 2001, Building: 1, Unit: 2, Room: 301, // 缺权属
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10040), resp.Base.Code, "部分提供（有房号无权属）应返回 10040")
	count, _ := mm.CountActiveByUserId(context.Background(), 1104)
	assert.Equal(t, int64(0), count, "部分提供不得创建 membership")
}

func TestJoinCommunity_Partial_OwnershipOnly_Returns10040(t *testing.T) {
	// 部分提供（有权属没房号）→ 10040，不建 membership
	svc, _ := certTestSvc(t)
	mm := membershipModel(svc)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1105, CommunityId: 2001, Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED, // 无房号
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10040), resp.Base.Code, "部分提供（有权属没房号）应返回 10040")
	count, _ := mm.CountActiveByUserId(context.Background(), 1105)
	assert.Equal(t, int64(0), count, "部分提供不得创建 membership")
}
