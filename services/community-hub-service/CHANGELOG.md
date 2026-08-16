# CHANGELOG — community-hub-service

## 2026-08-16 — QA 门禁修复（graph_freshness + memory-index 过期，运维类非源码缺陷）

### 做了什么（类型：运维操作，无源码逻辑改动）
- **graph_freshness FAIL 消除**：上一轮 QA（`_qa.md`）唯一 FAIL 为机械化检查第 10 项 graph_freshness——api-proto 子模块 commit `9c848cb`（since_days proto，ts=1786871984）晚于图谱同步时间戳（1786856530），图谱过期。修复：执行 `bash .harness/scripts/graph-sync.sh`（Neo4j HTTP/bolt 可达，增量同步成功，Proto 222 消息/96 RPC、Go 9 服务/27 表、TS 52 接口全部并入图谱，时间戳更新至 1786873166）。**非本服务源码缺陷**（graph_freshness 扫描全仓库提交非工作树 diff）；本工作树 go build/vet/test、json_string、错误码、跨服务导入等 18 项检查本就 PASS。
- **memory-index freshness FAIL 消除（同步暴露）**：门禁重跑时新暴露第 17 项 memory index 过期（记忆文件 `tdd-red-evidence-requires-fail-excerpt.md` 更新后索引未重建）。修复：执行 `bash .harness/scripts/memory-index-build.sh` 重建 `.memory-index.json`。
- **无源码/无 TDD 变更**：本轮为纯运维门禁修复，不涉及逻辑函数，无 RED/GREEN 证据要求。

### 验证结果（--service community-hub-service 全量重跑）
- 机械化检查：**19 PASS / 0 FAIL / 2 WARN**（graph_freshness + memory_index 均已转 PASS）。
- 剩余 2 WARN 为既有跨服务项，非本变更引入：proto→TS 对齐（identity.ts 5 字段 / moderation.d.ts reviewer_id 滞后，属其他服务/前端任务）；git hygiene（api-proto gitlink 无 .gitmodules 条目，Git 治理规范）。
- 门禁：`go build ./...` exit 0；`go test ./... -count=1` 13 包 ~124 测试 0 fail。

---

## 2026-08-16 — 移动端首页信息架构改造（mobile-homepage-content-revamp Task 1.1-1.5）

### 做了什么（类型：model 逻辑改造 + RPC 逻辑改造 + REST 透传 + migration 补表/索引）

**Migration（字段映射类，纯 DDL 无 TDD；执行 + DESCRIBE/EXPLAIN 验证归 Task 3.1 Owner 运维验证）**
- `migration/004_add_community_contacts.sql`（Task 1.1）：`CREATE TABLE IF NOT EXISTS community_contacts`，DDL 与 001_initial.sql / model/community_contact.go 完全对齐（8 列 + idx_community，InnoDB/utf8mb4）；**不预置种子数据**（REQ-CLP-2 场景 3 空态）；文件头声明 001 为 schema 单源、004 为运行库缺表幂等补救、结构漂移不自动修复需人工订正（场景 5）。
- `migration/005_content_posts_window_index.sql`（Task 1.2）：`idx_status_pinned_published (status, is_pinned, published_at)` 幂等守卫（`information_schema.statistics` 检查，MySQL 8.0 无 ADD INDEX IF NOT EXISTS）；列序等值 status + 等值 is_pinned + 范围 published_at，覆盖 ORDER BY 前导列减少 filesort（REQ-NTW-6 / ADR-3）；纯增量不影响缺省非窗口调用。

**Model（逻辑函数 TDD）**
- `model/content_post.go`（Task 1.3）：新增 `ContentPostListOption` 变参选项 + `WithTimeWindow(since)` 构造器（内部 `*contentPostListParams.since *time.Time`，nil=不过滤，additive）+ `ContentPostListOptionSince` 自省助手 + `buildWindowClause` 共享谓词构建。`FindListByCommunity` 签名改变参 `opts ...ContentPostListOption`——既有调用方/测试零改动即可编译；窗口选项存在时 count/list 两 SQL 追加 `and content_posts.published_at >= ? and content_posts.published_at <= ?`（下界 since、上界 `time.Now()`，参数化防注入 D12），NULL 恒不匹配下界、未来行被上界排除。

**RPC（逻辑函数 TDD）**
- `rpc/internal/logic/notice/listcontentpostslogic.go`（Task 1.4）：`since_days` 参数校验先于业务逻辑（fail-fast）——`<0 || >365` → 080005（以 Base 错误返回，非 gRPC err；r2-5：int32 wire 恒数字，非数字由 REST 网关解析层拒绝）；`==0`（缺省）不过滤（PC 管理列表兼容）；`>0` → `model.WithTimeWindow(time.Now().AddDate(0,0,-since_days))` 传入 `FindListByCommunity`。

