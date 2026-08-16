# Plan Review — content-post-generalization（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、Scenario 具体到实现者得出相同行为、术语一致）
**审查版本**: P1.3（fallback:r0:rc1）
**审查对象**: `.harness/changes/content-post-generalization/proposal.md` + `specs/{content-post-publish, content-post-read, content-post-moderation, content-post-permission, content-post-attachment-security}/spec.md`
**对照基准**: `docs/superpowers/specs/2026-08-16-content-post-design.md` + `.change.yaml` + 磁盘现状（migration/001_initial.sql、notice 模型与 create 逻辑、file.proto、community.proto）

## 摘要
- 🔴 MUST FIX: 3 / 🟡 SHOULD FIX: 7 / 🔵 INFO: 4

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | content-post-publish/spec.md REQ-CPB-1 / REQ-CPB-4 / REQ-CPB-5 / REQ-CPB-7 | **两阶段状态机（draft→submitted）与 status 枚举/入口/触发载体冲突，实现者无法得出相同行为**。REQ-CPB-1 定义 `status` 枚举仅 0=pending/1=approved/2=rejected/3=withdrawn，且场景「迁移后多小区发布写入成功」/ REQ-CPB-5 场景「网格员多小区发布成功」均写「插入即 status=approved」；而 REQ-CPB-4 要求实现 draft（可编辑）→ submitted（不可编辑可删）两阶段，REQ-CPB-7 的 Kafka 推送依赖「状态转换到 submitted」触发。spec 未定义：①draft/submitted 对应 status 的哪个枚举值；②CreateContentPostRequest 契约（REQ-CPB-5）无 status 字段、RPC 集合（Create/List/Get/Delete/Update）中无 submit 端点，draft 的创建入口与 draft→submitted 的触发载体（字段？RPC？）完全缺失；③「插入即 approved」与「draft 草稿可编辑」场景直接矛盾（谁创建 draft？）。REQ-CPB-4 场景「draft 状态重复提交产生重复帖」还暗示 Create=create+submit 一体，与 draft 场景再次冲突。 | 在 REQ-CPB-4/REQ-CPB-5 中显式定义：①CreateContentPost 的 entry 状态（draft 还是直接 approved）与请求是否携带 status/action 字段；②draft→submitted 的触发载体（如 UpdateContentPost.status 字段或独立 Submit RPC，须与 api-proto 契约一致）；③draft/submitted 在 `content_posts.status` 的映射值（如 draft=0 pending、submitted=0 pending 或新增枚举值），并统一 REQ-CPB-1 场景「插入即 approved」与两阶段状态机的表述；REQ-CPB-7 的推送时机随触发载体定义一并落定。 |
| 2 | content-post-read/spec.md REQ-CPR-1 / REQ-CPR-3 + content-post-publish/spec.md REQ-CPB-1 / REQ-CPB-8 | **published_at 写时机与「默认 approved 且无消费者」冲突，排序/跑马灯行为矛盾、场景不成立**。REQ-CPB-1 明确 published_at「创建 NULL、通过置 now（REUSE:notice-D27/D30）」，而本期无审核消费者（D16）status 默认 approved → 新帖 published_at 恒为 NULL。此时：①REQ-CPR-1 声称「rows failing the completeness gate (which may carry NULL published_at) SHALL be excluded before ordering so NULL never perturbs DESC」——但完整性谓词（REQ-CPB-8）只含 status=approved + count(approved attachments)==attachment_count，**不含 published_at 条件**，默认 approved 的新帖过门禁仍带 NULL published_at，会进排序并扰动 DESC，该断言在逻辑上不成立；②REQ-CPR-3 跑马灯窗口 `published_at >= now-15d` 在 published_at 恒 NULL 时恒空，「跑马灯返回 ≤10 条」场景在本期无法复现。 | 明确本期默认 approved 场景下 published_at 的置值规则，并同步三处：①若默认 approved 即视为审核通过，则在 submitted/创建时置 `published_at=now`；②若保持 NULL，须显式定义排序对 NULL 的处理（NULL 排最后、或排序前过滤 NULL）与跑马灯窗口对 NULL 的处理，并修正 REQ-CPR-1「NULL 不扰动 DESC」的论证（补 published_at 条件或改排序列）。三处（REQ-CPB-1 场景、REQ-CPR-1、REQ-CPR-3）保持一致，避免实现者歧义。 |
| 3 | content-post-read/spec.md REQ-CPR-3 + content-post-publish/spec.md REQ-CPB-1 | **跑马灯依赖 is_pinned，但数据模型未显式保留 is_pinned 且无置位机制，置顶场景无法复现**。REQ-CPR-3 要求 `order by is_pinned desc`（置顶优先），但 REQ-CPB-1 的「保留字段集」仅列 title/text/published_at/publisher_id + 新增 section_code/status/attachment_count，未列 is_pinned；proposal 验收标准与 design doc §2.1 亦未列 is_pinned。虽然 RENAME 会物理保留该列（migration 001 有 `is_pinned TINYINT DEFAULT 0`），但 spec 未声明其保留，且**全 spec 无任何需求定义 is_pinned 如何被置 true**（哪个 RPC/流程置顶？），实现者无法让「置顶优先」生效。 | 在 REQ-CPB-1 数据模型中显式声明保留 `is_pinned`（含默认值 0）；新增/明确 is_pinned 的置位机制（如 UpdateContentPost 支持 is_pinned 字段，或明确本期置顶来源沿用 notice 既有能力），并在 REQ-CPR-3 场景中给置顶帖给出可执行的 GIVEN（置顶由谁操作）。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 1 | content-post-publish REQ-CPB-1 vs REQ-CPB-3 / content-post-read REQ-CPR-2 | **存量附件读 fallback 与存量帖不可读矛盾**：REQ-CPB-1 场景「存量 notices 行不迁移」明确存量帖因无 content_post_scope 行而**不返回**；但 REQ-CPB-3 场景「存量附件行缺 review_status/file_id」与 REQ-CPR-2 场景「存量附件 file_id=0 回退 stored file_url」都描述「读到存量附件行」的读路径行为。存量帖不可读则存量附件永不可达，这些 fallback 场景不可达、相互矛盾。 | 统一口径：要么明确存量帖本期不可读、删除存量附件 fallback 场景；要么说明存量附件 fallback 针对「新帖中历史遗留附件」或预留路径，并给出一条可达的 GIVEN。 |
| 2 | content-post-permission REQ-CPP-2 | **division→community 授权集解析为设计门禁未定，场景假设超前**：REQ-CPP-2 场景「社区管理员选社区展开为小区快照」假设 division 授权可解析出 C1/C2 作为 publishable set，但正文明确该判定「SHALL be verified ... in the design review before coding（design gate）」——即当前未知能否成立，场景在 spec 阶段不可执行、且未定义失败时的错误行为。 | 要么把该判定从「待 design 验证」提升为 spec 明确结论（division 授权是否进入 AssertPublishScope 授权集），要么把场景标为待定并在场景中给出验证失败时的行为（错误码/拒绝语义），供实现者不阻塞。 |
| 3 | content-post-publish REQ-CPB-5 | **「≤100 targets」与 division_id 展开歧义**：REQ-CPB-5 说 community_ids/division_id「≤100 targets」，但 division_id 经 GetResidentialAreasByDivision 展开为多个小区，展开后的数量是否受 100 上限约束？上限按 community_ids 元素数还是展开后小区数？080003 超限按哪个口径触发？ | 明确 100 上限的计量口径（原始入参 vs 展开后快照数），并给出 division 展开超限时的行为与错误码。 |
| 4 | content-post-moderation REQ-CPM-4 | **Redis 消费者「跳过 content_post 任务」的精确判定标签未定义**：REQ-CPM-4 场景要求按「source_type/content_type indicating content post」跳过，但未定义精确值。现有代码 task 消息 source_type 为 "notice"（createnoticelogic.go:81/94），本期 content_posts 不再入 Redis，存量/残留任务携带的标签值到底是 "notice" 还是新值 "content_post" 未说明，实现者无法写出精确的 skip 判定。 | 明确 Redis 残留/存量 content-post 任务的 source_type 精确取值（及与 lostfound/user 的区分），并落到 REQ-CPM-4 场景的 GIVEN 中。 |
| 5 | content-post-read REQ-CPR-1 | **「optional role filter SHALL be retained with notice-compatible semantics」行为未定义**：只引用「notice 兼容语义」未在本 spec 给出 role filter 的输入/判定/对返回集的影响，实现者无法判断该 filter 过滤什么、来自哪个权限 RPC。 | 在本 spec 显式给出 role filter 的契约（入参、判定依据、对列表结果的作用）或引用到具体权限 RPC/语义定义。 |
| 6 | content-post-attachment-security REQ-CAS-1 / REQ-CAS-3 | **「reject all archive files (zip, rar) without exception」与 docx=ZIP 容器存在措辞张力**：REQ-CAS-1 白名单含 docx、禁止集含「所有 zip/rar」；而 docx 本质是 ZIP(OOML)。场景虽已消歧（扩展名层拒绝 .zip/.rar，ConfirmUpload 层 docx 特判），但正文 SHALL 表述「zip/rar 一律不接收」会让实现者误伤 docx 判定。 | 将正文 SHALL 措辞改为「扩展名层拒绝 .zip/.rar 扩展名；内容层仅 docx(OOML: ZIP+word/document.xml) 特判放行，其余 zip/rar 内容拒绝」，与场景一致。 |
| 7 | content-post-moderation REQ-CPM-2 | **契约版本字段未定义**：REQ-CPM-2 声明「契约变更需 version bump」，但消息契约（post_id/section_code/text/publisher_id/attachments[]）无 version 字段，bump 载体缺失。 | 在消息契约顶层增加 `version` 字段（如 int32）或明确 bump 的落地方式（topic 命名版本化？），使契约可演进。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-CPB-3 正文沿用「main_id」（proposal 与 design doc 的遗留说法），正文已写「post_id ... i.e. the main_id referenced by the user」，建议全链路统一为 post_id，删除 main_id 遗留术语，避免实现者误以为有两个关联字段。 |
| 2 | REQ-CPB-4 两阶段场景未显式说明 draft 状态可否删除（submitted 明确可删），建议补一句 draft 删除语义，消除歧义。 |
| 3 | REQ-CPB-7/REQ-CPM-1「Kafka 推送失败 MAY be retried by compensation/reconciliation path」的对账/补偿机制未定义（MAY 属允许性，本期不阻塞）；因本期 status 默认 approved、无消费者，消息丢失无可见影响，建议登记 BACKLOG。 |
| 4 | Kafka 消息契约（REQ-CPB-7/REQ-CPM-2 含 file_url）与 design doc §4.2（不含 file_url、attachments 无 review_status）存在差异；spec 为权威、符合 D7，建议同步 design doc 避免后续实现者以 design doc 为准。 |

---
VERDICT: REVISION
---
