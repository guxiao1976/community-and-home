# Plan Review — content-post-generalization（结构合理性视角）

**审查维度**: 职责边界 / 一致性（proposal 影响范围 ↔ specs 职责边界；各 capability 职责清晰无重叠）
**审查版本**: P1.3 fallback:r0:rc1
**审查时间**: 2026-08-16
**审查对象**: proposal.md + specs/{content-post-publish, content-post-read, content-post-moderation, content-post-permission, content-post-attachment-security}/spec.md，对照 services/ 实际设计文档与种子/migration 现状

## 摘要
- 🔴 MUST FIX: 3 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 1

## 发现

### 🔴 MUST FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/content-post-permission/spec.md` REQ-CPP-3 / REQ-CPP-1 / REQ-CPP-2 | **种子变更范围与现有种子矛盾，且与自身「SHALL NOT hold」断言不一致。** 现种子 `services/permission-service/scripts/init_permissions.sql:251-253` 显式授予 owner(1)/tenant(5) `community:notice:create-api`(421)，且 `:201-202` 将 421 的 `min_verf_level` 置为 0（注释明确「未认证业主/租户即可发布」）。REQ-CPP-1/2/3 均断言 owner/tenant「SHALL NOT hold the create permission / 只读 / can_publish=false」，但 REQ-CPP-3 的种子变更范围只列「property_admin 保留 421 + grid_worker 授 421 + 421 置 min_verf_level=2」，**未包含「撤销 (1,421)/(5,421)」与「把 421 的 min_verf_level 从现有 0 改为 2」**。按字面实现，owner/tenant 仍持 421，若其角色关联达 level-2 即可通过功能权限层发布，直接违反「业主只读」业务规则。 | REQ-CPP-3 显式补两条：① 撤销 `rel_role_permission` 中 `(1,421)/(5,421)`（删除现有 init_permissions.sql:252-253 的 421 绑定，注意保留 435/436）；② 声明 421 的 `min_verf_level` 由现有 0 提升到 2（现状见 init_permissions.sql:201-202，是行为变更非新增）。 |
| 2 | `specs/content-post-publish/spec.md` REQ-CPB-1 ↔ `specs/content-post-read/spec.md` REQ-CPR-1/REQ-CPR-3 | **数据模型保留字段集与下游特性依赖不一致。** REQ-CPB-1 声明的保留字段集仅 `title/text/published_at/publisher_id`，但 (a) REQ-CPR-3 跑马灯 `order by is_pinned desc` 依赖 `is_pinned` 列（现 `notices` 表 001_initial.sql + proto Notice 字段 8 均有）；(b) REQ-CPR-1「role 过滤保留」依赖 `role` 列；(c) 设计文档 §2.1 content_posts 含 `publisher/role` 两列。proposal/设计文档/REQ-CPB-1 三处数据模型定义均未声明 is_pinned/role/publisher，跑马灯与角色过滤的列依赖无任何 Requirement 背书，实现者按 REQ-CPB-1 枚举建模型可能遗漏，跑马灯/角色过滤不可用。 | REQ-CPB-1 显式补充保留字段：`is_pinned`（跑马灯置顶排序载体，REQ-CPR-3）、`role`（REQ-CPR-1 角色过滤）、`publisher`（展示名，设计文档 §2.1）；并同步修正设计文档 §2.1 的 content_posts schema 缺失项。 |
| 3 | `specs/content-post-publish/spec.md` REQ-CPB-4（draft 可编辑）| **草稿编辑路径（UpdateContentPost）无契约，attachment_count 同步责任无归属，D15 完整性判定骨干可能被破坏。** D9 状态机声明 draft 可编辑（含 text/attachments/scope，REQ-CPB-4），Proto 改名含 `UpdateContentPost`（proposal D4），但全 spec 无任何 Requirement 定义 update 路径行为：附件增删时 `attachment_count` 如何重算、新增附件 review_status 默认值、是否重跑 confirmed/归属/≤10/≤50MB 校验、scope 变更是否重跑 AssertPublishScope、submitted 禁编辑由哪个 RPC 承载。REQ-CPB-1/REQ-CAS-5 只在创建路径定义 attachment_count=绑定数。若草稿编辑删附件不重算 attachment_count → count(approved)!=attachment_count → 帖子永不展示；count 偏低又可能让「任一 rejected 不展示」失效。 | 在 publish capability 内新增「UpdateContentPost（draft 编辑）」显式 Requirement：定义 draft 状态下的附件增删/scope 变更契约 + `attachment_count` 重算与附件校验复跑、submitted 状态禁止编辑返回错误码；或明确限定草稿附件不可变更。使 D15 不变量（attachment_count 同步）有明确责任归属。 |

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 4 | REQ-CPB-7 ↔ REQ-CPM-2 | **Kafka content-review 契约在两个 capability 重复定义。** REQ-CPM-2 声明「single, stable payload definition」，但 REQ-CPB-7 场景 1 已完整定义同一 JSON（post_id/section_code/text/publisher_id/attachments[{file_id,file_type,review_status,file_url}]），双向漂移风险。 | 契约单源落在消费侧 REQ-CPM-2；REQ-CPB-7 改为引用 REQ-CPM-2，不重复枚举字段。 |
| 5 | REQ-CPB-6 ↔ REQ-CAS-5 | **绑定期总量上限（≤10 个/≤50MB → 080005）在两个 capability 重复声明**，职责重叠。 | 单源放在 publish 侧 REQ-CPB-6（绑定动作所有者）；REQ-CAS-5 只引用并保留 FileInfo 载体/落库职责，删除重复的绑定校验场景。 |
| 6 | 设计文档 §3.1 ↔ REQ-CPB-8 | **「任一附件 rejected → status=rejected」与 spec「不展示（status 不变，谓词隐藏）」语义不一致。** 实现者按设计文档写 status=rejected 会与 REQ-CPR 谓词式判定、REQ-CPB-8 场景（status=approved 但附件 rejected）冲突。 | spec 显式声明：附件 rejected 时 `content_posts.status` 是否置为 rejected（若置，需同步设计文档；若不置，需明确 status 保持与展示谓词的关系），消除二义。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 7 | `DeleteContentPost`（撤回）仅有场景级覆盖（REQ-CPB-4 两个场景），无独立 Requirement。建议在 publish capability 补一个显式 Requirement：仅发布者本人 + 全局生效（scope 行清理）+ 附件保留 + 错误码（080002 无权限），使删除 RPC 有明确契约落点。 |

## 问题跟踪表（状态：待修复）

| # | 问题 | 状态 |
|---|------|------|
| 1 | REQ-CPP-3 种子变更未含 owner/tenant 421 撤销与 min_verf_level 0→2 | 待修复 |
| 2 | REQ-CPB-1 数据模型未声明 is_pinned/role/publisher 保留 | 待修复 |
| 3 | 草稿编辑路径无契约，attachment_count 同步无归属 | 待修复 |
| 4 | Kafka 契约双源（REQ-CPB-7 / REQ-CPM-2） | 待修复 |
| 5 | 附件总量上限双源（REQ-CPB-6 / REQ-CAS-5） | 待修复 |
| 6 | 设计文档 §3.1 与 REQ-CPB-8 附件 rejected 语义不一致 | 待修复 |
| 7 | DeleteContentPost 无独立 Requirement | 待修复 |

---
VERDICT: REVISION
---
