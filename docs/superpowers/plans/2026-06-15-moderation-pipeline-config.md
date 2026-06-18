# 审核管线可配置化 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 AC 引擎→小模型→大模型级联审核管线改造为可配置化系统，支持选择提示词模板、编辑升级规则、即时测试，测试通过的配置可设为生产环境默认管线。

**Architecture:** 后端 moderation-service 新增 `mod_pipeline_config` 表 + CRUD 端点 + pipeline/test 执行端点；前端 Vue 3 改造 ModerationTest.vue 为配置测试工作台。复用 ai-model-service 已有模板/模型查询端点。

**Tech Stack:** Go (go-zero), MySQL, Vue 3 + Element Plus + TypeScript

---

## 文件结构总览

### 后端 (moderation-service)

```
services/moderation-service/
  migrations/
    002_pipeline_config.sql                    [CREATE]  迁移 SQL
  model/
    mod_pipeline_config_gen.go                 [CREATE]  goctl 风格数据模型
    mod_pipeline_config_model.go               [CREATE]  自定义辅助方法
  api/internal/
    config/config.go                           [MODIFY]  新增 PipelineDB 配置
    types/types.go                             [MODIFY]  新增 Pipeline 相关类型
    svc/service_context.go                     [MODIFY]  注入 PipelineModel + PipelineExecutor
    handler/
      routes.go                                [MODIFY]  注册 7 个新路由
      pipeline/
        create_pipeline_handler.go             [CREATE]
        update_pipeline_handler.go             [CREATE]
        delete_pipeline_handler.go             [CREATE]
        get_pipeline_handler.go                [CREATE]
        list_pipelines_handler.go              [CREATE]
        activate_pipeline_handler.go           [CREATE]
        pipeline_test_handler.go               [CREATE]
    logic/
      pipeline/
        create_pipeline_logic.go               [CREATE]
        update_pipeline_logic.go               [CREATE]
        delete_pipeline_logic.go               [CREATE]
        get_pipeline_logic.go                  [CREATE]
        list_pipelines_logic.go                [CREATE]
        activate_pipeline_logic.go             [CREATE]
        pipeline_test_logic.go                 [CREATE]
  internal/
    pipeline/
      config.go                                [CREATE]  PipelineConfig 类型 + 加载
      executor.go                              [CREATE]  PipelineExecutor 执行引擎
```

### 前端 (web/pc)

```
web/
  common/types/
    moderation.d.ts                            [MODIFY]  新增 Pipeline 类型
  pc/src/
    api/
      moderation.ts                            [MODIFY]  新增 pipeline API
      aimodel.ts                               [MODIFY]  新增模板/模型列表查询
    config/modules/
      moderation.config.ts                     [MODIFY]  路由+菜单更新
    views/moderation/
      ModerationConfigTest.vue                 [CREATE]  页面容器（替代 ModerationTest.vue）
    components/moderation/
      PipelineSelector.vue                     [CREATE]  管线选择器
      LayerConfigPanel.vue                     [CREATE]  三层配置面板
      EscalationRuleEditor.vue                 [CREATE]  升级规则编辑器
      TestInputArea.vue                        [CREATE]  测试输入区
      PipelineResultPanel.vue                  [CREATE]  结果展示面板
```

---

### Task 1: 数据库迁移

**Files:**
- Create: `services/moderation-service/migrations/002_pipeline_config.sql`

- [ ] **Step 1: 编写迁移 SQL**

```sql
-- Migration 002: Create mod_pipeline_config table for configurable moderation pipelines
-- Database: moderation_db

USE moderation_db;

CREATE TABLE IF NOT EXISTS mod_pipeline_config (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pipeline_id     VARCHAR(100) NOT NULL COMMENT '唯一标识，如 default_v1, strict_mode',
    pipeline_name   VARCHAR(200) NOT NULL COMMENT '显示名称，如"默认审核管线"',
    description     TEXT         COMMENT '描述',
    is_active       TINYINT      NOT NULL DEFAULT 1 COMMENT '1=启用 0=软删除',
    is_production   TINYINT      NOT NULL DEFAULT 0 COMMENT '1=生产环境使用',

    -- AC 引擎配置
    ac_enabled            TINYINT NOT NULL DEFAULT 1,
    ac_severity_threshold INT     NOT NULL DEFAULT 1 COMMENT '最低触发严重度(1高/2中/3低)',

    -- 小模型配置
    small_model_template_id VARCHAR(100) COMMENT 'FK→ai_model_db.am_prompt_template.template_id',
    small_model_config_key  VARCHAR(64)  COMMENT '可选，覆盖模板默认的 config_key',

    -- 大模型配置
    large_model_template_id VARCHAR(100) COMMENT 'FK→ai_model_db.am_prompt_template.template_id',
    large_model_config_key  VARCHAR(64)  COMMENT '可选，覆盖模板默认的 config_key',

    -- 升级规则（AC → 小模型）
    ac_to_small_condition  VARCHAR(50)  NOT NULL DEFAULT 'any_hit' COMMENT 'any_hit|severity_gte|category_in|never',
    ac_to_small_severity   INT          COMMENT 'severity_gte 时的阈值',
    ac_to_small_categories JSON         COMMENT 'category_in 时的分类列表',

    -- 升级规则（小模型 → 大模型）
    small_to_large_condition            VARCHAR(50)  NOT NULL DEFAULT 'confidence_lt' COMMENT 'confidence_lt|category_in|always|never',
    small_to_large_confidence_threshold DECIMAL(3,2) DEFAULT 0.90,
    small_to_large_categories           JSON         COMMENT 'category_in 时的分类列表',

    -- 终判逻辑
    final_verdict_logic VARCHAR(50) NOT NULL DEFAULT 'last_model_wins' COMMENT 'last_model_wins|large_overrides|small_overrides',

    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time  TIMESTAMP NULL DEFAULT NULL,

    UNIQUE INDEX idx_pipeline_id (pipeline_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审核管线配置';
```

- [ ] **Step 2: 执行迁移**

```bash
docker exec -i mysql mysql -uroot -proot123456 moderation_db < services/moderation-service/migrations/002_pipeline_config.sql
```

- [ ] **Step 3: 验证**

```bash
docker exec -i mysql mysql -uroot -proot123456 moderation_db -e "DESC mod_pipeline_config;"
```

---

### Task 2: Go 数据模型 — ModPipelineConfig

**Files:**
- Create: `services/moderation-service/model/mod_pipeline_config_gen.go`
- Create: `services/moderation-service/model/mod_pipeline_config_model.go`

- [ ] **Step 1: 编写生成模型 `mod_pipeline_config_gen.go`**

```go
package model

import (
	"database/sql"
	"time"
)

// ModPipelineConfig 审核管线配置
type ModPipelineConfig struct {
	Id                           uint64         `db:"id"`
	PipelineId                   string         `db:"pipeline_id"`
	PipelineName                 string         `db:"pipeline_name"`
	Description                  sql.NullString `db:"description"`
	IsActive                     int64          `db:"is_active"`
	IsProduction                 int64          `db:"is_production"`
	AcEnabled                    int64          `db:"ac_enabled"`
	AcSeverityThreshold          int64          `db:"ac_severity_threshold"`
	SmallModelTemplateId         sql.NullString `db:"small_model_template_id"`
	SmallModelConfigKey          sql.NullString `db:"small_model_config_key"`
	LargeModelTemplateId         sql.NullString `db:"large_model_template_id"`
	LargeModelConfigKey          sql.NullString `db:"large_model_config_key"`
	AcToSmallCondition           string         `db:"ac_to_small_condition"`
	AcToSmallSeverity            sql.NullInt64  `db:"ac_to_small_severity"`
	AcToSmallCategories          sql.NullString `db:"ac_to_small_categories"`
	SmallToLargeCondition        string         `db:"small_to_large_condition"`
	SmallToLargeConfidenceThresh sql.NullFloat64 `db:"small_to_large_confidence_threshold"`
	SmallToLargeCategories       sql.NullString `db:"small_to_large_categories"`
	FinalVerdictLogic            string         `db:"final_verdict_logic"`
	CreatedTime                  time.Time      `db:"created_time"`
	UpdatedTime                  time.Time      `db:"updated_time"`
	DeleteTime                   sql.NullTime   `db:"delete_time"`
}

// ModPipelineConfigModel 接口 — 遵循项目 goctl 模型模式
type ModPipelineConfigModel interface {
	Insert(ctx context.Context, data *ModPipelineConfig) (sql.Result, error)
	FindOne(ctx context.Context, id uint64) (*ModPipelineConfig, error)
	FindOneByPipelineId(ctx context.Context, pipelineId string) (*ModPipelineConfig, error)
	FindList(ctx context.Context, page, pageSize int) ([]*ModPipelineConfig, error)
	FindProduction(ctx context.Context) (*ModPipelineConfig, error)
	Update(ctx context.Context, data *ModPipelineConfig) error
	Delete(ctx context.Context, id uint64) error
	Count(ctx context.Context) (int64, error)
}

// defaultModPipelineConfigModel 默认实现（基于 sqlx）
type defaultModPipelineConfigModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewModPipelineConfigModel(conn sqlx.SqlConn) ModPipelineConfigModel {
	return &defaultModPipelineConfigModel{
		conn:  conn,
		table: "mod_pipeline_config",
	}
}

func (m *defaultModPipelineConfigModel) Insert(ctx context.Context, data *ModPipelineConfig) (sql.Result, error) {
	query := `INSERT INTO mod_pipeline_config (
		pipeline_id, pipeline_name, description, is_active, is_production,
		ac_enabled, ac_severity_threshold,
		small_model_template_id, small_model_config_key,
		large_model_template_id, large_model_config_key,
		ac_to_small_condition, ac_to_small_severity, ac_to_small_categories,
		small_to_large_condition, small_to_large_confidence_threshold, small_to_large_categories,
		final_verdict_logic
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query,
		data.PipelineId, data.PipelineName, data.Description, data.IsActive, data.IsProduction,
		data.AcEnabled, data.AcSeverityThreshold,
		data.SmallModelTemplateId, data.SmallModelConfigKey,
		data.LargeModelTemplateId, data.LargeModelConfigKey,
		data.AcToSmallCondition, data.AcToSmallSeverity, data.AcToSmallCategories,
		data.SmallToLargeCondition, data.SmallToLargeConfidenceThresh, data.SmallToLargeCategories,
		data.FinalVerdictLogic,
	)
}

