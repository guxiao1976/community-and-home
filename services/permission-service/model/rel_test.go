package model

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// SEE: [[testing-discipline]] — 关系表测试：RelUserRole 唯一约束、RelRolePermission 级联删除、幂等性

func TestRelUserRoleModel_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	mock.ExpectExec("insert into `rel_user_role`").
		WithArgs(int64(100), int64(1), "community", int64(5001)).
		WillReturnResult(sqlmock.NewResult(10, 1))

	rel := &RelUserRole{
		UserId:    100,
		RoleId:    1,
		ScopeType: "community",
		ScopeId:   5001,
	}

	id, err := m.Insert(context.Background(), rel)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 10 {
		t.Errorf("expected id 10, got %d", id)
	}
}

func TestRelUserRoleModel_BatchInsertUserRoles_Idempotent(t *testing.T) {
	// SEE: [[testing-discipline]] — BatchInsert 使用 INSERT IGNORE 实现幂等性
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	// 第一次插入成功
	mock.ExpectExec("insert ignore into `rel_user_role`").
		WithArgs(int64(100), int64(1), "community", int64(5001)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 第二次插入（重复）返回 0 rows affected，但不报错
	mock.ExpectExec("insert ignore into `rel_user_role`").
		WithArgs(int64(100), int64(2), "building", int64(6001)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	records := []*RelUserRole{
		{UserId: 100, RoleId: 1, ScopeType: "community", ScopeId: 5001},
		{UserId: 100, RoleId: 2, ScopeType: "building", ScopeId: 6001},
	}

	err = m.BatchInsertUserRoles(context.Background(), records)

	if err != nil {
		t.Errorf("expected nil error for idempotent insert, got %v", err)
	}
}

func TestRelUserRoleModel_FindActiveByUserId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	rows := sqlmock.NewRows([]string{"role_id", "role_code", "role_name", "is_system", "role_status", "description", "scope_type", "scope_id"}).
		AddRow(1, "owner", "业主", 0, 1, "业主角色", "community", 5001).
		AddRow(2, "property_admin", "物业管理员", 1, 1, "系统角色", "community", 5001)

	mock.ExpectQuery("SELECT ur.role_id, r.role_code, r.role_name, r.is_system, r.status as role_status, r.description").
		WithArgs(int64(100)).
		WillReturnRows(rows)

	result, err := m.FindActiveByUserId(context.Background(), 100)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 roles, got %d", len(result))
	}
	if result[0].RoleCode != "owner" {
		t.Errorf("expected first role 'owner', got '%s'", result[0].RoleCode)
	}
	if result[1].IsSystem != 1 {
		t.Errorf("expected second role is_system=1, got %d", result[1].IsSystem)
	}
}

func TestRelUserRoleModel_FindScopesByUserId(t *testing.T) {
	tests := []struct {
		name      string
		userId    int64
		scopeType string
		wantIds   []int64
	}{
		{"社区范围", 100, "community", []int64{5001, 5002}},
		{"楼栋范围", 100, "building", []int64{6001}},
		{"无结果", 999, "community", []int64{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			conn := sqlx.NewSqlConnFromDB(db)
			m := NewRelUserRoleModel(conn, nil)

			rows := sqlmock.NewRows([]string{"scope_id"})
			for _, id := range tt.wantIds {
				rows.AddRow(id)
			}

			mock.ExpectQuery("SELECT DISTINCT ur.scope_id FROM `rel_user_role` ur").
				WithArgs(tt.userId, tt.scopeType).
				WillReturnRows(rows)

			result, err := m.FindScopesByUserId(context.Background(), tt.userId, tt.scopeType)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(result) != len(tt.wantIds) {
				t.Errorf("expected %d scope_ids, got %d", len(tt.wantIds), len(result))
			}
		})
	}
}

func TestRelUserRoleModel_DeleteByUserIdAndRoleId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	mock.ExpectExec("delete from `rel_user_role` where user_id = \\? and role_id = \\? and scope_type = \\? and scope_id = \\?").
		WithArgs(int64(100), int64(1), "community", int64(5001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.DeleteByUserIdAndRoleId(context.Background(), 100, 1, "community", 5001)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRelUserRoleModel_CountByRoleId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	mock.ExpectQuery("select count\\(\\*\\) from `rel_user_role` where role_id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))

	count, err := m.CountByRoleId(context.Background(), 1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if count != 25 {
		t.Errorf("expected count 25, got %d", count)
	}
}

func TestRelRolePermissionModel_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelRolePermissionModel(conn, nil)

	mock.ExpectExec("insert into `rel_role_permission`").
		WithArgs(int64(1), int64(101)).
		WillReturnResult(sqlmock.NewResult(20, 1))

	rel := &RelRolePermission{
		RoleId:       1,
		PermissionId: 101,
	}

	id, err := m.Insert(context.Background(), rel)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 20 {
		t.Errorf("expected id 20, got %d", id)
	}
}

func TestRelRolePermissionModel_FindByRoleId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelRolePermissionModel(conn, nil)

	rows := sqlmock.NewRows([]string{"id", "role_id", "permission_id", "created_time"}).
		AddRow(1, 1, 101, time.Now()).
		AddRow(2, 1, 102, time.Now()).
		AddRow(3, 1, 103, time.Now())

	mock.ExpectQuery("select \\* from `rel_role_permission` where role_id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := m.FindByRoleId(context.Background(), 1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 permissions, got %d", len(result))
	}
}

func TestRelRolePermissionModel_DeleteByRoleId_CascadeDelete(t *testing.T) {
	// SEE: [[testing-discipline]] — 测试级联删除：删除角色时清空其权限关联
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelRolePermissionModel(conn, nil)

	mock.ExpectExec("delete from `rel_role_permission` where role_id = \\?").
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	err = m.DeleteByRoleId(context.Background(), 1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRelRolePermissionModel_BatchInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelRolePermissionModel(conn, nil)

	mock.ExpectExec("insert into `rel_role_permission`").
		WithArgs(int64(1), int64(101)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("insert into `rel_role_permission`").
		WithArgs(int64(1), int64(102)).
		WillReturnResult(sqlmock.NewResult(2, 1))

	records := []*RelRolePermission{
		{RoleId: 1, PermissionId: 101},
		{RoleId: 1, PermissionId: 102},
	}

	err = m.BatchInsert(context.Background(), records)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
