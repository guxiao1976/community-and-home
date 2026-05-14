package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmCostAlertConfigModel = (*customAmCostAlertConfigModel)(nil)

type (
	// AmCostAlertConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmCostAlertConfigModel.
	AmCostAlertConfigModel interface {
		amCostAlertConfigModel
	}

	customAmCostAlertConfigModel struct {
		*defaultAmCostAlertConfigModel
	}
)

// NewAmCostAlertConfigModel returns a model for the database table.
func NewAmCostAlertConfigModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmCostAlertConfigModel {
	return &customAmCostAlertConfigModel{
		defaultAmCostAlertConfigModel: newAmCostAlertConfigModel(conn, c, opts...),
	}
}
