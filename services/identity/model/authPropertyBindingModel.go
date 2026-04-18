package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthPropertyBindingModel = (*customAuthPropertyBindingModel)(nil)

type (
	// AuthPropertyBindingModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthPropertyBindingModel.
	AuthPropertyBindingModel interface {
		authPropertyBindingModel
		FindByUserId(ctx context.Context, userId int64) ([]*AuthPropertyBinding, error)
		FindActiveByUserId(ctx context.Context, userId int64) ([]*AuthPropertyBinding, error)
		FindByPropertyUnitId(ctx context.Context, propertyUnitId int64) ([]*AuthPropertyBinding, error)
		CheckPrimaryUser(ctx context.Context, propertyUnitId int64) (*AuthPropertyBinding, error)
		FindByUserAndProperty(ctx context.Context, userId, propertyUnitId int64) (*AuthPropertyBinding, error)
		FindOneByUserIdPropertyUnitId(ctx context.Context, userId, propertyUnitId int64) (*AuthPropertyBinding, error)
	}

	customAuthPropertyBindingModel struct {
		*defaultAuthPropertyBindingModel
	}
)

// NewAuthPropertyBindingModel returns a model for the database table.
func NewAuthPropertyBindingModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthPropertyBindingModel {
	return &customAuthPropertyBindingModel{
		defaultAuthPropertyBindingModel: newAuthPropertyBindingModel(conn, c, opts...),
	}
}

func (m *customAuthPropertyBindingModel) FindByUserId(ctx context.Context, userId int64) ([]*AuthPropertyBinding, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND delete_time IS NULL", authPropertyBindingRows, m.table)
	var resp []*AuthPropertyBinding
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customAuthPropertyBindingModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*AuthPropertyBinding, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND bind_status = 1 AND delete_time IS NULL", authPropertyBindingRows, m.table)
	var resp []*AuthPropertyBinding
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customAuthPropertyBindingModel) FindByPropertyUnitId(ctx context.Context, propertyUnitId int64) ([]*AuthPropertyBinding, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE property_unit_id = ? AND bind_status = 1 AND delete_time IS NULL", authPropertyBindingRows, m.table)
	var resp []*AuthPropertyBinding
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, propertyUnitId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customAuthPropertyBindingModel) CheckPrimaryUser(ctx context.Context, propertyUnitId int64) (*AuthPropertyBinding, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE property_unit_id = ? AND is_primary = 1 AND bind_status = 1 AND delete_time IS NULL LIMIT 1", authPropertyBindingRows, m.table)
	var resp AuthPropertyBinding
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, propertyUnitId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (m *customAuthPropertyBindingModel) FindByUserAndProperty(ctx context.Context, userId, propertyUnitId int64) (*AuthPropertyBinding, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND property_unit_id = ? AND delete_time IS NULL LIMIT 1", authPropertyBindingRows, m.table)
	var resp AuthPropertyBinding
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, userId, propertyUnitId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAuthPropertyBindingModel) FindOneByUserIdPropertyUnitId(ctx context.Context, userId, propertyUnitId int64) (*AuthPropertyBinding, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND property_unit_id = ? AND delete_time IS NULL LIMIT 1", authPropertyBindingRows, m.table)
	var resp AuthPropertyBinding
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, userId, propertyUnitId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
