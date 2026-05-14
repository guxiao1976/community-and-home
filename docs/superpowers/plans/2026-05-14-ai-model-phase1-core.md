# AI 模型统一管理服务 - Phase 1 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 AI 模型统一管理服务的核心功能，支持本地 Qwen 和 Claude 模型调用，提供基础的配置管理和日志记录。

**Architecture:** 采用渐进式改造方案，保持现有 Python 引擎独立运行，通过适配器模式统一管理内部和外部模型。使用双层接口设计（高层业务接口 + 底层通用接口），调用方指定模型，服务负责路由和成本管理。

**Tech Stack:** Go 1.21+, go-zero 1.6+, gRPC, MySQL 8.0, Redis 7.0, etcd 3.5, anthropic-sdk-go

---

## 文件结构规划

### 新建目录结构
```
services/ai-model-unified/
├── rpc/                          # gRPC 服务（端口 8084）
│   ├── ai_model.go              # 主入口
│   ├── etc/
│   │   └── ai-model.yaml        # 配置文件
│   ├── pb/
│   │   ├── ai_model.proto       # protobuf 定义
│   │   ├── ai_model.pb.go       # 生成的代码
│   │   └── ai_model_grpc.pb.go  # 生成的代码
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go        # 配置结构
│   │   ├── svc/
│   │   │   └── service_context.go  # 服务上下文
│   │   ├── server/
│   │   │   └── ai_model_server.go  # gRPC 服务器
│   │   ├── logic/
│   │   │   ├── moderate_text_logic.go      # 文本审核逻辑
│   │   │   ├── classify_words_logic.go     # 敏感词分类逻辑
│   │   │   ├── call_model_logic.go         # 通用调用逻辑
│   │   │   ├── get_models_logic.go         # 获取模型列表逻辑
│   │   │   └── health_check_logic.go       # 健康检查逻辑
│   │   ├── adapter/
│   │   │   ├── interface.go     # 适配器接口定义
│   │   │   ├── local_qwen.go    # 本地 Qwen 适配器
│   │   │   └── claude.go        # Claude 适配器
│   │   └── manager/
│   │       ├── model_manager.go  # 模型管理器
│   │       ├── cost_manager.go   # 成本管理器
│   │       └── template_manager.go  # 模板管理器
│   └── model/                    # 数据模型（goctl 生成）
│       ├── am_model_config_model.go
│       ├── am_health_check_model.go
│       ├── am_call_log_model.go
│       ├── am_prompt_template_model.go
│       ├── am_usage_statistics_model.go
│       ├── am_cost_alert_config_model.go
│       ├── am_alert_record_model.go
│       └── am_api_key_model.go
│
├── api/                          # HTTP API 服务（端口 8891）
│   ├── ai_model_api.go          # 主入口
│   ├── etc/
│   │   └── ai-model-api.yaml    # 配置文件
│   ├── ai_model.api             # API 定义
│   └── internal/
│       ├── config/
│       │   └── config.go
│       ├── svc/
│       │   └── service_context.go
│       ├── handler/
│       │   └── model_config/    # 模型配置管理 handlers
│       │       ├── get_model_configs_handler.go
│       │       ├── create_model_config_handler.go
│       │       ├── update_model_config_handler.go
│       │       ├── delete_model_config_handler.go
│       │       ├── test_model_connection_handler.go
│       │       └── toggle_model_status_handler.go
│       ├── logic/
│       │   └── model_config/    # 对应的 logic 层
│       │       ├── get_model_configs_logic.go
│       │       ├── create_model_config_logic.go
│       │       ├── update_model_config_logic.go
│       │       ├── delete_model_config_logic.go
│       │       ├── test_model_connection_logic.go
│       │       └── toggle_model_status_logic.go
│       └── types/
│           └── types.go          # API 类型定义（goctl 生成）
│
└── sql/
    └── ai_model_db.sql          # 数据库初始化脚本
```

---

## Task 1: 数据库设计与创建

**Files:**
- Create: `services/ai-model-unified/sql/ai_model_db.sql`

- [ ] **Step 1: 编写数据库创建脚本**

