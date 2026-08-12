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
		WithArgs(int64(100), int64(1), "community", int64(5001), int64(0), sqlmock.AnyArg(), sqlmock.AnyArg()).
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
		WithArgs(int64(100), int64(1), "community", int64(5001), int64(0), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 第二次插入（重复）返回 0 rows affected，但不报错
	mock.ExpectExec("insert ignore into `rel_user_role`").
		WithArgs(int64(100), int64(2), "building", int64(6001), int64(0), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	rows := sqlmock.NewRows([]string{"role_id", "role_code", "role_name", "is_system", "role_status", "description", "scope_type", "scope_id", "ur_status", "verified_at", "expires_at"}).
		AddRow(1, "owner", "业主", 0, 1, "业主角色", "community", 5001, 2, nil, nil).
		AddRow(2, "property_admin", "物业管理员", 1, 1, "系统角色", "community", 5001, 2, nil, nil)

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

// TestRelUserRoleModel_FindScopesByUserId_FiltersInactiveAndZeroScope
// T1.2: FindScopesByUserId 必须过滤 status IN (0,1,2) 且 scope_id != 0（空/全局占位不进 limited 并集）
func TestRelUserRoleModel_FindScopesByUserId_FiltersInactiveAndZeroScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	rows := sqlmock.NewRows([]string{"scope_id"}).AddRow(5001).AddRow(5002)
	mock.ExpectQuery("SELECT DISTINCT ur.scope_id FROM `rel_user_role` ur").
		WithArgs(int64(100), "community").
		WillReturnRows(rows)

	result, err := m.FindScopesByUserId(context.Background(), 100, "community")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 scope_ids, got %d", len(result))
	}
	// 验证 SQL 含状态过滤与 scope_id!=0（通过 ExpectQuery 匹配新 query 字符串）
}

// TestRelUserRoleModel_FindScopesByUserId_ZeroScopeExcluded
// T1.2: scope_id=0 行（empty/global 占位）不进结果 —— sqlmock 断言 query 含 scope_id != 0
func TestRelUserRoleModel_FindScopesByUserId_ZeroScopeExcluded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	mock.ExpectQuery("ur.scope_id != 0").
		WithArgs(int64(100), "community").
		WillReturnRows(sqlmock.NewRows([]string{"scope_id"}).AddRow(5001))

	_, err = m.FindScopesByUserId(context.Background(), 100, "community")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRelUserRoleModel_FindActiveRolesByUserId
// T1.2: FindActiveRolesByUserId 返回 status IN (0,1,2) 且未过期的 grants（含 scope/verified_at/ur_status）
func TestRelUserRoleModel_FindActiveRolesByUserId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	rows := sqlmock.NewRows([]string{"role_id", "role_code", "role_name", "is_system", "role_status", "description", "scope_type", "scope_id", "ur_status", "verified_at", "expires_at"}).
		AddRow(1, "owner", "业主", 0, 1, "业主角色", "community", 5001, 0, nil, nil).
		AddRow(9, "registered_user", "注册用户", 1, 1, "基角色", "", 0, 2, nil, nil)

	// 断言 query 含 status IN (0,1,2) 与过期过滤
	mock.ExpectQuery("ur.status IN \\(0,1,2\\)").
		WithArgs(int64(100)).
		WillReturnRows(rows)

	result, err := m.FindActiveRolesByUserId(context.Background(), 100)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 grants, got %d", len(result))
	}
	if result[0].ScopeType != "community" || result[0].ScopeId != 5001 {
		t.Errorf("expected community:5001, got %s:%d", result[0].ScopeType, result[0].ScopeId)
	}
	if result[0].URStatus != 0 {
		t.Errorf("expected ur_status=0, got %d", result[0].URStatus)
	}
	if result[1].URStatus != 2 || result[1].VerifiedAt.Valid {
		t.Errorf("registered_user should be status=2 with verified_at NULL, got status=%d verified=%v", result[1].URStatus, result[1].VerifiedAt.Valid)
	}
}

// TestRelUserRoleModel_FindActiveRolesByUserId_ExpiredExcluded
// T1.2: 已过期（expires_at <= NOW()）的 grant 被 SQL 排除 —— 断言 query 含过期过滤
func TestRelUserRoleModel_FindActiveRolesByUserId_ExpiredExcluded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	mock.ExpectQuery("expires_at IS NULL OR ur.expires_at > NOW").
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "role_code", "role_name", "is_system", "role_status", "description", "scope_type", "scope_id", "ur_status", "verified_at", "expires_at"}))

	result, err := m.FindActiveRolesByUserId(context.Background(), 100)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 grants for expired-only user, got %d", len(result))
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

