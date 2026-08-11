# Code Review — 内容审核服务（设计业务视角）

**审查时间**: 2026-06-16
**审查维度**: 设计一致性(#2)、代码质量(#4)、变更完整性(#8部分)
**审查范围**: 管线配置化功能 (2026-06-15) + QA 修复 (2026-06-16)，涉及 ~20 个文件

## 摘要
- 变更文件: ~20 (pipeline 包、API logic/handler、model、migration、types、middleware)
- 🔴 CRITICAL: 0 / 🟡 WARNING: 5 / 🔵 NOTE: 5

---

## 发现

### 🔴 CRITICAL

| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|

（无 CRITICAL 问题）

### 🟡 WARNING

| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| 1 | api/internal/svc/service_context.go:106 | #4 代码质量 | `getAiModelClient()` 被无条件调用创建 `PipelineExecutor`，内部使用 `zrpc.MustNewClient` 会在 `AiModelRpc` 未配置时 panic。此前 `getAiModelClient()` 仅在模板 ID 存在时才调用（有 guard），此处新增了无 guard 的调用路径，是启动行为回归。 | 在调用前加 guard：`if c.AiModelRpc.Etcd.Key != "" \|\| len(c.AiModelRpc.Endpoints) > 0` 时才创建 pipelineExecutor，否则设为 nil 并在测试端点中返回友好错误 |
| 2 | api/internal/logic/pipeline/create_pipeline_logic.go:39 | #4 代码质量 | `existing, _ := l.svcCtx.PipelineModel.FindOneByPipelineId(...)` 忽略了 DB 查询错误。若发生 DB 连接异常等非 "not found" 错误，代码会误判为"管线不存在"并继续执行 Insert，可能触发重复键约束（数据完整性依赖 DB 约束兜底，但错误消息会误导）。 | 应显式处理错误：`if err != nil && err != sql.ErrNoRows { return nil, err }` |
| 3 | api/internal/logic/pipeline/activate_pipeline_logic.go:26-30 | #4 代码质量 | `SetProduction` 返回任何错误都被包装为 `ErrPipelineNotFound`（"管线配置不存在"）。如果错误是 DB 连接失败或事务超时，用户会收到误导性错误消息。 | 区分处理：`errors.Is(err, sql.ErrNoRows)` → `ErrPipelineNotFound`；其他错误 → 保留原始错误或专用内部错误 |
| 4 | internal/pipeline/config.go:33-35, api/internal/logic/pipeline/get_pipeline_logic.go:34-37 | #4 代码质量 | 多处 `json.Unmarshal` 错误被忽略。如果 DB 中存储的 JSON 因历史原因损坏，代码会静默返回空的 `acCats`/`slCats` 切片，升级规则将不生效，排查困难。 | 至少记录 warn 日志：`if err := json.Unmarshal(...); err != nil { logx.Errorf("corrupt JSON: %v", err) }` |
| 5 | api/internal/logic/pipeline/create_pipeline_logic.go:45,49, update_pipeline_logic.go:68,78 | #4 代码质量 | `json.Marshal` 错误被忽略（`b, _ := json.Marshal(...)`）。虽然 `[]string` 类型的 Marshal 不会出错，但代码模式传递了"错误可忽略"的不良信号。 | 显式处理错误；或封装为辅助函数 `mustMarshalJSON(v interface{}) string` |

### 🔵 NOTE

| # | 文件:行号 | 维度 | 建议 |
|---|----------|------|------|
| 1 | docs/superpowers/specs/2026-06-15-moderation-pipeline-config-design.md §"生产环境集成" | #2 设计一致性 | 设计文档明确承诺"生产环境集成"——当 `is_production=1` 时，`TextEngine.Check()` 应读取管线配置中的升级规则替换硬编码 `HighConfThreshold`。目前 `PipelineExecutor` 仅服务于 `/pipeline/test` 测试端点，实际生产审核流量仍走原有的 `TextEngine.Check()` 硬编码路径。建议在后续迭代中完成此集成，并在 CHANGELOG/design.md 中标注当前实现范围。 |
| 2 | internal/pipeline/executor.go:81-84 | #4 代码质量 | AC 引擎结果中的 `Categories` 来自多条 `MatchDetail` 的 cat 字段拼接，未做去重。同一类别可能在多条匹配中出现，导致 `categories` 数组有重复值。建议添加去重逻辑。 |
| 3 | rpc/internal/svc/servicecontext.go | #2 设计一致性 | gRPC 服务上下文缺少 `PipelineModel` 和 `PipelineExecutor` 的注入。虽然管线配置管理是管理面功能（REST-only 合理），但如果未来 gRPC 客户端也需要按生产管线执行审核，此处需要补充。建议在 design.md 中明确标注"管线配置为 REST 管理面功能，gRPC 层暂不涉及"。 |
| 4 | migrations/002_pipeline_config.sql | #8 变更完整性 | Migration 未提供回滚脚本。当前迁移仅新增表（无破坏性变更），风险较低，但若有回滚需求（如生产环境部署失败），缺少 `DROP TABLE IF EXISTS mod_pipeline_config` 的回滚 SQL。建议在 migrations/ 目录中增加 `002_pipeline_config_rollback.sql`。 |
| 5 | api/internal/logic/pipeline/create_pipeline_logic.go:97-101 | #4 代码质量 | `newNullFloat64`/`newNullInt64` 将零值映射为 SQL NULL，导致 DB schema 的 `DEFAULT 0.90` 等默认值实际不被使用（INSERT 总是发送显式 NULL，覆盖 DEFAULT）。当前代码在读取侧（`FromDBModel`）将 NULL 重新映射回 0.90，行为正确但绕过了 DB 层的默认值。建议：要么在 INSERT 时不发送该列（让 DB DEFAULT 生效），要么在应用层统一设置默认值。 |

---

## 设计一致性检查 (#2)

- [x] 数据模型：`mod_pipeline_config` 表结构与设计文档一致（字段/类型/默认值/索引均匹配）
- [x] API 端点：7 个 REST 端点（CRUD + activate + test）与设计文档完全一致
- [x] 请求/响应类型：`PipelineTestReq`/`PipelineTestResp`/`LayerResult` 与设计文档一致
- [x] 升级条件枚举：4 种 AC→Small 条件 + 4 种 Small→Large 条件与设计文档一致
- [x] 终判逻辑枚举：`last_model_wins`/`large_overrides`/`small_overrides` 与设计文档一致
- [x] 执行流程：PipelineExecutor.Execute() 的 AC→Small→Large 三层执行顺序与设计文档流程图一致
- [x] fallthrough bug 修复（executor.go:251）：`large_overrides` 大模型未调用时不再 fallthrough 到 `small_overrides`，正确回退到 `lastModelWins()`
- [ ] 生产环境集成：设计文档 §"生产环境集成" 承诺的 `TextEngine.Check()` 适配未实现（见 NOTE #1）

## 代码质量检查 (#4)

- [x] 边界条件处理：`shouldEscalateSmallToLarge` 处理了 nil、not_called 边界；`lastModelWins` 处理了所有层未调用的回退
- [x] 错误处理：各 handler 统一使用 `responsex.Response`，错误码使用命名常量
- [x] 资源管理：PipelineExecutor 不持有需显式释放的资源，gRPC 连接由 go-zero zrpc 管理生命周期
- [x] 并发安全：`SetProduction` 使用事务 (`TransactCtx`) 原子化清除+设置生产管线标志，消除竞态
- [x] 竞态条件修复：activate 从 read-modify-write 改为 `SetProduction` 事务方法
- [ ] 部分 `json.Marshal`/`Unmarshal` 错误被忽略（见 WARNING #4、#5）
- [ ] `FindOneByPipelineId` DB 错误被忽略（见 WARNING #2）
- [ ] `SetProduction` 错误包装过于笼统（见 WARNING #3）
- [ ] `getAiModelClient()` 无条件调用可能 panic（见 WARNING #1）

## 变更完整性检查 (#8)

- [x] CHANGELOG 已更新：2026-06-15 管线配置化 + 2026-06-16 QA 修复两个条目，信息完整
- [x] Migration 提供：`002_pipeline_config.sql` 完整且与设计文档一致
- [x] Migration 安全性：仅新增表，不影响现有数据；`ENGINE=InnoDB` 支持在线 DDL
- [x] Proto 变更：无（复用已有 `aimodel/v1.ModerateText` gRPC 接口）
- [x] 前端变更：已记录（ModerationConfigTest.vue + 5 个组件 + 路由/菜单更新）
- [ ] 无 Migration 回滚脚本（见 NOTE #4）

---

**VERDICT: PASS**

**理由**: 我负责的维度（设计一致性 #2、代码质量 #4、变更完整性 #8）中未发现 CRITICAL 问题。5 个 WARNING 均不涉及数据丢失或安全风险。已记录的 QA 修复（fallthrough bug、竞态条件、恒真条件、响应格式不一致）均已在代码中正确体现。
