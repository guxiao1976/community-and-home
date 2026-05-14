package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmAlertRecordModel = (*customAmAlertRecordModel)(nil)

type (
	// AmAlertRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmAlertRecordModel.
	AmAlertRecordModel interface {
		amAlertRecordModel
	}

	customAmAlertRecordModel struct {
		*defaultAmAlertRecordModel
	}
)

// NewAmAlertRecordModel returns a model for the database table.
func NewAmAlertRecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmAlertRecordModel {
	return &customAmAlertRecordModel{
		defaultAmAlertRecordModel: newAmAlertRecordModel(conn, c, opts...),
	}
}
