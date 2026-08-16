# Design Review — content-post-generalization（interface-proto 视角 v3）

**审查维度**: 接口契约 + Proto 破坏性 + 依赖顺序
**审查对象**: `.harness/changes/content-post-generalization/design.md` + `tasks.md`
**对照**: proposal.md + 5 个 spec + `api-proto/api/community/v1/community.proto` + `api-proto/api/file/v1/file.proto` + `api-proto/api/permission/v1/permission.proto` + `api-proto/api/masterdata/v1/masterdata.proto` + `services/community-hub-service/rpc/internal/logic/notice/updatenoticelogic.go` + `services/community-hub-service/rpc/internal/logic/notice/getnoticelogic.go` + `services/community-hub-service/api/internal/types/types.go` + `services/permission-service/model/rel.go` + `services/permission-service/rpc/internal/logic/permission/scope.go` + `services/community-hub-service/rpc/internal/svc/servicecontext.go` + `web/mobile/src/pages.json` + `web/mobile/src/api/community.ts` + `web/mobile/src/pages/notice/*.vue` + `services/community-hub-service/rpc/internal/logic/lostfound/*.go`

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2

## 上一轮问题修复验证（v1 → v2 → v3）

| 轮次 | 问题 | 验证状态 |
|------|------|---------|
| v1 M1 | ContentPost 字段号冲突 | ✅ 已修复。实核对 community.proto：ContentPost 既有 1-12（content=4/role=5/publisher=6/created_at=10/updated_at=11）无重复，新增 section_code(13)/status(14)/attachment_count(15) 为未占用号；ContentPostAttachment 1-4 + 新 5/6/7 无冲突；FileInfo 1-10 + 新 11/12 无冲突 |
| v1 M2 | REST/Proto 枚举错位（透传） | ✅ 已修复。entry_status/Update.status 改 int32 三侧同号（0=draft/1=submitted），删除 ContentPostEntryStatus 枚举；Task 0.1/1.22/1.23 一致 |
| v1 M3 | D3 停 Redis 未覆盖 Update 路径 | ✅ 已修复。已核实 updatenoticelogic.go CreateAuditLog(~L72)+LpushCtx(~L97) 块存在，Task 1.11 整体移除 + Task 1.20 RED 断言扩双路径 |
| v2 M1 | 破坏面误评估（web/mobile 有消费方）+ wire 改名破坏 | ⚠️ 主体已修复，**遗留子问题见 MUST 2**。已实测：pages.json L10/34/66 注册 notice/notice-detail/notice-browse 页（L91-94 tabbar）；api/community.ts:119-136 读 res.notices/res.notice（getNoticeDetail 不传 community_id）；notice.vue:336/notice-browse.vue:110/notice-detail.vue:103 实际调用。R2 方案 a（wire 键保持 + /notices 路径）已落地且 wire 键逐一核对（notices/notice/content 保持） |
| v2 S1 | 响应 base=1 未显式声明 | ✅ 已修复。Task 0.1 各响应消息显式列 base=1，与既有响应（CreateNoticeResponse 等 base=1）一致 |
| v2 S2 | 头注释对齐遗漏 080004 | ⚠️ 保留动作正确，**标签错误见 SHOULD 1** |
| v2 S3 | publisher 档案查询缺 UserClient | ✅ 已修复。已核实 servicecontext.go 仅 Moderation/Permission/MasterData/SysConfig/Redis，无 UserClient；Task 1.9 补 UserRpc 接线（user-service GetUsersByIds 已存在） |
| v1 S4 | property_admin 421↔427/428 不对称 | ✅ 已登记 design §权限种子 + Task 3.2 + rbac-design.md §6.5 |

## 发现

### 🔴 MUST FIX

