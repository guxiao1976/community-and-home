## Why

社区平台需要集成大语言模型（LLM）能力来实现智能化功能，如敏感词自动分类、内容智能审核、用户问题自动回复等。当前各服务直接调用外部 AI API 存在以下问题：

1. **配置分散**：每个服务独立管理 API Key、模型配置，难以统一管理
2. **成本失控**：缺少统一的调用监控和成本管理，无法预警和控制费用
3. **重复开发**：每个服务都需要实现相同的 API 调用、错误处理、重试逻辑
4. **缺少监控**：无法统一查看模型健康状态、调用成功率、响应时间等指标
5. **提示词管理混乱**：提示词硬编码在各服务中，难以复用和优化

本项目创建独立的 AI-Model 微服务，提供统一的 AI 模型调用能力，支持多模型（Claude、GPT、Ollama 等），实现配置管理、健康监控、成本控制、提示词模板管理等功能。

## What Changes

- **新增 ai-model 微服务**：在 `services/ai-model/` 创建完整的微服务框架
- **新增 ai_model_db 数据库**：独立数据库存储模型配置、调用日志、统计数据
- **RPC 接口**：提供 `AIModelService` RPC 服务供其他微服务调用
- **HTTP 管理接口**：提供模型配置、健康检查、统计查询的 Web 管理界面
- **模型适配器**：实现 Claude、OpenAI、Ollama 等多种模型的统一调用接口
- **成本管理**：调用日志记录、成本统计、预警机制
- **提示词模板**：可复用的提示词模板管理

## Capabilities

### New Capabilities

- `model-config`: 模型配置管理，支持多模型（Claude、GPT、Ollama）、多 API Key、优先级和负载均衡
- `model-adapter`: 统一的模型调用适配器，封装不同厂商的 API 差异，提供统一接口
- `health-check`: 模型健康检查，定期探测模型可用性、响应时间、错误率
- `call-log`: 调用日志记录，记录每次调用的请求、响应、耗时、成本
- `usage-statistics`: 使用统计，按天/周/月汇总调用次数、Token 消耗、成本
- `cost-alert`: 成本预警，当日/月成本超过阈值时触发告警
- `prompt-template`: 提示词模板管理，支持变量替换、版本管理、A/B 测试
- `api-key-rotation`: API Key 轮换和管理，支持多 Key 负载均衡、自动失效检测
- `batch-call`: 批量调用优化，支持批量请求合并、并发控制、断点续传
- `rpc-service`: gRPC 服务接口，供其他微服务调用（CallModel、CallModelBatch、GetAvailableModels）
- `http-api`: HTTP 管理接口，提供 Web 管理界面（配置、监控、统计）

### Modified Capabilities

- `masterdata-sensitive-word`: Masterdata 服务的敏感词分类功能将通过 RPC 调用 AI-Model 服务

## Impact

- **数据库**：新建 ai_model_db 数据库，包含 8 张表（模型配置、健康检查、调用日志、提示词模板、使用统计、成本预警配置、预警记录、API Key 管理）
- **服务**：新增 ai-model-rpc（端口 8084）和 ai-model-api（端口 8891）两个服务
- **依赖**：新增 Go 依赖 anthropic-sdk-go、openai-go、ollama-go 等 AI SDK
- **配置**：新增 ai-model-api.yaml 和 ai-model-rpc.yaml 配置文件
- **基础设施**：需要配置外部 AI API 的网络访问权限
- **其他服务**：Masterdata 等服务需要添加 AI-Model RPC 客户端依赖
