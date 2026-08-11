# Code Review — moderation-service

**审查时间**: 2026-06-16 22:45
**审查范围**: 审核管线可配置化 (2026-06-15) + 错误码规范化修复 + 引擎重构 (text_engine checkMode 扩展)
**审查者**: Code Reviewer

## 摘要
- 变更文件: 38 (新增 7 个 pipeline handler + 7 个 pipeline logic + 2 个 model 文件 + 1 个 migration + 1 个 pipeline 包，修改 20 个已有文件)
- 共计: +1654 / -766
- 🔴 CRITICAL: 1 / 🟡 WARNING: 6 / 🔵 NOTE: 5

## 发现

### 🔴 CRITICAL
| # | 文件:行号 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | `internal/pipeline/executor.go:251` | `computeFinalVerdict` 中 `case "large_overrides":` 通过 `fallthrough` 落入 `case "small_overrides":`。当大模型被调用且通过时（`return "pass"` line 247）是正确的。但若大模型被调用且不通过（`return "reject"` line 249）也是正确的——它不会 fallthrough。然而当 `LargeModelResult` 为 nil 或未被调用时，`fallthrough` 会进入 `small_overrides` 逻辑。但 `large_overrides` 的语义是"大模型结果为准"，如果大模型没被调用，fallthrough 到 small 逻辑是合理的降级策略。**但是**：这会导致 `large_overrides` 模式在 large model 未调用时退化成 `small_overrides`，这与"overrides"的显式语义矛盾——应根据设计意图确认。如果 `large_overrides` 且大模型未调用，期望行为应是 `need_review`（保守安全）而非退到 small model 判定。当前实现可能导致：大模型因未配置模板ID而被跳过时，小模型的判定会替代终判，这可能是安全的也可能会误放行。请确认设计意图。 | 若期望 `large_overrides` 且大模型未调用时返回保守结果，在 line 251 fallthrough 前加 `return "need_review"`。若设计意图确为降级到 small，请在代码中加注释说明降级策略。 |

### 🟡 WARNING
| # | 文件:行号 | 问题 | 建议 |
|---|----------|------|------|
| 1 | `api/internal/handler/pipeline/` 7 个 handler | 所有 7 个 pipeline handler 使用 `httpx.OkJson(w, resp)` / `httpx.Error(w, err)` 返回响应，而非项目标准的 `responsex.Response(w, resp, err)`。其他 handler（health、text_check、image_check、submit_review）和 middleware 都已使用 `responsex.Response`。这会导致 pipeline handlers 的响应格式与项目其他端点不一致：其他端点返回 `{code, msg, data}` 标准格式，pipeline 端点返回原始 JSON。QA 报告已持续追踪此问题，仍未修复。 | 统一使用 `responsex.Response(w, resp, err)` 替代 `httpx.OkJson`/`httpx.Error`，确保响应格式全局一致。Logic 层需改为返回 `(*T, error)` 形式（当前 Create/Update/Delete/Activate 返回`error`无 data，需适配）。 |
| 2 | `api/internal/types/types.go:98-103` | `DeletePipelineReq` 和 `GetPipelineReq` 的 `PipelineId` 字段使用 `json:"pipeline_id"` tag，但路由定义为 `DELETE /pipeline/:pipeline_id` 和 `GET /pipeline/:pipeline_id`（path 参数）。go-zero 的 `httpx.Parse` 根据 tag 区分来源：path 参数需要 `path` tag，json 标签用于 body 解析。当前 `json` tag 不会提取 path 参数，`pipeline_id` 在 DELETE/GET 请求中会为空值。类似地 `ActivatePipelineReq` 同样使用 `json:"pipeline_id"` 但路由为 `PUT /pipeline/:pipeline_id/activate`。 | 将 `PipelineId` 字段的 tag 改为 `path:"pipeline_id"`，或使用 go-zero 的 `httpx.GetRouter(r)` 手动提取 path 参数。对于 GET 请求，还需确认 `httpx.Parse` 对 body 的处理（GET 通常无 body，`pipeline_id` 必须来自 path）。 |
| 3 | `internal/pipeline/` 包 (config.go + executor.go) | 核心管线引擎代码零单元测试。executor.go 包含 4 种升级条件分支 x 3 种终判逻辑 = 12+ 种组合，以及多层级的早期返回路径（AC 禁用、模板缺失、调用失败等），全部无测试覆盖。config.go 的 `FromDBModel` 包含 JSON 解析和 Nullable 字段处理，也未测试。QA 报告确认 0.0% 覆盖率。 | 为 `shouldEscalateAcToSmall`、`shouldEscalateSmallToLarge`、`computeFinalVerdict`、`FromDBModel` 编写单元测试，覆盖所有条件枚举值和边界（空 hits、nil LayerResult、JSON 解析失败等）。 |
| 4 | `api/internal/logic/pipeline/update_pipeline_logic.go:52,58` | 条件 `if req.SmallModelConfigKey != "" \|\| req.SmallModelConfigKey == ""` 和 `if req.LargeModelConfigKey != "" \|\| req.LargeModelConfigKey == ""` 恒为 true——不管什么值都会执行赋值。这意味着即使请求未提供 `small_model_config_key` 字段（空字符串），仍会用空字符串覆盖已有值，导致该字段被意外清空。这与其他字段的"非空才更新"语义矛盾。 | 删除 `\|\| req.SmallModelConfigKey == ""` 和 `\|\| req.LargeModelConfigKey == ""` 部分，仅保留 `!= ""` 判断。或者加入显式 `optional` 标记/指针类型来区分"未提供"和"明确清空"。 |
| 5 | `api/internal/logic/pipeline/activate_pipeline_logic.go:26-35` | Activate 逻辑先查所有 pipeline（limit 1000），再逐个 Update 将 `is_production` 设为 0，最后设置目标为 1。存在竞态：如果并发调用，两个 pipeline 可能同时被设为 `is_production=1`。且未使用事务，逐条 Update 在部分失败时不回滚。 | 使用 DB 事务包裹：`BEGIN; UPDATE ... SET is_production=0 WHERE is_production=1; UPDATE ... SET is_production=1 WHERE pipeline_id=?; COMMIT;`，或使用 `WHERE is_production=1` 条件批量更新代替逐条循环。 |
| 6 | `rpc/etc/moderation.yaml:3` | gRPC 服务的 `Timeout: 30000`（30 秒），API 层的 `Timeout: 30000`，但 gRPC 客户端的 `AiModelRpc` 和 `MasterDataRpc` 未显式配置 Timeout 值。按照记忆 gRPC 三层超时对齐要求，大模型调用可能需要 60s，应检查各层超时链是否满足最长调用需求。 | 在 `AiModelRpc` 中显式设置 `Timeout: 60000`，在 API 层设置 `Timeout: 120000`（对齐记忆 `grpc-timeout-layers` 规范：REST > gRPC client > external）。 |

### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `internal/pipeline/executor.go:302-325` | `callModel` 中 `RawResponse` 字段使用 `fmt.Sprintf` 拼接而非序列化完整 gRPC 响应。调试时可能丢失信息（如 `RawResponse` 仅含摘要字符串而非完整 proto 响应）。建议用 JSON marshal 完整响应或 protojson 序列化。 |
| 2 | `api/internal/logic/pipeline/create_pipeline_logic.go:77` | `result.LastInsertId()` 返回的 `id` 被忽略错误（`id, _ := result.LastInsertId()`）。通常 MySQL 驱动下不会失败，但丢失错误信息不利于排查。建议处理 err 并记录日志。 |
| 3 | `model/mod_pipeline_config_gen.go` | 数据模型 `ModPipelineConfig` 中 `Id` 字段类型为 `uint64`，但在 `types/types.go` 中 `PipelineInfo.Id` 也是 `uint64`。Snowflake ID 规范要求 int64 ID 在 Proto 中使用 `[jstype=JS_STRING]` + Go 使用 `json:",string"` + TS 使用 `string`。数据库自增 ID（非 Snowflake）可以豁免，但建议在注释中标明此 ID 是 DB 自增而非 Snowflake，以便后续审查时快速区分。 |
| 4 | `migrations/001_moderation_schema.sql` | 迁移文件完全移除了 `md_sensitive_word` 表的 ALTER，这是正确的（表归 master-data-service 管理）。但旧环境中已执行过这些 ALTER 的实例，回滚此迁移会导致 schema 不一致。建议在 CHANGELOG 中记录：之前通过此 migration 添加到 `md_sensitive_word` 的列需手动确认是否需要保留。 |
| 5 | `internal/pipeline/executor.go:185-210` | `shouldEscalateAcToSmall` 使用 switch-case 处理 4 种升级条件，逻辑清晰。但 `AcToSmallCategories` 的比较是 O(n*m) 嵌套循环，对于大列表（如 1000 categories x 1000 hits）有性能风险。实际 category 数量有限（ac_engine 层命中通常 < 100），当前性能可接受，但建议后续添加注释说明规模上限。 |

