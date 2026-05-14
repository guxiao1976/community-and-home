## 1. 数据库设计与创建

- [x] 1.1 编写 DDL 创建 ai_model_db 数据库
- [x] 1.2 编写 DDL 创建 am_model_config 表（模型配置）
- [x] 1.3 编写 DDL 创建 am_health_check 表（健康检查记录）
- [x] 1.4 编写 DDL 创建 am_call_log 表（调用日志）
- [x] 1.5 编写 DDL 创建 am_prompt_template 表（提示词模板）
- [x] 1.6 编写 DDL 创建 am_usage_statistics 表（使用统计）
- [x] 1.7 编写 DDL 创建 am_cost_alert_config 表（成本预警配置）
- [x] 1.8 编写 DDL 创建 am_alert_record 表（预警记录）
- [x] 1.9 编写 DDL 创建 am_api_key 表（API密钥管理）
- [x] 1.10 执行所有 DDL 语句创建数据库和表

## 2. 服务框架搭建

- [x] 2.1 创建 services/ai-model 目录结构
- [x] 2.2 使用 goctl 生成 ai-model-rpc 服务框架（端口 8084）
- [ ] 2.3 使用 goctl 生成 ai-model-api 服务框架（端口 8891）
- [x] 2.4 编写 ai-model.proto 定义 RPC 接口（CallModel、CallModelBatch、GetAvailableModels、HealthCheck）
- [x] 2.5 生成 protobuf 代码
- [ ] 2.6 配置 etcd 服务注册
- [x] 2.7 编写基础配置文件（ai-model-api.yaml、ai-model-rpc.yaml）

## 3. Model 层实现

- [x] 3.1 使用 goctl 生成 am_model_config 表的 model 代码
- [x] 3.2 使用 goctl 生成 am_health_check 表的 model 代码
- [x] 3.3 使用 goctl 生成 am_call_log 表的 model 代码
- [x] 3.4 使用 goctl 生成 am_prompt_template 表的 model 代码
- [x] 3.5 使用 goctl 生成 am_usage_statistics 表的 model 代码
- [x] 3.6 使用 goctl 生成 am_cost_alert_config 表的 model 代码
- [x] 3.7 使用 goctl 生成 am_alert_record 表的 model 代码
- [x] 3.8 使用 goctl 生成 am_api_key 表的 model 代码
- [ ] 3.9 编写自定义 model 方法（批量查询、统计等）

## 4. 模型适配器实现

- [x] 4.1 定义 ModelAdapter 接口（internal/adapter/interface.go）
- [x] 4.2 实现 ClaudeAdapter（internal/adapter/claude.go）
- [x] 4.3 添加 anthropic-sdk-go 依赖（使用原生 HTTP 客户端）
- [x] 4.4 实现 Claude API 调用逻辑（认证、请求、响应解析）
- [x] 4.5 实现 Claude 错误处理和重试机制
- [x] 4.6 实现 Claude Token 计数和成本估算
- [x] 4.7 实现 OpenAIAdapter（internal/adapter/openai.go）
- [x] 4.8 添加 openai-go 依赖（使用原生 HTTP 客户端）
- [x] 4.9 实现 OpenAI API 调用逻辑
- [x] 4.10 实现 OllamaAdapter（internal/adapter/ollama.go）
- [x] 4.11 添加 ollama-go 依赖（使用原生 HTTP 客户端）
- [x] 4.12 实现 Ollama 本地调用逻辑
- [ ] 4.13 编写适配器单元测试

## 5. 模型管理器实现

- [x] 5.1 实现 ModelManager（internal/manager/manager.go）
- [x] 5.2 实现模型配置加载和缓存
- [x] 5.3 实现模型选择逻辑（优先级、负载均衡）
- [x] 5.4 实现 API Key 轮换逻辑
- [x] 5.5 实现 API Key 加密存储（AES-256）
- [x] 5.6 实现模型健康检查调度器
- [x] 5.7 实现健康状态评估逻辑（Healthy/Degraded/Unhealthy）
- [x] 5.8 实现成本计算和统计（CostManager）
- [x] 5.9 实现成本预警检查
- [x] 5.10 实现 TemplateManager（提示词模板管理）
- [ ] 5.11 编写 ModelManager 单元测试

## 6. RPC 服务实现

- [x] 6.1 实现 CallModel RPC 方法（单次调用）
- [ ] 6.2 实现 CallModelBatch RPC 方法（批量调用）
- [x] 6.3 实现 GetAvailableModels RPC 方法
- [x] 6.4 实现 HealthCheck RPC 方法
- [x] 6.5 实现调用日志记录逻辑
- [x] 6.6 实现错误处理和响应封装
- [ ] 6.7 修复编译错误（model 字段名不匹配）
- [ ] 6.8 编写 RPC 服务集成测试

## 7. HTTP API 实现 - 模型配置管理

- [ ] 7.1 编写 ai-model.api 定义模型配置接口
- [ ] 7.2 实现 GET /api/ai-model/configs（获取模型列表）
- [ ] 7.3 实现 POST /api/ai-model/configs（添加模型）
- [ ] 7.4 实现 PUT /api/ai-model/configs/:id（更新模型）
- [ ] 7.5 实现 DELETE /api/ai-model/configs/:id（删除模型）
- [ ] 7.6 实现 POST /api/ai-model/configs/:id/test（测试模型连接）
- [ ] 7.7 实现 PUT /api/ai-model/configs/:id/enable（启用/禁用模型）

