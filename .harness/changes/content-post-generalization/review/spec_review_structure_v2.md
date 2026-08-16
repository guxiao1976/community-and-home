# Plan Review — content-post-generalization（结构合理性视角）

**审查维度**: 职责边界 / 一致性（proposal 影响范围 ↔ specs 职责边界；各 capability 职责清晰无重叠）
**审查版本**: P1.3 fallback:r1:rc1
**审查时间**: 2026-08-16
**审查对象**: proposal.md + specs/{content-post-publish, content-post-read, content-post-moderation, content-post-permission, content-post-attachment-security}/spec.md
**核验基准**: services/permission-service/scripts/init_permissions.sql、services/community-hub-service/migration/001_initial.sql、api-proto/api/community/v1/community.proto、docs/superpowers/specs/2026-08-16-content-post-design.md

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 3

## 上一轮（r0）MUST FIX 修复核验（全部已验证修复）

| # | r0 MUST FIX | r1 状态 | 核验依据 |
|---|------------|---------|---------|
| 1 | REQ-CPP-3 种子变更未含 owner/tenant 421 撤销与 min_verf_level 0→2 | ✅ 已修复 | REQ-CPP-3 item(3)/(4) 显式补「撤销 (1,421)/(5,421)（保留 435/436，init_permissions.sql:252-253）」「421 min_verf_level 由现有 0 提升到 2（init_permissions.sql:201-202）」，与实际种子（:201-202 置 0、:251-253 绑 (1,421)/(5,421)）逐条吻合 |
| 2 | REQ-CPB-1 数据模型未声明 is_pinned/role/publisher 保留 | ✅ 已修复 | REQ-CPB-1 显式声明 is_pinned（跑马灯载体，REQ-CPR-3 依赖）/role/publisher 三列为 RENAME 物理保留列，migration/001_initial.sql 实际含此三列 |
| 3 | 草稿编辑路径无契约，attachment_count 同步无归属 | ✅ 已修复 | 新增 REQ-CPB-9（UpdateContentPost：draft 编辑 + attachment_count 同事务重算 + scope 复校验 + is_pinned + submit 动作）+ REQ-CPB-10（DeleteContentPost）独立契约 |

另核验：r0 SHOULD FIX #4（Kafka 契约双源 → REQ-CPB-7 已改为引用 REQ-CPM-2 单源）、#5（总量上限双源 → REQ-CAS-5 已引用 REQ-CPB-6 单源）、#6（设计文档 §3.1 附件 rejected 语义 → 设计文档已同步「读路径不 mutate status，status=rejected 仅审核流回写」）均已修复。

## 发现

### 🔴 MUST FIX
（无）

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | REQ-CPB-7 ↔ REQ-CPM-1 | **Kafka at-least-once 重推机制在两个 capability 重复定义**。REQ-CPB-7（publish）与 REQ-CPM-1（moderation）各自声明「落库待推标记 + 定时重推 + 直到 ack 或 quarantine + pending-push 可观测」。本变更已为「契约载荷」（REQ-CPM-2 单源）与「总量上限」（REQ-CPB-6 单源）消除双源漂移，但重推机制本身仍双源——重推节奏/隔离阈值/metric 命名任一演进都会漂移。 | 重推机制单源落在生产者侧 REQ-CPB-7（publish capability 拥有推送动作与待推标记）；REQ-CPM-1 只声明「Kafka 侧保证 at-least-once 语义 + retention 覆盖空窗」，不重复声明扫描路径细节。 |
| 2 | REQ-CPB-10 ↔ REQ-CPB-7/REQ-CPM-1 | **pending-push 与删除/撤回的交互未定义**。REQ-CPB-10 允许 draft/submitted/approved 任一状态删除（软删 + status=withdrawn），但 REQ-CPB-7/REQ-CPM-1 的重推扫描对「记录为 pending-push 的帖」重投「直到 ack 或 quarantine」，未排除已软删/withdrawn 的帖。提交即隐式 approved 的本期路径下，submit 后删除、Kafka 不可用、重推扫描补投——可能对已撤回帖推送审核消息，产生「永不审核却推送」状态噪音。 | REQ-CPB-7 重推扫描显式排除软删/withdrawn（status=4 或 deleted_at 非空）的帖；REQ-CPB-10 删除路径显式声明清理/忽略该帖的 pending-push 标记。 |
| 3 | REQ-CPB-1/REQ-CPB-5 ↔ REQ-CPP-1 ↔ REQ-CPR-1 | **RBAC→发布角色映射被引用但无定义，role 列取值集合跨 capability 不统一**。REQ-CPB-1/REQ-CPB-5 引用「RBAC→publish-role mapping (REQ-CPP-1)」，但 REQ-CPP-1 只定义可发布角色（RBAC 名 grid_worker/community_admin/property_admin/committee），未定义到 role 列存值的映射；REQ-CPR-1 role 过滤按 NoticeRole 枚举语义（proto NOTICE_ROLE_GRID_OFFICER，场景用存值 grid_officer），与发布角色名 grid_worker 两套命名。实现者按 RBAC 名写 role 列、过滤按 NoticeRole 名查询，会静默失配（角色过滤不可用）。 | REQ-CPP-1（或 REQ-CPB-1）显式定义 RBAC→role 列存值映射（如 community_admin→community / property_admin→property / committee→committee / grid_worker→grid_officer）作为单源；REQ-CPR-1 role 过滤值集合对齐该映射。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 4 | REQ-CPB-7 场景 1 仍逐字段罗列 Kafka 载荷（post_id/section_code/text/publisher_id/version/attachments[...]），与「单源 REQ-CPM-2」声明轻微重复。建议场景仅断言「消息符合 REQ-CPM-2」，不重列字段。 |
| 5 | REQ-CAS-5（attachment-security）与 REQ-CPB-6/REQ-CPB-3（publish）均描述 file_type/file_id 从 FileInfo 回读并写入 content_post_attachments 的落库行为，落库职责双源。建议表结构定义归 REQ-CPB-3 单源，REQ-CAS-5 只引用。 |
| 6 | REQ-CPB-1 引用行号「migration/001_initial.sql:15-17」有轻微偏差（role/publisher/publisher_id 在 15-17，is_pinned 在 18），非实质性。 |

## 结构视角总体评估

- **capability 职责边界清晰**：publish（写）/ read（读）/ moderation（Kafka 基建+契约+审核）/ permission（发布权限+种子）/ attachment-security（上传安全）五能力互不重叠；单源机制（完整性谓词 REQ-CPB-8、Kafka 契约 REQ-CPM-2、总量上限 REQ-CPB-6、is_pinned 置位 REQ-CPB-9）均被下游显式引用，无重复权威。
- **proposal 影响范围 ↔ specs 服务职责边界一致**：6 服务 + docker-compose 在 5 个 spec 的职责边界表中均有归属且与 proposal 描述吻合；服务间仅 gRPC（跨服务通信经 GetUserRoles/GetFileUrl/GetResidentialAreasByDivision/AssertPublishScope，无直读他库）。
- **数据模型单源**：content_posts/content_post_scope/content_post_attachments 结构在 REQ-CPB-1/2/3 定义，read/moderation 均引用不重定义。
- **残留 SHOULD FIX 均为机制/映射级一致性问题**（重推双源、pending-push×delete 交互、role 映射未定义），不构成架构违反/数据丢失/业务不可用。

---
VERDICT: APPROVED
---
