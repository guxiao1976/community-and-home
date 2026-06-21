package model

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// SEE: [[testing-discipline]] — Model 层边界测试：FindByIds 空列表、FindList 分页边界、SysRole CRUD

func TestSysRoleModel_FindByIds_EmptyList(t *testing.T) {
	// RED: 验证空列表直接返回 nil
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysRoleModel(conn, nil)

	result, err := m.FindByIds(context.Background(), []int64{})

	if err != nil {
		t.Errorf("expected nil error for empty ids, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty ids, got %v", result)
	}
}

func TestSysRoleModel_FindByIds_SingleId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysRoleModel(conn, nil)

	rows := sqlmock.NewRows([]string{"id", "role_code", "role_name", "description", "is_system", "sort_order", "status", "created_by", "created_time", "updated_time", "delete_time"}).
		AddRow(1, "owner", "业主", sql.NullString{String: "业主角色", Valid: true}, 0, 1, 1, 100, time.Now(), time.Now(), sql.NullTime{})

	mock.ExpectQuery("select \\* from `sys_role` where id in \\(\\?\\) and delete_time is null").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := m.FindByIds(context.Background(), []int64{1})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 role, got %d", len(result))
	}
	if result[0].RoleCode != "owner" {
		t.Errorf("expected role_code 'owner', got '%s'", result[0].RoleCode)
	}
}

func TestSysRoleModel_FindByIds_MultipleIds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysRoleModel(conn, nil)

	rows := sqlmock.NewRows([]string{"id", "role_code", "role_name", "description", "is_system", "sort_order", "status", "created_by", "created_time", "updated_time", "delete_time"}).
		AddRow(1, "owner", "业主", sql.NullString{}, 0, 1, 1, 100, time.Now(), time.Now(), sql.NullTime{}).
		AddRow(2, "property_admin", "物业管理员", sql.NullString{}, 0, 2, 1, 100, time.Now(), time.Now(), sql.NullTime{})

	mock.ExpectQuery("select \\* from `sys_role` where id in \\(\\?,\\?\\) and delete_time is null").
		WithArgs(int64(1), int64(2)).
		WillReturnRows(rows)

	result, err := m.FindByIds(context.Background(), []int64{1, 2})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 roles, got %d", len(result))
	}
}

func TestSysRoleModel_FindList_Pagination(t *testing.T) {
	tests := []struct {
		name     string
		page     int64
		pageSize int64
		total    int64
		wantRows int
	}{
		{"第1页_10条", 1, 10, 25, 10},
		{"第2页_10条", 2, 10, 25, 10},
		{"第3页_10条", 3, 10, 25, 5},
		{"边界_第1页_0条", 1, 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			conn := sqlx.NewSqlConnFromDB(db)
			m := NewSysRoleModel(conn, nil)

			// Mock count query
			mock.ExpectQuery("select count\\(\\*\\) from `sys_role` where delete_time is null").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.total))

			// Mock data query
			rows := sqlmock.NewRows([]string{"id", "role_code", "role_name", "description", "is_system", "sort_order", "status", "created_by", "created_time", "updated_time", "delete_time"})
			for i := 0; i < tt.wantRows; i++ {
				rows.AddRow(int64(i+1), "code", "name", sql.NullString{}, 0, 1, 1, 100, time.Now(), time.Now(), sql.NullTime{})
			}
			mock.ExpectQuery("select \\* from `sys_role` where delete_time is null order by sort_order asc limit .* offset .*").
				WillReturnRows(rows)

			result, total, err := m.FindList(context.Background(), nil, tt.page, tt.pageSize)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if total != tt.total {
				t.Errorf("expected total %d, got %d", tt.total, total)
			}
			if len(result) != tt.wantRows {
				t.Errorf("expected %d rows, got %d", tt.wantRows, len(result))
			}
		})
	}
}

