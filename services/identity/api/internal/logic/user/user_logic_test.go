package user

import (
	"context"
	"database/sql"
	"testing"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"
	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/stretchr/testify/assert"
)

// --- Mock: AuthUserRoleModel ---

type mockAuthUserRoleModel struct {
	findActiveByUserIdFn    func(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error)
	findOneByUserIdRoleIdFn   func(ctx context.Context, userId, roleId int64) (*model.AuthUserRole, error)
	batchInsertUserRolesFn     func(ctx context.Context, userId int64, roleIds []int64) error
	deleteByUserIdAndRoleIdFn func(ctx context.Context, userId, roleId int64) error
}

func (m *mockAuthUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	return m.findActiveByUserIdFn(ctx, userId)
}
func (m *mockAuthUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*model.AuthUserRole, error) {
	return nil, nil
}
func (m *mockAuthUserRoleModel) FindOne(ctx context.Context, id int64) (*model.AuthUserRole, error) {
	return nil, model.ErrNotFound
}
func (m *mockAuthUserRoleModel) FindOneByUserIdRoleId(ctx context.Context, userId, roleId int64) (*model.AuthUserRole, error) {
	return m.findOneByUserIdRoleIdFn(ctx, userId, roleId)
}
func (m *mockAuthUserRoleModel) Insert(ctx context.Context, data *model.AuthUserRole) (sql.Result, error) {
	return nil, nil
}
func (m *mockAuthUserRoleModel) Update(ctx context.Context, data *model.AuthUserRole) error { return nil }
func (m *mockAuthUserRoleModel) Delete(ctx context.Context, id int64) error                   { return nil }
func (m *mockAuthUserRoleModel) BatchInsertUserRoles(ctx context.Context, userId int64, roleIds []int64) error {
	return m.batchInsertUserRolesFn(ctx, userId, roleIds)
}
func (m *mockAuthUserRoleModel) DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64) error {
	return m.deleteByUserIdAndRoleIdFn(ctx, userId, roleId)
}

// --- Mock: AuthRoleModel ---

type mockAuthRoleModel struct {
	findByIdsFn func(ctx context.Context, ids []int64) ([]*model.AuthRole, error)
	findOneFn   func(ctx context.Context, id int64) (*model.AuthRole, error)
}

func (m *mockAuthRoleModel) FindByIds(ctx context.Context, ids []int64) ([]*model.AuthRole, error) {
	return m.findByIdsFn(ctx, ids)
}
func (m *mockAuthRoleModel) FindOne(ctx context.Context, id int64) (*model.AuthRole, error) {
	return m.findOneFn(ctx, id)
}
func (m *mockAuthRoleModel) FindOneByCode(ctx context.Context, code string) (*model.AuthRole, error) {
	return nil, model.ErrNotFound
}
func (m *mockAuthRoleModel) Insert(ctx context.Context, data *model.AuthRole) (sql.Result, error) {
	return nil, nil
}
func (m *mockAuthRoleModel) Update(ctx context.Context, data *model.AuthRole) error { return nil }
func (m *mockAuthRoleModel) Delete(ctx context.Context, id int64) error                   { return nil }
func (m *mockAuthRoleModel) FindList(ctx context.Context, page, pageSize int32, status *int64) ([]*model.AuthRole, error) {
	return nil, nil
}
func (m *mockAuthRoleModel) Count(ctx context.Context, status *int64) (int64, error) { return 0, nil }

// --- Mock: AuthRolePermissionModel ---

type mockAuthRolePermissionModel struct {
	findByRoleIdFn func(ctx context.Context, roleId int64) ([]*model.AuthRolePermission, error)
}

func (m *mockAuthRolePermissionModel) FindByRoleId(ctx context.Context, roleId int64) ([]*model.AuthRolePermission, error) {
	return m.findByRoleIdFn(ctx, roleId)
}
func (m *mockAuthRolePermissionModel) FindOne(ctx context.Context, id int64) (*model.AuthRolePermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockAuthRolePermissionModel) FindOneByRoleIdPermissionId(ctx context.Context, roleId, permissionId int64) (*model.AuthRolePermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockAuthRolePermissionModel) Insert(ctx context.Context, data *model.AuthRolePermission) (sql.Result, error) {
	return nil, nil
}
func (m *mockAuthRolePermissionModel) Update(ctx context.Context, data *model.AuthRolePermission) error { return nil }
func (m *mockAuthRolePermissionModel) Delete(ctx context.Context, id int64) error                   { return nil }
func (m *mockAuthRolePermissionModel) DeleteByRoleId(ctx context.Context, roleId int64) error { return nil }
func (m *mockAuthRolePermissionModel) BatchInsert(ctx context.Context, roleId int64, permissionIds []int64) error {
	return nil
}

