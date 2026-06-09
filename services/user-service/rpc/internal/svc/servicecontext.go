package svc

import (
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ServiceContext 用户中心 RPC 服务上下文
type ServiceContext struct {
	Config config.Config

	Redis     *redis.Redis       // Redis 缓存客户端
	SysConfig *sysconfig.Client // 系统参数配置读取器

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

	// 初始化系统参数配置客户端（无 gRPC fallback，需单独配置）
	// 若要启用 Redis 未命中时回退到 master-data-service，可传入带 GetConfig RPC 的 FallbackFunc
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", nil)

	return &ServiceContext{
		Config:                     c,
		Redis:                      rds,
		SysConfig:                  sysCfg,
		UserBaseModel:              model.NewUserBaseModel(conn),
		UserCommunityMembershipModel: model.NewUserCommunityMembershipModel(conn),
		UserMembershipRoleModel:    model.NewUserMembershipRoleModel(conn),
		UserCertificationModel:     model.NewUserCertificationModel(conn),
		UserResidenceModel:         model.NewUserResidenceModel(conn),
	}
}