**REST（逻辑函数 TDD）**
- `api/internal/types/types.go` + `api/internal/logic/notice/listcontentpostslogic.go`（Task 1.5）：`ListContentPostsReq` 新增 `SinceDays int32 form:"since_days,optional"`；RPC 请求补 `SinceDays: req.SinceDays`（**必贯通**，漏传致移动端 30 天窗口静默失效，REVISION r2-2）；该 logic 补 `responsex.ToError(resp.GetBase())` 上抛（RPC 侧 080005 以 Base 错误返回，REST 层禁止静默吞错，与 getcontentpostlogic.go 既有模式一致）。

### TDD 证据（RED 摘录）
- Task 1.3（model）：`undefined: ContentPostListOptionSince` / `undefined: WithTimeWindow` / `too many arguments in call to m.FindListByCommunity`（编译期 RED）→ GREEN：`go test ./model/ -run 'TestContentPostListOption_WithTimeWindow|TestContentPostModel_FindListByCommunity'` PASS（窗口内行返回 + NULL/未来行被 SQL 谓词排除 + 无窗口缺省不过滤）。
- Task 1.4（RPC）：`Should NOT be empty, but was []`（since_days=30 未传窗口选项）+ `expected: 80005, actual: 0`（-1/366 未拦截）→ GREEN：`go test ./rpc/internal/logic/notice/ -run TestListContentPosts -count=1` 5 用例 PASS。
- Task 1.5（REST）：`unknown field SinceDays in struct literal`（编译期 RED）→ GREEN：`go test ./api/internal/logic/notice/ -run TestListContentPosts -count=1` 3 用例 PASS（30 透传 / 080005 上抛 / 成功映射 notices+total）。

### 影响
- 配置：无新增（复用既有 RPC/API 配置）。
- 依赖：无新增（复用 go-sqlmock 测试依赖）。
- 兼容：RPC/API 请求响应契约 additive 非破坏（since_days 新增可选字段，缺省 0 行为不变）；迁移 004/005 幂等。
- 跨服务：无新增依赖（api-proto since_days 字段已由全局 Claude 完成，Task 0.1-0.3）。
- 门禁：`go build ./...` + `go vet ./...` + `go test ./... -count=1` 全绿。

---

## 2026-08-16 — 运维验证修复（Task 6.1/6.2 发现，published_at 落库 + Kafka 推送非阻塞）

### 做了什么
- **`model/content_post.go` Insert SQL 补 `published_at` 列（真实 bug 修复）**：submit 路径逻辑层已置 `PublishedAt=sql.NullTime{NOW}`，但 Insert 列清单漏了该列 → published_at 恒 NULL → `FindMarquee`（`published_at >= since` 过滤）恒空。修复后 draft 写 NULL、submit 写 NOW()；配套补回归测试 `TestContentPostModel_Insert_SubmitWritesPublishedAt`（运维验证真实暴露，见 `_qa.md`/`_tdd_evidence.md`）。
- **`rpc/internal/kafkapush/producer.go` Push 独立短超时（真实 bug 修复）**：Push 原用 RPC 请求 ctx 同步写，Kafka 不可用时 kafka-go WriteMessages 重试耗尽请求 deadline → 客户端收到 `DeadlineExceeded`（尽管帖已提交成功）。修复：网络写/附件 URL 再生走 `context.WithTimeout(context.Background(), 3s)`，DB 落标（markPending/ack）走请求 ctx → 推送失败不阻塞发布（D20），客户端收到成功 + `kafka_push_status=1`(pending) 待重推。
- **`rpc/etc/communityhub.yaml` brokers 改 `localhost:29092`（部署修复）**：community-hub 为宿主机进程不在 compose 网络，原 `kafka:9092` 无法解析；docker-compose 为 Kafka 增 EXTERNAL listener（`localhost:29092` 宿主可达）。

### 验证（Task 6.2 端到端冒烟，真实 DB + Kafka）
- 发布 → status=2 + published_at=NOW + Kafka content-review 消息（REQ-CPM-2 契约）→ 列表/详情/跑马灯可见；R2 wire 键 `notices`/`notice`/`content` 保持。
- Kafka 不可用（D20）：发布成功（0.09s）+ pending 标记；恢复后扫描器补投（status 1→2）。
- 权限矩阵：owner 只读（发帖 99401）、纯 property_admin 发布成功 + DELETE/PUT 403（fail-closed 427/428）、审核完整性（rejected 附件帖读路径隐藏）。

## 2026-08-16 — 通用图文发布组件重构（content-post-generalization Task 1.1-1.23）

### 做了什么（类型：model 重构 + 逻辑改造 + 新增功能）

