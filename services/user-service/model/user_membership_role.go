package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserMembershipRole 角色表
type UserMembershipRole struct {
	Id           int64         `db:"id"`
	UserId       int64         `db:"user_id"`
	MembershipId sql.NullInt64 `db:"membership_id"`
	CommunityId  int64         `db:"community_id"`
	RoleCode     string        `db:"role_code"`
	VerfStatus   int64         `db:"verf_status"`
	VerifiedAt   sql.NullTime  `db:"verified_at"`
	ExpiresAt    sql.NullTime  `db:"expires_at"`
	CreatedTime  time.Time     `db:"created_time"`
	UpdatedTime  time.Time     `db:"updated_time"`
}

type UserMembershipRoleModel interface {
	Insert(ctx context.Context, data *UserMembershipRole) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserMembershipRole, error)
	FindByMembershipAndRole(ctx context.Context, membershipId int64, roleCode string) (*UserMembershipRole, error)
	FindByUserAndCommunity(ctx context.Context, userId, communityId int64) ([]*UserMembershipRole, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserMembershipRole, error)
	FindApprovedByUser(ctx context.Context, userId int64, communityId int64, roleCodes []string) ([]*UserMembershipRole, error)
	FindExpiredRoles(ctx context.Context) ([]*UserMembershipRole, error)
	UpdateVerfStatus(ctx context.Context, id int64, verfStatus int64, verifiedAt, expiresAt sql.NullTime) error
	UpdateVerfStatusOnly(ctx context.Context, id int64, verfStatus int64) error
}

type defaultUserMembershipRoleModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserMembershipRoleModel(conn sqlx.SqlConn) UserMembershipRoleModel {
	return &defaultUserMembershipRoleModel{
		conn:  conn,
		table: "user_membership_role",
	}
}

func (m *defaultUserMembershipRoleModel) Insert(ctx context.Context, data *UserMembershipRole) (sql.Result, error) {
	query := fmt.Sprintf(`INSERT INTO %s (id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, m.table)
	return m.conn.ExecCtx(ctx, query, data.Id, data.UserId, data.MembershipId, data.CommunityId, data.RoleCode, data.VerfStatus, data.VerifiedAt, data.ExpiresAt, data.CreatedTime, data.UpdatedTime)
}

func (m *defaultUserMembershipRoleModel) FindOne(ctx context.Context, id int64) (*UserMembershipRole, error) {
	query := fmt.Sprintf(`SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM %s WHERE id = ?`, m.table)
	var resp UserMembershipRole
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserMembershipRoleModel) FindByMembershipAndRole(ctx context.Context, membershipId int64, roleCode string) (*UserMembershipRole, error) {
	query := fmt.Sprintf(`SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM %s WHERE membership_id = ? AND role_code = ?`, m.table)
	var resp UserMembershipRole
	err := m.conn.QueryRowCtx(ctx, &resp, query, membershipId, roleCode)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserMembershipRoleModel) FindByUserAndCommunity(ctx context.Context, userId, communityId int64) ([]*UserMembershipRole, error) {
	query := fmt.Sprintf(`SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM %s WHERE user_id = ? AND community_id = ?`, m.table)
	var resp []*UserMembershipRole
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, communityId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserMembershipRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*UserMembershipRole, error) {
	query := fmt.Sprintf(`SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM %s WHERE user_id = ?`, m.table)
	var resp []*UserMembershipRole
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindApprovedByUser 查询用户已认证的角色，支持按 community_id 和 role_codes 过滤
// communityId=0 时不限制小区（用于 merchant 等全局角色）；roleCodes 为空时返回所有认证角色
func (m *defaultUserMembershipRoleModel) FindApprovedByUser(ctx context.Context, userId int64, communityId int64, roleCodes []string) ([]*UserMembershipRole, error) {
	args := []interface{}{userId, RoleVerfStatusApproved}
	where := "WHERE user_id = ? AND verf_status = ?"

	if communityId > 0 {
		where += " AND community_id = ?"
		args = append(args, communityId)
	}

	if len(roleCodes) > 0 {
		placeholders := make([]string, len(roleCodes))
		for i, rc := range roleCodes {
			placeholders[i] = "?"
			args = append(args, rc)
		}
		where += fmt.Sprintf(" AND role_code IN (%s)", strings.Join(placeholders, ","))
	}

	query := fmt.Sprintf("SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM %s %s", m.table, where)
	var resp []*UserMembershipRole
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserMembershipRoleModel) FindExpiredRoles(ctx context.Context) ([]*UserMembershipRole, error) {
	query := fmt.Sprintf(`SELECT id, user_id, membership_id, community_id, role_code, verf_status, verified_at, expires_at, created_time, updated_time FROM %s WHERE verf_status = ? AND expires_at IS NOT NULL AND expires_at < NOW()`, m.table)
	var resp []*UserMembershipRole
	err := m.conn.QueryRowsCtx(ctx, &resp, query, RoleVerfStatusApproved)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserMembershipRoleModel) UpdateVerfStatus(ctx context.Context, id int64, verfStatus int64, verifiedAt, expiresAt sql.NullTime) error {
	query := fmt.Sprintf(`UPDATE %s SET verf_status=?, verified_at=?, expires_at=?, updated_time=? WHERE id=?`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, verfStatus, verifiedAt, expiresAt, time.Now(), id)
	return err
}

func (m *defaultUserMembershipRoleModel) UpdateVerfStatusOnly(ctx context.Context, id int64, verfStatus int64) error {
	query := fmt.Sprintf(`UPDATE %s SET verf_status=?, updated_time=? WHERE id=?`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, verfStatus, time.Now(), id)
	return err
}
