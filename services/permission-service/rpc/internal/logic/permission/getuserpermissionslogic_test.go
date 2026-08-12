package permission

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGetUserPermissions_UnverifiedOwner_PublishCodesIncluded
// T1.5: GetUserPermissions 改用 FindActiveRolesByUserId（status IN (0,1,2)），未认证业主（status=0）的发布权限码必须出现在返回 codes 中
// SEE: [[is-system-no-permission-shortcut]] — 权限由 rel_role_permission 配置决定，未认证业主不因 is_system 而获全权限，但发布类（min_verf_level=0）须在列
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — QA 补测，RED 摘录见 CHANGELOG 2026-08-12 补测节
// 断言：status=0 grant 经 FindActiveRolesByUserId → FindByRoleId → FindByIds，发布权限码在返回 codes 中
func TestGetUserPermissions_UnverifiedOwner_PublishCodesIncluded(t *testing.T) {
	mockUserRole := new(MockUserRoleModel)
	mockRolePerm := new(MockRolePermissionModel)
	mockPerm := new(MockPermissionModel)

	// 未认证业主 grant（status=0，scope community:100）
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1001)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, RoleCode: "owner", ScopeType: "community", ScopeId: 100, URStatus: 0},
	}, nil)

	// owner 角色关联发布权限
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil)

	// 权限定义
	mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{
		{Id: 435, Code: "community:lostfound:create-api", Type: 3, MinVerfLevel: 0},
	}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
	}

	logic := NewGetUserPermissionsLogic(context.Background(), svcCtx)
	resp, err := logic.GetUserPermissions(&permissionv1.GetUserPermissionsRequest{UserId: 1001})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// GREEN：未认证业主发布权限码必须在列（T1.5 行为）
	assert.Contains(t, resp.PermissionCodes, "community:lostfound:create-api",
		"未认证业主发布权限码必须在列")

	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}

// TestGetUserPermissions_NoGrants_CodesNil — 空 grants 时返回 PermissionCodes=nil（不报错）
func TestGetUserPermissions_NoGrants_CodesNil(t *testing.T) {
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(999)).Return([]*model.UserRoleWithInfo{}, nil)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole}

	logic := NewGetUserPermissionsLogic(context.Background(), svcCtx)
	resp, err := logic.GetUserPermissions(&permissionv1.GetUserPermissionsRequest{UserId: 999})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.PermissionCodes)

	mockUserRole.AssertExpectations(t)
}

// TestGetUserPermissions_MultipleRoles_DedupeCodes — 多角色共享权限 ID 时，返回 codes 无重复
func TestGetUserPermissions_MultipleRoles_DedupeCodes(t *testing.T) {
	mockUserRole := new(MockUserRoleModel)
	mockRolePerm := new(MockRolePermissionModel)
	mockPerm := new(MockPermissionModel)

	// owner（status=0）+ committee（status=2）共享发布权限
	mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1002)).Return([]*model.UserRoleWithInfo{
		{RoleId: 1, RoleCode: "owner", ScopeType: "community", ScopeId: 100, URStatus: 0},
		{RoleId: 6, RoleCode: "committee", ScopeType: "community", ScopeId: 100, URStatus: 2},
	}, nil)

	mockRolePerm.On("FindByRoleId", mock.Anything, int64(1)).Return([]*model.RelRolePermission{
		{RoleId: 1, PermissionId: 435},
	}, nil)
	mockRolePerm.On("FindByRoleId", mock.Anything, int64(6)).Return([]*model.RelRolePermission{
		{RoleId: 6, PermissionId: 435},
		{RoleId: 6, PermissionId: 600},
	}, nil)

	// 实现端 permIdSet map 迭代顺序不保证 → 集合相等匹配
	mockPerm.On("FindByIds", mock.Anything, mock.MatchedBy(func(ids []int64) bool {
		if len(ids) != 2 {
			return false
		}
		got := map[int64]struct{}{ids[0]: {}, ids[1]: {}}
		_, has435 := got[435]
		_, has600 := got[600]
		return has435 && has600
	})).Return([]*model.SysPermission{
		{Id: 435, Code: "community:lostfound:create-api", MinVerfLevel: 0},
		{Id: 600, Code: "committee:election:vote", MinVerfLevel: 2},
	}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel:       mockUserRole,
		RolePermissionModel: mockRolePerm,
		PermissionModel:     mockPerm,
	}

	logic := NewGetUserPermissionsLogic(context.Background(), svcCtx)
	resp, err := logic.GetUserPermissions(&permissionv1.GetUserPermissionsRequest{UserId: 1002})

	assert.NoError(t, err)
	assert.Len(t, resp.PermissionCodes, 2, "两个权限码无重复")
	assert.Contains(t, resp.PermissionCodes, "community:lostfound:create-api")
	assert.Contains(t, resp.PermissionCodes, "committee:election:vote")

	mockUserRole.AssertExpectations(t)
	mockRolePerm.AssertExpectations(t)
	mockPerm.AssertExpectations(t)
}