创建文件 `services/ai-model-unified/sql/ai_model_db.sql`：

```sql
-- 创建数据库
CREATE DATABASE IF NOT EXISTS ai_model_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai_model_db;

-- 1. 模型配置表
CREATE TABLE am_model_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_name VARCHAR(100) NOT NULL UNIQUE COMMENT '模型标识：qwen/claude-3-5-sonnet/gpt-4',
    model_type VARCHAR(20) NOT NULL COMMENT '模型类型：local/cloud',
    provider VARCHAR(50) NOT NULL COMMENT '提供商：qwen/claude/openai/ollama',
    display_name VARCHAR(200) COMMENT '显示名称',
    description TEXT COMMENT '模型描述',
    capabilities JSON COMMENT '能力列表：["text_moderation","classification"]',
    
    -- 连接配置
    endpoint VARCHAR(500) COMMENT 'API端点或Python引擎地址',
    api_key_id BIGINT COMMENT '关联的API Key ID（外部模型）',
    timeout INT DEFAULT 30000 COMMENT '超时时间(ms)',
    max_retries INT DEFAULT 2 COMMENT '最大重试次数',
    
    -- 成本配置
    cost_per_input_token DECIMAL(10, 6) DEFAULT 0 COMMENT '每1K输入token成本',
    cost_per_output_token DECIMAL(10, 6) DEFAULT 0 COMMENT '每1K输出token成本',
    
    -- 状态
    status TINYINT DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
    health_status VARCHAR(20) DEFAULT 'unknown' COMMENT '健康状态：healthy/degraded/unhealthy/unknown',
    
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time TIMESTAMP NULL DEFAULT NULL,
    
    INDEX idx_provider (provider),
    INDEX idx_status (status),
    INDEX idx_delete_time (delete_time)
) COMMENT '模型配置表';

-- 2. API密钥管理表
CREATE TABLE am_api_key (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    key_name VARCHAR(100) NOT NULL COMMENT '密钥名称',
    provider VARCHAR(50) NOT NULL COMMENT '提供商：claude/openai/ollama',
    api_key VARCHAR(500) NOT NULL COMMENT 'API密钥（AES加密存储）',
    masked_key VARCHAR(50) COMMENT '脱敏显示：sk-***abc123',
    
    -- 配额和限制
    daily_quota INT DEFAULT 0 COMMENT '每日配额（0=无限制）',
    monthly_quota INT DEFAULT 0 COMMENT '每月配额（0=无限制）',
    priority INT DEFAULT 0 COMMENT '优先级：数字越大优先级越高',
    
    -- 状态
    status TINYINT DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
    failure_count INT DEFAULT 0 COMMENT '连续失败次数',
    last_used_time TIMESTAMP NULL COMMENT '最后使用时间',
    last_failure_time TIMESTAMP NULL COMMENT '最后失败时间',
    
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time TIMESTAMP NULL DEFAULT NULL,
    
    INDEX idx_provider (provider),
    INDEX idx_status (status),
    INDEX idx_priority (priority DESC)
) COMMENT 'API密钥管理表';

-- 3. 健康检查记录表
CREATE TABLE am_health_check (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL COMMENT 'healthy/degraded/unhealthy',
    latency BIGINT COMMENT '响应时间(ms)',
    error_msg TEXT COMMENT '错误信息',
    check_time TIMESTAMP NOT NULL,
    
    INDEX idx_model_time (model_name, check_time DESC),
    INDEX idx_check_time (check_time DESC)
) COMMENT '健康检查记录表';

-- 4. 调用日志表
CREATE TABLE am_call_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_name VARCHAR(100) NOT NULL,
    caller VARCHAR(100) COMMENT '调用方服务名',
    request_type VARCHAR(50) COMMENT '请求类型：moderate_text/classify_words/call_model',
    
    -- Token和成本
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    cost DECIMAL(10, 6) DEFAULT 0 COMMENT '本次调用成本',
    
    -- 性能
    latency BIGINT COMMENT '耗时(ms)',
    success TINYINT DEFAULT 1 COMMENT '是否成功：1=成功，0=失败',
    error_msg TEXT COMMENT '错误信息',
    
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_model_time (model_name, created_time DESC),
    INDEX idx_caller (caller),
    INDEX idx_created_time (created_time DESC)
) COMMENT '调用日志表';

-- 5. 使用统计表（按天汇总）
CREATE TABLE am_usage_statistics (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_name VARCHAR(100) NOT NULL,
    stat_date DATE NOT NULL COMMENT '统计日期',
    
    -- 调用统计
    total_calls INT DEFAULT 0,
    success_calls INT DEFAULT 0,
    failed_calls INT DEFAULT 0,
    
    -- Token统计
    total_tokens BIGINT DEFAULT 0,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0,
    
    -- 成本统计
    total_cost DECIMAL(10, 2) DEFAULT 0,
    
    -- 性能统计
    avg_latency BIGINT DEFAULT 0 COMMENT '平均耗时(ms)',
    
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_model_date (model_name, stat_date),
    INDEX idx_stat_date (stat_date DESC)
) COMMENT '使用统计表';

-- 6. 提示词模板表
CREATE TABLE am_prompt_template (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    template_id VARCHAR(100) NOT NULL UNIQUE COMMENT '模板ID',
    template_name VARCHAR(200) NOT NULL COMMENT '模板名称',
    category VARCHAR(50) COMMENT '分类：moderation/classification/qa',
    content TEXT NOT NULL COMMENT '模板内容，支持{{variable}}变量',
    variables JSON COMMENT '变量列表：["words","categories"]',
    
    -- 版本管理
    version VARCHAR(20) DEFAULT 'v1.0',
    is_active TINYINT DEFAULT 1 COMMENT '是否激活：1=激活，0=未激活',
    
    -- 使用统计
    usage_count INT DEFAULT 0 COMMENT '使用次数',
    
    description TEXT COMMENT '模板说明',
    
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time TIMESTAMP NULL DEFAULT NULL,
    
    INDEX idx_category (category),
    INDEX idx_active (is_active)
) COMMENT '提示词模板表';

-- 7. 成本预警配置表
CREATE TABLE am_cost_alert_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    daily_limit DECIMAL(10, 2) DEFAULT 0 COMMENT '日成本上限（0=不限制）',
    monthly_limit DECIMAL(10, 2) DEFAULT 0 COMMENT '月成本上限（0=不限制）',
    alert_threshold DECIMAL(5, 2) DEFAULT 0.8 COMMENT '预警阈值：0.8表示达到80%时预警',
    
    -- 通知配置
    enable_notification TINYINT DEFAULT 1 COMMENT '是否启用通知',
    notification_channels JSON COMMENT '通知渠道：["dingtalk","email"]',
    
    status TINYINT DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
    
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT '成本预警配置表';

-- 8. 预警记录表
CREATE TABLE am_alert_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    alert_type VARCHAR(20) NOT NULL COMMENT '预警类型：daily/monthly',
    threshold DECIMAL(10, 2) NOT NULL COMMENT '阈值',
    actual_value DECIMAL(10, 2) NOT NULL COMMENT '实际值',
    message TEXT COMMENT '预警消息',
    
    -- 处理状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态：pending/notified/resolved',
    notified_time TIMESTAMP NULL COMMENT '通知时间',
    resolved_time TIMESTAMP NULL COMMENT '解决时间',
    
    alert_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_alert_time (alert_time DESC),
    INDEX idx_status (status)
) COMMENT '预警记录表';

-- 插入初始数据
-- 本地 Qwen 模型配置
INSERT INTO am_model_config (
    model_name, model_type, provider, display_name, description,
    capabilities, endpoint, cost_per_input_token, cost_per_output_token, status
) VALUES (
    'qwen', 'local', 'qwen', 'Qwen2.5-3B-Instruct', '本地部署的千问模型，用于内容审核',
    '["text_moderation", "classification"]', 'http://localhost:8001', 0.00002, 0.00002, 1
);

-- Claude 模型配置（需要配置 API Key）
INSERT INTO am_model_config (
    model_name, model_type, provider, display_name, description,
    capabilities, endpoint, cost_per_input_token, cost_per_output_token, status
) VALUES (
    'claude-3-5-sonnet', 'cloud', 'claude', 'Claude 3.5 Sonnet', 'Anthropic Claude 3.5 Sonnet 模型',
    '["text_generation", "text_moderation", "classification", "qa"]', 
    'https://api.anthropic.com', 0.003, 0.015, 0
);

-- 默认成本预警配置
INSERT INTO am_cost_alert_config (daily_limit, monthly_limit, alert_threshold) 
VALUES (100.00, 3000.00, 0.8);

-- 敏感词分类模板
INSERT INTO am_prompt_template (
    template_id, template_name, category, content, variables, version
) VALUES (
    'sensitive_word_classify', '敏感词分类模板', 'classification',
    '请对以下敏感词进行分类，分类包括：政治敏感、色情低俗、暴力恐怖、违法犯罪、人身攻击、广告营销、其他。

要求：
1. 严格按照JSON数组格式返回
2. 每个词必须分配一个分类
3. 格式：[{"word": "词1", "category": "分类1", "confidence": 0.95}, ...]

敏感词列表：
{{words}}',
    '["words"]', 'v1.0'
);
```

