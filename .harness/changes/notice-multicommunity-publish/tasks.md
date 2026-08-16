# Tasks: 通知模块多小区发布 + 通栏跑马灯 + 附件安全

> **对执行 Agent 的指令**: 每个 Task 独立可测，按 TDD 执行（先写测试→看失败→写实现→看通过）。精确到文件路径。
> 依赖顺序: Proto → Migration/模型 → Logic → Handler/Routes → 前端。迁移/种子为「Owner 运维验证」任务，不走 harness-pipeline（无 DB 环境），派发时显式排除。

---

## 全局 / Proto（由全局 Claude 执行，不分发）

### Task 0.1: community.proto 消息字段演化
- **文件**: `api-proto/api/community/v1/community.proto`
- [ ] 头注释错误码块对齐实际语义（D29）：`080003 — 单次发布目标数超限`、`080005 — 参数无效（含小区ID无效）`、`080006 — 目标小区超出发布者数据范围`；剔除陈旧「080003 寻失发布次数已达上限」（该语义实际码为 080007）
- [ ] `CreateNoticeRequest` 新增 `repeated int64 community_ids = 8 [jstype = JS_STRING]`（多小区目标）
- [ ] `CreateNoticeRequest` 新增 `int64 division_id = 9 [jstype = JS_STRING]`（仅 community_admin，值域 `md_administrative_division.id`）
- [ ] `CreateNoticeRequest.community_id`(1) 标记 deprecated（保留 wire 兼容，服务端不回退）
- [ ] `CreateNoticeRequest.role`(4)/`publisher_id`(6) 标记 deprecated（服务端从 JWT 派生，请求体不信任，评审 SHOULD 4）；`publisher`(5) **不 deprecated**（请求体展示字符串，评审 S1）
- [ ] `GetNoticeRequest` 新增 `int64 community_id = 2 [jstype = JS_STRING]`（必填请求上下文，缺失 080005）
- [ ] `NoticeAttachment` 新增 `string file_type = 5`
- [ ] 所有 int64 均 `[jstype = JS_STRING]`（SEE [[proto-jstype]]）；deprecated 是机器可读指令，勿当占位注释（SEE [[go-deprecated-directive-not-test-comment]]）
- [ ] 保持兼容（新字段号），`buf breaking-check` 通过

### Task 0.2: community.proto 新增 RPC（GetPublishPermission / GetMarqueeNotices）
- **文件**: `api-proto/api/community/v1/community.proto`
- [ ] `NoticeService` 新增 `rpc GetPublishPermission(GetPublishPermissionRequest) returns (GetPublishPermissionResponse)`（D3）
- [ ] 新增 `message GetPublishPermissionRequest {}`（空，身份经 JWT metadata）
- [ ] 新增 `message GetPublishPermissionResponse { common.v1.BaseResp base = 1; bool can_publish = 2; repeated NoticeRole publishable_roles = 3; }`
- [ ] `NoticeService` 新增 `rpc GetMarqueeNotices(GetMarqueeNoticesRequest) returns (GetMarqueeNoticesResponse)`（D12）
- [ ] 新增 `message GetMarqueeNoticesRequest { int64 community_id = 1 [jstype = JS_STRING]; }`
- [ ] 新增 `message NoticeMarqueeItem { int64 id = 1 [jstype = JS_STRING]; string title = 2; }`
- [ ] 新增 `message GetMarqueeNoticesResponse { common.v1.BaseResp base = 1; repeated NoticeMarqueeItem items = 2; }`
- [ ] `make ci` 通过（lint + breaking-check + generate）

### Task 0.3: file.proto 变更（FileInfo 扩展 + 错误码对齐）
- **文件**: `api-proto/api/file/v1/file.proto`
- [ ] 头注释错误码块对齐实际常量（D11）：`070001 文件不存在 / 070002 文件访问被拒绝 / 070003 文件操作失败 / 070004 文件类型不支持 / 070005 文件大小超限`；修正漂移的「070002 上传失败 / 070003 文件类型不支持 / 070004 文件大小超限 / 070005 bucket 不存在」
- [ ] `FileInfo` 新增 `string file_type = 11`（白名单规范类型，ConfirmUpload magic-bytes 层产出）
- [ ] `FileInfo` 新增 `bool confirmed = 12`（上传流程完成标志）
- [ ] 保持兼容（新字段号），`make ci` 通过

### Task 0.4: CHANGELOG 登记 + 生成 + CI 门禁
- **文件**: `api-proto/CHANGELOG.md`
- [ ] 登记 community/v1 全部变更（含头注释错误码语义迁移 D29）
- [ ] 登记 file/v1 全部变更（含头注释语义迁移 D11、070004/070005 登记）
- [ ] **登记语义破坏项（评审 SHOULD 6）**：`GetNotice` 新增必填 `community_id`（缺失→080005，未升级消费方全 080005）+ legacy `CreateNotice` 仅传 `community_id`(1) 不再接受（→080005，不回退）；注明兼容期与本变更的同步升级范围（web/mobile 已升级、web/pc 无 notice 消费方）
- [ ] **登记附件 file_url 语义（评审 S4）**：`NoticeAttachment.file_url` 为**短期预签名 URL**（发布时快照，3600s~7 天），详情读路径按 `file_id` 重生，勿当永久链接——见 design.md §GetNotice
- [ ] `cd api-proto && make ci` → lint 0 errors + breaking-check 无破坏性 + 生成同步
- [ ] 通知受影响服务：community-hub-service / file-service / permission-service / web/mobile（生成代码已本地可见）

---

## community-hub-service

### Task 1.1: Migration 003（notice_scope + 去 NOT NULL + file_type + file_id）
- **创建**: `services/community-hub-service/migration/003_multi_community_notice.sql`
- [ ] `ALTER TABLE notices MODIFY community_id BIGINT DEFAULT NULL`（D1 弃用列，兼容期保留不写入）
- [ ] `ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL`（D30，创建写 NULL、审核通过回调置 now）
- [ ] 新建 `notice_scope` 表：复合 PK `(notice_id, community_id)`（= uk_notice_community 唯一约束）+ `KEY idx_scope_community (community_id, notice_id)`（读索引，双列 NOT NULL）
- [ ] `ALTER TABLE notice_attachments ADD COLUMN file_type VARCHAR(20) DEFAULT NULL`
- [ ] `ALTER TABLE notice_attachments ADD COLUMN file_id BIGINT DEFAULT 0`（设计评审 data-model v3 M1：附件预签名 URL 重生载体，S4 闭环；CreateNotice 写、GetNotice 按 file_id 重生）
- [ ] `ALTER TABLE notice_attachments MODIFY file_url VARCHAR(1024) NOT NULL`（设计评审 data-model v4 S4：新行写占位空串不落 URL 快照，file_id 为权威载体；加宽防存量/边界长 URL 触发 VARCHAR(500) ERROR 1406 → 整事务回滚）
- [ ] 兼容期保留 `idx_community`/`idx_published`（标记 deprecated，不删）
- **Owner 运维验证**: 三态库执行 + 迁移先于功能上线门禁（见 Task 5.1/5.2）
- **SEE**: [[migration-must-execute]] — 提交后必须执行；[[unique-index-migration-dup-precheck]] 排除（新表无存量重复，见 design.md「不适用记忆」）

### Task 1.2: NoticeScopeModel 新建 + ServiceContext 注册
- **创建**: `services/community-hub-service/model/notice_scope.go` + `notice_scope_test.go`
- **修改**: `services/community-hub-service/rpc/internal/svc/servicecontext.go`
- [ ] 定义 `NoticeScope` struct（notice_id/community_id/created_at，复合主键）
- [ ] 定义 `NoticeScopeModel` 接口：`InsertBatch(ctx, noticeId, communityIds []int64) error`（批量插入，业务层先 dedupe）、`DeleteByNoticeId(ctx, noticeId) error`（物理删除，撤回用）、`FindCommunityIdsByNoticeId(ctx, noticeId) ([]int64, error)`
- [ ] **RED**: table-driven 测试（正常批量插入 + 空列表 + 重复目标去重后单行 + 物理删除后查无）→ 确认 FAIL
- [ ] **GREEN**: 实现（InsertBatch 对去重后 targets；DeleteByNoticeId 物理 DELETE；FindCommunityIdsByNoticeId 按 notice_id 查询）→ 确认 PASS
- [ ] ServiceContext 注册 `NoticeScopeModel: model.NewNoticeScopeModel(conn)`

