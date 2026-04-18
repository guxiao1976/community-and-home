package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdConfigurationModel = (*customMdConfigurationModel)(nil)

type (
	// MdConfigurationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMdConfigurationModel.
	MdConfigurationModel interface {
		mdConfigurationModel
		FindByModule(ctx context.Context, module string, limit, offset int) ([]*MdConfiguration, int64, error)
		FindAll(ctx context.Context, limit, offset int) ([]*MdConfiguration, int64, error)
	}

	customMdConfigurationModel struct {
		*defaultMdConfigurationModel
	}
)

// NewMdConfigurationModel returns a model for the database table.
func NewMdConfigurationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdConfigurationModel {
	return &customMdConfigurationModel{
		defaultMdConfigurationModel: newMdConfigurationModel(conn, c, opts...),
	}
}

func (m *customMdConfigurationModel) FindOneByConfigKey(ctx context.Context, configKey string) (*MdConfiguration, error) {
	var resp MdConfiguration
	query := fmt.Sprintf("select %s from %s where `config_key` = ? limit 1", mdConfigurationRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, configKey)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMdConfigurationModel) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*MdConfiguration, int64, error) {
	var resp []*MdConfiguration
	var total int64

	query := fmt.Sprintf("select %s from %s", mdConfigurationRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s", m.table)

	if category != "" {
		query += " where `category` = ?"
		countQuery += " where `category` = ?"
	}

	query += " order by id desc limit ? offset ?"

	var err error
	if category != "" {
		err = m.QueryRowsNoCacheCtx(ctx, &resp, query, category, limit, offset)
		err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery, category)
		if err2 != nil {
			return nil, 0, err2
		}
	} else {
		err = m.QueryRowsNoCacheCtx(ctx, &resp, query, limit, offset)
		err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery)
		if err2 != nil {
			return nil, 0, err2
		}
	}

	switch err {
	case nil:
		return resp, total, nil
	case sqlx.ErrNotFound:
		return nil, 0, nil
	default:
		return nil, 0, err
	}
}

func (m *customMdConfigurationModel) FindByModule(ctx context.Context, module string, limit, offset int) ([]*MdConfiguration, int64, error) {
	var resp []*MdConfiguration
	var total int64

	query := fmt.Sprintf("select %s from %s where `module` = ? order by id desc limit ? offset ?", mdConfigurationRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s where `module` = ?", m.table)

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, module, limit, offset)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery, module)
	if err2 != nil {
		return nil, 0, err2
	}

	return resp, total, nil
}

func (m *customMdConfigurationModel) FindAll(ctx context.Context, limit, offset int) ([]*MdConfiguration, int64, error) {
	var resp []*MdConfiguration
	var total int64

	query := fmt.Sprintf("select %s from %s order by id desc limit ? offset ?", mdConfigurationRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s", m.table)

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, limit, offset)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery)
	if err2 != nil {
		return nil, 0, err2
	}

	return resp, total, nil
}