**数据模型（字段映射类 + 读/写逻辑）**
- `migration/003_content_posts_generalize.sql`（Task 1.1，一次性 RENAME 勿重跑）：`notices`→`content_posts`（content→`text` + `section_code/status/attachment_count`）+ `published_at`/`community_id` 去 NOT NULL + Kafka 待推列（`kafka_push_status/kafka_push_retries/kafka_push_last_error/kafka_pushed_at`）+ 新建 `content_post_scope`（复合 PK + idx_scope_community）+ `notice_attachments`→`content_post_attachments`（notice_id→post_id + review_status/file_id/file_type + file_url 加宽）；同步 `docs/design.md` §数据模型。
- `model/content_post.go`（rename 自 notice.go，Task 1.2/1.5/1.6）：`Content`→`Text`、`CommunityId`→`*int64`、`PublishedAt`→`sql.NullTime`；导出状态常量（StatusDraft..Withdrawn / KafkaPushNone..Done）；`IsReviewComplete` 审核完整性单一谓词；读查询（`FindListByCommunity` scope JOIN + 完整性子查询走 idx_notice / `FindOneReviewComplete` / `FindMarquee`）+ 写路径（Insert 显式状态、UpdateContent 三列、**UpdateIsPinned 独立列**、UpdateStatusAndPublish、UpdateAttachmentCount、Withdraw、UpdateKafkaPushStatus、FindPendingPush）+ Tx 变体。
- `model/content_post_scope.go`（Task 1.3）：InsertBatch/FindCommunityIdsByPostId/DeleteByPostId + Tx。
- `model/content_post_attachment.go`（rename 自 notice_attachment.go，Task 1.4）：post_id/review_status/file_id/file_type + InsertBatch/FindByPostId/DeleteByPostId + Tx。
- `rpc/internal/svc/servicecontext.go`：注册新模型 + Conn + UserClient（R5）+ FileClient + KafkaProducer/Rescanner（D20）。

**scope 基础设施（逻辑函数 TDD）**
- `rpc/internal/logic/scope/scope.go`：`AssertCommunitiesScope`（多目标单次批量 all-or-nothing）。
- `rpc/internal/logic/scope/division.go`（Task 1.7）：`ExpandDivisionCommunities`（guard + GetResidentialAreasByDivision status=1）+ `ResolveAdminDivision`（R1 grounded：经既有 community grant 派生唯一 division，URStatus==2 过滤）。
- `rpc/internal/logic/scope/userctx.go`（Task 1.8）：`PublishRolesFrom`（level-2 + 优先序）+ `PublishRoleToString`（RBAC→DB role 列）+ `IsLevel2Grant`。

**写逻辑（逻辑函数 TDD）**
- `createcontentpostlogic.go`（Task 1.10）：板块白名单 + title/text 非空 + community_ids 去重 + 社区管理员 division 展开（>100→080003）+ 附件绑定（GetFileUrl confirmed/user_id 归属/≤10/≤50MB）+ 单次批量 AssertCommunitiesScope + JWT/RBAC/档案身份派生 + 单事务落库 + entry=submitted 隐式通过（status=2 + published_at=NOW + kafka_push_status=1）→ 事务提交后 Producer.Push；**不再 LPUSH Redis**（D3）。
- `updatecontentpostlogic.go`（Task 1.11，V5 presence 权威实现）：Title/Text/SectionCode/IsPinned 用 proto3 `optional` 指针判 presence、附件/scope 以 HasAttachmentChange/HasScopeChange 标志；授权分流（(a) 内容编辑先作者校验 080002 / (b) 仅 is_pinned 跳过作者校验改验 PublishRolesFrom+AssertCommunitiesScope）；**UpdateIsPinned 独立列更新防正文清空**；attachment_count 同事务重算（空集=0）；submit 动作 UpdateStatusAndPublish + Producer.Push；**整体移除 CreateAuditLog + LpushCtx**（评审 M3）。
- `deletecontentpostlogic.go`（Task 1.12）：仅发布者本人 080002 + Withdraw 单语句原子（scope/附件保留）。

**读逻辑（逻辑函数 TDD）**
- `listcontentpostslogic.go`（Task 1.13）：FilterAllowed + role 枚举→DB 列映射收敛 helper.go 单源。
- `getcontentpostlogic.go`（Task 1.14）：community_id RPC 必填 + FindOneReviewComplete + scope 匹配 + 附件 file_url 重生（file_id 权威 / 0 回退 stored）+ `ResolveReadableCommunityForCompat`（委托共享 `internal/contentcompat`）。
- `getmarqueenoticeslogic.go`（Task 1.15）：FindMarquee（15 天含端点、≤10、置顶优先）+ FilterAllowed。
- `getpublishpermissionlogic.go`（Task 1.16）：level-2 判定（含 property_admin）+ ContentPostRole 映射。

