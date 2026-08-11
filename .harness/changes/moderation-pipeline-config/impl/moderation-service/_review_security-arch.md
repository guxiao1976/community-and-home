# Code Review — 内容审核服务（安全架构视角）

**审查时间**: 2026-06-16
**审查维度**: 架构一致性(#1)、安全性(#5)、变更完整性(#8)
**审查范围**: git diff 2f1aa24..HEAD + 未暂存变更，涉及 42 个文件（含 22 个新增文件）

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 3 / 🔵 NOTE: 4

## 发现

### 🔴 CRITICAL
| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|
| （无） | — | — | — | — |

### 🟡 WARNING
| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| 1 | `docs/design.md` | #8 变更完整性 | service 自身 `docs/design.md` 未更新，未包含新增的管线配置化功能（PipelineExecutor、PipelineConfig 类型、7 个 REST 端点、mod_pipeline_config 表）。CHANGELOG 引用了外部 spec `docs/superpowers/specs/2026-06-15-moderation-pipeline-config-design.md`，但服务级 design.md 应同步更新，补充管线配置管理、升级条件枚举、终判逻辑等的说明。 | 在 `docs/design.md` 中新增章节描述可配置管线系统，或至少在「已知问题」中注明管线已配置化 |
| 2 | `api/internal/logic/pipeline/create_pipeline_logic.go:656-667` | #5 安全性 | `newNullInt64(0)` 和 `newNullFloat64(0)` 将零值视为 NULL。当 `AcToSmallSeverity` 或 `SmallToLargeConfidenceThresh` 的合法值为 0 时（如 severity=0 表示不设置阈值），零值会被错误地存储为 NULL，导致数据丢失。 | 改为使用显式的 `has` bool 参数（如 `newNullInt64(v int64, has bool)`），或要求前端在 omitempty 场景不发送该字段而非发送零值 |
| 3 | `api/internal/types/types.go:134,146` | #1 架构一致性 | `CreatePipelineResp.Id` 和 `PipelineInfo.Id` 类型为 `uint64`。虽然这是 DB 自增 ID（非 Snowflake），但项目中所有对外暴露的数字 ID 统一使用 `int64` + `json:",string"` 标签会更一致，且避免 JS 端大整数精度丢失（自增 ID 在数据量大时可能超过 2^53）。 | 将 `Id` 统一为 `int64`，添加 `json:",string"` 标签 |

### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `internal/pipeline/executor.go:3019` | `smLayer.RawResponse = smResult.RawResponse` — `RawResponse` 包含完整的 LLM 返回内容，可能包含用户输入文本的回显。当前仅序列化 `is_safe/confidence/model/template` 四个字段（见 `toRawResponseJSON`），不包含原文，这是安全的。但如果未来扩展 `RawResponse` 内容，注意不要回显用户 PII 敏感文本。 |
| 2 | `api/internal/types/types.go:137-143` | 新增 `SubmitReviewReq`/`SubmitReviewResp` 类型但对应的 `submit_review_handler.go` 仍为 stub 实现（返回 "not implemented"）。如果类型已定义但无用，建议添加注释说明预留用途，避免前端误用。 |
| 3 | `api/internal/logic/pipeline/pipeline_test_logic.go:870-881` | ad-hoc 模式创建临时 PipelineConfig 时硬编码了升级条件（`any_hit`/`confidence_lt`/`last_model_wins`），这些是合理的默认值。但如果未来升级条件的合法枚举值扩展，此处应与 CreatePipeline 使用相同的常量默认值来源。 |
| 4 | `migrations/002_pipeline_config.sql` | 迁移编号使用顺序数字 `002`，当前项目也使用 `001`。建议未来逐渐过渡到时间戳命名（如 `20260615_pipeline_config.sql`），避免并行开发时编号冲突。 |

## 架构一致性检查

- [x] **Proto 规范** — 无新增 Proto 定义；已有 `aimodel/v1` 和 `masterdata/v1` Proto 通过 gRPC 使用，符合 Proto 统一管理的规则
- [x] **gRPC 通信** — WordStore 通过 gRPC 调用 master-data-service（C2），LLM 客户端通过 gRPC 调用 ai-model-service（C1），符合「服务间通信仅 gRPC」规则
- [x] **跨服务 DB 访问** — ✅ 已消除：WordStore 不再直连 masterdata_db（DB 直读代码已删除），迁移 001 也清理了 masterdata_db 的 ALTER 语句（归属 master-data-service）。PipelineExecutor 操作 moderation_db（本服务自己管理），合规
- [x] **configx.MustLoad** — `api/moderation.go:23` 使用 `configx.MustLoad` 加载配置，符合全局规则
- [x] **错误码规范** — 所有错误码使用 `errx.NewCodeError` 或 `errx.NewUnauthorizedError` 命名常量，无硬编码数字
- [x] **YAML 密钥** — 硬编码凭据已替换为环境变量占位符（`${JWT_ACCESS_SECRET}`、`${MYSQL_USER}`、`${MYSQL_PASSWORD}`、`${REDIS_PASSWORD}`），符合安全规则

## 变更完整性检查

- [x] **CHANGELOG.md** — 已更新，覆盖 C1/C2 架构修复、管线配置化、QA 修复、错误码规范化等全部变更
- [x] **Proto CHANGELOG** — 无 Proto 变更，无需记录
- [x] **Migration** — `migrations/002_pipeline_config.sql` 已提供 `mod_pipeline_config` 表的完整 DDL；`migrations/001_moderation_schema.sql` 已清理不再修改 masterdata_db
- [⚠] **design.md** — 服务设计文档未同步更新，缺失管线配置化功能的描述（WARNING #1）

## 安全专项检查

| 检查项 | 结果 | 说明 |
|--------|:---:|------|
| 硬编码密钥 | ✅ PASS | YAML 配置全部使用环境变量 `${...}`，代码中无硬编码密钥（`harness-checks.sh` 也确认） |
| SQL 注入防护 | ✅ PASS | 所有 SQL 查询使用参数化占位符 `?`，通过 `ExecCtx`/`QueryRowCtx`/`QueryRowsCtx` 传递参数。审计日志表的 `fmt.Sprintf` 仅用于表名插值（表名来自常量 `"mod_audit_log"`） |
| 输入校验 | ✅ PASS | `TextCheck`: content 非空 + ≤10000 字符限制；`ImageCheck`: file 必填 + 内容类型白名单 + ≤10MB；`PipelineTest`: content 非空；`PipelineId`: 路径参数 + DB 查询前校验存在性 |
| JWT 鉴权 | ✅ PASS | Pipeline 路由通过 `rest.WithJwt` + `AuthMiddleware` 双重保护。AuthMiddleware 已从桩实现升级为真正的 JWT claims 提取（user_id/username/administrative_code），且 user_id=0 时拒绝访问 |
| 敏感信息日志 | ✅ PASS | 错误日志不包含用户输入原文（仅记录 pipeline_id、错误描述等结构化信息）。审核日志 `ContentSummary` 截断至 100 字符 |

---
VERDICT: PASS
---

**理由**: 我负责的三个维度（架构一致性、安全性、变更完整性）中无 CRITICAL 问题。C1/C2 架构债务（HTTP→gRPC、DB直读→gRPC）已完全消除，硬编码凭据已替换为环境变量，SQL 注入防护到位，输入校验完善。3 个 WARNING 均为非阻塞性建议（design.md 同步更新、NULL 零值边界处理、ID 类型一致性）。服务安全架构合规。
