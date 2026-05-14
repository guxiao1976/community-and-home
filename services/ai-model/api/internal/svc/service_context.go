package svc

import (
	"community-and-home/services/ai-model/api/internal/config"
	"community-and-home/services/ai-model/rpc/aimodel"
	"community-and-home/services/ai-model/rpc/model"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	AiModelRpc aimodel.AiModel

	// 数据库连接
	DB sqlx.SqlConn

	// Model 层
	ApiKeyModel          model.AmApiKeyModel
	PromptTemplateModel  model.AmPromptTemplateModel
	CallLogModel         model.AmCallLogModel
	CostAlertConfigModel model.AmCostAlertConfigModel
	AlertRecordModel     model.AmAlertRecordModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	// 构造 cache 配置
	var cacheConf cache.CacheConf
	if len(c.CacheRedis) > 0 {
		cacheConf = make(cache.CacheConf, len(c.CacheRedis))
		for i, redisConf := range c.CacheRedis {
			cacheConf[i] = cache.NodeConf{
				RedisConf: redisConf,
				Weight:    100,
			}
		}
	}

	return &ServiceContext{
		Config:     c,
		AiModelRpc: aimodel.NewAiModel(zrpc.MustNewClient(c.AiModelRpc)),

		// 数据库连接
		DB: conn,

		// 初始化 Model 层
		ApiKeyModel:          model.NewAmApiKeyModel(conn, cacheConf),
		PromptTemplateModel:  model.NewAmPromptTemplateModel(conn, cacheConf),
		CallLogModel:         model.NewAmCallLogModel(conn, cacheConf),
		CostAlertConfigModel: model.NewAmCostAlertConfigModel(conn, cacheConf),
		AlertRecordModel:     model.NewAmAlertRecordModel(conn, cacheConf),
	}
}
