# Design Review — notice-multicommunity-publish（接口契约+Proto 视角）v5

**审查模式**: 模式一.5 设计评审 · **视角**: interface-proto（接口契约 / Proto 破坏性 / 鉴权-幂等-错误码 / 依赖顺序）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md`（07-32）+ `tasks.md`（07-33）——即 v4 评审（07-28）之后架构设计师修订的最新版本
**对照基准**: v4 评审问题清单 + proposal.md + 6 个 spec + 磁盘真相（community/file/permission proto、init_permissions.sql、routes.go、checkpermissionlogic.go、masterdata division 展开、moderation task_handler.go、web/mobile community.ts、model/notice.go、updatenoticelogic.go）
**审查时间**: 2026-08-16
> 注：任务指令给定落盘名 v2，但 review 目录已有 v1–v4 且"旧版不删"，故本轮按版本递增写 v5。

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2
- v4 的 2 个 MUST FIX（读/写权限矩阵）+ SHOULD 3/4 + INFO 5/6 **已全部落地**（426/427/428 新增、422 扩展全部移动端角色、parent_id 显式 423/424/426→410、427/428→420、425→310、幻影 435 措辞、rbac-design §6.5 补登），本轮复核保持。
- **VERDICT: REVISION**（1 个 MUST FIX：UpdateNotice 编辑路径未适配多小区数据模型——编译破坏 + 新通知不可编辑 + 428 授权契约自相矛盾）

## 已核实（与磁盘真相一致，设计正确）

- **Proto 全部为兼容新增，字段号落在空闲位，`buf breaking-check` 可通过**：`CreateNoticeRequest.community_ids(8)/division_id(9)`（现状 1-7 占用，8/9 空闲）、`GetNoticeRequest.community_id(2)`（现状仅 id=1）、`NoticeAttachment.file_type(5)`（现状 1-4）、`FileInfo.file_type(11)/confirmed(12)`（现状 1-10）。`community_id(1)/role(4)/publisher_id(6)` 标 deprecated 为兼容新增；`publisher(5)` 不 deprecated（展示字符串，NOT NULL 列来源显式）——与既有 wire 兼容。
- **`GetPublishPermissionResponse.publishable_roles` 用 `NoticeRole` 枚举**（community.proto:56-62：UNSPECIFIED=0/COMMUNITY=1/COMMITTEE=2/PROPERTY=3/GRID_OFFICER=4），design 映射 grid_worker→GRID_OFFICER、community_admin→COMMUNITY、committee→COMMITTEE 与枚举一致；REST `PublishableRoles []int32` 与枚举 Go 类型对齐。
- **附件校验载体契约成立**：`GetFileUrlResponse = { base, download_url, FileInfo file }`（file.proto:173-178），`FileInfo` 含 `user_id(2)/file_size(7)` 及新增 `file_type(11)/confirmed(12)`——CreateNotice 绑定校验（confirmed + user_id==JWT + file_type 回读）与 GetNotice 预签名 URL 重生（download_url）一次 RPC 双满足，零新增 RPC（D24 成立）。
- **批量 scope 校验契约成立**：`AssertPublishScopeRequest = { user_id=1, repeated ScopeRef targets=2 }`（permission.proto:257-264）——Task 1.6 单次批量（非逐目标 N+1）正确。
- **level-2 发布判定可基于 RPC 输出**：`UserRoleInfo` 含 `role/scope_type/scope_id/status(0~4)/verified_at/expires_at`（permission.proto:353-366）+ `GetUserRoles` RPC 存在——GetPublishPermission/GetUserRoles 角色派生不直读 rel_user_role，grpc-only-comms 遵守。
- **回调复用现有消息**：`UpdateModerationStatusRequest = { id=1, moderation_status=2 }` 注释 `1=machine_pass,2=machine_fail,3=human_pass,4=human_fail`（community.proto:314-317）——REQ-NP-MOD-2 复用、`IsModerationPassed IN(1,3)` 单一谓词语义与回调/读查询共用自洽。
- **错误码对齐判定准确（D11/D29）**：file.proto 头注释现状「070002 上传失败/070003 文件类型不支持/070004 文件大小超限/070005 bucket 不存在」与实际常量（errcode.go 70001 FileNotFound/70002 AccessDenied/70003 OperationFailed）确漂移；community 常量 CodeOverLimit=80003（超限）/CodeInvalidParam=80005/CodeScopeDenied=80006/CodeSectionQuotaExceeded=80007 存在——080003→"单次发布目标数超限"、080007 承载"寻失发布次数已达上限"语义、080006 已存在，设计判定正确。
- **权限种子现状核对**：421 现绑 (2,3,6,1,5)、grid_worker(4) 未持、422 仅绑 (9,1,5)（line 233/258-259）、**423-429 全部空闲**、435 幻影（无 INSERT，仅 rel_role_permission (1,435)/(5,435) + line 202 UPDATE 引用）、436 真实 INSERT——「授 4 / 回收 2/1/5 / 置 min_verf_level=2」+「426/427/428 新增 + 422 扩展全部移动端角色 + 423/424/425」目标与现状一致。
- **fail-closed 鉴权语义核实**：checkpermissionlogic.go `permissionDefMinLevel(needle)` 缺 def → `Allowed=false` → 403——v4 MUST 1/2 判据成立，本次设计补齐 426/427/428 后各 REST 端点权限码齐备。
- **中间件归属核实**：permission-service routes.go:124 `rest.WithMiddleware(serverCtx.PermMiddleware.Handle, routes...)` + line 128 `rest.WithPrefix("/api/perm")` 包裹全部 /api/perm 路由——425 data-scopes 复用同一中间件组成立；community-hub routes.go notice 路由（POST/GET `/notices` + GET/PUT/DELETE `/notices/:id`）全部在 `PermMiddleware` 组内。
- **division 展开契约核实**：masterdata `GetResidentialAreasByDivision` 的 `case in.CommunityDivId > 0` → FindByCommunityDivId，否则默认 FindAll（≤1000）；`status==1` 时过滤 `SubmissionStatus!=2`——design guard `divisionID<=0 → 080005` 杜绝进入 FindAll 过度展开，正确（评审 SHOULD 5 落地）。
- **REST 绑定**：`CommunityIds []string`（encoding/json `,string` 不支持 slice，[jstype=JS_STRING] 正确）+ `DivisionId int64 json:",string"`（标量支持）+ GetNotice/GetMarqueeNotices `form:"community_id"`（GET query 走 ParseForm，`json` 标签被 skip）——与 go-zero httpx.Parse 行为一致（v1/v2 已核实）。
- **web/mobile**：community.ts 现仅 getNoticeList/getNoticeDetail/getContacts/getLostFoundList，**无 createNotice**（INFO 3 判定正确，属新增非扩展）；moderation task_handler.go 仅映射 1=machine_pass/2=machine_fail（status=3 为潜在路径，IsModerationPassed IN(1,3) 统一封死）。
- **依赖顺序**：Proto(0) → 迁移/模型(1.1-1.5) → scope 基础设施(1.6) → 写/读逻辑(1.7-1.13) → 注册/API(1.14-1.17) → 前端(4) → 运维验证(5)，基础设施→核心→辅助→前端合规；Proto 变更独立成任务（0.1/0.2/0.3）且 Migration（1.1/2.5）独立，任务粒度刚性校验通过。
- **幂等语义**：CreateNotice 不幂等（D25，前端提交中禁用）、UpdateNoticeModerationStatus 幂等（覆盖写），与 REQ-NP-7/REQ-NP-MOD-4 一致。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | design.md §UpdateNotice + tasks.md（缺 UpdateNotice 适配 Task）+ `updatenoticelogic.go:42` | **编辑路径未适配多小区数据模型，设计自相矛盾且现有代码将编译破坏。** 磁盘证据：(a) `updatenoticelogic.go:42` 调 `scope.CheckPublishScope(l.ctx, client, notice.CommunityId)`（签名 `communityID int64`，scope.go:82）；Task 1.3 将 `Notice.CommunityId` 由 `int64`→`*int64`（model/notice.go:14）后该调用**编译失败**，而 tasks.md **无任何 Task 修改 updatenoticelogic.go**（design §UpdateNotice 明确写"逻辑无代码变更，仅回归测试确认"）——对照 deletenoticelogic:38 有 Task 1.12 重写、updatemoderationstatuslogic:45 有 Task 1.13 适配（legacy 行 *int64 取值 + 新行 FindCommunityIdsByNoticeId），唯独 updatenoticelogic 成孤儿。(b) 语义上：新多小区通知 `community_id=NULL`（弃用列不写入）→ 传 NULL→0 给 `CheckPublishScope` → `AssertPublishScope(target=0)` 未知节点 fail-closed deny → **新通知永远无法编辑**，与 REQ-NP-MOD-3"保持通知级语义可编辑"冲突。(c) 契约自相矛盾：§权限种子 428 行称"真正越权判定交业务层 080002 作者校验 / 通知级语义"，§UpdateNotice 却称"无新增越权判定"，而 UpdateNotice 现无作者校验（仅 CheckPublishScope 数据范围判定）——编辑与撤回（DeleteNotice 收窄为作者 080002）授权不对称，428 绑全部移动端角色后编辑授权面未定义。 | 新增 Task（或并入 Task 1.13 同款模式）重写 `updatenoticelogic.go`：新模型 reverse-lookup `FindCommunityIdsByNoticeId(id)` → `AssertCommunitiesScope(user, communities)` 单次批量（与 Create/Delete 一致）；legacy 行回退解引用 `notice.CommunityId`（*int64 取值）；**显式定案 428 授权语义**（作者校验 080002 与 427 对齐，或发布 scope 校验，二选一不可悬空）；design §UpdateNotice 删除"逻辑无代码变更"表述；Task 5.2 补"新多小区通知可编辑 + 非作者编辑拒绝"冒烟断言。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 2 | design.md §Proto（D29 错误码对齐）+ §DeleteNotice | `DeleteNotice` 非作者撤回复用 **080002**（`CodePublishDenied`，头注释"无发布权限（未认证对应社区角色）"），实际语义为"无权限操作此通知（非作者）"，与头注释不吻合。D29 正对齐 community/v1 头注释错误码块（Task 0.1），应一并扩写 080002 语义，否则调用方（前端错误文案映射 080002 → "未认证发布角色"）会对非作者撤回提示错误。 | Task 0.1 D29 将 080002 头注释扩为「无权限（未认证对应社区角色 / 非作者无操作权）」或显式登记 080002 现亦承载"非作者撤回"语义。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 3 | design §CreateNotice 校验顺序第 4 步 / Task 1.6 `ExpandDivisionCommunities`：`DivisionId` 值域仅 guard `divisionID<=0`，未校验节点确为 `community_div` 类型。md_administrative_division 含 county/street/community_div 三层，`GetResidentialAreasByDivision` 的 `CommunityDivId>0` 分支对 county/street id 返回空集 → 空展开 080005 兜底可接受；建议 design 注明"非 community_div 类型节点 → 空展开 → 080005"这一兜底依赖，防执行方误以为任意 md 节点 id 均可传。 |
| 4 | design §DeleteNotice / Task 1.12：`Notice.PublisherId` 为 `*int64`（可空）。存量行若 publisher_id 为空，作者校验 `publisher_id == JWT` 恒 false → 无人可撤回该通知。建议 Task 1.12 注明"既有 CreateNotice 恒写 publisher_id=JWT，存量非空"的先验；如确存 NULL 行，声明不可撤回或回退 legacy 数据范围判定。 |

## 问题跟踪表

| 状态 | 说明 |
|------|------|
| 已修复（v1 MUST 1/2/3 + SHOULD 4/5/6/7、v2 MUST 1 + SHOULD 2、v3 SHOULD 1/2、v4 MUST 1/2 + SHOULD 3/4 + INFO 5/6） | 全部落地并经磁盘证据复核保持（见"已核实"） |
| 待修复（本轮 MUST 1） | UpdateNotice 编辑路径授权契约未定义 + 编译破坏 + 新通知不可编辑——阻塞级 |
| 待修复（本轮 SHOULD 2 + INFO 3/4） | 080002 头注释语义扩写；division 类型兜底注明；publisher_id NULL 先验（不阻塞） |

## 报告自检
1. 本轮 MUST FIX 定位到具体文件（`updatenoticelogic.go:42` / design §UpdateNotice / tasks 缺 Task）并附可落地修复，非空泛 ✅
2. 接口契约维度覆盖：Proto 破坏性（字段号/新 RPC/枚举）、REST 绑定（form/json/`,string`）、**鉴权（421-428 种子+中间件 fail-closed 命中）**、错误码语义（D11/D29/080002 复用）、幂等（不幂等声明）、必填回归（GetNotice community_id）、依赖顺序、路由冲突（静态先于 :id）、跨服务语义破坏登记 ✅

---
VERDICT: REVISION
---

## 结论
接口契约与 Proto 维度整体自洽且与磁盘真相高度一致：Proto 变更全部为兼容新增、错误码对齐判定（D11/D29）准确、附件载体（GetFileUrl+FileInfo）一次 RPC 双满足、批量 AssertPublishScope 与 level-2 判定契约成立、v4 的权限矩阵 MUST 1/2（426/427/428 + 422 扩展）已逐条落地并经种子/路由核实。唯一阻塞点是 **UpdateNotice 编辑路径**：Task 1.3 将 `Notice.CommunityId` 改 `*int64` 后，`updatenoticelogic.go:42` 的 `CheckPublishScope(notice.CommunityId)` 编译即断且无任何 Task 覆盖（design 误称"逻辑无代码变更"）；语义上新多小区通知（community_id=NULL）经 fail-closed scope 校验永远无法编辑，与 REQ-NP-MOD-3 冲突；且 428 绑全部移动端角色的授权语义在 design 内部自相矛盾（"080002 作者校验" vs "无新增越权判定"）。需架构设计师补 UpdateNotice 适配 Task 并定案 428 授权契约后复审。
