package model

import (
	"context"
	"database/sql"
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

// scope 三态语义常量（access-data-permission 阶段①）
//
//	三态：global（全局放行）/ limited（限定节点）/ empty（无数据范围）
//	scope_id=0 约定为「非实体占位」：empty 行（scope_type=''）零贡献；global 行全放行
//
// SEE: [[is-system-no-permission-shortcut]] — 空(empty) ≠ global，杜绝「空当 global」灾难
const (
	ScopeTypeGlobal    = "global"    // 全局放行（审核员 / sys_admin / moderation 系统身份）
	ScopeTypeEmpty     = ""          // 空数据范围（仅 registered_user 基角色）
	ScopeTypeCommunity = "community" // 限定：小区
	ScopeTypeBuilding  = "building"  // 限定：楼栋
	ScopeTypeUnit      = "unit"      // 限定：单元
	ScopeTypeGrid      = "grid"      // 限定：网格
)

// RelUserRole 用户-角色关联表（含数据范围，spec/permission.md 新增 scope_type + scope_id）
type RelUserRole struct {
	Id          int64        `db:"id"`
	UserId      int64        `db:"user_id"`
	RoleId      int64        `db:"role_id"`
	ScopeType   string       `db:"scope_type"`  // community/building/unit/grid
	ScopeId     int64        `db:"scope_id"`    // 对应层级的实体 ID
	Status      int64        `db:"status"`      // 个体角色生命周期: 0=未认证 1=待审 2=已认证 3=已驳回 4=已过期
	VerifiedAt  sql.NullTime `db:"verified_at"` // 个体认证通过时间
	ExpiresAt   sql.NullTime `db:"expires_at"`  // 个体角色到期时间, NULL=永久
	CreatedTime time.Time    `db:"created_at"`
}

// UserRoleWithInfo 用户角色详细信息（联表查询结果）
type UserRoleWithInfo struct {
	RoleId      int64        `db:"role_id"`
	RoleCode    string       `db:"role_code"`
	RoleName    string       `db:"role_name"`
	IsSystem    int64        `db:"is_system"`
	Status      int64        `db:"role_status"` // sys_role.status（角色定义状态）
	Description string       `db:"description"`
	Platforms   string       `db:"platforms"` // 允许登录的端，逗号分隔：pc,mobile；空=未声明（fail-open）
	ScopeType   string       `db:"scope_type"`
	ScopeId     int64        `db:"scope_id"`
	URStatus    int64        `db:"ur_status"` // rel_user_role.status（个体角色生命周期）
	VerifiedAt  sql.NullTime `db:"verified_at"`
	ExpiresAt   sql.NullTime `db:"expires_at"`
}

type RelUserRoleModel interface {
	Insert(ctx context.Context, data *RelUserRole) (int64, error)
	// InsertIgnore 幂等插入（INSERT IGNORE）：唯一键冲突不报错（T1.6）
	InsertIgnore(ctx context.Context, data *RelUserRole) error
	FindByUserId(ctx context.Context, userId int64) ([]*RelUserRole, error)
	FindByRoleId(ctx context.Context, roleId int64) ([]*RelUserRole, error)
	// FindActiveByUserId 联表查询返回角色完整信息（仅已认证 status=2）
	FindActiveByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error)
	// FindActiveRolesByUserId 联表查询活跃 grants（status IN (0,1,2) 且未过期）
	// 返回含 scope_type/scope_id/verified_at/ur_status，供能力分层聚合
	FindActiveRolesByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error)
	// FindScopesByUserId 根据 user_id + scope_type 返回 scope_ids
	FindScopesByUserId(ctx context.Context, userId int64, scopeType string) ([]int64, error)
	// DeleteByUserIdAndRoleId 删除用户指定角色
	DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64, scopeType string, scopeId int64) error
	// BatchInsertUserRoles 批量插入（使用 INSERT IGNORE）
	BatchInsertUserRoles(ctx context.Context, records []*RelUserRole) error
	// CountByRoleId 统计角色被分配给多少用户
	CountByRoleId(ctx context.Context, roleId int64) (int64, error)
	// CountActiveByRoleAndScope 统计某角色在某作用域（scope_type+scope_id）下的活跃 grants 数量
	//（status IN (0,1,2)：未认证/待审/已认证；驳回 3 与过期 4 不计）。
	// excludeUserId > 0 时排除该用户（幂等重复申请不计入，避免误拒）。
	CountActiveByRoleAndScope(ctx context.Context, roleId int64, scopeType string, scopeId, excludeUserId int64) (int64, error)
	// UpdateRoleStatus 更新用户角色的生命周期状态（认证通过/驳回/过期）
	UpdateRoleStatus(ctx context.Context, userId, roleId int64, scopeType string, scopeId, status int64, verifiedAt, expiresAt sql.NullTime) error
	// FindAllByUserId 联表查询用户所有角色（含个体生命周期状态，不过滤）
	FindAllByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error)
}

type defaultRelUserRoleModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewRelUserRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) RelUserRoleModel {
	return &defaultRelUserRoleModel{conn: conn, table: "`rel_user_role`"}
}

func (m *defaultRelUserRoleModel) Insert(ctx context.Context, data *RelUserRole) (int64, error) {
	query := fmt.Sprintf("insert into %s (user_id, role_id, scope_type, scope_id, status, verified_at, expires_at) values (?, ?, ?, ?, ?, ?, ?)", m.table)
	res, err := m.conn.ExecCtx(ctx, query, data.UserId, data.RoleId, data.ScopeType, data.ScopeId, data.Status, data.VerifiedAt, data.ExpiresAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// InsertIgnore 幂等插入（INSERT IGNORE，T1.6）
// uk_user_role_scope 唯一键冲突时静默跳过（不报错），支撑注册自动分配/Join 自动授权幂等
func (m *defaultRelUserRoleModel) InsertIgnore(ctx context.Context, data *RelUserRole) error {
	query := fmt.Sprintf("insert ignore into %s (user_id, role_id, scope_type, scope_id, status, verified_at, expires_at) values (?, ?, ?, ?, ?, ?, ?)", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.UserId, data.RoleId, data.ScopeType, data.ScopeId, data.Status, data.VerifiedAt, data.ExpiresAt)
	return err
}

func (m *defaultRelUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*RelUserRole, error) {
	var list []*RelUserRole
	err := m.conn.QueryRowsCtx(ctx, &list, fmt.Sprintf("select * from %s where user_id = ?", m.table), userId)
	return list, err
}

// FindActiveByUserId 联表查询活跃角色（INNER JOIN sys_role）
// 只返回已认证（status=2）且未过期的个体角色
func (m *defaultRelUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error) {
	var list []*UserRoleWithInfo
	query := fmt.Sprintf(`
		SELECT ur.role_id, r.role_code, r.role_name, r.is_system, r.status as role_status, r.description, r.platforms,
		       ur.scope_type, ur.scope_id, ur.status as ur_status, ur.verified_at, ur.expires_at
		FROM %s ur
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND r. deleted_at IS NULL AND r.status = 1
		  AND ur.status = 2 AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
	`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, userId)
	return list, err
}

// FindScopesByUserId 查询用户在某 scope_type 下的所有 scope_id（spec/permission.md GetDataScopes）
// T1.2 三态语义：仅取 status IN (0,1,2) 且 scope_id != 0 的 limited 并集
//
//	scope_id=0 是 empty/global 占位，不进并集（REQ-A / REQ-1.2）
func (m *defaultRelUserRoleModel) FindScopesByUserId(ctx context.Context, userId int64, scopeType string) ([]int64, error) {
	var scopeIds []int64
	query := fmt.Sprintf(`
		SELECT DISTINCT ur.scope_id FROM %s ur
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND ur.scope_type = ? AND r.deleted_at IS NULL AND r.status = 1
		  AND ur.status IN (0,1,2) AND ur.scope_id != 0
	`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &scopeIds, query, userId, scopeType)
	return scopeIds, err
}

// FindActiveRolesByUserId 联表查询活跃 grants（T1.2 新增）
// 返回 status IN (0,1,2) 且未过期（expires_at IS NULL OR > NOW()）的 grants
// 含 scope_type/scope_id/verified_at/ur_status，供 CheckPermission 能力分层聚合（T1.5）
func (m *defaultRelUserRoleModel) FindActiveRolesByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error) {
	var list []*UserRoleWithInfo
	query := fmt.Sprintf(`
		SELECT ur.role_id, r.role_code, r.role_name, r.is_system, r.status as role_status, r.description, r.platforms,
		       ur.scope_type, ur.scope_id, ur.status as ur_status, ur.verified_at, ur.expires_at
		FROM %s ur
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND r.deleted_at IS NULL AND r.status = 1
		  AND ur.status IN (0,1,2) AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		ORDER BY ur.id
	`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, userId)
	return list, err
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
		query := fmt.Sprintf("insert ignore into %s (user_id, role_id, scope_type, scope_id, status, verified_at, expires_at) values (?, ?, ?, ?, ?, ?, ?)", m.table)
		if _, err := m.conn.ExecCtx(ctx, query, r.UserId, r.RoleId, r.ScopeType, r.ScopeId, r.Status, r.VerifiedAt, r.ExpiresAt); err != nil {
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

// CountActiveByRoleAndScope 统计某角色在某作用域（scope_type+scope_id）下的活跃 grants 数量
// 活跃 = status IN (0,1,2)（未认证/待审/已认证；驳回 3 与过期 4 不计）。
// excludeUserId > 0 时追加 `user_id != ?` 排除该用户（幂等重复申请不计入，避免误拒）。
// 支撑 AssignRole 每小区 community_admin 上限 3 人（用户拍板 2026-08-17）。
func (m *defaultRelUserRoleModel) CountActiveByRoleAndScope(ctx context.Context, roleId int64, scopeType string, scopeId, excludeUserId int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where role_id = ? and scope_type = ? and scope_id = ? and status in (0,1,2)", m.table)
	args := []any{roleId, scopeType, scopeId}
	if excludeUserId > 0 {
		query += " and user_id != ?"
		args = append(args, excludeUserId)
	}
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

// FindByRoleId 查询持有指定角色的所有用户角色记录
// need_human 修复：原 SQL 显式列清单引用了 rel_user_role 不存在的 assign_time 列（live 库仅
// id/user_id/role_id/scope_type/scope_id/status/verified_at/expires_at/created_at）→ MySQL 1054
// 阻断 updaterolelogic.invalidateRoleCache 缓存失效（安全漏洞）。改用 `select *`（与 FindByUserId 一致，
// RelUserRole 以 db tag 映射，go-zero sqlx 支持），不再引用不存在的列。
func (m *defaultRelUserRoleModel) FindByRoleId(ctx context.Context, roleId int64) ([]*RelUserRole, error) {
	var list []*RelUserRole
	query := fmt.Sprintf("select * from %s where role_id = ?", m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, roleId)
	return list, err
}

var _ = time.Now

// UpdateRoleStatus 更新用户角色的生命周期状态（认证通过/驳回/过期）
func (m *defaultRelUserRoleModel) UpdateRoleStatus(ctx context.Context, userId, roleId int64, scopeType string, scopeId, status int64, verifiedAt, expiresAt sql.NullTime) error {
	query := fmt.Sprintf("update %s set status = ?, verified_at = ?, expires_at = ? where user_id = ? and role_id = ? and scope_type = ? and scope_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, verifiedAt, expiresAt, userId, roleId, scopeType, scopeId)
	return err
}

// FindAllByUserId 联表查询用户所有角色（含个体生命周期状态，不过滤 status）
func (m *defaultRelUserRoleModel) FindAllByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error) {
	var list []*UserRoleWithInfo
	query := fmt.Sprintf(`
		SELECT ur.role_id, r.role_code, r.role_name, r.is_system, r.status as role_status, r.description, r.platforms,
		       ur.scope_type, ur.scope_id, ur.status as ur_status, ur.verified_at, ur.expires_at
		FROM %s ur
		INNER JOIN sys_role r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND r. deleted_at IS NULL
		ORDER BY ur.id
	`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, userId)
	return list, err
}