### Task 1.3: Notice / NoticeAttachment 模型字段调整（nullable + file_type）
- **修改**: `services/community-hub-service/model/notice.go` + `model/notice_attachment.go`
- [ ] `Notice.CommunityId` 由 `int64` → `*int64`（列已可空；弃用列不再写入）
- [ ] `Notice.PublishedAt` 由 `time.Time` → `sql.NullTime`（D30：创建 NULL、审核通过置 now；严禁 `time.Time{}` 零值写 DATETIME，SEE [[restore-compensation-zero-time]]）
- [ ] `Insert` SQL 移除 `community_id` 与 `published_at` 两列（不写入弃用/可空列）
- [ ] `NoticeAttachment` 增加 `FileType string \`db:"file_type"\``；`Insert` 包含 file_type 列
- [ ] `NoticeAttachment` 增加 `FileId int64 \`db:"file_id"\``（设计评审 data-model v3 M1：重生预签名 URL 载体）；`Insert` 包含 file_id 列（来自 attachment_ids 透传）
- [ ] **同步共享转换器 `rpc/internal/logic/notice/helper.go` 的 `toProtoNotice`（设计评审 data-model v4 S1——getnoticelogic/listnoticeslogic 共用，模型字段类型变更后 converter 编译即断）**：`CommunityId` 适配 `*int64`（null 感知，值由请求/scope 注入，Task 1.8/1.9 覆盖来源）；`PublishedAt` 用 `sql.NullTime` null 感知取值（NULL 不落 0 时间戳，SEE [[restore-compensation-zero-time]]——原 `n.PublishedAt.Unix()` 对 sql.NullTime 无此方法）
- [ ] 同步 `services/community-hub-service/docs/design.md` §数据模型（design_consistency 门禁，防 model 与标准迁移源漂移）
- **TDD**: 无逻辑代码可不写测试（模型字段调整由下游查询任务覆盖）；converter 适配由 `go build ./...` 编译门禁验证

### Task 1.4: NoticeModel 读查询重构（scope JOIN + 通过态谓词）
- **修改**: `services/community-hub-service/model/notice.go` + `notice_test.go`
- [ ] **新增共享谓词**：常量 `ModerationStatusMachinePass=1` / `ModerationStatusHumanPass=3` + **导出函数 `func IsModerationPassed(status int64) bool`**（`status==1 || status==3`，评审 S1——跨包复用必须导出，勿用未导出 `isModerationPassed`），供 rpc 层（Task 1.5/1.13）与读查询共用（评审 M1）
- [ ] `FindListByCommunity(ctx, communityId, role, offset, limit)` — `notices JOIN notice_scope ON notices.id=notice_scope.notice_id`，**显式投影 `notices.*` + `notice_scope.community_id`（右表限定/别名，防 `select *` 双 community_id 列按列序取到弃用 NULL，设计评审 data-model v4 S3）**，`scope.community_id=?` + **通过态 `moderation_status IN (1,3)`（IsModerationPassed）** + `deleted_at IS NULL` + 可选 role + `order by is_pinned desc, published_at desc` + 分页（复用既有 count 语义）
- [ ] **GetNotice 读路径收敛（评审 S2）**：**不新增 `FindOneByCommunity`**（防死代码/重复查询）——GetNotice 采用规范路径「`FindOnePublished(id)`（通过态 IN(1,3)，改造既有方法）+ `NoticeScopeModel` 匹配 `(id, community_id)`」，与 Task 1.8 对齐
- [ ] `FindMarquee(ctx, communityId, since, limit)` — JOIN scope、**通过态 IN (1,3)**、`published_at >= since`（含端点）、`order by is_pinned desc, published_at desc`、`limit 10`
- [ ] **RED**: table-driven 测试（正常 + 分页边界 + role 筛选 + scope 外小区查无 + **status=3 human_pass 出现在列表/详情/跑马灯** + status=2/4 过滤 + pending(published_at=NULL) 不进入倒序 + `IsModerationPassed(1/3)=true、(2/4/0)=false`）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- [ ] **REFACTOR**: 清理重复 SQL 拼装，保持测试绿
- **SEE**: [[moderation-status-write-without-read-gating]]（通过态谓词 IN(1,3)，评审 M1 修订）

### Task 1.5: NoticeModel 审核通过置 published_at + 撤回（scope 物理删）
- **修改**: `services/community-hub-service/model/notice.go` + `notice_test.go`
- [ ] `UpdateModerationStatusPass(ctx, id, status)` — **pass 判定用 `IsModerationPassed(status)`（1/3，评审 S1 导出符号）** → `set moderation_status=?, moderation_time=NOW(), published_at=NOW()`（pass 时置 published_at，D27/D30；status=3 human_pass 同样置 published_at，评审 M1）
- [ ] 保留 `UpdateModerationStatus(ctx, id, status)`（非 pass 分支：仅状态+时间）
- [ ] `Withdraw(ctx, id)` 语义由逻辑层组合：`SoftDelete` + `NoticeScopeModel.DeleteByNoticeId`（见 Task 1.12，本任务仅确保两个 model 方法可独立调用并同事务）
- [ ] **RED**: 测试 UpdateModerationStatusPass 置 published_at（status=1 与 status=3 两分支）+ 非 pass（2/4）不置 → FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.6: scope 包多目标校验 + division 展开 helper
- **修改**: `services/community-hub-service/rpc/internal/logic/scope/scope.go`、`services/community-hub-service/rpc/internal/logic/scope/userctx.go`
- **创建**: `services/community-hub-service/rpc/internal/logic/scope/scope_test.go`、`services/community-hub-service/rpc/internal/logic/scope/division.go`
- [ ] `AssertCommunitiesScope(ctx, client, userID, targets []int64) error` — **单次批量 `AssertPublishScope`**（一次携带全部 target ScopeRef，`AssertPublishScopeRequest.repeated targets`；任一目标越权/不可解析 → 060007→080006 一次映射整体返回 all-or-nothing；未知节点 fail-closed（D31）；**避免逐目标 N+1 RPC**，评审 SHOULD 3）
- [ ] `ExpandDivisionCommunities(ctx, mdClient, divisionID) ([]int64, error)` — **先 guard `divisionID<=0 → 080005`**（fail-closed，杜绝进入 masterdata 默认 FindAll 分支过度展开，评审 SHOULD 5）；调 masterdata `GetResidentialAreasByDivision(community_div_id=divisionID, status=1)`，返回审核通过小区 id 列表；RPC 传输错误原样返回
- [ ] **RED**: 测试（单目标越权→080006；多目标任一越权→整体 080006（一次批量调用语义）；目标不存在→080006 fail-closed；GLOBAL 放行；division 展开成功/展开空/**division<=0→080005**/传输错误）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[grpc-timeout-layers]]（AssertPublishScope 内嵌 master-data ResolveScopeAncestors，三层超时对齐）；[[rpc-identity-spoofing-loopback-isolation]]（user_id 经 metadata，不信任 body）

