package manager

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/internal/adapter"
	"github.com/guxiao/community-and-home/services/ai-model/rpc/model"
)

type ModelManager struct {
	db              sqlx.SqlConn
	modelConfigModel model.AmModelConfigModel
	apiKeyModel     model.AmApiKeyModel
	healthCheckModel model.AmHealthCheckModel

	adapters        map[int64]adapter.ModelAdapter
	adaptersMutex   sync.RWMutex

	encryptionKey   []byte
}

func NewModelManager(db sqlx.SqlConn, cacheConf cache.CacheConf, encryptionKey string) *ModelManager {
	key := []byte(encryptionKey)
	if len(key) != 32 {
		key = make([]byte, 32)
		copy(key, encryptionKey)
	}

	return &ModelManager{
		db:              db,
		modelConfigModel: model.NewAmModelConfigModel(db, cacheConf),
		apiKeyModel:     model.NewAmApiKeyModel(db, cacheConf),
		healthCheckModel: model.NewAmHealthCheckModel(db, cacheConf),
		adapters:        make(map[int64]adapter.ModelAdapter),
		encryptionKey:   key,
	}
}

func (m *ModelManager) GetAdapter(ctx context.Context, modelID int64) (adapter.ModelAdapter, error) {
	m.adaptersMutex.RLock()
	if adp, exists := m.adapters[modelID]; exists {
		m.adaptersMutex.RUnlock()
		return adp, nil
	}
	m.adaptersMutex.RUnlock()

	config, err := m.modelConfigModel.FindOne(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("find model config: %w", err)
	}

	if config.Status != 1 {
		return nil, fmt.Errorf("model is not enabled")
	}

	apiKey, err := m.getAPIKey(ctx, config.Provider)
	if err != nil {
		return nil, fmt.Errorf("get API key: %w", err)
	}

	endpoint := ""
	if config.Endpoint.Valid {
		endpoint = config.Endpoint.String
	}

	adapterConfig := &adapter.ModelConfig{
		Provider:    config.Provider,
		ModelName:   config.ModelName,
		APIEndpoint: endpoint,
		APIKey:      apiKey,
		MaxTokens:   int32(4096),
		Timeout:     time.Duration(config.Timeout) * time.Millisecond,
		RetryCount:  int32(config.MaxRetries),
	}

	var adp adapter.ModelAdapter
	switch config.Provider {
	case "claude":
		adp = adapter.NewClaudeAdapter(adapterConfig)
	case "openai":
		adp = adapter.NewOpenAIAdapter(adapterConfig)
	case "ollama":
		adp = adapter.NewOllamaAdapter(adapterConfig)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	m.adaptersMutex.Lock()
	m.adapters[modelID] = adp
	m.adaptersMutex.Unlock()

	return adp, nil
}

func (m *ModelManager) GetAvailableModels(ctx context.Context, provider string) ([]*model.AmModelConfig, error) {
	query := "SELECT * FROM am_model_config WHERE status = 1"
	args := []interface{}{}

	if provider != "" {
		query += " AND provider = ?"
		args = append(args, provider)
	}

	query += " ORDER BY priority ASC, id ASC"

	var configs []*model.AmModelConfig
	err := m.db.QueryRowsCtx(ctx, &configs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query models: %w", err)
	}

	return configs, nil
}

func (m *ModelManager) RecordHealthCheck(ctx context.Context, modelID int64, status string, latency int64, errorMsg string) error {
	healthCheck := &model.AmHealthCheck{
		ModelId:   modelID,
		Status:    status,
		LatencyMs: sql.NullInt64{Int64: latency, Valid: true},
		ErrorMsg:  sql.NullString{String: errorMsg, Valid: errorMsg != ""},
	}

	_, err := m.healthCheckModel.Insert(ctx, healthCheck)
	if err != nil {
		return fmt.Errorf("insert health check: %w", err)
	}

	updateQuery := "UPDATE am_model_config SET health_status = ?, last_health_check = ? WHERE id = ?"
	_, err = m.db.ExecCtx(ctx, updateQuery, status, time.Now(), modelID)
	if err != nil {
		return fmt.Errorf("update model health status: %w", err)
	}

	return nil
}

func (m *ModelManager) PerformHealthCheck(ctx context.Context, modelID int64) error {
	adp, err := m.GetAdapter(ctx, modelID)
	if err != nil {
		m.RecordHealthCheck(ctx, modelID, "unhealthy", 0, err.Error())
		return err
	}

	startTime := time.Now()
	err = adp.HealthCheck(ctx)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		m.RecordHealthCheck(ctx, modelID, "unhealthy", latency, err.Error())
		return err
	}

	m.RecordHealthCheck(ctx, modelID, "healthy", latency, "")
	return nil
}

func (m *ModelManager) StartHealthCheckScheduler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runHealthChecks(ctx)
		}
	}
}

func (m *ModelManager) runHealthChecks(ctx context.Context) {
	models, err := m.GetAvailableModels(ctx, "")
	if err != nil {
		logx.Errorf("failed to get available models: %v", err)
		return
	}

	for _, model := range models {
		go func(modelID int64) {
			if err := m.PerformHealthCheck(ctx, modelID); err != nil {
				logx.Errorf("health check failed for model %d: %v", modelID, err)
			}
		}(model.Id)
	}
}

func (m *ModelManager) getAPIKey(ctx context.Context, provider string) (string, error) {
	query := "SELECT * FROM am_api_key WHERE provider = ? AND status = 1 ORDER BY priority ASC, last_used_time ASC LIMIT 1"

	var apiKey model.AmApiKey
	err := m.db.QueryRowCtx(ctx, &apiKey, query, provider)
	if err != nil {
		return "", fmt.Errorf("query API key: %w", err)
	}

	decryptedKey, err := m.decryptAPIKey(apiKey.ApiKey)
	if err != nil {
		return "", fmt.Errorf("decrypt API key: %w", err)
	}

	updateQuery := "UPDATE am_api_key SET last_used_time = ? WHERE id = ?"
	_, err = m.db.ExecCtx(ctx, updateQuery, time.Now(), apiKey.Id)
	if err != nil {
		logx.Errorf("failed to update API key usage: %v", err)
	}

	return decryptedKey, nil
}

func (m *ModelManager) encryptAPIKey(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (m *ModelManager) decryptAPIKey(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (m *ModelManager) InvalidateCache(modelID int64) {
	m.adaptersMutex.Lock()
	delete(m.adapters, modelID)
	m.adaptersMutex.Unlock()
}

func (m *ModelManager) GetModelConfig(ctx context.Context, modelID int64) (*model.AmModelConfig, error) {
	config, err := m.modelConfigModel.FindOne(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("find model config: %w", err)
	}
	return config, nil
}

func (m *ModelManager) CreateModelConfig(ctx context.Context, config *model.AmModelConfig) (sql.Result, error) {
	result, err := m.modelConfigModel.Insert(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("insert model config: %w", err)
	}
	return result, nil
}

func (m *ModelManager) UpdateModelConfig(ctx context.Context, config *model.AmModelConfig) error {
	err := m.modelConfigModel.Update(ctx, config)
	if err != nil {
		return fmt.Errorf("update model config: %w", err)
	}

	// 清除缓存的adapter，下次调用时会重新创建
	m.InvalidateCache(config.Id)

	return nil
}

func (m *ModelManager) DeleteModelConfig(ctx context.Context, modelID int64) error {
	err := m.modelConfigModel.Delete(ctx, modelID)
	if err != nil {
		return fmt.Errorf("delete model config: %w", err)
	}

	// 清除缓存的adapter
	m.InvalidateCache(modelID)

	return nil
}
