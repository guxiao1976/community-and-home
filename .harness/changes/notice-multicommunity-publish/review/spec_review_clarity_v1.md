# Plan Review — notice-multicommunity-publish（清晰可执行视角 · rc3 独立重审）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、Scenario 具体到实现者得出相同行为、术语一致）
**审查版本**: P1.3（fallback:r0:rc3）— 与历史轮次哈希不同，按磁盘最新内容独立重审，未沿用旧轮结论
**审查对象**: proposal.md + specs/{notice-publish, notice-read, publish-permission, attachment-security, notice-moderation, notice-mobile}/spec.md
**对照基准（已逐一核实磁盘实码）**: community.proto 头注释错误码块 / CreateNoticeRequest(1-7) / GetNoticeRequest / NoticeAttachment / NoticeRole 枚举、file.proto 头错误码块 + errcode.go 70001-70003、permission-service init_permissions.sql（role 1/2/3/4/5/6/8 的 421 绑定 + §4.2 min_verf_level）+ grantSatisfiedLevel(helpers.go)、masterdata.proto GetResidentialAreasByDivision(community_div_id=3, status=4)、community-hub types.go 错误码登记 + scope.go:24(80006) + section_quota.go:13(80007)、001_initial.sql(notices/notice_attachments/published_at)、model/notice.go + createnoticelogic.go、file-service GetFileUrl rpc/api 层

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 4

## 上轮（r2:rc1，即 clarity v3）问题修复验证

| v3 # | 问题 | 状态 |
|:---:|------|:---:|
| MF-1 | community.proto 头注释漂移未登记（漏 D11 同类处置） | **已修复** — D29 已登记（proposal 决策日志 + §影响范围 api-proto 行 + REQ-NP-3 正文），并核验 D29 事实准确：types.go 登记 080003=超限/080005=参数无效/080007=板块配额；CodeOverLimit(80003) 全仓未被 emit；section_quota.go:13 确为 80007；community.proto 头「080003 寻失发布次数已达上限」确属陈旧漂移 |
| SHOULD 1 | division 展开未指定 status 参数 | **已修复** — REQ-NP-3/REQ-NP-4 显式「community_div_id=division_id AND status=1 (approved only)」，masterdata.proto:175 status「0=all,1=approved only」吻合 |
| SHOULD 2 | 附件校验载体 RPC 未钉死 | **已修复** — REQ-NP-6 钉死 GetFileUrl(file_id)（不新增 RPC）+ confirmed==true + user_id 归属；已核 RPC 层 getfileurllogic 不做 owner 过滤、返回完整 FileInfo(含 user_id 字段2)，服务间调用可行 |
| SHOULD 3 | division 选项数据源与 GetDataScopes 契约冲突 | **已修复** — REQ-NM-5 明示 GetDataScopes 现 scope_type 仅 community/building/unit/grid、无 community_div，division 选项改走 masterdata division 树，扩展 GetDataScopes 记为 design gate（D17/REV-17） |
| SHOULD 4 | sys_admin 写路径是否拦截未声明 | **已修复** — REQ-PP-1 + 场景显式「sys_admin 写路径不额外拦截（D16）」，与 REQ-PP-3 一致 |
| INFO 1 | is_system 措辞 | **已修复** — REQ-PP-1 改为「explicit seed binding for role 8」，与 init_permissions.sql `SELECT 8, id FROM sys_permission WHERE status=1` 一致 |
| INFO 2 | GetNotice community_id 字段号未入 spec | **未修复**（残留，见本轮 INFO-1） |
| INFO 3 | GetMarqueeNotices 响应结构未定义 | **未修复**（REQ-NR-3 明示设计阶段定，可接受，见本轮 INFO-2） |

