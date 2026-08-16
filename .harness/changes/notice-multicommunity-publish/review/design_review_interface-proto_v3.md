# Design Review — notice-multicommunity-publish（接口契约+Proto 视角）v3

**审查模式**: 模式一.5 设计评审 · **视角**: interface-proto（接口契约 / Proto 破坏性 / 依赖顺序 / 鉴权-幂等-错误码）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`（v3，修订后）
**对照基准**: v2 评审问题清单 + proposal.md + 6 个 spec + 磁盘真相（community/file/permission proto、init_permissions.sql、routes.go、GetFileUrl/GetDataScopes/GetUserRoles/AssertPublishScope 契约、web/mobile 消费方）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2
- 上一轮 MUST FIX 1 项 + SHOULD 1 项：**全部已修复**（见「问题跟踪表」）
- **VERDICT: APPROVED**（无 MUST FIX，本轮接口契约无阻塞问题）

## 上一轮问题修复验证（v2 接口契约 MUST/SHOULD 逐条核对）

| # | v2 问题 | 修复落地证据 | 状态 |
|---|---------|-------------|:---:|
| MUST 1 | GetNotice community_id 用 `json` 标签，GET query 绑定被 go-zero 跳过 → 详情页恒 080005 全挂 | tasks.md Task 1.15 `GetNoticeReq.CommunityId` 改 **`form:"community_id"`**（禁 json，注明 `unmarshaler.go:571 usingDifferentKeys("form", field)` skip 行为）+ RED 断言 `GET /notices/:id?community_id=456` 绑定；Task 1.16 marquee `GetMarqueeNoticesReq.CommunityId` 同用 form 标签 + 绑定断言；design §GetNotice/§GetMarqueeNotices 补 REST 绑定说明。仓库既有惯例 `ListNoticesReq form:"community_id"` 一致。 | ✅ 已修复 |
| SHOULD 2 | AssertPublishScope 判据放宽的 blast radius（lostfound/contacts 共享调用方）未 acknowledge | tasks.md Task 3.1 验证门禁追加条目 4/5（community_div=D1 时 lostfound 发布 C1 → allowed、contacts upsert C1 → allowed，或显式隔离）；design §Design Gate 新增「共享 blast radius 声明」+ ADR 行（评审 SHOULD 2 显式声明）。 | ✅ 已修复 |

## 已核实（设计正确、与磁盘真相一致的部分）

- **Proto 全部为兼容新增，字段号无冲突**：`CreateNoticeRequest.community_ids(8)/division_id(9)`（现状 1-7 占用，8/9 空闲）、`GetNoticeRequest.community_id(2)`（现状 id=1）、`NoticeAttachment.file_type(5)`（现状 1-4）、`FileInfo.file_type(11)/confirmed(12)`（现状 1-10）。`buf breaking-check` 可通过。
- **附件校验载体契约成立**：`GetFileUrlResponse = { base, download_url, FileInfo file }`（file.proto:173-178），`FileInfo` 含 `user_id(2)` 及新增 `file_type(11)/confirmed(12)` —— CreateNotice 绑定校验（confirmed + user_id==JWT + file_type 回读）与 GetNotice 附件预签名 URL 重生（download_url）**一次 RPC 双满足**，零新增 RPC（D24 成立）。
- **批量 scope 校验契约成立**：`AssertPublishScopeRequest.repeated ScopeRef targets`（permission.proto:257-265）—— Task 1.6 单次批量（非 N+1）正确。
- **level-2 发布判定可基于 RPC 输出**：`UserRoleInfo` 含 `role.code/status(0未认证~4过期)/verified_at/expires_at`（permission.proto:353-366），无直读 rel_user_role。
- **错误码对齐判定准确**：file.proto 头注释现状「070002 上传失败/070003 文件类型不支持/070004 文件大小超限/070005 bucket 不存在」与实际常量（errcode.go 70001 FileNotFound/70002 AccessDenied/70003 OperationFailed）确漂移，D11 目标（070004 类型不支持/070005 大小超限）正确；community.proto 头注释「080003 寻失发布次数已达上限」确陈旧（实际码 080007，D29 判定正确），080005 现状「小区ID无效」→ 新「参数无效（含小区ID无效）」语义收敛。
- **权限种子编号核对**：423/424/425 在 init_permissions.sql 中**全部空闲**（现状 4xx 占用 400/410/411/420/421/422/430/431/432/433/434/436）；421 现绑 (2,3,6,1,5)、grid_worker(4) 未持、422 仅绑 (9,1,5) ——「授 4 / 回收 2/1/5 / 置 min_verf_level=2」+「425 全角色」目标与现状一致。
- **425 PermMiddleware 归属核实**：permission-service `routes.go:124` `rest.WithMiddleware(serverCtx.PermMiddleware.Handle, routes...)` + `rest.WithPrefix("/api/perm")` 包裹全部 /api/perm 路由——「新端点走同一中间件组」声明成立；RouteRegistry 登记（Task 3.3）与既有 auto-discover 惯例一致。
- **语义破坏面收敛**：web/pc 无 notice 消费方（grep 零命中）；web/mobile 在本变更内同步升级（详情页 Task 4.5 携带 community_id）；无其他服务直连 GetNotice/CreateNotice RPC——GetNotice 必填 community_id + legacy CreateNotice 不回退 community_id(1) 的破坏面已核实收敛，CHANGELOG 已登记（Task 0.4）。
- **路由注册**：`GET /notices/marquee` / `/publish-permission` 静态路径先于 `:id` 注册已由 Task 1.17 显式处理；go-zero 静态段优先于参数段，双保险。
- **响应契约**：`Notice.attachments`(12) 已存在（community.proto:77），GetNotice 输出 `Notice + attachments[]（含 file_type）` 契约自洽，无需新增字段。
- **依赖顺序**：Proto(0) → 迁移/模型(1.1-1.5) → scope 基础设施(1.6) → 写/读逻辑(1.7-1.13) → 注册/API(1.14-1.17) → 前端(4) → 运维验证(5)，基础设施→核心→辅助→前端合规。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 1 | tasks.md Task 3.2（+design §权限种子） | **422（ListNotices 浏览列表）绑定未扩展到发布者角色**。本变更新增 423/424/425 绑定**全部移动端角色**，但 422（`GET /api/community/notices`，浏览页数据源）仍仅绑 registered_user(9)/owner(1)/tenant(5)（init_permissions.sql:233/258-259）。REQ-NM-1「更多→浏览」经新 marquee（423 全角色）把 grid_worker/community_admin/committee 引导到浏览页，但浏览页调 getNoticeList → 422 → 这几位（本变更刚授 421 的发布者）PermMiddleware 403。属跨功能一致性缺口（421/423 与 422 绑定集分裂）。 | Task 3.2 把 422 一并扩展绑定全部移动端社区角色（与 423/424/425 同批），或在 design §权限种子显式声明「发布者角色不浏览列表」并由 Task 4.4 客户端对 marquee「更多」入口按 can_publish 或角色隐藏（避免 403 空页）。二选一，不可悬空。 |
| 2 | tasks.md Task 3.2 | **新增权限码 423/424/425 的 `parent_id` 未指定**。sys_permission INSERT 需 parent_id：423/424 自然挂 410（community:read，同 422/411）；425 是 `community:*` 代码（community:data-scopes:read-api）但路由在 `/api/perm/*`（permission:menu 树）——执行方需自行推断，菜单树放置不确定（尤其 425）。 | Task 3.2 为三条新码显式标注 parent_id（423/424→410；425 建议挂 310 permission:read 或新增 data-scope 菜单节点，明示归属），并同步 Task 5.1 种子验证断言。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 3 | Task 3.2「435（寻失发布）保持 min_verf_level=0 不动」引用的事实准确性：init_permissions.sql 中 **sys_permission 无 id=435 行**（`community:lostfound:create-api` 仅出现在 line 202 的 UPDATE WHERE code IN...，无 INSERT 定义；rel_role_permission 却绑定 (1,435),(5,435) 指向幻影 id）。属既有种子数据完整性瑕疵（非本变更引入）。建议 Task 3.2 措辞改为「lostfound:create-api 权限无 sys_permission 行，保持现状不动」，并核 Task 5.1 种子验证断言不与幻影 435 冲突。 |
| 4 | Task 4.6 发布表单的 `division_id` 空值语义：REST `DivisionId int64 json:"division_id,string"`，若移动端对非 community_admin 发送 `division_id: ""`（空串）而非**省略字段**，`encoding/json ,string` 会解包报错（REST 层 4xx），进不到 RPC 080005。建议 Task 4.1/4.6 注明「不适用时省略 division_id 字段，不发送空串」；RPC 侧 double-empty 判 080005 已兜底。 |

## 问题跟踪表

| 状态 | 说明 |
|------|------|
| 已修复（上轮 MUST 1 + SHOULD 2，已验证） | 逐条见「上一轮问题修复验证」 |
| 已修复（v1 MUST 1/2/3 + SHOULD 4/5/6/7，此前已验证） | — |
| 待修复（本轮 SHOULD 1/2） | 422 绑定集分裂 + 423/424/425 parent_id 未指定（不阻塞，建议下一轮或 BACKLOG 落实） |

## 报告自检
1. 本轮 MUST FIX：0；SHOULD/INFO 均定位到具体 Task 并附可落地建议 ✅
2. 接口契约维度覆盖：Proto 破坏性（字段号/新 RPC）、REST 绑定（form/json）、鉴权（421/423/424/425 种子+中间件）、错误码语义（D11/D29/080005 收敛）、幂等（不幂等声明）、必填回归（GetNotice community_id）、依赖顺序、路由冲突 ✅

---
VERDICT: APPROVED
---

## 结论
v3 设计接口契约维度无阻塞问题。v2 的 MUST FIX（GetNotice community_id 用 `json` 标签 → form 标签修正）与 SHOULD 2（AssertPublishScope 共享 blast radius 声明+回归）已逐条落地并经磁盘证据核实；Proto 全部为兼容新增且字段号落在空闲位、附件校验载体（GetFileUrl 返回 FileInfo + download_url）一次 RPC 双满足、错误码对齐判定（D11/D29）、425 data-scopes 鉴权钉死+PermMiddleware 归属、web/pc 无消费方的语义破坏面收敛——全部与代码真相一致。仅余两条不阻塞的建议级发现（422 浏览列表未随 421/423 一并扩展给发布者角色、新权限码 parent_id 未指定），建议随 421/423/424/425 种子改动一并落实或登记 BACKLOG。