func (m *defaultModPipelineConfigModel) FindOne(ctx context.Context, id uint64) (*ModPipelineConfig, error) {
	query := `SELECT * FROM mod_pipeline_config WHERE id = ? AND delete_time IS NULL LIMIT 1`
	var resp ModPipelineConfig
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultModPipelineConfigModel) FindOneByPipelineId(ctx context.Context, pipelineId string) (*ModPipelineConfig, error) {
	query := `SELECT * FROM mod_pipeline_config WHERE pipeline_id = ? AND delete_time IS NULL LIMIT 1`
	var resp ModPipelineConfig
	err := m.conn.QueryRowCtx(ctx, &resp, query, pipelineId)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultModPipelineConfigModel) FindList(ctx context.Context, page, pageSize int) ([]*ModPipelineConfig, error) {
	offset := (page - 1) * pageSize
	query := `SELECT * FROM mod_pipeline_config WHERE delete_time IS NULL ORDER BY updated_time DESC LIMIT ?, ?`
	var resp []*ModPipelineConfig
	err := m.conn.QueryRowsCtx(ctx, &resp, query, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultModPipelineConfigModel) FindProduction(ctx context.Context) (*ModPipelineConfig, error) {
	query := `SELECT * FROM mod_pipeline_config WHERE is_production = 1 AND is_active = 1 AND delete_time IS NULL LIMIT 1`
	var resp ModPipelineConfig
	err := m.conn.QueryRowCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultModPipelineConfigModel) Update(ctx context.Context, data *ModPipelineConfig) error {
	query := `UPDATE mod_pipeline_config SET
		pipeline_name = ?, description = ?, is_active = ?, is_production = ?,
		ac_enabled = ?, ac_severity_threshold = ?,
		small_model_template_id = ?, small_model_config_key = ?,
		large_model_template_id = ?, large_model_config_key = ?,
		ac_to_small_condition = ?, ac_to_small_severity = ?, ac_to_small_categories = ?,
		small_to_large_condition = ?, small_to_large_confidence_threshold = ?, small_to_large_categories = ?,
		final_verdict_logic = ?
		WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query,
		data.PipelineName, data.Description, data.IsActive, data.IsProduction,
		data.AcEnabled, data.AcSeverityThreshold,
		data.SmallModelTemplateId, data.SmallModelConfigKey,
		data.LargeModelTemplateId, data.LargeModelConfigKey,
		data.AcToSmallCondition, data.AcToSmallSeverity, data.AcToSmallCategories,
		data.SmallToLargeCondition, data.SmallToLargeConfidenceThresh, data.SmallToLargeCategories,
		data.FinalVerdictLogic,
		data.Id,
	)
	return err
}

func (m *defaultModPipelineConfigModel) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE mod_pipeline_config SET delete_time = NOW() WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

func (m *defaultModPipelineConfigModel) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM mod_pipeline_config WHERE delete_time IS NULL`
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}
```

- [ ] **Step 2: 编写辅助方法 `mod_pipeline_config_model.go`**

```go
package model

import "database/sql"

// NullStringToString safely converts sql.NullString to string
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// NullFloat64ToFloat64 safely converts sql.NullFloat64 to float64
func nullFloat64ToFloat64(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

// NullInt64ToInt64 safely converts sql.NullInt64 to int64
func nullInt64ToInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}
```

- [ ] **Step 3: 验证编译**

```bash
cd services/moderation-service && go build ./model/...
```

---

### Task 3: Config 更新 — 新增 PipelineDB 数据源

**Files:**
- Modify: `services/moderation-service/api/internal/config/config.go`

- [ ] **Step 1: 在 `api/internal/config/config.go` 的 Config struct 中确认 `LogDataSource` 字段可用于 moderation_db**

`LogDataSource` 已存在（连接 moderation_db）。Pipeline 配置也存储在 moderation_db，无需新增数据源字段。

- [ ] **Step 2: 验证编译**

```bash
cd services/moderation-service && go build ./...
```

---

### Task 4: API 类型定义 — Pipeline 请求/响应类型

**Files:**
- Modify: `services/moderation-service/api/internal/types/types.go`

- [ ] **Step 1: 在 types.go 末尾追加 Pipeline 相关类型**

```go
// ========== 管线配置管理类型 ==========

type CreatePipelineReq struct {
	PipelineId                 string   `json:"pipeline_id"`
	PipelineName               string   `json:"pipeline_name"`
	Description                string   `json:"description,optional"`
	AcEnabled                  int64    `json:"ac_enabled,optional,default=1"`
	AcSeverityThreshold        int64    `json:"ac_severity_threshold,optional,default=1"`
	SmallModelTemplateId       string   `json:"small_model_template_id,optional"`
	SmallModelConfigKey        string   `json:"small_model_config_key,optional"`
	LargeModelTemplateId       string   `json:"large_model_template_id,optional"`
	LargeModelConfigKey        string   `json:"large_model_config_key,optional"`
	AcToSmallCondition         string   `json:"ac_to_small_condition,optional,default=any_hit"`
	AcToSmallSeverity          int64    `json:"ac_to_small_severity,optional"`
	AcToSmallCategories        []string `json:"ac_to_small_categories,optional"`
	SmallToLargeCondition      string   `json:"small_to_large_condition,optional,default=confidence_lt"`
	SmallToLargeConfidenceThresh float64 `json:"small_to_large_confidence_threshold,optional,default=0.90"`
	SmallToLargeCategories     []string `json:"small_to_large_categories,optional"`
	FinalVerdictLogic          string   `json:"final_verdict_logic,optional,default=last_model_wins"`
}

type CreatePipelineResp struct {
	Id         uint64 `json:"id"`
	PipelineId string `json:"pipeline_id"`
}

type UpdatePipelineReq struct {
	Id                         uint64   `json:"id"`
	PipelineName               string   `json:"pipeline_name,optional"`
	Description                string   `json:"description,optional"`
	IsActive                   int64    `json:"is_active,optional"`
	IsProduction               int64    `json:"is_production,optional"`
	AcEnabled                  int64    `json:"ac_enabled,optional"`
	AcSeverityThreshold        int64    `json:"ac_severity_threshold,optional"`
	SmallModelTemplateId       string   `json:"small_model_template_id,optional"`
	SmallModelConfigKey        string   `json:"small_model_config_key,optional"`
	LargeModelTemplateId       string   `json:"large_model_template_id,optional"`
	LargeModelConfigKey        string   `json:"large_model_config_key,optional"`
	AcToSmallCondition         string   `json:"ac_to_small_condition,optional"`
	AcToSmallSeverity          int64    `json:"ac_to_small_severity,optional"`
	AcToSmallCategories        []string `json:"ac_to_small_categories,optional"`
	SmallToLargeCondition      string   `json:"small_to_large_condition,optional"`
	SmallToLargeConfidenceThresh float64 `json:"small_to_large_confidence_threshold,optional"`
	SmallToLargeCategories     []string `json:"small_to_large_categories,optional"`
	FinalVerdictLogic          string   `json:"final_verdict_logic,optional"`
}

type DeletePipelineReq struct {
	PipelineId string `json:"pipeline_id"`
}

type GetPipelineReq struct {
	PipelineId string `json:"pipeline_id"`
}

type ListPipelinesReq struct {
	Page     int `json:"page,optional,default=1"`
	PageSize int `json:"page_size,optional,default=10"`
}

type PipelineInfo struct {
	Id                           uint64   `json:"id"`
	PipelineId                   string   `json:"pipeline_id"`
	PipelineName                 string   `json:"pipeline_name"`
	Description                  string   `json:"description"`
	IsActive                     int64    `json:"is_active"`
	IsProduction                 int64    `json:"is_production"`
	AcEnabled                    int64    `json:"ac_enabled"`
	AcSeverityThreshold          int64    `json:"ac_severity_threshold"`
	SmallModelTemplateId         string   `json:"small_model_template_id"`
	SmallModelConfigKey          string   `json:"small_model_config_key"`
	LargeModelTemplateId         string   `json:"large_model_template_id"`
	LargeModelConfigKey          string   `json:"large_model_config_key"`
	AcToSmallCondition           string   `json:"ac_to_small_condition"`
	AcToSmallSeverity            int64    `json:"ac_to_small_severity"`
	AcToSmallCategories          []string `json:"ac_to_small_categories"`
	SmallToLargeCondition        string   `json:"small_to_large_condition"`
	SmallToLargeConfidenceThresh float64  `json:"small_to_large_confidence_threshold"`
	SmallToLargeCategories       []string `json:"small_to_large_categories"`
	FinalVerdictLogic            string   `json:"final_verdict_logic"`
	CreatedTime                  string   `json:"created_time"`
	UpdatedTime                  string   `json:"updated_time"`
}

type ListPipelinesResp struct {
	List     []PipelineInfo `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type ActivatePipelineReq struct {
	PipelineId string `json:"pipeline_id"`
}

// ========== 管线测试类型 ==========

type PipelineTestReq struct {
	Content                 string  `json:"content"`
	PipelineId              string  `json:"pipeline_id,optional"`
	SmallModelTemplateId    string  `json:"small_model_template_id,optional"`
	LargeModelTemplateId    string  `json:"large_model_template_id,optional"`
	SmallToLargeConfidence  float64 `json:"small_to_large_confidence,optional"`
}

type LayerResult struct {
	Called        bool     `json:"called"`
	SkippedReason string   `json:"skipped_reason,omitempty"`
	Passed        bool     `json:"passed"`
	RiskLevel     string   `json:"risk_level"`
	Confidence    float64  `json:"confidence"`
	Categories    []string `json:"categories,omitempty"`
	Reason        string   `json:"reason,omitempty"`
	LatencyMs     int64    `json:"latency_ms"`
	MatchedWords  []string `json:"matched_words,omitempty"`
	ModelUsed     string   `json:"model_used,omitempty"`
	TemplateID    string   `json:"template_id,omitempty"`
	RawResponse   string   `json:"raw_response,omitempty"`
}

type PipelineTestResp struct {
	PipelineId       string       `json:"pipeline_id"`
	AcResult         *LayerResult `json:"ac_result"`
	SmallModelResult *LayerResult `json:"small_model_result"`
	LargeModelResult *LayerResult `json:"large_model_result"`
	FinalVerdict     string       `json:"final_verdict"`
	TotalLatencyMs   int64        `json:"total_latency_ms"`
}
```

- [ ] **Step 2: 验证编译**

```bash
cd services/moderation-service && go build ./api/internal/types/...
```

---

### Task 5: PipelineExecutor — 管线执行引擎

**Files:**
- Create: `services/moderation-service/internal/pipeline/config.go`
- Create: `services/moderation-service/internal/pipeline/executor.go`

- [ ] **Step 1: 编写 `config.go` — 管线配置运行时表示**

```go
package pipeline

import (
	"encoding/json"

	"community-moderation-service/model"
)

// PipelineConfig 管线配置的运行时表示（从 DB 模型解析）
type PipelineConfig struct {
	PipelineId                   string   `json:"pipeline_id"`
	PipelineName                 string   `json:"pipeline_name"`
	AcEnabled                    bool     `json:"ac_enabled"`
	AcSeverityThreshold          int      `json:"ac_severity_threshold"`
	SmallModelTemplateId         string   `json:"small_model_template_id"`
	SmallModelConfigKey          string   `json:"small_model_config_key"`
	LargeModelTemplateId         string   `json:"large_model_template_id"`
	LargeModelConfigKey          string   `json:"large_model_config_key"`
	AcToSmallCondition           string   `json:"ac_to_small_condition"`
	AcToSmallSeverity            int      `json:"ac_to_small_severity"`
	AcToSmallCategories          []string `json:"ac_to_small_categories"`
	SmallToLargeCondition        string   `json:"small_to_large_condition"`
	SmallToLargeConfidenceThresh float64  `json:"small_to_large_confidence_threshold"`
	SmallToLargeCategories       []string `json:"small_to_large_categories"`
	FinalVerdictLogic            string   `json:"final_verdict_logic"`
}

// FromDBModel 从 DB 模型转换为运行时配置
func FromDBModel(m *model.ModPipelineConfig) *PipelineConfig {
	var acCats, slCats []string
	if m.AcToSmallCategories.Valid {
		json.Unmarshal([]byte(m.AcToSmallCategories.String), &acCats)
	}
	if m.SmallToLargeCategories.Valid {
		json.Unmarshal([]byte(m.SmallToLargeCategories.String), &slCats)
	}

	smallConfThresh := 0.90
	if m.SmallToLargeConfidenceThresh.Valid {
		smallConfThresh = m.SmallToLargeConfidenceThresh.Float64
	}

	return &PipelineConfig{
		PipelineId:                   m.PipelineId,
		PipelineName:                 m.PipelineName,
		AcEnabled:                    m.AcEnabled == 1,
		AcSeverityThreshold:          int(m.AcSeverityThreshold),
		SmallModelTemplateId:         model.NullStringToString(m.SmallModelTemplateId),
		SmallModelConfigKey:          model.NullStringToString(m.SmallModelConfigKey),
		LargeModelTemplateId:         model.NullStringToString(m.LargeModelTemplateId),
		LargeModelConfigKey:          model.NullStringToString(m.LargeModelConfigKey),
		AcToSmallCondition:           m.AcToSmallCondition,
		AcToSmallSeverity:            int(model.NullInt64ToInt64(m.AcToSmallSeverity)),
		AcToSmallCategories:          acCats,
		SmallToLargeCondition:        m.SmallToLargeCondition,
		SmallToLargeConfidenceThresh: smallConfThresh,
		SmallToLargeCategories:       slCats,
		FinalVerdictLogic:            m.FinalVerdictLogic,
	}
}
```

- [ ] **Step 2: 编写 `executor.go` — 管线执行引擎**

```go
package pipeline

import (
	"context"
	"time"

	"community-moderation-service/internal/engine"
	"community-moderation-service/internal/llm"

	aimodelv1 "github.com/guxiao1976/api-proto/gen/go/aimodel/v1"
)

// Executor 管线执行引擎
type Executor struct {
	textEngine      *engine.TextEngine
	aiModelClient   aimodelv1.AiModelServiceClient
}

// NewExecutor 创建管线执行引擎
func NewExecutor(textEngine *engine.TextEngine, aiModelClient aimodelv1.AiModelServiceClient) *Executor {
	return &Executor{
		textEngine:    textEngine,
		aiModelClient: aiModelClient,
	}
}

// LayerResult 单层执行结果
type LayerResult struct {
	Called        bool     `json:"called"`
	SkippedReason string   `json:"skipped_reason,omitempty"`
	Passed        bool     `json:"passed"`
	RiskLevel     string   `json:"risk_level"`
	Confidence    float64  `json:"confidence"`
	Categories    []string `json:"categories,omitempty"`
	Reason        string   `json:"reason,omitempty"`
	LatencyMs     int64    `json:"latency_ms"`
	MatchedWords  []string `json:"matched_words,omitempty"`
	ModelUsed     string   `json:"model_used,omitempty"`
	TemplateID    string   `json:"template_id,omitempty"`
	RawResponse   string   `json:"raw_response,omitempty"`
}

// PipelineResult 管线完整执行结果
type PipelineResult struct {
	PipelineId       string       `json:"pipeline_id"`
	AcResult         *LayerResult `json:"ac_result"`
	SmallModelResult *LayerResult `json:"small_model_result"`
	LargeModelResult *LayerResult `json:"large_model_result"`
	FinalVerdict     string       `json:"final_verdict"`
	TotalLatencyMs   int64        `json:"total_latency_ms"`
}

// Execute 执行完整审核管线
func (e *Executor) Execute(ctx context.Context, config *PipelineConfig, content string) (*PipelineResult, error) {
	startTime := time.Now()
	result := &PipelineResult{PipelineId: config.PipelineId}

	// ------- 第一层：AC 引擎 -------
	var acHits []engine.MatchDetail
	if config.AcEnabled {
		acStart := time.Now()
		modResult, err := e.textEngine.Check(ctx, content, "post", "ac_only")
		acLatency := time.Since(acStart).Milliseconds()

		acLayer := &LayerResult{
			Called:    true,
			LatencyMs: acLatency,
		}
		if err != nil {
			acLayer.Passed = false
			acLayer.RiskLevel = "high"
			acLayer.Reason = "AC引擎执行异常: " + err.Error()
		} else {
			acHits = modResult.Details
			acLayer.Passed = modResult.Pass
			acLayer.RiskLevel = modResult.RiskLevel
			acLayer.Reason = modResult.Reason
			for _, d := range modResult.Details {
				if d.MatchedText != "" {
					acLayer.MatchedWords = append(acLayer.MatchedWords, d.MatchedText)
				}
			}
		}
		result.AcResult = acLayer

		// 判断是否升级到小模型
		if !e.shouldEscalateAcToSmall(config, acHits) {
			result.AcResult.SkippedReason = "" // AC已执行，不做跳过标记
			result.SmallModelResult = &LayerResult{
				Called: false, SkippedReason: "AC引擎未触发升级条件",
			}
			result.LargeModelResult = &LayerResult{
				Called: false, SkippedReason: "AC引擎未触发升级条件",
			}
			result.FinalVerdict = e.computeFinalVerdict(config, result)
			result.TotalLatencyMs = time.Since(startTime).Milliseconds()
			return result, nil
		}
	} else {
		result.AcResult = &LayerResult{Called: false, SkippedReason: "AC引擎已禁用"}
	}

	// ------- 第二层：小模型 -------
	if config.SmallModelTemplateId == "" {
		result.SmallModelResult = &LayerResult{Called: false, SkippedReason: "未配置小模型模板"}
		result.LargeModelResult = &LayerResult{Called: false, SkippedReason: "上层未完成"}
		result.FinalVerdict = e.computeFinalVerdict(config, result)
		result.TotalLatencyMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	smStart := time.Now()
	smResult, err := e.callModel(ctx, config.SmallModelTemplateId, config.SmallModelConfigKey, content)
	smLatency := time.Since(smStart).Milliseconds()

	smLayer := &LayerResult{
		Called:     true,
		LatencyMs:  smLatency,
		TemplateID: config.SmallModelTemplateId,
	}
	if err != nil {
		smLayer.Passed = false
		smLayer.RiskLevel = "medium"
		smLayer.Reason = "小模型调用异常: " + err.Error()
	} else {
		smLayer.Passed = smResult.IsSafe
		smLayer.Confidence = smResult.Confidence
		smLayer.Categories = smResult.Categories
		smLayer.Reason = smResult.Reason
		smLayer.ModelUsed = smResult.ModelUsed
		smLayer.RawResponse = smResult.RawResponse
		smLayer.RiskLevel = smResult.RiskLevel
	}
	result.SmallModelResult = smLayer

	// 判断是否升级到大模型
	if !e.shouldEscalateSmallToLarge(config, smLayer) {
		result.LargeModelResult = &LayerResult{
			Called: false, SkippedReason: "小模型置信度足够，无需升级",
		}
		result.FinalVerdict = e.computeFinalVerdict(config, result)
		result.TotalLatencyMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	// ------- 第三层：大模型 -------
	if config.LargeModelTemplateId == "" {
		result.LargeModelResult = &LayerResult{Called: false, SkippedReason: "未配置大模型模板"}
		result.FinalVerdict = e.computeFinalVerdict(config, result)
		result.TotalLatencyMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	lmStart := time.Now()
	lmResult, err := e.callModel(ctx, config.LargeModelTemplateId, config.LargeModelConfigKey, content)
	lmLatency := time.Since(lmStart).Milliseconds()

	lmLayer := &LayerResult{
		Called:     true,
		LatencyMs:  lmLatency,
		TemplateID: config.LargeModelTemplateId,
	}
	if err != nil {
		lmLayer.Passed = false
		lmLayer.RiskLevel = "medium"
		lmLayer.Reason = "大模型调用异常: " + err.Error()
	} else {
		lmLayer.Passed = lmResult.IsSafe
		lmLayer.Confidence = lmResult.Confidence
		lmLayer.Categories = lmResult.Categories
		lmLayer.Reason = lmResult.Reason
		lmLayer.ModelUsed = lmResult.ModelUsed
		lmLayer.RawResponse = lmResult.RawResponse
		lmLayer.RiskLevel = lmResult.RiskLevel
	}
	result.LargeModelResult = lmLayer

	result.FinalVerdict = e.computeFinalVerdict(config, result)
	result.TotalLatencyMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// shouldEscalateAcToSmall 判断 AC 引擎结果是否需要升级到小模型
func (e *Executor) shouldEscalateAcToSmall(config *PipelineConfig, hits []engine.MatchDetail) bool {
	switch config.AcToSmallCondition {
	case "never":
		return false
	case "any_hit":
		return len(hits) > 0
	case "severity_gte":
		for _, h := range hits {
			if h.Severity >= config.AcToSmallSeverity {
				return true
			}
		}
		return false
	case "category_in":
		for _, h := range hits {
			for _, cat := range config.AcToSmallCategories {
				if h.Category == cat {
					return true
				}
			}
		}
		return false
	default:
		return len(hits) > 0 // 默认 any_hit
	}
}

// shouldEscalateSmallToLarge 判断小模型结果是否需要升级到大模型
func (e *Executor) shouldEscalateSmallToLarge(config *PipelineConfig, smResult *LayerResult) bool {
	if smResult == nil || !smResult.Called {
		return false
	}
	switch config.SmallToLargeCondition {
	case "never":
		return false
	case "always":
		return true
	case "confidence_lt":
		return smResult.Confidence < config.SmallToLargeConfidenceThresh
	case "category_in":
		for _, cat := range smResult.Categories {
			for _, target := range config.SmallToLargeCategories {
				if cat == target {
					return true
				}
			}
		}
		return false
	default:
		return smResult.Confidence < config.SmallToLargeConfidenceThresh // 默认 confidence_lt
	}
}

// computeFinalVerdict 根据终判逻辑计算最终结果
func (e *Executor) computeFinalVerdict(config *PipelineConfig, result *PipelineResult) string {
	switch config.FinalVerdictLogic {
	case "large_overrides":
		if result.LargeModelResult != nil && result.LargeModelResult.Called {
			if result.LargeModelResult.Passed {
				return "pass"
			}
			return "reject"
		}
		// fall through to last_model_wins
	case "small_overrides":
		if result.SmallModelResult != nil && result.SmallModelResult.Called {
			if result.SmallModelResult.Passed {
				return "pass"
			}
			return "reject"
		}
		if result.AcResult != nil && result.AcResult.Called {
			if result.AcResult.Passed {
				return "pass"
			}
			return "reject"
		}
		return "need_review"
	default: // "last_model_wins"
		// 倒序查找最后一个被调用的层
		if result.LargeModelResult != nil && result.LargeModelResult.Called {
			if result.LargeModelResult.Passed {
				return "pass"
			}
			return "reject"
		}
		if result.SmallModelResult != nil && result.SmallModelResult.Called {
			if result.SmallModelResult.Passed {
				return "pass"
			}
			return "reject"
		}
		if result.AcResult != nil && result.AcResult.Called {
			if result.AcResult.Passed {
				return "pass"
			}
			return "reject"
		}
		return "need_review"
	}
}

// modelCallResult 模型调用结果
type modelCallResult struct {
	IsSafe      bool
	Confidence  float64
	Categories  []string
	Reason      string
	ModelUsed   string
	RiskLevel   string
	RawResponse string
}

// callModel 通过 gRPC 调用 ai-model-service 的 ModerateText
func (e *Executor) callModel(ctx context.Context, templateID, configKey, content string) (*modelCallResult, error) {
	req := &aimodelv1.ModerateTextRequest{
		Content:       content,
		CallerService: "moderation-service",
		TemplateId:    templateID,
	}

	resp, err := e.aiModelClient.ModerateText(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &modelCallResult{
		IsSafe:      resp.IsSafe,
		Confidence:  float64(resp.Confidence),
		Categories:  resp.Categories,
		Reason:      resp.Reason,
		ModelUsed:   resp.ModelUsed,
		RiskLevel:   resp.RiskLevel,
		RawResponse: resp.Reason, // ModerateTextResponse 没有 raw_response 字段，用 reason 代替
	}
	return result, nil
}
```

- [ ] **Step 3: 验证编译**

```bash
cd services/moderation-service && go build ./internal/pipeline/...
```

---

### Task 6: Pipeline CRUD Logic — 管线配置管理逻辑层

**Files:**
- Create: `services/moderation-service/api/internal/logic/pipeline/create_pipeline_logic.go`
- Create: `services/moderation-service/api/internal/logic/pipeline/update_pipeline_logic.go`
- Create: `services/moderation-service/api/internal/logic/pipeline/delete_pipeline_logic.go`
- Create: `services/moderation-service/api/internal/logic/pipeline/get_pipeline_logic.go`
- Create: `services/moderation-service/api/internal/logic/pipeline/list_pipelines_logic.go`
- Create: `services/moderation-service/api/internal/logic/pipeline/activate_pipeline_logic.go`

- [ ] **Step 1: 编写 `create_pipeline_logic.go`**

```go
package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"

	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"
	"community-moderation-service/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePipelineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePipelineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePipelineLogic {
	return &CreatePipelineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePipelineLogic) CreatePipeline(req *types.CreatePipelineReq) (*types.CreatePipelineResp, error) {
	if req.PipelineId == "" {
		return nil, ErrPipelineIdRequired
	}
	if req.PipelineName == "" {
		return nil, ErrPipelineNameRequired
	}

	// 检查 pipeline_id 是否已存在
	existing, _ := l.svcCtx.PipelineModel.FindOneByPipelineId(l.ctx, req.PipelineId)
	if existing != nil {
		return nil, ErrPipelineIdDuplicate
	}

	var acCats, slCats sql.NullString
	if len(req.AcToSmallCategories) > 0 {
		b, _ := json.Marshal(req.AcToSmallCategories)
		acCats = sql.NullString{String: string(b), Valid: true}
	}
	if len(req.SmallToLargeCategories) > 0 {
		b, _ := json.Marshal(req.SmallToLargeCategories)
		slCats = sql.NullString{String: string(b), Valid: true}
	}

	data := &model.ModPipelineConfig{
		PipelineId:                   req.PipelineId,
		PipelineName:                 req.PipelineName,
		Description:                  newNullString(req.Description),
		IsActive:                     1,
		IsProduction:                 0,
		AcEnabled:                    req.AcEnabled,
		AcSeverityThreshold:          req.AcSeverityThreshold,
		SmallModelTemplateId:         newNullString(req.SmallModelTemplateId),
		SmallModelConfigKey:          newNullString(req.SmallModelConfigKey),
		LargeModelTemplateId:         newNullString(req.LargeModelTemplateId),
		LargeModelConfigKey:          newNullString(req.LargeModelConfigKey),
		AcToSmallCondition:           req.AcToSmallCondition,
		AcToSmallSeverity:            newNullInt64(req.AcToSmallSeverity),
		AcToSmallCategories:          acCats,
		SmallToLargeCondition:        req.SmallToLargeCondition,
		SmallToLargeConfidenceThresh: newNullFloat64(req.SmallToLargeConfidenceThresh),
		SmallToLargeCategories:       slCats,
		FinalVerdictLogic:            req.FinalVerdictLogic,
	}

	result, err := l.svcCtx.PipelineModel.Insert(l.ctx, data)
	if err != nil {
		l.Logger.Errorf("insert pipeline config failed: %v", err)
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &types.CreatePipelineResp{
		Id:         uint64(id),
		PipelineId: req.PipelineId,
	}, nil
}

func newNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func newNullInt64(v int64) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v, Valid: true}
}

func newNullFloat64(v float64) sql.NullFloat64 {
	if v == 0 {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: v, Valid: true}
}
```

- [ ] **Step 2: 编写 `update_pipeline_logic.go`**

```go
package pipeline

import (
	"context"
	"encoding/json"

	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePipelineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePipelineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePipelineLogic {
	return &UpdatePipelineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePipelineLogic) UpdatePipeline(req *types.UpdatePipelineReq) error {
	if req.Id == 0 {
		return ErrPipelineIdRequired
	}

	existing, err := l.svcCtx.PipelineModel.FindOne(l.ctx, req.Id)
	if err != nil {
		l.Logger.Errorf("pipeline not found: id=%d, err=%v", req.Id, err)
		return ErrPipelineNotFound
	}

	// 只更新非零值字段
	if req.PipelineName != "" {
		existing.PipelineName = req.PipelineName
	}
	if req.Description != "" {
		existing.Description = newNullString(req.Description)
	}
	if req.IsActive != 0 {
		existing.IsActive = req.IsActive
	}
	if req.IsProduction != 0 {
		existing.IsProduction = req.IsProduction
	}
	if req.AcEnabled != 0 {
		existing.AcEnabled = req.AcEnabled
	}
	if req.AcSeverityThreshold != 0 {
		existing.AcSeverityThreshold = req.AcSeverityThreshold
	}
	if req.SmallModelTemplateId != "" {
		existing.SmallModelTemplateId = newNullString(req.SmallModelTemplateId)
	}
	existing.SmallModelConfigKey = newNullString(req.SmallModelConfigKey)
	if req.LargeModelTemplateId != "" {
		existing.LargeModelTemplateId = newNullString(req.LargeModelTemplateId)
	}
	existing.LargeModelConfigKey = newNullString(req.LargeModelConfigKey)
	if req.AcToSmallCondition != "" {
		existing.AcToSmallCondition = req.AcToSmallCondition
	}
	if req.AcToSmallSeverity != 0 {
		existing.AcToSmallSeverity = newNullInt64(req.AcToSmallSeverity)
	}
	if req.AcToSmallCategories != nil {
		b, _ := json.Marshal(req.AcToSmallCategories)
		existing.AcToSmallCategories = newNullString(string(b))
	}
	if req.SmallToLargeCondition != "" {
		existing.SmallToLargeCondition = req.SmallToLargeCondition
	}
	if req.SmallToLargeConfidenceThresh != 0 {
		existing.SmallToLargeConfidenceThresh = newNullFloat64(req.SmallToLargeConfidenceThresh)
	}
	if req.SmallToLargeCategories != nil {
		b, _ := json.Marshal(req.SmallToLargeCategories)
		existing.SmallToLargeCategories = newNullString(string(b))
	}
	if req.FinalVerdictLogic != "" {
		existing.FinalVerdictLogic = req.FinalVerdictLogic
	}

	return l.svcCtx.PipelineModel.Update(l.ctx, existing)
}
```

- [ ] **Step 3: 编写 `delete_pipeline_logic.go`**

```go
package pipeline

import (
	"context"
	"errors"

	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrPipelineIdRequired = errors.New("pipeline_id 不能为空")
	ErrPipelineNameRequired = errors.New("pipeline_name 不能为空")
	ErrPipelineIdDuplicate  = errors.New("pipeline_id 已存在")
	ErrPipelineNotFound     = errors.New("管线配置不存在")
)

type DeletePipelineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePipelineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePipelineLogic {
	return &DeletePipelineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePipelineLogic) DeletePipeline(req *types.DeletePipelineReq) error {
	existing, err := l.svcCtx.PipelineModel.FindOneByPipelineId(l.ctx, req.PipelineId)
	if err != nil {
		return ErrPipelineNotFound
	}
	return l.svcCtx.PipelineModel.Delete(l.ctx, existing.Id)
}
```

- [ ] **Step 4: 编写 `get_pipeline_logic.go`**

```go
package pipeline

import (
	"context"
	"encoding/json"

	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"
	"community-moderation-service/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPipelineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPipelineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPipelineLogic {
	return &GetPipelineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPipelineLogic) GetPipeline(req *types.GetPipelineReq) (*types.PipelineInfo, error) {
	m, err := l.svcCtx.PipelineModel.FindOneByPipelineId(l.ctx, req.PipelineId)
	if err != nil {
		return nil, ErrPipelineNotFound
	}
	return toPipelineInfo(m), nil
}

func toPipelineInfo(m *model.ModPipelineConfig) *types.PipelineInfo {
	var acCats, slCats []string
	if m.AcToSmallCategories.Valid {
		json.Unmarshal([]byte(m.AcToSmallCategories.String), &acCats)
	}
	if m.SmallToLargeCategories.Valid {
		json.Unmarshal([]byte(m.SmallToLargeCategories.String), &slCats)
	}

	smallConf := 0.0
	if m.SmallToLargeConfidenceThresh.Valid {
		smallConf = m.SmallToLargeConfidenceThresh.Float64
	}

	return &types.PipelineInfo{
		Id:                           m.Id,
		PipelineId:                   m.PipelineId,
		PipelineName:                 m.PipelineName,
		Description:                  model.NullStringToString(m.Description),
		IsActive:                     m.IsActive,
		IsProduction:                 m.IsProduction,
		AcEnabled:                    m.AcEnabled,
		AcSeverityThreshold:          m.AcSeverityThreshold,
		SmallModelTemplateId:         model.NullStringToString(m.SmallModelTemplateId),
		SmallModelConfigKey:          model.NullStringToString(m.SmallModelConfigKey),
		LargeModelTemplateId:         model.NullStringToString(m.LargeModelTemplateId),
		LargeModelConfigKey:          model.NullStringToString(m.LargeModelConfigKey),
		AcToSmallCondition:           m.AcToSmallCondition,
		AcToSmallSeverity:            model.NullInt64ToInt64(m.AcToSmallSeverity),
		AcToSmallCategories:          acCats,
		SmallToLargeCondition:        m.SmallToLargeCondition,
		SmallToLargeConfidenceThresh: smallConf,
		SmallToLargeCategories:       slCats,
		FinalVerdictLogic:            m.FinalVerdictLogic,
		CreatedTime:                  m.CreatedTime.Format("2006-01-02 15:04:05"),
		UpdatedTime:                  m.UpdatedTime.Format("2006-01-02 15:04:05"),
	}
}
```

- [ ] **Step 5: 编写 `list_pipelines_logic.go`**

```go
package pipeline

import (
	"context"

	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPipelinesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPipelinesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPipelinesLogic {
	return &ListPipelinesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPipelinesLogic) ListPipelines(req *types.ListPipelinesReq) (*types.ListPipelinesResp, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	list, err := l.svcCtx.PipelineModel.FindList(l.ctx, req.Page, req.PageSize)
	if err != nil {
		l.Logger.Errorf("list pipelines failed: %v", err)
		return nil, err
	}

	total, err := l.svcCtx.PipelineModel.Count(l.ctx)
	if err != nil {
		l.Logger.Errorf("count pipelines failed: %v", err)
		return nil, err
	}

	result := make([]types.PipelineInfo, 0, len(list))
	for _, m := range list {
		result = append(result, *toPipelineInfo(m))
	}

	return &types.ListPipelinesResp{
		List:     result,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
```

- [ ] **Step 6: 编写 `activate_pipeline_logic.go`**

```go
package pipeline

import (
	"context"

	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActivatePipelineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivatePipelineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivatePipelineLogic {
	return &ActivatePipelineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActivatePipelineLogic) ActivatePipeline(req *types.ActivatePipelineReq) error {
	target, err := l.svcCtx.PipelineModel.FindOneByPipelineId(l.ctx, req.PipelineId)
	if err != nil {
		return ErrPipelineNotFound
	}

	// 将所有配置的 is_production 清 0
	// 注意：这个操作需要事务保证，简化实现为逐条更新
	allPipelines, err := l.svcCtx.PipelineModel.FindList(l.ctx, 1, 1000)
	if err != nil {
		return err
	}
	for _, p := range allPipelines {
		if p.IsProduction == 1 {
			p.IsProduction = 0
			l.svcCtx.PipelineModel.Update(l.ctx, p)
		}
	}

	// 设置目标配置为生产
	target.IsProduction = 1
	return l.svcCtx.PipelineModel.Update(l.ctx, target)
}
```

- [ ] **Step 7: 验证编译**

```bash
cd services/moderation-service && go build ./api/internal/logic/pipeline/...
```

---

### Task 7: Pipeline Test Logic — 管线测试逻辑层

**Files:**
- Create: `services/moderation-service/api/internal/logic/pipeline/pipeline_test_logic.go`

- [ ] **Step 1: 编写 `pipeline_test_logic.go`**

```go
package pipeline

import (
	"context"
	"errors"

	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"
	"community-moderation-service/internal/pipeline"

	"github.com/zeromicro/go-zero/core/logx"
)

var errContentRequired = errors.New("测试文本内容不能为空")

type PipelineTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPipelineTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PipelineTestLogic {
	return &PipelineTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PipelineTestLogic) PipelineTest(req *types.PipelineTestReq) (*types.PipelineTestResp, error) {
	if req.Content == "" {
		return nil, errContentRequired
	}

	var config *pipeline.PipelineConfig

	if req.PipelineId != "" {
		m, err := l.svcCtx.PipelineModel.FindOneByPipelineId(l.ctx, req.PipelineId)
		if err != nil {
			return nil, ErrPipelineNotFound
		}
		config = pipeline.FromDBModel(m)
	} else {
		smallConf := req.SmallToLargeConfidence
		if smallConf == 0 {
			smallConf = 0.90
		}
		config = &pipeline.PipelineConfig{
			PipelineId:                   "adhoc",
			PipelineName:                 "临时配置",
			AcEnabled:                    true,
			AcSeverityThreshold:          1,
			SmallModelTemplateId:         req.SmallModelTemplateId,
			LargeModelTemplateId:         req.LargeModelTemplateId,
			AcToSmallCondition:           "any_hit",
			SmallToLargeCondition:        "confidence_lt",
			SmallToLargeConfidenceThresh: smallConf,
			FinalVerdictLogic:            "last_model_wins",
		}
	}

	executor := l.svcCtx.PipelineExecutor
	result, err := executor.Execute(l.ctx, config, req.Content)
	if err != nil {
		l.Logger.Errorf("pipeline test failed: %v", err)
		return nil, err
	}

	return &types.PipelineTestResp{
		PipelineId:       result.PipelineId,
		AcResult:         convertLayerResult(result.AcResult),
		SmallModelResult: convertLayerResult(result.SmallModelResult),
		LargeModelResult: convertLayerResult(result.LargeModelResult),
		FinalVerdict:     result.FinalVerdict,
		TotalLatencyMs:   result.TotalLatencyMs,
	}, nil
}

func convertLayerResult(lr *pipeline.LayerResult) *types.LayerResult {
	if lr == nil {
		return nil
	}
	return &types.LayerResult{
		Called:        lr.Called,
		SkippedReason: lr.SkippedReason,
		Passed:        lr.Passed,
		RiskLevel:     lr.RiskLevel,
		Confidence:    lr.Confidence,
		Categories:    lr.Categories,
		Reason:        lr.Reason,
		LatencyMs:     lr.LatencyMs,
		MatchedWords:  lr.MatchedWords,
		ModelUsed:     lr.ModelUsed,
		TemplateID:    lr.TemplateID,
		RawResponse:   lr.RawResponse,
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
cd services/moderation-service && go build ./api/internal/logic/pipeline/...
```

---

### Task 8: Pipeline CRUD Handlers — 管线配置管理 Handler 层

**Files:**
- Create: `services/moderation-service/api/internal/handler/pipeline/create_pipeline_handler.go`
- Create: `services/moderation-service/api/internal/handler/pipeline/update_pipeline_handler.go`
- Create: `services/moderation-service/api/internal/handler/pipeline/delete_pipeline_handler.go`
- Create: `services/moderation-service/api/internal/handler/pipeline/get_pipeline_handler.go`
- Create: `services/moderation-service/api/internal/handler/pipeline/list_pipelines_handler.go`
- Create: `services/moderation-service/api/internal/handler/pipeline/activate_pipeline_handler.go`

- [ ] **Step 1: 编写 `create_pipeline_handler.go`**

```go
package pipeline

import (
	"net/http"

	"community-moderation-service/api/internal/logic/pipeline"
	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreatePipelineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreatePipelineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pipeline.NewCreatePipelineLogic(r.Context(), svcCtx)
		resp, err := l.CreatePipeline(&req)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
```

- [ ] **Step 2: 编写 `update_pipeline_handler.go`**

```go
package pipeline

import (
	"net/http"

	"community-moderation-service/api/internal/logic/pipeline"
	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdatePipelineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdatePipelineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pipeline.NewUpdatePipelineLogic(r.Context(), svcCtx)
		if err := l.UpdatePipeline(&req); err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, nil)
	}
}
```

- [ ] **Step 3: 编写 `delete_pipeline_handler.go`**

```go
package pipeline

import (
	"net/http"

	"community-moderation-service/api/internal/logic/pipeline"
	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeletePipelineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeletePipelineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pipeline.NewDeletePipelineLogic(r.Context(), svcCtx)
		if err := l.DeletePipeline(&req); err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, nil)
	}
}
```

- [ ] **Step 4: 编写 `get_pipeline_handler.go`**

```go
package pipeline

import (
	"net/http"

	"community-moderation-service/api/internal/logic/pipeline"
	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPipelineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPipelineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pipeline.NewGetPipelineLogic(r.Context(), svcCtx)
		resp, err := l.GetPipeline(&req)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
```

- [ ] **Step 5: 编写 `list_pipelines_handler.go`**

```go
package pipeline

import (
	"net/http"

	"community-moderation-service/api/internal/logic/pipeline"
	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListPipelinesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListPipelinesReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pipeline.NewListPipelinesLogic(r.Context(), svcCtx)
		resp, err := l.ListPipelines(&req)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
```

- [ ] **Step 6: 编写 `activate_pipeline_handler.go`**

```go
package pipeline

import (
	"net/http"

	"community-moderation-service/api/internal/logic/pipeline"
	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ActivatePipelineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ActivatePipelineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pipeline.NewActivatePipelineLogic(r.Context(), svcCtx)
		if err := l.ActivatePipeline(&req); err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, nil)
	}
}
```

- [ ] **Step 7: 验证编译**

```bash
cd services/moderation-service && go build ./api/internal/handler/pipeline/...
```

---

### Task 9: Pipeline Test Handler

**Files:**
- Create: `services/moderation-service/api/internal/handler/pipeline/pipeline_test_handler.go`

- [ ] **Step 1: 编写 `pipeline_test_handler.go`**

```go
package pipeline

import (
	"net/http"

	"community-moderation-service/api/internal/logic/pipeline"
	"community-moderation-service/api/internal/svc"
	"community-moderation-service/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PipelineTestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PipelineTestReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pipeline.NewPipelineTestLogic(r.Context(), svcCtx)
		resp, err := l.PipelineTest(&req)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
```

- [ ] **Step 2: 验证编译**

```bash
cd services/moderation-service && go build ./api/internal/handler/pipeline/...
```

---

### Task 10: Route Registration — 注册新路由

**Files:**
- Modify: `services/moderation-service/api/internal/handler/routes.go`

- [ ] **Step 1: 在 `routes.go` 中添加 pipeline 路由组**

在现有的三个 `server.AddRoutes` 调用之后（或中间），添加新的 pipeline 路由组：

```go
// Pipeline configuration routes (JWT protected)
server.AddRoutes(
    rest.WithMiddlewares(
        []rest.Middleware{serverCtx.AuthMiddleware},
        []rest.Route{
            {
                Method:  http.MethodPost,
                Path:    "/pipeline",
                Handler: pipelinehdlr.CreatePipelineHandler(serverCtx),
            },
            {
                Method:  http.MethodPut,
                Path:    "/pipeline",
                Handler: pipelinehdlr.UpdatePipelineHandler(serverCtx),
            },
            {
                Method:  http.MethodDelete,
                Path:    "/pipeline/:pipeline_id",
                Handler: pipelinehdlr.DeletePipelineHandler(serverCtx),
            },
            {
                Method:  http.MethodGet,
                Path:    "/pipeline/:pipeline_id",
                Handler: pipelinehdlr.GetPipelineHandler(serverCtx),
            },
            {
                Method:  http.MethodGet,
                Path:    "/pipelines",
                Handler: pipelinehdlr.ListPipelinesHandler(serverCtx),
            },
            {
                Method:  http.MethodPut,
                Path:    "/pipeline/:pipeline_id/activate",
                Handler: pipelinehdlr.ActivatePipelineHandler(serverCtx),
            },
            {
                Method:  http.MethodPost,
                Path:    "/pipeline/test",
                Handler: pipelinehdlr.PipelineTestHandler(serverCtx),
            },
        }...,
    ),
    rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
    rest.WithPrefix("/api/moderation"),
)
```

- [ ] **Step 2: 在 `routes.go` 顶部添加 import**

```go
import (
    // ... 现有 imports ...
    pipelinehdlr "community-moderation-service/api/internal/handler/pipeline"
)
```

- [ ] **Step 3: 验证编译**

```bash
cd services/moderation-service && go build ./api/...
```

---

### Task 11: ServiceContext Wiring — 注入 Pipeline 依赖

**Files:**
- Modify: `services/moderation-service/api/internal/svc/service_context.go`

- [ ] **Step 1: 在 ServiceContext struct 中新增字段**

```go
type ServiceContext struct {
    Config         config.Config
    AuthMiddleware rest.Middleware
    WordStore      *wordstore.WordStore
    TextEngine     *engine.TextEngine
    ImageEngine    *engine.ImageEngine
    AuditLogger    *auditlog.AuditLogger
    // 新增
    PipelineModel    model.ModPipelineConfigModel
    PipelineExecutor *pipeline.Executor
}
```

- [ ] **Step 2: 在 NewServiceContext 构造函数中注入**

在现有构造函数末尾（`return &ServiceContext{...}` 之前），添加：

```go
// Pipeline model & executor
pipelineModel := model.NewModPipelineConfigModel(sqlx.NewMysql(c.LogDataSource))
pipelineExecutor := pipeline.NewExecutor(textEngine, getAiModelClient())

// 在 return 语句中添加：
return &ServiceContext{
    // ... 现有字段 ...
    PipelineModel:    pipelineModel,
    PipelineExecutor: pipelineExecutor,
}
```

- [ ] **Step 3: 添加 import**

```go
import (
    // ... 现有 imports ...
    "community-moderation-service/internal/pipeline"
    "community-moderation-service/model"
)
```

- [ ] **Step 4: 验证编译**

```bash
cd services/moderation-service && go build ./...
```

---

### Task 12: 前端类型定义 — Pipeline 类型

**Files:**
- Modify: `web/common/types/moderation.d.ts`

- [ ] **Step 1: 在 moderation.d.ts 末尾追加类型**

```typescript
// ========== 管线配置类型 ==========

export interface PipelineConfig {
  id: string;
  pipeline_id: string;
  pipeline_name: string;
  description: string;
  is_active: number;
  is_production: number;
  ac_enabled: number;
  ac_severity_threshold: number;
  small_model_template_id: string;
  small_model_config_key: string;
  large_model_template_id: string;
  large_model_config_key: string;
  ac_to_small_condition: string;
  ac_to_small_severity: number;
  ac_to_small_categories: string[];
  small_to_large_condition: string;
  small_to_large_confidence_threshold: number;
  small_to_large_categories: string[];
  final_verdict_logic: string;
  created_time: string;
  updated_time: string;
}

export interface PipelineListResponse {
  list: PipelineConfig[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreatePipelineRequest {
  pipeline_id: string;
  pipeline_name: string;
  description?: string;
  ac_enabled?: number;
  ac_severity_threshold?: number;
  small_model_template_id?: string;
  small_model_config_key?: string;
  large_model_template_id?: string;
  large_model_config_key?: string;
  ac_to_small_condition?: string;
  ac_to_small_severity?: number;
  ac_to_small_categories?: string[];
  small_to_large_condition?: string;
  small_to_large_confidence_threshold?: number;
  small_to_large_categories?: string[];
  final_verdict_logic?: string;
}

export interface UpdatePipelineRequest {
  id: string;
  pipeline_name?: string;
  description?: string;
  is_active?: number;
  is_production?: number;
  ac_enabled?: number;
  ac_severity_threshold?: number;
  small_model_template_id?: string;
  small_model_config_key?: string;
  large_model_template_id?: string;
  large_model_config_key?: string;
  ac_to_small_condition?: string;
  ac_to_small_severity?: number;
  ac_to_small_categories?: string[];
  small_to_large_condition?: string;
  small_to_large_confidence_threshold?: number;
  small_to_large_categories?: string[];
  final_verdict_logic?: string;
}

// ========== 管线测试类型 ==========

export interface PipelineTestRequest {
  content: string;
  pipeline_id?: string;
  small_model_template_id?: string;
  large_model_template_id?: string;
  small_to_large_confidence?: number;
}

export interface PipelineLayerResult {
  called: boolean;
  skipped_reason?: string;
  passed: boolean;
  risk_level: string;
  confidence: number;
  categories?: string[];
  reason?: string;
  latency_ms: number;
  matched_words?: string[];
  model_used?: string;
  template_id?: string;
  raw_response?: string;
}

export interface PipelineTestResponse {
  pipeline_id: string;
  ac_result: PipelineLayerResult | null;
  small_model_result: PipelineLayerResult | null;
  large_model_result: PipelineLayerResult | null;
  final_verdict: string;
  total_latency_ms: number;
}
```

---

### Task 13: 前端 API 层 — Pipeline CRUD + 测试 API

**Files:**
- Modify: `web/pc/src/api/moderation.ts`
- Modify: `web/pc/src/api/aimodel.ts`

- [ ] **Step 1: 在 `moderation.ts` 追加 Pipeline API**

```typescript
import type {
  PipelineConfig,
  PipelineListResponse,
  CreatePipelineRequest,
  UpdatePipelineRequest,
  PipelineTestRequest,
  PipelineTestResponse,
} from '@common/types/moderation';
import type { PaginationParams } from '@common/types/common';

// ========== Pipeline CRUD ==========

export function createPipeline(data: CreatePipelineRequest) {
  return request.post<{ id: string; pipeline_id: string }>(
    '/api/moderation/pipeline',
    data
  );
}

export function updatePipeline(data: UpdatePipelineRequest) {
  return request.put<null>('/api/moderation/pipeline', data);
}

export function deletePipeline(pipelineId: string) {
  return request.delete<null>(`/api/moderation/pipeline/${pipelineId}`);
}

export function getPipeline(pipelineId: string) {
  return request.get<PipelineConfig>(`/api/moderation/pipeline/${pipelineId}`);
}

export function listPipelines(params?: PaginationParams) {
  return request.get<PipelineListResponse>('/api/moderation/pipelines', { params });
}

export function activatePipeline(pipelineId: string) {
  return request.put<null>(`/api/moderation/pipeline/${pipelineId}/activate`);
}

// ========== Pipeline Test ==========

export function testPipeline(data: PipelineTestRequest) {
  return request.post<PipelineTestResponse>(
    '/api/moderation/pipeline/test',
    data,
    { timeout: 60000 }
  );
}
```

- [ ] **Step 2: 在 `aimodel.ts` 追加模板/模型便捷查询**

在现有文件末尾追加：

```typescript
// 获取审核类型的模板列表（用于下拉选择）
export function getModerationTemplates() {
  return request.get<PaginatedResponse<PromptTemplate>>('/api/v1/templates', {
    params: { category: 'moderation', page: 1, page_size: 50 }
  });
}

// 获取健康可用的模型列表（用于下拉选择）
export function getAvailableModels() {
  return request.get<{ models: ModelConfig[] }>('/api/v1/models');
}
```

---

### Task 14: 前端路由 & 菜单 — 模块配置更新

**Files:**
- Modify: `web/pc/src/config/modules/moderation.config.ts`

- [ ] **Step 1: 更新菜单和路由**

将原来的 `ModerationTest` 替换为 `ModerationConfigTest`：

```typescript
import type { ModuleConfig } from '../types';
import { Warning, Monitor } from '@element-plus/icons-vue';

export const moderationModule: ModuleConfig = {
  name: 'moderation',

  menu: {
    path: '/moderation',
    title: '内容审核',
    icon: Warning,
    children: [
      {
        path: '/moderation/config-test',
        title: '配置测试',
        icon: Monitor
      }
    ]
  },

  routes: [
    {
      path: '/moderation/config-test',
      name: 'ModerationConfigTest',
      component: () => import('@/views/moderation/ModerationConfigTest.vue'),
      meta: { title: '配置测试', requiresAuth: true }
    }
  ]
};
```

---

### Task 15: PipelineSelector 组件 — 管线选择器

**Files:**
- Create: `web/pc/src/components/moderation/PipelineSelector.vue`

- [ ] **Step 1: 编写组件**

```vue
<template>
  <div class="pipeline-selector">
    <el-select
      v-model="selectedId"
      placeholder="选择管线配置"
      :loading="loading"
      @change="handleSelect"
      style="width: 240px;"
    >
      <el-option
        v-for="p in pipelines"
        :key="p.pipeline_id"
        :label="`${p.pipeline_name}${p.is_production ? ' (生产中)' : ''}`"
        :value="p.pipeline_id"
      />
    </el-select>

    <el-button @click="handleNew" :icon="Plus">新建</el-button>
    <el-button
      @click="handleCopy"
      :disabled="!selectedId"
      :icon="CopyDocument"
    >
      复制
    </el-button>
    <el-button
      type="warning"
      @click="handleActivate"
      :disabled="!selectedId"
      :icon="CircleCheck"
    >
      设为生产
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, CopyDocument, CircleCheck } from '@element-plus/icons-vue';
import { listPipelines, createPipeline, activatePipeline, getPipeline } from '@/api/moderation';
import type { PipelineConfig } from '@common/types/moderation';

const emit = defineEmits<{
  (e: 'select', config: PipelineConfig): void;
  (e: 'new'): void;
}>();

const selectedId = ref('');
const pipelines = ref<PipelineConfig[]>([]);
const loading = ref(false);

const loadPipelines = async () => {
  loading.value = true;
  try {
    const resp = await listPipelines({ page: 1, page_size: 50 });
    pipelines.value = resp.list;
    // 默认选中生产配置
    const prod = resp.list.find(p => p.is_production === 1);
    if (prod && !selectedId.value) {
      selectedId.value = prod.pipeline_id;
      emit('select', prod);
    }
  } catch (e: any) {
    ElMessage.error(e.message || '加载管线列表失败');
  } finally {
    loading.value = false;
  }
};

const handleSelect = async (pipelineId: string) => {
  try {
    const config = await getPipeline(pipelineId);
    emit('select', config);
  } catch (e: any) {
    ElMessage.error(e.message || '加载管线配置失败');
  }
};

const handleNew = () => {
  emit('new');
};

const handleCopy = async () => {
  if (!selectedId.value) return;
  try {
    const config = await getPipeline(selectedId.value);
    const newId = `${config.pipeline_id}_copy_${Date.now()}`;
    await createPipeline({
      pipeline_id: newId,
      pipeline_name: `${config.pipeline_name} (副本)`,
      description: config.description,
      ac_enabled: config.ac_enabled,
      ac_severity_threshold: config.ac_severity_threshold,
      small_model_template_id: config.small_model_template_id,
      small_model_config_key: config.small_model_config_key,
      large_model_template_id: config.large_model_template_id,
      large_model_config_key: config.large_model_config_key,
      ac_to_small_condition: config.ac_to_small_condition,
      ac_to_small_severity: config.ac_to_small_severity,
      ac_to_small_categories: config.ac_to_small_categories,
      small_to_large_condition: config.small_to_large_condition,
      small_to_large_confidence_threshold: config.small_to_large_confidence_threshold,
      small_to_large_categories: config.small_to_large_categories,
      final_verdict_logic: config.final_verdict_logic,
    });
    ElMessage.success('配置复制成功');
    await loadPipelines();
    selectedId.value = newId;
    handleSelect(newId);
  } catch (e: any) {
    ElMessage.error(e.message || '复制失败');
  }
};

const handleActivate = async () => {
  if (!selectedId.value) return;
  try {
    await ElMessageBox.confirm(
      '将此管线配置设为生产环境默认配置？',
      '确认操作',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    );
    await activatePipeline(selectedId.value);
    ElMessage.success('已设为生产配置');
    await loadPipelines();
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '操作失败');
    }
  }
};

const reset = () => {
  selectedId.value = '';
};

onMounted(loadPipelines);

defineExpose({ loadPipelines, reset });
</script>

<style scoped>
.pipeline-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}
</style>
```

---

### Task 16: LayerConfigPanel 组件 — 三层配置面板

**Files:**
- Create: `web/pc/src/components/moderation/LayerConfigPanel.vue`

- [ ] **Step 1: 编写组件**

```vue
<template>
  <el-row :gutter="16" class="layer-config-panel">
    <!-- AC 引擎 -->
    <el-col :span="8">
      <el-card shadow="never">
        <template #header>
          <div class="layer-header">
            <span>AC 引擎</span>
            <el-switch v-model="localConfig.ac_enabled" :active-value="1" :inactive-value="0" />
          </div>
        </template>
        <el-form label-width="80px" size="small">
          <el-form-item label="严重度≥">
            <el-select v-model="localConfig.ac_severity_threshold" :disabled="!localConfig.ac_enabled" style="width: 100%;">
              <el-option :value="1" label="1 - 高危" />
              <el-option :value="2" label="2 - 中危" />
              <el-option :value="3" label="3 - 低危" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>

    <!-- 小模型 -->
    <el-col :span="8">
      <el-card shadow="never">
        <template #header>
          <div class="layer-header">
            <span>小模型</span>
            <el-switch
              v-model="smallEnabled"
              @change="onSmallToggle"
            />
          </div>
        </template>
        <el-form label-width="60px" size="small">
          <el-form-item label="模板">
            <el-select
              v-model="localConfig.small_model_template_id"
              :disabled="!smallEnabled"
              placeholder="选择提示词模板"
              style="width: 100%;"
            >
              <el-option
                v-for="t in templates"
                :key="t.template_id"
                :label="`${t.template_name} v${t.version}`"
                :value="t.template_id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="模型">
            <el-select
              v-model="localConfig.small_model_config_key"
              :disabled="!smallEnabled"
              placeholder="默认使用模板关联模型"
              clearable
              style="width: 100%;"
            >
              <el-option
                v-for="m in models"
                :key="m.config_key"
                :label="m.display_name"
                :value="m.config_key"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>

    <!-- 大模型 -->
    <el-col :span="8">
      <el-card shadow="never">
        <template #header>
          <div class="layer-header">
            <span>大模型</span>
            <el-switch
              v-model="largeEnabled"
              @change="onLargeToggle"
            />
          </div>
        </template>
        <el-form label-width="60px" size="small">
          <el-form-item label="模板">
            <el-select
              v-model="localConfig.large_model_template_id"
              :disabled="!largeEnabled"
              placeholder="选择提示词模板"
              style="width: 100%;"
            >
              <el-option
                v-for="t in templates"
                :key="t.template_id"
                :label="`${t.template_name} v${t.version}`"
                :value="t.template_id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="模型">
            <el-select
              v-model="localConfig.large_model_config_key"
              :disabled="!largeEnabled"
              placeholder="默认使用模板关联模型"
              clearable
              style="width: 100%;"
            >
              <el-option
                v-for="m in models"
                :key="m.config_key"
                :label="m.display_name"
                :value="m.config_key"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue';
import { getModerationTemplates, getAvailableModels } from '@/api/aimodel';
import type { PromptTemplate, ModelConfig } from '@common/types/aimodel';

export interface LayerConfigValues {
  ac_enabled: number;
  ac_severity_threshold: number;
  small_model_template_id: string;
  small_model_config_key: string;
  large_model_template_id: string;
  large_model_config_key: string;
}

const props = defineProps<{
  modelValue: LayerConfigValues;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', v: LayerConfigValues): void;
}>();

const templates = ref<PromptTemplate[]>([]);
const models = ref<ModelConfig[]>([]);

const localConfig = reactive<LayerConfigValues>({ ...props.modelValue });
const smallEnabled = ref(props.modelValue.small_model_template_id !== '');
const largeEnabled = ref(props.modelValue.large_model_template_id !== '');

watch(localConfig, (v) => emit('update:modelValue', { ...v }), { deep: true });

watch(() => props.modelValue, (v) => {
  Object.assign(localConfig, v);
  smallEnabled.value = v.small_model_template_id !== '';
  largeEnabled.value = v.large_model_template_id !== '';
});

const onSmallToggle = (val: boolean) => {
  if (!val) localConfig.small_model_template_id = '';
};
const onLargeToggle = (val: boolean) => {
  if (!val) localConfig.large_model_template_id = '';
};

onMounted(async () => {
  try {
    const [tResp, mResp] = await Promise.all([
      getModerationTemplates(),
      getAvailableModels()
    ]);
    templates.value = tResp.list || [];
    models.value = (mResp as any).models || mResp.list || [];
  } catch (e) {
    console.error('加载模板/模型列表失败', e);
  }
});
</script>

<style scoped>
.layer-config-panel {
  margin-bottom: 16px;
}
.layer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

---

### Task 17: EscalationRuleEditor 组件 — 升级规则编辑器

**Files:**
- Create: `web/pc/src/components/moderation/EscalationRuleEditor.vue`

- [ ] **Step 1: 编写组件**

```vue
<template>
  <el-card shadow="never" class="escalation-editor">
    <template #header><span>升级规则</span></template>

    <!-- AC → 小模型 -->
    <div class="rule-row">
      <span class="rule-label">AC → 小模型：</span>
      <el-select v-model="localRules.ac_to_small_condition" style="width: 140px;" @change="onAcConditionChange">
        <el-option value="any_hit" label="任何命中" />
        <el-option value="severity_gte" label="严重度 ≥" />
        <el-option value="category_in" label="分类包含" />
        <el-option value="never" label="从不" />
      </el-select>
      <template v-if="localRules.ac_to_small_condition === 'severity_gte'">
        <el-input-number v-model="localRules.ac_to_small_severity" :min="1" :max="3" style="width: 80px; margin-left: 8px;" />
      </template>
      <template v-if="localRules.ac_to_small_condition === 'category_in'">
        <el-select
          v-model="localRules.ac_to_small_categories"
          multiple
          filterable
          allow-create
          placeholder="输入分类名"
          style="width: 240px; margin-left: 8px;"
        />
      </template>
    </div>

    <!-- 小模型 → 大模型 -->
    <div class="rule-row">
      <span class="rule-label">小模型 → 大模型：</span>
      <el-select v-model="localRules.small_to_large_condition" style="width: 140px;" @change="onSlConditionChange">
        <el-option value="confidence_lt" label="置信度 &lt;" />
        <el-option value="category_in" label="分类包含" />
        <el-option value="always" label="总是" />
        <el-option value="never" label="从不" />
      </el-select>
      <template v-if="localRules.small_to_large_condition === 'confidence_lt'">
        <el-input-number
          v-model="localRules.small_to_large_confidence_threshold"
          :min="0" :max="1" :step="0.05" :precision="2"
          style="width: 100px; margin-left: 8px;"
        />
      </template>
      <template v-if="localRules.small_to_large_condition === 'category_in'">
        <el-select
          v-model="localRules.small_to_large_categories"
          multiple filterable allow-create
          placeholder="输入分类名"
          style="width: 240px; margin-left: 8px;"
        />
      </template>
    </div>

    <!-- 终判逻辑 -->
    <div class="rule-row">
      <span class="rule-label">终判逻辑：</span>
      <el-select v-model="localRules.final_verdict_logic" style="width: 180px;">
        <el-option value="last_model_wins" label="最后模型判定" />
        <el-option value="large_overrides" label="大模型覆盖" />
        <el-option value="small_overrides" label="小模型覆盖" />
      </el-select>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue';

export interface EscalationRuleValues {
  ac_to_small_condition: string;
  ac_to_small_severity: number;
  ac_to_small_categories: string[];
  small_to_large_condition: string;
  small_to_large_confidence_threshold: number;
  small_to_large_categories: string[];
  final_verdict_logic: string;
}

const props = defineProps<{
  modelValue: EscalationRuleValues;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', v: EscalationRuleValues): void;
}>();

const localRules = reactive<EscalationRuleValues>({
  ac_to_small_condition: props.modelValue.ac_to_small_condition || 'any_hit',
  ac_to_small_severity: props.modelValue.ac_to_small_severity || 1,
  ac_to_small_categories: props.modelValue.ac_to_small_categories || [],
  small_to_large_condition: props.modelValue.small_to_large_condition || 'confidence_lt',
  small_to_large_confidence_threshold: props.modelValue.small_to_large_confidence_threshold || 0.90,
  small_to_large_categories: props.modelValue.small_to_large_categories || [],
  final_verdict_logic: props.modelValue.final_verdict_logic || 'last_model_wins',
});

watch(localRules, (v) => emit('update:modelValue', { ...v }), { deep: true });

watch(() => props.modelValue, (v) => {
  Object.assign(localRules, {
    ac_to_small_condition: v.ac_to_small_condition || 'any_hit',
    ac_to_small_severity: v.ac_to_small_severity || 1,
    ac_to_small_categories: v.ac_to_small_categories || [],
    small_to_large_condition: v.small_to_large_condition || 'confidence_lt',
    small_to_large_confidence_threshold: v.small_to_large_confidence_threshold || 0.90,
    small_to_large_categories: v.small_to_large_categories || [],
    final_verdict_logic: v.final_verdict_logic || 'last_model_wins',
  });
});

const onAcConditionChange = () => {
  // 切换条件时清空不相关参数
  if (localRules.ac_to_small_condition !== 'severity_gte') {
    localRules.ac_to_small_severity = 1;
  }
  if (localRules.ac_to_small_condition !== 'category_in') {
    localRules.ac_to_small_categories = [];
  }
};

const onSlConditionChange = () => {
  if (localRules.small_to_large_condition !== 'confidence_lt') {
    localRules.small_to_large_confidence_threshold = 0.90;
  }
  if (localRules.small_to_large_condition !== 'category_in') {
    localRules.small_to_large_categories = [];
  }
};
</script>

<style scoped>
.escalation-editor {
  margin-bottom: 16px;
}
.rule-row {
  display: flex;
  align-items: center;
  margin-bottom: 14px;
}
.rule-label {
  width: 140px;
  color: #606266;
  font-size: 14px;
  flex-shrink: 0;
}
</style>
```

---

### Task 18: TestInputArea 组件 — 测试输入区

**Files:**
- Create: `web/pc/src/components/moderation/TestInputArea.vue`

- [ ] **Step 1: 编写组件**

```vue
<template>
  <el-card shadow="never" class="test-input-area">
    <template #header><span>测试区域</span></template>
    <el-input
      v-model="content"
      type="textarea"
      :rows="6"
      placeholder="请输入要测试的文本内容（不超过500字）"
      maxlength="500"
      show-word-limit
    />
    <div class="test-actions">
      <el-button type="primary" :loading="loading" :disabled="!content.trim()" @click="handleTest">
        执行测试
      </el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const emit = defineEmits<{
  (e: 'test', content: string): void;
}>();

const content = ref('');
const loading = ref(false);

const handleTest = () => {
  emit('test', content.value);
};

const handleReset = () => {
  content.value = '';
};

const setLoading = (val: boolean) => { loading.value = val; };

defineExpose({ setLoading });
</script>

<style scoped>
.test-input-area {
  margin-bottom: 16px;
}
.test-actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}
</style>
```

---

### Task 19: PipelineResultPanel 组件 — 结果展示面板

**Files:**
- Create: `web/pc/src/components/moderation/PipelineResultPanel.vue`

- [ ] **Step 1: 编写组件**

```vue
<template>
  <div v-if="result" class="pipeline-result">
    <el-card shadow="never">
      <template #header>
        <div class="result-header">
          <span>执行结果</span>
          <el-tag :type="verdictTagType">{{ verdictLabel }}</el-tag>
        </div>
      </template>

      <el-row :gutter="16">
        <!-- AC 引擎结果 -->
        <el-col :span="8">
          <LayerResultCard
            title="AC 引擎"
            :result="result.ac_result"
            :show-matched-words="true"
          />
        </el-col>

        <!-- 小模型结果 -->
        <el-col :span="8">
          <LayerResultCard
            title="小模型"
            :result="result.small_model_result"
          />
        </el-col>

        <!-- 大模型结果 -->
        <el-col :span="8">
          <LayerResultCard
            title="大模型"
            :result="result.large_model_result"
          />
        </el-col>
      </el-row>

      <!-- 总览表 -->
      <el-table :data="overviewRows" border style="margin-top: 16px;">
        <el-table-column prop="layer" label="审核层" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="called" label="调用情况" width="100">
          <template #default="{ row }">
            <el-tag :type="row.called ? 'success' : 'info'" size="small">
              {{ row.called ? '已调用' : '未调用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="passed" label="结果" width="100">
          <template #default="{ row }">
            <template v-if="row.called">
              <el-tag :type="row.passed ? 'success' : 'danger'" size="small">
                {{ row.passed ? '通过' : '未通过' }}
              </el-tag>
            </template>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="confidence" label="置信度" width="100">
          <template #default="{ row }">
            <span v-if="row.called && row.confidence !== undefined">
              {{ (row.confidence * 100).toFixed(0) }}%
            </span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="latency" label="耗时" width="100">
          <template #default="{ row }">
            <span v-if="row.called">{{ row.latency }}ms</span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="detail" label="详情" min-width="200">
          <template #default="{ row }">
            <span v-if="row.detail" style="color: #909399; font-size: 13px;">{{ row.detail }}</span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { PipelineTestResponse, PipelineLayerResult } from '@common/types/moderation';

const props = defineProps<{
  result: PipelineTestResponse | null;
}>();

const verdictLabel = computed(() => {
  const m: Record<string, string> = { pass: '通过', reject: '拒绝', need_review: '需人工审核' };
  return m[props.result?.final_verdict || ''] || props.result?.final_verdict || '';
});

const verdictTagType = computed(() => {
  const m: Record<string, string> = { pass: 'success', reject: 'danger', need_review: 'warning' };
  return m[props.result?.final_verdict || ''] || 'info';
});

const buildRow = (key: string, label: string, lr: PipelineLayerResult | null) => {
  if (!lr) return { layer: key, label, called: false, passed: false, detail: '' };
  return {
    layer: key,
    label,
    called: lr.called,
    passed: lr.passed,
    confidence: lr.confidence,
    latency: lr.latency_ms,
    detail: lr.called
      ? (lr.reason || lr.matched_words?.join(', ') || (lr.passed ? '' : '未通过'))
      : (lr.skipped_reason || ''),
  };
};

const overviewRows = computed(() => {
  if (!props.result) return [];
  return [
    buildRow('ac', 'AC引擎', props.result.ac_result),
    buildRow('small', '小模型', props.result.small_model_result),
    buildRow('large', '大模型', props.result.large_model_result),
  ];
});
</script>

<style scoped>
.pipeline-result {
  margin-top: 16px;
}
.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

**还需要 `LayerResultCard` 子组件。创建 `LayerResultCard.vue`：**

```vue
<template>
  <div class="layer-card" :class="{ skipped: !result || !result.called }">
    <div class="layer-title">{{ title }}</div>
    <div v-if="result && result.called" class="layer-body">
      <div class="layer-row">
        <el-tag :type="result.passed ? 'success' : 'danger'" size="small">
          {{ result.passed ? '通过' : '未通过' }}
        </el-tag>
        <el-tag style="margin-left: 4px;" size="small">
          {{ result.risk_level || 'unknown' }}
        </el-tag>
      </div>
      <div class="layer-row" v-if="result.confidence !== undefined">
        置信度: {{ (result.confidence * 100).toFixed(0) }}%
      </div>
      <div class="layer-row" v-if="result.matched_words?.length">
        命中: {{ result.matched_words.join(', ') }}
      </div>
      <div class="layer-row" v-if="result.categories?.length">
        分类: {{ result.categories.join(', ') }}
      </div>
      <div class="layer-row" v-if="result.model_used">
        模型: {{ result.model_used }}
      </div>
      <div class="layer-row text-muted">
        耗时: {{ result.latency_ms }}ms
      </div>
      <el-button
        v-if="result.raw_response"
        size="small"
        text
        type="primary"
        @click="copyJson(result.raw_response)"
      >
        复制原始JSON
      </el-button>
    </div>
    <div v-else class="layer-skipped">
      {{ result?.skipped_reason || '未调用' }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import type { PipelineLayerResult } from '@common/types/moderation';

defineProps<{
  title: string;
  result: PipelineLayerResult | null;
  showMatchedWords?: boolean;
}>();

const copyJson = (data: string) => {
  navigator.clipboard.writeText(data).then(() => {
    ElMessage.success('已复制到剪贴板');
  }).catch(() => {
    ElMessage.error('复制失败');
  });
};
</script>

<style scoped>
.layer-card {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 12px;
  min-height: 160px;
}
.layer-card.skipped {
  background-color: #f5f7fa;
}
.layer-title {
  font-weight: 600;
  margin-bottom: 8px;
  font-size: 14px;
}
.layer-row {
  margin-bottom: 4px;
  font-size: 13px;
}
.text-muted {
  color: #909399;
}
.layer-skipped {
  color: #c0c4cc;
  font-size: 13px;
  margin-top: 24px;
  text-align: center;
}
</style>
```

---

### Task 20: ModerationConfigTest.vue — 页面容器

**Files:**
- Create: `web/pc/src/views/moderation/ModerationConfigTest.vue`

- [ ] **Step 1: 编写页面组件**

```vue
<template>
  <div class="config-test-page">
    <el-card>
      <template #header>
        <div class="page-header">
          <span>内容审核配置测试</span>
          <el-button type="primary" :loading="saving" @click="handleSave">保存配置</el-button>
        </div>
      </template>

      <!-- 管线选择器 -->
      <PipelineSelector
        ref="selectorRef"
        @select="onPipelineSelect"
        @new="onNewPipeline"
      />

      <!-- 三层配置 -->
      <LayerConfigPanel v-model="layerConfig" />

      <!-- 升级规则 -->
      <EscalationRuleEditor v-model="escalationRules" />

      <!-- 测试输入 -->
      <TestInputArea ref="testInputRef" @test="onTest" />

      <!-- 执行结果 -->
      <PipelineResultPanel :result="testResult" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import PipelineSelector from '@/components/moderation/PipelineSelector.vue';
import LayerConfigPanel from '@/components/moderation/LayerConfigPanel.vue';
import type { LayerConfigValues } from '@/components/moderation/LayerConfigPanel.vue';
import EscalationRuleEditor from '@/components/moderation/EscalationRuleEditor.vue';
import type { EscalationRuleValues } from '@/components/moderation/EscalationRuleEditor.vue';
import TestInputArea from '@/components/moderation/TestInputArea.vue';
import PipelineResultPanel from '@/components/moderation/PipelineResultPanel.vue';
import { testPipeline, createPipeline, updatePipeline } from '@/api/moderation';
import type { PipelineConfig, PipelineTestResponse } from '@common/types/moderation';

const saving = ref(false);
const testResult = ref<PipelineTestResponse | null>(null);
const selectorRef = ref<InstanceType<typeof PipelineSelector>>();
const testInputRef = ref<InstanceType<typeof TestInputArea>>();
const currentPipelineId = ref('');

const layerConfig = reactive<LayerConfigValues>({
  ac_enabled: 1,
  ac_severity_threshold: 1,
  small_model_template_id: '',
  small_model_config_key: '',
  large_model_template_id: '',
  large_model_config_key: '',
});

const escalationRules = reactive<EscalationRuleValues>({
  ac_to_small_condition: 'any_hit',
  ac_to_small_severity: 1,
  ac_to_small_categories: [],
  small_to_large_condition: 'confidence_lt',
  small_to_large_confidence_threshold: 0.90,
  small_to_large_categories: [],
  final_verdict_logic: 'last_model_wins',
});

const onPipelineSelect = (config: PipelineConfig) => {
  currentPipelineId.value = config.pipeline_id;
  Object.assign(layerConfig, {
    ac_enabled: config.ac_enabled,
    ac_severity_threshold: config.ac_severity_threshold,
    small_model_template_id: config.small_model_template_id,
    small_model_config_key: config.small_model_config_key,
    large_model_template_id: config.large_model_template_id,
    large_model_config_key: config.large_model_config_key,
  });
  Object.assign(escalationRules, {
    ac_to_small_condition: config.ac_to_small_condition,
    ac_to_small_severity: config.ac_to_small_severity,
    ac_to_small_categories: config.ac_to_small_categories || [],
    small_to_large_condition: config.small_to_large_condition,
    small_to_large_confidence_threshold: config.small_to_large_confidence_threshold,
    small_to_large_categories: config.small_to_large_categories || [],
    final_verdict_logic: config.final_verdict_logic,
  });
};

const onNewPipeline = () => {
  currentPipelineId.value = '';
  Object.assign(layerConfig, {
    ac_enabled: 1, ac_severity_threshold: 1,
    small_model_template_id: '', small_model_config_key: '',
    large_model_template_id: '', large_model_config_key: '',
  });
  Object.assign(escalationRules, {
    ac_to_small_condition: 'any_hit', ac_to_small_severity: 1,
    ac_to_small_categories: [],
    small_to_large_condition: 'confidence_lt',
    small_to_large_confidence_threshold: 0.90,
    small_to_large_categories: [],
    final_verdict_logic: 'last_model_wins',
  });
};

const onTest = async (content: string) => {
  testInputRef.value?.setLoading(true);
  testResult.value = null;
  try {
    const resp = await testPipeline({
      content,
      pipeline_id: currentPipelineId.value || undefined,
      small_model_template_id: layerConfig.small_model_template_id || undefined,
      large_model_template_id: layerConfig.large_model_template_id || undefined,
      small_to_large_confidence: escalationRules.small_to_large_confidence_threshold,
    });
    testResult.value = resp;
    ElMessage.success('测试完成');
  } catch (e: any) {
    ElMessage.error(e.message || '测试失败');
  } finally {
    testInputRef.value?.setLoading(false);
  }
};

const handleSave = async () => {
  if (!currentPipelineId.value) {
    // 新建
    const newId = `pipeline_${Date.now()}`;
    saving.value = true;
    try {
      await createPipeline({
        pipeline_id: newId,
        pipeline_name: `管线配置_${new Date().toLocaleDateString()}`,
        ac_enabled: layerConfig.ac_enabled,
        ac_severity_threshold: layerConfig.ac_severity_threshold,
        small_model_template_id: layerConfig.small_model_template_id,
        small_model_config_key: layerConfig.small_model_config_key,
        large_model_template_id: layerConfig.large_model_template_id,
        large_model_config_key: layerConfig.large_model_config_key,
        ac_to_small_condition: escalationRules.ac_to_small_condition,
        ac_to_small_severity: escalationRules.ac_to_small_severity,
        ac_to_small_categories: escalationRules.ac_to_small_categories,
        small_to_large_condition: escalationRules.small_to_large_condition,
        small_to_large_confidence_threshold: escalationRules.small_to_large_confidence_threshold,
        small_to_large_categories: escalationRules.small_to_large_categories,
        final_verdict_logic: escalationRules.final_verdict_logic,
      });
      currentPipelineId.value = newId;
      ElMessage.success('配置已保存');
      selectorRef.value?.loadPipelines();
    } catch (e: any) {
      ElMessage.error(e.message || '保存失败');
    } finally {
      saving.value = false;
    }
  } else {
    // 更新现有（需要先查询 id）
    saving.value = true;
    try {
      // 获取现有配置的 id
      const { getPipeline } = await import('@/api/moderation');
      const existing = await getPipeline(currentPipelineId.value);
      await updatePipeline({
        id: existing.id,
        pipeline_name: existing.pipeline_name,
        ac_enabled: layerConfig.ac_enabled,
        ac_severity_threshold: layerConfig.ac_severity_threshold,
        small_model_template_id: layerConfig.small_model_template_id,
        small_model_config_key: layerConfig.small_model_config_key,
        large_model_template_id: layerConfig.large_model_template_id,
        large_model_config_key: layerConfig.large_model_config_key,
        ac_to_small_condition: escalationRules.ac_to_small_condition,
        ac_to_small_severity: escalationRules.ac_to_small_severity,
        ac_to_small_categories: escalationRules.ac_to_small_categories,
        small_to_large_condition: escalationRules.small_to_large_condition,
        small_to_large_confidence_threshold: escalationRules.small_to_large_confidence_threshold,
        small_to_large_categories: escalationRules.small_to_large_categories,
        final_verdict_logic: escalationRules.final_verdict_logic,
      });
      ElMessage.success('配置已更新');
      selectorRef.value?.loadPipelines();
    } catch (e: any) {
      ElMessage.error(e.message || '更新失败');
    } finally {
      saving.value = false;
    }
  }
};
</script>

<style scoped>
.config-test-page {
  padding: 20px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

---

### Task 21: 最终验证

- [ ] **Step 1: 后端编译验证**

```bash
cd services/moderation-service && go build ./...
```

- [ ] **Step 2: 前端类型检查**

```bash
cd web/pc && npm run build
```

- [ ] **Step 3: 执行数据库迁移**

```bash
docker exec -i mysql mysql -uroot -proot123456 moderation_db < services/moderation-service/migrations/002_pipeline_config.sql
```

- [ ] **Step 4: 重启 moderation-service**

```bash
# 重启 API 和 RPC 服务以加载新路由
```

- [ ] **Step 5: 手动验证**

1. 打开配置测试页面 `/moderation/config-test`
2. 验证模板下拉列表加载
3. 新建一个管线配置，保存
4. 输入测试文本，点击执行测试
5. 验证 AC引擎/小模型/大模型 逐层结果显示
6. 修改升级规则，验证不同条件的行为
7. 设为生产配置，验证生效

---

### 可选：Task 22 — 生产 TextEngine 集成

当 `mod_pipeline_config` 表中有 `is_production=1` 的记录时，让 `TextEngine` 使用管线的升级规则替代硬编码阈值。

**Files:**
- Modify: `services/moderation-service/internal/engine/text_engine.go`
- Modify: `services/moderation-service/api/internal/svc/service_context.go`

- [ ] **Step 1: 在 TextEngine 中新增可选管线配置字段并修改 modelReview 方法**

在 `TextEngine` struct 中新增字段 `pipelineConfig *pipeline.PipelineConfig`。当 `pipelineConfig` 非空时，`modelReview` 方法使用管线配置中的升级阈值替代 `HighConfThreshold`。

- [ ] **Step 2: 工厂方法新增 WithPipelineConfig 选项或 SetPipelineConfig 方法**

- [ ] **Step 3: 在 ServiceContext 初始化时加载生产管线配置**

```go
// 尝试加载生产管线配置
prodConfig, err := pipelineModel.FindProduction(context.Background())
if err == nil {
    textEngine.SetPipelineConfig(pipeline.FromDBModel(prodConfig))
}
```