| # | 位置 | 问题 | 修复建议 |
|---|------|------|---------|
| 1 | `design.md` §UpdateContentPost（L184-193）+ `tasks.md` Task 1.11（L166-173） | **UpdateContentPost 授权模型自相矛盾，REQ-CPB-9(f) 能力不可达。** Task 1.11 顶层先做**作者校验**「`publisher_id == JWT user_id`（非发布者编辑 → 080002）」，而 REQ-CPB-9(f)（spec content-post-publish/spec.md:208,235-238 + 场景「授权操作者置顶 approved 帖」）要求**非发布者、持发布角色且数据范围覆盖帖小区的授权操作者**可对 submitted/approved 帖置 is_pinned。两者直接冲突：非发布者操作者在作者校验处即被 080002 拒绝，永远到不了 is_pinned 分支——跑马灯置顶能力（REQ-CPR-3 依赖）对非发布者操作者不可实现；且「非 draft 不可编辑 → 080005」门禁未区分「仅 is_pinned」与「内容编辑」，若照字面实现，approved 帖 is_pinned-only 更新也会被 080005 拒掉。Task 1.11 的 RED 测试清单只有「is_pinned 置位授权/越权」，无「非发布者操作者置顶 approved 帖」用例，实现者无从落实分支。 | 在 design.md §UpdateContentPost + Task 1.11 显式定义授权分支：(a) 含内容/附件/scope 字段的更新 → 仅 draft + 发布者（作者校验 080002）；(b) 仅 is_pinned 且帖 submitted/approved → 跳过作者校验，改验 `PublishRolesFrom` 非空 + `AssertCommunitiesScope`（数据范围覆盖帖小区），越权 → 080006；(c) 非 draft 的内容编辑 → 080005。并在 Task 1.11 RED 增用例：非发布者操作者（持发布角色 + scope 覆盖）置顶 approved 帖成功 / scope 不覆盖 → 080006。 |
| 2 | `design.md` §GetContentPost「REST 兼容回退」（L215）+ `tasks.md` Task 1.14/1.23（L195/277） | **多小区用户移动端详情页回归：R2「移动端现行通知消费方在迁移后仍可用」未完全兑现。** `ResolveSingleCommunityForCompat` 要求「恰一个 distinct 小区 → 使用；多/无 → 080005」。现网存在真实多小区用户：grid_worker 是多小区角色（apply_role_logic 每服务小区一条 community grant）、多房产业主可持多条 community grant——这类用户点击 notice-detail（notice-detail.vue:103 `getNoticeDetail(id)`，仅传 id，无 community_id）在迁移后**恒 080005**，详情页不可用。而当前行为（getnoticelogic.go 反查帖 community_id → `FilterAllowed(userID, community_id)`）对任意小区数用户均可读（LIMITED 语义）。设计选型比现状**更窄**，且未在 BACKLOG / Task 6.2 回归清单登记该收窄。 | 二选一，建议前者：**(a) 回退改反查**——REST 薄代理层缺 community_id 时按 `content_post_scope` 反查该帖所属小区集，`FilterAllowed` 逐一判定任一允许即放行（与现网行为一致、不泄露，无小区用户维持 080001/080005），而非依赖用户 grant 恰好唯一；**(b)** 若坚持 grant 派生，须在 design.md + BACKLOG 显式登记「多小区用户（含 grid_worker）迁移后 notice-detail 缺参 080005」为已知回归，并在 Task 6.2 补一条多小区用户详情冒烟。 |

### 🟡 SHOULD FIX

| # | 位置 | 问题 | 建议 |
|---|------|------|------|
| 1 | `tasks.md` Task 0.1（L29）头注释对齐 | **080004 保留动作正确，但标签「便民联络不存在，ContactService 仍用」与代码不符。** 已核实 080004 = `CodeLostFoundMiss`（types.go:19，寻失记录不存在），唯一使用方是 lostfound（resolvelostfoundlogic.go:33 / updatemoderationstatuslogic.go:39 / getlostfoundlogic.go:34 均返回 80004「寻失记录不存在」）；contact 逻辑 grep 无任何 080004 引用。Task 0.1 若照「便民联络不存在」落注释，将把社区.proto 现存漂移继续固化（与「头注释对齐实际语义」目标自相矛盾）。 | 头注释对齐为 `080004 寻失记录不存在（LostFoundService 仍用）`，保留理由改引 LostFoundService 而非 ContactService；080004 行本身保留不动。 |
| 2 | `design.md` §权限种子（L356）+ Task 3.2（L379） | **「全部移动端角色」集合内 property_admin(2) 已因 platforms='pc' 被排除，但 421 授权保留，形成移动端「可发布不可编辑/撤回」缺口仍只靠 080002 作者校验兜底**（v1 SHOULD 已登记，v3 维持确认——创建即发布者，作者校验可兜底创建后操作，不阻塞；但请在 Task 6.2 权限冒烟显式补一条「property_admin 身份调用 427/428 → 403（fail-closed）」，把该不对称纳入回归断言，防后续放开 427/428 到 property_admin 时静默变更）。 | 非阻塞登记强化：Task 6.2 权限矩阵冒烟显式断言 property_admin 对 427/428 为 403（移动端面），形成 fail-closed 回归护栏。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | **详情拒绝错误码变化**：现网 GetNotice scope 外拒绝返回 `DenyBase`（=080006 数据权限拒绝，scope.go:66），新 GetContentPost 改映射 080001。移动端 fetchDetail 仅置 notice=null 不分支，无运行期影响；建议 design §GetContentPost 注释一句「scope 外/不存在统一 080001（含原 080006 拒绝路径），避免与写路径 080006 混淆」。 |
| 2 | **`CreateContentPostReq.CommunityIds []string` 逐条 `strconv.ParseInt`**：Task 1.23 已含转换，但未定义非数字串的失败语义（转 int64 失败应 080005 还是忽略）。建议在 Task 1.23 RED 补一条「community_ids 含非数字 → 080005」用例，防实现时静默忽略产生空范围误过校验。 |