### Task 1.7: CreateNoticeLogic 重写（多小区 / division 展开 / 附件绑定 / 单事务）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/createnoticelogic.go` + `createnoticelogic_test.go`、`services/community-hub-service/rpc/internal/logic/notice/helper.go`（roleToString 保留）
- [ ] 校验顺序落地（design.md §CreateNotice）：title/content → scope 载体恰一（双载/双空/legacy 单 community_id 无 community_ids → 080005）→ community_ids 去重 & >100 → 080003 → division 展开（仅 community_admin；**division<=0 → 080005**；空展开 080005，>100 080003）
- [ ] 附件绑定（REQ-NP-6/REQ-AS-6）：逐 attachment_id（**自 RPC `attachment_ids` 透传，含 api 层 Task 1.15 修正**）调 file `GetFileUrl` → FileInfo `confirmed==true` 且 `user_id==JWT`；数量 ≤10；`Σ file_size ≤ 50MB` → 否则 080005；`file_type`/`file_name`/`file_size`/`download_url` 从 FileInfo 回读
- [ ] 数据权限：`AssertCommunitiesScope(user, targets)`（**单次批量**，Task 1.6）→ 080006（含目标级解析失败，D31）
- [ ] `publisher_id`/`role` 从 JWT 派生（**评审 S3 修订——JWT 仅含 user_id，实际角色必须显式调 permission `GetUserRoles(user_id)` 解析** → RBAC→NoticeRole 映射：grid_worker→GRID_OFFICER、community_admin→COMMUNITY、committee→COMMITTEE，**复用 Task 1.11 的判定/映射 helper**；多角色取授权 421 的发布角色优先，选择顺序 grid_worker>community_admin>committee），不信任请求体
- [ ] **`publisher` 取请求体展示字符串落库（NOT NULL 列，评审 S1）+ 非空校验：缺省 → 080005（评审 I3，防 INSERT NOT NULL 变 500）**
- [ ] 单事务落库：notices（community_id=NULL、published_at=NULL、**publisher=请求体展示字符串**、role=GetUserRoles 映射、publisher_id=JWT）+ NoticeScopeModel.InsertBatch + NoticeAttachment 批量 Insert（**file_url 写占位空串 `''`——file_id 为权威重生载体，不落预签名 URL 快照（防 VARCHAR(500) 截断 ERROR 1406，设计评审 data-model v4 S4）；file_name/file_size/file_type 自 FileInfo 回读、file_id=attachment_ids 透传**）
- [ ] 保持异步 moderation 入队（既有 Redis LPUSH 流程，不阻塞发布）
- [ ] **RED**: table-driven 测试覆盖正常成功 + 空范围 080005 + 双载 080005 + 非 admin 传 division 080005 + >100 080003 + publisher 空 080005 + 无发布角色（由功能层测试覆盖，RPC 层 mock 不设）→ 越权 080006 + 目标不存在 080006 + 附件未确认/他人文件 080005 + 附件超量/超总 080005 + GetUserRoles 解析角色落库 + JWT 伪造 role 被纠正 → FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[snake-camel-field-mismatch]]（TS 消费 community_ids）；[[grpc-only-comms]]（附件校验经 file GetFileUrl，不直连 uploaded_file）

### Task 1.8: GetNoticeLogic 重写（community_id 上下文 + scope 校验）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/getnoticelogic.go` + `getnoticelogic_test.go`
- [ ] `community_id` 必填：缺失/空 → 080005（D15）
- [ ] `FindOnePublished(id)` 未找到 → 080001；notice_scope 匹配 `(id, community_id)` 缺失 → 080001（scope 外不泄露）；`FilterAllowed(userID, community_id)` false → 080001
- [ ] 响应 `Notice.community_id` = 请求小区（由 notice_scope 派生，不读弃用列）；attachments 携带 `file_type`
- [ ] **附件 file_url 重生（评审 S4 + data-model v3 M1 闭环）**：详情读路径对每个附件按 `notice_attachments.file_id` 调 file `GetFileUrl(file_id)` 重生预签名 URL（file_type/file_name/file_size 同源回读）；**DB 中 `file_url` 为发布时短期预签名快照（3600s~7 天），勿当永久链接**——旧通知附件在详情页须可下载；**兼容期 file_id=0/NULL 的存量行回退返回 stored file_url（尽力而为，与 Q1 兼容期一致）**
- [ ] **RED**: 测试（正常 + 缺失 community_id 080005 + 不存在 080001 + scope 外 080001 + 未过审 080001 + 附件 file_type 返回 + **附件 file_url 为 GetFileUrl 重生后的新预签名 URL 断言**）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.9: ListNoticesLogic 改造（scope JOIN + community_id 派生）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/listnoticeslogic.go` + `listnoticeslogic_test.go`
- [ ] `FilterAllowed`（GLOBAL/LIMITED/EMPTY 语义保留）→ false 返回空列表
- [ ] `FindListByCommunity`（scope JOIN）+ `order by is_pinned desc, published_at desc`（审核锚定 D27/D32）+ 分页
- [ ] 响应每条 `community_id` = 请求小区；`role` 筛选语义保留
- [ ] **RED**: 测试（正常 + 分页边界 + role 筛选 + 越权读空列表 + 多小区通知只出现在其发布的小区 + **JOIN 投影断言：多小区通知在 C2 列表返回 community_id=C2（不取弃用 NULL 列，设计评审 data-model v4 S3）**）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.10: GetMarqueeNoticesLogic 新建
- **创建**: `services/community-hub-service/rpc/internal/logic/notice/getmarqueenoticeslogic.go` + `getmarqueenoticeslogic_test.go`
- [ ] `community_id` 必填：缺失/空/0 → 080005（D15）
- [ ] `FilterAllowed` → false 返回空列表（与列表读路径一致）
- [ ] `FindMarquee(community_id, now-15*24h, 10)`：置顶优先 + published_at 倒序 + 封顶 10 条 + 仅审核通过；15 天边界含端点（D32）
- [ ] 输出 `NoticeMarqueeItem[]{id, title}`
- [ ] **RED**: 测试（正常 ≤10 + 置顶优先 + 倒序 + 15 天边界含端点 + >15 天排除 + 未过审排除 + 空态 + 缺失 080005）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.11: GetPublishPermissionLogic 新建
- **创建**: `services/community-hub-service/rpc/internal/logic/notice/getpublishpermissionlogic.go` + `getpublishpermissionlogic_test.go`
- [ ] `userID = UserIDFromCtx`；0 → can_publish=false（防御，认证中间件兜底）
- [ ] 调 permission `GetUserRoles(user_id)`；level-2 判定：`role.Code ∈ {grid_worker, community_admin, committee}` 且 `status==2` 且 `verified_at>0` 且 `expires_at==0 OR expires_at>now`（基于 RPC 输出，禁止直读 rel_user_role，SEE [[grpc-only-comms]]）
- [ ] 命中 → can_publish=true + 映射 NoticeRole（grid_worker→GRID_OFFICER、community_admin→COMMUNITY、committee→COMMITTEE）
- [ ] property_admin / owner / tenant / merchant / sys_admin → can_publish=false（D16/D26）；sys_admin 写路径不额外拦截（design.md §GetPublishPermission）
- [ ] **RED**: 测试（网格员已认证通过 + status=2 但 verified_at NULL 拒绝 + 角色过期拒绝 + property_admin/owner 拒绝 + sys_admin 拒绝 + GetUserRoles 传输错误 fail-closed）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[auto-grant-unverified-grant-confers-scope-level0]]（min_verf_level=2 与 level-2 判定一致）

### Task 1.12: DeleteNoticeLogic 收窄（仅发布者本人 + scope 物理删 + 附件保留）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/deletenoticelogic.go` + `deletenoticelogic_test.go`
- [ ] `FindOne(id)` 未找到 → 080001
- [ ] 作者校验：`notice.PublisherId == JWT user_id`，否则 **080002**（收窄：原 CheckPublishScope 数据范围判定 → 仅发布者本人，行为回归，REQ-NP-5）
- [ ] **单事务（设计评审 data-model v4 S2 钉死）**：`conn.Transact(func(session sqlx.Session) error)` 传共享 session 给 `SoftDelete`（notices）+ `NoticeScopeModel.DeleteByNoticeId`（物理删）——model 方法签名改为接受 session/executor（或提供 Transact 变体），**禁止各自用独立 `m.conn.ExecCtx`**（中间失败会留半态：scope 已删而 notices 未软删 / 软删成功而 scope 孤儿行）；**不删 notice_attachments 行、不删 MinIO 对象**（D28）
- [ ] **RED**: 测试（发布者撤回全局生效 + 非发布者 080002 + 不存在 080001 + 附件保留断言 + **失败注入测试：第二阶段 DeleteByNoticeId 报错 → 整体回滚、notices 未软删、无孤儿 scope 行**）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.13: UpdateNoticeModerationStatusLogic（published_at 通过回调）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/updatemoderationstatuslogic.go` + `updatemoderationstatuslogic_test.go`
- [ ] `FindOne(id)` → 080001
- [ ] 系统身份 scope 校验适配：legacy 行用 `notice.CommunityId`（*int64 取值），新行用 `FindCommunityIdsByNoticeId` 首个小区 id 作为 target；无 scope 行跳过（系统身份 global 放行）
- [ ] **pass 判定复用 model 层 `IsModerationPassed(status)`（1=machine_pass / 3=human_pass，与读查询同一谓词，评审 M1）** → `UpdateModerationStatusPass`（置 published_at=NOW，D27/D30）；非 pass → `UpdateModerationStatus`
- [ ] **RED**: 测试（首次 pass=1 置 published_at + **pass=3 human_pass 置 published_at** + 编辑后重过审更新 published_at + fail/pending(2/4) 不置 + 回调目标不存在 080001 + **status=3 置 published_at 后经 FindListByCommunity/FindOnePublished/FindMarquee 可见**）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[rpc-callback-must-check-response-base]]（回调消费方检查响应 Base，不只看 gRPC err）；[[moderation-status-write-without-read-gating]]（读路径恒过滤通过态 IN(1,3)，与回调 pass 判定同一谓词，评审 M1）

### Task 1.14: RPC server 注册新方法
- **修改**: `services/community-hub-service/rpc/internal/server/communityhubserver.go`
- [ ] `GetPublishPermission` / `GetMarqueeNotices` 方法委托对应 Logic
- [ ] `go build ./...` 通过

### Task 1.15: API 层类型同步（types.go + attachment_ids 透传）
- **修改**: `services/community-hub-service/api/internal/types/types.go` + `services/community-hub-service/api/internal/logic/notice/createnoticelogic.go` + 相关测试
- [ ] CreateNoticeRequest 增 **`CommunityIds []string \`json:"community_ids"\``**（**REST 层 string 形式 Snowflake ID，不能 `[]int64`**——encoding/json `,string` 不支持 slice，移动端发 `["1001","1002"]` 解 `[]int64` 会 `cannot unmarshal string into ... of type int64`，评审 MUST 1）；`DivisionId int64 \`json:"division_id,string"\``（**标量**才支持 `,string`，SEE [[proto-jstype]]）
- [ ] **新增 `AttachmentIds []string \`json:"attachment_ids"\``（现 types.go 无此字段，评审 MUST 2）**
- [ ] NoticeAttachment 增 `FileType string \`json:"file_type"\``
- [ ] **GetNoticeRequest 增 `CommunityId int64 \`form:"community_id"\``（评审 MUST 1 修订——GET query 绑定必须用 `form` 标签，仓库惯例同 `ListNoticesReq form:"community_id"`；`id` 保持 `path:"id"`）**。**禁用 `json` 标签**：go-zero httpx.Parse 对 GET 走 ParseForm（formUnmarshaler），`unmarshaler.go:571 usingDifferentKeys("form", field)` 对无 form 标签字段直接 skip → `json` 标签不生效 → `CommunityId` 恒 0 → 每次详情调用 080005 全挂。可用 `form:"community_id,optional"` + **api 逻辑显式判空 → 080005**
- [ ] API create 逻辑代理：`community_ids` 逐个 `strconv.ParseInt` 转 int64 → RPC `community_ids`；`division_id` 透传；**`attachment_ids` 透传到 RPC `attachment_ids`（修复现丢弃）**；GetNotice 代理透传 community_id 上下文（**RPC 入参 `community_id` 填 req.CommunityId（form 绑定）**）
- [ ] **RED**: api 层测试（`community_ids=["1001","1002"]` 解码并转 int64 成功；`attachment_ids` 透传 RPC 参数断言；非数字 community_id 解析失败 → 080005；**GET `/notices/:id?community_id=456` → 断言 req.CommunityId 被绑定 =456（form 标签生效）**；**GET `/notices/:id` 缺 community_id → 080005**）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- [ ] 现有 api 逻辑测试随字段同步（snake_case 对齐，SEE [[snake-camel-field-mismatch]]）

