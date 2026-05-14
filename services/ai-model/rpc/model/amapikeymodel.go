package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmApiKeyModel = (*customAmApiKeyModel)(nil)

type (
	// AmApiKeyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmApiKeyModel.
	AmApiKeyModel interface {
		amApiKeyModel
	}

	customAmApiKeyModel struct {
		*defaultAmApiKeyModel
	}
)

// NewAmApiKeyModel returns a model for the database table.
func NewAmApiKeyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmApiKeyModel {
	return &customAmApiKeyModel{
		defaultAmApiKeyModel: newAmApiKeyModel(conn, c, opts...),
	}
}