- [ ] **Step 2: 执行数据库脚本**

```bash
mysql -u root -p < services/ai-model-unified/sql/ai_model_db.sql
```

预期输出：数据库和表创建成功，无错误

- [ ] **Step 3: 验证数据库创建**

```bash
mysql -u root -p -e "USE ai_model_db; SHOW TABLES;"
```

预期输出：显示 8 张表

- [ ] **Step 4: 验证初始数据**

```bash
mysql -u root -p -e "USE ai_model_db; SELECT model_name, provider, status FROM am_model_config;"
```

预期输出：显示 qwen 和 claude-3-5-sonnet 两条记录

- [ ] **Step 5: 提交**

```bash
git add services/ai-model-unified/sql/ai_model_db.sql
git commit -m "feat(ai-model): 创建数据库表结构和初始数据

- 创建 ai_model_db 数据库
- 创建 8 张表：模型配置、API密钥、健康检查、调用日志、统计、模板、预警配置、预警记录
- 插入初始数据：Qwen 和 Claude 模型配置、默认预警配置、敏感词分类模板

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: RPC 服务框架搭建

**Files:**
- Create: `services/ai-model-unified/rpc/pb/ai_model.proto`
- Create: `services/ai-model-unified/rpc/etc/ai-model.yaml`
- Create: `services/ai-model-unified/rpc/internal/config/config.go`

- [ ] **Step 1: 创建目录结构**

```bash
mkdir -p services/ai-model-unified/rpc/{pb,etc,internal/{config,svc,server,logic,adapter,manager},model}
```

- [ ] **Step 2: 编写 protobuf 定义**

创建 `services/ai-model-unified/rpc/pb/ai_model.proto`：

```protobuf
syntax = "proto3";

