package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmCallLogModel = (*customAmCallLogModel)(nil)

type (
	// AmCallLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmCallLogModel.
	AmCallLogModel interface {
		amCallLogModel
	}

	customAmCallLogModel struct {
		*defaultAmCallLogModel
	}
)

// NewAmCallLogModel returns a model for the database table.
func NewAmCallLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmCallLogModel {
	return &customAmCallLogModel{
		defaultAmCallLogModel: newAmCallLogModel(conn, c, opts...),
	}
}
