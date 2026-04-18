package svc

import (
	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/guxiao/community-and-home/services/identity/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                         config.Config
	AuthUserModel                  model.AuthUserModel
	AuthRoleModel                  model.AuthRoleModel
	AuthPermissionModel            model.AuthPermissionModel
	AuthRolePermissionModel        model.AuthRolePermissionModel
	AuthUserRoleModel              model.AuthUserRoleModel
	AuthPropertyUnitModel          model.AuthPropertyUnitModel
	AuthPropertyBindingModel       model.AuthPropertyBindingModel
	AuthHomeownerVerificationModel model.AuthHomeownerVerificationModel
	AuthFamilyModel                model.AuthFamilyModel
	AuthFamilyMemberModel          model.AuthFamilyMemberModel
	AuthUploadedFileModel          model.AuthUploadedFileModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:                         c,
		AuthUserModel:                  model.NewAuthUserModel(conn, c.Cache),
		AuthRoleModel:                  model.NewAuthRoleModel(conn, c.Cache),
		AuthPermissionModel:            model.NewAuthPermissionModel(conn, c.Cache),
		AuthRolePermissionModel:        model.NewAuthRolePermissionModel(conn, c.Cache),
		AuthUserRoleModel:              model.NewAuthUserRoleModel(conn, c.Cache),
		AuthPropertyUnitModel:          model.NewAuthPropertyUnitModel(conn, c.Cache),
		AuthPropertyBindingModel:       model.NewAuthPropertyBindingModel(conn, c.Cache),
		AuthHomeownerVerificationModel: model.NewAuthHomeownerVerificationModel(conn, c.Cache),
		AuthFamilyModel:                model.NewAuthFamilyModel(conn, c.Cache),
		AuthFamilyMemberModel:          model.NewAuthFamilyMemberModel(conn, c.Cache),
		AuthUploadedFileModel:          model.NewAuthUploadedFileModel(conn, c.Cache),
	}
}

