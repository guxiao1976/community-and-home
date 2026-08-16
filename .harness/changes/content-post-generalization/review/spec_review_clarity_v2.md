# Plan Review — content-post-generalization（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、Scenario 具体到实现者得出相同行为、术语一致）
**审查版本**: P1.3（fallback:r1:rc1）
**审查对象**: `.harness/changes/content-post-generalization/proposal.md` + `specs/{content-post-publish, content-post-read, content-post-moderation, content-post-permission, content-post-attachment-security}/spec.md`
**对照基准**: `.change.yaml` + 磁盘现状（migration/001_initial.sql、model/notice.go、createnoticelogic.go、helper.go、community.proto、file.proto、init_permissions.sql）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 6

## 上一轮（r0，v1 报告）问题修复验证

| 上轮问题 | 状态 | 验证依据 |
|---|---|---|
| MUST FIX #1 两阶段状态机（draft/submitted 枚举、入口、触发载体缺失） | ✅ 已修复 | REQ-CPB-1 明确枚举 0=draft/1=submitted/2=approved/3=rejected/4=withdrawn；REQ-CPB-4(b) draft→submitted 由 `UpdateContentPostRequest.status=submitted` 触发、无独立 Submit RPC；REQ-CPB-5 Create 入口 status（draft 默认/submitted 立即提交）；REQ-CPB-4(c)/REQ-CPB-9(d) submit 即隐式通过置 approved+NOW；四 spec 表述一致 |
| MUST FIX #2 published_at 写时机与排序/跑马灯矛盾 | ✅ 已修复 | REQ-CPB-1 显式「本期 submit 即置 NOW()（D16 隐式通过）」；REQ-CPR-1 排序断言改为「完整性门禁要求 status=approved，本期仅 submit 隐式通过路径可达 → published_at 恒非 NULL」+ NULLS LAST 防御；REQ-CPR-3 跑马灯窗口本期非空 |
| MUST FIX #3 is_pinned 数据模型未保留且无置位机制 | ✅ 已修复 | REQ-CPB-1 显式保留 `is_pinned TINYINT DEFAULT 0`；REQ-CPB-9(f) 定义置位机制（draft 由发布者 / submitted·approved 由授权操作者）；REQ-CPR-3 场景给出可执行 GIVEN |
| SHOULD FIX #1-7、INFO #1-4 | ✅ 已修复 | 存量附件 fallback 改「保留防御路径非本期可达」；division 授权行为结论化（fail-closed 080006）；≤100 上限按展开快照计量；Redis 跳过判定钉死 `source_type="notice"`（已核代码 createnoticelogic.go:94）；role 过滤语义显式定义；docx=ZIP 措辞消歧（扩展名层 vs 内容层）；契约 version 字段落地；main_id 遗留、draft 删除语义、at-least-once 升级、契约 file_url 差异均处理 |