### Task 1.16: API 新端点 handler + logic 代理（marquee / publish-permission）
- **创建**: `services/community-hub-service/api/internal/handler/notice/getmarqueenoticeshandler.go`、`services/community-hub-service/api/internal/logic/notice/getmarqueenoticeslogic.go`、`services/community-hub-service/api/internal/handler/notice/getpublishpermissionhandler.go`、`services/community-hub-service/api/internal/logic/notice/getpublishpermissionlogic.go`
- **修改**: `services/community-hub-service/api/internal/types/types.go`（新增请求/响应类型，评审 I1 防遗漏）
- [ ] **types.go 新增**：`GetMarqueeNoticesReq { CommunityId int64 \`form:"community_id"\` }`（**form 标签，GET query 绑定，评审 MUST 1——勿用 json/path，同 GetNotice**）+ `GetMarqueeNoticesResp { Items []NoticeMarqueeItemInfo \`json:"items"\` }`；`GetPublishPermissionReq`（空）/ `GetPublishPermissionResp { CanPublish bool \`json:"can_publish"\`; PublishableRoles []int32 \`json:"publishable_roles"\` }`（缺失类型不可自行命名，与 proto 契约对齐）
- [ ] 各 handler 解析 JWT user_id + 注入出站 metadata（`util.WithUserID`，复用既有模式）
- [ ] marquee logic：**community_id 经 form 绑定（req.CommunityId）** → 代理 RPC → 返回 items；publish-permission logic：代理 RPC → 返回 can_publish + roles
- [ ] API 层测试：**GET `/notices/marquee?community_id=456` → 断言 req.CommunityId 被绑定 =456**；community_id 缺失 → 080005；正常返回
- **SEE**: [[snake-camel-field-mismatch]]（REST JSON 字段 snake_case）

### Task 1.17: routes.go 注册新 REST 端点
- **修改**: `services/community-hub-service/api/internal/handler/routes.go`
- [ ] 注册 `GET /api/community/notices/marquee`、`GET /api/community/notices/publish-permission`（**静态路径先于 `:id` 注册**，避免被 `:id` 通配吞掉）
- [ ] 置于 PermMiddleware 中间件组内（与既有 notice 路由一致）
- [ ] `go build ./...` + 路由冒烟（curl 两个端点确认不被 `:id` 抢占）

---

## file-service

### Task 2.1: 错误码 70004/70005 登记
- **修改**: `services/file-service/rpc/internal/errx/errcode.go`、`services/file-service/rpc/internal/logic/file/errcode.go`、`services/file-service/api/internal/logic/errcode.go`、`services/file-service/api/internal/logic/file/errcode.go`
- [ ] 新增 `ErrCodeUnsupportedFileType = 70004`（070004 不支持的文件类型）
- [ ] 新增 `ErrCodeFileSizeExceeded = 70005`（070005 文件大小超限）
- [ ] **70001-70003 不重编号**（ErrCodeFileOperationFailed 保持 70003，D11）
- **SEE**: [[error-code-collision-and-namespace-alignment]]（同整数双语义禁止）；[[error-code-literal-bypasses-qa-gate]]（用命名常量，禁裸数字）

### Task 2.2: 白名单/禁止集/10MB + GetUploadUrl L1 快速拒绝
- **创建**: `services/file-service/internal/guard/whitelist.go` + `whitelist_test.go`（通用校验器沉淀，供 L1/L2 共用）
- **修改**: `services/file-service/rpc/internal/logic/file/getuploadurllogic.go` + `getuploadurllogic_test.go`
- [ ] 常量：白名单 {png,jpg,jpeg,gif,pdf,doc,docx}、禁止集 {exe,bat,sh,cmd,com,msi,apk,js,vbs,ps1,py,pl,php}、`MaxSingleFileSize = 10MB`
- [ ] `ValidateFileName(fileName string) (ext string, err error)` — 扩展名大小写不敏感；无扩展名/点文件/非白名单 → ErrCodeUnsupportedFileType；禁止集/zip/rar → 同码
- [ ] `ValidateFileSize(size int64) error` — >10MB → ErrCodeFileSizeExceeded；=10MB 放行
- [ ] GetUploadUrl 在生成预签名 URL 前调 `ValidateFileName(in.FileName)` + `ValidateFileSize(in.FileSize)` 快速拒绝
- [ ] entity_type 覆盖：全局基线 + 可细化（本 Task 仅落全局基线，覆盖机制见 Task 2.6）
- [ ] **RED**: 测试（白名单放行 + exe/sh/js 拒绝 + zip/rar 拒绝 + 大小写绕过无效 + 无扩展名/点文件拒绝 + 12MB 拒绝 + 恰 10MB 放行）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[migration-must-execute]] 不适用（本任务无迁移）

