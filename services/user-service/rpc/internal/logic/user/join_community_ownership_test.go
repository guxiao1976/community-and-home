package user

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	permissionmocks "github.com/guxiao1976/api-proto/gen/go/permission/v1/mocks"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 3.2: JoinCommunity ownership + 自动授权
// - ownership ∈ {OWNED, RENTED}，UNSPECIFIED → 10040
// - membership 落库后 AssignRole(owner|tenant, 'community', community_id, status=0)
// - 授权失败 → 补偿恢复 membership 并返回失败（不留「有成员无 scope」）
// =============================================================================

// mockJoinRoles 构造 owner=1 / tenant=5 / registered_user=9 的角色表响应
func mockJoinRoles() *permissionv1.ListRolesResponse {
	return &permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{
			{Id: 1, Code: "owner"},
			{Id: 2, Code: "property_admin"},
			{Id: 3, Code: "community_admin"},
			{Id: 4, Code: "grid_worker"},
			{Id: 5, Code: "tenant"},
			{Id: 6, Code: "committee"},
			{Id: 9, Code: "registered_user"},
		},
	}
}

// expectNoVerifiedRoles 默认 GetUserRoles 返回空（未认证，触发频次校验路径）
func expectNoVerifiedRoles(m *permissionmocks.MockPermissionServiceClient) {
	m.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).
		Return(&permissionv1.GetUserRolesResponse{}, nil).AnyTimes()
}

func TestJoinCommunity_Owned_AssignsOwnerGrant(t *testing.T) {
	// J-O-01: 自有 → owner 角色自动授权（scope_type='community', scope_id=community_id, status=0）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()

	var assigned *permissionv1.AssignRoleRequest
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.AssignRoleRequest, _ ...interface{}) (*permissionv1.AssignRoleResponse, error) {
			assigned = req
			return &permissionv1.AssignRoleResponse{}, nil
		}).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1001, CommunityId: 2001, Building: 1, Unit: 2, Room: 301,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	require.NotNil(t, assigned, "OWNED 加入应触发 AssignRole")
	assert.Equal(t, int64(1), assigned.RoleId, "owner role_id=1")
	assert.Equal(t, "community", assigned.ScopeType)
	assert.Equal(t, int64(2001), assigned.ScopeId)
	require.NotNil(t, assigned.Status)
	assert.Equal(t, int32(0), *assigned.Status, "自动授权为未认证 status=0")

	// membership 已落库且 active
	count, _ := mm.CountActiveByUserId(context.Background(), 1001)
	assert.Equal(t, int64(1), count)
}

func TestJoinCommunity_Rented_AssignsTenantGrant(t *testing.T) {
	// J-O-02: 租住 → tenant 角色自动授权
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)

	createTestUser(t, ub, 1002, "phone_1002")
	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()

	var assigned *permissionv1.AssignRoleRequest
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.AssignRoleRequest, _ ...interface{}) (*permissionv1.AssignRoleResponse, error) {
			assigned = req
			return &permissionv1.AssignRoleResponse{}, nil
		}).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1002, CommunityId: 2002, Building: 3, Unit: 1, Room: 502,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_RENTED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	require.NotNil(t, assigned, "RENTED 加入应触发 AssignRole")
	assert.Equal(t, int64(5), assigned.RoleId, "tenant role_id=5")
	assert.Equal(t, "community", assigned.ScopeType)
	assert.Equal(t, int64(2002), assigned.ScopeId)
	require.NotNil(t, assigned.Status)
	assert.Equal(t, int32(0), *assigned.Status)
}

func TestJoinCommunity_AssignFailure_CompensatesAndFails(t *testing.T) {
	// J-O-03: 授权失败 → join 失败，且无「有成员无 scope」（membership 被补偿为非 active）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1003, "phone_1003")
	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("permission service unavailable")).Times(1)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1003, CommunityId: 2003, Building: 5, Unit: 1, Room: 601,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.Error(t, err, "授权失败必须使 join 失败")
	require.Nil(t, resp)

	// 补偿：无 active membership（不留「有成员无 scope」）
	count, _ := mm.CountActiveByUserId(context.Background(), 1003)
	assert.Equal(t, int64(0), count, "授权失败后不得有 active membership 而无 scope")
	ms, ferr := mm.FindByUserAndCommunity(context.Background(), 1003, 2003)
	require.NoError(t, ferr)
	assert.Equal(t, int64(model.MembershipBindStatusLeft), ms.BindStatus, "membership 应被补偿为 left")
}

func TestJoinCommunity_MissingOwnership_Returns10040(t *testing.T) {
	// J-O-04: ownership 缺失（UNSPECIFIED）→ 10040，不建 membership
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1004, "phone_1004")

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1004, CommunityId: 2004, Building: 1, Unit: 0, Room: 101,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10040), resp.Base.Code, "ownership 缺失应返回 10040")

	count, _ := mm.CountActiveByUserId(context.Background(), 1004)
	assert.Equal(t, int64(0), count, "ownership 缺失不得创建 membership")
	_ = permMock // UNSPECIFIED 路径不触发授权
}

func TestJoinCommunity_DuplicateActive_NoDoubleAssign(t *testing.T) {
	// J-O-05: 重复加入（已 active）→ 10007，且不重复授权
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1005, "phone_1005")
	createTestMembership(t, mm, 5001, 1005, 2005)
	expectNoVerifiedRoles(permMock)
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).
		Return(&permissionv1.AssignRoleResponse{}, nil).Times(0)

	logic := NewJoinCommunityLogic(context.Background(), svc)
	resp, err := logic.JoinCommunity(&userv1.JoinCommunityRequest{
		UserId: 1005, CommunityId: 2005, Building: 1, Unit: 1, Room: 101,
		Ownership: userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(10007), resp.Base.Code, "重复加入应返回 10007")

	count, _ := mm.CountActiveByUserId(context.Background(), 1005)
	assert.Equal(t, int64(1), count, "不得产生重复 membership")
}
