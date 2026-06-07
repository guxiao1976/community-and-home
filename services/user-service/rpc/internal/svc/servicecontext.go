package svc

import (
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ServiceContext 用户中心 RPC 服务上下文
type ServiceContext struct {
	Config config.Config

	Redis *redis.Redis // Redis 缓存客户端

	UserBaseModel               model.UserBaseModel
	UserCommunityMembershipModel model.UserCommunityMembershipModel
	UserMembershipRoleModel      model.UserMembershipRoleModel
	UserCertificationModel       model.UserCertificationModel
	UserResidenceModel           model.UserResidenceModel
}

// NewServiceContext 创建服务上下文，初始化所有数据库模型、Redis 和加密组件
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	// 初始化全局 AES 加密
	if c.AesKey != "" {
		if err := crypto.InitAES(c.AesKey); err != nil {
			panic("init AES failed: " + err.Error())
		}
	}

	// 初始化 Redis
	var rds *redis.Redis
	if c.Cache.Host != "" {
		rds = c.Cache.NewRedis()
	}

	return &ServiceContext{
		Config:                     c,
		Redis:                      rds,
		UserBaseModel:              model.NewUserBaseModel(conn),
		UserCommunityMembershipModel: model.NewUserCommunityMembershipModel(conn),
		UserMembershipRoleModel:    model.NewUserMembershipRoleModel(conn),
		UserCertificationModel:     model.NewUserCertificationModel(conn),
		UserResidenceModel:         model.NewUserResidenceModel(conn),
	}
}