### Task 2.3: magic-bytes 内容嗅探器（doc/docx/pdf/图片 + 容器子类型拒绝）
- **创建**: `services/file-service/internal/guard/magic.go` + `magic_test.go`
- [ ] `SniffType(buf []byte) (string, bool)` 返回规范扩展名 + 是否识别
- [ ] 规则：png `89 50 4E 47`、jpg `FF D8 FF`、gif `47 49 46 38`、pdf `%PDF`（读文件头字节）
- [ ] **doc = OLE2/CFB（D0 CF 11 E0 A1 B1 1A E1）且内含 `WordDocument` 流** → doc（仅 CFB 头不充分；`.msi/.xls/.ppt` 无 WordDocument 流 → 不映射 → 拒绝）
- [ ] **docx = ZIP（PK 头）+ 含 `word/document.xml` 部件** → docx（仅 `[Content_Types].xml` 不充分；`.xlsx/.pptx` → 不映射 → 拒绝）
- [ ] 其他 OLE2 子类型（msi/xls/ppt）与其他 OOXML 子类型（xlsx/pptx）与通用 zip/rar → 返回 (not recognized)（由调用方 070004）
- [ ] **RED**: table-driven 测试（真实 doc 放行 + msi 改 doc 拒绝 + 真实 docx 放行 + xlsx 改 docx 拒绝 + 通用 zip 拒绝 + 改名 png 传 exe（PE MZ）拒绝）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[migration-must-execute]] 不适用；文档见 design.md §附件安全（D18）

### Task 2.4: ConfirmUpload L2 回读校验 + file_type/confirmed 落库
- **修改**: `services/file-service/rpc/internal/logic/file/confirmuploadlogic.go` + `confirmuploadlogic_test.go`、`services/file-service/rpc/internal/logic/file/helper.go`
- [ ] 回读 MinIO 实际对象（`RawMinio` 或 `MinioCli` GetObject）前若干字节 → `SniffType` → 映射规范扩展名
- [ ] 嗅探类型与声明扩展名一致才放行；不一致 → 070004（类型主码；大小超限为次码）
- [ ] File 记录写 `FileType`（嗅探映射）+ `Confirmed=true`
- [ ] `toProtoFile` 填充 `FileType`/`Confirmed`
- [ ] **RED**: 测试（正常确认成功 + 声明 png 实为 exe 拒绝 070004 + 声明与魔数一致放行 + FileInfo 返回 file_type/confirmed）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 2.5: File 模型扩展 + Migration 002（file_type/confirmed 列）
- **修改**: `services/file-service/model/file.go`、`services/file-service/model/filemodel.go`
- **创建**: `services/file-service/migration/002_file_guard.sql`
- **粒度说明（评审 INFO 4）**: Migration 002 与 File 模型扩展合并同一任务——迁移即模型列的 DB 载体（同一列变更的落库端，GORM 非 AutoMigrate 须显式迁移），独立运维验证由 Task 5.1（三态库执行）兜底
- [ ] `File` struct 增 `FileType string`、`Confirmed bool`（GORM 非 AutoMigrate，须显式迁移）
- [ ] `002_file_guard.sql`：**首行 `USE file_db;`（与 001 一致，设计评审 data-model v4 INFO I3——001 用 file_db，迁移必须带 DB 上下文）**；`ALTER TABLE uploaded_file ADD COLUMN file_type VARCHAR(20) DEFAULT NULL, ADD COLUMN confirmed TINYINT NOT NULL DEFAULT 1`
- [ ] `FileModel.Insert` 包含 file_type/confirmed 列
- [ ] `toProtoFile` 已在 Task 2.4 覆盖；此处补 model 层测试（Insert 后读回字段）
- **SEE**: [[migration-must-execute]] — 提交后必须执行；[[snake-camel-field-mismatch]]（TS 消费 file_type）

### Task 2.6: entity_type 覆盖机制（全局基线不弱化）
- **修改**: `services/file-service/internal/guard/whitelist.go`（覆盖注册）+ `whitelist_test.go`、`services/file-service/rpc/internal/logic/file/getuploadurllogic.go`
- [ ] `RegisterEntityTypeOverride(entityType string, cfg Override)` — 追加允许类型；**禁止集不可放行、10MB 硬上限不可放宽**（REQ-AS-4 不变量）
- [ ] GetUploadUrl 按 `in.EntityType` 查覆盖：有 → 基线上叠加；无 → 全局基线
- [ ] **RED**: 测试（基线默认生效 + override 追加合法类型 + override 试图放宽 10MB/放行 exe 被拒 070005/070004 + 既有 avatar/verification/lostfound/contacts 上传不回归）→ FAIL
- [ ] **GREEN**: 实现 → PASS

---

## permission-service

### Task 3.1: AssertPublishScope 判据扩展（division→community 授权落位，design gate）
- **修改**: `services/permission-service/rpc/internal/logic/permission/assertpublishscopelogic.go`、`services/permission-service/rpc/internal/logic/permission/scope.go`
- **创建**: `services/permission-service/rpc/internal/logic/permission/assertpublishscope_division_test.go`
- [ ] 新增 `resolvePublishScope(ctx, urm, userId)` 变体：收集 `community` ∪ `community_div` 双 scope_type 授权并集（供 targetCovered 使用）；`resolveUserScope` 保持（GetDataScopes 读路径不动）
- [ ] `AssertPublishScopeLogic.AssertPublishScope` 改用 `resolvePublishScope`；`targetCovered` 逻辑不变（祖先链 ∩ ids）
- [ ] **design gate 验证门禁（编码前必须通过）**：单测/集成验收——
  1. community_admin 持 `community_div=D1` grant 发布 C1（`community_div_id=D1`，祖先链含 D1）→ allowed
  2. 发布 D1 外小区 C2 → 060007 denied
  3. 目标小区不存在（ResolveScopeAncestors found=false）→ denied（安全拒绝未知节点）
  4. community 单 scope grant（grid_worker）发布行为不回归（原语义保持）
  5. **共享调用方回归（评审 SHOULD 2——AssertPublishScope 是共享 RPC，还被 lostfound 创建 createlostfoundlogic.go、contacts upsert upsertcontactslogic.go 调用）**：community_admin 持 `community_div=D1` 时（a）lostfound 发布到 C1（祖先链含 D1）→ allowed、（b）contacts upsert 到 C1 → allowed（如产品语义要求隔离则显式声明并在 resolvePublishScope 内按调用方隔离）；见 design §Design Gate「共享 blast radius 声明」
- [ ] **RED**: 以上测试先写并确认 FAIL（现状判据拒绝 division；lostfound/contacts 调用方同路径回归）→ **GREEN**: 实现 → PASS
- **SEE**: [[is-system-no-permission-shortcut]]（global 无字段短路）；[[grpc-timeout-layers]]（AssertPublishScope 内嵌 ResolveScopeAncestors 超时对齐）；[[tdd-red-evidence-requires-fail-excerpt]]（RED 证据含实际 FAIL 摘录）

