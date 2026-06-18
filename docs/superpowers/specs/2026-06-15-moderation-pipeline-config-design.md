# 审核管线可配置化 — 设计文档

## 概述

将当前硬编码的 AC 引擎 → 小模型 → 大模型级联审核管线改造为**可配置、可测试、可落地**的管线配置系统。用户在"配置测试"页面上选择提示词模板、编辑升级规则、即时测试，测试通过的配置可保存并设为生产环境默认管线。

## 动机

- **现状问题**：提示词模板 ID 写死在 moderation-service YAML 配置中，升级阈值（`HighConfThreshold=0.9`）硬编码在 `text_engine.go`，用户无法调整
- **目标**：审核管线行为由配置驱动，非运营人员可自行调试和优化审核策略

## 数据模型

### 新表：`mod_pipeline_config`（moderation_db）

```sql
CREATE TABLE mod_pipeline_config (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    pipeline_id     VARCHAR(100) NOT NULL COMMENT '唯一标识，如 default_v1, strict_mode',
    pipeline_name   VARCHAR(200) NOT NULL COMMENT '显示名称，如"默认审核管线"',
    description     TEXT         COMMENT '描述',
    is_active       TINYINT      NOT NULL DEFAULT 1 COMMENT '1=启用 0=软删除',
    is_production   TINYINT      NOT NULL DEFAULT 0 COMMENT '1=生产环境使用',

    -- AC 引擎配置
    ac_enabled              TINYINT NOT NULL DEFAULT 1,
    ac_severity_threshold   INT     NOT NULL DEFAULT 1 COMMENT '最低触发严重度(1高/2中/3低)',

    -- 小模型配置
    small_model_template_id VARCHAR(100) COMMENT 'FK→ai_model_db.am_prompt_template.template_id',
    small_model_config_key  VARCHAR(64)  COMMENT '可选，覆盖模板默认的 config_key',

    -- 大模型配置
    large_model_template_id VARCHAR(100) COMMENT 'FK→ai_model_db.am_prompt_template.template_id',
    large_model_config_key  VARCHAR(64)  COMMENT '可选，覆盖模板默认的 config_key',

    -- 升级规则（AC → 小模型）
    ac_to_small_condition     VARCHAR(50) NOT NULL DEFAULT 'any_hit'
        COMMENT 'any_hit|severity_gte|category_in|never',
    ac_to_small_severity      INT         COMMENT 'severity_gte 时的阈值',
    ac_to_small_categories    JSON        COMMENT 'category_in 时的分类列表',

    -- 升级规则（小模型 → 大模型）
    small_to_large_condition             VARCHAR(50) NOT NULL DEFAULT 'confidence_lt'
        COMMENT 'confidence_lt|category_in|always|never',
    small_to_large_confidence_threshold  DECIMAL(3,2) DEFAULT 0.90,
    small_to_large_categories            JSON COMMENT 'category_in 时的分类列表',

    -- 终判逻辑
    final_verdict_logic VARCHAR(50) NOT NULL DEFAULT 'last_model_wins'
        COMMENT 'last_model_wins|large_overrides|small_overrides',

    created_time  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time   TIMESTAMP NULL DEFAULT NULL,

    UNIQUE INDEX idx_pipeline_id (pipeline_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审核管线配置';
```

### 升级条件枚举

| 条件值             | 适用层级                   | 含义               | 配套参数                                  |
| --------------- | ---------------------- | ---------------- | ------------------------------------- |
| `any_hit`       | AC→Small               | 只要 AC 命中任何敏感词即升级 | 无                                     |
| `severity_gte`  | AC→Small               | 命中词严重度 ≥ 阈值时升级   | `ac_to_small_severity`                |
| `category_in`   | AC→Small / Small→Large | 命中分类在指定列表中时升级    | `*_categories` (JSON)                 |
| `confidence_lt` | Small→Large            | 小模型置信度低于阈值时升级    | `small_to_large_confidence_threshold` |
| `always`        | Small→Large            | 无条件升级（每单都过大模型）   | 无                                     |
| `never`         | AC→Small / Small→Large | 永不升级（跳过该层）       | 无                                     |

### 终判逻辑枚举

| 值                 | 含义                    |
| ----------------- | --------------------- |
| `last_model_wins` | 最后一层有结果的模型判定为终判（默认）   |
| `large_overrides` | 大模型判定覆盖小模型（即使小模型已判合规） |
| `small_overrides` | 小模型判定覆盖大模型            |

