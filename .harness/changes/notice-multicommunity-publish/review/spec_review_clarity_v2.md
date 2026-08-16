# Plan Review — notice-multicommunity-publish（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、Scenario 具体到实现者得出相同行为、术语一致）
**审查版本**: P1.3（fallback:r1:rc3）— 与历史轮次哈希不同，按磁盘最新内容独立重审，未沿用旧轮结论
**审查对象**: proposal.md + specs/{notice-publish, notice-read, publish-permission, attachment-security, notice-moderation, notice-mobile}/spec.md
**对照基准（已逐一核实磁盘实码）**:
- community.proto 头错误码块（行 29-33 陈旧「080003 寻失发布次数已达上限」/「080005 小区ID无效」）+ CreateNoticeRequest(1-7)/GetNoticeRequest(1)/NoticeAttachment(1-4)/NoticeRole 枚举
- community-hub types.go（80001-80007 语义登记）+ scope.go:21-24（060007→080006 映射、CodePublishScopeDenied=80006）+ section_quota.go:13（CodeSectionQuotaExceeded=80007）+ 001_initial.sql:12/19/23/24（community_id/published_at NOT NULL + idx_community/idx_published）+ createnoticelogic.go:62（PublishedAt 创建时即设）+ updatemoderationstatuslogic.go（通过回调钩子）
- permission init_permissions.sql（role 1/2/3/4/5/6/8 的 421 绑定 + §4.2 min_verf_level=0）+ helpers.go grantSatisfiedLevel（level-2=status==2 AND verified_at NOT NULL AND 未过期）+ rbac-design §6.5（grid_worker 发布通知 ✅ 拦截）
- masterdata.proto GetResidentialAreasByDivisionReq(community_div_id=3, status=4) + getresidentialareasbydivisionlogic.go（CommunityDivId>0 分支 + status=1 过滤 SubmissionStatus!=2，county_id/street_id 非必需）
- permission.proto GetDataScopesRequest scope_type 注释（community/building/unit/grid，无 community_div）+ GetUserRoles/UserRoleInfo(status/verified_at/expires_at)
- file.proto FileInfo(1-10，file_type=11/confirmed=12 空位) + file-service errcode.go（70001/70002/70003，70004/70005 空位）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 4

## 上轮（r2:rc1，即 clarity v3）问题修复验证

| v3 # | 问题 | 状态 |
|:---:|------|:---:|
| MF-1 | community/v1 头注释错误码块漂移未登记（080003/080005 双语义） | **已修复** — D29 已登记（proposal 决策日志 + §影响范围 api-proto 行 + REQ-NP-3 正文），并核验 D29 事实准确：community.proto 头行 31/33 确为陈旧注释（实码 80007=CodeSectionQuotaExceeded=寻失配额，types.go 登记 80003=CodeOverLimit=超限、全仓非测试代码未 emit）；spec 语义与 types.go/scope.go 实际登记一致 |
| SHOULD 1 | division 展开未指定 status 参数 | **已修复** — REQ-NP-3/REQ-NP-4 显式「community_div_id=division_id AND status=1 (approved only)」；已核 getresidentialareasbydivisionlogic.go:52 `in.Status==1 && c.SubmissionStatus!=2 → continue`，且 county_id/street_id 在 CommunityDivId>0 分支非必需（switch 分支） |
| SHOULD 2 | 附件校验载体 RPC 未钉死 | **已修复** — REQ-NP-6 钉死 GetFileUrl(file_id)（不新增 RPC）+ confirmed==true + user_id 归属；FileInfo 含 user_id(字段2) 供归属校验 |
| SHOULD 3 | division 选项数据源与 GetDataScopes 契约冲突 | **已修复** — REQ-NM-5 明示 GetDataScopes 现 scope_type 仅 community/building/unit/grid、无 community_div（permission.proto:243 注释吻合），division 选项改走 masterdata division 树（GetDivisionTree RPC 实存，masterdata.proto:88），扩展 GetDataScopes 记为 design gate（D17/REV-17）。**残留见本轮 SHOULD-2（county 过滤源未定义 + 场景括号残余）** |
| SHOULD 4 | sys_admin 写路径是否拦截未声明 | **已修复** — REQ-PP-1 + 场景显式「sys_admin 写路径不额外拦截（D16）」；已核 init_permissions.sql:180 `SELECT 8, id FROM sys_permission WHERE status=1`（role 8 全权限绑定属实） |

上轮 MUST FIX 已闭环；本轮按磁盘最新内容独立重审，无新增 MUST FIX，残留 4 个 SHOULD（均为精度/边界补强，不阻塞实现者得出唯一行为）。