### Task 3.2: 权限种子变更 REQ-PP-4 + 读/写权限矩阵补齐（评审 interface-proto v4 MUST 1/2 + SHOULD 3 + INFO 5/6）
- **修改**: `services/permission-service/scripts/init_permissions.sql` + `docs/specs/rbac-design.md`
- **写路径**:
  - [ ] grid_worker(4) 授 421：`INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES (4, 421)`
  - [ ] 421 置 `min_verf_level=2`：`UPDATE sys_permission SET min_verf_level = 2 WHERE code = 'community:notice:create-api'`（原 §4.2 置 0）
  - [ ] **回收 421（显式 DELETE，INSERT IGNORE 无法撤销，SEE [[insert-ignore-swallows-errors]]）**：property_admin(2)、owner(1)、tenant(5) 各 `DELETE FROM rel_role_permission WHERE role_id=? AND permission_id=421`
  - [ ] **新增 `427` `DELETE:/api/community/notices/:id`（code `community:notice:delete-api`，parent_id=420）——撤回端点补权限码（评审 MUST 2：现无任何码，fail-closed 下全体 403、DeleteNotice 080002 作者校验走不到）**
  - [ ] **新增 `428` `PUT:/api/community/notices/:id`（code `community:notice:update-api`，parent_id=420）——编辑端点补权限码（评审 MUST 2，UpdateNotice 回归测试可达）**
