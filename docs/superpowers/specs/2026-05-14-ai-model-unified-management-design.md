# AI 模型统一管理服务 - 设计方案

**日期**: 2026-05-14  
**版本**: v1.0  
**状态**: 已批准

## 1. 概述

### 1.1 背景

社区平台需要集成大语言模型（LLM）能力来实现智能化功能，如敏感词自动分类、内容智能审核、用户问题自动回复等。当前存在以下问题：

1. **现有 AI-Model 服务**：专注于内容审核，使用本地 Qwen2.5-3B 模型（GPU 推理）
2. **新需求**：需要调用外部 AI API（Claude、OpenAI）实现更多场景
3. **管理分散**：缺少统一的配置管理、成本监控、健康检查

### 1.2 目标

创建统一的 AI 模型管理服务，实现：

- **统一管理**：内部模型（Qwen）和外部模型（Claude/OpenAI/Ollama）统一管理
- **双层接口**：高层业务接口（ModerateText）+ 底层通用接口（CallModel）
- **调用方指定模型**：调用方明确指定使用哪个模型
- **统一成本核算**：本地模型按资源成本估算，外部 API 按实际费用
- **完整管理界面**：模型配置、健康监控、成本统计、提示词模板、API Key 管理

### 1.3 非目标

- 不实现模型训练和微调功能
- 不实现向量数据库和 RAG 功能（后续独立服务）
- 不实现流式响应（SSE）功能（Phase 1 仅支持同步调用）
- 不实现复杂的提示词工程工具（仅基础模板管理）

---

## 2. 整体架构

### 2.1 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI-Model Service (统一服务)                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────┐              ┌──────────────────┐         │
│  │  ai-model-api    │              │  ai-model-rpc    │         │
│  │  (HTTP: 8891)    │              │  (gRPC: 8084)    │         │
│  │                  │              │                  │         │
│  │  管理界面后端     │              │  业务调用接口     │         │
│  └────────┬─────────┘              └────────┬─────────┘         │
│           │                                 │                   │
│           └─────────────┬───────────────────┘                   │
│                         │                                       │
│              ┌──────────▼──────────┐                            │
│              │   Business Layer    │                            │
│              ├─────────────────────┤                            │
│              │ ModelManager        │ 模型配置、选择、健康检查    │
│              │ CostManager         │ 成本计算、统计、预警        │
│              │ TemplateManager     │ 提示词模板管理             │
│              └──────────┬──────────┘                            │
│                         │                                       │
│              ┌──────────▼──────────┐                            │
│              │   Adapter Layer     │                            │
│              ├─────────────────────┤                            │
│              │ LocalQwenAdapter    │ → Python引擎(8001)         │
│              │ ClaudeAdapter       │ → Claude API               │
│              │ OpenAIAdapter       │ → OpenAI API (Phase 3)     │
│              │ OllamaAdapter       │ → Ollama (Phase 3)         │
│              └─────────────────────┘                            │
│                                                                   │
└───────────────────────────┬───────────────────────────────────────┘
                            │
                ┌───────────▼───────────┐
                │  ai_model_db (MySQL)  │
                ├───────────────────────┤
                │ am_model_config       │ 模型配置
                │ am_health_check       │ 健康检查记录
                │ am_call_log           │ 调用日志
                │ am_prompt_template    │ 提示词模板
                │ am_usage_statistics   │ 使用统计
                │ am_cost_alert_config  │ 成本预警配置
                │ am_alert_record       │ 预警记录
                │ am_api_key            │ API密钥管理
                └───────────────────────┘

外部依赖：
┌──────────────────────┐
│ Python Engine (8001) │ ← 现有服务，保持独立
│ - Qwen2.5-3B        │
│ - GPU 推理          │
└──────────────────────┘

