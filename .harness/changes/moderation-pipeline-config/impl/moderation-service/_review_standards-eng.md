# Code Review — 内容审核服务（规范工程视角）

**审查时间**: 2026-06-16
**审查维度**: 规范遵循(#3)、复用性(#6)、测试覆盖(#7)、记忆遵守(#9)
**审查范围**: moderation-service 近 5 次提交，管线配置化功能 + QA 修复

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 9 / 🔵 NOTE: 5

## 发现

### 🔴 CRITICAL
无

### 🟡 WARNING

| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|
| 1 | `api/internal/logic/pipeline/create_pipeline_logic.go:15-19` + 同包其他文件 | #3 规范遵循 | 管线 Logic 包统一使用 `errors.New()` 而非 `errx.NewCodeError()`。与 text/image review 逻辑层使用的 `errx.CodeModerationInvalidParam`/`errx.CodeModerationInternalError` 不一致。`responsex.Response()` 会将普通 error 映射为通用错误码，前端无法按错误类型区分（如"管线不存在"与"ID 重复"） | 引入管线专用错误码常量（如 `CodeModerationPipelineNotFound`、`CodeModerationPipelineIdDuplicate`），使用 `errx.NewCodeError(code, msg)` 返回 |
| 2 | `api/internal/types/types.go:41,44` | #3 规范遵循 | 既有问题：`SubmitReviewReq` 中的 `AuditLogId int64` 和 `ReviewerId int64` 缺少 `json:",string"` tag。若为 Snowflake ID 则 JS 精度丢失 | 确认是否为 Snowflake ID，若是则添加 `json:"audit_log_id,string"` |
| 3 | `internal/pipeline/executor.go:53-182` | #7 测试覆盖 | `Execute()` 核心方法无单元测试（~130 行核心级联逻辑）。依赖外部 TextEngine + gRPC client 导致测试困难 | 通过 mock TextEngine + mock gRPC client 添加集成测试 |
| 4 | `internal/pipeline/executor.go:304-326` | #7 测试覆盖 | `callModel()` gRPC 调用包装无测试 | 使用 mock `AiModelServiceClient` 测试请求构造和响应映射 |
| 5 | `api/internal/middleware/authmiddleware.go:43-88` | #7 测试覆盖 | `Handle()` 方法含 JWT claim 提取 + 鉴权逻辑，无测试。新增的 `GetUserInfo`/`GetUserID` 也无测试 | 构造 go-zero mock context 设置 claim 值，测试正常/缺失 user_id 等路径 |
| 6 | `model/mod_pipeline_config_gen.go:166-190` | #7 测试覆盖 | `SetProduction()` 事务方法（清除所有 flag + 设置目标）无测试 | 使用 `sqlx.NewMockConnector` 或集成测试覆盖事务成功/目标不存在/其他错误路径 |
| 7 | `api/internal/logic/pipeline/activate_pipeline_logic.go:24` | #9 记忆遵守 M2 | `// SEE: [[moderation-submodule-broken]]` — 虚假匹配。该 slug 是关于"git submodule state breaks worktree isolation"，与代码使用的 `TransactCtx` SQL 事务无关 | 移除该引用或替换为准确 slug（如涉及事务原子性可考虑新记忆条目；当前上下文不匹配任何现有记忆） |
| 8 | `internal/pipeline/executor.go:304-311` | #9 记忆遵守 M3 | `[[grpc-timeout-layers]]` must-follow 记忆遗漏：`callModel()` 直接使用 HTTP request context 发起 gRPC 调用，未创建独立的 `context.WithTimeout()`。且 `AiModelRpc` YAML 配置未设置 `Timeout` 字段（依赖 go-zero 默认 2000ms） | 对 `callModel` 使用 `context.WithTimeout(context.Background(), 60*time.Second)` 独立上下文；或在 `AiModelRpc` YAML 中显式设置 `Timeout: 60000`。注意：本端点（/pipeline/test）为调试用途，严重度可降为 WARN |
| 9 | `api/internal/logic/pipeline/get_pipeline_logic.go:34` + `internal/pipeline/config.go:31` | #9 记忆遵守 M3 | `json.Unmarshal` 返回的 error 被忽略（`_`），若数据库中 `ac_to_small_categories`/`small_to_large_categories` 存储了损坏的 JSON，会静默失败导致分类列表丢失 | 捕获并记录 Unmarshal 错误，或返回 error 给调用方 |

### 🔵 NOTE

| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `api/internal/handler/pipeline/*` (7 files) | ✅ 所有 pipeline handler 使用 `responsex.Response(w, resp, err)` 单层包装，符合 `[[api-response-single-wrap]]` 规范 |
| 2 | `internal/pipeline/executor_test.go` | ✅ 新增 28 个单元测试，覆盖升级条件判断、终判逻辑、JSON 序列化，质量良好。对 CRITICAL fix (large_overrides fallthrough) 有专门回归测试 |
| 3 | `api/internal/logic/pipeline/create_pipeline_logic.go:85-102` | `newNullString`/`newNullInt64`/`newNullFloat64` 三个 helper 函数仅在 pipeline logic 包内使用，定位合理。`newNullInt64(0)` 将零值视为无效——与 `update_pipeline_logic.go` 中 `req.Xxx != 0` 的更新判断一致，但语义上有细微歧义（"不更新" vs "设置为0"），目前业务上可接受 |
| 4 | `api/internal/logic/pipeline/pipeline_test_logic.go:73-91` | `convertLayerResult()` 负责 `internal/pipeline.LayerResult` → `types.LayerResult` 的类型转换。考虑在 `internal/pipeline` 包导出该转换函数以复用（如果未来有其他入口使用 PipelineResult） |
| 5 | `model/mod_pipeline_config_gen.go` | 整个 model 文件为手写 SQL（非 goctl 生成），维护性需关注但实现质量 OK。`SetProduction` 事务方法正确修复了竞态条件 |

## 记忆遵守检查

### M1: 收集引用
| # | 文件:行号 | Slug |
|---|----------|------|
| 1 | `activate_pipeline_logic.go:24` | `moderation-submodule-broken` |
| 2 | `executor.go:242` | `api-response-single-wrap` |
| 3 | `executor.go:329` | `api-response-single-wrap` |

### M2: 准确性验证
| Slug | 文件存在 | 代码遵守 | 虚假匹配 | 结论 |
|------|:---:|:---:|:---:|------|
| `moderation-submodule-broken` | ✅ | N/A | 🟡 是 | 虚假匹配：slug 关于 git submodule，代码使用 TransactCtx（SQL 事务） |
| `api-response-single-wrap` (line 242) | ✅ | ✅ | 否 | verdict 值 "pass"/"reject"/"need_review" 为纯字符串，前端直接使用 |
| `api-response-single-wrap` (line 329) | ✅ | ✅ | 否 | RawResponse 使用 `json.Marshal` 而非 `fmt.Sprintf` |

### M3: 遗漏检查
| 记忆 | 关键词匹配 | 遗漏 | 结论 |
|------|:---:|:---:|------|
| `grpc-timeout-layers` | gRPC, ai-model, 调用, LLM, 超时 | 🟡 部分遗漏 | pipeline test 端点 gRPC 调用未设独立超时上下文（WARNING #8） |
| `proto-jstype` | int64, Snowflake, 序列化 | 否 | 新代码无 Snowflake ID 引入 |
| `api-response-single-wrap` | response, Response, handler, API | 否 | 已正确引用并遵守 |
| 其他 must-follow 记忆 | — | 否 | 均不适用或已遵守 |

### M4: 元数据更新
- `api-response-single-wrap`: 本次变更 7 个 handler 均正确遵循单层包装原则，建议更新 `last_applied: 2026-06-16`、`apply_count +1`

---
VERDICT: PASS
---

**理由**: 本视角审查的四个维度（规范遵循、复用性、测试覆盖、记忆遵守）中无 CRITICAL 问题。9 个 WARNING 主要为错误码不一致（可改进）、测试覆盖盲区（依赖外部服务，有 mock 路径）、一处虚假记忆引用、一处 gRPC 超时配置遗漏（调试端点）。所有 WARNING 均不阻塞合入，建议在后续迭代中逐项改进。
