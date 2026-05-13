package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthUserRoleModel = (*customAuthUserRoleModel)(nil)

type (
	AuthUserRoleModel interface {
		authUserRoleModel
		FindByUserId(ctx context.Context, userId int64) ([]*AuthUserRole, error)
		DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64) error
		BatchInsertUserRoles(ctx context.Context, userId int64, roleIds []int64) error
		FindActiveByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error)
	}

	customAuthUserRoleModel struct {
		*defaultAuthUserRoleModel
	}

	UserRoleWithInfo struct {
		Id          int64  `db:"id"`
		UserId      int64  `db:"user_id"`
		RoleId      int64  `db:"role_id"`
		RoleName    string `db:"role_name"`
		RoleCode    string `db:"role_code"`
		IsSystem    int64  `db:"is_system"`
		RoleStatus  int64  `db:"role_status"`
		Description string `db:"description"`
	}
)

func NewAuthUserRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthUserRoleModel {
	return &customAuthUserRoleModel{
		defaultAuthUserRoleModel: newAuthUserRoleModel(conn, c, opts...),
	}
}

func (m *customAuthUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*AuthUserRole, error) {
	var list []*AuthUserRole
	query := fmt.Sprintf("select %s from %s where user_id = ?", authUserRoleRows, m.table)
	err := m.QueryRowsNoCache(&list, query, userId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customAuthUserRoleModel) DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64) error {
	query := fmt.Sprintf("delete from %s where user_id = ? and role_id = ?", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, userId, roleId)
	return err
}

func (m *customAuthUserRoleModel) BatchInsertUserRoles(ctx context.Context, userId int64, roleIds []int64) error {
	if len(roleIds) == 0 {
		return nil
	}
	query := fmt.Sprintf("insert ignore into %s (user_id, role_id, created_time) values ", m.table)
	args := make([]interface{}, 0, len(roleIds)*3)
	for i, roleId := range roleIds {
		if i > 0 {
			query += ", "
		}
		query += "(?, ?, NOW())"
		args = append(args, userId, roleId)
	}
	_, err := m.ExecNoCacheCtx(ctx, query, args...)
	return err
}

func (m *customAuthUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*UserRoleWithInfo, error) {
	var list []*UserRoleWithInfo
	query := fmt.Sprintf(
		`SELECT ur.id, ur.user_id, ur.role_id, r.name AS role_name, r.code AS role_code,
		        r.is_system, r.status AS role_status, r.description
		 FROM %s ur
		 INNER JOIN auth_role r ON ur.role_id = r.id
		 WHERE ur.user_id = ? AND r.delete_time IS NULL AND r.status = 1`,
		m.table,
	)
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, userId)
	if err != nil {
		return nil, err
	}
	return list, nil
}