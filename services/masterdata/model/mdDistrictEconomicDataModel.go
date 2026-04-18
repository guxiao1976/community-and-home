package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdDistrictEconomicDataModel = (*customMdDistrictEconomicDataModel)(nil)

type (
	// MdDistrictEconomicDataModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMdDistrictEconomicDataModel.
	MdDistrictEconomicDataModel interface {
		mdDistrictEconomicDataModel
	}

	customMdDistrictEconomicDataModel struct {
		*defaultMdDistrictEconomicDataModel
	}
)

// NewMdDistrictEconomicDataModel returns a model for the database table.
func NewMdDistrictEconomicDataModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdDistrictEconomicDataModel {
	return &customMdDistrictEconomicDataModel{
		defaultMdDistrictEconomicDataModel: newMdDistrictEconomicDataModel(conn, c, opts...),
	}
}
