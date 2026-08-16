# Design Review — notice-multicommunity-publish（接口契约+Proto 视角）v2

**审查模式**: 模式一.5 设计评审 · **视角**: interface-proto（接口契约 / Proto 破坏性 / 依赖顺序 / 鉴权-幂等-错误码）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`（v2，修订后）
**对照基准**: v1 评审问题清单 + proposal.md + 6 个 spec + 磁盘真相（community/file/permission/masterdata proto、go-zero v1.10.1 httpx/mapping 源码、types.go、handler.go、init_permissions.sql、master-data GetResidentialAreasByDivision）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 1（新增，v1 未覆盖）/ 🟡 SHOULD FIX: 1 / 🔵 INFO: 2
- 上一轮 MUST FIX 3 项：**全部已修复**（见「问题跟踪表」）

## 上一轮问题修复验证（v1 接口契约 MUST/SHOULD 逐条核对）

### v1 🔴 MUST FIX（3 项，全部修复 ✅）
| # | v1 问题 | 修复落地证据 | 状态 |
|---|---------|-------------|:---:|
| MUST 1 | `CommunityIds []int64 json:"community_ids"` 无法解析 TS string 形式 Snowflake ID | tasks.md Task 1.15 已改 `[]string` + 逐个 `strconv.ParseInt`；design.md §CreateNotice「REST/JSON 绑定（评审 MUST 1/2 修订）」显式说明 `,string` 不支持 slice。已核实 proto 侧 `community_ids=8 repeated int64 [jstype=JS_STRING]`、移动端 Task 4.1 `community_ids: string[]`，链路闭环。 | ✅ 已修复 |
| MUST 2 | API 层缺 `AttachmentIds` 字段与透传（附件功能静默失效） | tasks.md Task 1.15 新增 `AttachmentIds []string json:"attachment_ids"` + 「透传到 RPC attachment_ids（修复现丢弃）」；design.md 同声明。已核实 proto `CreateNoticeRequest.attachment_ids(7)` 存在、现有 types.go CreateNoticeReq 确无该字段（30-35 行）、现有 api create 逻辑丢弃——修复方向正确。 | ✅ 已修复 |
| MUST 3 | `/api/perm/data-scopes` 鉴权未钉死且「复用 422」排除发布者 | tasks.md Task 3.2 新增专用权限码 `425 GET:/api/perm/data-scopes`（community:data-scopes:read-api）+ 绑定全量移动端角色；Task 3.3 去掉「或经既有 read 权限放行」二义措辞 + 钉死 PermMiddleware 归属 permission-service。已核实 init_permissions.sql：422 仅绑 (9,1,5)，grid_worker(4)/community_admin(3)/committee(6) 确不在内——425 专用码正确，不复用 422 的结论成立。 | ✅ 已修复 |

### v1 🟡 SHOULD FIX（4 项，全部落地 ✅）
- SHOULD 4（role/publisher_id deprecated）→ Task 0.1 已标 deprecated（`publisher`(5) 不 deprecated，评审 S1 一致）。
- SHOULD 5（division<=0 guard）→ Task 1.6 `ExpandDivisionCommunities` 先 guard `divisionID<=0→080005`。已核实 masterdata `getresidentialareasbydivisionlogic.go:39-40` default 分支确走 `FindAll`（最多 1000 小区）——guard 必要且落地正确。
- SHOULD 6（CHANGELOG 语义破坏登记）→ Task 0.4 已显式登记 GetNotice 必填 community_id + legacy CreateNotice 不再接受 community_id(1)，含兼容期与回退行为。
- SHOULD 7（重审 published_at 副作用）→ design.md ADR 已记录「重审重新锚定可见日，属可感知行为副作用（评审 SHOULD 7 显式声明）」。
- （data-model v1）M1 通过态谓词 IN(1,3) → Task 1.4 共享谓词 `isModerationPassed` + 读查询/回调复用，已核实。

## 已核实（设计正确、与磁盘真相一致的部分）

- Proto 全部为**兼容新增**：`CreateNoticeRequest.community_ids(8)/division_id(9)`、`GetNoticeRequest.community_id(2)`、`NoticeAttachment.file_type(5)`、`FileInfo.file_type(11)/confirmed(12)` 均落在空闲字段号（已读 community.proto / file.proto 核对）；新 RPC `GetPublishPermission`/`GetMarqueeNotices` 非破坏；`buf breaking-check` 可通过。
- 错误码消歧判定准确：file 侧实际常量 70001/70002/70003（FileNotFound/AccessDenied/OperationFailed）与 design D11「070004 文件类型不支持 / 070005 文件大小超限」目标一致；community 侧实际常量 80001-80007（types.go 18-23 行）——D29「剔除陈旧 080003 寻失发布次数已达上限（实际码 080007）」经 `section_quota.go:13`（CodeSectionQuotaExceeded=80007）+ `createlostfoundlogic.go:55-56`（寻失发布配额映射 80007）核实成立；080003 为通用「超限（复用）」与「单次发布目标数超限」新语义兼容。
- 批量 scope 校验契约成立：`AssertPublishScopeRequest.repeated targets`（permission.proto:257-265）——Task 1.6 单次批量调用（非 N+1）设计正确。
- level-2 发布判定基于 RPC 输出成立：`GetUserRolesResponse.UserRoleInfo` 含 `role.code/status/verified_at/expires_at`（permission.proto:347-369），无需直读 rel_user_role。
- 附件校验载体契约自洽：`FileInfo.user_id(2)` 已存在并经 `GetFileUrl` 返回，D24「confirmed + user_id 归属 + file_type 回读」无需新 RPC。
- division 展开契约：masterdata `GetResidentialAreasByDivision`（CommunityDivId>0 + status==1 → SubmissionStatus!=2 过滤）与 design 一致；default 分支 FindAll 的过度展开风险已由 division<=0 guard 封堵。
- 权限种子核对：421 现绑 (2,3,6,1,5)、grid_worker(4) 未持、422 绑 (9,1,5)——「授 4 / 回收 2/1/5」+「425 全角色」目标与现状一致。
- web/pc 无 notice 消费方（grep 无命中）、web/mobile 在本变更内同步升级——GetNotice 必填 community_id 的语义破坏面收敛，CHANGELOG 已登记。
- 依赖顺序合规：Proto(0) → 迁移/模型(1.1-1.5) → scope 基础设施(1.6) → 写/读逻辑(1.7-1.13) → 注册/API(1.14-1.17) → 前端(4) → 运维验证(5)；基础设施→核心→辅助→前端。
- REST 静态路径 `GET /api/community/notices/marquee` / `publish-permission` 先于 `:id` 注册（Task 1.17）+ PermMiddleware 组内注册，路由抢占风险已显式处理。

## 发现

### 🔴 MUST FIX（1）

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | tasks.md Task 1.15（`GetNoticeRequest 增 CommunityId int64 json:"community_id,string"`）+ design.md §GetNotice | **GetNotice 的 community_id 在 REST 层无法绑定 → 详情页 080005 全挂**。GetNotice 是 GET `/api/community/notices/:id`，移动端（Task 4.5「请求携带当前小区 community_id 上下文」）必须以 query 传 community_id；但 Task 1.15 给 GetNoticeReq 标了 `json` 标签。已读 go-zero v1.10.1 源码确认：`httpx.Parse` 对 GET 走 `ParseForm`（formUnmarshaler，key="form"），`unmarshaler.go:571 usingDifferentKeys("form", field)` 对无 `form` 标签的字段（`field.Tag.Lookup("form")` 失败）**直接 skip**，`json` 标签在 GET query 绑定上完全不生效。结果 `CommunityId` 恒为 0 → api 逻辑透传 0 → RPC GetNoticeLogic（Task 1.8「缺失/空→080005」）→ 每次详情调用 080005，通知详情页整体不可用。本仓库 GET query 绑定的既有惯例即 `form` 标签（ListNoticesReq `form:"community_id"`、Task 1.16 marquee「解析 community_id query」），Task 1.15 的 `json` 标签与之矛盾。 | Task 1.15 GetNoticeReq 改为 `CommunityId int64 form:"community_id"`（可选则 `form:"community_id,optional"`，api 逻辑显式判空→080005）；同步检查 Task 1.16 marquee handler 类型：`GetMarqueeNoticesReq.CommunityId` 必须同样用 `form:"community_id"`（不要用 json/path）；补 api 层 RED 测试：GET `/notices/:id?community_id=456` 断言 req.CommunityId 被绑定，缺失→080005。 |

### 🟡 SHOULD FIX（1）

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 2 | design.md §Design Gate + tasks.md Task 3.1 | **AssertPublishScope 判据放宽的 blast radius 未acknowledge**：`AssertPublishScope` 是共享 RPC，除通知创建（新 AssertCommunitiesScope）外还被 lostfound 创建（createlostfoundlogic.go）、contacts upsert（upsertcontactslogic.go）调用。`resolvePublishScope` 收集 `community ∪ community_div` 后，community_admin 的 division grant 将**同时放行** lostfound/contacts 写到 division 下小区——这大概率是修复（division 授权本就该生效），但 design/task 只按「通知发布」框定该变更，Task 3.1 的 RED/验证门禁也只覆盖通知场景（C1/C2/目标不存在/grid_worker 不回归），未对 lostfound/contacts 调用方补 division 授权回归。属跨服务共享语义变更，按「跨服务变更加倍谨慎」须显式声明。 | Task 3.1 在 Design Gate 验证门禁追加 2 条回归：community_admin 持 `community_div=D1` 时（a）lostfound 发布到 C1（祖先链含 D1）→ allowed、（b）contacts upsert 到 C1 → allowed（或按产品语义拒绝则显式声明并在 resolvePublishScope 内按调用方隔离）；design §Design Gate 注明该变更对 lostfound/contacts 写入路径的连带生效。 |

### 🔵 INFO（2）

| # | 建议 |
|---|------|
| 3 | web/mobile/src/api/community.ts 当前**没有** `createNotice` 函数（仅 getNoticeList/getNoticeDetail/getContacts/getLostFoundList，已核实）。Task 4.1 措辞「CreateNoticeRequest 参数增 community_ids/division_id」隐含「扩展既有函数」，但并无既有函数可扩；发布页（Task 4.6）依赖的 createNotice client 必须**新增**。建议 Task 4.1 改为「新增 createNotice(…) client（含 community_ids/division_id/attachment_ids 参数）」，避免执行方搜索不到既有函数而困惑。 |
| 4 | Task 2.5 将 Migration 002 与 File 模型扩展合并在同一任务（数据层内部耦合），而社区侧 Task 1.1 的 Migration 003 是独立任务——两侧拆分粒度不一致。任务粒度规则「Migration 变更必须独立成任务」严格解读下 Task 2.5 属灰色。若团队严格执法应拆出独立任务；若保留耦合，建议在 Task 2.5 注明理由（迁移即模型列的 DB 载体，Task 5.1 独立运维验证已兜底）。 |

## 问题跟踪表

| 状态 | 说明 |
|------|------|
| 待修复（本轮） | MUST 1（GetNotice community_id 绑定） + SHOULD 2（AssertPublishScope 共享 blast radius） |
| 已修复（上轮 MUST 1/2/3 + SHOULD 4/5/6/7，已验证） | 逐条见「上一轮问题修复验证」 |
| 已修复（data-model v1 M1，评审共享） | 通过态谓词 IN(1,3) + isModerationPassed 单一谓词 |

---
VERDICT: REVISION（存在 1 个 MUST FIX）
---

## 结论
v2 设计质量显著收敛：v1 三个接口契约 MUST FIX（CommunityIds []string、AttachmentIds 透传、data-scopes 425 钉死）均已正确落地并经磁盘证据核实，错误码消歧（D11/D29）、批量 AssertPublishScope、GetUserRoles level-2 判定、division 展开 guard、CHANGELOG 语义破坏登记等设计判断全部与代码真相一致。仅存一个新增 MUST FIX：Task 1.15 给 GetNoticeReq 的 community_id 标 `json` 标签，而 GET query 绑定在 go-zero 下只认 `form` 标签（已读源码确认 skip 行为），执行方照 Task 字面实现将导致通知详情页每次调用 080005 整体不可用——需改为 `form` 标签并补 api 层绑定测试后回传架构设计师修订。
