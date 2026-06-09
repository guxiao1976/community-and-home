package svc

import (
	"context"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/config"
	sysconfig "github.com/guxiao1976/community-common/v2/pkg/sysconfig"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
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
	sysCfg := sysconfig.MustInit(c.SysConfigRedis, "", func(ctx context.Context, key string) (*sysconfig.ConfigValue, error) {
		// gRPC fallback to master-data GetConfig
		conn, err := zrpc.NewClient(c.MasterDataRpc)
		if err != nil {
			return nil, err
		}
		client := masterdatav1.NewMasterdataServiceClient(conn.Conn())
		resp, err := client.GetConfig(ctx, &masterdatav1.GetConfigReq{ConfigKey: key})
		if err != nil {
			return nil, err
		}
		return &sysconfig.ConfigValue{
			Value: resp.ConfigValue,
			Type:  resp.ValueType,
			Desc:  resp.Description,
		}, nil
	})

	return &ServiceContext{
		Config:                c,
		SysConfig:             sysCfg,
		NoticeModel:           model.NewNoticeModel(conn),
		NoticeAttachmentModel: model.NewNoticeAttachmentModel(conn),
		CommunityContactModel: model.NewCommunityContactModel(conn),
		LostFoundItemModel:    model.NewLostFoundItemModel(conn),
	}
}