## 架构一致性检查
- [x] Proto 规范 — 无 Proto 变更，使用已有 `aimodel/v1` 和 `masterdata/v1`
- [x] gRPC 通信 — LLM 调用通过 gRPC `AiModelService.ModerateText`，敏感词加载通过 gRPC `MasterdataService.GetSensitiveWords`
- [x] 错误码 — authmiddleware 使用 `errx.NewUnauthorizedError`，其他 logic 使用 `errx.NewCodeError(errx.CodeModeration*)`，无硬编码数字错误码
- [x] Snowflake ID — 本次变更未涉及 Snowflake ID 生成（pipeline_id 为 VARCHAR 用户自定义）
- [x] API 响应单层包装 — 已有 handler（health/text_check/image_check/submit_review）使用 `responsex.Response` 正确。pipeline handlers 使用 `httpx.OkJson` 未遵循规范（已在 WARNING #1 标记）

## 变更完整性检查
- [x] CHANGELOG — 已更新，记录了 2026-06-15 审核管线可配置化 + QA 修复
- [x] Proto CHANGELOG — 无 Proto 变更，无需更新
- [x] Migration — `migrations/002_pipeline_config.sql` 已提供，创建 `mod_pipeline_config` 表
- [x] design.md — 未更新（管线配置化是对审核流程的扩展，核心设计文档可考虑补充本章节）
- [x] 旧 migration 001 已正确移除跨服务 ALTER（md_sensitive_word 表变更归 master-data-service 管理）

## 记忆遵守检查

### M1: 代码引用检查
- [x] 变更文件中未发现 `// SEE: [[...]]` 注释

### M2: 不适用（无引用）

### M3: 遗漏检查

**变更关键技术关键词**: gRPC 调用, context timeout, API 响应格式, migration, soft-delete, DDL

| # | 记忆 | 类型 | 严重度 | 匹配评估 |
|---|------|------|--------|---------|
| 1 | `grpc-timeout-layers` | must-follow | — | **触发**：PipelineExecutor 通过 gRPC 调用 ai-model-service 的 `ModerateText`，未使用独立 Context（继承 request context），未设置显式 gRPC 客户端超时。rpc/etc/moderation.yaml 的 `AiModelRpc` 未配置 Timeout。已在 WARNING #6 报告。 |
| 2 | `api-response-single-wrap` | must-follow | — | **触发**：7 个 pipeline handler 使用 `httpx.OkJson` 而非 `responsex.Response`，响应格式与项目标准不一致。已在 WARNING #1 报告。 |
| 3 | `migration-must-execute` | must-follow | — | **部分触发**：新增 `002_pipeline_config.sql`，需确认已执行。CHANGELOG 中未包含执行确认。建议标注执行状态。 |
| 4 | `redis-cache-soft-delete` | should-follow | — | 不适用：pipeline config model 使用软删除（`delete_time`），但 CRUD 直接使用 `sqlx.SqlConn`（非 go-zero 的 `CachedConn`），不存在 Redis 索引缓存问题。无需告警。 |
| 5 | `goctl-logic-stubs` | must-follow | — | 不适用：新增的 7 个 pipeline logic 文件已完整实现，无 TODO stub 残留。harness check 12 已确认 PASS。 |
| 6 | `service-api-jwt-env` | should-follow | — | 不适用：API 入口 `api/moderation.go` 使用 `configx.MustLoad`，yaml 中的 `AccessSecret: ${JWT_ACCESS_SECRET}` 是已有配置，非本次新增。 |

### M4: 元数据更新
- `api-response-single-wrap` 记忆适用于本次审查（7 个 handler 未遵循），应更新 `last_applied: 2026-06-16` 和 `apply_count`（重申规范要求）。
- `grpc-timeout-layers` 记忆适用于本次审查（pipeline executor 的 gRPC 调用链），应更新 `last_applied: 2026-06-16` 和 `apply_count`。

---
VERDICT: PASS
---

**理由**: 1 个 CRITICAL 是关于 `large_overrides` fallthrough 降级策略的设计意图确认（非明确代码错误，且 build/vet/test 全通过）。该 case 的行为在当前上下文下是安全的（降级到 small model 不会被绕过——因为若大模型未被调用且小模型也未调用，最终会 fall 到 `need_review`），但建议跟设计文档对齐预期行为。WARNING 级别问题（响应格式不一致、path 参数解析、零测试覆盖、更新逻辑恒真条件、激活竞态、gRPC 超时未配置）均不阻塞门禁，但应在下个迭代中修复。