## 8. HTTP API 实现 - 健康检查与监控

- [ ] 8.1 实现 GET /api/ai-model/health（获取所有模型健康状态）
- [ ] 8.2 实现 GET /api/ai-model/health/:id（获取单个模型健康状态）
- [ ] 8.3 实现 POST /api/ai-model/health/:id/check（手动触发健康检查）

## 9. HTTP API 实现 - 调用统计

- [ ] 9.1 实现 GET /api/ai-model/statistics（获取使用统计）
- [ ] 9.2 实现 GET /api/ai-model/statistics/daily（按天统计）
- [ ] 9.3 实现 GET /api/ai-model/statistics/cost（成本统计）
- [ ] 9.4 实现 GET /api/ai-model/call-logs（获取调用日志）

## 10. HTTP API 实现 - 提示词模板管理

- [ ] 10.1 实现 GET /api/ai-model/templates（获取模板列表）
- [ ] 10.2 实现 POST /api/ai-model/templates（创建模板）
- [ ] 10.3 实现 PUT /api/ai-model/templates/:id（更新模板）
- [ ] 10.4 实现 DELETE /api/ai-model/templates/:id（删除模板）
- [ ] 10.5 实现 POST /api/ai-model/templates/:id/render（渲染模板）
- [ ] 10.6 实现模板变量替换逻辑

## 11. HTTP API 实现 - 成本预警

- [ ] 11.1 实现 GET /api/ai-model/alerts/config（获取预警配置）
- [ ] 11.2 实现 PUT /api/ai-model/alerts/config（更新预警配置）
- [ ] 11.3 实现 GET /api/ai-model/alerts/records（获取预警记录）
- [ ] 11.4 实现预警触发逻辑
- [ ] 11.5 实现预警通知（日志记录，预留钉钉/邮件接口）

## 12. API Key 管理

- [ ] 12.1 实现 GET /api/ai-model/api-keys（获取 Key 列表，脱敏显示）
- [ ] 12.2 实现 POST /api/ai-model/api-keys（添加 Key）
- [ ] 12.3 实现 PUT /api/ai-model/api-keys/:id（更新 Key）
- [ ] 12.4 实现 DELETE /api/ai-model/api-keys/:id（删除 Key）
- [ ] 12.5 实现 Key 失效检测逻辑
- [ ] 12.6 实现 Key 轮换调度器

## 13. 批量调用优化

- [ ] 13.1 实现批量请求合并逻辑
- [ ] 13.2 实现批量结果解析和验证
- [ ] 13.3 实现部分失败处理逻辑
- [ ] 13.4 实现并发控制（限制同时调用数）
- [ ] 13.5 实现断点续传机制
- [ ] 13.6 编写批量调用测试

## 14. 前端管理界面

- [ ] 14.1 创建 AI 模型管理页面路由
- [ ] 14.2 实现模型配置列表页面
- [ ] 14.3 实现模型添加/编辑表单
- [ ] 14.4 实现模型测试功能
- [ ] 14.5 实现健康状态监控页面
- [ ] 14.6 实现调用统计图表（ECharts）
- [ ] 14.7 实现成本统计和预警配置页面
- [ ] 14.8 实现提示词模板管理页面
- [ ] 14.9 实现 API Key 管理页面

## 15. Masterdata 服务集成

- [ ] 15.1 在 Masterdata 服务中添加 AI-Model RPC 客户端依赖
- [ ] 15.2 在 masterdata-api.yaml 中配置 AI-Model RPC 地址
- [ ] 15.3 实现敏感词分类接口（POST /api/masterdata/sensitive-words/classify）
- [ ] 15.4 实现批量分类逻辑（调用 AI-Model RPC）
- [ ] 15.5 实现分类结果保存（更新 category_by_model 字段）
- [ ] 15.6 在前端添加"敏感词分类"Tab 页面
- [ ] 15.7 实现提示词输入和模型选择界面
- [ ] 15.8 实现分类进度显示
- [ ] 15.9 实现分类结果预览和人工修正
- [ ] 15.10 测试完整的敏感词分类流程

## 16. 测试与文档

- [ ] 16.1 编写单元测试（覆盖率 > 70%）
- [ ] 16.2 编写集成测试
- [ ] 16.3 编写 API 文档
- [ ] 16.4 编写部署文档
- [ ] 16.5 编写使用手册
- [ ] 16.6 性能测试（并发调用、批量处理）
- [ ] 16.7 安全测试（API Key 加密、提示词注入）

## 17. 部署与上线

- [ ] 17.1 更新 docker-compose.yml 添加 ai-model 服务
- [ ] 17.2 更新 start_project.sh 脚本
- [ ] 17.3 配置生产环境数据库
- [ ] 17.4 配置外部 AI API 网络访问
- [ ] 17.5 配置监控和告警
- [ ] 17.6 执行数据库迁移
- [ ] 17.7 部署服务到测试环境
- [ ] 17.8 测试环境验证
- [ ] 17.9 部署到生产环境
- [ ] 17.10 生产环境验证
