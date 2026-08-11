package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ==================== RelRolePermission ====================

// RelRolePermission 角色-权限关联表
type RelRolePermission struct {
	Id           int64     `db:"id"`
	RoleId       int64     `db:"role_id"`
	PermissionId int64     `db:"permission_id"`
	CreatedTime  time.Time `db:"created_at"`
}

type RelRolePermissionModel interface {
	Insert(ctx context.Context, data *RelRolePermission) (int64, error)
	FindByRoleId(ctx context.Context, roleId int64) ([]*RelRolePermission, error)
	DeleteByRoleId(ctx context.Context, roleId int64) error
	BatchInsert(ctx context.Context, records []*RelRolePermission) error
}

type defaultRelRolePermissionModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewRelRolePermissionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RelRolePermissionModel {
	return &defaultRelRolePermissionModel{conn: conn, table: "`rel_role_permission`"}
}

func (m *defaultRelRolePermissionModel) Insert(ctx context.Context, data *RelRolePermission) (int64, error) {
	query := fmt.Sprintf("insert into %s (role_id, permission_id) values (?, ?)", m.table)
	res, err := m.conn.ExecCtx(ctx, query, data.RoleId, data.PermissionId)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultRelRolePermissionModel) FindByRoleId(ctx context.Context, roleId int64) ([]*RelRolePermission, error) {
	var list []*RelRolePermission
	err := m.conn.QueryRowsCtx(ctx, &list, fmt.Sprintf("select * from %s where role_id = ?", m.table), roleId)
	return list, err
}

func (m *defaultRelRolePermissionModel) DeleteByRoleId(ctx context.Context, roleId int64) error {
	_, err := m.conn.ExecCtx(ctx, fmt.Sprintf("delete from %s where role_id = ?", m.table), roleId)
	return err
}

func (m *defaultRelRolePermissionModel) BatchInsert(ctx context.Context, records []*RelRolePermission) error {
	for _, r := range records {
		if _, err := m.Insert(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// ==================== RelUserRole ====================

// RelUserRole 用户-角色关联表（含数据范围，spec/permission.md 新增 scope_type + scope_id）
type RelUserRole struct {
	Id          int64     `db:"id"`
	UserId      int64     `db:"user_id"`
	RoleId      int64     `db:"role_id"`
	ScopeType   string    `db:"scope_type"` // community/building/unit/grid
	ScopeId     int64     `db:"scope_id"`   // 对应层级的实体 ID
	CreatedTime time.Time `db:"created_at"`
}

// UserRoleWithInfo 用户角色详细信息（联表查询结果）
type UserRoleWithInfo struct {
	RoleId      int64  `db:"role_id"`
	RoleCode    string `db:"role_code"`
	RoleName    string `db:"role_name"`
	IsSystem    int64  `db:"is_system"`
	Status      int64  `db:"role_status"`
	Description string `db:"description"`
	ScopeType   string `db:"scope_type"`
	ScopeId     int64  `db:"scope_id"`
}

type RelUserRoleModel interface {
	Insert(ctx context.Context, data *RelUserRole) (int64, error)
	FindByUserId(ctx context.Context, userId int64) ([]*RelUserRole, error)
	FindByRoleId(ctx context.Context, roleId int64) ([]*RelUserRole, error)
	// FindActiveByUserId 联表查询返回角色完整信息
	FindActiveByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error)
	// FindScopesByUserId 根据 user_id + scope_type 返回 scope_ids
	FindScopesByUserId(ctx context.Context, userId int64, scopeType string) ([]int64, error)
	// DeleteByUserIdAndRoleId 删除用户指定角色
	DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64, scopeType string, scopeId int64) error
	// BatchInsertUserRoles 批量插入（使用 INSERT IGNORE）
	BatchInsertUserRoles(ctx context.Context, records []*RelUserRole) error
	// CountByRoleId 统计角色被分配给多少用户
	CountByRoleId(ctx context.Context, roleId int64) (int64, error)
}

type defaultRelUserRoleModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewRelUserRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RelUserRoleModel {
	return &defaultRelUserRoleModel{conn: conn, table: "`rel_user_role`"}
}

func (m *defaultRelUserRoleModel) Insert(ctx context.Context, data *RelUserRole) (int64, error) {
	query := fmt.Sprintf("insert into %s (user_id, role_id, scope_type, scope_id) values (?, ?, ?, ?)", m.table)
	res, err := m.conn.ExecCtx(ctx, query, data.UserId, data.RoleId, data.ScopeType, data.ScopeId)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultRelUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*RelUserRole, error) {
	var list []*RelUserRole
	err := m.conn.QueryRowsCtx(ctx, &list, fmt.Sprintf("select * from %s where user_id = ?", m.table), userId)
	return list, err
}

// FindActiveByUserId 联表查询活跃角色（INNER JOIN sys_role）
func (m *defaultRelUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error) {
	var list []*UserRoleWithInfo
	query := fmt.Sprintf(`
		SELECT ur.role_id, r.role_code, r.role_name, r.is_system, r.status as role_status, r.description,
		       ur.scope_type, ur.scope_id
		FROM %s ur
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND r. deleted_at IS NULL AND r.status = 1
	`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, userId)
	return list, err
}

// FindScopesByUserId 查询用户在某 scope_type 下的所有 scope_id（spec/permission.md GetDataScopes）
func (m *defaultRelUserRoleModel) FindScopesByUserId(ctx context.Context, userId int64, scopeType string) ([]int64, error) {
	var scopeIds []int64
	query := fmt.Sprintf(`
		SELECT DISTINCT ur.scope_id FROM %s ur
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND ur.scope_type = ? AND r. deleted_at IS NULL AND r.status = 1
	`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &scopeIds, query, userId, scopeType)
	return scopeIds, err
}

// DeleteByUserIdAndRoleId 删除用户指定角色+作用域
func (m *defaultRelUserRoleModel) DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64, scopeType string, scopeId int64) error {
	_, err := m.conn.ExecCtx(ctx, fmt.Sprintf("delete from %s where user_id = ? and role_id = ? and scope_type = ? and scope_id = ?", m.table),
		userId, roleId, scopeType, scopeId)
	return err
}

// BatchInsertUserRoles 批量插入（INSERT IGNORE 幂等）
func (m *defaultRelUserRoleModel) BatchInsertUserRoles(ctx context.Context, records []*RelUserRole) error {
	for _, r := range records {
		query := fmt.Sprintf("insert ignore into %s (user_id, role_id, scope_type, scope_id) values (?, ?, ?, ?)", m.table)
		if _, err := m.conn.ExecCtx(ctx, query, r.UserId, r.RoleId, r.ScopeType, r.ScopeId); err != nil {
			return err
		}
	}
	return nil
}

// CountByRoleId 统计角色被分配给多少用户
func (m *defaultRelUserRoleModel) CountByRoleId(ctx context.Context, roleId int64) (int64, error) {
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, fmt.Sprintf("select count(*) from %s where role_id = ?", m.table), roleId)
	return count, err
}

// FindByRoleId 查询持有指定角色的所有用户角色记录
func (m *defaultRelUserRoleModel) FindByRoleId(ctx context.Context, roleId int64) ([]*RelUserRole, error) {
	var list []*RelUserRole
	query := fmt.Sprintf("SELECT id, user_id, role_id, scope_type, scope_id, assign_time FROM %s WHERE role_id = ?", m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, roleId)
	return list, err
}

var _ = time.Now
