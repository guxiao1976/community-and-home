package logic

import (
	"context"
	"database/sql"
	"testing"

	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockRpcUserRoleModel struct {
	findActiveByUserIdFn func(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error)
}

func (m *mockRpcUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	return m.findActiveByUserIdFn(ctx, userId)
}
func (m *mockRpcUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*model.AuthUserRole, error) {
	return nil, nil
}
func (m *mockRpcUserRoleModel) FindOne(ctx context.Context, id int64) (*model.AuthUserRole, error) {
	return nil, model.ErrNotFound
}
func (m *mockRpcUserRoleModel) FindOneByUserIdRoleId(ctx context.Context, userId, roleId int64) (*model.AuthUserRole, error) {
	return nil, model.ErrNotFound
}
func (m *mockRpcUserRoleModel) Insert(ctx context.Context, data *model.AuthUserRole) (sql.Result, error) {
	return nil, nil
}
func (m *mockRpcUserRoleModel) Update(ctx context.Context, data *model.AuthUserRole) error { return nil }
func (m *mockRpcUserRoleModel) Delete(ctx context.Context, id int64) error                   { return nil }
func (m *mockRpcUserRoleModel) BatchInsertUserRoles(ctx context.Context, userId int64, roleIds []int64) error {
	return nil
}
func (m *mockRpcUserRoleModel) DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64) error {
	return nil
}

type mockRpcRolePermModel struct {
	findByRoleIdFn func(ctx context.Context, roleId int64) ([]*model.AuthRolePermission, error)
}

func (m *mockRpcRolePermModel) FindByRoleId(ctx context.Context, roleId int64) ([]*model.AuthRolePermission, error) {
	return m.findByRoleIdFn(ctx, roleId)
}
func (m *mockRpcRolePermModel) FindOne(ctx context.Context, id int64) (*model.AuthRolePermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockRpcRolePermModel) FindOneByRoleIdPermissionId(ctx context.Context, roleId, permissionId int64) (*model.AuthRolePermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockRpcRolePermModel) Insert(ctx context.Context, data *model.AuthRolePermission) (sql.Result, error) {
	return nil, nil
}
func (m *mockRpcRolePermModel) Update(ctx context.Context, data *model.AuthRolePermission) error { return nil }
func (m *mockRpcRolePermModel) Delete(ctx context.Context, id int64) error                   { return nil }
func (m *mockRpcRolePermModel) DeleteByRoleId(ctx context.Context, roleId int64) error       { return nil }
func (m *mockRpcRolePermModel) BatchInsert(ctx context.Context, roleId int64, permissionIds []int64) error {
	return nil
}

type mockRpcPermModel struct {
	findByIdsFn func(ctx context.Context, ids []int64) ([]*model.AuthPermission, error)
}

func (m *mockRpcPermModel) FindByIds(ctx context.Context, ids []int64) ([]*model.AuthPermission, error) {
	return m.findByIdsFn(ctx, ids)
}
func (m *mockRpcPermModel) FindAll(ctx context.Context, status *int64) ([]*model.AuthPermission, error) {
	return nil, nil
}
func (m *mockRpcPermModel) FindOne(ctx context.Context, id int64) (*model.AuthPermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockRpcPermModel) FindOneByCode(ctx context.Context, code string) (*model.AuthPermission, error) {
	return nil, model.ErrNotFound
}
func (m *mockRpcPermModel) Insert(ctx context.Context, data *model.AuthPermission) (sql.Result, error) {
	return nil, nil
}
func (m *mockRpcPermModel) Update(ctx context.Context, data *model.AuthPermission) error { return nil }
func (m *mockRpcPermModel) Delete(ctx context.Context, id int64) error                   { return nil }

func makeRpcSvcCtx(ur model.AuthUserRoleModel, rp model.AuthRolePermissionModel, p model.AuthPermissionModel) *svc.ServiceContext {
	return &svc.ServiceContext{
		AuthUserRoleModel:       ur,
		AuthRolePermissionModel: rp,
		AuthPermissionModel:     p,
	}
}

