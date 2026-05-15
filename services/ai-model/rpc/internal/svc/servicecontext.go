package svc

import (
	"community-and-home/services/ai-model/rpc/internal/config"
	"community-and-home/services/ai-model/rpc/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config           config.Config
	ModelConfigModel model.AmModelConfigModel
	ApiKeyModel      model.AmApiKeyModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)

	// 构造 cache 配置
	logx.Infof("CacheRedis config length: %d", len(c.CacheRedis))
	var cacheConf cache.CacheConf
	if len(c.CacheRedis) > 0 {
		cacheConf = make(cache.CacheConf, len(c.CacheRedis))
		for i, redisConf := range c.CacheRedis {
			logx.Infof("Redis config[%d]: Host=%s, Type=%s, Pass=%s", i, redisConf.Host, redisConf.Type, redisConf.Pass)
			cacheConf[i] = cache.NodeConf{
				RedisConf: redisConf,
				Weight:    100,
			}
		}
		logx.Infof("Cache config created with %d nodes", len(cacheConf))
	} else {
		logx.Info("No cache redis configured, using nil cache")
	}

	return &ServiceContext{
		Config:           c,
		ModelConfigModel: model.NewAmModelConfigModel(conn, cacheConf),
		ApiKeyModel:      model.NewAmApiKeyModel(conn, cacheConf),
	}
}