**Kafka 基建（Task 1.17-1.19）**
- `rpc/internal/config/config.go` + `etc/communityhub.yaml`：Kafka（brokers/topic/retry）+ UserRpc + FileRpc；go.mod +`segmentio/kafka-go`。
- `rpc/internal/kafkapush/producer.go`（Task 1.18）：ContentReviewMessage 契约（REQ-CPM-2 单源）+ Producer.Push（先提交后推送；成功 ack 置 2 / 失败置 pending + last_error 落库）。
- `rpc/internal/kafkapush/rescanner.go`（Task 1.19）：定时重推复用 Producer.Push（内含 GetFileUrl 重生 file_url）+ 超阈值保留 pending + pending-count 可观测。

**接口（Task 1.21-1.23）**
- `rpc/internal/server/communityhubserver.go` + `rpc/communityhub.go`：NoticeService→ContentPostService 改名 + GetPublishPermission/GetMarqueeNotices + 移除 UpdateNoticeModerationStatus。
- `api/internal/types/types.go`（Task 1.22）：ContentPost 类型 + pointer/标志位 presence 字段 + form 标签 + **R2 wire 兼容（notices/notice/content 键保持）**。
- `api/internal/logic/notice/*` + handler + routes（Task 1.23）：RPC 代理（community_ids []string→[]int64 + entry_status/status 同号映射 + **Update presence 指针/标志转发**）+ **详情 community_id 兼容回退（R2：scope 反查 + 逐小区 FilterAllowed 任一允许即放行）** + marquee/publish-permission 静态路径先于 :id 注册。

### 关键设计决策
- **R2 wire 兼容**：REST 响应键保持 `notices`/`notice`/`content`（移动端 tabbar/浏览/详情现行消费方零改动），RPC/proto/DB 通用化改名 `text` 分轨。
- **详情 community_id 兼容回退**：RPC 层严格必填（080005），回退只落 REST 薄代理层（contentcompat scope 反查 + FilterAllowed 任一允许即放行）——多小区用户迁移后详情不 080005。
- **V5 presence 语义**：UpdateContentPost 分支判定以 proto3 `optional` 指针/标志位为准，禁止 value 非空启发式（取消置顶 `*false`、清空全部附件 `attachment_count→0` 确定性可达）。
- **is_pinned 独立列更新**：置顶/取消置顶一律走 `UpdateIsPinned`，禁止复用 UpdateContent 传空 title/text（防清空已发布帖正文）。
- **停 Redis 只推 Kafka（D3）**：content_posts Create/Update(submit) 不再 LPUSH `moderation:task:queue`；lostfound/user 仍走 Redis（双轨过渡）。

### TDD
- 含逻辑函数任务均先写失败测试（RED）再实现（GREEN）：scope 包（division/roles/userctx）、写逻辑（create/update/delete）、读逻辑（list/get/marquee/publish-permission）、Kafka（producer/rescanner）、contentcompat、API 代理（presence 转发 + compat 回退）。RED 摘录留档于测试注释（080006/080005/080002 映射、attachment_count 重算、is_pinned 操作者路径）。
- 字段映射类（model struct/SQL/Tx 变体/proto 透出）测试绿即可（content_post_test/content_post_scope_test/content_post_attachment_test sqlmock 锁定 SQL）。

### 影响
- 配置：yaml 新增 `UserRpc`/`FileRpc`/`Kafka` 段；API yaml 新增 `DataSource`（仅 contentcompat 只读查询）。
- 依赖：go.mod 新增 `segmentio/kafka-go`（v0.4.51）。
- 兼容：REST 路径保持 `/api/community/notices` + wire 键保持（R2）；RPC/proto/DB 通用化改名（破坏性，预期登记 Task 0.1）。
- 跨服务：新增依赖 user-service `GetUsersByIds`、file-service `GetFileUrl`、master-data `GetResidentialArea/GetResidentialAreasByDivision`、permission `GetUserRoles/GetDataScopes`。
- 门禁：harness-checks 全绿（除 file-service proto→TS 的 FileInfo.confirmed/file_type 同步属 file-service 任务范围）；`go build ./...` + `go vet ./...` + `go test ./...` 全绿。

---

## 2026-08-13 — 板块发布配额（access-control Task 4.1-4.4）

### 做了什么
- `model/lost_found_item.go`：新增 `CountQuotaOccupied(ctx, publisherId, communityId, typ)`（Task 4.1），
  谓词 `deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)` —— 待审(0)+通过(1) 同占配额，
  驳回(2)/已解决(resolved)/已删除(deleted_at 非空) 释放；口径「用户×小区×板块」按目标小区计。
- `rpc/internal/logic/scope/section_quota.go`：新增 `CheckSectionQuota` + `quotaAllowed`（Task 4.2），
  消费 master-data `GetSectionQuota`（configured=false 不限），计数 `>= max_count` → 80007。
- `rpc/internal/logic/lostfound/createlostfoundlogic.go`（Task 4.3）：`AssertPublishScope` 之后、`Insert` 之前
  挂载 `CheckSectionQuota`，超限返回 080007，传输/DB 错误原样传播（fail-closed）。
