# 全局经验索引

> Agent 启动时读取本文件，根据当前任务上下文精读相关记忆文件。
> 格式：`- [标题](文件.md) — 适用范围, 严重程度, 类型, 触发关键词`

- [测试纪律](global/testing-discipline.md) — 所有代码改动, 硬性约束, 流程, 测试/PR/提交/改动
- [is_system 不再授予全权限](permission-service/is-system-no-permission-shortcut.md) — permission-service, must-follow, `decision`, `is_system CheckPermission RBAC 系统角色 短路`
- [种子数据 API path 必须匹配实际路由](permission-service/permission-seed-api-path-must-match-routes.md) — permission-service, must-follow, `pitfall`, `seed path 前缀 CheckPermission 匹配 路由`
- [FindByRoleId 查询不存在的 assign_time 列 → MySQL 1054](permission-service/need-human-findbyroleid-assign_time.md) — permission-service, must-follow, `pitfall`, `FindByRoleId assign_time rel_user_role 1054 Unknown column select * need_human 缓存失效`
>
> 类型说明：`pitfall`=踩过的坑 | `guideline`=编码/架构规范 | `process`=流程约束 | `decision`=技术决策 | `model`=数据模型

## 必须遵守 (must-follow)

- [Proto int64 字段必须加 jstype=JS_STRING](api-proto/proto-jstype.md) — api-proto, must-follow, `guideline`, `proto int64 jstype JS_STRING Snowflake`
- [服务间通信仅通过 gRPC](global/grpc-only-comms.md) — all, must-follow, `guideline`, `gRPC 服务间调用 直连数据库`
- [API 调用前验证路由存在 + 禁止静默吞错](global/verify-api-before-calling.md) — all, must-follow, `pitfall`, `API 路由 404 catch 空 静默`
- [提交前必须通过机械化检查](global/pre-commit-checks.md) — all, must-follow, `process`, `提交 commit 门禁 gate harness-checks QA 检查`
- [手机号加密存储读取必须解密](global/phone-encryption.md) — all, must-follow, `pitfall`, `phone 手机号 AES 加密 解密 乱码`
- [Migration 文件提交后必须执行](global/migration-must-execute.md) — all, must-follow, `pitfall`, `migration DDL ALTER TABLE column schema 1054 unknown column`
- [API 响应必须单层包装](global/api-response-single-wrap.md) — all, must-follow, `pitfall`, `API 响应 response.Success 双层嵌套 BaseResponse goctl Response 格式`
- [大模型 API 连接测试实现规范](ai-model-service/llm-connection-test.md) — ai-model-service, must-follow, `pitfall`, `连接测试 test connection 大模型 API endpoint LLM DeepSeek POST GET 404`
- [Vue 模板中不能出现嵌套的 {{ 字面量](web/vue-template-nested-interpolation.md) — web, must-follow, `pitfall`, `Vue 模板 {{ 插值 花括号 template parse error`
- [goctl 生成的 Logic 空壳必须实现后才能交付](global/goctl-logic-stubs.md) — all, must-follow, `pitfall`, `TODO stub 空壳 goctl silent success 假成功 harness check`
- [gRPC 调用链需要三层超时对齐](global/grpc-timeout-layers.md) — all, must-follow, `pitfall`, `gRPC timeout DeadlineExceeded 超时 LLM 大模型调用`
- [端点自动补全逻辑必须在所有调用路径共享](ai-model-service/endpoint-auto-complete.md) — ai-model-service, must-follow, `guideline`, `endpoint auto-complete base URL v1/chat v1/messages`
- [软删除 + Redis 索引缓存 → 唯一索引冲突](global/redis-cache-soft-delete.md) — all, should-follow, `pitfall`, `Redis cache 缓存 软删除 unique index stale`
- [DeepSeek 响应可能包含 thinking 块](ai-model-service/deepseek-thinking-response.md) — ai-model-service, must-follow, `pitfall`, `DeepSeek thinking Claude adapter response content block`
- [编辑回显数据完整性——表单字段 8 层全链路检查](global/edit-form-data-integrity.md) — all, must-follow, `pitfall`, `编辑 回显 表单 下拉框 combobox model_type 数据丢失 proto types.go TS`
- [Submodule 缺少 .gitmodules 导致 worktree 不可用](global/moderation-submodule-broken.md) — all, should-follow, `pitfall`, `submodule worktree .gitmodules moderation-service master-data-service`
- [服务 API 层启动需要 JWT_ACCESS_SECRET 环境变量](global/service-api-jwt-env.md) — all, should-follow, `pitfall`, `JWT_ACCESS_SECRET env 启动 API layer panic secret`
- [CheckText RPC 已接入管线配置](moderation-service/moderation-checktext-pipeline.md) — moderation-service, must-follow, `decision`, `CheckText 管线 pipeline 审核接口 RPC 生产管线 is_production`
- [qwen2.5:3b 不适合内容审核初筛](global/qwen-3b-unsuitable-for-moderation.md) — all, must-follow, `pitfall`, `qwen 小模型 本地模型 审核模型选型`
- [模板变量名必须用 {{content}}](ai-model-service/ai-model-template-variable-content.md) — ai-model-service, must-follow, `pitfall`, `模板 变量 {{input}} {{content}} prompt ModerateText`
- [gRPC 4MB 上限 vs 5万+ 敏感词](moderation-service/grpc-max-msg-size-sensitive-words.md) — moderation-service, must-follow, `pitfall`, `gRPC ResourceExhausted 敏感词 MaxCallRecvMsgSize`
- [每次代码变更必须执行验证清单](global/change-verification-checklist.md) — all, must-follow, `process`, `验证 交付 编译 部署 前后端 遗漏 忘了 改了 修改 变更 build deploy`
- [Harness Pipeline 需要参数校验防止 undefined 字符串化](global/harness-pipeline-undefined-guard.md) — all, should-follow, `pitfall`, `harness pipeline undefined 参数 args serviceName serviceDir task 校验`
- [管线数据可信度：模拟数据 ≠ 真实数据严重性](global/pipeline-data-trust-simulated-vs-real.md) — all, must-follow, `pitfall`, `模拟数据 测试数据 管线 数据可信度 63次 时间戳 数据质量`
- [Monorepo 端口冲突检测与启动顺序](global/monorepo-port-management.md) — all, must-follow, `pitfall`, `端口 冲突 port monorepo 8087 8088 start.sh stop.sh smoketest`
- [TypeScript erasableSyntaxOnly 与 enum 冲突](web/typescript-erasable-syntax-enum-conflict.md) — web, must-follow, `pitfall`, `TypeScript erasableSyntaxOnly enum TS1294 npm run build vue-tsc 编译失败`
- [QA 的 TDD RED 证据必须包含实际 FAIL 输出摘录](global/tdd-red-evidence-requires-fail-excerpt.md) — all, must-follow, `process`, `TDD RED GREEN 测试未提交 证据摘录 undefined 编译失败 QA FAIL`

## 应该遵守 (should-follow)

- [前端管线接入模式：条件分支而非独立文件](global/frontend-pipeline-integration-pattern.md) — all, should-follow, `guideline`, `前端 管线 pipeline Vue TypeScript isFrontend vitest npm`
- [错误码必须语义唯一，禁止同码异义](global/error-code-collision-and-namespace-alignment.md) — all, should-follow, `guideline`, `错误码 error code 60006 060007 冲突 复用 语义 NewBaseRespWithError 同码异义 namespace`
- [INSERT IGNORE 静默吞掉非唯一键错误导致假成功](global/insert-ignore-swallows-errors.md) — all, should-follow, `pitfall`, `INSERT IGNORE 幂等 唯一键 静默 假成功 RowsAffected ON DUPLICATE KEY AssignRole`
- [缓存 not-found sentinel 前必须区分 ErrNoRows 与瞬时 DB 错误](global/notfound-cache-sentinel-vs-transient-error.md) — all, should-follow, `pitfall`, `缓存 not-found sentinel ErrNoRows 瞬时错误 DB抖动 30分钟 min_verf_level 冷缓存`
- [对既有数据加唯一索引必须先查重](global/unique-index-migration-dup-precheck.md) — all, should-follow, `pitfall`, `唯一索引 迁移 ALTER duplicate 重复数据 阻塞部署 ADD UNIQUE uk_user_role_scope`

- [前端可视化开发流程](web/frontend-visual-development-workflow.md) — web/mobile, should-follow, `process`, `UI 页面 设计 样式 视觉`
- [Harness 架构决策记录](global/harness-architecture-decisions.md) — all, should-follow, `decision`, `harness .harness 驾驭工程 目录结构 架构决策 四支柱`

## 参考信息 (info)

- [系统名称通过环境变量可配置](web/system-name-configurable.md) — web, info, `reference`, `系统名称 标题 VITE_APP_TITLE 配置`