// TestRelUserRoleModel_InsertIgnore_Idempotent — INSERT IGNORE 唯一键冲突幂等语义（T1.6 AssignRole/注册自动分配）
// 断言 SQL 为 `insert ignore into`：成功返回 NewResult(1,1)、重复键返回 NewResult(0,0)，均须 nil error（幂等）
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — QA 补测，RED 摘录见 CHANGELOG 2026-08-12 补测节
func TestRelUserRoleModel_InsertIgnore_Idempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewRelUserRoleModel(conn, nil)

	rel := &RelUserRole{
		UserId:    100,
		RoleId:    1,
		ScopeType: "community",
		ScopeId:   5001,
		Status:    0,
	}

	// GREEN：断言 SQL 为 `insert ignore into`（INSERT IGNORE 幂等语义）
	// 第一次插入成功
	mock.ExpectExec("insert ignore into `rel_user_role`").
		WithArgs(int64(100), int64(1), "community", int64(5001), int64(0), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 重复键（uk_user_role_scope 唯一键冲突）：INSERT IGNORE 返回 0 rows affected 但不报错（幂等）
	mock.ExpectExec("insert ignore into `rel_user_role`").
		WithArgs(int64(100), int64(1), "community", int64(5001), int64(0), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := m.InsertIgnore(context.Background(), rel); err != nil {
		t.Errorf("first insert: expected nil error, got %v", err)
	}
	if err := m.InsertIgnore(context.Background(), rel); err != nil {
		t.Errorf("duplicate insert: expected nil error (INSERT IGNORE idempotent), got %v", err)
	}
}

// TestRelUserRoleModel_FindByRoleId
// need_human 修复回归：FindByRoleId 原 SQL `SELECT id, user_id, role_id, scope_type, scope_id, assign_time FROM ...`
// 引用了 rel_user_role 表不存在的 assign_time 列（live 库仅 id/user_id/role_id/scope_type/scope_id/status/
// verified_at/expires_at/created_at），MySQL 1054 Unknown column 阻断 updaterolelogic.invalidateRoleCache 缓存失效
// （安全漏洞：角色权限变更后授权残留最长 30 分钟）。
// 修复后 SQL 与 FindByUserId 一致改为 `select *`（db tag 映射，go-zero sqlx 支持）。
// RED: 修复前 query 为 `SELECT id, ..., assign_time FROM `rel_user_role` WHERE role_id = ?`，
// 与正则 `select \* from `rel_user_role` where role_id = \?` 不匹配 → sqlmock 报 query not expected → FAIL。
// SEE: [[need_human-findbyroleid-assign_time]]
func TestRelUserRoleModel_FindByRoleId(t *testing.T) {
	tests := []struct {
		name     string
		roleId   int64
		rows     *sqlmock.Rows
		wantLen  int
		wantUser int64
		wantRole int64
	}{
		{
			name:   "命中：列映射到 RelUserRole 全部字段",
			roleId: 1,
			rows: sqlmock.NewRows([]string{"id", "user_id", "role_id", "scope_type", "scope_id", "status", "verified_at", "expires_at", "created_at"}).
				AddRow(10, 100, 1, "community", 5001, 2, nil, nil, time.Now()).
				AddRow(11, 100, 9, "", 0, 0, nil, nil, time.Now()),
			wantLen:  2,
			wantUser: 100,
			wantRole: 1,
		},
		{
			name:     "未命中：返回空 list",
			roleId:   999,
			rows:     sqlmock.NewRows([]string{"id", "user_id", "role_id", "scope_type", "scope_id", "status", "verified_at", "expires_at", "created_at"}),
			wantLen:  0,
			wantUser: 0,
			wantRole: 0,
		},
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

			// 断言 SQL 为 `select * from `rel_user_role` where role_id = ?`，不含不存在的 assign_time 列
			mock.ExpectQuery("select \\* from \\`rel_user_role\\` where role_id = \\?").
				WithArgs(tt.roleId).
				WillReturnRows(tt.rows)

			result, err := m.FindByRoleId(context.Background(), tt.roleId)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.wantLen {
				t.Fatalf("expected %d rows, got %d", tt.wantLen, len(result))
			}
			if tt.wantLen > 0 {
				if result[0].UserId != tt.wantUser {
					t.Errorf("expected user_id %d, got %d", tt.wantUser, result[0].UserId)
				}
				if result[0].RoleId != tt.wantRole {
					t.Errorf("expected role_id %d, got %d", tt.wantRole, result[0].RoleId)
				}
				if result[0].ScopeType != "community" || result[0].ScopeId != 5001 {
					t.Errorf("expected scope community:5001, got %s:%d", result[0].ScopeType, result[0].ScopeId)
				}
			}
		})
	}
}