## 发现

### 🔴 MUST FIX

无。

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| 1 | notice-publish REQ-NP-3（信任边界段） | **`publisher`（发布单位/人名称，CreateNoticeRequest 字段 5 / Notice 字段 6）信任边界未定义**。REQ-NP-3 只声明 `role`/`publisher_id` 不信任请求体（JWT/实际角色派生），对 `publisher` 显示名只字未提。字面实现→客户端可伪造显示名（如网格员伪造「社区居委会」）；安全实现→从 JWT 用户资料派生。同一 spec 两种实现得出不同行为，且本变更主题恰是「身份字段不信任请求体」，遗漏显突兀。 | 在 REQ-NP-3 信任边界段显式二选一：(a) `publisher` 也由服务端从 JWT 用户身份/名称派生（推荐，与 role/publisher_id 同标准）；或 (b) 明确声明 `publisher` 为客户端展示名、可信任并沿用现状。写入后实现无歧义。 |
| 2 | notice-mobile REQ-NM-5（division 选项 + 场景） | **division 选项「filtered to the user's county region」的 county 派生源未定义 + 场景 GIVEN「publishable scope is a division grant (permission GetDataScopes)」残余矛盾**。county 来源（division grant 的 scope_id 对应行政区？用户住宅小区经祖先链？data scope？）未指定，前端实现者无法确定选项过滤集；且场景括号 (permission GetDataScopes) 与本需求正文「GetDataScopes 不返回 community_div」直接矛盾（v1 SHOULD-1 已点，未完全清除）。 | 定义 county 派生源（如经 permission GetUserRoles 中 community_div scope 的 scope_id 取行政区，或用户住宅小区经 masterdata 祖先链）；删除场景中误导性的「(permission GetDataScopes)」括号，改为「(division grant)」或直接删括号，与 D17/REV-17 design gate 的关系写清。 |
| 3 | notice-publish REQ-NP-3（校验顺序） | **CreateNotice 多类违规同时命中时错误码优先级未声明**。080005（请求形状）/080003（目标数>100）/080006（scope 越权/目标不存在）映射各自明确，但三者同现时的求值顺序未定义（如 150 目标且含越权：count 先查→080003，scope 先查→080006，实现者得出不同码）。file 域已定主次码（REQ-AS-3 类型为主/大小次），notice-create 域无对应声明。 | 在 REQ-NP-3 显式声明校验顺序与优先级：请求形状校验（080005）→ 目标数上限（080003）→ AssertPublishScope（080006）逐层短路；或任一等价钉死的求值序，写入后前端错误映射唯一。 |
| 4 | publish-permission REQ-PP-1 | **GetPublishPermission 返回的「可发布角色列表」成员口径未定义**：是用户持有的全部可发布角色（含未达 level-2 者），还是仅 level-2 活跃者可入列？can_publish 有 level-2 条件，roles 列表成员无明确条件，实现者两种口径都成立。 | 明确 roles 列表口径：与 can_publish 同条件（仅 level-2 活跃的可发布角色入列）或列出全部持有点选角色并在响应中由 can_publish 表达活跃性，二选一写入。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | attachment-security REQ-AS-5「GetFileUrl **(or ListFiles)**」残留：REQ-NP-6 已钉死 GetFileUrl 为绑定校验载体，REQ-AS-7 亦言明字段经 GetFileUrl (and ListFiles) 返回（FileInfo 为共享消息，扩展天然双侧生效）——歧义已极低，建议统一措辞为「GetFileUrl（FileInfo 为共享消息，ListFiles 同步返回扩展字段）」即可。 |
| 2 | notice-publish REQ-NP-5 非发布者撤回复用 080002（types.go 命名「无发布权限」）：author-identity 拒绝与功能权限拒绝共码，语义略宽、前端错误文案需自行区分（v1 INFO-3 残留，可接受）。 |
| 3 | notice-read REQ-NR-3 GetMarqueeNotices 响应消息结构（如 NoticeMarqueeItem id+title）明示设计阶段定——可接受的设计交接点，设计评审须钉死（v3 INFO-3 残留）。 |
| 4 | notice-read REQ-NR-2 未在 spec 正文标注 GetNoticeRequest.community_id 字段号（2，proposal §影响范围有）——建议 spec 内自足写出（v3 INFO-2 残留）。 |

## 问题跟踪表

| # | 状态 |
|---|------|
| — | 本轮无 MUST FIX |

---
VERDICT: APPROVED
---