// --- Mock: AuthPermissionModel ---

type mockAuthPermissionModel struct {
	findByIdsFn func(ctx context.Context, ids []int64) ([]*model.AuthPermission, error)
	findAllFn   func(ctx context.Context, status *int64) ([]*model.AuthPermission, error)
}

func (m *mockAuthPermissionModel) FindByIds(ctx context.Context, ids []int64) ([]*model.AuthPermission, error) {
	return m.findByIdsFn(ctx, ids)
}
func (m *mockAuthPermissionModel) FindAll(ctx context.Context, status *int64) ([]*model.AuthPermission, error) {
	return m.findAllFn(ctx, status)
}
func (m *mockAuthPermissionModel) FindOne(ctx context.Context, id int64) (*model.AuthPermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockAuthPermissionModel) FindOneByCode(ctx context.Context, code string) (*model.AuthPermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockAuthPermissionModel) Insert(ctx context.Context, data *model.AuthPermission) (sql.Result, error) {
	return nil, nil
}
func (m *mockAuthPermissionModel) Update(ctx context.Context, data *model.AuthPermission) error { return nil }
func (m *mockAuthPermissionModel) Delete(ctx context.Context, id int64) error                   { return nil }

// --- Helper ---

func makeSvcCtx(ur model.AuthUserRoleModel, r model.AuthRoleModel, rp model.AuthRolePermissionModel, p model.AuthPermissionModel) *svc.ServiceContext {
	return &svc.ServiceContext{
		AuthUserRoleModel:      ur,
		AuthRoleModel:          r,
		AuthRolePermissionModel: rp,
		AuthPermissionModel:    p,
	}
}

// ========== GetUserPermissionsLogic ==========

func TestGetUserPermissions_WithRoles(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{
				{RoleId: 1, RoleCode: "community_manager"},
				{RoleId: 2, RoleCode: "reviewer"},
			}, nil
		},
	}
	rp := &mockAuthRolePermissionModel{
		findByRoleIdFn: func(_ context.Context, roleId int64) ([]*model.AuthRolePermission, error) {
			switch roleId {
			case 1:
				return []*model.AuthRolePermission{{PermissionId: 10}, {PermissionId: 11}}, nil
			case 2:
				return []*model.AuthRolePermission{{PermissionId: 20}, {PermissionId: 11}}, nil
			default:
				return nil, nil
			}
		},
	}
	pm := &mockAuthPermissionModel{
		findByIdsFn: func(_ context.Context, ids []int64) ([]*model.AuthPermission, error) {
			all := map[int64]*model.AuthPermission{
				10: {Id: 10, Code: "user:list", Type: 1, Status: 1},
				11: {Id: 11, Code: "user:create", Type: 2, Status: 1},
				20: {Id: 20, Code: "community:view", Type: 1, Status: 1},
				21: {Id: 21, Code: "community:edit", Type: 2, Status: 2},
			}
			out := make([]*model.AuthPermission, 0, len(ids))
			for _, id := range ids {
				if p, ok := all[id]; ok {
					out = append(out, p)
				}
			}
			return out, nil
		},
		findAllFn: func(_ context.Context, _ *int64) ([]*model.AuthPermission, error) {
			return []*model.AuthPermission{
				{Id: 10, Code: "user:list", Type: 1, Status: 1},
				{Id: 11, Code: "user:create", Type: 2, Status: 1},
				{Id: 20, Code: "community:view", Type: 1, Status: 1},
			}, nil
		},
	}

	l := NewGetUserPermissionsLogic(context.Background(), makeSvcCtx(ur, nil, rp, pm))
	resp, err := l.GetUserPermissions(&types.GetUserPermissionsReq{UserId: 5})

	assert.NoError(t, err)
	assert.ElementsMatch(t, resp.Permissions, []string{"user:list", "user:create", "community:view"})
	// All active permissions become tree nodes (3 active, 1 disabled excluded)
	assert.Len(t, resp.Menus, 3)
}

