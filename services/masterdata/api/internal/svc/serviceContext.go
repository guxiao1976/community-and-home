package svc

import (
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/config"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/sync"
	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                        config.Config
	MdAdministrativeDivisionModel model.MdAdministrativeDivisionModel
	MdResidentialAreaModel        model.MdResidentialAreaModel
	MdConfigurationModel          model.MdConfigurationModel
	MdSensitiveWordModel          model.MdSensitiveWordModel
	MdDivisionStatisticsModel     model.MdDivisionStatisticsModel
	SubmissionRecordModel         model.SubmissionRecordModel
	SyncEngine                    *sync.SyncEngine
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	ctx := &ServiceContext{
		Config:                        c,
		MdAdministrativeDivisionModel: model.NewMdAdministrativeDivisionModel(conn, c.Cache),
		MdResidentialAreaModel:        model.NewMdResidentialAreaModel(conn, c.Cache),
		MdConfigurationModel:          model.NewMdConfigurationModel(conn, c.Cache),
		MdSensitiveWordModel:          model.NewMdSensitiveWordModel(conn, c.Cache),
		MdDivisionStatisticsModel:     model.NewMdDivisionStatisticsModel(conn, c.Cache),
		SubmissionRecordModel:         model.NewSubmissionRecordModel(conn),
	}
	ctx.SyncEngine = sync.NewSyncEngine(c.AMapKey, ctx.MdAdministrativeDivisionModel, ctx.MdResidentialAreaModel)
	return ctx
}