## API 设计

### moderation-service 新增端点

| 方法       | 路径                                               | 说明               |
| -------- | ------------------------------------------------ | ---------------- |
| `POST`   | `/api/moderation/pipeline`                       | 创建管线配置           |
| `PUT`    | `/api/moderation/pipeline`                       | 更新管线配置           |
| `DELETE` | `/api/moderation/pipeline/:pipeline_id`          | 软删除              |
| `GET`    | `/api/moderation/pipeline/:pipeline_id`          | 获取单条详情           |
| `GET`    | `/api/moderation/pipelines`                      | 列表（分页）           |
| `PUT`    | `/api/moderation/pipeline/:pipeline_id/activate` | 设为生产配置           |
| `POST`   | `/api/moderation/pipeline/test`                  | 按管线配置执行审核，返回逐层结果 |

### POST /api/moderation/pipeline/test

请求体 `PipelineTestReq`：

```go
type PipelineTestReq struct {
    Content    string `json:"content"`                // 测试文本，必填
    PipelineID string `json:"pipeline_id,optional"`   // 管线ID，选填
    // ad-hoc 参数（pipeline_id 为空时使用）
    SmallModelTemplateID    string  `json:"small_model_template_id,optional"`
    LargeModelTemplateID    string  `json:"large_model_template_id,optional"`
    SmallToLargeConfidence  float64 `json:"small_to_large_confidence,optional"`
}
```

响应体 `PipelineTestResp`：

```go
type PipelineTestResp struct {
    PipelineID     string        `json:"pipeline_id"`
    AcResult       *LayerResult  `json:"ac_result"`
    SmallModelResult *LayerResult `json:"small_model_result"`
    LargeModelResult *LayerResult `json:"large_model_result"`
    FinalVerdict   string        `json:"final_verdict"`    // pass / reject / need_review
    TotalLatencyMs int           `json:"total_latency_ms"`
}

type LayerResult struct {
    Called        bool     `json:"called"`
    SkippedReason string   `json:"skipped_reason,omitempty"`
    Passed        bool     `json:"passed"`
    RiskLevel     string   `json:"risk_level"`
    Confidence    float64  `json:"confidence"`
    Categories    []string `json:"categories,omitempty"`
    Reason        string   `json:"reason,omitempty"`
    LatencyMs     int      `json:"latency_ms"`
    // AC 特有
    MatchedWords  []string `json:"matched_words,omitempty"`
    // 模型特有
    ModelUsed     string   `json:"model_used,omitempty"`
    TemplateID    string   `json:"template_id,omitempty"`
    RawResponse   string   `json:"raw_response,omitempty"`
}
```

### 聚合已有端点（前端直接调用 ai-model-service）

| 方法    | 路径                                      | 用途     |
| ----- | --------------------------------------- | ------ |
| `GET` | `/api/v1/templates?category=moderation` | 模板下拉列表 |
| `GET` | `/api/v1/models`                        | 模型下拉列表 |

### 后端执行流程（PipelineTestLogic）

```
POST /api/moderation/pipeline/test
  │
  ├─ 1. 解析管线配置
  │     pipeline_id 非空 → 从 DB 加载完整配置
  │     pipeline_id 为空 → 用 ad-hoc 参数 + 默认值拼装
  │
  ├─ 2. AC 引擎
  │     └─ Normalize → AC Match → Whitelist Filter → Split Detect
  │     返回 LayerResult{match_words, severity}
  │
  ├─ 3. 判断 ac_to_small_condition
  │     ├─ "never"          → 跳过小模型
  │     ├─ "any_hit"        → 有命中即升级
  │     ├─ "severity_gte"   → severity >= 阈值 升级
  │     └─ "category_in"    → 命中词分类在列表中 升级
  │
  ├─ 4. 小模型
  │     └─ gRPC → ai-model-service.ModerateText(
  │           template_id,
  │           config_key (可选覆盖),
  │           content
  │        )
  │     返回 LayerResult{passed, confidence, categories, reason}
  │
  ├─ 5. 判断 small_to_large_condition
  │     ├─ "never"          → 跳过，小模型结果=终判
  │     ├─ "always"         → 无条件升级
  │     ├─ "confidence_lt"  → confidence < 阈值 升级
  │     └─ "category_in"    → 小模型输出分类在列表中 升级
  │
  ├─ 6. 大模型
  │     └─ gRPC → ai-model-service.ModerateText(...)
  │     返回 LayerResult{passed, confidence, categories, reason}
  │
  └─ 7. 终判 (final_verdict_logic)
        ├─ "last_model_wins"  → 最后一层结果
        ├─ "large_overrides"  → 大模型判定覆盖
        └─ "small_overrides"  → 小模型判定覆盖
```

