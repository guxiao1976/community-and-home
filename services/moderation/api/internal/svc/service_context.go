package svc

import (
	"github.com/guxiao/community-and-home/services/moderation/api/internal/config"
	"github.com/guxiao/community-and-home/services/moderation/api/internal/middleware"
	"github.com/guxiao/community-and-home/services/moderation/internal/auditlog"
	"github.com/guxiao/community-and-home/services/moderation/internal/engine"
	"github.com/guxiao/community-and-home/services/moderation/internal/llm"
	"github.com/guxiao/community-and-home/services/moderation/internal/wordstore"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config         config.Config
	AuthMiddleware rest.Middleware
	WordStore      *wordstore.WordStore
	TextEngine     *engine.TextEngine
	ImageEngine    *engine.ImageEngine
	AuditLogger    *auditlog.AuditLogger
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.DataSource)
	logDB := sqlx.NewMysql(c.LogDataSource)

	// Word store: loads sensitive words from masterdata_db
	ws := wordstore.NewWordStore(db,
		wordstore.WithSyncInterval(c.ACEngine.SyncInterval),
		wordstore.WithPinyinExpand(c.ACEngine.PinyinExpand, c.ACEngine.PinyinMaxVariants),
		wordstore.WithSplitSeparators(c.ACEngine.SplitSeparators),
	)

	// Build normalizer
	normalizer := ws.GetNormalizer()

	// Build split detector (needed by text engine)
	splitDetector := ws.GetSplitDetector()

	// LLM clients (empty implementations for now)
	var smallModel, largeModel llm.LLMClient
	if c.SmallModel.Enable {
		smallModel = llm.NewOllamaClient()
	}
	if c.LargeModel.Enable {
		largeModel = llm.NewRemoteLLMClient(c.LargeModel.Endpoint)
	}

	// Text engine
	textEngine := engine.NewTextEngine(
		normalizer,
		ws.GetACMachine(),
		ws.GetWhitelist(),
		splitDetector,
		smallModel,
		largeModel,
		c.SmallModel.HighConfThreshold,
	)

	// Image engine (image hash not yet active)
	var imageEngine *engine.ImageEngine
	if c.ImageHash.Enable {
		// Image hasher will be initialized when image hash module is fully implemented
		imageEngine = engine.NewImageEngine(nil, smallModel, largeModel)
	}

	// Audit logger
	auditLogger := auditlog.NewAuditLogger(logDB)

	return &ServiceContext{
		Config:         c,
		AuthMiddleware: middleware.NewAuthMiddleware().Handle,
		WordStore:      ws,
		TextEngine:     textEngine,
		ImageEngine:    imageEngine,
		AuditLogger:    auditLogger,
	}
}