func TestGetUserPermissions_NoRoles(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, nil
		},
	}
	l := NewGetUserPermissionsLogic(context.Background(), makeSvcCtx(ur, nil, nil, nil))
	resp, err := l.GetUserPermissions(&types.GetUserPermissionsReq{UserId: 5})

	assert.NoError(t, err)
	assert.Empty(t, resp.Permissions)
	assert.Empty(t, resp.Menus)
}

func TestGetUserPermissions_ExcludesDisabled(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 1}}, nil
		},
	}
	rp := &mockAuthRolePermissionModel{
		findByRoleIdFn: func(_ context.Context, _ int64) ([]*model.AuthRolePermission, error) {
			return []*model.AuthRolePermission{{PermissionId: 99}}, nil
		},
	}
	pm := &mockAuthPermissionModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthPermission, error) {
			return []*model.AuthPermission{{Id: 99, Code: "disabled:perm", Status: 2}}, nil
		},
		findAllFn: func(_ context.Context, _ *int64) ([]*model.AuthPermission, error) { return nil, nil },
	}

	l := NewGetUserPermissionsLogic(context.Background(), makeSvcCtx(ur, nil, rp, pm))
	resp, err := l.GetUserPermissions(&types.GetUserPermissionsReq{UserId: 5})

	assert.NoError(t, err)
	assert.Empty(t, resp.Permissions)
}

func TestGetUserPermissions_DBError(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, sql.ErrConnDone
		},
	}
	l := NewGetUserPermissionsLogic(context.Background(), makeSvcCtx(ur, nil, nil, nil))
	_, err := l.GetUserPermissions(&types.GetUserPermissionsReq{UserId: 5})
	assert.Error(t, err)
}

// ========== GetUserRolesLogic ==========

func TestGetUserRoles_Success(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{
				{RoleId: 1, RoleName: "超级管理员", RoleCode: "super_admin", IsSystem: 1, RoleStatus: 1, Description: "系统角色"},
				{RoleId: 2, RoleName: "社区经理", RoleCode: "community_mgr", IsSystem: 0, RoleStatus: 1, Description: "自定义角色"},
			}, nil
		},
	}
	l := NewGetUserRolesLogic(context.Background(), makeSvcCtx(ur, nil, nil, nil))
	resp, err := l.GetUserRoles(&types.GetUserRolesReq{Id: 5})

	assert.NoError(t, err)
	assert.Len(t, resp.Roles, 2)
	assert.Equal(t, int32(1), resp.Roles[0].IsSystem)
	assert.Equal(t, int32(0), resp.Roles[1].IsSystem)
}

func TestGetUserRoles_Empty(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) { return nil, nil },
	}
	l := NewGetUserRolesLogic(context.Background(), makeSvcCtx(ur, nil, nil, nil))
	resp, err := l.GetUserRoles(&types.GetUserRolesReq{Id: 99})

	assert.NoError(t, err)
	assert.Empty(t, resp.Roles)
}

// ========== AssignUserRolesLogic ==========

func TestAssignUserRoles_Success(t *testing.T) {
	rm := &mockAuthRoleModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthRole, error) {
			return []*model.AuthRole{{Id: 2, Status: 1}, {Id: 3, Status: 1}}, nil
		},
	}
	ur := &mockAuthUserRoleModel{
		batchInsertUserRolesFn: func(ctx context.Context, userId int64, roleIds []int64) error {
			assert.Equal(t, int64(5), userId)
			assert.Equal(t, []int64{2, 3}, roleIds)
			return nil
		},
	}
	l := NewAssignUserRolesLogic(context.Background(), makeSvcCtx(ur, rm, nil, nil))
	resp, err := l.AssignUserRoles(&types.AssignUserRolesReq{Id: 5, RoleIds: []int64{2, 3}})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestAssignUserRoles_EmptyRoleIds(t *testing.T) {
	l := NewAssignUserRolesLogic(context.Background(), makeSvcCtx(nil, nil, nil, nil))
	resp, err := l.AssignUserRoles(&types.AssignUserRolesReq{Id: 5, RoleIds: []int64{}})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestAssignUserRoles_NonExistentRole(t *testing.T) {
	rm := &mockAuthRoleModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthRole, error) {
			return []*model.AuthRole{{Id: 2, Status: 1}}, nil
		},
	}
	l := NewAssignUserRolesLogic(context.Background(), makeSvcCtx(nil, rm, nil, nil))
	_, err := l.AssignUserRoles(&types.AssignUserRolesReq{Id: 5, RoleIds: []int64{2, 999}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do not exist")
}

func TestAssignUserRoles_InactiveRole(t *testing.T) {
	rm := &mockAuthRoleModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthRole, error) {
			return []*model.AuthRole{{Id: 2, Code: "cm", Status: 1}, {Id: 3, Code: "off", Status: 2}}, nil
		},
	}
	l := NewAssignUserRolesLogic(context.Background(), makeSvcCtx(nil, rm, nil, nil))
	_, err := l.AssignUserRoles(&types.AssignUserRolesReq{Id: 5, RoleIds: []int64{2, 3}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestAssignUserRoles_DBInsertError(t *testing.T) {
	rm := &mockAuthRoleModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthRole, error) {
			return []*model.AuthRole{{Id: 2, Status: 1}}, nil
		},
	}
	ur := &mockAuthUserRoleModel{
		batchInsertUserRolesFn: func(_ context.Context, _ int64, _ []int64) error {
			return sql.ErrConnDone
		},
	}
	l := NewAssignUserRolesLogic(context.Background(), makeSvcCtx(ur, rm, nil, nil))
	_, err := l.AssignUserRoles(&types.AssignUserRolesReq{Id: 5, RoleIds: []int64{2}})

	assert.Error(t, err)
}