## 前端设计

### 页面改造

将现有 `/moderation/test`（ModerationTest.vue）升级为 **配置测试工作台**，路由改为 `/moderation/config-test`。

### 页面布局

```
┌────────────────────────────────────────────────────┐
│  内容审核配置测试                          [保存配置] │
├────────────────────────────────────────────────────┤
│  ┌─ 管线配置选择 ────────────────────────────────┐ │
│  │ [默认审核管线 ▼]  [新建] [复制] [设为生产]      │ │
│  └────────────────────────────────────────────────┘ │
│                                                      │
│  ┌─ 审核层配置（三列布局）───────────────────────┐  │
│  │  AC 引擎         小模型           大模型         │  │
│  │  ☑启用           ☑启用            ☑启用         │  │
│  │  严重度≥[1]      模板下拉         模板下拉       │  │
│  │                  模型下拉         模型下拉       │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ┌─ 升级规则 ───────────────────────────────────┐  │
│  │  AC→小模型: [任何命中 ▼]  阈值/分类动态出现     │  │
│  │  小→大模型: [置信度< ▼]  [0.90]                 │  │
│  │  终判逻辑:  [最后模型判定 ▼]                     │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ┌─ 测试区域 ───────────────────────────────────┐  │
│  │  文本输入框（≤500字）    [执行测试] [重置]      │  │
│  └────────────────────────────────────────────────┘  │
│                                                      │
│  ┌─ 执行结果（三层并排卡片）─────────────────────┐  │
│  │  AC引擎结果   小模型结果    大模型结果           │  │
│  │  调用状态     调用状态      调用状态             │  │
│  │  命中词列表   置信度/分类   置信度/分类          │  │
│  │  耗时         耗时/模型名   耗时/模型名          │  │
│  │  [复制JSON]   [复制JSON]    [复制JSON]          │  │
│  │                                                  │  │
│  │                终判: ❌/✅                       │  │
│  └────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────┘
```

### 组件拆分

```
ModerationConfigTest.vue          (页面容器)
├── PipelineSelector.vue          (管线选择器)
├── LayerConfigPanel.vue          (三层配置面板)
│   ├── AcEngineConfig.vue
│   ├── SmallModelConfig.vue
│   └── LargeModelConfig.vue
├── EscalationRuleEditor.vue      (升级规则编辑器)
├── TestInputArea.vue             (测试文本+执行按钮)
└── PipelineResultPanel.vue       (三层结果并排展示)
```

### 关键交互

- **模板下拉**：调用 `GET /api/v1/templates?category=moderation`，选项格式 `{template_name} v{version}`，值 `template_id`
- **模型下拉**：调用 `GET /api/v1/models`，过滤 `health_status=healthy && status=1`
- **升级规则联动**：选择不同条件类型时，动态显示对应参数输入（阈值数字框 / 分类标签选择器 / 无参数）
- **设为生产**：二次确认弹窗 → `PUT /api/moderation/pipeline/:id/activate`（后端将其他 `is_production` 清 0，当前设 1）

### 路由和菜单

- 路由路径：`/moderation/config-test`（原名 `/moderation/test`）
- 菜单位置：内容审核 → 配置测试
- 菜单图标：保持 `Monitor`

## 生产环境集成

当 `mod_pipeline_config` 中存在 `is_production=1` 的记录时：

- `moderation-service` 启动时从 DB 加载生产管线配置
- `TextEngine.Check()` 不再使用硬编码的 `HighConfThreshold`，改为读取管线配置中的升级规则
- 若不存在生产配置，fallback 到现有硬编码默认值（向后兼容）

## 涉及服务

| 服务                 | 改动范围                                                                |
| ------------------ | ------------------------------------------------------------------- |
| moderation-service | 新增 pipeline CRUD 端点 + pipeline/test 端点 + Go 数据模型 + PipelineExecutor |
| ai-model-service   | 无需改动（复用已有 template/model 查询端点）                                      |
| 前端 (web/pc)        | 改造 ModerationTest.vue → ModerationConfigTest.vue + 5 个新组件 + 路由/菜单更新 |