- `rpc/internal/svc/servicecontext.go`：新增 `MasterDataClient`（GetSectionQuota 客户端）。
- 错误码 80007（Task 4.4）：`api/internal/types/types.go` 登记 `CodeSectionQuotaExceeded`；
  常量实现在 `scope.CodeSectionQuotaExceeded`。

### 关键设计决策
- **板块口径**：配额按板块（`sys_section_quota.section_type`，种子 `lost_found=5`）配置；`lost_found_items`
  表即 lost_found 板块唯一承载（`type` 列仅区分 lost/found 子类，二者同属 lost_found 板块共同占配额），
  故 `CountQuotaOccupied` 不按 `type` 过滤，按「用户×小区×板块」统计整板占配额内容。`typ` 参数保留为
  多板块扩展位（对应 tasks.md SQL 中的 `type=?`，此处以板块为粒度收敛，避免 `type='lost_found'` 恒空）。
- **校验顺序**：功能权限（PermMiddleware）→ 数据权限（AssertPublishScope）→ 配额（CheckSectionQuota）→ 落库。
- **fail-closed**：`GetSectionQuota`/`CountQuotaOccupied` 传输或 DB 错误原样返回，不静默放行。

### TDD
- 含逻辑函数（CheckSectionQuota/quotaAllowed/CountQuotaOccupied/CreateLostFound 挂载）均先测试后实现：
  `section_quota_test.go`（未配置/4-5/5-5/6-5/传输错误/DB 错误）、`lost_found_item_quota_test.go`
  （sqlmock 锁定谓词 + DB 错误）、`publishscope_test.go` 新增 `TestCreateLostFound_QuotaExceeded`。
- RED 摘录（行为型，已留档于测试注释）：`expected: 80007, actual: 0`（达上限未拦截直接落库）。

### 影响
- 配置：无新增（复用 `MasterDataRpc` 目标 `masterdata.rpc`）。
- 依赖：`go.mod` 新增测试依赖 `go-sqlmock v1.5.2`（model 层 sqlmock 边界测试）；`testify` 由间接转直接。
- 兼容：RPC/API 请求响应契约未变；新增 080007 错误码（原 80001-80006 未占用码位）。
- 跨服务：依赖 master-data `GetSectionQuota` RPC（proto 已由 Owner 生成，逻辑侧 Task 5.2 并行交付）；
  调用失败 fail-closed 传播。
- 门禁：harness-checks 18 PASS / 0 FAIL / 0 WARN；`go build ./...` + `go vet ./...` + `go test ./...` 全绿。

---

## 2026-08-13 — 审核可见性门禁：读路径仅返回审核通过内容

### 做了什么
- `model/notice.go` + `model/lost_found_item.go`：
  - `FindList` 查询加 `moderation_status = 1` 过滤（列表只返回审核通过内容）
  - 新增 `FindOnePublished`（读路径专用，仅返回 moderation_status=1；`FindOne` 保留给写接口 / 审核回调内部使用）
- `getnoticelogic.go` + `getlostfoundlogic.go`：读路径 `FindOne` → `FindOnePublished`
- 新增 TDD 测试：`TestGetNotice_FilterByModerationStatus` / `TestGetLostFound_FilterByModerationStatus`

### 为什么
数据权限 Wave 阶段④遗留（summary.md WARN）：读列表/详情只按数据范围（scope）过滤，未按审核状态过滤，
普通用户能看到「待审核/拒绝」的内容。本变更落地「最小实现」：全员读路径仅见审核通过内容。

### 影响
- 读路径语义：moderation_status=0（待审核）/2（拒绝）内容对普通用户不可见（表现为「不存在」80001/80004）
- 写接口 / moderation 回调内部仍用 FindOne（不受影响）
- 门禁：harness-checks 16 PASS / 0 FAIL / 0 WARN

## 2026-08-12 — TDD 证据补齐（8 个包装函数无测试，QA TDD FAIL → 修复）

### 背景
终局 LLM QA 的 TDD 证据表发现：上一轮（Get-by-ID 数据范围 + RPC 身份伪造修复）核心函数
（GetNotice/GetLostFound scope 过滤、API CallCtx+ToError、回环绑定、scope 包）全部有测试，
但**同批被连带修改的 8 个包装函数无对应测试**，判定 QA FAIL（TDD 证据不足）：
- RPC `ListLostFoundLogic.ListLostFound` / `ListContactsLogic.ListContacts`（新增 FilterAllowed scope 过滤）
- API `ListNotices/ListLostFound/ListContactsLogic`（新增 CallCtx 身份注入）
- API `UpdateNotice/DeleteNotice/ResolveLostFoundLogic`（新增 CallCtx + responsex.ToError 透出 080006）

