package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmModelConfigModel = (*customAmModelConfigModel)(nil)

type (
	// AmModelConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmModelConfigModel.
	AmModelConfigModel interface {
		amModelConfigModel
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
