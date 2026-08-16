# Plan Review — content-post-generalization（覆盖完整性 视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别 / NEEDS CLARIFICATION 遗漏
**审查版本**: fallback:r0:rc1（首轮，无历史轮次缓存）
**审查对象**: proposal.md + specs/content-post-publish / read / moderation / permission / attachment-security 共 5 个 spec.md
**核对基线**: .change.yaml、docs/superpowers/specs/2026-08-16-content-post-design.md、community-hub-service/migration/001_initial.sql、002_add_moderation_status.sql、model/notice.go、api-proto/api/community/v1/community.proto

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3

说明：本 change 无 `request.md`（以 proposal.md + .change.yaml 作为需求基线）。决策点 D1-D18 + REUSE:notice-D1/D3/D4/D5/D11/D12/D13/D14/D19/D23/D24/D25/D27/D29/D30/D31/D32 均已映射到各 Requirement 且有正/异常场景，决策覆盖整体完整。主要缺口集中在「两阶段状态机的数据模型映射」「本期无消费者路径下 published_at 的置值路径」两个决策点，以及若干保留字段与边界场景。

## 发现

### 🔴 MUST FIX
| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| MF-1 | content-post-publish/spec.md REQ-CPB-1 + REQ-CPB-4 | **两阶段状态机（draft/submitted）的持久化表示未定义**。REQ-CPB-1 定义 `status TINYINT` 枚举仅为 0=pending/1=approved/2=rejected/3=withdrawn，不含 draft/submitted；REQ-CPB-4 却引入 draft→submitted→approved 完整状态机。draft 与 submitted 落哪个字段、如何区分存储（draft=pending? submitted=pending? 需独立 publish_state 列?）完全未定义。proposal 仅列 5 个 RPC（CreateContentPost/ListContentPosts/GetContentPost/DeleteContentPost/UpdateContentPost），无 SubmitContentPost，则 draft→submitted 转换的触发载体（UpdateContentPost 某状态/动作字段?）与 submitted 触发 Kafka 推送（REQ-CPB-7）的机制均未定义。设计文档 §2.1/§3.1 同样无 draft/submitted。实现者无法确定性落地该验收标准「两阶段发布：draft 可编辑 → submitted 不可编辑但可删 →（默认 approved）」。 | 明确 draft/submitted 的存储映射（扩展 status 枚举或新增 `publish_state` 列）；定义 draft→submitted 转换的触发载体（如 UpdateContentPost 携带 submit 动作/状态字段）与 submitted 时推送 Kafka 的时机；定义 draft 更新下 title/text/attachments/scope 的可改范围与 attachment_count 重算规则。 |
| MF-2 | content-post-publish/spec.md REQ-CPB-1 场景1 + content-post-read/spec.md REQ-CPR-1/REQ-CPR-3 + content-post-moderation/spec.md REQ-CPM-5 | **本期无消费者路径下 published_at 恒为 NULL → 跑马灯恒空 + 列表排序锚点失效（业务不可用）**。REQ-CPB-1 场景1 明确 published_at 创建时置 NULL、「由审核通过回调置值」；REQ-CPM-5 明确本期不实现消费者（D18）。则本期创建的所有帖 published_at 恒为 NULL：REQ-CPR-3 跑马灯窗口 `published_at >= now-15*24h` 在 SQL 中对 NULL 不成立 → 跑马灯恒空（与验收「跑马灯返回 ≤10 条置顶优先」直接矛盾）；REQ-CPR-1 排序锚点 `published_at DESC` 全为 NULL（DESC 被扰动）。REUSE:notice-D27「审核锚定」与 D16「无消费者默认 approved」在本期路径下自相矛盾，两个新增 RPC 之一的 GetMarqueeNotices（D5）按字面实现将永远返回空。 | 明确本期 published_at 的置值路径，二选一：(a) status 默认 approved 时提交即置 published_at=NOW()（审核锚定的隐式通过），并修订 REQ-CPB-1 场景1「published_at 置空」表述；或 (b) 读路径对 gate 通过但 published_at NULL 的帖回退 created_at 做排序/窗口锚点。并同步 REQ-CPR-3 窗口语义。 |
### 🟡 SHOULD FIX
| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| SF-1 | content-post-publish/spec.md REQ-CPB-1 | **content_posts 保留字段清单不完整**：REQ-CPB-1 只列 title/text/published_at/publisher_id，但读路径依赖 `is_pinned`（REQ-CPR-3 跑马灯 `order by is_pinned desc`）、`role`/`publisher`（REQ-CPR-1 可选 role 过滤，且 001_initial.sql 中二者为 NOT NULL）、`deleted_at`（REQ-CPR-1「排除软删帖」）。模型迁移遗漏清单导致实现者不确定这些列是否保留/是否仍 NOT NULL 写入。 | 在 REQ-CPB-1 补全保留字段全集（is_pinned/role/publisher/deleted_at/created_at/updated_at）及约束处理；明确新插入时 role/publisher 的取值来源（role 从实际角色派生已有，publisher 展示名未定义来源）。 |
| SF-2 | content-post-publish/spec.md REQ-CPB-4 | **draft 编辑下附件增删的 attachment_count 重算无场景覆盖**：REQ-CPB-4 允许 draft 更新 attachments，但无场景覆盖「draft 移除 N 个附件后 attachment_count 是否/如何递减」。D15 完整性判定依赖该计数，陈旧计数会导致已提交帖 gate 失效。 | 增加 draft 附件增删场景，明确 attachment_count 在 draft 编辑时与附件行同事务重算，并在提交（submitted）时冻结。 |
| SF-3 | content-post-publish/spec.md REQ-CPB-5 / content-post-permission/spec.md REQ-CPP-2 | **REQ-CPB-5 三个边界只有文本、无异常场景**：(1) community_ids 与 division_id 同时传 → 互斥拒绝；(2) division_id 仅限 community_admin，其他角色传入 → 拒绝；(3) 目标数 >100 → 080003。REQ-CPP-2 中 committee 作为可发布角色（D6）无正向场景（仅有 property_admin/community_admin/owner）。 | 为上述边界各补 1 个异常场景；补 committee 本小区发布正向场景。 |
| SF-4 | content-post-publish/spec.md REQ-CPB-4 / content-post-read/spec.md REQ-CPR-1 | **删除（撤回）落库语义未定 + Update/Delete RPC 契约无独立 Requirement**：`status=withdrawn(3)` 在模型定义但无任何场景使用；DeleteContentPost 走软删（deleted_at，现 model/notice.go SoftDelete 语义，REQ-CPR-1「排除软删帖」）还是置 status=withdrawn？二者关系未说明。proposal 列的 UpdateContentPost/DeleteContentPost 无专属 Requirement 定义请求/响应契约与可改字段集（仅散见 REQ-CPB-4 场景）。 | 明确 delete 的落库路径（deleted_at vs status=withdrawn，或二者）及对 scope/attachment 行的影响；补 UpdateContentPost/DeleteContentPost 契约 Requirement（可改字段、不可改字段、返回结构）。 |
### 🔵 INFO
| # | 建议 |
|---|------|
| I-1 | content-review topic 无消费者阶段消息堆积的保留策略未说明；REQ-CPB-7 提到补偿/重试「MAY」但本期是否实现补偿路径未定。建议明确 topic retention 与本期不做补偿。 |
| I-2 | 「存量 notices 行不迁移 → 存量通知在移动端读路径全量不可见」的用户可见影响仅在 REQ-CPB-1 存量场景声明，proposal 风险评估未单列该业务影响。建议在 proposal 风险/验收显式登记。 |
| I-3 | proto 中 `UpdateNoticeModerationStatus` RPC（现 community.proto:52）在改名后是否保留/移除未在 proposal api-proto 影响范围声明（REQ-CPM-4 仅说消费者不再回调）。建议在 api-proto 影响范围明确该 RPC 处理（保留兼容 vs 移除）。 |

## 问题跟踪表
| 问题 | 状态 |
|------|------|
| MF-1 状态机存储映射 | 待修复 |
| MF-2 published_at 置值路径 | 待修复 |
| SF-1 保留字段清单 | 待修复 |
| SF-2 attachment_count 重算 | 待修复 |
| SF-3 边界场景缺失 | 待修复 |
| SF-4 删除语义/RPC 契约 | 待修复 |

---
VERDICT: REVISION
---
