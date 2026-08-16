# Plan Review — notice-multicommunity-publish（业务有效性视角）

**审查维度**: 业务自洽 / 非功能（安全·性能·兼容） / 合规
**审查版本**: P1.3（fallback:r0:rc3）— 磁盘最新内容独立审查（spec 已于 8-15 夜间更新，本轮不沿用旧轮结论）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3
- VERDICT: **REVISION**

## 上轮（v3）MUST FIX / SHOULD FIX 复验（磁盘最新内容）

| # | 上轮问题 | 当前磁盘状态 | 结论 |
|---|---------|-------------|------|
| 1 | v3 CRITICAL：magic-bytes 容器识别过宽（.msi→doc / .xlsx→docx 绕过白名单） | **已修复**：REQ-AS-3 现要求 .doc 需 OLE2/CFB **且含 `WordDocument` 流**、.docx 需 ZIP **且含 `word/document.xml` 部件**；补「.msi 改名 .doc → 070004」「.xlsx 改名 .docx → 070004」验收场景；doc/doosx 内容签名显式放行 | ✅ |
| 2 | community_ids 未声明 [jstype=JS_STRING] | REQ-NP-3 已显式声明 `[jstype=JS_STRING]`（硬性约束 #3） | ✅ |
| 3 | 附件绑定校验未点名 file-service RPC | REQ-NP-6 已点名 `GetFileUrl(file_id)`→扩展 FileInfo 校验 confirmed + user_id 归属（D24） | ✅ |
| 4 | DeleteNotice 行为收窄未登记 | REQ-NP-5 已登记为行为回归（仅发布者本人，删全局 notice_scope） | ✅ |
| 5 | 080005/080006 前端错误契约未区分 | REQ-NP-3 已显式区分（080005=参数/请求形状、080006=数据权限、080002=功能权限） | ✅ |

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|-----------|------|---------|
| 1 | specs/notice-moderation/spec.md REQ-NP-MOD-4 + specs/notice-publish/spec.md REQ-NP-1（数据模型一致性） | **published_at 语义迁移（D27）与 `notices.published_at DATETIME NOT NULL` 的 schema 约束冲突，未在迁移契约中消解**。现 schema（migration/001_initial.sql:19）`published_at DATETIME NOT NULL`，本变更把 published_at 语义改为「审核通过时设置、待审时未设置」（REQ-NP-MOD-4 场景 1：`created at T0 ... no published_at set`）。但 REQ-NP-1 迁移仅声明 `ALTER TABLE notices MODIFY community_id BIGINT DEFAULT NULL`，**未把 published_at 去 NOT NULL 或定义待审期间的占位值**。若按 spec 字面实现（待审时 published_at 不设），INSERT 违反 NOT NULL 约束 → 创建主链路直接失败（业务不可用，且未迁移场景门禁只覆盖 community_id 不覆盖 published_at）。 | 在 REQ-NP-1 / REQ-NP-MOD-4 显式补充 published_at 的迁移与落库契约：① 迁移把 published_at 改为 `DATETIME DEFAULT NULL`（兼容待审不设），或 ② 定义待审期间 published_at=创建时占位、审核通过回调时覆盖为通过时刻；并在异常场景中登记「未迁移 published_at 时创建失败」门禁。二选一，spec 必须唯一解释。 |
| 2 | proposal.md 验收标准 line 114 + specs/notice-read/spec.md REQ-NR-3 场景标题 vs 场景正文（业务自洽） | **跑马灯 15 天窗口的验收标准与场景正文自相矛盾**。proposal 验收标准「审核滞留 >15 天通过的通知不入跑马灯但入浏览列表（窗口从可见日起算）」；REQ-NR-3 场景标题同样写「审核滞留 >15 天通过的通知不入跑马灯」，但**该场景正文**明确：「created 20 days ago whose approval lands today (published_at = approval = now) → the notice **IS included** in the marquee (published_at = approval day, within the 15-day window)」。即：审核滞留>15 天才通过的（创建 20 天、今天通过）→ published_at=今天 → 按 D27 语义**应入**跑马灯，与验收标准/场景标题「不入」直接冲突。验收方与实现方将得出相反结论。 | 统一表述为「审核**通过后至今** >15 天（published_at 距今 >15 天）的通知不入跑马灯但入浏览列表」；删除「审核滞留 >15 天」这种以创建-通过间隔为锚的歧义表述，正文与标题/验收标准全部以 published_at（审核通过日）为窗口锚（D27）。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|-----------|------|------|
| 1 | specs/notice-publish/spec.md REQ-NP-6 | 附件绑定校验经 GetFileUrl 逐附件调用，但未定义：同一 file_id 被多个通知引用、重复出现在 attachment_ids、附件已被删除（GetFileUrl 404）时统一归为 080005 的边界；且未提「单通知 ≤10 个/≤50MB 总量上限与 10MB 单文件硬上限在绑定侧是否再次强制」。 | 补附件重复引用/已被删除/总量超限的处置边界（统一 080005），明确总量校验用 FileInfo.file_size 而非客户端声明。 |
| 2 | specs/attachment-security/spec.md REQ-AS-3 | 两层校验第二层「回读 MinIO 实际对象 magic-bytes」依赖 file-service 能按 object_key 回读对象并解析容器（doc OLE2 流 / docx ZIP 部件）——实现成本高且 ConfirmUpload 阶段对象尚未正式落元数据，需明确回读的时序与失败（MinIO 不可达/解析异常）时的 fail-closed 语义。 | 明确第二层在 ConfirmUpload 成功后的回读时序、解析失败/IO 失败一律拒绝（fail-closed），并补一个「MinIO 回读不可用 → 070004 拒绝且不入库」场景。 |
| 3 | specs/publish-permission/spec.md REQ-PP-1 | can_publish 判定仅基于角色状态（status/verified/expires），未纳入 421 功能权限与数据范围非空：持有角色但 421 被回收（如 D26 后 property_admin）或数据范围为空（EMPTY scope）的用户，前端显示可发布入口但每次发布必失败（080002/080006）。 | 建议 GetPublishPermission 返回时叠加「持 421 且数据范围非空」判定（或明确前端以 can_publish + 数据范围选项共同决定入口），避免空转入口。 |
| 4 | specs/notice-publish/spec.md REQ-NP-3 vs REQ-NP-4 | division_id 展开后的授权判定依赖 masterdata `GetResidentialAreasByDivision`（community_div_id 展开）与 permission `AssertPublishScope` 的授权集解析（design gate REV-17）——spec 已挂 design gate，但未定义「展开返回小区数>100」与「division 展开后部分小区越权」在 division 场景下的错误码优先级（080003 vs 080006）。 | 在 division 展开场景明确校验顺序与错误码优先级（如先展开→count>100→080003；再逐目标 AssertPublishScope→越权→080006），与 community_ids 直传路径一致。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-AS-4「既有 entity_type 上传不回归」依赖「现网无 >10MB 或非白名单类型」事实：建议实现前逐 entity_type（avatar/verification/lostfound/contacts/notice）核对现有允许类型与大小，逐项验证无回归。 |
| 2 | GetResidentialAreasByDivision 现实现硬编码 page=1/pageSize=1000（getresidentialareasbydivisionlogic.go:34），division 展开小区 >100 时 spec 要求 080003 拒绝——建议设计阶段确认展开上限与 100-cap 的关系，避免超限被静默截断。 |
| 3 | notice_scope 物理删除（撤回）+ notices 软删：建议 migration 对 notice_scope 建 `uk_notice_community` 唯一约束的删除场景做幂等校验（重复撤回不报错），与 REQ-NP-5 撤回场景配套。 |

## 问题跟踪表

| 编号 | 问题 | 状态 |
|------|------|------|
| 1 | published_at NOT NULL vs D27「待审不设」无迁移/占位契约 | 待修复（MUST） |
| 2 | 跑马灯窗口验收标准与场景正文矛盾（审核滞留 vs 通过后） | 待修复（MUST） |
| 3 | 附件绑定边界（重复引用/已删/总量）未定义 | 待修复（SHOULD） |
| 4 | 第二层回读 fail-closed 语义未定义 | 待修复（SHOULD） |
| 5 | can_publish 未叠加 421/数据范围非空 | 待修复（SHOULD） |
| 6 | division 展开错误码优先级未定义 | 待修复（SHOULD） |
| 7-9 | 见 INFO | 建议采纳 |

---
VERDICT: **REVISION**（存在 ≥1 MUST FIX —— published_at 迁移契约缺失将导致创建主链路业务不可用；跑马灯窗口验收与场景正文自相矛盾，二选一票否决）
---
