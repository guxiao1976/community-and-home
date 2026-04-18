package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthPermissionModel = (*customAuthPermissionModel)(nil)

type (
	// AuthPermissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthPermissionModel.
	AuthPermissionModel interface {
		authPermissionModel
		FindByIds(ctx context.Context, ids []int64) ([]*AuthPermission, error)
		FindAll(ctx context.Context, status *int64) ([]*AuthPermission, error)
	}

	customAuthPermissionModel struct {
		*defaultAuthPermissionModel
	}
)

// NewAuthPermissionModel returns a model for the database table.
func NewAuthPermissionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthPermissionModel {
	return &customAuthPermissionModel{
		defaultAuthPermissionModel: newAuthPermissionModel(conn, c, opts...),
	}
}

func (m *customAuthPermissionModel) FindByIds(ctx context.Context, ids []int64) ([]*AuthPermission, error) {
	if len(ids) == 0 {
		return []*AuthPermission{}, nil
	}

	var permissions []*AuthPermission
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id IN (%s) AND delete_time IS NULL ORDER BY sort_order ASC",
		authPermissionRows, m.table, strings.Join(placeholders, ","))

	err := m.QueryRowsNoCacheCtx(ctx, &permissions, query, args...)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (m *customAuthPermissionModel) FindAll(ctx context.Context, status *int64) ([]*AuthPermission, error) {
	var permissions []*AuthPermission
	query := fmt.Sprintf("SELECT %s FROM %s WHERE delete_time IS NULL", authPermissionRows, m.table)
	args := []interface{}{}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY sort_order ASC"

	err := m.QueryRowsNoCacheCtx(ctx, &permissions, query, args...)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

