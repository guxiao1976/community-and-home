package svc

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"community-and-home/services/ai-model/rpc/internal/config"
	"community-and-home/services/ai-model/rpc/internal/manager"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	ModelManager    *manager.ModelManager
	CostManager     *manager.CostManager
	TemplateManager *manager.TemplateManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.DataSource)

	// 构造 cache 配置
	var cacheConf cache.CacheConf
	if len(c.CacheRedis) > 0 {
		cacheConf = make(cache.CacheConf, len(c.CacheRedis))
		for i, redisConf := range c.CacheRedis {
			cacheConf[i] = cache.NodeConf{
				RedisConf: redisConf,
				Weight:    100, // 默认权重
			}
		}
	}

	return &ServiceContext{
		Config:          c,
		DB:              db,
		ModelManager:    manager.NewModelManager(db, cacheConf, c.EncryptionKey),
		CostManager:     manager.NewCostManager(db, cacheConf),
		TemplateManager: manager.NewTemplateManager(db, cacheConf),
	}
}
