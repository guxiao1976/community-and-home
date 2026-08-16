# Plan Review — content-post-generalization（业务有效性视角）

**审查维度**: 业务自洽、非功能（安全/性能/兼容）、合规、架构冲突/技术债/依赖风险
**审查版本**: P1.3 (fallback:r0:rc1) — 按磁盘最新内容独立审查
**审查对象**: proposal.md + specs/content-post-{publish,read,moderation,permission,attachment-security}/spec.md + 设计文档 docs/superpowers/specs/2026-08-16-content-post-design.md + 现网 notices 模型/迁移

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 4

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | content-post-read/spec.md REQ-CPR-1（列表排序）+ REQ-CPR-3（跑马灯窗口/排序）；content-post-publish/spec.md REQ-CPB-1（published_at 创建置 NULL，由审核通过回调置 now）| **published_at 本期恒 NULL → GetMarqueeNotices 恒空 + 列表排序不确定（spec 内部矛盾）**。REQ-CPB-1 明确 published_at 由「审核通过回调」置 now（REUSE:notice-D27），但本期 D16/D18 明确无审核消费者、无回写回调（status 默认 approved）——因此本期所有 content_posts.published_at 恒为 NULL。REQ-CPR-3 跑马灯 15 天窗口 `published_at >= now-15*24h` 恒为 false → 本期在途交付物（D5，验收标准「GetMarqueeNotices 新 RPC 可用」）功能不可用；REQ-CPR-1 列表按 published_at DESC 排序对全 NULL 值不确定、分页不稳定（其「先排除未过完整性门禁的 NULL 行」假设在本期不成立——本期所有新帖均过门禁且均 NULL published_at）。对比原 notice 设计（有 Redis 消费者回调置 published_at），本变更复制其排序假设但删除了回调。另：REQ-CPR-3 依赖 `is_pinned desc` 排序，但 REQ-CPB-1 字段清单未列出 is_pinned 保留（RENAME 物理保留，spec 未声明）。| 明确本期（无消费者）的锚定语义：submit 且 status 默认 approved 时置 `published_at = now`（通过锚定=提交时刻，消费者上线后仍按 D27 覆盖），并在 REQ-CPB-1 显式列出 `is_pinned` 为保留列；或显式定义不依赖 published_at 的回退排序（如 created_at DESC）并显式接受跑马灯本期为空。二选一，消除 spec 内部矛盾。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 2 | content-post-publish/spec.md REQ-CPB-4（draft 可编辑 attachments）/ REQ-CPB-8（attachment_count 完整性判定）/ REQ-CAS-5 | **attachment_count 在 draft 编辑附件集合时的重算不变量未定义**。REQ-CPB-4 允许 draft 增删 attachments，但 spec 未规定 attachment_count 同步重算。若沿用创建时计数，draft 删附件后 count(approved) < attachment_count → 帖子提交后永远过不了完整性门禁（REQ-CPB-8）静默不可见，属数据一致性隐患。 | 显式声明「附件集合每次变更时 attachment_count 重算」为不变量；或读路径判定用实时 count 对账存储计数兜底。 |
| 3 | content-post-publish/spec.md REQ-CPB-7（best-effort + MAY retry）/ content-post-moderation/spec.md REQ-CPM-1/REQ-CPM-5 | **Kafka best-effort 推送 + 「MAY retry」为弱约束 → 审核消息可静默丢失，形成审核盲区**。推送失败仅打日志且「MAY be retried」，本期无消费者无感知；未来消费者（D18）上线后，这些帖子永久无审核记录却因默认 approved 一直可见，与变更核心目标（内容级审核）冲突。 | 把补偿/重推定义为准入门禁（如 content_posts 落库待推标记 + 定时扫描重推，或落库 outbox），至少将 MAY 改为 SHALL 并登记「推送失败=该帖永不审核」的显式业务风险与可观测指标。 |
| 4 | content-post-publish/spec.md REQ-CPB-1（字段清单）/ REQ-CPB-5（JWT 身份信任边界） | **content_posts 模型描述缺失 role/publisher 列，publisher 展示名来源未定义（展示名伪造向量）**。RENAME 保留 notices 的 `role VARCHAR(20) NOT NULL` / `publisher VARCHAR(100) NOT NULL`（migration/001_initial.sql:15-16），但 REQ-CPB-1 字段清单不含这两列；现有实现 `Publisher: in.Publisher` 直接信任请求体展示名（createnoticelogic.go:59）。新 spec 只约束 role/publisher_id 从 JWT 派生（REQ-CPB-5），未约束展示名来源——若沿用现状，发布者可任意伪造展示名（冒充他人/官方），与「JWT 实际身份为准」的安全姿态不一致。 | REQ-CPB-1 显式列出 role/publisher 列保留及来源：role 由 RBAC→发布角色映射派生、publisher 展示名取用户真实档案（禁请求体信任），并补充伪造展示名被纠正的异常场景。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 5 | 存量 notices 迁移后不可见（D2 用户拍板，spec REQ-CPB-1 场景3 已承认不返回）：上线瞬间存量通知全部从列表/详情/跑马灯消失，属业务可用性回归但已拍板。建议在 proposal 影响范围/风险节显式登记「存量通知迁移后不可见」影响面 + BACKLOG 回填项，避免上线侧措手不及。 |
| 6 | content_posts.status 迁移加列默认值未定：INSERT 必须显式置 1（approved）；legacy 行 status 为 NULL/0 因无 scope 行不可见，无影响但实现需注意。 |
| 7 | Kafka 消息 attachments[].review_status 在推送时为默认 approved（审核前值），与「审核结果」语义不符易误导；file_url 为短时效预签名 URL，消费者读取时需再生（契约已声明可再生，仅提示）。 |
| 8 | GetPublishPermission 依赖的权限码 421 名称仍为 `community:notice:create-api`（未随 CreateContentPost 改名），技术债提示，可在设计评审确认是否保留旧码名。 |

## 问题跟踪表

| # | 状态 |
|---|------|
| 1 | 待修复 |
| 2 | 待修复 |
| 3 | 待修复 |
| 4 | 待修复 |

---
VERDICT: **REVISION**（存在 1 项 MUST FIX，spec 内部矛盾：published_at 本期恒 NULL 使在途交付物跑马灯不可用 + 列表排序不确定）
---