根因：Generator 只给变更的 headline 函数写测试（本轮=Get-by-ID），把同构的列表/更新/删除/解决
包装函数当作"同逻辑已被代表函数覆盖"，但 Go 测试隔离使这些函数从未被任何测试引用；机械门禁
（harness-checks.go_test / tdd-evidence-validator）只查包级测试文件与 RED 列文本格式，无按函数维度门禁。

### 修复内容（仅补测试 + fake，无生产代码改动）
- 新建 `rpc/internal/logic/lostfound/listlostfound_filter_test.go`：`TestListLostFound_FilterByScope`
  四态（EMPTY 不查库 / LIMITED 未命中不查库 / LIMITED 命中查询 / GLOBAL 查询），照抄
  `rpc/internal/logic/notice/listnotices_filter_test.go` 模板；复用已存在的 fakeReadPerm/readPerm/ctxWithUserID。
- 新建 `rpc/internal/logic/contact/listcontacts_filter_test.go`：`TestListContacts_FilterByScope` 同四态；
  本文件新增 contact 包的 `fakeReadPerm`/`listPerm`（GetDataScopes fake）。
- 新建 `api/internal/logic/notice/listnoticeslogic_test.go`、`api/internal/logic/lostfound/listlostfoundlogic_test.go`、
  `api/internal/logic/contact/listcontactslogic_test.go`：`TestList*_InjectsIdentity`，照抄
  getnoticelogic_test.go 的 `TestGet*_InjectsIdentity` 模板（出站 metadata 断言 user_id=42）。
- 新建 `api/internal/logic/notice/updatenoticelogic_test.go`、`deletenoticelogic_test.go`、
  `api/internal/logic/lostfound/resolvelostfoundlogic_test.go`：`Test*_InjectJWTAndSurfaceBaseError`
  双路径（成功 Base → nil；080006 Base → errx.FromError Code==80006 + 身份注入），照抄
  `api/internal/logic/contact/upsertcontactslogic_test.go` 的 InjectJWT + SurfaceBaseError 模板。
- fake 扩展：`rpc/internal/logic/lostfound/publishscope_test.go` 的 `fakeLostFoundModel` 补
  `FindList`（findListCalled 标志）；`rpc/internal/logic/contact/upsertcontactslogic_test.go` 的
  `fakeContactModel` 补 `FindByCommunityId`（findByCalled 标志）。

### TDD 证据（RED 摘录，均通过 `git stash` 回退生产文件到 HEAD 复现真实 FAIL 后恢复）
- RPC 过滤（HEAD 无 FilterAllowed，EMPTY/LIMITED-未命中仍查库）：
  - `listlostfound_filter_test.go:64: Error: Not equal: expected: false`（findListCalled 应为 false）
  - `listcontacts_filter_test.go:83: Error: Not equal: expected: false`
- API 列表（HEAD 无 CallCtx，出站 ctx 无 metadata）：
  - `listnoticeslogic_test.go:48 / listlostfoundlogic_test.go:48 / listcontactslogic_test.go:47:
    Error: Should be true / 出站 metadata 必须存在`
- API 写路径（HEAD 无 CallCtx + 无 ToError，080006 Base 被静默吞掉）：
  - `updatenoticelogic_test.go:59 / deletenoticelogic_test.go:59 / resolvelostfoundlogic_test.go:59:
    Error: An error is expected but got nil.`（越权 080006 未透出）
  - 同测试注入子用例：`Error: Should be true / 出站 metadata 必须存在`

### 影响
- 生产代码零改动（8 个逻辑文件维持工作树原状）；仅新增 8 个测试文件 + 扩展 2 个 fake
- 门禁：`go build ./...` + `go vet ./...` 全绿；`go test ./... -count=1` 全绿（8 个新测试函数全 PASS）
- 修复后 TDD 证据表：26 个新增/修改函数全部有测试命中，无 ❌

---

## 2026-08-12 — 多视角评审修复（Get-by-ID 数据范围 + RPC 身份伪造，TDD）

### 背景
阶段④交付后多视角审查 1/3 PASS：安全架构 FAIL(2 CRITICAL)、设计业务 FAIL(1 CRITICAL)。
本条目修复全部 3 项 CRITICAL。

### 修复内容

**1. GetNotice/GetLostFound 单条读取补数据范围过滤（security-arch + design-biz 双重 CRITICAL）**
- 根因：T4.6/REQ-1.6 读过滤只挂载于 List（ListNotices/ListLostFound/ListContacts），
  Get-by-ID 直接 `FindOne(id)` 返回完整内容（notice 全文/附件、lost_found 的
  description/contact_phone），无 FilterAllowed/AssertCommunityScope。LIMITED/EMPTY 用户
  被列表过滤后仍可按 ID（Snowflake 时间有序可枚举 / 分享链接）越权读取，违背
  『注册用户读不到小区内部内容』。