上轮 MUST FIX 已闭环；本轮按磁盘最新内容独立重审，新增 2 个 MUST FIX（均未在历史轮次出现）。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|---------|------|---------|
| 1 | notice-moderation REQ-NP-MOD-4 / notice-publish REQ-NP-1 / proposal §影响范围(community-hub-service 迁移行) / 001_initial.sql:19 | **D27「published_at 审核通过时设置」与 schema `published_at DATETIME NOT NULL` 直接矛盾，迁移清单未处置**。REQ-NP-MOD-4 首场景 GIVEN「a notice created at time T0 with moderation pending and **no published_at set**」，即创建时 published_at 不写；但 001_initial.sql:19 为 `published_at DATETIME NOT NULL`，model/notice.go:52-56 插入语句含 published_at，现状 createnoticelogic.go:62 创建时写入 `PublishedAt: time.Now()`。proposal 迁移清单只列「community_id 去 NOT NULL + notice_scope + notice_attachments.file_type」，**未将 published_at 置可空/未定义创建时占位值**。字面实现 → 创建 INSERT 违反 NOT NULL → 发布主链路不可用（业务不可用）；折中实现（创建时占位 now、审核回调覆盖）又违背场景文本「no published_at set」。两种实现得出不同行为与不同 schema。 | 在 REQ-NP-MOD-4 与 proposal 迁移清单中钉死二选一并写入验收：(a) 迁移 `ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL`，创建时写 NULL、审核通过回调置 now（须说明 pending 行在 idx_published/排序中的表现）；或 (b) 显式规定创建时占位值（如 now）、审核通过回调覆盖为可见日（D27 最终语义满足），并把该占位语义写入场景，删除「no published_at set」表述。同时补 pending 期间 published_at 对 idx_published 排序影响说明。 |
| 2 | notice-publish REQ-NP-3（错误码区分段 vs「目标小区不存在」场景） | **同一输入类两处给出互斥错误码**。REQ-NP-3 加粗「Error-code distinction」段：080005 参数无效 含「**unknown-but-invalid params**」；而同需求「目标小区不存在（安全拒绝未知节点）」场景：community_ids=[99999]（md_residential_area 中不存在）→「treated as uncovered by AssertPublishScope … rejected with **080006**」。实现者读区分段→不存在目标判 080005；读场景→判 080006，前端错误处理/后端产出分叉，SHALL 无唯一解释。 | 消歧并钉死优先级：目标级解析失败（不存在/不可解析的 community_id）统一 080006（fail-closed，与 scope 未覆盖同处理，与 REQ-NP-4「展开后目标越权→080006」一致）；080005 仅保留请求形状类（空/双载、非 admin 传 division、division_id 值非法、附件引用/超限）。若「unknown-but-invalid params」指 division_id 非法值，改为显式「非法 division_id 值」并注明与 080006 的边界。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| 1 | notice-mobile REQ-NM-5 | **division 选项过滤基准「filtered to the user's county region」的 county 来源未定义**：由住宅小区→祖先 county？由 division grant(community_div scope)？由 data scope？未指定。且场景 GIVEN「publishable scope is a division grant **(permission GetDataScopes)**」与正文「GetDataScopes 不返回 community_div」直接矛盾，实现者无法确定 division 选项集与过滤依据。 | 定义 county 派生源（如 rel_user_role 中 community_div scope 的 scope_id 对应行政区，或用户住宅小区经 ResolveScopeAncestors 取 county 祖先），移除场景中误导性的「(permission GetDataScopes)」括号，与 D17/REV-17 design gate 的关系写清。 |
| 2 | attachment-security REQ-AS-5 | **载体双源残余**：REQ-AS-5 写「GetFileUrl **(or ListFiles)** 契约」，REQ-NP-6 已钉死仅 GetFileUrl；「or ListFiles」给实现者留第二条路径，与钉死结论不一致。 | REQ-AS-5 统一为 GetFileUrl（删除 or ListFiles），或说明 ListFiles 仅作备用且与 confirmed/user_id/file_type 字段一致性约束同 REQ-AS-7。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-NR-2 未在 spec 内标注 GetNoticeRequest.community_id 的字段号（2，proposal §影响范围有）；建议 spec 内自足写出（v3 INFO-2 残留）。 |
| 2 | GetMarqueeNotices 响应消息结构由设计阶段定（REQ-NR-3 已明示）——可接受，建议设计评审补（v3 INFO-3 残留）。 |
| 3 | REQ-NP-5 非发布者撤回用 080002（types.go 命名「无发布权限」）——代码复用但语义略宽（author-identity 与功能权限混用），建议 spec/前端错误文案区分「非发布者撤回」与「无发布权限」。 |
| 4 | REQ-AS-5「遗留数据 file_type 为空 MAY 记录空值」场景在 REQ-NP-6 confirmed==true 门槛下几乎不可达（新文件经 ConfirmUpload 同时写 confirmed+file_type）；建议删除或加注仅限历史数据。 |

## 问题跟踪表

| # | 状态 |
|---|------|
| 1（published_at NOT NULL vs D27） | 待修复 |
| 2（未知目标 080005 vs 080006 冲突） | 待修复 |

---
VERDICT: REVISION
---