// ========== CheckPermissionLogic ==========

func TestCheckPermission_SystemRoleBypass(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 1, IsSystem: 1}}, nil
		},
	}
	l := NewCheckPermissionLogic(context.Background(), makeRpcSvcCtx(ur, nil, nil))
	resp, err := l.CheckPermission(&pb.CheckPermissionReq{UserId: 1, Action: "delete", Resource: "user"})

	assert.NoError(t, err)
	assert.True(t, resp.Allowed)
}

func TestCheckPermission_MatchingPermission(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 2, IsSystem: 0}}, nil
		},
	}
	rp := &mockRpcRolePermModel{
		findByRoleIdFn: func(_ context.Context, _ int64) ([]*model.AuthRolePermission, error) {
			return []*model.AuthRolePermission{{PermissionId: 10}}, nil
		},
	}
	pm := &mockRpcPermModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthPermission, error) {
			return []*model.AuthPermission{{Id: 10, Code: "user:list", Status: 1}}, nil
		},
	}
	l := NewCheckPermissionLogic(context.Background(), makeRpcSvcCtx(ur, rp, pm))
	resp, err := l.CheckPermission(&pb.CheckPermissionReq{UserId: 1, Action: "user", Resource: "list"})

	assert.NoError(t, err)
	assert.True(t, resp.Allowed)
}

func TestCheckPermission_NoMatchingPermission(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 2, IsSystem: 0}}, nil
		},
	}
	rp := &mockRpcRolePermModel{
		findByRoleIdFn: func(_ context.Context, _ int64) ([]*model.AuthRolePermission, error) {
			return []*model.AuthRolePermission{{PermissionId: 10}}, nil
		},
	}
	pm := &mockRpcPermModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthPermission, error) {
			return []*model.AuthPermission{{Id: 10, Code: "user:list", Status: 1}}, nil
		},
	}
	l := NewCheckPermissionLogic(context.Background(), makeRpcSvcCtx(ur, rp, pm))
	resp, err := l.CheckPermission(&pb.CheckPermissionReq{UserId: 1, Action: "user", Resource: "delete"})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckPermission_NoRoles(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, nil
		},
	}
	l := NewCheckPermissionLogic(context.Background(), makeRpcSvcCtx(ur, nil, nil))
	resp, err := l.CheckPermission(&pb.CheckPermissionReq{UserId: 1, Action: "user", Resource: "list"})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckPermission_DBError(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, sql.ErrConnDone
		},
	}
	l := NewCheckPermissionLogic(context.Background(), makeRpcSvcCtx(ur, nil, nil))
	resp, err := l.CheckPermission(&pb.CheckPermissionReq{UserId: 1, Action: "user", Resource: "list"})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckPermission_DisabledPermissionExcluded(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 2, IsSystem: 0}}, nil
		},
	}
	rp := &mockRpcRolePermModel{
		findByRoleIdFn: func(_ context.Context, _ int64) ([]*model.AuthRolePermission, error) {
			return []*model.AuthRolePermission{{PermissionId: 10}}, nil
		},
	}
	pm := &mockRpcPermModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthPermission, error) {
			return []*model.AuthPermission{{Id: 10, Code: "user:list", Status: 2}}, nil
		},
	}
	l := NewCheckPermissionLogic(context.Background(), makeRpcSvcCtx(ur, rp, pm))
	resp, err := l.CheckPermission(&pb.CheckPermissionReq{UserId: 1, Action: "user", Resource: "list"})

	assert.NoError(t, err)
	assert.False(t, resp.Allowed)
}

// ========== GetUserPermissionsLogic ==========

