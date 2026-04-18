package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthPropertyUnitModel = (*customAuthPropertyUnitModel)(nil)

type (
	// AuthPropertyUnitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthPropertyUnitModel.
	AuthPropertyUnitModel interface {
		authPropertyUnitModel
	}

	customAuthPropertyUnitModel struct {
		*defaultAuthPropertyUnitModel
	}
)

// NewAuthPropertyUnitModel returns a model for the database table.
func NewAuthPropertyUnitModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthPropertyUnitModel {
	return &customAuthPropertyUnitModel{
		defaultAuthPropertyUnitModel: newAuthPropertyUnitModel(conn, c, opts...),
	}
}