- 修复（RPC 层）：
  - `rpc/internal/logic/notice/getnoticelogic.go` / `lostfound/getlostfoundlogic.go`：
    `FindOne` 后 reverse-lookup 内容 community_id → `scope.FilterAllowed(userID, communityID)`；
    不通过 → `scope.DenyBase()`(080006) + 空内容，且不查询附件/不返回越权内容；
    传输错误原样传播（fail-closed）。
- 修复（API 层）：
  - `api/internal/logic/notice/getnoticelogic.go` / `lostfound/getlostfoundlogic.go`：
    补 `CallCtx` 注入（与 List 一致，使 JWT 身份经 metadata 到达 RPC 读过滤）；
    `responsex.ToError(resp.GetBase())` 透出 080006/80001 业务错误（此前静默返回 200 空）。
- 回归测试（先 RED 后 GREEN，断言越权 Get → 080006 + 空 + 不查附件）：
  - `rpc/internal/logic/notice/getnoticelogic_test.go`（5 用例：LIMITED 命中/未命中、EMPTY、无身份 fail-closed、GLOBAL）
  - `rpc/internal/logic/lostfound/getlostfoundlogic_test.go`（同矩阵 5 用例）
  - `api/internal/logic/notice/getnoticelogic_test.go`、`api/internal/logic/lostfound/getlostfoundlogic_test.go`
    （身份注入 + 080006 透出）

**2. RPC 身份伪造风险（security-arch CRITICAL，仓库级模式）**
- 根因：数据权限身份经未认证的 gRPC metadata 传输，RPC 绑定 `0.0.0.0:8088` 且无服务间鉴权，
  `UserIDFromCtx` 盲信入站 user_id → 局域网 / Docker 桥接网络对端可注入任意身份，击穿数据权限。
- 修复（本服务可落地部分）：
  - `rpc/etc/communityhub.yaml`：`ListenOn: 0.0.0.0:8088 → 127.0.0.1:8088`（回环绑定）。
    go-zero `figureOutListenOn` 对非 0.0.0.0 host 原样注册 etcd；单机拓扑（scripts/start.sh
    全部 go run 于宿主）下 API 网关 + moderation 回调可正常发现。阻断局域网 / Docker 桥接对端。
  - `rpc/internal/logic/scope/userctx.go`：信任边界文档化（metadata 盲信的安全前提 = 网络隔离），
    并标注仓库级根治方向（服务凭据/mTLS + unary 拦截器，涉及 common/ 与全部 9 服务，Owner 协调）。
  - `rpc/internal/config/config_test.go`：配置不变式测试 `TestRpcConfig_BindsLoopback`，
    ListenOn 回退 0.0.0.0 即 FAIL（RED 已复现，`Should be true / host="0.0.0.0"`）。
  - 沉淀记忆 `global/rpc-identity-spoofing-loopback-isolation.md`（仓库级 9 服务 0.0.0.0 无鉴权模式），
    已建索引 + 登记 MEMORY.md should-follow。

### TDD 证据（RED 摘录）
- `expected: 80006, actual: 0` + `Expected nil, but got: &communityv1.Notice{...CommunityId:200...}`
  —— 越权 Get 原本返回内容（GetNotice 4 个拒绝用例全 FAIL）。
- `Should be true / 出站 metadata 必须存在`、`An error is expected but got nil` —— API 未注入身份/未透出 080006。
- `Should be true / RPC 必须绑定回环（127.0.0.1/localhost），当前 host="0.0.0.0"` —— 配置不变式。

### 影响
- 配置：rpc/etc/communityhub.yaml ListenOn 0.0.0.0 → 127.0.0.1（脚本端口探测 netstat `:8088` 仍匹配，无破坏）
- 兼容：RPC/API 请求响应契约未变（身份仍经 gRPC metadata）；Get-by-ID 越权时由 200 空 → 080006 错误，属预期修复
- 依赖：无新增；复用 scope.FilterAllowed / scope.DenyBase
- 门禁：harness-checks 16 PASS / 0 FAIL / 0 WARN；`go build ./...` + `go test ./...`（含 -count=1 全量）全绿

---

## 2026-08-12 — 数据权限消费方（access-data-permission 阶段④ T4.0-T4.8，TDD）