## 发现

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 1 | content-post-permission REQ-CPP-1 / content-post-publish REQ-CPB-1 / content-post-read REQ-CPR-1 | **角色命名与存储 role 列取值映射缺失**：REQ-CPP-1 发布角色集用 sys_role 角色名（grid_worker/community_admin/property_admin/committee，与 init_permissions.sql role_code 一致），REQ-CPB-1 要求 `role` 经「RBAC→publish-role mapping」派生但未定义该映射；现有代码（helper.go `roleToString`）与 `NoticeRole` 枚举的存储值是 grid_officer/community/property/committee，REQ-CPR-1 场景亦用「role=grid_officer」。spec 未给出 角色名→role 列值 映射，实现者可能落库 "grid_worker"，导致 REQ-CPR-1 本期交付的 role 过滤按 NoticeRole 枚举匹配时结果不一致（两实现者产出不同数据）。 | 在 REQ-CPB-1（或 REQ-CPR-1）显式给出映射表：grid_worker→"grid_officer"、community_admin→"community"、property_admin→"property"、committee→"committee"（与 NoticeRole 枚举一致），并统一 REQ-CPP/REQ-CPB 场景中的角色命名。 |
| 2 | content-post-publish REQ-CPB-1 场景「板块非法值被拒」 | **section_code 本期注册板块白名单未枚举**：仅声明 notice 为首个板块、repair 等未来板块 allowed，proposal out_of_scope 又说 repair「本期仅预留不实现」，未明确本期注册白名单与 repair 本期可否作为创建入参，实现者无法判定 section_code="repair" 是合法还是 080005。 | 显式枚举本期板块白名单（如 {notice, repair} 或仅 {notice}），并说明 repair 本期是否可创建（无板块逻辑是否仍放行）。 |
| 3 | content-post-publish REQ-CPB-9(f) / content-post-read REQ-CPR-3 | **多小区帖 is_pinned 置位作用域未定义**：(f) 要求操作者数据范围覆盖「帖子的小区」即可置顶，但多小区帖（scope C1+C2）单一 is_pinned 列同时影响所有小区跑马灯，未定义操作者 scope 需覆盖任一/全部小区，也未给出多小区置顶的 GIVEN。 | 明确多小区帖 is_pinned 的置位判定（如 scope 覆盖任一关联小区即可，或须覆盖全部），并在 REQ-CPR-3 补多小区置顶场景。 |
| 4 | content-post-publish REQ-CPB-9(e)/(f) | **is_pinned 例外未显式化**：(e) 规定「status != draft 的任何 content edit 均 080005 拒绝」，(f) 又规定 submitted/approved 帖可由授权操作者置 is_pinned；正文未声明 is_pinned 是 (e) 的例外，实现者可能误读为 approved 帖无法更新 is_pinned。 | 将 (e) 措辞限定为内容字段（title/text/attachments/scope/section_code），显式声明 is_pinned 置位按 (f) 为例外路径。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-CPB-1「draft 或 submitted-pending 状态 carries published_at NULL」与本期 submit 即置 NOW() 并存；建议补一句「本期 submit→approved 恒非 NULL，NULL 锚定规则属未来消费者阶段（D27）」，消除残留措辞歧义。 |
| 2 | REQ-CPM-1 content-review retention「survive until a consumer exists」无最小保留期（exact retention is config）；成功推送后 pending-push 标记即清除，已推送未消费消息无重推兜底，建议给最小保留基线（如 30d）作 config 默认并登记。 |
| 3 | REQ-CPB-7 未显式 Kafka 推送相对事务提交的顺序；建议明确「事务内记 pending-push 标记 → 事务提交后发送 → 成功清除标记」，避免先发后回滚产生孤儿消息。 |
| 4 | REQ-CPB-1 引用 migration/001_initial.sql:15-17 为 is_pinned/role/publisher 保留列；实际 is_pinned 在 18 行（15-17 为 role/publisher/publisher_id），建议修正行号引用。 |
| 5 | REQ-CPM-2 契约含 text 不含 title，内容审核（关键字/大模型）通常需审标题；建议明确 title 是否纳入契约或登记 BACKLOG。 |
| 6 | FileInfo 现有字段号 1-10，REQ-CAS-5 声明 file_type(11)/confirmed(12) 非破坏新增与磁盘现状一致（核 file.proto），仅作确认记录。 |

## 结论

spec P1.3（fallback:r1:rc1）已完整修复上一轮 3 项 MUST FIX 与全部 SHOULD/INFO：两阶段状态机（枚举/入口/触发载体）、published_at 隐式通过锚定、is_pinned 置位机制、Kafka 契约（version+可再生 file_url）与 at-least-once 均落到可执行粒度，跨五 spec 术语与场景一致。本期无 MUST FIX；遗留 4 项 SHOULD 属边界歧义（角色取值映射、板块白名单、多小区置顶作用域、is_pinned 例外措辞），不阻塞进入设计评审，建议并入设计/实现约束。

---
VERDICT: APPROVED
---
