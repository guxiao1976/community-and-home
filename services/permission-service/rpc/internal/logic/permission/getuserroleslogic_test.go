package permission

import (
	"context"
	"database/sql"
	"testing"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGetUserRoles_Success 测试成功获取用户角色
func TestGetUserRoles_Success(t *testing.T) {
	// Setup
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).
		Return([]*model.UserRoleWithInfo{
			{
				RoleId:      1,
				RoleCode:    "owner",
				RoleName:    "业主",
				Description: "小区业主",
				IsSystem:    0,
				Status:      1,
				Platforms:   "pc,mobile",
				ScopeType:   "community",
				ScopeId:     100,
				URStatus:    2,
			},
			{
				RoleId:      2,
				RoleCode:    "property_admin",
				RoleName:    "物业管理员",
				Description: "物业管理人员",
				IsSystem:    0,
				Status:      1,
				Platforms:   "pc",
				ScopeType:   "building",
				ScopeId:     200,
				URStatus:    2,
			},
		}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
	}

	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{
		UserId: 1001,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.Roles, 2)

	// 验证第一个角色
	assert.Equal(t, int64(1), resp.Roles[0].Role.Id)
	assert.Equal(t, "owner", resp.Roles[0].Role.Code)
	assert.Equal(t, "业主", resp.Roles[0].Role.Name)
	assert.Equal(t, "community", resp.Roles[0].ScopeType)
	assert.Equal(t, int64(100), resp.Roles[0].ScopeId)
	// Platforms 透出断言：owner 双端 → ["pc","mobile"]
	// SEE: [[is-system-no-permission-shortcut]] — platforms 为配置属性，仅透出供 auth 端准入判定
	assert.Equal(t, []string{"pc", "mobile"}, resp.Roles[0].Role.Platforms)

	// 验证第二个角色
	assert.Equal(t, int64(2), resp.Roles[1].Role.Id)
	assert.Equal(t, "property_admin", resp.Roles[1].Role.Code)
	assert.Equal(t, "building", resp.Roles[1].ScopeType)
	// Platforms 透出断言：property_admin 单端 → ["pc"]
	assert.Equal(t, []string{"pc"}, resp.Roles[1].Role.Platforms)

	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_NoRoles 测试用户无角色
func TestGetUserRoles_NoRoles(t *testing.T) {
	// Setup
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).
		Return([]*model.UserRoleWithInfo{}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
	}

	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{
		UserId: 1001,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Nil(t, resp.Roles)

	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_Error 覆盖 FindAllByUserId 返回 error → Roles nil（不 panic）
func TestGetUserRoles_Error(t *testing.T) {
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).Return(nil, assert.AnError)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole}
	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{UserId: 1001})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Nil(t, resp.Roles)
	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_LifecycleTimestamps 覆盖 VerifiedAt/ExpiresAt Valid 分支透传 Unix 秒
func TestGetUserRoles_LifecycleTimestamps(t *testing.T) {
	verifiedAt := time.Now().Add(-24 * time.Hour)
	expiresAt := time.Now().Add(24 * time.Hour)

	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).
		Return([]*model.UserRoleWithInfo{
			{
				RoleId:     1,
				RoleCode:   "owner",
				RoleName:   "业主",
				Status:     1,
				Platforms:  "pc",
				ScopeType:  "community",
				ScopeId:    100,
				URStatus:   2,
				VerifiedAt: sql.NullTime{Time: verifiedAt, Valid: true},
				ExpiresAt:  sql.NullTime{Time: expiresAt, Valid: true},
			},
		}, nil)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole}
	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{UserId: 1001})
	assert.NoError(t, err)
	assert.Len(t, resp.Roles, 1)
	assert.Equal(t, verifiedAt.Unix(), resp.Roles[0].VerifiedAt, "VerifiedAt 应透传 Unix 秒")
	assert.Equal(t, expiresAt.Unix(), resp.Roles[0].ExpiresAt, "ExpiresAt 应透传 Unix 秒")
	assert.Equal(t, int32(2), resp.Roles[0].Status)
	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_SystemRole 测试系统角色
func TestGetUserRoles_SystemRole(t *testing.T) {
	// Setup
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(1001)).
		Return([]*model.UserRoleWithInfo{
			{
				RoleId:      999,
				RoleCode:    "super_admin",
				RoleName:    "超级管理员",
				Description: "系统超级管理员",
				IsSystem:    1, // 系统角色
				Status:      1,
				ScopeType:   "system",
				ScopeId:     0,
			},
		}, nil)

	svcCtx := &svc.ServiceContext{
		UserRoleModel: mockUserRole,
	}

	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	// Execute
	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{
		UserId: 1001,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Len(t, resp.Roles, 1)
	assert.True(t, resp.Roles[0].Role.IsSystem) // 系统角色标记

	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_ErrWithRoles — err 非 nil 但 roles 非空（杀 L27 的 ||→&& 变异）
func TestGetUserRoles_ErrWithRoles(t *testing.T) {
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(3001)).
		Return([]*model.UserRoleWithInfo{{RoleId: 1, RoleCode: "owner", URStatus: 0}}, assert.AnError)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole}
	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{UserId: 3001})
	assert.NoError(t, err) // err 存在时函数吞掉，返回空 Roles
	assert.Len(t, resp.Roles, 0)
	mockUserRole.AssertExpectations(t)
}

// TestGetUserRoles_VerifiedAndExpiresTimestamps
// 覆盖 VerifiedAt.Valid 与 ExpiresAt.Valid 的分支（L37/L40），杀 r.VerifiedAt.Valid 变异族
func TestGetUserRoles_VerifiedAndExpiresTimestamps(t *testing.T) {
	now := time.Now()
	mockUserRole := new(MockUserRoleModel)
	mockUserRole.On("FindAllByUserId", mock.Anything, int64(2001)).
		Return([]*model.UserRoleWithInfo{
			{
				RoleId: 1, RoleCode: "owner", RoleName: "业主",
				URStatus:  2,
				VerifiedAt: sql.NullTime{Time: now, Valid: true}, // 已验证
				ExpiresAt:  sql.NullTime{Time: now.Add(24 * time.Hour), Valid: true},
			},
			{
				RoleId: 2, RoleCode: "registered_user", RoleName: "注册用户",
				URStatus: 2,
				// VerifiedAt/ExpiresAt 均 Invalid
			},
		}, nil)

	svcCtx := &svc.ServiceContext{UserRoleModel: mockUserRole}
	logic := NewGetUserRolesLogic(context.Background(), svcCtx)

	resp, err := logic.GetUserRoles(&permissionv1.GetUserRolesRequest{UserId: 2001})
	assert.NoError(t, err)
	assert.Len(t, resp.Roles, 2)
	// 已验证角色：VerifiedAt/ExpiresAt 填充
	assert.Equal(t, now.Unix(), resp.Roles[0].VerifiedAt)
	assert.Equal(t, now.Add(24*time.Hour).Unix(), resp.Roles[0].ExpiresAt)
	// 未验证角色：均为 0
	assert.Equal(t, int64(0), resp.Roles[1].VerifiedAt)
	assert.Equal(t, int64(0), resp.Roles[1].ExpiresAt)
	mockUserRole.AssertExpectations(t)
}
