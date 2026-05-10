package svc

import (
	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                        config.Config
	MdAdministrativeDivisionModel model.MdAdministrativeDivisionModel
	MdResidentialAreaModel        model.MdResidentialAreaModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	cacheConf := c.Cache
	opts := cache.WithExpiry(3600)

	return &ServiceContext{
		Config:                        c,
		MdAdministrativeDivisionModel: model.NewMdAdministrativeDivisionModel(conn, cacheConf, opts),
		MdResidentialAreaModel:        model.NewMdResidentialAreaModel(conn, cacheConf, opts),
	}
}
