package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdSensitiveWordModel = (*customMdSensitiveWordModel)(nil)

type (
	// MdSensitiveWordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMdSensitiveWordModel.
	MdSensitiveWordModel interface {
		mdSensitiveWordModel
		FindOneByWord(ctx context.Context, word string) (*MdSensitiveWord, error)
		FindByCategory(ctx context.Context, category string, limit, offset int) ([]*MdSensitiveWord, int64, error)
	}

	customMdSensitiveWordModel struct {
		*defaultMdSensitiveWordModel
	}
)

// NewMdSensitiveWordModel returns a model for the database table.
func NewMdSensitiveWordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdSensitiveWordModel {
	return &customMdSensitiveWordModel{
		defaultMdSensitiveWordModel: newMdSensitiveWordModel(conn, c, opts...),
	}
}

func (m *customMdSensitiveWordModel) FindOneByWord(ctx context.Context, word string) (*MdSensitiveWord, error) {
	var resp MdSensitiveWord
	query := fmt.Sprintf("select %s from %s where `word` = ? limit 1", mdSensitiveWordRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, word)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMdSensitiveWordModel) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*MdSensitiveWord, int64, error) {
	var resp []*MdSensitiveWord
	var total int64

	query := fmt.Sprintf("select %s from %s", mdSensitiveWordRows, m.table)
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
