package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmModelConfigModel = (*customAmModelConfigModel)(nil)

type (
	// AmModelConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmModelConfigModel.
	AmModelConfigModel interface {
		amModelConfigModel
		FindList(ctx context.Context, provider string, status int64, page, pageSize int64) ([]*AmModelConfig, error)
		Count(ctx context.Context, provider string, status int64) (int64, error)
	}

	customAmModelConfigModel struct {
		*defaultAmModelConfigModel
	}
)

// NewAmModelConfigModel returns a model for the database table.
func NewAmModelConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmModelConfigModel {
	return &customAmModelConfigModel{
		defaultAmModelConfigModel: newAmModelConfigModel(conn, c, opts...),
	}
}

// FindList queries model configurations with pagination and filtering
func (m *customAmModelConfigModel) FindList(ctx context.Context, provider string, status int64, page, pageSize int64) ([]*AmModelConfig, error) {
	var conditions []string
	var args []interface{}

	if provider != "" {
		conditions = append(conditions, "provider = ?")
		args = append(args, provider)
	}

	if status > 0 {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY id DESC LIMIT ? OFFSET ?", amModelConfigRows, m.table, whereClause)
	args = append(args, pageSize, offset)

	var resp []*AmModelConfig
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Count returns the total count of model configurations matching the filters
func (m *customAmModelConfigModel) Count(ctx context.Context, provider string, status int64) (int64, error) {
	var conditions []string
	var args []interface{}

	if provider != "" {
		conditions = append(conditions, "provider = ?")
		args = append(args, provider)
	}

	if status > 0 {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", m.table, whereClause)

	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
