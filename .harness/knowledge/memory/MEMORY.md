# 全局经验索引

> Agent 启动时读取本文件，根据当前任务上下文精读相关记忆文件。
> 格式：`- [标题](文件.md) — 适用范围, 严重程度, 类型, 触发关键词`

- [测试纪律](testing-discipline.md) — 所有代码改动, 硬性约束, 流程, 测试/PR/提交/改动
>
> 类型说明：`pitfall`=踩过的坑 | `guideline`=编码/架构规范 | `process`=流程约束 | `decision`=技术决策 | `model`=数据模型

## 必须遵守 (must-follow)

- [Proto int64 字段必须加 jstype=JS_STRING](proto-jstype.md) — api-proto, must-follow, `guideline`, `proto int64 jstype JS_STRING Snowflake`
- [服务间通信仅通过 gRPC](grpc-only-comms.md) — all, must-follow, `guideline`, `gRPC 服务间调用 直连数据库`
- [API 调用前验证路由存在 + 禁止静默吞错](verify-api-before-calling.md) — all, must-follow, `pitfall`, `API 路由 404 catch 空 静默`
- [提交前必须通过机械化检查](pre-commit-checks.md) — all, must-follow, `process`, `提交 commit 门禁 gate harness-checks QA 检查`
- [手机号加密存储读取必须解密](phone-encryption.md) — all, must-follow, `pitfall`, `phone 手机号 AES 加密 解密 乱码`
- [Migration 文件提交后必须执行](migration-must-execute.md) — all, must-follow, `pitfall`, `migration DDL ALTER TABLE column schema 1054 unknown column`
- [API 响应必须单层包装](api-response-single-wrap.md) — all, must-follow, `pitfall`, `API 响应 response.Success 双层嵌套 BaseResponse goctl Response 格式`
- [大模型 API 连接测试实现规范](llm-connection-test.md) — ai-model-service, must-follow, `pitfall`, `连接测试 test connection 大模型 API endpoint LLM DeepSeek POST GET 404`
- [Vue 模板中不能出现嵌套的 {{ 字面量](vue-template-nested-interpolation.md) — web, must-follow, `pitfall`, `Vue 模板 {{ 插值 花括号 template parse error`
- [goctl 生成的 Logic 空壳必须实现后才能交付](goctl-logic-stubs.md) — all, must-follow, `pitfall`, `TODO stub 空壳 goctl silent success 假成功 harness check`
- [gRPC 调用链需要三层超时对齐](grpc-timeout-layers.md) — all, must-follow, `pitfall`, `gRPC timeout DeadlineExceeded 超时 LLM 大模型调用`
- [端点自动补全逻辑必须在所有调用路径共享](endpoint-auto-complete.md) — ai-model-service, must-follow, `guideline`, `endpoint auto-complete base URL v1/chat v1/messages`
- [软删除 + Redis 索引缓存 → 唯一索引冲突](redis-cache-soft-delete.md) — all, should-follow, `pitfall`, `Redis cache 缓存 软删除 unique index stale`
- [DeepSeek 响应可能包含 thinking 块](deepseek-thinking-response.md) — ai-model-service, must-follow, `pitfall`, `DeepSeek thinking Claude adapter response content block`
- [编辑回显数据完整性——表单字段 8 层全链路检查](edit-form-data-integrity.md) — all, must-follow, `pitfall`, `编辑 回显 表单 下拉框 combobox model_type 数据丢失 proto types.go TS`
- [Submodule 缺少 .gitmodules 导致 worktree 不可用](moderation-submodule-broken.md) — all, should-follow, `pitfall`, `submodule worktree .gitmodules moderation-service master-data-service`
- [服务 API 层启动需要 JWT_ACCESS_SECRET 环境变量](service-api-jwt-env.md) — all, should-follow, `pitfall`, `JWT_ACCESS_SECRET env 启动 API layer panic secret`
- [CheckText RPC 已接入管线配置](moderation-checktext-pipeline.md) — moderation-service, must-follow, `decision`, `CheckText 管线 pipeline 审核接口 RPC 生产管线 is_production`
- [qwen2.5:3b 不适合内容审核初筛](qwen-3b-unsuitable-for-moderation.md) — all, must-follow, `pitfall`, `qwen 小模型 本地模型 审核模型选型`
- [模板变量名必须用 {{content}}](ai-model-template-variable-content.md) — ai-model-service, must-follow, `pitfall`, `模板 变量 {{input}} {{content}} prompt ModerateText`
- [gRPC 4MB 上限 vs 5万+ 敏感词](grpc-max-msg-size-sensitive-words.md) — moderation-service, must-follow, `pitfall`, `gRPC ResourceExhausted 敏感词 MaxCallRecvMsgSize`
- [每次代码变更必须执行验证清单](change-verification-checklist.md) — all, must-follow, `process`, `验证 交付 编译 部署 前后端 遗漏 忘了 改了 修改 变更 build deploy`
- [Harness Pipeline 需要参数校验防止 undefined 字符串化](harness-pipeline-undefined-guard.md) — all, should-follow, `pitfall`, `harness pipeline undefined 参数 args serviceName serviceDir task 校验`
- [管线数据可信度：模拟数据 ≠ 真实数据严重性](pipeline-data-trust-simulated-vs-real.md) — all, must-follow, `pitfall`, `模拟数据 测试数据 管线 数据可信度 63次 时间戳 数据质量`
- [Monorepo 端口冲突检测与启动顺序](monorepo-port-management.md) — all, must-follow, `pitfall`, `端口 冲突 port monorepo 8087 8088 start.sh stop.sh smoketest`

## 应该遵守 (should-follow)

- [前端管线接入模式：条件分支而非独立文件](frontend-pipeline-integration-pattern.md) — all, should-follow, `guideline`, `前端 管线 pipeline Vue TypeScript isFrontend vitest npm`

- [前端可视化开发流程](frontend-visual-development-workflow.md) — web/mobile, should-follow, `process`, `UI 页面 设计 样式 视觉`
- [Harness 架构决策记录](harness-architecture-decisions.md) — all, should-follow, `decision`, `harness .harness 驾驭工程 目录结构 架构决策 四支柱`

## 参考信息 (info)

- [系统名称通过环境变量可配置](system-name-configurable.md) — web, info, `reference`, `系统名称 标题 VITE_APP_TITLE 配置`