┌──────────────────────┐
│ External AI APIs     │
│ - Claude API         │
│ - OpenAI API         │
│ - Ollama (本地)      │
└──────────────────────┘
```

### 2.2 关键设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| **服务范围** | 统一管理所有 AI 能力 | 提供统一接口、配置、监控 |
| **接口设计** | 双层接口（高层+底层） | 高层封装业务逻辑，底层提供灵活性 |
| **本地模型集成** | Python 引擎作为 Adapter | 改动最小，风险低 |
| **模型选择** | 调用方指定模型 | 调用方完全控制，服务只负责路由 |
| **成本管理** | 统一成本核算 | 本地按资源估算，外部按实际费用 |
| **数据库** | 独立数据库 ai_model_db | 遵循微服务数据独立原则 |
| **前端界面** | 完整管理界面 | 配置、监控、统计、模板、Key 管理 |
| **实施策略** | 分三个阶段 | 降低风险，每阶段可独立交付 |

---

## 3. RPC 接口设计

### 3.1 服务定义

```protobuf
service AiModel {
  // ========== 高层业务接口 ==========
  
  // 文本审核（封装业务逻辑）
  rpc ModerateText(TextModerationRequest) returns (TextModerationResponse);
  
  // 敏感词分类（批量）
  rpc ClassifySensitiveWords(SensitiveWordsRequest) returns (SensitiveWordsResponse);
  
  // 问答（预留）
  rpc AnswerQuestion(QuestionRequest) returns (QuestionResponse);
  
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
```

### 3.2 核心消息定义

```protobuf
// 文本审核请求
message TextModerationRequest {
  string content = 1;                    // 待审核文本
  string model = 2;                      // 指定模型：qwen/claude/openai
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
  string model_used = 7;                 // 实际使用的模型
  double cost = 8;                       // 本次调用成本
}

// 敏感词分类请求
message SensitiveWordsRequest {
  repeated string words = 1;             // 待分类的敏感词列表
  string model = 2;                      // 指定模型
  string template_id = 3;                // 提示词模板ID（可选）
  map<string, string> variables = 4;     // 模板变量（可选）
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
  string category = 2;                   // 分类结果
  double confidence = 3;
  bool success = 4;
  string error = 5;                      // 失败原因
}

// 通用模型调用请求
message ModelCallRequest {
  string model = 1;                      // 模型标识：qwen/claude-3-5-sonnet/gpt-4
  string prompt = 2;                     // 提示词
  map<string, string> parameters = 3;    // 模型参数（temperature等）
  string template_id = 4;                // 使用模板（可选）
  map<string, string> variables = 5;     // 模板变量
}

message ModelCallResponse {
  string content = 1;                    // 模型返回内容
  int32 input_tokens = 2;
  int32 output_tokens = 3;
  double cost = 4;
  int64 latency_ms = 5;
  string model_used = 6;
}
```

---

## 4. 适配器层设计

### 4.1 统一适配器接口

```go
type ModelAdapter interface {
    // 调用模型
    Call(ctx context.Context, req *CallRequest) (*CallResponse, error)
    
    // 健康检查
    HealthCheck(ctx context.Context) error
    
    // 估算成本（调用前）
    EstimateCost(req *CallRequest) (float64, error)
    
    // 获取模型信息
    GetModelInfo() *ModelInfo
}

type CallRequest struct {
    Prompt      string
    Parameters  map[string]interface{}
    Timeout     time.Duration
}

type CallResponse struct {
    Content      string
    InputTokens  int32
    OutputTokens int32
    Cost         float64
    Latency      time.Duration
    ModelVersion string
}

type ModelInfo struct {
    Name         string
    Type         string   // local/cloud
    Provider     string   // qwen/claude/openai
    Capabilities []string
    CostPerInputToken  float64
    CostPerOutputToken float64
}
```

### 4.2 LocalQwenAdapter

- **功能**：通过 HTTP 调用现有 Python 引擎（8001 端口）
- **成本计算**：基于 GPU 时间估算（RTX 5060 功耗 120W * 电费 ¥0.6/kWh）
- **健康检查**：调用 Python 引擎的 /health 接口

### 4.3 ClaudeAdapter

- **功能**：调用 Claude API（anthropic-sdk-go）
- **成本计算**：按实际 Token 计费（输入 $3/MTok，输出 $15/MTok）
- **健康检查**：发送简单测试请求

### 4.4 OpenAIAdapter（Phase 3）

- **功能**：调用 OpenAI API（openai-go）
- **成本计算**：按实际 Token 计费
- **健康检查**：发送简单测试请求

### 4.5 OllamaAdapter（Phase 3）

- **功能**：调用本地 Ollama（ollama-go）
- **成本计算**：按资源成本估算
- **健康检查**：检查 Ollama 服务状态

---

## 5. 业务层设计

### 5.1 ModelManager（模型管理器）

**职责**：
- 管理所有模型适配器
- 根据模型名称返回对应的适配器
- 健康检查调度（每 5 分钟）
- 健康状态评估（Healthy/Degraded/Unhealthy）
- 获取可用模型列表

**健康状态判断**：
- Healthy：健康检查成功，响应时间 < 5s
- Degraded：健康检查成功，响应时间 >= 5s
- Unhealthy：健康检查失败

### 5.2 CostManager（成本管理器）

**职责**：
- 记录每次调用日志
- 更新统计数据（按天汇总）
- 检查成本预警（日/月）
- 触发预警通知

**成本预警逻辑**：
- 达到阈值（默认 80%）时触发预警
- 记录预警记录
- 发送通知（日志记录，预留钉钉/邮件接口）

### 5.3 TemplateManager（提示词模板管理器）

**职责**：
- 获取模板（带 Redis 缓存）
- 渲染模板（变量替换）
- 验证模板变量完整性

**模板格式**：
```
请对以下敏感词进行分类，分类包括：{{categories}}

敏感词列表：
{{words}}
```

---

## 6. 数据库设计

### 6.1 表结构

**am_model_config（模型配置表）**
- 存储所有模型（本地+外部）的配置信息
- 包含连接配置、成本配置、健康状态

**am_api_key（API密钥管理表）**
- 加密存储外部 API 密钥（AES-256）
- 支持优先级和配额管理
- 失效检测和自动切换

**am_health_check（健康检查记录表）**
- 记录每次健康检查结果
- 用于健康历史查询和趋势分析

**am_call_log（调用日志表）**
- 记录每次调用的详细信息
- 用于审计、分析、成本统计

**am_usage_statistics（使用统计表）**
- 按天汇总调用数据
- 减少查询压力，提高统计效率

**am_prompt_template（提示词模板表）**
- 存储可复用的提示词模板
- 支持变量替换和版本管理

**am_cost_alert_config（成本预警配置表）**
- 配置日/月成本上限
- 配置预警阈值和通知渠道

**am_alert_record（预警记录表）**
- 记录所有预警事件
- 跟踪处理状态

### 6.2 初始数据

```sql
-- 本地 Qwen 模型
INSERT INTO am_model_config (model_name, model_type, provider, display_name, 
    endpoint, cost_per_input_token, cost_per_output_token, status)
VALUES ('qwen', 'local', 'qwen', 'Qwen2.5-3B-Instruct', 
    'http://localhost:8001', 0.00002, 0.00002, 1);

-- Claude 模型（需配置 API Key）
INSERT INTO am_model_config (model_name, model_type, provider, display_name,
    endpoint, cost_per_input_token, cost_per_output_token, status)
VALUES ('claude-3-5-sonnet', 'cloud', 'claude', 'Claude 3.5 Sonnet',
    'https://api.anthropic.com', 0.003, 0.015, 0);

-- 默认成本预警配置
INSERT INTO am_cost_alert_config (daily_limit, monthly_limit, alert_threshold)
VALUES (100.00, 3000.00, 0.8);
```

---

## 7. HTTP API 设计

### 7.1 API 分组

**模型配置管理** (`/api/ai-model/configs`)
- GET /configs - 获取模型列表
- POST /configs - 添加模型
- PUT /configs/:id - 更新模型
- DELETE /configs/:id - 删除模型
- POST /configs/:id/test - 测试连接
- PUT /configs/:id/status - 启用/禁用

**健康检查与监控** (`/api/ai-model/health`)
- GET /health - 获取所有模型健康状态
- GET /health/:model_name/history - 获取健康历史
- POST /health/:model_name/check - 手动触发检查

**调用统计** (`/api/ai-model/statistics`)
- GET /statistics - 获取使用统计
- GET /statistics/cost - 获取成本统计
- GET /call-logs - 获取调用日志
- GET /call-logs/:id - 获取日志详情

**提示词模板管理** (`/api/ai-model/templates`)
- GET /templates - 获取模板列表
- POST /templates - 创建模板
- PUT /templates/:id - 更新模板
- DELETE /templates/:id - 删除模板
- POST /templates/:id/render - 渲染模板

**成本预警** (`/api/ai-model/alerts`)
- GET /alerts/config - 获取预警配置
- PUT /alerts/config - 更新预警配置
- GET /alerts/records - 获取预警记录
- PUT /alerts/records/:id/resolve - 标记已处理

**API Key 管理** (`/api/ai-model/api-keys`)
- GET /api-keys - 获取 Key 列表（脱敏）
- POST /api-keys - 添加 Key
- PUT /api-keys/:id - 更新 Key
- DELETE /api-keys/:id - 删除 Key
- POST /api-keys/:id/test - 测试 Key

### 7.2 统一响应格式

所有 API 使用 `responsex.Response` 返回：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 实际数据
  }
}
```

---

## 8. 前端管理界面

### 8.1 页面结构

1. **仪表板**：概览卡片、健康状态、趋势图表
2. **模型配置管理**：模型列表、添加/编辑、测试连接
3. **健康监控**：实时状态、健康历史、手动检查
4. **调用统计**：统计概览、趋势图表、成本分析、调用日志
5. **提示词模板管理**：模板列表、创建/编辑、测试
6. **成本预警**：预警配置、预警记录、成本预测
7. **API Key 管理**：Key 列表、添加/编辑、测试

### 8.2 技术栈

- **框架**：Vue 3 + TypeScript + Composition API
- **UI 库**：Element Plus
- **图表**：ECharts
- **HTTP 客户端**：Axios
- **状态管理**：Pinia

### 8.3 关键功能

- **实时更新**：健康状态每 30 秒自动刷新
- **数据可视化**：调用趋势、成本趋势、成本占比
- **交互友好**：表格支持筛选、排序、分页
- **状态指示**：使用颜色和图标直观显示健康状态

---

## 9. 实施计划

### Phase 1: 核心功能（2-3周）

**目标**：实现基础的模型调用和管理功能，支持本地 Qwen 和 Claude。

**交付物**：
- ✅ 可运行的 ai-model-rpc 服务（8084）
- ✅ 可运行的 ai-model-api 服务（8891）
- ✅ 支持 Qwen 和 Claude 两个模型
- ✅ 支持 ModerateText、CallModel 等核心接口
- ✅ 基础的模型配置管理 API
- ✅ 调用日志记录和成本统计

### Phase 2: 管理功能（1-2周）

**目标**：完善管理功能，提供完整的前端管理界面。

**交付物**：
- ✅ 完整的 HTTP 管理 API
- ✅ 完整的前端管理界面
- ✅ 健康检查调度器（后台任务）
- ✅ 成本预警机制
- ✅ 提示词模板系统
- ✅ API Key 加密存储和管理

### Phase 3: 扩展和优化（1-2周）

**目标**：添加更多模型支持，优化性能，完善文档。

**交付物**：
- ✅ 支持 4 种模型（Qwen、Claude、OpenAI、Ollama）
- ✅ 批量调用优化
- ✅ API Key 自动轮换
- ✅ 性能优化和压测报告
- ✅ Masterdata 服务集成完成
- ✅ 完整的文档
- ✅ 生产环境部署

---

## 10. 风险和缓解措施

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| **现有 Python 引擎不稳定** | 高 | 中 | 1. 增加重试机制<br>2. 实现降级到 Claude<br>3. 监控健康状态 |
| **外部 API 限流** | 中 | 高 | 1. 实现请求队列<br>2. 多 API Key 轮换<br>3. 本地模型作为备份 |
| **成本超预算** | 高 | 中 | 1. 实时成本监控<br>2. 预警机制<br>3. 每日配额限制 |
| **数据库性能瓶颈** | 中 | 低 | 1. 调用日志异步写入<br>2. 定期归档历史数据<br>3. 读写分离 |
| **API Key 泄露** | 高 | 低 | 1. AES-256 加密存储<br>2. 数据库访问控制<br>3. 定期轮换 Key |
| **提示词注入攻击** | 中 | 中 | 1. 输入转义和清洗<br>2. 系统提示词防护<br>3. 内容审计 |

---

## 11. 成功标准

**Phase 1**：
- ✅ RPC 服务可正常调用，响应时间 < 2s
- ✅ 本地 Qwen 和 Claude 都能正常工作
- ✅ 调用日志正确记录
- ✅ 成本计算准确

**Phase 2**：
- ✅ 前端界面完整可用
- ✅ 健康检查每 5 分钟自动执行
- ✅ 成本预警能正确触发
- ✅ 所有管理功能正常

**Phase 3**：
- ✅ 支持 4 种模型
- ✅ 批量调用性能提升 > 80%
- ✅ 单元测试覆盖率 > 70%
- ✅ 文档完整
- ✅ 生产环境稳定运行

---

## 12. 附录

### 12.1 端口分配

- ai-model-rpc: 8084
- ai-model-api: 8891
- Python Engine: 8001（现有）

### 12.2 依赖服务

- MySQL 8.0（ai_model_db）
- Redis 7.0（缓存）
- etcd 3.5（服务注册）
- Python Engine（现有，保持独立）

### 12.3 相关文档

- API 文档：`docs/api/ai-model-service.md`（待创建）
- 部署文档：`services/ai-model/DEPLOYMENT.md`（待创建）
- 使用手册：`services/ai-model/USER_GUIDE.md`（待创建）
- 架构文档：`services/ai-model/ARCHITECTURE.md`（待创建）