package ai_model;
option go_package = "./pb";

// AI 模型统一管理服务
service AiModel {
  // ========== 高层业务接口 ==========
  
  // 文本审核
  rpc ModerateText(TextModerationRequest) returns (TextModerationResponse);
  
  // 敏感词分类（批量）
  rpc ClassifySensitiveWords(SensitiveWordsRequest) returns (SensitiveWordsResponse);
  
  // ========== 底层通用接口 ==========
  
  // 通用模型调用（单次）
  rpc CallModel(ModelCallRequest) returns (ModelCallResponse);
  
  // 通用模型调用（批量）
  rpc CallModelBatch(ModelBatchRequest) returns (ModelBatchResponse);
  
  // 获取可用模型列表
  rpc GetAvailableModels(GetModelsRequest) returns (GetModelsResponse);
  
  // 健康检查
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

// ============ 文本审核 ============
message TextModerationRequest {
  string content = 1;                    // 待审核文本
  string model = 2;                      // 指定模型：qwen/claude-3-5-sonnet
  repeated string check_categories = 3;  // 检查维度（可选）
  string context = 4;                    // 上下文（可选）
}

message TextModerationResponse {
  bool is_safe = 1;
  string risk_level = 2;                 // high/medium/low/safe
  repeated string categories = 3;
  string reason = 4;
  double confidence = 5;
  int64 latency_ms = 6;
  string model_used = 7;
  double cost = 8;
}

// ============ 敏感词分类 ============
message SensitiveWordsRequest {
  repeated string words = 1;
  string model = 2;
  string template_id = 3;
  map<string, string> variables = 4;
}

message SensitiveWordsResponse {
  repeated WordClassification results = 1;
  int32 total = 2;
  int32 success = 3;
  int32 failed = 4;
  double total_cost = 5;
  int64 latency_ms = 6;
}

message WordClassification {
  string word = 1;
  string category = 2;
  double confidence = 3;
  bool success = 4;
  string error = 5;
}

// ============ 通用模型调用 ============
message ModelCallRequest {
  string model = 1;
  string prompt = 2;
  map<string, string> parameters = 3;
  string template_id = 4;
  map<string, string> variables = 5;
}

message ModelCallResponse {
  string content = 1;
  int32 input_tokens = 2;
  int32 output_tokens = 3;
  double cost = 4;
  int64 latency_ms = 5;
  string model_used = 6;
}

// ============ 批量调用 ============
message ModelBatchRequest {
  string model = 1;
  repeated string prompts = 2;
  map<string, string> parameters = 3;
}

message ModelBatchResponse {
  repeated ModelCallResponse results = 1;
  int32 total = 2;
  int32 success = 3;
  int32 failed = 4;
  double total_cost = 5;
}

// ============ 获取模型列表 ============
message GetModelsRequest {}

message GetModelsResponse {
  repeated ModelInfo models = 1;
}

message ModelInfo {
  string name = 1;
  string type = 2;
  string provider = 3;
  repeated string capabilities = 4;
  double cost_per_input_token = 5;
  double cost_per_output_token = 6;
  string health_status = 7;
}

// ============ 健康检查 ============
message HealthCheckRequest {}

message HealthCheckResponse {
  string status = 1;
  map<string, ModelHealthStatus> models = 2;
  int64 uptime_seconds = 3;
}

message ModelHealthStatus {
  string name = 1;
  string status = 2;
  int64 avg_latency_ms = 3;
  int64 total_requests = 4;
}
```

- [ ] **Step 3: 生成 protobuf 代码**

```bash
cd services/ai-model-unified/rpc
protoc --go_out=. --go-grpc_out=. pb/ai_model.proto
```

预期输出：生成 `pb/ai_model.pb.go` 和 `pb/ai_model_grpc.pb.go`

- [ ] **Step 4: 编写配置文件**

创建 `services/ai-model-unified/rpc/etc/ai-model.yaml`：

```yaml
Name: ai-model.rpc
ListenOn: 0.0.0.0:8084

# etcd 服务注册
Etcd:
  Hosts:
    - localhost:2379
  Key: ai-model.rpc

# MySQL 配置
DataSource: root:password@tcp(localhost:3306)/ai_model_db?charset=utf8mb4&parseTime=true&loc=Local

# Redis 配置
CacheRedis:
  - Host: localhost:6379
    Type: node

# 日志配置
Log:
  ServiceName: ai-model-rpc
  Mode: console
  Level: info
```

- [ ] **Step 5: 编写配置结构**

创建 `services/ai-model-unified/rpc/internal/config/config.go`：

```go
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	CacheRedis []struct {
		Host string
		Type string
	}
}
```

- [ ] **Step 6: 提交**

```bash
git add services/ai-model-unified/rpc/
git commit -m "feat(ai-model): 创建 RPC 服务框架

- 定义 protobuf 接口（高层业务接口 + 底层通用接口）
- 生成 protobuf 代码
- 配置文件和配置结构

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

由于实施计划非常长，我将其分为多个文档。这是 Phase 1 的前两个任务。

计划已保存到 `docs/superpowers/plans/2026-05-14-ai-model-phase1-core.md`。

**两种执行选项：**

**1. Subagent-Driven（推荐）** - 我为每个任务派发一个新的子代理，任务间进行审查，快速迭代

**2. Inline Execution** - 在当前会话中使用 executing-plans 执行任务，批量执行并设置检查点

你选择哪种方式？