## 已核实无问题（架构一致性 / 契约核对）

- ✅ **字段号全量唯一**：ContentPost 1-15、ContentPostAttachment 1-7、FileInfo 1-12、各新请求/响应消息字段号逐一与 community.proto/file.proto 现有定义核对无冲突。
- ✅ **响应 base=1 全对齐**：既有 ListNoticesResponse/GetNoticeResponse 等 base=1，新契约各响应显式 base=1。
- ✅ **UpdateModerationStatusRequest/Response 保留正确**：community.proto:229 LostFoundService.UpdateLostFoundModerationStatus 仍用，移除 UpdateNoticeModerationStatus RPC 后消息不删。
- ✅ **moderation-service 为唯一外部 gRPC 消费方**：全仓 NoticeServiceClient 仅 moderation-service（servicecontext.go:38,147-151 / task_handler.go:38,44）+ community-hub 自身 API 代理层（同步改名）；Task 4.1 移除接线成立。
- ✅ **web/pc 无 notice 消费方**：grep 为空，破坏面断言属实。
- ✅ **R2 wire 兼容键逐一成立**：ListContentPostsResp.notices / GetContentPostResp.notice / 帖体 content / community_id / attachments 各键与移动端 Notice 接口及 notice.vue/notice-browse.vue/notice-detail.vue 读取字段一致；新增键（section_code/status/attachment_count/file_type/file_id/review_status）additive 无冲突。
- ✅ **entry_status/Update.status 三侧同号**：REST 0/1 ↔ proto 0/1 ↔ DB（draft 入口 0 / submitted 入口 2 隐式通过），从根因消除枚举错位。
- ✅ **UserRoleInfo 支撑**：permission.proto GetUserRoles 返回 Role+scope_type(2)/scope_id(3)/status(4)/verified_at(5)/expires_at(6)——level-2 判定与 ResolveAdminDivision（community_admin + scope_type='community' + scope_id!=0）均 grounded。
- ✅ **R1 grounded 实核**：rel.go:77-82 scope_type 常量无 community_div；apply_role_logic 将 community_admin 绑 scope_type='community'/scope_id=communityId；permission scope.go 已有 resolveUserScope/grantActive 供 Task 3.1 扩展；masterdata GetResidentialAreasByDivisionReq 含 community_div_id(3)+status(4,0=all/1=approved) 支撑「division + status=1 展开」。
- ✅ **共享 blast radius 声明**：AssertPublishScope 被 lostfound/contacts 复用，Task 3.1 门禁场景 5 含共享调用方回归（lostfound C1 allowed / contacts C1 allowed / owner 越权 denied）。
- ✅ **Kafka 契约单源**：REQ-CPM-2 唯一权威，REQ-CPB-7 引用；attachments 空数组非 null；file_url 可再生；version=1 供未来协商。
- ✅ **REST 路由顺序**：marquee/publish-permission 静态路径先于 :id 注册（Task 1.23 已含）——防 `GET /notices/marquee` 被 `:id` 通配吞掉。
- ✅ **迁移/种子/依赖顺序**：Proto(0.1-0.3) → Kafka 基建(0.4) → 模型/写/读/Kafka/接口 → file(2) → permission(3) → moderation(4) → 运维(6)；003 一次性 RENAME 勿重跑（R4）标注到位；task 单服务/单代码层级/独立可测。

## 问题跟踪表

| # | 状态 | 说明 |
|---|------|------|
| 1 | 待修复（本轮新增） | UpdateContentPost 授权模型自相矛盾：作者校验 080002 使 REQ-CPB-9(f) 非发布者操作者置顶不可达 |
| 2 | 待修复（本轮新增） | R2 详情兼容回退仅支持单小区用户，多小区用户（grid_worker/多房产业主）notice-detail 迁移后 080005 |
| 3 | 待修复（本轮新增） | Task 0.1 080004 头注释标签「便民联络不存在/ContactService 仍用」与代码不符（实为寻失记录不存在，LostFoundService 用） |
| 4 | 待确认（强化） | property_admin 421↔427/428 不对称回归断言未显式落 Task 6.2（SHOULD 2） |

---
VERDICT: REVISION
---
