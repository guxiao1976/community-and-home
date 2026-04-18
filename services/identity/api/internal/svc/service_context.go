// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"github.com/casbin/casbin/v2"
	"github.com/guxiao/community-and-home/services/identity/api/internal/config"
	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	Enforcer                       *casbin.Enforcer
	MinIOClient                    *minio.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	// Initialize Casbin enforcer (optional, only if ModelPath is configured)
	var enforcer *casbin.Enforcer
	if c.Casbin.ModelPath != "" {
		var err error
		enforcer, err = casbin.NewEnforcer(c.Casbin.ModelPath, c.Casbin.PolicyAdapter)
		if err != nil {
			panic(err)
		}
	}

	// Initialize MinIO client
	var minioClient *minio.Client
	if c.MinIO.Endpoint != "" {
		var err error
		minioClient, err = minio.New(c.MinIO.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(c.MinIO.AccessKeyID, c.MinIO.SecretAccessKey, ""),
			Secure: c.MinIO.UseSSL,
		})
		if err != nil {
			panic(err)
		}
	}

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
		Enforcer:                       enforcer,
		MinIOClient:                    minioClient,
	}
}
