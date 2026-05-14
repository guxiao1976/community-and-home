package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmHealthCheckModel = (*customAmHealthCheckModel)(nil)

type (
	// AmHealthCheckModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmHealthCheckModel.
	AmHealthCheckModel interface {
		amHealthCheckModel
	}

	customAmHealthCheckModel struct {
		*defaultAmHealthCheckModel
	}
)

// NewAmHealthCheckModel returns a model for the database table.
func NewAmHealthCheckModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmHealthCheckModel {
	return &customAmHealthCheckModel{
		defaultAmHealthCheckModel: newAmHealthCheckModel(conn, c, opts...),
	}
}