### 做了什么
- **T4.0 身份注入通道**：新增 `api/internal/util/userctx.go`（`JWTUserID(ctx) (int64,error)` 提取 JWT、`WithUserID` 注入出站 gRPC metadata）；`ServiceContext.CallCtx` 统一封装「提取 JWT → 注入 metadata」，供所有 API→RPC 调用使用
- **T4.1 publisher_id 规范化**：CreateLostFound/CreateNotice API 逻辑用 JWT user_id 覆盖 gRPC 请求 publisher_id，忽略客户端 body 值（防伪造）
- **T4.2-4.4 AssertPublishScope 挂载（写接口）**：新增 `rpc/internal/logic/scope` 包（`AssertCommunityScope` / `CheckPublishScope` / `CheckSystemPublishScope` / `FilterAllowed`）；RPC 层落库前对目标 community_id 校验，`allowed=false` → 080006，映射 permission 060007 → 080006；覆盖 CreateLostFound/ResolveLostFound/CreateNotice/UpdateNotice/DeleteNotice/UpsertContacts
- **T4.5 moderation 回调身份校验（S4）**：UpdateNotice/UpdateLostFoundModerationStatus 先 reverse-lookup 内容 community_id（查不到拒绝），以系统身份（system_user_id=0 常量，global scope）调 AssertPublishScope（服务身份回调放行，不按作者 scope）
- **T4.6 读列表按 GetDataScopes 过滤**：新增 `rpc/internal/logic/scope/filter.go`；ListNotices/ListLostFound/ListContacts 按 GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表过滤（空列表在逻辑层返回，不拼空 IN）；API 列表逻辑同步注入 metadata
- **T4.7 错误码 080006 注册**：`api/internal/types/types.go` 登记 08xxxx 常量（080002 功能权限 / 080006 数据权限分层语义）；`rpc/internal/logic/scope.CodePublishScopeDenied=80006`
- RPC 配置新增 `PermissionRpc`（communityhub.yaml + config.go + servicecontext.go 挂 permission 客户端）

### 关键设计决策
- **身份链路**：JWT（rest.WithJwt 注入 ctx）→ `JWTUserID` 提取 → `WithUserID` 注入出站 gRPC metadata → RPC 层 `UserIDFromCtx` 读取，用于 AssertPublishScope/GetDataScopes。不信任客户端 body 的 publisher_id/userId
- **校验顺序**：功能权限（PermMiddleware，中间件链先于 handler）→ 数据权限（RPC 落库前）→ 落库
- **系统身份**：system_user_id=0 是合法 global 身份（moderation 回调），走同一 grant 判定路径，无代码级短路；与 CheckPublishScope 的「userID==0 拒绝」分支区分（经 CheckSystemPublishScope 直接 AssertCommunityScope）
- **fail-closed**：读过滤 userID==0 恒拒绝（0 是系统身份，禁止用户读路径借用）；写接口无身份 → 080006
- API 写逻辑（Resolve/Update/Delete/Upsert）同步补 `responsex.ToError(resp.GetBase())`，使 RPC 业务错误（080006）透出客户端（此前忽略 Base.Code 的既有缺陷在本变更触及处修复）

### TDD 证据
- RED 摘录示例：`expected: 80006, actual: 0` / `expected: 1001, actual: 999999`（伪造 publisher_id 未被覆盖）/ `undefined: JWTUserID`（新函数编译期 RED）
- 新增测试文件（11 个）：util/userctx_test、svc/servicecontext_test、api logic（lostfound/notice create、contact upsert）、rpc logic（lostfound publishscope、notice publishscope、notice updatemoderationstatus、contact upsert、scope scope/filter、notice listnotices_filter）

### 影响
- 配置：rpc/etc/communityhub.yaml 新增 `PermissionRpc`（permission.rpc）
- 依赖：无新外部模块；复用 api-proto permission/v1 生成代码（AssertPublishScope/GetDataScopes，契约已提交 031f4e4+c245c09）
- 兼容：不触碰 api-proto/ 与 common/；RPC 写/读接口请求参数与响应契约未变（身份经 gRPC metadata 传递，非 proto 字段）
- 门禁：harness-checks 16 PASS / 0 FAIL / 0 WARN；`go build ./...` + `go test ./...` 全绿

---

## 2026-06-06 — 服务初始化

### 做了什么
- 创建 community-hub-service 微服务，实现社区枢纽功能
- 实现 RPC + REST API 双层架构（端口 8087/8887）
- 实现 NoticeService（通知公告 CRUD + 软删除）
- 实现 ContactService（便民联络列表 + 批量更新）
- 实现 LostFoundService（寻失互助 CRUD + 标记解决）
- 创建 4 张数据库表（notices, notice_attachments, community_contacts, lost_found_items）
- 使用 go-zero sqlx 风格数据模型（4 个 Model）
- 所有 int64 ID 使用 json:",string" 标签
- 使用 configx.MustLoad 加载配置（支持 ${ENV_VAR}）
- 使用 Snowflake 生成分布式唯一 ID

### 为什么
社区平台需要小区信息汇聚中心，提供通知公告、便民联络、寻失互助等社区内容场景

### 影响
- Proto: api-proto/api/community/v1/community.proto（已定义）
- 调用方: 无（新服务，暂无外部 gRPC 调用方）
- 数据库: 新增 community_hub_db 库，4 张表
- 关联: go.work 已添加 services/community-hub-service
