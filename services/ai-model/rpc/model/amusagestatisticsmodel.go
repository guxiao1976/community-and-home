package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmUsageStatisticsModel = (*customAmUsageStatisticsModel)(nil)

type (
	// AmUsageStatisticsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmUsageStatisticsModel.
	AmUsageStatisticsModel interface {
		amUsageStatisticsModel
	}

	customAmUsageStatisticsModel struct {
		*defaultAmUsageStatisticsModel
	}
)

// NewAmUsageStatisticsModel returns a model for the database table.
func NewAmUsageStatisticsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmUsageStatisticsModel {
	return &customAmUsageStatisticsModel{
		defaultAmUsageStatisticsModel: newAmUsageStatisticsModel(conn, c, opts...),
	}
}
