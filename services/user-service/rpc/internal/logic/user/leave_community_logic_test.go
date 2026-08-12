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
// Task 3.3: LeaveCommunity 撤销授权
// - membership 置 left 后双调 RevokeRole(owner_role_id + tenant_role_id, 'community', community_id)（幂等）
// - 失败 → 恢复 bind_status=active 并返回失败
// =============================================================================

func TestLeaveCommunity_RevokesOwnerAndTenant(t *testing.T) {
	// L-R-01: 退出撤销该小区 owner+tenant 双角色授权
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1001, "phone_1001")
	createTestMembership(t, mm, 5001, 1001, 2001)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()

	var revoked []*permissionv1.RevokeRoleRequest
	permMock.EXPECT().RevokeRole(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.RevokeRoleRequest, _ ...interface{}) (*permissionv1.RevokeRoleResponse, error) {
			revoked = append(revoked, req)
			return &permissionv1.RevokeRoleResponse{}, nil
		}).Times(2)

	logic := NewLeaveCommunityLogic(context.Background(), svc)
	resp, err := logic.LeaveCommunity(&userv1.LeaveCommunityRequest{UserId: 1001, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	// 双调断言：owner(1) + tenant(5)，均指向该小区
	require.Len(t, revoked, 2, "应双调 RevokeRole（owner + tenant）")
	roleIDs := make(map[int64]bool)
	for _, r := range revoked {
		assert.Equal(t, int64(1001), r.UserId)
		require.NotNil(t, r.ScopeType)
		assert.Equal(t, "community", *r.ScopeType)
		require.NotNil(t, r.ScopeId)
		assert.Equal(t, int64(2001), *r.ScopeId)
		roleIDs[r.RoleId] = true
	}
	assert.True(t, roleIDs[1], "应撤销 owner(role_id=1)")
	assert.True(t, roleIDs[5], "应撤销 tenant(role_id=5)")

	// membership 已置 left
	ms, _ := mm.FindByUserAndCommunity(context.Background(), 1001, 2001)
	assert.Equal(t, int64(model.MembershipBindStatusLeft), ms.BindStatus)
}

func TestLeaveCommunity_OtherCommunityPreserved(t *testing.T) {
	// L-R-02: 退出某小区只撤销该小区授权，其他小区 membership 保留
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1002, "phone_1002")
	createTestMembership(t, mm, 5002, 1002, 2001)
	createTestMembership(t, mm, 5003, 1002, 2002)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()

	permMock.EXPECT().RevokeRole(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.RevokeRoleRequest, _ ...interface{}) (*permissionv1.RevokeRoleResponse, error) {
			// 只允许撤销 2001（不得触碰 2002）
			assert.Equal(t, int64(2001), *req.ScopeId, "不得撤销其他小区的角色")
			return &permissionv1.RevokeRoleResponse{}, nil
		}).Times(2)

	logic := NewLeaveCommunityLogic(context.Background(), svc)
	resp, err := logic.LeaveCommunity(&userv1.LeaveCommunityRequest{UserId: 1002, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	ms1, _ := mm.FindByUserAndCommunity(context.Background(), 1002, 2001)
	assert.Equal(t, int64(model.MembershipBindStatusLeft), ms1.BindStatus, "退出的 2001 应 left")
	ms2, _ := mm.FindByUserAndCommunity(context.Background(), 1002, 2002)
	assert.Equal(t, int64(model.MembershipBindStatusActive), ms2.BindStatus, "未退出的 2002 应保留 active")
}

func TestLeaveCommunity_RevokeFailure_RestoresAndFails(t *testing.T) {
	// L-R-03: 撤销失败 → leave 失败并恢复 bind_status=active
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1003, "phone_1003")
	createTestMembership(t, mm, 5004, 1003, 2001)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockJoinRoles(), nil).AnyTimes()
	permMock.EXPECT().RevokeRole(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("permission service unavailable")).Times(1)

	logic := NewLeaveCommunityLogic(context.Background(), svc)
	resp, err := logic.LeaveCommunity(&userv1.LeaveCommunityRequest{UserId: 1003, CommunityId: 2001})
	require.Error(t, err, "撤销失败必须使 leave 失败")
	require.Nil(t, resp)

	// 补偿：bind_status 恢复为 active
	ms, _ := mm.FindByUserAndCommunity(context.Background(), 1003, 2001)
	assert.Equal(t, int64(model.MembershipBindStatusActive), ms.BindStatus, "撤销失败后应恢复 bind_status=active")
}

func TestLeaveCommunity_DuplicateLeave_Returns10005(t *testing.T) {
	// L-R-04: 重复 leave（已退出）→ 10005，且不重复撤销
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)
	mm := membershipModel(svc)

	createTestUser(t, ub, 1004, "phone_1004")
	ms := createTestMembership(t, mm, 5005, 1004, 2001)
	ms.BindStatus = model.MembershipBindStatusLeft

	permMock.EXPECT().RevokeRole(gomock.Any(), gomock.Any()).
		Return(&permissionv1.RevokeRoleResponse{}, nil).Times(0)

	logic := NewLeaveCommunityLogic(context.Background(), svc)
	resp, err := logic.LeaveCommunity(&userv1.LeaveCommunityRequest{UserId: 1004, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(10005), resp.Base.Code, "重复 leave 应返回 10005")
}
