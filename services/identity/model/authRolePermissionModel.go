package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthRolePermissionModel = (*customAuthRolePermissionModel)(nil)

type (
	AuthRolePermissionModel interface {
		authRolePermissionModel
		FindByRoleId(ctx context.Context, roleId int64) ([]*AuthRolePermission, error)
		DeleteByRoleId(ctx context.Context, roleId int64) error
		BatchInsert(ctx context.Context, roleId int64, permissionIds []int64) error
	}

	customAuthRolePermissionModel struct {
		*defaultAuthRolePermissionModel
	}
)

func NewAuthRolePermissionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthRolePermissionModel {
	return &customAuthRolePermissionModel{
		defaultAuthRolePermissionModel: newAuthRolePermissionModel(conn, c, opts...),
	}
}

func (m *customAuthRolePermissionModel) FindByRoleId(ctx context.Context, roleId int64) ([]*AuthRolePermission, error) {
	var list []*AuthRolePermission
	query := fmt.Sprintf("select %s from %s where role_id = ?", authRolePermissionRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &list, query, roleId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (m *customAuthRolePermissionModel) DeleteByRoleId(ctx context.Context, roleId int64) error {
	query := fmt.Sprintf("delete from %s where role_id = ?", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, roleId)
	return err
}

func (m *customAuthRolePermissionModel) BatchInsert(ctx context.Context, roleId int64, permissionIds []int64) error {
	if len(permissionIds) == 0 {
		return nil
	}

	query := fmt.Sprintf("insert into %s (role_id, permission_id, created_time) values ", m.table)
	args := make([]interface{}, 0, len(permissionIds)*3)

	for i, permissionId := range permissionIds {
		if i > 0 {
			query += ", "
		}
		query += "(?, ?, NOW())"
		args = append(args, roleId, permissionId)
	}

	_, err := m.ExecNoCacheCtx(ctx, query, args...)
	return err
}