func TestGetUserPermissions_Success(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{
				{RoleId: 1, RoleCode: "admin"},
				{RoleId: 2, RoleCode: "viewer"},
			}, nil
		},
	}
	rp := &mockRpcRolePermModel{
		findByRoleIdFn: func(_ context.Context, roleId int64) ([]*model.AuthRolePermission, error) {
			switch roleId {
			case 1:
				return []*model.AuthRolePermission{{PermissionId: 10}, {PermissionId: 11}}, nil
			case 2:
				return []*model.AuthRolePermission{{PermissionId: 20}}, nil
			default:
				return nil, nil
			}
		},
	}
	pm := &mockRpcPermModel{
		findByIdsFn: func(_ context.Context, ids []int64) ([]*model.AuthPermission, error) {
			all := map[int64]*model.AuthPermission{
				10: {Id: 10, Code: "user:list", Status: 1},
				11: {Id: 11, Code: "user:create", Status: 1},
				20: {Id: 20, Code: "community:view", Status: 1},
			}
			out := make([]*model.AuthPermission, 0, len(ids))
			for _, id := range ids {
				if p, ok := all[id]; ok {
					out = append(out, p)
				}
			}
			return out, nil
		},
	}
	l := NewGetUserPermissionsLogic(context.Background(), makeRpcSvcCtx(ur, rp, pm))
	resp, err := l.GetUserPermissions(&pb.GetUserPermissionsReq{UserId: 5})

	assert.NoError(t, err)
	assert.ElementsMatch(t, resp.Permissions, []string{"user:list", "user:create", "community:view"})
}

func TestGetUserPermissions_NoRoles(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, nil
		},
	}
	l := NewGetUserPermissionsLogic(context.Background(), makeRpcSvcCtx(ur, nil, nil))
	resp, err := l.GetUserPermissions(&pb.GetUserPermissionsReq{UserId: 5})

	assert.NoError(t, err)
	assert.Empty(t, resp.Permissions)
}

func TestGetUserPermissions_DBError(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, sql.ErrConnDone
		},
	}
	l := NewGetUserPermissionsLogic(context.Background(), makeRpcSvcCtx(ur, nil, nil))
	_, err := l.GetUserPermissions(&pb.GetUserPermissionsReq{UserId: 5})

	assert.Error(t, err)
}

func TestGetUserPermissions_FiltersDisabled(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 1}}, nil
		},
	}
	rp := &mockRpcRolePermModel{
		findByRoleIdFn: func(_ context.Context, _ int64) ([]*model.AuthRolePermission, error) {
			return []*model.AuthRolePermission{{PermissionId: 10}, {PermissionId: 11}}, nil
		},
	}
	pm := &mockRpcPermModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthPermission, error) {
			return []*model.AuthPermission{
				{Id: 10, Code: "active:perm", Status: 1},
				{Id: 11, Code: "disabled:perm", Status: 2},
			}, nil
		},
	}
	l := NewGetUserPermissionsLogic(context.Background(), makeRpcSvcCtx(ur, rp, pm))
	resp, err := l.GetUserPermissions(&pb.GetUserPermissionsReq{UserId: 5})

	assert.NoError(t, err)
	assert.Equal(t, []string{"active:perm"}, resp.Permissions)
}

func TestGetUserPermissions_RolePermQueryErrorSkipped(t *testing.T) {
	ur := &mockRpcUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 1}, {RoleId: 2}}, nil
		},
	}
	rp := &mockRpcRolePermModel{
		findByRoleIdFn: func(_ context.Context, roleId int64) ([]*model.AuthRolePermission, error) {
			if roleId == 1 {
				return nil, sql.ErrConnDone
			}
			return []*model.AuthRolePermission{{PermissionId: 20}}, nil
		},
	}
	pm := &mockRpcPermModel{
		findByIdsFn: func(_ context.Context, _ []int64) ([]*model.AuthPermission, error) {
			return []*model.AuthPermission{{Id: 20, Code: "user:list", Status: 1}}, nil
		},
	}
	l := NewGetUserPermissionsLogic(context.Background(), makeRpcSvcCtx(ur, rp, pm))
	resp, err := l.GetUserPermissions(&pb.GetUserPermissionsReq{UserId: 5})

	assert.NoError(t, err)
	assert.Equal(t, []string{"user:list"}, resp.Permissions)
}