- **读路径**:
  - [ ] **422 `GET:/api/community/notices` 扩展绑定全部移动端社区角色**（现仅 (9,1,5)；补 grid_worker 4 / committee 6 / merchant 7 / community_admin 3 / sys_admin 8，评审 MUST 1——发布角色经 marquee「更多→浏览」进入浏览页不得 403）
  - [ ] 新增读接口权限码（path 与实际 REST 路由一致，SEE [[permission-seed-api-path-must-match-routes]]）：`423 GET:/api/community/notices/marquee`（community:notice:read-marquee-api，**parent_id=410**）、`424 GET:/api/community/notices/publish-permission`（community:notice:publish-permission-api，**parent_id=410**）、**`426 GET:/api/community/notices/:id`（community:notice:read-detail-api，**parent_id=410**）——详情端点补权限码（评审 MUST 1：现无任何码 → 全体 403，GetNotice 080001/080005 契约不可达）**、**`425 GET:/api/perm/data-scopes`（`community:data-scopes:read-api`，**parent_id=310**——评审 MUST 3 + SHOULD 3：data-scopes 端点鉴权钉死不复用 422；parent_id 显式指定 310，AutoDiscover 对 /api/perm/* 会 miss 落到 /data-scopes 菜单，勿依赖自动发现，见 Task 3.3）**
  - [ ] **423/424/425/426** 绑定全部移动端社区角色（registered_user 9 / owner 1 / tenant 5 / grid_worker 4 / committee 6 / merchant 7 / community_admin 3 / sys_admin 8 全量）；**427/428 同批绑定全部移动端社区角色**（真正越权判定交业务层 080002 作者校验 / 通知级语义）
- **权限码 parent_id 汇总（评审 SHOULD 3，防孤儿节点）**: 423/424/426 → 410（community:read，同 422）；427/428 → 420（community:notice，同 421）；425 → 310（permission:read，path 在 /api/perm/*）
- **幻影 435 措辞（评审 INFO 5）**: 改为「`community:lostfound:create-api` **无 sys_permission 行**，保持现状不动」（原措辞「435 保持 min_verf_level=0 不动」易误读为存在 id=435 行——磁盘核实种子无 INSERT 定义，仅 line 202 UPDATE 引用 + rel_role_permission (1,435)/(5,435) 指向幻影）；436（contact:upsert）有真实行保持不动；本变更不动 owner/tenant 的 435/436
- **rbac-design.md §6.5 矩阵补登（评审 INFO 6）**: `docs/specs/rbac-design.md` §6.5 角色验收矩阵新增通知读/写权限行——422/423/424/426（读）+ 427/428（写）+ 421 发布角色集变更（grid_worker 授、property_admin/owner/tenant 回收），与 §权限种子矩阵一致
- [ ] 种子末尾验证查询阈值随回收/新增更新（property_admin perm_count、owner 等——回收 421 后下降，新增读码后 owner/tenant/grid_worker/committee 等上升；阈值 `>=` 保底仍过，但断言须精确到具体码）
- [ ] 幂等：整段可重复执行（guard + INSERT IGNORE + 幂等 DELETE）
- [ ] `services/permission-service/scripts/init_permissions.sql` 提交后按部署编排执行（SEE [[cross-service-seed-deployment-order]]）
- **SEE**: [[is-system-no-permission-shortcut]]（sys_admin 全权限经 rel_role_permission 配置，非短路）；[[auto-grant-unverified-grant-confers-scope-level0]]（min_verf_level=2 收窄未认证发布）；[[permission-seed-api-path-must-match-routes]]

### Task 3.3: 权限服务 API 暴露 GetDataScopes（移动端范围选项数据源）
- **创建**: `services/permission-service/api/internal/handler/getdatascopeshandler.go`、`services/permission-service/api/internal/logic/getdatascopeslogic.go`
- **修改**: `services/permission-service/api/internal/handler/routes.go`、`services/permission-service/api/internal/handler/handlers.go`
- [ ] `GET /api/perm/data-scopes?scope_type=community`：解析 JWT user_id → 代理 RPC `GetDataScopes` → 返回 `scope_ids` + `state`
- [ ] 仅接受 `scope_type=community`（不返回 community_div，REQ-NM-5 注：division 选项经 masterdata 树，D17）
- [ ] **鉴权（钉死，评审 MUST 3）**：绑定权限码 **`425 GET:/api/perm/data-scopes`**（`community:data-scopes:read-api`，**parent_id=310**，Task 3.2 已登记 + 绑定全量移动端角色）——**去掉「或经既有 read 权限放行」二义措辞**；**PermMiddleware 归属 = permission-service**（/api/perm 全部路由统一走 `serverCtx.PermMiddleware.Handle`，routes.go 末尾已包裹，本端点沿用同一中间件组）
- [ ] routes.go 注册（含 RouteRegistry 登记 `GET /api/perm/data-scopes`）+ handlers 接线；`go build ./...`
- [ ] API 测试：**验收「grid_worker 调 data-scopes 返回其 community scope_ids」**；正常返回 scope_ids；scope_type 非法 → 错误；未登录 → 认证中间件拒绝
- **说明**: 该端点供移动端发布表单网格员多选小区选项（REQ-NM-5「as today」数据源落地）；返回调用者自身 scope_ids（self-scope 读），绑定全量移动端社区角色安全且覆盖发布者（grid_worker/community_admin/committee），不复用 422（其绑定集排除发布者）

---

## master-data-service

> 只读复用，无代码变更。division→小区展开（`GetResidentialAreasByDivision` community_div_id>0 + status=1）与 `ResolveScopeAncestors` 已存在；验证见 Task 5.1/5.2（Owner 运维验证）。

---

## moderation-service

> 只读复用，无代码变更。UpdateNoticeModerationStatus 回调行为（pass 置 published_at）在 community-hub 侧实现（Task 1.13）；moderation-service 本身不改。

---

## web/mobile（前端，Q6 仅移动端）

### Task 4.1: API client 扩展（community.ts + 类型）
- **修改**: `web/mobile/src/api/community.ts`
- [ ] `NoticeAttachment` 增 `file_type: string`
- [ ] **新增 `createNotice(...)` client（评审 INFO 3——`web/mobile/src/api/community.ts` 当前无 createNotice 函数，仅 getNoticeList/getNoticeDetail，属新增非扩展）**：POST `/api/community/notices`，参数 `community_ids: string[]` / `division_id: string` / `attachment_ids: string[]`（int64 用 string，防 JS 精度丢失，SEE [[snake-camel-field-mismatch]]）；**division_id 不适用时省略字段、不发送空串（评审 SHOULD 4——`encoding/json ,string` 解 `division_id:""` 空串在 REST 层 4xx，进不到 RPC 080005）**
- [ ] 新增 `getPublishPermission()` → `GET /api/community/notices/publish-permission`（返回 can_publish + roles）
- [ ] 新增 `getMarqueeNotices(communityId)` → `GET /api/community/notices/marquee?community_id=xxx`（返回 NoticeMarqueeItem[]：id/title）
- [ ] 新增文件上传 client：`getUploadUrl` → `POST /api/files/upload-url`、`confirmUpload` → `POST /api/files/confirm`（复用既有 file-service 契约）
- [ ] 新增数据范围选项：`getDataScopes()` → `GET /api/perm/data-scopes?scope_type=community`；division 选项复用 `web/mobile/src/api/user.ts` 的 `getDivisions`/division 树
- [ ] 类型不重复定义：优先复用 `web/common` 共享层（SEE [[web-common-type-reuse-no-redefine]]）

### Task 4.2: 附件前端一致预校验（共享校验器）
- **创建**: `web/mobile/src/utils/attachmentGuard.ts` + `attachmentGuard.spec.ts`
- [ ] `validateAttachment(file): { ok, reason }` — 白名单扩展名（大小写不敏感）+ 单文件 ≤10MB（REQ-AS-1/AS-2 镜像）
- [ ] `validateNoticeAttachments(files): { ok, reason }` — ≤10 个 且 总大小 ≤50MB（REQ-AS-6 镜像，D23）
- [ ] **不弱化后端不变量**：前端预校验仅 UX，最终安全边界在后端（REQ-NM-6 场景）
- [ ] 组件测试：exe 拒绝、zip 拒绝、12MB 拒绝、第 11 个附件拒绝、总超 50MB 拒绝、恰边界放行
- **SEE**: [[frontend-business-rule-hardcode]]（校验区间与后端/proto 对齐，避免漂移）

### Task 4.3: NoticeMarquee 组件 + 首页接入
- **创建**: `web/mobile/src/components/NoticeMarquee.vue`
- **修改**: `web/mobile/src/pages/notice/notice.vue`
- [ ] 组件消费 `getMarqueeNotices(currentCommunityId)`；渲染通栏滚动标题；「更多 ›」→ 浏览页；点击标题 → 详情页；空数据占位不报错（REQ-NM-1）
- [ ] 切换小区后重新拉取（随 `communityStore.currentCommunityId` 响应式）
- [ ] 首页用 NoticeMarquee 替换现有内联跑马灯；15 天窗口/置顶/封顶 10 由后端契约保证，前端不再 re-derive（REQ-NM-1）
- [ ] 组件自持数据拉取，页面不重复实现（REQ-NM-7）

### Task 4.4: NoticeList 组件 + 浏览页改造
- **创建**: `web/mobile/src/components/NoticeList.vue`
- **修改**: `web/mobile/src/pages/notice-browse/notice-browse.vue`
- [ ] 组件按 `published_at` DESC 渲染列表 + 分页加载（滚动到底追加、耗尽停止，REQ-NM-2）
- [ ] 数据源与 REQ-NR-1 一致（ListNotices scope 过滤）；点击 → 详情
- [ ] 组件自持数据拉取；空态/加载失败降级不破坏宿主页（REQ-NM-7）

### Task 4.5: NoticeDetail 组件 + 详情页改造
- **创建**: `web/mobile/src/components/NoticeDetail.vue`
- **修改**: `web/mobile/src/pages/notice-detail/notice-detail.vue`
- [ ] 展示 title/content/published_at（审核通过日，D27）/附件列表（file_name/file_url/file_size/**file_type**，可点击打开/下载，REQ-NM-3）
- [ ] 请求携带当前小区 community_id 上下文；080001 → 空态（不泄露跨小区/未审内容）
- [ ] 组件自持数据拉取；`file_type` 展示（pdf/doc/docx/图片标识）

### Task 4.6: NoticePublisher 组件 + 发布页
- **创建**: `web/mobile/src/components/NoticePublisher.vue`、`web/mobile/src/pages/notice-publish/notice-publish.vue`
- **修改**: `web/mobile/src/pages.json`（注册发布页路由）
- [ ] 表单：标题/正文/附件（经 Task 4.2 校验 + 上传）/范围选择
- [ ] 范围选择：grid_worker 多选小区（`getDataScopes()`，REQ-NM-5）；community_admin 选 division（提交 division_id，后端展开，选项经 masterdata division 树）；committee 固定本小区（不可编辑，community_ids=[本小区]）；property_admin 无入口（D26，由 mine 页 Task 4.7 兜底）；**非 community_admin 不发送 division_id 字段（省略，勿传空串，评审 SHOULD 4）**
- [ ] **提交中禁用防双击**（D25，REQ-NM-5/REQ-NP-7）：in-flight 置 disabled，忽略二次点击，仅发一次 CreateNotice
- [ ] 空标题/正文前端校验拦截
- [ ] 成功 → 成功 toast + 离开发布页；失败 → 错误提示（080003/080005/080006 等文案映射）
- [ ] 附件上传失败/超限在前端预校验拦截，不发起 CreateNotice（REQ-NM-6）
- **SEE**: [[async-submit-double-guard]]（异步提交防重复）

### Task 4.7: 【我的】页发布入口（can_publish 驱动）
- **修改**: `web/mobile/src/pages/mine/mine.vue`
- [ ] 拉取 `getPublishPermission()`；`can_publish==true` → 渲染「发布通知」入口 → 导航发布页；false → 不渲染（REQ-NM-4，前端**不做**角色码判断，REQ-PP-2）
- [ ] 未登录/接口失败 → 按 can_publish=false 处理（入口隐藏，单一行为）
- [ ] 组件测试：can_publish true/false 两分支渲染

---

## Owner 运维验证（不走 harness-pipeline，编码后由 Owner 单独执行）

### Task 5.1: 三态库迁移验证
- [ ] community-hub-service `003_multi_community_notice.sql`：新库/旧库升级/生产 三态执行，验证 `notices.community_id`/`published_at` 可空（DESC 确认 DEFAULT NULL）、`notice_scope` 表 + 双索引存在、`notice_attachments.file_type`/`file_id` 列存在 + **`file_url` 已加宽 VARCHAR(1024)（data-model v4 S4）**
- [ ] file-service `002_file_guard.sql`：`uploaded_file` 增 `file_type`/`confirmed` 列
- [ ] permission-service `init_permissions.sql` 重跑幂等（评审 MUST 1/2 + SHOULD 3 断言）：
  - grid_worker(4) 持 421、421 `min_verf_level=2`、property_admin(2)/owner(1)/tenant(5) **无** 421
  - **426 `GET:/api/community/notices/:id` 存在且绑定全部移动端角色**（owner/community_admin/grid_worker/tenant/committee/merchant/sys_admin/registered_user 全有）
  - **422 扩展后 grid_worker(4)/community_admin(3)/committee(6)/merchant(7)/sys_admin(8) 均持 422**
  - **427 `DELETE:/api/community/notices/:id`、428 `PUT:/api/community/notices/:id` 存在且绑定全部移动端角色**
  - 423/424/425 绑定就绪；**425 parent_id=310、426 parent_id=410、427/428 parent_id=420（SHOULD 3，无孤儿节点）**
  - **幻影 435 不新增 sys_permission 行**（断言不与 (1,435)/(5,435) 幻影绑定冲突）
- **SEE**: [[migration-must-execute]]、[[cross-service-seed-deployment-order]]

### Task 5.2: 部署顺序 + 端到端冒烟
- [ ] 迁移先于功能上线：未迁移时多小区发布 INSERT 被 NOT NULL 拒绝（REQ-NP-1 异常门禁）→ 迁移后成功
- [ ] 端到端冒烟（真实 DB）：网格员多小区发布成功 + 跑马灯返回 + 详情/列表 scope 过滤 + 附件白名单/magic-bytes 拒绝 + 撤回仅发布者本人
- [ ] **权限矩阵冒烟（评审 MUST 1/2）**：全部移动端角色（owner/grid_worker/community_admin/tenant/committee/merchant/registered_user 等）调 `GET /notices/:id`（详情）与 `GET /notices`（浏览列表）**不 403**；grid_worker 经 marquee「更多→浏览」进入浏览页不 403；**撤回 DELETE /notices/:id 可达**（发布者本人成功全局撤回；非发布者经中间件放行后由 080002 作者校验拒绝——验证 427 权限码让 REST 层放行、业务层 080002 兜底）
- [ ] **通过态一致性冒烟（评审 M1）**：DB 层模拟一条 `moderation_status=3`（human_pass）+ `published_at` 已置的通知，确认其出现在列表/详情/跑马灯（与 status=1 同等可见）；status=2/4 不出现在任何读路径
- [ ] 确认 design_consistency：`bash .harness/scripts/check-design-consistency.sh --all` 无本变更引入的 WARN

---

## Tasks Self-Review 记录

- **占位符扫描**: 无 `<任务描述>`/TBD/TODO；全部精确到文件路径 ✅
- **TDD 覆盖**: 所有含逻辑任务（1.2/1.4/1.5/1.6/1.7/1.8/1.9/1.10/1.11/1.12/1.13/1.15/2.2/2.3/2.4/2.6/3.1/4.2/4.7）含 RED→GREEN（前端组件测试策略按项目惯例）；Migration（1.1/2.5）与 Proto（0.x）无逻辑代码不写 TDD，但含验证门禁 ✅
- **依赖顺序**: Proto(0) → 迁移/模型(1.1-1.5) → scope 基础设施(1.6) → 写/读逻辑(1.7-1.13) → 注册/API(1.14-1.17) → 前端(4) → 运维验证(5) ✅
- **独立可测**: 每 Task 文件数 ≤3、单代码层级、可独立编译测试 ✅
- **记忆引用检查**: 20 个相关记忆，全部注入 design.md 且高风险 Task（迁移/种子/Design Gate/附件安全/防重）已标注 `SEE:` 引用；`[[unique-index-migration-dup-precheck]]`/`[[notfound-cache-sentinel-vs-transient-error]]`/`[[redis-cache-soft-delete]]` 主动排除并记理由 ✅
- **设计评审修订回执（本轮 MUST FIX 逐条落地）**:
  - 评审 data-model M1 → Task 1.4（isModerationPassed 共享谓词 + IN(1,3) 读查询 + status=3 测试）、Task 1.5/1.13（pass 判定复用谓词）、Task 5.2（status=3 可见性冒烟）、design §可见性门禁
  - 评审 interface-proto MUST 1 → Task 1.15 `CommunityIds []string` + strconv 代理 + 修正 repeated `,string` 误导
  - 评审 interface-proto MUST 2 → Task 1.15 `AttachmentIds []string` + 透传 RPC + api 层测试
  - 评审 interface-proto MUST 3 → Task 3.2 `425 GET:/api/perm/data-scopes` + Task 3.3 去二义措辞/钉死 PermMiddleware + 验收 grid_worker scope_ids
  - SHOULD 全落地：publisher 来源显式（Task 1.7）、批量 AssertPublishScope（Task 1.6）、division<=0 guard（Task 1.6）、role/publisher_id deprecated（Task 0.1）、CHANGELOG 语义破坏登记（Task 0.4）、重审 published_at 副作用（design ADR）、confirmed 存量注记（design §ConfirmUpload）✅
- **设计评审修订回执（本轮 interface-proto v2 + data-model v2 逐条落地）**:
  - 评审 interface-proto v2 🔴 MUST 1（GetNotice community_id REST 绑定）→ Task 1.15 `GetNoticeReq.CommunityId` 改 **`form:"community_id"`**（禁 json，GET query 走 ParseForm、json 标签被 skip → 恒 0 → 080005 全挂）+ RED 断言 `?community_id=456` 绑定；Task 1.16 marquee `GetMarqueeNoticesReq.CommunityId` 同用 form 标签 + 绑定断言；design §GetNotice/§GetMarqueeNotices 补 REST 绑定说明 ✅
  - 评审 interface-proto v2 🟡 SHOULD 2（AssertPublishScope 共享 blast radius）→ Task 3.1 验证门禁追加 lostfound/contacts 调用方回归（community_div=D1 时两写路径 allowed）；design §Design Gate 新增「共享 blast radius 声明」+ ADR 行 ✅
  - 评审 interface-proto v2 🔵 INFO 3 → Task 4.1 改「新增 createNotice client」（community.ts 现无此函数，非扩既有）✅
  - 评审 interface-proto v2 🔵 INFO 4 → Task 2.5 补「粒度说明」理由（迁移即模型列 DB 载体，Task 5.1 独立运维验证兜底）✅
  - 评审 data-model v2 🟡 S1（isModerationPassed 未导出）→ Task 1.4/1.5/1.13 + design §可见性门禁 统一命名 **`IsModerationPassed`**（导出）✅
  - 评审 data-model v2 🟡 S2（FindOneByCommunity 冗余）→ Task 1.4 删除 FindOneByCommunity，收敛 GetNotice 为 `FindOnePublished` + notice_scope 匹配 ✅
  - 评审 data-model v2 🟡 S3（role 派生须 GetUserRoles）→ Task 1.7 + design §CreateNotice 显式补 GetUserRoles 调用 + 多角色选择顺序 ✅
  - 评审 data-model v2 🟡 S4（file_url 短期预签名）→ Task 1.8 附件 file_url 按 file_id 重生 + design §GetNotice 语义声明 + Task 0.4 CHANGELOG 登记 ✅
  - 评审 data-model v2 🔵 I1/I2/I3 → Task 1.16 补 types.go 请求类型清单；design notice_scope 迁移注释注明纯关联表物理删除语义；Task 1.7 publisher 非空校验 080005 ✅
- **设计评审修订回执（本轮 interface-proto v4 + data-model v4 逐条落地）**:
  - 评审 interface-proto v4 🔴 MUST 1（通知读路径权限矩阵不完整：详情 426 缺失、422 仅绑 (9,1,5)）→ **design 新增 §权限种子 读/写权限矩阵**；Task 3.2 新增 **426 GET:/api/community/notices/:id** + **422 扩展全部移动端角色**（同 423/424/425 批次）；§GetNotice/§ListNotices/§GetMarqueeNotices 补鉴权行；Task 5.1/5.2 断言同步 ✅
  - 评审 interface-proto v4 🔴 MUST 2（通知写路径权限矩阵缺失：DELETE/PUT 无权限码）→ Task 3.2 新增 **427 DELETE:/api/community/notices/:id** + **428 PUT:/api/community/notices/:id** 绑定全部移动端角色（真正越权判定交 080002 作者校验）；§DeleteNotice/§UpdateNotice 补鉴权行；Task 5.2 撤回冒烟同步 ✅
  - 评审 interface-proto v4 🟡 SHOULD 3（423/424/425 parent_id 未指定）→ Task 3.2 显式 parent_id：423/424/426→410、427/428→420、425→310；Task 3.3 补 parent_id=310；design §权限种子 parent_id 汇总 ✅
  - 评审 interface-proto v4 🟡 SHOULD 4（division_id 空串 REST 层 4xx）→ Task 4.1/4.6 注明「不适用时省略 division_id 字段，不发送空串」；RPC double-empty 判 080005 兜底 ✅
  - 评审 interface-proto v4 🔵 INFO 5（幻影 435）→ Task 3.2 措辞改为「lostfound:create-api 无 sys_permission 行，保持现状不动」；Task 5.1 断言不与幻影冲突 ✅
  - 评审 interface-proto v4 🔵 INFO 6（rbac-design §6.5 矩阵补登）→ Task 3.2 新增 rbac-design.md §6.5 补登 421/422/423/424/425/426/427/428 矩阵 ✅
  - 评审 data-model v4 🟡 S1（toProtoNotice 转换器未随模型类型变更）→ Task 1.3 补 toProtoNotice 适配（*int64 + sql.NullTime null 感知 + community_id 由 scope 注入）；design §GetNotice 补转换器同步说明 ✅
  - 评审 data-model v4 🟡 S2（撤回单事务机制未指定）→ Task 1.12 钉死 `conn.Transact` 传共享 session（model 方法签名改接受 session）+ 失败注入测试（无孤儿 scope 行）；design §DeleteNotice 同步 ✅
  - 评审 data-model v4 🟡 S3（JOIN 双 community_id 列投影陷阱）→ Task 1.4 显式投影 `notices.*` + `notice_scope.community_id`；Task 1.9 RED 补「C2 列表返回 community_id=C2」断言；design §ListNotices 同步 ✅
  - 评审 data-model v4 🟡 S4（file_url VARCHAR(500) 截断）→ Task 1.1 migration 003 加宽 `file_url VARCHAR(1024)`；Task 1.7 新行写占位空串（file_id 权威载体）；design §CreateNotice/§数据模型 同步 ✅