func TestSysRoleModel_FindList_WithStatusFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysRoleModel(conn, nil)

	status := int64(1)

	mock.ExpectQuery("select count\\(\\*\\) from `sys_role` where delete_time is null and status = \\?").
		WithArgs(status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	rows := sqlmock.NewRows([]string{"id", "role_code", "role_name", "description", "is_system", "sort_order", "status", "created_by", "created_time", "updated_time", "delete_time"}).
		AddRow(1, "owner", "业主", sql.NullString{}, 0, 1, 1, 100, time.Now(), time.Now(), sql.NullTime{})

	mock.ExpectQuery("select \\* from `sys_role` where delete_time is null and status = \\? order by sort_order asc limit .* offset .*").
		WithArgs(status).
		WillReturnRows(rows)

	result, total, err := m.FindList(context.Background(), &status, 1, 10)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 role, got %d", len(result))
	}
	if result[0].Status != 1 {
		t.Errorf("expected status 1, got %d", result[0].Status)
	}
}

func TestSysRoleModel_Insert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysRoleModel(conn, nil)

	mock.ExpectExec("insert into `sys_role`").
		WithArgs("test_code", "测试角色", sql.NullString{String: "描述", Valid: true}, int64(0), int64(10), int64(1), int64(100)).
		WillReturnResult(sqlmock.NewResult(123, 1))

	role := &SysRole{
		RoleCode:    "test_code",
		RoleName:    "测试角色",
		Description: sql.NullString{String: "描述", Valid: true},
		IsSystem:    0,
		SortOrder:   10,
		Status:      1,
		CreatedBy:   100,
	}

	id, err := m.Insert(context.Background(), role)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id != 123 {
		t.Errorf("expected id 123, got %d", id)
	}
}

func TestSysRoleModel_SoftDelete_SystemRoleProtected(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysRoleModel(conn, nil)

	// SoftDelete 的 SQL 包含 is_system = 0 条件，系统角色不会被删除
	mock.ExpectExec("update `sys_role` set delete_time = now\\(\\) where id = \\? and is_system = 0").
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = m.SoftDelete(context.Background(), 1)

	if err != nil {
		t.Errorf("expected nil error (0 rows affected is ok), got %v", err)
	}
}

func TestSysPermissionModel_FindByIds_EmptyList(t *testing.T) {
	// SEE: [[testing-discipline]] — Permission 空列表边界测试
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysPermissionModel(conn, nil)

	result, err := m.FindByIds(context.Background(), []int64{})

	if err != nil {
		t.Errorf("expected nil error for empty ids, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty ids, got %v", result)
	}
}

func TestSysPermissionModel_FindWithFilter_NoFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysPermissionModel(conn, nil)

	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "code", "type", "path", "icon", "sort_order", "status", "created_time", "updated_time"}).
		AddRow(1, sql.NullInt64{}, "用户管理", "user:read", 1, sql.NullString{String: "/api/user", Valid: true}, sql.NullString{}, 1, 1, time.Now(), time.Now())

	mock.ExpectQuery("select \\* from `sys_permission` where 1=1 order by sort_order asc").
		WillReturnRows(rows)

	result, err := m.FindWithFilter(context.Background(), nil, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 permission, got %d", len(result))
	}
}

func TestSysPermissionModel_FindWithFilter_TypeFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewSysPermissionModel(conn, nil)

	typeFilter := int64(3) // API 类型

	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "code", "type", "path", "icon", "sort_order", "status", "created_time", "updated_time"}).
		AddRow(1, sql.NullInt64{}, "API权限", "api:test", 3, sql.NullString{String: "/api/test", Valid: true}, sql.NullString{}, 1, 1, time.Now(), time.Now())

	mock.ExpectQuery("select \\* from `sys_permission` where 1=1 and type = \\? order by sort_order asc").
		WithArgs(typeFilter).
		WillReturnRows(rows)

	result, err := m.FindWithFilter(context.Background(), &typeFilter, nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 permission, got %d", len(result))
	}
	if result[0].Type != 3 {
		t.Errorf("expected type 3, got %d", result[0].Type)
	}
}
