package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ==================== SysRole ====================

// SysRole 角色定义表（spec/permission.md 数据模型）
type SysRole struct {
	Id          int64          `db:"id"`
	RoleCode    string         `db:"role_code"` // owner/property_admin/community_admin/grid_worker
	RoleName    string         `db:"role_name"`
	Description sql.NullString `db:"description"`
	IsSystem    int64          `db:"is_system"` // 1=系统角色，不可删除
	SortOrder   int64          `db:"sort_order"`
	Status      int64          `db:"status"`    // 1=启用 0=禁用
	Platforms   string         `db:"platforms"` // 允许登录的端，逗号分隔：pc,mobile；空=未声明（fail-open）
	CreatedBy   int64          `db:"created_by"`
	CreatedTime time.Time      `db:"created_at"`
	UpdatedTime time.Time      `db:"updated_at"`
	DeleteTime  sql.NullTime   `db:"deleted_at"`
}

// roleSortFieldWhitelist ORDER BY 白名单字面量（二次防御，杜绝用户原始输入拼入 SQL）。
// 与 rpc/internal/logic/permission/sort.go 的 roleSortFieldWhitelist 需保持同步（各自独立定义，双处同步）。
// SEE: [[error-code-literal-bypasses-qa-gate]] — 纵深防御：即使 RPC 层校验被绕过，此处仍回落默认值。
var roleSortFieldWhitelist = map[string]struct{}{
	"id":         {},
	"role_code":  {},
	"role_name":  {},
	"sort_order": {},
	"status":     {},
	"created_at": {},
	"updated_at": {},
}

type SysRoleModel interface {
	Insert(ctx context.Context, data *SysRole) (int64, error)
	FindOne(ctx context.Context, id int64) (*SysRole, error)
	FindByCode(ctx context.Context, code string) (*SysRole, error)
	FindByIds(ctx context.Context, ids []int64) ([]*SysRole, error)
	// FindList 分页查询；sortField/sortOrder 为 RPC 层已校验值，可为空。
	// 空 sortField → 默认首键 sort_order；恒追加 id asc 平局决胜。
	FindList(ctx context.Context, status *int64, page, pageSize int64, sortField, sortOrder string) ([]*SysRole, int64, error)
	Update(ctx context.Context, data *SysRole) error
	SoftDelete(ctx context.Context, id int64) error
}

type defaultSysRoleModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewSysRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysRoleModel {
	return &defaultSysRoleModel{conn: conn, table: "`sys_role`"}
}

func (m *defaultSysRoleModel) Insert(ctx context.Context, data *SysRole) (int64, error) {
	query := fmt.Sprintf("insert into %s (role_code, role_name, description, is_system, sort_order, status, platforms, created_by) values (?, ?, ?, ?, ?, ?, ?, ?)", m.table)
	res, err := m.conn.ExecCtx(ctx, query, data.RoleCode, data.RoleName, data.Description, data.IsSystem, data.SortOrder, data.Status, data.Platforms, data.CreatedBy)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultSysRoleModel) FindOne(ctx context.Context, id int64) (*SysRole, error) {
	var v SysRole
	err := m.conn.QueryRowCtx(ctx, &v, fmt.Sprintf("select * from %s where id = ? and  deleted_at is null", m.table), id)
	return &v, err
}

func (m *defaultSysRoleModel) FindByCode(ctx context.Context, code string) (*SysRole, error) {
	var v SysRole
	err := m.conn.QueryRowCtx(ctx, &v, fmt.Sprintf("select * from %s where role_code = ? and  deleted_at is null limit 1", m.table), code)
	return &v, err
}

func (m *defaultSysRoleModel) FindByIds(ctx context.Context, ids []int64) ([]*SysRole, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	var list []*SysRole
	query := fmt.Sprintf("select * from %s where id in (%s) and  deleted_at is null", m.table, strings.Join(placeholders, ","))
	err := m.conn.QueryRowsCtx(ctx, &list, query, args...)
	return list, err
}

func (m *defaultSysRoleModel) FindList(ctx context.Context, status *int64, page, pageSize int64, sortField, sortOrder string) ([]*SysRole, int64, error) {
	where := "where  deleted_at is null"
	args := make([]interface{}, 0)
	if status != nil {
		where += " and status = ?"
		args = append(args, *status)
	}
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, fmt.Sprintf("select count(*) from %s %s", m.table, where), args...); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var list []*SysRole
	query := fmt.Sprintf("select * from %s %s %s limit %d offset %d", m.table, where, orderByClause(sortField, sortOrder), pageSize, offset)
	err := m.conn.QueryRowsCtx(ctx, &list, query, args...)
	return list, total, err
}

// orderByClause 组装 ORDER BY 子句（白名单字面量二次防御）。
// - 字段转小写比对白名单，非白名单回落默认首键 sort_order
// - 方向仅 asc/desc 字面量二值分支，非 desc 回落 asc
// - 空字段默认 sort_order
// - 恒追加 id asc 平局决胜（tiebreaker），保证分页稳定性
// 用户原始输入（含注入载荷）永不拼入 SQL（REQ-6 安全）。
func orderByClause(sortField, sortOrder string) string {
	field := strings.ToLower(sortField)
	if _, ok := roleSortFieldWhitelist[field]; !ok {
		field = "sort_order"
	}
	direction := "asc"
	if strings.ToLower(sortOrder) == "desc" {
		direction = "desc"
	}
	return fmt.Sprintf("order by %s %s, id asc", field, direction)
}

