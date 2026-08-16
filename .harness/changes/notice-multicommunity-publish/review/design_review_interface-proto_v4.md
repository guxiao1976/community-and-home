# Design Review — notice-multicommunity-publish（接口契约+Proto 视角）v4

**审查模式**: 模式一.5 设计评审 · **视角**: interface-proto（接口契约 / Proto 破坏性 / 鉴权-幂等-错误码 / 依赖顺序）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`（07-19 修订后；v3 评审 07-14 之后又有更新）
**对照基准**: v3 评审问题清单 + proposal.md + 6 个 spec + 磁盘真相（community/file/permission proto、init_permissions.sql、routes.go、common/pkg/middleware/permission.go、permission-service checkpermissionlogic.go、web/mobile api/pages）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2
- **VERDICT: REVISION**（存在 ≥2 MUST FIX：通知读/写端点的权限矩阵在 fail-closed 鉴权下使本变更核心交付（详情页/撤回/浏览）不可用）

## 已核实（与磁盘真相一致，设计正确）

- **Proto 全部为兼容新增，字段号无冲突**：`CreateNoticeRequest.community_ids(8)/division_id(9)`（现状 1-7 占用，8/9 空闲）、`GetNoticeRequest.community_id(2)`（现状 id=1）、`NoticeAttachment.file_type(5)`（现状 1-4）、`FileInfo.file_type(11)/confirmed(12)`（现状 1-10）。`buf breaking-check` 可通过。
- **附件校验载体契约成立**：`GetFileUrlResponse = { base, download_url, FileInfo file }`（file.proto:173-178），`FileInfo` 含 `user_id(2)/file_size(7)` 及新增 `file_type(11)/confirmed(12)` —— CreateNotice 绑定校验（confirmed + user_id==JWT + file_type 回读）与 GetNotice 附件预签名 URL 重生（download_url）一次 RPC 双满足，零新增 RPC（D24 成立）。
- **批量 scope 校验契约成立**：`AssertPublishScopeRequest.repeated ScopeRef targets`（permission.proto:257-265）—— Task 1.6 单次批量（非 N+1）正确。
- **level-2 发布判定可基于 RPC 输出**：`UserRoleInfo` 含 `role.code/status(0未认证~4过期)/verified_at/expires_at`（permission.proto:353-366），无直读 rel_user_role。
- **错误码对齐判定准确**：file.proto 头注释现状「070002 上传失败/070003 文件类型不支持/070004 文件大小超限/070005 bucket 不存在」与实际常量（errcode.go 70001 FileNotFound/70002 AccessDenied/70003 OperationFailed）确漂移，D11 目标正确；community.proto 头注释「080003 寻失发布次数已达上限」确陈旧（实际码 080007），D29 判定正确。
- **权限种子编号核对**：421 现绑 (2,3,6,1,5)、grid_worker(4) 未持、min_verf_level 现 0（§4.2 UPDATE 置 0）；423/424/425 全部空闲 ——「授 4 / 回收 2/1/5 / 置 min_verf_level=2」+「423/424/425 全角色」计划与现状一致。
- **425 PermMiddleware 归属核实**：permission-service `routes.go:107-128` `rest.WithPrefix("/api/perm")` + `rest.WithMiddleware(serverCtx.PermMiddleware.Handle, routes...)` 包裹全部 /api/perm 路由 —— 425 端点复用同一中间件组成立；community-hub `routes.go` notice 路由同样在 PermMiddleware 组内。
- **语义破坏面收敛**：web/pc 无 notice 消费方（v3 已核实）；web/mobile 本变更内同步升级；GetNotice 必填 community_id + legacy CreateNotice 不回退 community_id(1) 已登记 CHANGELOG（Task 0.4）。
- **REST 绑定**：`CommunityIds []string`（encoding/json `,string` 不支持 slice，[jstype=JS_STRING] 正确）+ `DivisionId int64 json:",string"`（标量支持）+ GetNotice/GetMarqueeNotices `form:"community_id"`（GET query 走 ParseForm，`json` 标签被 skip）—— 与 go-zero httpx.Parse 行为一致。
- **依赖顺序**：Proto(0) → 迁移/模型(1.1-1.5) → scope 基础设施(1.6) → 写/读逻辑(1.7-1.13) → 注册/API(1.14-1.17) → 前端(4) → 运维验证(5)，合规。
- **幂等语义**：CreateNotice 声明不幂等（D25，前端提交中禁用）、UpdateNoticeModerationStatus 幂等（覆盖写），与 spec REQ-NP-7/REQ-NP-MOD-4 一致。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | tasks.md Task 3.2 + design §权限种子（设计漏项） | **通知读路径权限矩阵不完整，fail-closed 下核心读交付不可用。** 已核实磁盘真相：(a) `GET:/api/community/notices/:id`（详情）在 sys_permission **无任何权限码**（种子中 `notices/:id` 零匹配；对照 user 112 / role 212 均有显式种子详情读码并绑定角色），`checkpermissionlogic.go:55-59` `permissionDefMinLevel(needle)` 找不到 def → `Allowed=false` → **所有用户对详情端点一律 403**，GetNotice 的 080001/080005 契约永远不可达；(b) `GET:/api/community/notices`（浏览，422）仅绑 (9,1,5)（种子 :222/233/258-259），本变更新授 421/423 的 grid_worker(4)/community_admin(3)/committee(6) 均无 422 → **marquee「更多→浏览」(REQ-NM-1/Task 4.3) 把发布角色引向浏览页 (Task 4.4) 时 403**（v3 SHOULD 1 的 422 分裂在 07-19 修订后仍未处理）。设计自身端到端验收（Task 5.2「详情/列表 scope 过滤」、REQ-NM-1「点击标题→详情」、Task 4.5 NoticeDetail）将直接失败。 | Task 3.2 补齐读权限矩阵：新增并绑定详情读码（如 `426 GET:/api/community/notices/:id`，绑定全部移动端角色，同 423/424/425 批次）；**将 422 一并扩展绑定全部移动端角色**（与 423/424/425 同批，避免再出 425 因「复用 422 排除发布者」而新建码、422 本身却仍排除发布者的自相矛盾）；或显式声明发布角色不浏览 + 客户端隐藏 marquee「更多」入口，二选一不可悬空。Task 5.1 种子验证断言同步补 426/422 扩展。 |
| 2 | tasks.md Task 3.2 + design §DeleteNotice/§UpdateNotice（设计漏项） | **通知写路径权限矩阵缺失，撤回（REQ-NP-5 本变更核心交付）不可用。** 已核实磁盘真相：`DELETE:/api/community/notices/:id`（撤回，Task 1.12）与 `PUT:/api/community/notices/:id`（UpdateNotice）在 sys_permission **均无权限码**（种子零匹配）→ fail-closed 下所有用户 403，DeleteNotice 的 080001/080002 契约、UpdateNotice 回归测试在 REST 层全部被中间件拦截在前；Task 1.12 的「作者校验 080002」永远走不到（逻辑层测试虽绿，端到端撤回功能不可用），Task 5.2「撤回仅发布者本人」冒烟将失败。 | Task 3.2 新增并绑定 `DELETE:/api/community/notices/:id` 与 `PUT:/api/community/notices/:id` 权限码。授权面建议：DELETE/PUT 绑定全部移动端角色（真正越权判定交 080002 作者校验，与 design 语义一致），或显式声明撤回/编辑不在移动端开放并隐藏入口 + 客户端不做入口。Task 5.1/5.2 断言同步。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 3 | tasks.md Task 3.2（v3 SHOULD 2 仍存在） | 新权限码 423/424/425 的 `parent_id` 未指定。sys_permission 需 parent_id：423/424 自然挂 410（community:read，同 422）；425 code 为 `community:data-scopes:read-api` 但路由在 `/api/perm/*`，菜单树归属不确定（执行方需自行推断，AutoDiscover 的 findParent 对 /api/perm/data-scopes 按 resource=perm→parts[3]=data-scopes → resourceToMenu miss → 挂 /data-scopes 菜单，与既有 410 树不一致）。 | Task 3.2 显式标注三条新码 parent_id（423/424→410；425 建议挂 310 permission:read 或新增 data-scope 节点并明示），同步 Task 5.1 种子验证断言。 |
| 4 | tasks.md Task 4.1/4.6（v3 INFO 4 仍存在） | REST `DivisionId int64 json:"division_id,string"`：移动端若对非 community_admin 发送 `division_id: ""`（空串）而非省略字段，`encoding/json ,string` 解包空串报错（REST 层 4xx），进不到 RPC 080005。 | Task 4.1/4.6 注明「不适用时省略 division_id 字段，不发送空串」；RPC 侧 double-empty 判 080005 已兜底。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 5 | Task 3.2「435（寻失发布）保持 min_verf_level=0 不动」引用的事实准确性：sys_permission 无 id=435 行（`community:lostfound:create-api` 仅出现在 line 202 的 UPDATE WHERE code IN，无 INSERT 定义；rel_role_permission 却绑 (1,435),(5,435) 指向幻影 id）。建议措辞改为「lostfound:create-api 权限无 sys_permission 行，保持现状不动」，并核 Task 5.1 断言不与幻影 435 冲突。 |
| 6 | design §权限种子引用 rbac-design.md §6.5 验收矩阵，但当前 rbac-design.md 中无 422/423/424/425 的 notice 读权限矩阵行（grep 零命中）——建议随本变更将 422/423/424/425/426 权限码写入 rbac-design.md §6.5（.change.yaml 已列该文件为修订对象）。 |

## 问题跟踪表

| 状态 | 说明 |
|------|------|
| 已修复（v1 MUST 1/2/3 + SHOULD 4/5/6/7、v2 MUST 1 + SHOULD 2、v3 MUST 0） | 此前轮次已验证落地，本轮复核保持 |
| 待修复（本轮 MUST 1/2） | 读/写路径权限矩阵缺失——详见「发现」MUST 1/2，阻塞级 |
| 待修复（本轮 SHOULD 3/4） | 423/424/425 parent_id + division_id 空串绑定（不阻塞） |
| 待修复（本轮 INFO 5/6） | 幻影 435 措辞 + rbac-design §6.5 矩阵补登 |

## 报告自检
1. 本轮 MUST FIX 均定位到具体文件（seed / Task 章节）并附可落地修复，非空泛 ✅
2. 接口契约维度覆盖：Proto 破坏性（字段号/新 RPC）、REST 绑定（form/json）、**鉴权（421/422/423/424/425 种子+中间件，含 fail-closed 命中）**、错误码语义（D11/D29/080005 收敛）、幂等（不幂等声明）、必填回归（GetNotice community_id）、依赖顺序、路由冲突 ✅

---
VERDICT: REVISION
---

## 结论
设计在 Proto 契约（字段号/新 RPC 全部兼容、GetFileUrl+FileInfo 附件载体一次 RPC 双满足、批量 AssertPublishScope、level-2 判定）、错误码对齐（D11/D29）、REST 绑定（form/`,string`）等维度与磁盘真相一致，且 v1-v3 的 MUST/SHOULD 已逐条落地。但本变更把通知模块的**读（详情/浏览/跑马灯）与写（发布/撤回）**做成移动端核心交付，而权限种子任务（Task 3.2）只管理 421/423/424/425，**漏掉了 `GET/PUT/DELETE /api/community/notices/:id` 与 422 扩展**：在 fail-closed 的 CheckPermission（exact-path、缺 def 即拒）下，详情页对全体用户 403、撤回对全体用户 403、浏览页对本变更新授发布角色 403——直接击穿设计自身的端到端验收（Task 5.2）与核心交付（NoticeDetail/Task 4.5、撤回/Task 1.12、marquee→浏览）。需架构设计师在 Task 3.2 补齐读/写权限矩阵后复审。
