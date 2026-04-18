package svc

import (
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/config"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                        config.Config
	MdAdministrativeDivisionModel model.MdAdministrativeDivisionModel
	MdCommunityModel              model.MdCommunityModel
	MdAuditLogModel               model.MdAuditLogModel
	MdConfigurationModel          model.MdConfigurationModel
	MdSensitiveWordModel          model.MdSensitiveWordModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	cacheConf := c.Cache
	opts := cache.WithExpiry(3600)

	return &ServiceContext{
		Config:                        c,
		MdAdministrativeDivisionModel: model.NewMdAdministrativeDivisionModel(conn, cacheConf, opts),
		MdCommunityModel:              model.NewMdCommunityModel(conn, cacheConf, opts),
		MdAuditLogModel:               model.NewMdAuditLogModel(conn, cacheConf, opts),
		MdConfigurationModel:          model.NewMdConfigurationModel(conn, cacheConf, opts),
		MdSensitiveWordModel:          model.NewMdSensitiveWordModel(conn, cacheConf, opts),
	}
}