func (m *defaultSysRoleModel) Update(ctx context.Context, data *SysRole) error {
	// D6: 补 sort_order 落库（既有 Update 遗漏该列，导致前端编辑排序不生效）
	// 参数顺序与占位符一一对应：role_name, description, status, platforms, sort_order, id
	_, err := m.conn.ExecCtx(ctx, fmt.Sprintf("update %s set role_name = ?, description = ?, status = ?, platforms = ?, sort_order = ?, updated_at = now() where id = ?", m.table),
		data.RoleName, data.Description, data.Status, data.Platforms, data.SortOrder, data.Id)
	return err
}

func (m *defaultSysRoleModel) SoftDelete(ctx context.Context, id int64) error {
	_, err := m.conn.ExecCtx(ctx, fmt.Sprintf("update %s set  deleted_at = now() where id = ? and is_system = 0", m.table), id)
	return err
}

// ==================== SysPermission ====================

// SysPermission 权限定义表（树形结构）
type SysPermission struct {
	Id        int64          `db:"id"`
	ParentId  sql.NullInt64  `db:"parent_id"`
	Name      string         `db:"name"`
	Code      string         `db:"code"` // 全局唯一：user:read, user:write
	Type      int64          `db:"type"` // 1=菜单 2=按钮 3=API
	Path      sql.NullString `db:"path"` // API 路径
	Icon      sql.NullString `db:"icon"`
	SortOrder int64          `db:"sort_order"`
	Status    int64          `db:"status"` // 1=启用 0=禁用
	// MinVerfLevel 能力层级（T1.1 迁移）：0=持角色+数据范围即可, 2=需已认证(默认0)
	MinVerfLevel int64     `db:"min_verf_level"`
	CreatedTime  time.Time `db:"created_at"`
	UpdatedTime  time.Time `db:"updated_at"`
}

type SysPermissionModel interface {
	FindAll(ctx context.Context) ([]*SysPermission, error)
	FindByIds(ctx context.Context, ids []int64) ([]*SysPermission, error)
	FindByCode(ctx context.Context, code string) (*SysPermission, error)
	// FindByPath 按 API path（含 METHOD 前缀，如 "GET:/api/users"）查权限定义
	FindByPath(ctx context.Context, path string) (*SysPermission, error)
	FindWithFilter(ctx context.Context, typeFilter, statusFilter *int64) ([]*SysPermission, error)
}

type defaultSysPermissionModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewSysPermissionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysPermissionModel {
	return &defaultSysPermissionModel{conn: conn, table: "`sys_permission`"}
}

func (m *defaultSysPermissionModel) FindAll(ctx context.Context) ([]*SysPermission, error) {
	var list []*SysPermission
	err := m.conn.QueryRowsCtx(ctx, &list, fmt.Sprintf("select * from %s where status = 1 order by sort_order asc", m.table))
	return list, err
}

func (m *defaultSysPermissionModel) FindByIds(ctx context.Context, ids []int64) ([]*SysPermission, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	var list []*SysPermission
	query := fmt.Sprintf("select * from %s where id in (%s) and status = 1", m.table, strings.Join(placeholders, ","))
	err := m.conn.QueryRowsCtx(ctx, &list, query, args...)
	return list, err
}

func (m *defaultSysPermissionModel) FindByCode(ctx context.Context, code string) (*SysPermission, error) {
	var v SysPermission
	err := m.conn.QueryRowCtx(ctx, &v, fmt.Sprintf("select * from %s where code = ? and status = 1 limit 1", m.table), code)
	return &v, err
}

// FindByPath 按 API path（含 METHOD 前缀）查权限定义（T1.5 能力分层 perm:def 缓存回源）
func (m *defaultSysPermissionModel) FindByPath(ctx context.Context, path string) (*SysPermission, error) {
	var v SysPermission
	err := m.conn.QueryRowCtx(ctx, &v, fmt.Sprintf("select * from %s where path = ? and status = 1 limit 1", m.table), path)
	return &v, err
}

// FindWithFilter 根据 type/status 筛选权限列表（构建树之前调用）
//
//	参数为 nil 表示不限制该维度，传入 -1 以外的任何值均可筛选
func (m *defaultSysPermissionModel) FindWithFilter(ctx context.Context, typeFilter, statusFilter *int64) ([]*SysPermission, error) {
	where := "where 1=1"
	args := make([]interface{}, 0)

	if typeFilter != nil {
		where += " and type = ?"
		args = append(args, *typeFilter)
	}
	if statusFilter != nil {
		where += " and status = ?"
		args = append(args, *statusFilter)
	}

	var list []*SysPermission
	query := fmt.Sprintf("select * from %s %s order by sort_order asc", m.table, where)
	err := m.conn.QueryRowsCtx(ctx, &list, query, args...)
	return list, err
}
