package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthRoleModel = (*customAuthRoleModel)(nil)

type (
	// AuthRoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthRoleModel.
	AuthRoleModel interface {
		authRoleModel
		FindList(ctx context.Context, page, pageSize int32, status *int64) ([]*AuthRole, error)
		Count(ctx context.Context, status *int64) (int64, error)
		FindByIds(ctx context.Context, ids []int64) ([]*AuthRole, error)
	}

	customAuthRoleModel struct {
		*defaultAuthRoleModel
	}
)

// NewAuthRoleModel returns a model for the database table.
func NewAuthRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthRoleModel {
	return &customAuthRoleModel{
		defaultAuthRoleModel: newAuthRoleModel(conn, c, opts...),
	}
}

func (m *customAuthRoleModel) FindList(ctx context.Context, page, pageSize int32, status *int64) ([]*AuthRole, error) {
	var roles []*AuthRole
	offset := (page - 1) * pageSize

	query := "SELECT id, name, code, description, is_system, sort_order, status, created_by, created_time, updated_time, delete_time FROM auth_role WHERE delete_time IS NULL"
	args := []interface{}{}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY sort_order ASC, id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	err := m.QueryRowsNoCacheCtx(ctx, &roles, query, args...)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (m *customAuthRoleModel) Count(ctx context.Context, status *int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM auth_role WHERE delete_time IS NULL"
	args := []interface{}{}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *customAuthRoleModel) FindByIds(ctx context.Context, ids []int64) ([]*AuthRole, error) {
	if len(ids) == 0 {
		return []*AuthRole{}, nil
	}

	var roles []*AuthRole
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT id, name, code, description, is_system, sort_order, status, created_by, created_time, updated_time, delete_time FROM auth_role WHERE id IN (%s) AND delete_time IS NULL ORDER BY sort_order ASC",
		strings.Join(placeholders, ","))

	err := m.QueryRowsNoCacheCtx(ctx, &roles, query, args...)
	if err != nil {
		return nil, err
	}

	return roles, nil
}