// ========== RemoveUserRoleLogic ==========

func TestRemoveUserRole_Success(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findOneByUserIdRoleIdFn: func(_ context.Context, _ int64, _ int64) (*model.AuthUserRole, error) {
			return &model.AuthUserRole{Id: 100, UserId: 5, RoleId: 2}, nil
		},
		deleteByUserIdAndRoleIdFn: func(_ context.Context, _ int64, _ int64) error {
			return nil
		},
	}
	rm := &mockAuthRoleModel{
		findOneFn: func(_ context.Context, id int64) (*model.AuthRole, error) {
			return &model.AuthRole{Id: 2, Code: "community_mgr", IsSystem: 0}, nil
		},
	}
	l := NewRemoveUserRoleLogic(context.Background(), makeSvcCtx(ur, rm, nil, nil))
	resp, err := l.RemoveUserRole(&types.RemoveUserRoleReq{Id: 5, RoleId: 2})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestRemoveUserRole_NotAssigned(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findOneByUserIdRoleIdFn: func(_ context.Context, _ int64, _ int64) (*model.AuthUserRole, error) {
			return nil, model.ErrNotFound
		},
	}
	l := NewRemoveUserRoleLogic(context.Background(), makeSvcCtx(ur, nil, nil, nil))
	_, err := l.RemoveUserRole(&types.RemoveUserRoleReq{Id: 5, RoleId: 99})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not have role")
}

func TestRemoveUserRole_SystemRole(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findOneByUserIdRoleIdFn: func(_ context.Context, _ int64, _ int64) (*model.AuthUserRole, error) {
			return &model.AuthUserRole{Id: 100, UserId: 5, RoleId: 1}, nil
		},
	}
	rm := &mockAuthRoleModel{
		findOneFn: func(_ context.Context, id int64) (*model.AuthRole, error) {
			return &model.AuthRole{Id: 1, Code: "super_admin", IsSystem: 1}, nil
		},
	}
	l := NewRemoveUserRoleLogic(context.Background(), makeSvcCtx(ur, rm, nil, nil))
	_, err := l.RemoveUserRole(&types.RemoveUserRoleReq{Id: 5, RoleId: 1})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove system role")
}

func TestRemoveUserRole_DeleteError(t *testing.T) {
	ur := &mockAuthUserRoleModel{
		findOneByUserIdRoleIdFn: func(_ context.Context, _ int64, _ int64) (*model.AuthUserRole, error) {
			return &model.AuthUserRole{Id: 100}, nil
		},
		deleteByUserIdAndRoleIdFn: func(_ context.Context, _ int64, _ int64) error {
			return sql.ErrConnDone
		},
	}
	rm := &mockAuthRoleModel{
		findOneFn: func(_ context.Context, id int64) (*model.AuthRole, error) {
			return &model.AuthRole{Id: 2, Code: "custom", IsSystem: 0}, nil
		},
	}
	l := NewRemoveUserRoleLogic(context.Background(), makeSvcCtx(ur, rm, nil, nil))
	_, err := l.RemoveUserRole(&types.RemoveUserRoleReq{Id: 5, RoleId: 2})

	assert.Error(t, err)
}
