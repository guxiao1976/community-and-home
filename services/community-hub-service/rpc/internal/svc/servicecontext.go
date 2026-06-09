package svc

import (
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/config"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ServiceContext 社区枢纽 RPC 服务上下文
type ServiceContext struct {
	Config       config.Config
	SysConfig    *sysconfig.Client
	NoticeModel            model.NoticeModel
	NoticeAttachmentModel  model.NoticeAttachmentModel
	CommunityContactModel  model.CommunityContactModel
	LostFoundItemModel     model.LostFoundItemModel
}

// NewServiceContext 创建服务上下文，初始化所有数据库模型
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	// 初始化系统参数配置客户端
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", nil)

	return &ServiceContext{
		Config:                c,
		SysConfig:             sysCfg,
		NoticeModel:           model.NewNoticeModel(conn),
		NoticeAttachmentModel: model.NewNoticeAttachmentModel(conn),
		CommunityContactModel: model.NewCommunityContactModel(conn),
		LostFoundItemModel:    model.NewLostFoundItemModel(conn),
	}
}
