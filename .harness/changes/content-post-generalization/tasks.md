# Tasks: 通用图文发布组件重构（content_posts 通用化 + 内容级审核 + Kafka）

> **对执行 Agent 的指令**: 每个 Task 独立可测，按 TDD 执行（先写测试→看失败→写实现→看通过）。精确到文件路径。
> 依赖顺序: Proto → Kafka 基建 → Migration/模型 → scope 基础设施 → 身份派生接线 → 写逻辑 → 读逻辑 → Kafka 推送 → 接口/API → 权限种子 → moderation → 运维验证。迁移/种子为「Owner 运维验证」任务，不走 harness-pipeline（无 DB 环境），派发时显式排除。
> 设计参考: `.harness/changes/content-post-generalization/design.md` + `docs/superpowers/specs/2026-08-16-content-post-design.md`
> **REVISION v3（对照设计评审 data-model v2 / interface-proto v2）**：
> - R1 Design Gate grounded：`community_div` scope_type 在代码库不存在 → `ResolveAdminDivision`（Task 1.7）改经既有 `scope_type='community'` grant 派生 division；`resolvePublishScope`（Task 3.1）改社区管理员角色感知展开。
> - R2 破坏面修正 + REST wire 兼容：web/mobile 存在活跃通知消费方 → REST 响应键保持 `notices`/`notice`/`content`（Task 1.22/1.23）+ 详情 community_id 兼容回退（Task 1.14/1.23）。
> - R5 publisher 档案查询接线缺口 → 新增 Task 1.9（user-service 客户端 + UserRpc 配置）。
>
> **REVISION v4（对照设计评审 data-model v3 / interface-proto v3，本轮闭合 design↔tasks 漂移）**：
> - V4-1 [MUST] UpdateContentPost 授权分流：Task 1.11 消除「作者校验先行使 is_pinned 操作者置顶不可达」——(a) 内容/附件/scope 编辑路径先作者校验 080002；(b) 仅 is_pinned 路径跳过作者校验改验 `PublishRolesFrom` + `AssertCommunitiesScope`（非作者操作者不 080002，scope 不覆盖 080006）。
> - V4-2 [MUST] 详情 community_id 兼容回退改 **scope 反查**（取消 grant 唯一假设）：Task 1.14/1.23 `ResolveReadableCommunityForCompat`（`FindCommunityIdsByPostId` → 逐小区 `FilterAllowed` 任一允许即放行），多小区用户（grid_worker/多房产业主）迁移后详情不 080005。
> - V4-3 [SHOULD] 080004 头注释标签修正：`寻失记录不存在（LostFoundService 仍用）`——Task 0.1 + design.md §Proto。
> - V4-4 [SHOULD] Kafka 推送触发点/顺序显式落位：Task 1.10/1.11「事务提交成功后调用 `Producer.Push`（先提交后推送，提交失败不推送）」。
> - V4-5 [SHOULD] Task 6.2 补 property_admin 调 427/428 → 403 fail-closed 回归断言。
> - V4-6 [INFO] Task 1.23 补 community_ids 含非数字 → 080005 用例；design.md §GetContentPost 注明 scope 外统一 080001（含原 080006 拒绝路径）。
>
> **REVISION v5（本轮，对照设计评审 data-model v4 / interface-proto v4 逐条修订）**：
> - V5-1 [MUST] **UpdateContentPostRequest presence 语义显式化**（interface v4 MUST 1）：Task 0.1 定义 title/text/section_code/is_pinned 用 proto3 `optional`（Go `*string`/`*bool`）；community_ids/attachment_ids 为「全量替换集」+ bool 标志 `has_scope_change`/`has_attachment_change` 区分「未携带=不改」与「空数组=清空」；Task 1.11 分支判定改 presence/标志位（补「取消置顶」「清空全部附件」RED/GREEN 用例）；Task 1.22 REST 类型改 pointer 字段；Task 1.23 代理按 pointer/标志转发。
> - V5-2 [MUST] **is_pinned-only 持久化独立化**（data-model v4 M1）：Task 1.6 新增 `UpdateIsPinned`（独立列更新，不碰 title/text/section_code）+「更新后正文不变」断言；Task 1.11 (b) 分支 draft 与 submitted/approved 统一走 `UpdateIsPinned`，禁复用 `UpdateContent` 传空 title/text；(a) 分支 draft 编辑声明内容字段全量替换语义 + title/text 非空不变量（防「仅改附件」覆盖正文为空）。
> - V5-3 [SHOULD] 详情兼容回退错误码消歧（data-model v4 S1）：Task 1.14/1.23 统一「帖有 scope 但全部不可读 → 080001；帖无任何 scope（数据异常）→ 080005」。
> - V5-4 [SHOULD] Migration 003 部分失败恢复（data-model v4 S2）：Task 6.1 补恢复指引（先 RENAME 回对齐再修重跑；中间态禁直接重跑）。
> - V5-5 [SHOULD] 预期破坏清单补 `content(4)→text(4)` 改名（interface v4 SHOULD 1）：Task 0.1。
> - V5-6 [INFO] 契约收敛：Task 0.1 attachment_ids 统一 int64 JS_STRING + 显式列齐四个响应消息 + Kafka 消费者按 post_id 去重注记；Task 1.13 role 映射收敛 helper.go 单源；Task 1.5 完整性谓词子查询注明走 idx_notice。

---

## 全局 / Proto（由全局 Claude 执行，不分发）

### Task 0.1: community.proto 通用化改名（ContentPostService + 消息契约）
- **文件**: `api-proto/api/community/v1/community.proto`
- [ ] `NoticeService` → `ContentPostService`（D4 直接改名）；`CreateNotice`/`ListNotices`/`GetNotice`/`UpdateNotice`/`DeleteNotice` 直接改名 `CreateContentPost`/`ListContentPosts`/`GetContentPost`/`UpdateContentPost`/`DeleteContentPost`（不做兼容别名）
- [ ] **移除 `UpdateNoticeModerationStatus` RPC（D21）**；`UpdateModerationStatusRequest/Response` 消息**保留**（LostFoundService 仍用）
- [ ] `enum NoticeRole` → `ContentPostRole`（值不变：UNSPECIFIED/COMMUNITY/COMMITTEE/PROPERTY/GRID_OFFICER）
- [ ] `Notice` → `ContentPost`：**保留既有字段号 1-12 不动**（id=1/community_id=2/title=3/publisher=6/publisher_id=7/is_pinned=8/published_at=9/created_at=10/updated_at=11/attachments=12）；`content`(4) → `text`（仅改名，wire 兼容）、`role`(5) 改 `ContentPostRole`（字段号保持 5，publisher 占 6——**勿标 6**）；**新增字段一律追加新号：`section_code`(13, string)、`status`(14, int32 值语义 0=draft/1=submitted/2=approved/3=rejected/4=withdrawn，REVISION)、`attachment_count`(15, int32)**（评审 M1：原 section_code(3)/status(10)/attachment_count(11) 与 title/created_at/updated_at 冲突、role 误标 6，protoc 编译即失败）
- [ ] `NoticeAttachment` → `ContentPostAttachment`：新增 `file_type`(5, string)、`file_id`(6, int64 JS_STRING)、`review_status`(7, int32 值语义 0=pending/1=approved/2=rejected)
- [ ] `CreateContentPostRequest`：`section_code`(1)、`title`(2)、`text`(3)、`entry_status`(4, **int32 值语义 0=draft 默认/1=submitted 立即提交**——**REST/Proto/DB 三侧同号，删除 `ContentPostEntryStatus` 枚举**（评审 M2 根因：REST 1=submitted ↔ proto 1=DRAFT 使立即提交被当 draft）)、`repeated community_ids`(5, int64 JS_STRING)、`repeated attachment_ids`(6, **int64 JS_STRING=file_id——接口 v4 INFO 1 统一载体内型，对齐 `ContentPostAttachment.file_id`**；REST 层 `[]string` 由代理转换)；**无 division_id 字段（REVISION-10/A2）、无 role/publisher_id/publisher 请求字段（JWT 派生）**；本消息为**全新契约、全新字段号**（不沿旧 CreateNoticeRequest 字段号，评审 I6）
- [ ] `ListContentPostsRequest`：`community_id`(1)、`role`(2, ContentPostRole 可选)、`section_code`(3, string 可选)、`page`(4)/`page_size`(5)；`ListContentPostsResponse`：**`common.v1.BaseResp base = 1`**（评审 SHOULD 1——与既有 ListNoticesResponse 等 base=1 对齐，go-zero API 代理依赖 base 包裹错误码）+ `posts`(2, repeated ContentPost) + `total`(3, int64 JS_STRING)
- [ ] `GetContentPostRequest`：`id`(1) + **`community_id`(2, int64 JS_STRING RPC 必填请求上下文，缺失 080005；REST 兼容回退只落薄代理层，RPC 保持严格必填——R2）**
- [ ] `UpdateContentPostRequest`（**V5 presence 语义权威定义——评审 interface v4 MUST 1，仓内先例 permission.proto/user.proto 大量 `optional`**）：`id`(1, int64 JS_STRING)、**`optional string title`(2) / `optional string text`(3) / `optional string section_code`(4)**（proto3 `optional` → Go `*string`，presence 可判：**携带空串=清空，但 title/text 空 → 080005**，见 design §UpdateContentPost）、**`repeated int64 community_ids`(5, JS_STRING 全量替换集) + `bool has_scope_change`(9)**（`true`→全量替换 scope、**空集 → 080005**；`false`→不改）、**`repeated int64 attachment_ids`(6, JS_STRING 全量替换集) + `bool has_attachment_change`(7)**（`true`→全量替换附件、**空集=清空全部附件 attachment_count→0**；`false`→不改）、**`optional bool is_pinned`(8)**（**`*true`=置顶 / `*false`=取消置顶**，双向置位；缺失=不改）、`int32 status`(10, **值语义 0=无提交动作（编辑路径，默认）/1=submit（提交动作）**，RPC 校验：1→submit、0→编辑（仅 draft 可编辑）、其他值→080005；**与 REST 同号，删除 `ContentPostEntryStatus` 枚举**（评审 M2）；**proto 注释标注「action」语义，区别于 `ContentPost.status` 的「state」语义（评审 I2）**)；**分支判定以 presence/标志位为准（指针 nil / 标志 false = 不改），禁止 value 非空启发式**（repeated 无法用 `optional`，故加 bool 标志区分「未携带=不改」与「空数组=清空」）
- [ ] `DeleteContentPostRequest`：`id`(1)
- [ ] **响应消息显式字段（评审 interface v4 INFO 2）**：`CreateContentPostResponse { common.v1.BaseResp base = 1; int64 id = 2 [jstype=JS_STRING]; }`（对齐既有 CreateNoticeResponse）；`GetContentPostResponse { base = 1; ContentPost content_post = 2; }`（REST wire 键 `notice` 由 API 层映射，与移动端 `res.notice` 一致）；`UpdateContentPostResponse { base = 1; }` / `DeleteContentPostResponse { base = 1; }`（对齐既有 Update/DeleteNoticeResponse）
- [ ] 新增 `GetPublishPermissionRequest {}`（空）/ `GetPublishPermissionResponse { base = 1; bool can_publish = 2; repeated ContentPostRole publishable_roles = 3; }`
- [ ] 新增 `GetMarqueeNoticesRequest { int64 community_id = 1 [jstype=JS_STRING]; }` + `ContentPostMarqueeItem { int64 id = 1 [jstype=JS_STRING]; string title = 2; }` + `GetMarqueeNoticesResponse { base = 1; repeated ContentPostMarqueeItem items = 2; }`（命名对齐 ContentPostService，评审 S2；**板块固定 notice，注释注明非通用跑马灯——评审 INFO 1**）
- [ ] 头注释错误码块对齐实际语义：`080001 内容帖不存在 / 080002 无发布权限或非帖作者（功能权限层先于 scope 校验；Update/Delete 为作者归属校验）(评审 S1) / 080003 发布目标数量超限 / 080005 请求参数不合法 / 080006 目标级解析失败或越权`；**保留 080004（寻失记录不存在，LostFoundService 仍用——`CodeLostFoundMiss` types.go:19，唯一使用方 lostfound；contact 逻辑无 080004 引用，勿照旧标签「便民联络不存在/ContactService 仍用」固化漂移——评审 SHOULD 2）**
- [ ] 所有 int64 均 `[jstype=JS_STRING]`（SEE [[proto-jstype]]）；deprecated 是机器可读指令勿当占位注释（SEE [[go-deprecated-directive-not-test-comment]]）
- [ ] **字段号唯一性复核（评审 M1）**：`buf breaking-check`/`buf lint` 确认 ContentPost 消息无重复字段号（1-15 全唯一）；entry_status/status 均为 int32 三侧同号、无 `ContentPostEntryStatus` 枚举残留（评审 M2）
- [ ] **预期破坏清单登记（评审 INFO 3 + interface v4 SHOULD 1）**：记录 `buf breaking-check` 预期 fail 项（ContentPostService 改名 / UpdateNoticeModerationStatus 移除 / `ContentPost.role(5)` 字段类型改 ContentPostRole 的 `FIELD_SAME_TYPE` 类 / enum NoticeRole→ContentPostRole / **`ContentPost.content(4)`→`text(4)` 字段改名——buf WIRE_JSON 类 FIELD_SAME_NAME/FIELD_SAME_JSON_NAME 预期 fail，proto 层 JSON 名改 `text` 属有意，wire 兼容由 REST 层 `json:"content"` 承担**）供 `make ci` 人工核对
- **SEE**: [[proto-jstype]]；D4 破坏性变更（改名+移除回调）影响评估见 design.md §Proto（R2：REST wire 键保持，移动端运行期不破坏）

### Task 0.2: file.proto FileInfo 扩展 + 错误码对齐
- **文件**: `api-proto/api/file/v1/file.proto`
- [ ] 头注释错误码块对齐实际常量（REUSE:notice-D11）：`070001 文件不存在 / 070002 文件访问被拒绝 / 070003 文件操作失败 / 070004 文件类型不支持 / 070005 文件大小超限`；修正漂移的「070002 上传失败 / 070003 文件类型不支持 / 070004 文件大小超限 / 070005 bucket 不存在」
- [ ] `FileInfo` 新增 `string file_type = 11`（白名单规范类型，ConfirmUpload magic-bytes 层产出）
- [ ] `FileInfo` 新增 `bool confirmed = 12`（上传流程完成标志，REQ-CAS-5/REUSE:notice REQ-AS-7）
- [ ] 保持兼容（新字段号），`buf breaking-check` 通过

### Task 0.3: 生成 + CI + CHANGELOG
- **文件**: `api-proto/CHANGELOG.md`
- [ ] 登记 community/v1 全部变更（含 D4 改名 + D21 移除回调 + 消息字段 + **UpdateContentPostRequest presence 语义（title/text/section_code/is_pinned 改 proto3 `optional`、community_ids/attachment_ids 加 `has_scope_change`/`has_attachment_change` bool 标志，V5）**；**标注破坏性 + 如实影响评估（R2）**：web/mobile 存在活跃通知消费方（pages.json tabbar `pages/notice/notice` + notice-detail/notice-browse 注册页；api/community.ts getNoticeList 读 res.notices、getNoticeDetail 读 res.notice 不传 community_id；notice.vue/notice-browse.vue/notice-detail.vue 实际调用）——**REST wire 键保持 + 详情 community_id 兼容回退使移动端运行期不破坏**；web/pc 无消费方；moderation-service 同步移除接线）
- [ ] 登记 file/v1 全部变更（FileInfo +file_type/confirmed + 070004/070005 头注释对齐）
- [ ] **登记附件 file_url 语义**：`ContentPostAttachment.file_url` 为短期预签名 URL（新行占位空串，`file_id` 权威载体重生，勿当永久链接——见 design.md §GetContentPost）
- [ ] `cd api-proto && make ci` → lint 0 errors + breaking-check 通过（community.proto 为预期破坏性，对照 Task 0.1 预期破坏清单人工确认）+ 生成同步
- [ ] 通知受影响服务：community-hub-service / file-service / permission-service / moderation-service（生成代码已本地可见）

### Task 0.4: docker-compose Kafka 基建（单节点 KRaft + content-review topic）
- **文件**: `docker-compose.yml`
- [ ] 新增 `kafka` 服务（image `bitnami/kafka:3.7` 或等位 KRaft 单节点镜像）：`process.roles=broker,controller`、`node.id=1`、controller.quorum.voters=1@kafka:9093、app-network 固定 IP **172.19.0.8**（网段 172.19.0.0/24，避开既有 0.2-0.7）
- [ ] 数据卷持久化 `./data/kafka-data`（容器重启后 broker 健康、数据不丢，REQ-CPM-1 场景）
- [ ] advertised listeners：`kafka:9092`（compose 网络内可解析）+ 本机映射 `9092:9092`；`KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE=false`
- [ ] 创建 `content-review` topic（partition=1、replication=1，经 `KAFKA_CREATE_TOPICS="content-review:1:1"` 或等位 init 脚本）；**retention 覆盖消费者上线空窗**（D17/REVISION——本期无消费者，消息须存活到消费者存在或 pending-push 重推；retention 为配置项，设计约定「存活至消费者或重推」）
- [ ] healthcheck + 依赖服务 `depends_on: kafka: condition: service_healthy` 接线（community-hub 等推送服务在 Kafka 就绪后再启）
- [ ] 端口冲突预检（9092/9093 与既有服务无冲突，SEE [[monorepo-port-management]]）；`docker compose up -d` 冒烟：broker 健康 + topic 存在
- **SEE**: [[monorepo-port-management]]；[[cross-service-seed-deployment-order]]（基建先于服务启动）

---

## community-hub-service

### Task 1.1: Migration 003（notices → content_posts 通用化 + scope + 附件演化 + Kafka 待推列）
- **创建**: `services/community-hub-service/migration/003_content_posts_generalize.sql`
- [ ] **⚠️ 003 为一次性 RENAME 迁移（R4）**：MySQL 无 `RENAME TABLE IF EXISTS`，**仅执行一次，勿重跑**；重跑报错为预期（与 001/002 幂等风格不同属有意，见 design.md §数据模型）
- [ ] `RENAME TABLE notices TO content_posts` + `CHANGE COLUMN content \`text\` TEXT NOT NULL`（反引号包裹防解析歧义，评审 I7）+ 新增 `section_code VARCHAR(30) NOT NULL DEFAULT 'notice'` / `status TINYINT NOT NULL DEFAULT 0`（0=draft/1=submitted/2=approved/3=rejected/4=withdrawn，REVISION）/ `attachment_count INT NOT NULL DEFAULT 0`
- [ ] `ALTER TABLE content_posts MODIFY published_at DATETIME DEFAULT NULL`（D1 审核锚定，去 NOT NULL）
- [ ] `ALTER TABLE content_posts MODIFY community_id BIGINT DEFAULT NULL`（REVISION-9/coverage MUST-1：弃用列去 NOT NULL，迁移先于功能上线门禁）
- [ ] 新增 Kafka at-least-once 待推列：`kafka_push_status TINYINT NOT NULL DEFAULT 0`（0=无待推 1=pending-push 2=已推）/ `kafka_push_retries INT NOT NULL DEFAULT 0` / `kafka_push_last_error VARCHAR(500) DEFAULT NULL` / `kafka_pushed_at DATETIME DEFAULT NULL`（D20）
- [ ] 新建 `content_post_scope` 表：`post_id`+`community_id` 双 NOT NULL + 复合 PK `(post_id, community_id)` + `KEY idx_scope_community (community_id, post_id)`（读索引，REQ-CPB-2）
- [ ] `RENAME TABLE notice_attachments TO content_post_attachments` + `CHANGE COLUMN notice_id post_id BIGINT NOT NULL`（post_id 全链一致）+ 新增 `review_status TINYINT NOT NULL DEFAULT 1`（0=pending/1=approved/2=rejected，本期默认 approved，D14）/ `file_id BIGINT DEFAULT 0`（重生载体）/ `file_type VARCHAR(20) DEFAULT NULL`
- [ ] `ALTER TABLE content_post_attachments MODIFY file_url VARCHAR(1024) NOT NULL`（防 ERROR 1406；新行占位空串，file_id 权威载体）
- [ ] 兼容期保留 `idx_community`/`idx_published`（deprecated，不删）
- [ ] **同步 `services/community-hub-service/docs/design.md` §数据模型**（design_consistency 门禁——model 列 vs 标准迁移源一致）
- **Owner 运维验证**: 三态库执行 + 迁移先于功能上线门禁（见 Task 6.1/6.2）
- **SEE**: [[migration-must-execute]]（提交后必须执行；**003 一次性勿重跑**）；[[unique-index-migration-dup-precheck]] 排除（新表无存量重复，见 design.md「不适用记忆」）

### Task 1.2: ContentPost 模型重构（rename + 字段演化）
- **修改**: `services/community-hub-service/model/notice.go` → rename `content_post.go`（struct `Notice` → `ContentPost`）
- [ ] struct 字段：`Content` → `Text`；`CommunityId int64` → `*int64`（弃用列可空不写入）；`PublishedAt time.Time` → `sql.NullTime`（去 NOT NULL，严禁 `time.Time{}` 零值，SEE [[restore-compensation-zero-time]]）；新增 `SectionCode string` / `Status int64` / `AttachmentCount int64` / `KafkaPushStatus int64` / `KafkaPushRetries int64` / `KafkaPushLastError sql.NullString` / `KafkaPushedAt sql.NullTime`（db tag 对齐迁移 003）
- [ ] `Insert` SQL 移除 `community_id`/`published_at` 两列不写入（弃用/可空列；published_at 由 submit 路径显式写）、显式写 `section_code/status/attachment_count/kafka_push_status`
- [ ] 新增导出常量：`StatusDraft=0 / StatusSubmitted=1 / StatusApproved=2 / StatusRejected=3 / StatusWithdrawn=4`（REVISION 权威契约）+ `KafkaPushNone=0 / KafkaPushPending=1 / KafkaPushDone=2`（供 rpc 层复用）
- **TDD**: 无逻辑代码可不写测试（字段调整由下游查询/写入任务覆盖）；`go build ./...` 编译门禁验证

### Task 1.3: ContentPostScopeModel + ServiceContext 注册
- **创建**: `services/community-hub-service/model/content_post_scope.go` + `content_post_scope_test.go`
- **修改**: `services/community-hub-service/rpc/internal/svc/servicecontext.go`
- [ ] `ContentPostScope` struct（post_id/community_id/created_at，复合主键）+ `ContentPostScopeModel` 接口：`InsertBatch(ctx, postId, communityIds []int64) error`（批量插入，业务层先 dedupe）、`FindCommunityIdsByPostId(ctx, postId) ([]int64, error)`、`DeleteByPostId(ctx, postId) error`（撤回/scope 重写用）
- [ ] ServiceContext 注册 `ContentPostScopeModel`
- [ ] **RED**: table-driven 测试（InsertBatch 正常 + 空列表 + 去重后单行 + FindCommunityIdsByPostId + DeleteByPostId 后查无）→ 确认 FAIL
- [ ] **GREEN**: 实现 → 确认 PASS
- **SEE**: [[snake-camel-field-mismatch]]（db tag 与 Go 字段 snake_case 对齐，prose 引用）

### Task 1.4: ContentPostAttachmentModel + ServiceContext 注册
- **修改**: `services/community-hub-service/model/notice_attachment.go` → rename `content_post_attachment.go`（struct `NoticeAttachment` → `ContentPostAttachment`：`NoticeId` → `PostId` + 新增 `ReviewStatus`/`FileId`/`FileType`）、`services/community-hub-service/rpc/internal/svc/servicecontext.go`
- [ ] `ContentPostAttachment` struct + `ContentPostAttachmentModel` 接口：`InsertBatch(ctx, attachments []*ContentPostAttachment) error`（批量）、`FindByPostId(ctx, postId) ([]*ContentPostAttachment, error)`、`DeleteByPostId(ctx, postId) error`（draft 附件集合重写）
- [ ] ServiceContext 注册 `ContentPostAttachmentModel`
- [ ] **RED**: table-driven 测试（InsertBatch/FindByPostId/DeleteByPostId）→ 确认 FAIL
- [ ] **GREEN**: 实现 → 确认 PASS

### Task 1.5: ContentPostModel 读查询重构（scope JOIN + 审核完整性谓词）
- **修改**: `services/community-hub-service/model/content_post.go` + `content_post_test.go`（rename 自 notice_test.go）
- [ ] **新增共享谓词** `func IsReviewComplete(status int64, approvedAttachments, attachmentCount int64) bool`（正文 `status==StatusApproved` 且 `approvedAttachments==attachmentCount`，REQ-CPB-8 读路径单源谓词，导出供 rpc 层复用）
- [ ] `FindListByCommunity(ctx, communityId int64, sectionCode, role string, offset, limit int64) ([]*ContentPost, int64, error)` — `content_posts JOIN content_post_scope ON content_posts.id=content_post_scope.post_id`，**显式投影 `content_posts.*` + `content_post_scope.community_id`（右表限定，防 `select *` 双 community_id 列取到弃用 NULL）**，`scope.community_id=?` + `content_posts.deleted_at IS NULL` + 可选 `section_code=?`/`role=?` + **`status=2`（正文 approved 谓词前置）+ 附件完整性子查询（count(attachments WHERE review_status=approved)==attachment_count，关联标量子查询走 `idx_notice(post_id)`，注释注明勿退化为全表扫描——评审 data-model v4 I4）** + `order by is_pinned desc, published_at desc`（NULLS LAST 防御） + 分页
- [ ] `FindOneReviewComplete(ctx, id int64) (*ContentPost, error)`（替代 `FindOnePublished`：deleted_at IS NULL + status=2 + 附件完整性谓词）
- [ ] `FindMarquee(ctx, communityId int64, since time.Time, limit int64) ([]*ContentPost, error)` — JOIN scope、status=2 + 附件完整性、`published_at >= since`（含端点）、`order by is_pinned desc, published_at desc`、`limit 10`
- [ ] **RED**: table-driven 测试（正常 + 分页边界 + section_code/role 筛选 + scope 外小区查无 + **附件 rejected → 谓词不成立不返回** + 已审附件数<attachment_count 不返回 + 无附件帖返回 + withdrawn/deleted 不返回 + status=2 且完整返回 + NULL published_at 排最后）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[moderation-status-write-without-read-gating]]（读路径恒应用完整性谓词，不 mutate status）

### Task 1.6: ContentPostModel 写路径（Insert 显式状态 + UpdateIsPinned 独立列更新 + Withdraw + kafka_push 状态）
- **修改**: `services/community-hub-service/model/content_post.go` + `content_post_test.go`
- [ ] `Insert(ctx, n *ContentPost) (int64, error)` — 显式写 `section_code/status/attachment_count/kafka_push_status`，community_id/published_at 不写（published_at 由 submit 路径显式传参）
- [ ] `UpdateContent(ctx, id int64, title, text, sectionCode string) error`（draft 正文编辑三列 title/text/section_code，不碰 status；**is_pinned 不再由本方法写——一律走 `UpdateIsPinned`**，防「仅改附件/scope 传空 title/text」覆盖正文，评审 data-model v4 M1(c)）
- [ ] **`UpdateIsPinned(ctx, id int64, isPinned int32) error`（V5 新增——评审 data-model v4 M1 修复）**：**独立列更新 `UPDATE content_posts SET is_pinned=? WHERE id=? AND deleted_at IS NULL`，仅写 is_pinned 列，不碰 title/text/section_code**——is_pinned-only 路径（(b) 分支 draft/submitted/approved 置顶/取消置顶）一律走本方法；**禁止复用 UpdateContent 传空 title/text（会把已发布帖正文清空，数据丢失）**
- [ ] `UpdateStatusAndPublish(ctx, id int64, status int64, publishedAt time.Time) error`（submit：status=2 + published_at=NOW() 原子；与 `kafka_push_status=1` 同事务，D16/D20）
- [ ] `Withdraw(ctx, id int64) error`（软删 + `status=withdrawn(4)`，单事务由逻辑层组合，REQ-CPB-10）
- [ ] `UpdateKafkaPushStatus(ctx, id int64, pushStatus, retries int64, lastErr sql.NullString, pushedAt sql.NullTime) error`（重推扫描器回调）
- [ ] `FindPendingPush(ctx, limit int64) ([]*ContentPost, error)`（重推扫描：`kafka_push_status=1` 且 `deleted_at IS NULL`）
- [ ] **RED**: 测试（Insert 显式 status/attachment_count + community_id/published_at 不写；UpdateContent 不动 status/is_pinned；**UpdateIsPinned 置顶(1)与取消置顶(0)后再读回：is_pinned 为新值且 title/text/section_code 与更新前一致（正文不变断言）**；UpdateStatusAndPublish 置 2+NOW；Withdraw 置 withdrawn；UpdateKafkaPushStatus 各字段；FindPendingPush 只挑 pending）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[restore-compensation-zero-time]]（published_at sql.NullTime，零值禁写）

### Task 1.7: scope 包多目标校验 + division 展开（AssertCommunitiesScope / ExpandDivisionCommunities / ResolveAdminDivision，R1 grounded）
- **修改**: `services/community-hub-service/rpc/internal/logic/scope/scope.go`
- **创建**: `services/community-hub-service/rpc/internal/logic/scope/division.go`、`services/community-hub-service/rpc/internal/logic/scope/scope_test.go`
- [ ] `AssertCommunitiesScope(ctx, client, userID, targets []int64) error` — **单次批量 `AssertPublishScope`**（一次携带全部 target ScopeRef；任一越权/不可解析 → 060007→080006 一次映射整体拒绝，all-or-nothing；未知节点 fail-closed，REUSE:notice-D31；**community_admin 的 division 子树目标由 Task 3.1 的 `resolvePublishScope` 角色感知展开覆盖**）
- [ ] `ExpandDivisionCommunities(ctx, mdClient, divisionID int64) ([]int64, error)` — **先 guard `divisionID<=0 → 080005`**（fail-closed，杜绝进入 masterdata 默认分支过度展开）；调 masterdata `GetResidentialAreasByDivision(community_div_id=divisionID, status=1)`，返回 approved 小区 id 列表；展开空 → 080005（REVISION-10/A2 语义：community_admin 唯一管辖 division 展开为空）
- [ ] **`ResolveAdminDivision(ctx, permClient, mdClient, userID) (int64, error)`（R1 重写 + 评审 M2 状态过滤——不再依赖不存在的 community_div scope_type，且防过期/驳回 grant 越权）**：经 permission `GetUserRoles(user_id)` 取 **role_code=community_admin 且 scope_type='community' 且 scope_id!=0 且 URStatus==2（level-2 等价：status==2 且 verified_at NOT NULL 且未过期，与 421 写路径门槛同语义，禁止直读 rel_user_role）** 的 grant 列表（scope_id 为 communityId 集）→ 逐条 masterdata `GetResidentialArea(scope_id).community_div_id` → 收集 **distinct division 集**；**空 → 080005（非 admin 走 community_ids 直传路径）；>1 个 distinct division → 080005（「唯一管辖」契约守卫，fail-closed，评审 I4 语义保留）** → 返回唯一 division。**已过期(4)/已驳回(3)的 community_admin grant 不计入（URStatus 过滤）——即使该用户另有 level-2 发布角色（committee/grid_worker 也持 421），也不能用失效的 community_admin grant 驱动 division 展开（评审 M2 权限提升修复）**
- [ ] **RED**: 测试（单目标越权→080006；多目标任一越权→整体 080006；目标不存在→080006 fail-closed；division<=0→080005；展开空→080005；ResolveAdminDivision：单 community grant → 经 GetResidentialArea 得唯一 division / 无 community_admin grant → 080005 / **两个 community grant 映射到不同 division → 080005** / **两个 community grant 映射到同一 division → 合并为唯一 division 放行** / GetResidentialArea 传输错误 → fail-closed / **过期(4) community_admin grant + 另一 level-2 角色 → 不展开（080005 或按有效 grant 直传）** / **驳回(3) grant 不计入**）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[grpc-timeout-layers]]（AssertPublishScope/GetUserRoles/GetResidentialArea 内嵌跨服务 RPC 三层超时对齐）；[[rpc-identity-spoofing-loopback-isolation]]（user_id 经 metadata 不信任 body）；[[grpc-only-comms]]（禁止直读 rel_user_role）

### Task 1.8: 发布角色派生（PublishRolesFrom + RBAC→role 映射）
- **修改**: `services/community-hub-service/rpc/internal/logic/scope/userctx.go`
- **创建**: `services/community-hub-service/rpc/internal/logic/scope/userctx_test.go`
- [ ] `PublishRolesFrom(ctx, permClient, userID) ([]string, error)` — 经 `GetUserRoles` 取实际持有发布角色 code（grid_worker/community_admin/committee/property_admin，D6），供 role 列映射派生（REVISION REQ-CPB-5：JWT 仅含 user_id，必须显式 GetUserRoles）；多角色按发布角色优先序返回
- [ ] `PublishRoleToString(roleCode string) string` — RBAC code → role 列值映射（grid_worker→grid_officer、community_admin→community、committee→committee、property_admin→property）
- [ ] **RED**: 测试（多角色优先序 + 无发布角色空集 + GetUserRoles 传输错误 fail-closed + role 映射各分支）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[grpc-only-comms]]（经 GetUserRoles，禁止直读 rel_user_role）

### Task 1.9: publisher 展示名档案查询接线（user-service 客户端 + UserRpc 配置，R5 新增）
- **修改**: `services/community-hub-service/rpc/internal/config/config.go`（+`UserRpc zrpc.RpcClientConf`）、`services/community-hub-service/rpc/etc/communityhub.yaml`（+`UserRpc:` 段指向 user-service）、`services/community-hub-service/rpc/internal/svc/servicecontext.go`（注册 `userv1.UserServiceClient`）
- [ ] ServiceContext 新增 `UserClient userv1.UserServiceClient`（经 `zrpc.NewClient(c.UserRpc)`，与既有 Permission/MasterData 客户端同模式；community-hub 现无 UserClient——仅 Moderation/Permission/MasterData/SysConfig，评审 SHOULD 3）
- [ ] config 结构 + yaml 接线；`go build ./...` 通过
- [ ] 冒烟：GetUsersByIds 可调（真实 user-service 已提供该 RPC，纯接线无 Proto 变更）
- **SEE**: [[grpc-only-comms]]（档案经 gRPC 查询，禁请求体信任）；[[cross-service-seed-deployment-order]]（接线纳入部署编排）

### Task 1.10: CreateContentPostLogic（发布主路径：多小区 + 附件绑定 + 单事务 + 入口状态）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/createnoticelogic.go` → rename `createcontentpostlogic.go` + `createcontentpostlogic_test.go`、`services/community-hub-service/rpc/internal/logic/notice/helper.go`（roleToString 保留 + toProtoContentPost 适配）
- [ ] 校验顺序落地（design.md §CreateContentPost）：section_code ∈ 板块集（`notice`）→ 080005；title/text 非空 → 080005；community_ids 去重、空 → 080005
- [ ] **目标集解析（R1）**：grid_worker/committee/property_admin 直传 community_ids；community_admin → `ResolveAdminDivision`（Task 1.7，经既有 community grant 派生唯一 division）+ `ExpandDivisionCommunities` 展开（快照）；**展开后长度 >100 → 080003**（REVISION 按展开快照计量）
- [ ] 附件绑定（REQ-CPB-6 单源）：逐 attachment_id 调 file `GetFileUrl(file_id)` → `FileInfo.confirmed==true` 且 `user_id==JWT`；数量 ≤10；`Σ file_size ≤ 50MB` → 否则 080005（附件引用无效/超限）；`file_type`/`file_name`/`file_size`/`download_url` 自 FileInfo 回读
- [ ] 数据权限：`AssertCommunitiesScope(user, targets)`（单次批量，Task 1.7）→ 080006
- [ ] `publisher_id`/`role`/`publisher` 从身份派生（REVISION REQ-CPB-5）：`publisher_id`=JWT user_id；`role`=`PublishRolesFrom` 映射（Task 1.8）；`publisher`=用户真实档案展示名（**经 Task 1.9 的 `UserClient` 调 `GetUsersByIds` 查询**；禁请求体信任，堵展示名伪造向量）
- [ ] **单事务落库**：content_posts（community_id=NULL、section_code、title/text、status=入口（draft=0 / submitted=2）、attachment_count、is_pinned=0、role/publisher/publisher_id 派生、published_at=submitted?NOW():NULL、**kafka_push_status=submitted?1:0**）+ `ContentPostScopeModel.InsertBatch` + `ContentPostAttachmentModel.InsertBatch`（file_id=attachment_ids、file_type/file_name/file_size 自 FileInfo、file_url=占位空串、review_status=1）
- [ ] **entry=submitted 立即提交**：同事务置 status=2 + published_at=NOW() + kafka_push_status=1（隐式通过 D16）→ **事务提交成功后调用 Task 1.18 `Producer.Push`（先提交后推送，提交失败不推送——评审 data-model v3 S1 触发点/顺序显式落位）**
- [ ] **不再 LPUSH Redis `moderation:task:queue`（D3，content_posts 只走 Kafka）**——移除 CreateAuditLog + LPUSH 推送（停推逻辑独立 Task 1.20，本任务仅确保 Create 主路径不依赖 Redis 入队）
- [ ] **RED**: table-driven 测试（正常 draft + 正常 submitted 立即通过 + section_code 非法 080005 + title/text 空 080005 + 空范围 080005 + community_admin 展开成功/展开空 080005/展开>100 080003 + 越权 080006 + 目标不存在 080006 + 附件未确认/他人文件 080005 + 附件超量/超总 080005 + role/publisher 由 JWT/档案派生断言（伪造 body 被纠正）+ submitted 落 kafka_push_status=1）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[grpc-only-comms]]（附件校验经 file GetFileUrl、档案经 user GetUsersByIds，不直连 uploaded_file/user_base）

### Task 1.11: UpdateContentPostLogic（draft 编辑 + attachment_count 重算 + is_pinned + submit，V5 presence 语义）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/updatenoticelogic.go` → rename `updatecontentpostlogic.go` + `updatecontentpostlogic_test.go`
- [ ] `FindOne(id)` 未找到 → 080001
- [ ] **presence 输入语义（V5 权威定义——评审 interface v4 MUST 1）**：请求字段区分「未携带=不改」与「携带（含空串/空集/false）」——`Title/Text/SectionCode/IsPinned` 为 **proto3 `optional` 生成 `*string`/`*bool`（指针 != nil = 携带）**；附件/scope 以 `HasAttachmentChange`/`HasScopeChange` bool 标志判别（**`false`=不改，`true`=全量替换集，空集=清空**）。**分支判定以 presence/标志位为准，禁止 value 非空启发式**（proto3 标量无 presence）
- [ ] **授权分流（评审 data-model v3 M1 / interface v3 MUST 1 + V5 修订——消除「作者校验 vs 操作者置顶」互斥，REQ-CPB-9(f) 能力可达）**：按请求形状分流——
  - **(a) 内容/附件/scope 编辑路径**（`Title!=nil` 或 `Text!=nil` 或 `SectionCode!=nil` 或 `HasScopeChange==true` 或 `HasAttachmentChange==true`，或 `status==1`）→ **先作者校验**（`publisher_id == JWT user_id`，非发布者 → **080002**），再按 draft/非 draft 走 080005（仅 draft 可内容编辑）。
  - **(b) 仅 is_pinned 路径**（`IsPinned!=nil` 且 Title/Text/SectionCode 均 nil、HasScopeChange/HasAttachmentChange 均 false、status==0）→ **跳过作者校验**，改验操作者授权：draft 帖 → 发布者即可；submitted/approved 帖 → `PublishRolesFrom` 非空 且 `AssertCommunitiesScope`（数据范围覆盖帖小区），满足即放行（**非作者操作者不 080002**），scope 不覆盖 → **080006**；**`*true` 置顶 / `*false` 取消置顶均走此路径（presence 可判，取消置顶确定性可达——评审 MUST 1(a)）**
- [ ] **is_pinned 持久化（V5 修订——评审 data-model v4 M1 修复）**：置顶/取消置顶统一走 Task 1.6 `UpdateIsPinned`（独立列更新，draft 与 submitted/approved 均如此，**不碰 title/text/section_code**）——**禁止复用 UpdateContent 传空 title/text（会把已发布帖正文清空，数据丢失）**；draft 编辑请求含 is_pinned 时同样经 `UpdateIsPinned`
- [ ] **draft 编辑**（status==draft(0)，走 (a) 分支）：
  - **内容字段全量替换语义（V5 显式声明——评审 MUST 1(c)）**：`Title`/`Text`/`SectionCode` 任一 presence → 对应列按请求值覆盖；**携带空串 → 080005（title/text 非空不变量，与创建同规则）**；未携带列保持现值——**「仅改附件/scope」时 Title/Text/SectionCode 均 nil → 正文列不动（不得传空串覆盖）**
  - 附件集合变更（`HasAttachmentChange==true`）→ **`ContentPostAttachmentModel.DeleteByPostId` + `InsertBatch` 全量替换 + 同事务重算 `attachment_count`（新绑定数；**空集=清空全部附件 → attachment_count=0**，D19 不变量可归零——评审 MUST 1(b)）** + 复跑完整绑定校验（REQ-CPB-6）→ 080005 超限整体拒绝
  - scope 变更（`HasScopeChange==true`）→ 复跑 `AssertCommunitiesScope`（新目标集，任一越权 → 080006；**空集 → 080005**，帖必须 ≥1 scope 小区）+ `ContentPostScopeModel` 重写
- [ ] **submit 动作**（`status==1`，走 (a) 分支）：仅 draft 可提交；同事务 `UpdateStatusAndPublish(id, approved, NOW())` + 置 `kafka_push_status=1` → **事务提交成功后调用 Task 1.18 `Producer.Push`（先提交后推送，提交失败不推送——评审 data-model v3 S1 触发点/顺序显式落位）**；`status` 其他值（≠0/1）→ 080005（审核结果仅审核流写入）；**并发重复 submit（两请求同读 draft）可双推 content-review——at-least-once 容忍，后续消费者按 `post_id` 幂等去重（评审 INFO I3），不设防重锁**
- [ ] **不再 LPUSH（D3，评审 M3）**：`updatecontentpostlogic.go`（rename 自 updatenoticelogic.go）**整体移除**原 `ModerationClient.CreateAuditLog`（~L72）+ `RedisClient.LpushCtx(ctx, "moderation:task:queue", ...)`（~L97）块——submit 路径只推 Kafka（Task 1.18），**不既推 Kafka 又 LPUSH Redis**；移除 moderationv1 client / RedisClient 未用依赖引用；`go build ./...` 通过
- [ ] **非 draft 不可内容编辑**：`status != draft(0)` 的**内容/附件/scope 变更**（非 is_pinned-only）→ **080005（仅 draft 可编辑）**；`attachment_count`/scope 不变（all-or-nothing）
- [ ] **RED**: table-driven 测试（draft 编辑成功 + attachment_count 重算（删 1 剩 1 → count=1）+ **清空全部附件（HasAttachmentChange=true + 空 attachment_ids → 附件行全删 + attachment_count=0，评审 MUST 1(b)）** + **取消置顶（IsPinned=*false on approved 帖 → is_pinned=0 且 title/text 不变，经 UpdateIsPinned 断言，评审 MUST 1(a)）** + 新增附件超限 080005 整体不变 + **title/text 携带空串 → 080005（评审 MUST 1(c)）** + scope 越权 080006 + **scope 空集（HasScopeChange=true + 空 community_ids）→ 080005** + submit 成功（status=2+published_at=NOW+kafka_push_status=1 + **moderation:task:queue 无新元素——不调用 LPUSH，D3/M3** + **提交成功后调用 Producer.Push 断言**）+ submitted/approved 内容编辑 080005 + status 非法值 080005 + **is_pinned-only 分支：draft 发布者置位成功（正文不变断言）/ 非发布者操作者（持发布角色 + scope 覆盖）置顶 approved 帖成功（不 080002，REQ-CPB-9(f)）/ scope 不覆盖 → 080006 / 请求含 is_pinned+内容字段 → 走 (a) 作者校验 080002**）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.12: DeleteContentPostLogic（撤回，仅发布者本人）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/deletenoticelogic.go` → rename `deletecontentpostlogic.go` + `deletecontentpostlogic_test.go`
- [ ] `FindOne(id)` 未找到 → 080001
- [ ] 作者校验：`publisher_id == JWT user_id`，否则 **080002**（仅发布者本人，REUSE:notice-D19）
- [ ] 可用状态：draft/submitted/approved 均可删（REVISION——draft 可删、submitted 不可编辑但可删）
- [ ] **单事务**：`conn.Transact(func(session sqlx.Session) error)` 传共享 session——`Withdraw`（软删 + status=withdrawn(4)）；**content_post_scope 行与附件行全部保留**（帖的撤回由主表软删+withdrawn 表达，REQ-CPB-10）；**不推 Kafka**
- [ ] **RED**: 测试（发布者撤回全局生效 + 非发布者 080002 + 不存在 080001 + 附件/scope 保留断言 + withdrawn 后读路径不可见 + **失败注入测试：Withdraw 报错 → 整体回滚、无半态**）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.13: ListContentPostsLogic（scope 过滤 + 板块/role 筛选）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/listnoticeslogic.go` → rename `listcontentpostslogic.go` + `listcontentpostslogic_test.go`、`services/community-hub-service/rpc/internal/logic/notice/helper.go`（`toProtoContentPost`：CommunityId 用 *int64/scope 注入派生、PublishedAt sql.NullTime null 感知、Status/AttachmentCount/SectionCode 填充）
- [ ] `FilterAllowed`（GLOBAL/LIMITED/EMPTY 语义保留）→ false 返回空列表
- [ ] `FindListByCommunity`（section_code/role 筛选 + 完整性谓词）+ `order by is_pinned desc, published_at desc`（NULLS LAST）+ 分页
- [ ] 响应每条 `community_id`=请求小区（scope 派生，不读弃用列）；role 筛选保留 notice 兼容语义，**枚举→DB 列映射收敛 helper.go 单一函数（V5，评审 data-model v4 I2——与写侧 `PublishRoleToString` 同一字符串集合，防两份映射漂移）**
- [ ] **RED**: 测试（正常 + 分页边界 + section_code/role 筛选 + 越权空列表 + 多小区帖只在其发布小区出现 + **JOIN 投影断言：C2 列表返回 community_id=C2** + 未完整帖不返回）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.14: GetContentPostLogic（community_id 上下文 + scope 校验 + 附件重生 + REST 兼容回退）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/getnoticelogic.go` → rename `getcontentpostlogic.go` + `getcontentpostlogic_test.go`
- [ ] `community_id` **RPC 契约必填**：缺失/空 → 080005
- [ ] **REST 兼容回退（R2，评审 interface v3 MUST 2 修复——取消 grant 唯一假设，改 scope 反查）**：RPC 层保持 `community_id` **严格必填**（缺失/空 → 080005）；兼容回退只落 Task 1.23 的 REST handler。本逻辑任务提供 `ResolveReadableCommunityForCompat(ctx, contentPostModel, permClient, postId) (int64, error)` 辅助函数 + 测试：`FindOneReviewComplete(postId)` 未找到 → 080001 → `ContentPostScopeModel.FindCommunityIdsByPostId(postId)` 取帖所属小区集 → 对每小区 `FilterAllowed(userID, community_id)`（与现网 getnoticelogic.go 反查 `community_id → FilterAllowed` 的 LIMITED 语义一致）——**任意小区允许即放行（多小区 grid_worker / 多房产业主迁移后详情页仍可用，不 080005）**；全部不读取 → 080001（不泄露，与 RPC 层未过审同）；帖无任何 scope 小区（数据异常）→ 080005
- [ ] `FindOneReviewComplete(id)` 未找到 → 080001；scope 匹配 `(id, community_id)` 缺失 → 080001；`FilterAllowed(userID, community_id)` false → 080001（scope 外/不存在/未完整统一 080001，不泄露）
- [ ] 响应 `ContentPost.community_id`=请求小区（scope 派生，不读弃用列）；attachments 携带 `review_status`/`file_type`/`file_id`
- [ ] **附件 file_url 重生**：逐附件按 `content_post_attachments.file_id` 调 file `GetFileUrl(file_id)` 重生预签名 URL（file_type/file_name/file_size 同源回读）；兼容期 file_id=0/NULL 行回退 stored file_url（防御，存量帖本期不可达）
- [ ] **RED**: 测试（正常含附件 review_status/file_type/file_id + **附件 file_url 为 GetFileUrl 重生后的新预签名 URL 断言** + 缺失 community_id 080005 + 不存在/scope 外/未完整 080001 + file_id=0 回退 stored file_url + **`ResolveReadableCommunityForCompat`：帖 scope 多小区中任一可读 → 该小区（多小区用户不 080005）/ **全部不可读 → 080001（V5 消歧，评审 data-model v4 S1——与 RPC 层 scope 外统一 080001 一致，不泄露）** / 帖无 scope（数据异常）→ 080005 / FindOneReviewComplete 未找到 → 080001 / FilterAllowed 传输错误 fail-closed**）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.15: GetMarqueeNoticesLogic（新建，跑马灯）
- **创建**: `services/community-hub-service/rpc/internal/logic/notice/getmarqueenoticeslogic.go` + `getmarqueenoticeslogic_test.go`
- [ ] `community_id` 必填：缺失/空/0 → 080005
- [ ] `FilterAllowed` → false 返回空列表（与列表读路径一致）
- [ ] `FindMarquee(community_id, now-15*24h, 10)`：置顶优先 + published_at 倒序 + 封顶 10 + 完整性谓词；15 天边界含端点（REUSE:notice-D32）；板块固定 notice（无 section_code 入参，评审 INFO 1）
- [ ] 输出 `ContentPostMarqueeItem[]{id, title}`；空态返回空列表
- [ ] **RED**: 测试（正常 ≤10 + 置顶优先 + 倒序 + 15 天边界含端点 + >15 天排除 + 未完整排除 + 空态 + 缺失 080005）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 1.16: GetPublishPermissionLogic（can_publish + 可发布角色含 property_admin）
- **创建**: `services/community-hub-service/rpc/internal/logic/notice/getpublishpermissionlogic.go` + `getpublishpermissionlogic_test.go`
- [ ] `userID = UserIDFromCtx`；0 → can_publish=false（防御，认证中间件兜底）
- [ ] 调 permission `GetUserRoles(user_id)`；level-2 判定：`role.Code ∈ {grid_worker, community_admin, property_admin, committee}`（D6——property_admin 保留）且 `status==2` 且 `verified_at>0` 且 `expires_at==0 OR expires_at>now`（基于 RPC 输出，禁止直读 rel_user_role，SEE [[grpc-only-comms]]）
- [ ] 命中 → can_publish=true + 映射 `ContentPostRole`（grid_worker→GRID_OFFICER、community_admin→COMMUNITY、committee→COMMITTEE、property_admin→PROPERTY）
- [ ] owner/tenant/merchant/sys_admin → can_publish=false
- [ ] **RED**: 测试（网格员已认证通过 + property_admin 通过（D6）+ committee 通过 + status=2 但 verified_at NULL 拒绝 + 角色过期拒绝 + owner/tenant 拒绝 + GetUserRoles 传输错误 fail-closed）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[auto-grant-unverified-grant-confers-scope-level0]]（min_verf_level=2 与 level-2 判定一致）

### Task 1.17: Kafka 配置接入（config + yaml + 依赖）
- **创建**: `services/community-hub-service/rpc/etc/communityhub.yaml`（Kafka 配置段：`Kafka: {Brokers: [kafka:9092], Topic: content-review}`）
- **修改**: `services/community-hub-service/rpc/internal/config/config.go`（+`Kafka KafkaConf`：brokers []string / topic string / retry interval）、`services/community-hub-service/go.mod`（+`github.com/segmentio/kafka-go`，go mod tidy）
- [ ] 配置结构定义 + yaml 接线 + 依赖引入；`go build ./...` 通过
- [ ] 说明：Kafka 配置为「配置修改独立任务」（不并入业务逻辑），brokers 指向 docker-compose `kafka:9092`

### Task 1.18: Kafka Producer + content-review 契约 + at-least-once 推送
- **创建**: `services/community-hub-service/rpc/internal/kafkapush/producer.go` + `producer_test.go`
- **修改**: `services/community-hub-service/rpc/internal/svc/servicecontext.go`（初始化 Producer）
- [ ] `ContentReviewMessage` struct（JSON 契约 REQ-CPM-2 单源）：`version int32=1` / `post_id int64 string` / `section_code` / `text` / `publisher_id int64 string` / `attachments []{file_id, file_type, review_status, file_url}`——**file_url 为可再生预签名 URL（发布时经 GetFileUrl 生成）**；无附件帖推 `attachments: []`（空数组非 null）
- [ ] `Producer.Push(ctx, post *model.ContentPost, attachments []*model.ContentPostAttachment) error` — 打包 JSON → `segmentio/kafka-go` Writer 推送 `content-review` topic（acks=all）；**成功 → `UpdateKafkaPushStatus(2, ...)`；失败 → 置 `kafka_push_status=1` + `kafka_push_last_error` 落库（D20，不阻塞发布，status 已 approved 可见）**
- [ ] **RED**: 测试（成功推送后 UpdateKafkaPushStatus(2) + 失败后置 pending + last_error 记录 + **contract 字段断言（version/post_id string/section_code/text/publisher_id/attachments 含 file_url + review_status 快照）** + 空附件数组 + push 不 panic）→ FAIL（producer 用可 mock 的 Writer 接口）
- [ ] **GREEN**: 实现 → PASS
- **集成验证**: 真实 Kafka 投递见 Task 6.2（Owner 运维验证）
- **SEE**: [[best-effort-compensation-must-log]]（推送失败补偿不可静默丢弃）；[[grpc-max-msg-size-sensitive-words]]（未来消费者经 Kafka 拉取，不受 gRPC 4MB 限制）

### Task 1.19: Kafka Rescanner（定时重推，at-least-once）
- **创建**: `services/community-hub-service/rpc/internal/kafkapush/rescanner.go` + `rescanner_test.go`
- [ ] `Rescanner`（定时 Ticker，如 1 分钟，启动 goroutine 随 servicecontext 接线）：`FindPendingPush` → 逐条**复用 Task 1.18 `Producer.Push`**（内含 GetFileUrl 重生 file_url，避免重推携带过期预签名 URL，评审 I8）→ ack 置 2
- [ ] 重试超阈值（如 3 次）→ 保留 pending + 日志/quarantine 语义（BACKLOG 回填「消费者上线前 pending-push 积压处置」）；**pending-push 计数可观测**（指标/日志）
- [ ] **RED**: 测试（pending→重推→ack 置 2 + 推送失败继续 pending + 超阈值保留 + 计数可观测）→ FAIL（rescanner 用 sqlmock + mock producer）
- [ ] **GREEN**: 实现 → PASS
- **SEE**: [[best-effort-compensation-must-log]]（重推失败不静默丢弃）

### Task 1.20: 停 content_posts 的 Redis 队列推送（D3，Create + Update/submit 双路径）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/createcontentpostlogic.go` + `updatecontentpostlogic.go`（Create 路径本任务移除；**Update/submit 路径的移除在 Task 1.11 落**，本任务负责双路径 RED 断言）
- [ ] 移除 CreateContentPost 中对 `CreateAuditLog`（moderationv1）与 `RedisClient.LpushCtx(ctx, "moderation:task:queue", ...)` 的调用（D3——content_posts 只走 Kafka，不再 LPUSH Redis）
- [ ] **Update/submit 路径**：`updatecontentpostlogic.go` 的 CreateAuditLog + LPUSH 块已在 Task 1.11 整体移除（评审 M3——submit 不再「既推 Kafka 又 LPUSH Redis」）；本任务交叉验证 updatecontentpostlogic.go 无残留
- [ ] 确认 lostfound/user 等其他来源的 Redis 入队流程**保持不变**（D3 双轨：不动 createlostfoundlogic 等）
- [ ] 移除 `createcontentpostlogic.go` 中 moderationv1 client / RedisClient 的未用依赖引用；`go build ./...` 通过
- [ ] **RED→GREEN**: 断言 **CreateContentPost 与 UpdateContentPost(submit) 成功落库后 `moderation:task:queue` 均无新元素**（Redis mock 断言两路径均不调用 LPUSH）→ 实现 → PASS（评审 M3：断言扩展为双路径）

### Task 1.21: RPC server 注册 ContentPostService 新方法
- **修改**: `services/community-hub-service/rpc/internal/server/communityhubserver.go`
- [ ] 方法改名：`CreateNotice`→`CreateContentPost`、`ListNotices`→`ListContentPosts`、`GetNotice`→`GetContentPost`、`UpdateNotice`→`UpdateContentPost`、`DeleteNotice`→`DeleteContentPost`；委托对应 Logic
- [ ] 新增 `GetPublishPermission` / `GetMarqueeNotices` 方法委托对应 Logic
- [ ] 移除 `UpdateNoticeModerationStatus` 方法（D21——RPC 已从 proto 移除；`updatemoderationstatuslogic.go` 删除）
- [ ] `go build ./...` 通过

### Task 1.22: API types.go 同步（ContentPost 类型 + 新字段 + form 标签 + R2 wire 兼容）
- **修改**: `services/community-hub-service/api/internal/types/types.go`
- [ ] **R2 wire 兼容（评审 MUST）**：REST 响应 **wire 键保持既有**——`ListContentPostsResp` JSON 键保持 `notices`（数组）+ `total`；`GetContentPostResp` JSON 键保持 `notice`（单对象）；帖体正文 JSON 键保持 `content`（**Go 字段 `Text string json:"content"`**——proto/DB 用 `text`，REST wire 用 `content`，显式映射，见 design §REST wire 兼容/ADR）；`role`/`publisher`/`publisher_id`/`is_pinned`/`published_at`/`created_at`/`updated_at`/`community_id`/`attachments[]`（`id/file_name/file_url/file_size`）键全部保持
- [ ] `CreateNoticeReq` → `CreateContentPostReq`：`SectionCode string`（json:"section_code"）、`Title`、`Text`（json:"content"——**REST 请求 wire 键保持 content 与移动端旧调用一致**）、`EntryStatus int32`（json:"entry_status"：0=draft 默认/1=submitted——**与 proto int32 同号，无数值变换**，评审 M2）、`CommunityIds []string`（json:"community_ids"——**REST 层 string 形式 Snowflake ID，不能 []int64**，encoding/json `,string` 不支持 slice，SEE [[proto-jstype]]）、`AttachmentIds []string`（json:"attachment_ids"）、`IsPinned bool`（json:"is_pinned,optional"）；移除 `Role/Publisher/PublisherId`（服务端派生）
- [ ] `ListNoticesReq` → `ListContentPostsReq`：`CommunityId int64 form:"community_id"`、`SectionCode string form:"section_code,optional"`、`Role int32 form:"role,optional"`、Page/PageSize；`ListContentPostsResp`：`Notices []ContentPostInfo json:"notices"` + `Total int64 json:"total,string"`
- [ ] `GetContentPostReq`：`Id int64 path:"id"` + **`CommunityId int64 form:"community_id,optional"`**（GET query 必须用 form 标签，禁 json/path——go-zero ParseForm skip 无 form 标签字段 → 恒 0 → 080005 全挂；**optional 供 R2 兼容回退，RPC 层仍必填**）
- [ ] `GetContentPostResp`：`Notice ContentPostInfo json:"notice"`
- [ ] `UpdateContentPostReq`（**V5 presence 语义——评审 interface v4 MUST 1/S3，pointer/标志位可判；仓内先例 user-service `*string json:"...,optional"` / master-data `*int64`**）：`Id int64 path:"id"` + `Title *string json:"title,optional"` / `Text *string json:"content,optional"`（REST wire 键保持 content，R2；**指针 != nil = 携带，携带空串 → 080005 由 RPC 侧判**）/ `SectionCode *string json:"section_code,optional"` + `CommunityIds []string json:"community_ids,optional"` + `HasScopeChange bool json:"has_scope_change,optional"`（true=全量替换 scope、空集 080005）+ `AttachmentIds []string json:"attachment_ids,optional"` + `HasAttachmentChange bool json:"has_attachment_change,optional"`（true=全量替换附件、空集=清空）+ `IsPinned *bool json:"is_pinned,optional"`（`*true` 置顶 / `*false` 取消置顶）+ `Status int32 json:"status,optional"`（值语义 0=无提交动作（编辑）/1=submit，与 proto 同号；其他值 080005）；**「未携带」与「置空」以指针 nil / 标志 false 区分，禁止值类型 + omitempty 启发式**
- [ ] `NoticeInfo` → `ContentPostInfo`：新增 `SectionCode`/`Status`/`AttachmentCount`（新键 additive）；`Text string json:"content"`（R2 wire 键）；`NoticeAttachmentInfo` → `ContentPostAttachmentInfo`：+`FileType`/`FileId`/`ReviewStatus`（新键 additive）
- [ ] 新增 `GetMarqueeNoticesReq { CommunityId int64 form:"community_id" }` + `GetMarqueeNoticesResp { Items []ContentPostMarqueeItemInfo json:"items" }`；`GetPublishPermissionReq`（空）/`GetPublishPermissionResp { CanPublish bool; PublishableRoles []int32 }`
- [ ] 错误码常量 `CodeNoticeNotFound` → `CodeContentPostNotFound`（080001）语义保留
- **SEE**: [[snake-camel-field-mismatch]]（TS/REST JSON 字段 snake_case，prose 引用；**wire 键 content↔Go Text 分轨是有意为之非不匹配，R2**）；[[api-required-field-marked-optional]]（RPC community_id 必填勿误标 optional；REST 兼容 optional 只落薄代理）

### Task 1.23: API handler/logic 代理 + routes 注册（含 R2 兼容回退）
- **修改**: `services/community-hub-service/api/internal/handler/routes.go`、`services/community-hub-service/api/internal/logic/notice/createnoticelogic.go` → `createcontentpostlogic.go`、`deletenoticelogic.go` → `deletecontentpostlogic.go`、`getnoticelogic.go` → `getcontentpostlogic.go`、`listnoticeslogic.go` → `listcontentpostslogic.go`、`updatenoticelogic.go` → `updatecontentpostlogic.go`、`api/internal/handler/notice/*`
- **创建**: `services/community-hub-service/api/internal/logic/notice/getmarqueenoticeslogic.go`、`getpublishpermissionlogic.go` + 对应 handler（`api/internal/handler/notice/getmarqueenoticeshandler.go`、`getpublishpermissionhandler.go`）
- [ ] API create/update/delete/list/get logic：代理到 RPC `ContentPostService` 对应方法；`community_ids`/`attachment_ids` 逐个 `strconv.ParseInt` 转 int64（REST `[]string` → RPC `[]int64 JS_STRING`，接口 v4 INFO 1 对齐）；**`entry_status`/`status` 同号数值映射（REST 0↔RPC 0=draft、REST 1↔RPC 1=submitted，数值即语义、无枚举偏移——禁止裸「透传」，评审 M2）**；`section_code` 直传；Get/Marquee 透传 `community_id`（form 绑定值）
- [ ] **Update 代理按 presence 转发（V5，评审 interface v4 MUST 1/S3）**：`Title/Text/SectionCode/IsPinned` **指针非 nil 才填 RPC optional 字段**（nil 不填，presence 语义不丢失）；`HasScopeChange`/`HasAttachmentChange` 标志 + 对应数组直传（true 时空数组也转发——「清空」语义必须到达 RPC 层）；`Status` 同号映射转发；**禁止「指针解引用 + omitempty 后裸透传」将 *false/*空串 坍缩成未携带**
- [ ] **详情 community_id 兼容回退（R2，评审 interface v3 MUST 2 修复）**：`GET /notices/:id` handler 在 `req.CommunityId==0` 时调 `ResolveReadableCommunityForCompat`（Task 1.14）——**按帖 scope 反查 + 逐小区 FilterAllowed 任一允许即放行** → 注入该小区调 RPC GetContentPost；080001/080005 透传（移动端 `getNoticeDetail(id)` 不传 community_id、多小区用户仍可用，不 080005）
- [ ] marquee / publish-permission handler + logic（新增代理，复用 `util.WithUserID` 注入出站 metadata 模式；响应 `ContentPostMarqueeItemInfo`）
- [ ] routes.go：REST 路径**保持 `/api/community/notices`**（本期不改，R2 wire 兼容；权限码 422-428 path 一致，design §ADR）；`GET /notices/marquee`、`GET /notices/publish-permission` **静态路径先于 `:id` 注册**（防被 `:id` 通配吞掉）；置于 PermMiddleware 中间件组内
- [ ] **RED**: api 层测试（community_ids=["1001","1002"] 解码转 int64 成功；**community_ids 含非数字 → 080005（评审 INFO 2，防静默忽略产生空范围误过校验）**；attachment_ids 透传 RPC 断言；entry_status/submitted 透传；**Update presence 转发断言（V5）：Title=nil → RPC 不填 title 字段；Title=ptr("新标题") → 填；IsPinned=ptr(false) → RPC is_pinned=false（取消置顶语义不坍缩）；HasAttachmentChange=true + AttachmentIds=[] → RPC 空数组 + 标志 true（清空语义到达）；HasAttachmentChange=false + AttachmentIds 非空 → 标志 false（不改）**；**GET /notices/:id?community_id=456 → req.CommunityId 绑定=456（form 生效）**；**GET /notices/:id 缺 community_id → 经 `ResolveReadableCommunityForCompat`：帖 scope 多小区任一可读 → 注入成功响应 / 全部不可读 → 080001 / 帖无 scope → 080005（R2，多小区用户详情可用，不 080005）**；marquee `?community_id=456` 绑定；缺失 → 080005）→ FAIL
- [ ] **GREEN**: 实现 → PASS
- [ ] `go build ./...` + 路由冒烟（curl 确认 marquee/publish-permission 不被 `:id` 抢占）
- **SEE**: [[snake-camel-field-mismatch]]（REST JSON 字段 snake_case，prose 引用；wire 键 content 分轨 R2）；[[permission-seed-api-path-must-match-routes]]（path 与实际 REST 路由一致）

---

## file-service

### Task 2.1: 错误码 70004/70005 登记 + 头注释对齐
- **修改**: `services/file-service/rpc/internal/errx/errcode.go`、`services/file-service/rpc/internal/logic/file/errcode.go`、`services/file-service/api/internal/logic/errcode.go`、`services/file-service/api/internal/logic/file/errcode.go`
- [ ] 新增 `ErrCodeUnsupportedFileType = 70004`（070004 不支持的文件类型）
- [ ] 新增 `ErrCodeFileSizeExceeded = 70005`（070005 文件大小超限）
- [ ] **70001-70003 不重编号**（`ErrCodeFileOperationFailed` 保持 70003，REUSE:notice-D11）
- **SEE**: [[error-code-collision-and-namespace-alignment]]（同整数双语义禁止）；[[error-code-literal-bypasses-qa-gate]]（用命名常量，禁裸数字）

### Task 2.2: 白名单/禁止集/10MB + GetUploadUrl L1 快速拒绝
- **创建**: `services/file-service/internal/guard/whitelist.go` + `whitelist_test.go`（通用校验器沉淀，供 L1/L2 共用）
- **修改**: `services/file-service/rpc/internal/logic/file/getuploadurllogic.go` + `getuploadurllogic_test.go`
- [ ] 常量：白名单 {png,jpg,jpeg,gif,pdf,doc,docx}、禁止集 {exe,bat,sh,cmd,com,msi,apk,js,vbs,ps1,py,pl,php}、`MaxSingleFileSize = 10MB`
- [ ] `ValidateFileName(fileName string) (ext string, err error)` — 扩展名大小写不敏感；无扩展名/点文件/非白名单 → ErrCodeUnsupportedFileType；禁止集/zip/rar → 同码（REQ-CAS-1：zip/rar 扩展名层全部拒绝）
- [ ] `ValidateFileSize(size int64) error` — >10MB → ErrCodeFileSizeExceeded；=10MB 放行（REQ-CAS-2）
- [ ] GetUploadUrl 生成预签名 URL 前调 `ValidateFileName` + `ValidateFileSize` 快速拒绝
- [ ] **RED**: 测试（白名单放行 + exe/sh/js 拒绝 + zip/rar 拒绝 + 大小写绕过无效 + 无扩展名/点文件拒绝 + 12MB 拒绝 + 恰 10MB 放行）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 2.3: magic-bytes 内容嗅探器（doc/docx/pdf/图片 + 容器子类型拒绝）
- **创建**: `services/file-service/internal/guard/magic.go` + `magic_test.go`
- [ ] `SniffType(buf []byte) (string, bool)` 返回规范扩展名 + 是否识别
- [ ] 规则：png `89 50 4E 47`、jpg `FF D8 FF`、gif `47 49 46 38`、pdf `%PDF`（读文件头字节）
- [ ] **doc = OLE2/CFB（D0 CF 11 E0 A1 B1 1A E1）且内含 `WordDocument` 流** → doc（仅 CFB 头不充分；msi/xls/ppt 无 WordDocument 流 → 不映射 → 拒绝）
- [ ] **docx = ZIP（PK 头）+ 含 `word/document.xml` 部件** → docx（xlsx/pptx 无该部件 → 不映射 → 拒绝；REVISION 消歧：docx 为唯一 zip 内容特判）
- [ ] 其他 OLE2（msi/xls/ppt）与其他 OOXML（xlsx/pptx）与通用 zip/rar 内容 → (not recognized)（调用方 070004，REQ-CAS-3）
- [ ] **RED**: table-driven 测试（真实 doc 放行 + msi 改 doc 拒绝 + 真实 docx 放行 + xlsx 改 docx 拒绝 + 通用 zip 拒绝 + 改名 png 传 exe（PE MZ）拒绝）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 2.4: ConfirmUpload L2 回读校验 + file_type/confirmed 落库
- **修改**: `services/file-service/rpc/internal/logic/file/confirmuploadlogic.go` + `confirmuploadlogic_test.go`、`services/file-service/rpc/internal/logic/file/helper.go`
- [ ] 回读 MinIO 实际对象（`RawMinio`/`MinioCli` GetObject）前若干字节 → `SniffType` → 映射规范扩展名（REQ-CAS-3 L2）
- [ ] 嗅探类型与声明扩展名一致才放行；不一致 → 070004（类型主码；大小超限为次码）
- [ ] File 记录写 `FileType`（嗅探映射）+ `Confirmed=true`
- [ ] `toProtoFile` 填充 `FileType`/`Confirmed`
- [ ] **RED**: 测试（正常确认成功 + 声明 png 实为 exe 拒绝 070004 + 声明与魔数一致放行 + FileInfo 返回 file_type/confirmed）→ FAIL
- [ ] **GREEN**: 实现 → PASS

### Task 2.5: File 模型扩展 + Migration 002（file_type/confirmed 列）
- **修改**: `services/file-service/model/file.go`、`services/file-service/model/filemodel.go`
- **创建**: `services/file-service/migration/002_file_guard.sql`
- [ ] `File` struct 增 `FileType string`（gorm column file_type）、`Confirmed bool`（gorm column confirmed，GORM 非 AutoMigrate 须显式迁移）
- [ ] `002_file_guard.sql`：**首行 `USE file_db;`（与 001 一致，迁移必须带 DB 上下文）**；`ALTER TABLE uploaded_file ADD COLUMN file_type VARCHAR(20) DEFAULT NULL, ADD COLUMN confirmed TINYINT NOT NULL DEFAULT 1`（存量行免嗅探即 confirmed，REUSE:notice D24 存量折中）
- [ ] `FileModel.Insert` 包含 file_type/confirmed 列
- [ ] model 层测试（Insert 后读回字段）
- **SEE**: [[migration-must-execute]]（提交后必须执行）；[[snake-camel-field-mismatch]]（prose 引用）

### Task 2.6: entity_type 覆盖机制（全局基线不弱化）
- **修改**: `services/file-service/internal/guard/whitelist.go`（覆盖注册）+ `whitelist_test.go`、`services/file-service/rpc/internal/logic/file/getuploadurllogic.go`
- [ ] `RegisterEntityTypeOverride(entityType string, cfg Override)` — 追加允许类型；**禁止集不可放行、10MB 硬上限不可放宽**（REQ-CAS-4 不变量）
- [ ] GetUploadUrl 按 `in.EntityType` 查覆盖：有 → 基线上叠加；无 → 全局基线（content_posts 附件本期走全局基线）
- [ ] **RED**: 测试（基线默认生效 + override 追加合法类型 + override 试图放宽 10MB/放行 exe 被拒 070005/070004 + 既有 avatar/verification/lostfound/contacts 上传不回归）→ FAIL
- [ ] **GREEN**: 实现 → PASS

---

## permission-service

### Task 3.1: AssertPublishScope 判据扩展（社区管理员角色感知展开，design gate——R1 grounded 重写）
- **修改**: `services/permission-service/rpc/internal/logic/permission/assertpublishscopelogic.go`、`services/permission-service/rpc/internal/logic/permission/scope.go`
- **创建**: `services/permission-service/rpc/internal/logic/permission/assertpublishscope_division_test.go`
- [ ] **前提核实（R1）**：`scope_type='community_div'` 在代码库不存在（rel.go:77-82 常量仅 global/""/community/building/unit/grid；apply_role_logic.go:15,66 将 community_admin 绑定 scope_type='community'、scope_id=communityId）→ **不新增 community_div scope_type**（评审选项 2）
- [ ] 新增 `resolvePublishScope(ctx, urm, mdClient, userId) (DataScopeState, []int64)` 变体（供 AssertPublishScope 使用）：
  - 基线：收集 `scope_type='community'` grant（同 `resolveUserScope`，ids 并集，含 GLOBAL 支配短路）
  - **社区管理员角色感知展开**：若用户持有 community_admin 角色（经 grants 的角色信息判断），对每个 community grant 的 scope_id（communityId）→ masterdata `GetResidentialArea(scope_id).community_div_id` → `GetResidentialAreasByDivision(division, status=1)` 展开为 approved 小区子树，**并入 ids**
  - **非 community_admin（owner/tenant/committee/property_admin/grid_worker）语义完全不变**（精确小区授权，不展开）
  - `targetCovered` 逻辑不变（祖先链 ∩ ids）；`resolveUserScope` 保持（GetDataScopes 读路径不动）
- [ ] `AssertPublishScopeLogic.AssertPublishScope` 改用 `resolvePublishScope`（scope.go 新增 helper + 复用既有 `grantActive`）
- [ ] **design gate 验证门禁（编码前必须通过）**：单测/集成验收——
  1. community_admin 持 `scope_type='community'` grant（scope_id=C_admin，C_admin ∈ division D1）发布 D1 内另一小区 C1 → allowed（**division 子树展开生效**）
  2. 发布 D1 外小区 C2 → 060007 denied
  3. 目标小区不存在（ResolveScopeAncestors found=false）→ denied（安全拒绝未知节点）
  4. **非 community_admin 不回归**：grid_worker/committee/property_admin/owner 持精确 community grant 发布授权小区 → 保持现状语义（不展开）；owner 发布到授权外小区 → denied
  5. **共享调用方回归（评审 SHOULD 2——AssertPublishScope 是共享 RPC，还被 lostfound 创建、contacts upsert 调用）**：community_admin 持 community grant（C_admin）时（a）lostfound 发布到 C1（同 division）→ allowed、（b）contacts upsert 到 C1 → allowed；owner 发布 C1 外 → denied；见 design §Design Gate「共享 blast radius 声明」
  6. 社区管理员多 community grant 映射不同 division → 并集展开（community-hub 侧仍由 Task 1.7 唯一 division 守卫兜底）
- [ ] **RED**: 以上测试先写并确认 FAIL（现状 `resolveUserScope` 不展开 division → 场景 1/5 denied）→ **GREEN**: 实现 → PASS
- **SEE**: [[is-system-no-permission-shortcut]]（global 无字段短路）；[[grpc-timeout-layers]]（AssertPublishScope 内嵌 ResolveScopeAncestors + GetResidentialArea 超时对齐）；[[tdd-red-evidence-requires-fail-excerpt]]（RED 证据含实际 FAIL 摘录）

### Task 3.2: 权限种子变更（REQ-CPP-3 REVISION）+ 读/写权限矩阵补齐
- **修改**: `services/permission-service/scripts/init_permissions.sql` + `docs/specs/rbac-design.md`
- **写路径**:
  - [ ] **property_admin(2) 保留 421（D6——不做回收，推翻 notice D26）**：不执行 DELETE
  - [ ] grid_worker(4) 授 421：`INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES (4, 421)`
  - [ ] 421 置 `min_verf_level=2`：`UPDATE sys_permission SET min_verf_level = 2 WHERE code = 'community:notice:create-api'`（行为变更 0→2，REVISION）
  - [ ] **撤销 owner(1)/tenant(5) 的 421：显式 `DELETE FROM rel_role_permission WHERE (role_id, permission_id) IN ((1,421),(5,421))`（INSERT IGNORE 无法撤销，SEE [[insert-ignore-swallows-errors]]）**；**保留 (1,435)/(1,436)/(5,435)/(5,436)**
  - [ ] 新增 `427 DELETE:/api/community/notices/:id`（code `community:notice:delete-api`，parent_id=420）绑定全部移动端角色
  - [ ] 新增 `428 PUT:/api/community/notices/:id`（code `community:notice:update-api`，parent_id=420）绑定全部移动端角色
- **读路径**:
  - [ ] **422 `GET:/api/community/notices` 扩展绑定全部移动端角色**（现仅 (9,1,5)；补 grid_worker 4/committee 6/merchant 7/community_admin 3/sys_admin 8）
  - [ ] 新增 `423 GET:/api/community/notices/marquee`（community:notice:read-marquee-api，parent_id=410）、`424 GET:/api/community/notices/publish-permission`（community:notice:publish-permission-api，parent_id=410）、`426 GET:/api/community/notices/:id`（community:notice:read-detail-api，parent_id=410）——均绑定全部移动端角色
  - [ ] parent_id 汇总：423/424/426 → 410；427/428 → 420（防孤儿节点，path 与实际 REST 路由一致，SEE [[permission-seed-api-path-must-match-routes]]）
- [ ] **幻影 435 措辞**：`community:lostfound:create-api` 无 sys_permission 行，保持现状不动；436 有真实行保持不动；本变更不动 owner/tenant 的 435/436
- [ ] **rbac-design.md §6.5 矩阵补登**：421 发布角色集变更（property_admin 保留/grid_worker 授/owner·tenant 撤销 + min_verf_level=2）+ 422 扩展 + 423/424/426/427/428 新增 + **property_admin 绑 421 不绑 427/428 的不对称注明（platforms='pc' 走后续 PC 接线，080002 作者校验兜底，评审 SHOULD #4）**
- [ ] 种子末尾验证查询阈值随回收/新增更新（owner/tenant 撤销 421 后下降、grid_worker/community_admin/committee 新增读码后上升；断言精确到具体码）
- [ ] 幂等：整段可重复执行（guard + INSERT IGNORE + 幂等 DELETE）；提交后按部署编排执行（SEE [[cross-service-seed-deployment-order]]）
- **SEE**: [[is-system-no-permission-shortcut]]（sys_admin 全权限经 rel_role_permission）；[[auto-grant-unverified-grant-confers-scope-level0]]（min_verf_level=2 收窄未认证发布）；[[permission-seed-api-path-must-match-routes]]

---

## moderation-service

### Task 4.1: Redis 消费者跳过 source_type="notice" + 移除 NoticeServiceClient 接线（D4/D21）
- **修改**: `services/moderation-service/rpc/internal/consumer/task_handler.go`、`services/moderation-service/rpc/internal/consumer/task_consumer.go`、`services/moderation-service/rpc/internal/svc/servicecontext.go`
- [ ] `TaskHandler` 移除 `communityNotice communityv1.NoticeServiceClient` 字段（D4——`NoticeService` 已改名 `ContentPostService`，客户端类型不存在；`UpdateNoticeModerationStatus` RPC 已移除 D21）
- [ ] `NewTaskHandler` 签名去掉 `communityNotice` 参数；`servicecontext.go` 移除 `CommunityHubNotice` 客户端创建与注入
- [ ] `callbackModerationStatus` switch：**`case "notice":` 跳过（不回调 NoticeService，日志记录「source_type=notice 残留任务已跳过」）**——content_posts 不再走 Redis，任何 notice 任务为残留（REVISION 精确跳过判定）；lostfound/user 路由不变（D3）
- [ ] **RED→GREEN**: 测试（source_type="notice" 任务被跳过不调用回调 + lostfound 任务仍走既有路径 + 无 notice client 时正常编译）→ 实现 → PASS
- **SEE**: [[rpc-callback-must-check-response-base]]（既有回调检查响应 Base 语义；本任务移除 notice 回调路径）

---

## master-data-service

> 只读复用，无代码变更。division→小区展开（`GetResidentialAreasByDivision` community_div_id>0 + status=1）、`GetResidentialArea`（community_div_id 派生，R1）、`ResolveScopeAncestors` 均已存在；验证见 Task 6.1/6.2（Owner 运维验证）。

---

## 前端

> **无接线任务（Q10 + R2 wire 兼容）**：web/mobile 本期不改，但**存在活跃通知消费方（R2 已核实）**——本设计通过 REST wire 键保持（notices/notice/content）+ 详情 community_id 兼容回退保证移动端现行通知 tab/浏览/详情运行期不破坏；前端各板块展示差异化后续单独做，不接线通用组件。

---

## Owner 运维验证（不走 harness-pipeline，编码后由 Owner 单独执行）

### Task 6.1: 三态库迁移验证
- [ ] community-hub-service `003_content_posts_generalize.sql`：新库/旧库升级/生产 三态执行（**003 为一次性 RENAME，勿重跑；重跑报错为预期——R4**），验证 `notices`→`content_posts`（DESC 确认 title/text/is_pinned/role/publisher 保留 + section_code/status/attachment_count/kafka_push_* 存在 + `published_at`/`community_id` DEFAULT NULL）+ `content_post_scope` 表 + 双索引存在 + `content_post_attachments` 含 post_id/review_status/file_id/file_type + `file_url` 加宽 VARCHAR(1024)
- [ ] **003 部分失败恢复演练（V5，评审 data-model v4 S2）**：若 003 中途某条 ALTER 失败（表已 RENAME、缺新列的半完成态）——**禁止直接重跑完整脚本**（RENAME 已发生报「表不存在/重名」）；先 `RENAME TABLE content_posts TO notices`（或按失败语句前状态手动对齐）回到可重入状态，再修复后重跑完整 003
- [ ] **迁移先于功能上线门禁（REVISION-9）**：未迁移 `community_id` 去 NOT NULL 时新发布 INSERT 被拒绝 → 迁移后成功（REQ-CPB-1 场景）
- [ ] **R1 前提核实**：真实 DB `SELECT DISTINCT scope_type FROM rel_user_role` 确认**无 `community_div` 行**（grounded 前提；如意外存在需先评估）
- [ ] file-service `002_file_guard.sql`：`uploaded_file` 增 `file_type`/`confirmed` 列
- [ ] permission-service `init_permissions.sql` 重跑幂等断言：
  - property_admin(2) **持** 421、grid_worker(4) **持** 421、owner(1)/tenant(5) **无** 421（(1,421)/(5,421) 删除生效，435/436 保留）、421 `min_verf_level=2`
  - 422 扩展后 grid_worker(4)/community_admin(3)/committee(6)/merchant(7)/sys_admin(8) 均持 422
  - 423/424/426/427/428 存在且绑定全部移动端角色；parent_id=410/410/410/420/420（无孤儿节点）
- **SEE**: [[migration-must-execute]]、[[cross-service-seed-deployment-order]]

### Task 6.2: 部署顺序 + Kafka 冒烟 + 端到端 + R2 移动端兼容回归
- [ ] `docker compose up -d`：Kafka（KRaft 单节点）启动健康、`content-review` topic 存在、数据卷持久（容器重启后 broker 健康）；community-hub 在 Kafka 就绪后再启
- [ ] 端到端冒烟（真实 DB）：网格员多小区发布成功（draft→submit 隐式通过 status=approved + published_at=NOW()）→ **Kafka content-review topic 收到符合 REQ-CPM-2 契约的消息（含 file_url + version）** → 列表/详情/跑马灯可见
- [ ] **community_admin division 展开冒烟（R1）**：社区管理员（持 `scope_type='community'` grant，scope_id=C_admin）发布 → 自动展开至其 division 全部 approved 小区（content_post_scope 多行）→ 各目标小区列表可见；多 division grant → 080005
- [ ] **Kafka 不可用冒烟（D20）**：停 Kafka 发布不阻塞（帖可见）、帖进入 pending-push、恢复后重推扫描补投、pending-push 计数可观测
- [ ] 不再 LPUSH Redis `moderation:task:queue`（content_posts）→ moderation-service 消费者对 source_type="notice" 残留任务跳过不报错；lostfound 审核流不变
- [ ] 审核完整性：附件 rejected 帖读路径不返回（谓词隐藏，status 不被 mutate）；无附件帖恒展示
- [ ] **R2 移动端兼容回归（评审 MUST）**：迁移后 `GET /api/community/notices?community_id=X` 返回键 **`notices`**（非 posts）；`GET /api/community/notices/:id`（**不传 community_id**）经兼容回退返回键 **`notice`** 且正文键 **`content`**——与 `api/community.ts` `getNoticeList`/`getNoticeDetail` 及 notice.vue/notice-browse.vue/notice-detail.vue 读取字段逐一对照（模拟移动端现行调用不 404/不 undefined）
- [ ] 权限矩阵冒烟：全部移动端角色调 `GET /notices/:id`、`GET /notices`、`GET /notices/marquee`、`GET /notices/publish-permission` 不 403；**property_admin（platforms='pc' 不绑 427/428）身份调用 `DELETE /notices/:id` / `PUT /notices/:id` → 403（fail-closed 回归护栏，评审 interface v3 SHOULD 2——防后续放开 427/428 到 property_admin 时静默变更）**；**撤销后 owner/tenant 直调创建被拒 080002（含 level-2 业主）**；property_admin 本小区发布成功（D6）、grid_worker 多小区、community_admin division 展开、committee 本小区
- [ ] 附件安全冒烟：白名单外 070004、单文件 >10MB 070005、zip/rar 扩展名拒绝、改名 png 传 exe 拦截、doc/docx 容器签名放行、单帖 ≤10 个/≤50MB（080005）
- [ ] 确认 design_consistency：`bash .harness/scripts/check-design-consistency.sh --all` 无本变更引入的 WARN

---

## Tasks Self-Review 记录

- **REVISION v5 修订（对照设计评审 data-model v4 / interface-proto v4 的 MUST/SHOULD/INFO，逐条落进 Task）**:
  - **MUST-1（interface v4 MUST 1）UpdateContentPostRequest presence 未定义** → Task 0.1 定义 title/text/section_code/is_pinned 用 proto3 `optional`（`*string`/`*bool`）+ community_ids/attachment_ids 加 `has_scope_change`(9)/`has_attachment_change`(7) bool 标志区分「未携带=不改」与「空数组=清空」；Task 1.11 分支判定改 presence/标志位（**取消置顶 `*false` / 清空全部附件 `attachment_count→0` / scope 空集 080005 / title·text 空串 080005** RED/GREEN 用例）；Task 1.22 REST 类型改 pointer 字段（与仓内 user-service/master-data 先例一致）；Task 1.23 代理按 pointer/标志转发（禁止解引用+omitempty 坍缩 presence）✅
  - **MUST-2（data-model v4 M1）is_pinned-only 路径持久化缺失** → Task 1.6 新增 `UpdateIsPinned`（独立列更新 `SET is_pinned=?`，不碰 title/text/section_code）+「更新后正文不变」断言；Task 1.11 (b) 分支 draft 与 submitted/approved 统一走 `UpdateIsPinned`，禁复用 `UpdateContent` 传空 title/text；(a) 分支 draft 编辑声明内容字段全量替换语义 + title/text 非空不变量（防「仅改附件」覆盖正文为空）✅
  - **SHOULD-1（data-model v4 S1）详情回退 080001/080005 冲突** → Task 1.14/1.23 统一「帖有 scope 但全部不可读 → 080001；帖无任何 scope（数据异常）→ 080005」，design.md §GetContentPost 措辞对齐 ✅
  - **SHOULD-2（data-model v4 S2）003 部分失败恢复缺失** → Task 6.1 补恢复指引（先 RENAME 回对齐再修重跑；半完成态禁直接重跑）+ design.md §数据模型 同步 ✅
  - **SHOULD-3（data-model v4 S3）is_pinned-only 判别依赖 presence** → 已由 MUST-1 的 optional + 标志位修复覆盖（Task 1.11/1.22/1.23）✅
  - **SHOULD-4（interface v4 SHOULD 1）预期破坏清单漏 content→text 改名** → Task 0.1 补 `ContentPost.content(4)→text(4)`（buf FIELD_SAME_NAME/FIELD_SAME_JSON_NAME 预期 fail，wire 由 REST `json:"content"` 承担）✅
  - **INFO**：Task 0.1 attachment_ids 统一 int64 JS_STRING（INFO 1）+ 显式列齐四个响应消息（INFO 2）+ Kafka 消费者按 post_id 去重注记（I3，design §Kafka 契约）；Task 1.13 role 映射收敛 helper.go 单源（I2）；Task 1.5 完整性子查询注明走 idx_notice（I4）✅
- **REVISION v4 修订（对照设计评审 data-model v3 / interface-proto v3 的 MUST/SHOULD，闭合 design↔tasks 漂移）**:
  - **MUST-1（data-model v3 M1 / interface v3 MUST 1）UpdateContentPost 授权模型自相矛盾** → Task 1.11 改「授权分流」：先 `FindOne` 再按请求形状分流——(a) 内容/附件/scope 编辑路径先作者校验（非发布者 080002）；(b) 仅 is_pinned 路径跳过作者校验改验 `PublishRolesFrom` 非空 + `AssertCommunitiesScope` 覆盖（非作者操作者置顶 approved 帖放行，scope 不覆盖 080006）；RED 增「非发布者操作者置顶成功 / 请求含 is_pinned+内容字段走 (a) 080002」用例 ✅
  - **MUST-2（interface v3 MUST 2）R2 详情回退仅支持单小区用户** → Task 1.14/1.23 弃 `ResolveSingleCommunityForCompat`（grant 唯一假设）改 `ResolveReadableCommunityForCompat`（scope 反查 + 逐小区 FilterAllowed 任一允许即放行）——多小区用户迁移后详情不 080005；与 design.md §GetContentPost 反查方案对齐 ✅
  - **SHOULD-1（interface v3 SHOULD 1）080004 标签漂移** → Task 0.1 + design.md §Proto 改「080004 寻失记录不存在（LostFoundService 仍用，CodeLostFoundMiss types.go:19）」，不固化「便民联络不存在/ContactService 仍用」旧标签 ✅
  - **SHOULD-2（data-model v3 S1）Kafka Producer.Push 触发点/顺序未落任务** → Task 1.10（entry=submitted）/1.11（submit 动作）各加「事务提交成功后调用 Task 1.18 Producer.Push（先提交后推送，提交失败不推送）」✅
  - **SHOULD-3（interface v3 SHOULD 2）property_admin 427/428 不对称回归断言** → Task 6.2 权限矩阵冒烟显式补「property_admin 调 DELETE/PUT /notices/:id → 403（fail-closed 护栏）」✅
  - INFO：Task 1.23 补 community_ids 非数字 → 080005 用例（INFO 2）；design.md §GetContentPost 注明统一 080001 含原 080006 拒绝路径（INFO 1）✅
- **REVISION v3 修订（对照设计评审 data-model v2 / interface-proto v2，逐条落进 Task）**:
  - **R1 [MUST] Design Gate grounded**（community_div scope_type 不存在）→ Task 1.7 `ResolveAdminDivision` 改经既有 `scope_type='community'` grant 派生 division（GetResidentialArea → community_div_id + 唯一 division 守卫）；Task 3.1 `resolvePublishScope` 改社区管理员角色感知展开（division 子树并入 ids），`resolveUserScope` 读路径不动；Task 6.1 加 `SELECT DISTINCT scope_type` 前提核实 ✅
  - **R2 [MUST] 破坏面评估修正 + REST wire 兼容**（web/mobile 活跃通知消费方：tabbar pages/notice/notice + notice-detail/notice-browse + api/community.ts getNoticeList/getNoticeDetail）→ Task 0.1 响应消息补 `base=1`（评审 SHOULD 1）；Task 0.3 CHANGELOG 如实登记消费方 + wire 兼容；Task 1.22 wire 键保持 `notices`/`notice`/`content` + 新增键 additive；Task 1.14/1.23 详情 community_id 兼容回退（初版 `ResolveSingleCommunityForCompat`，**REVISION v4 已改为 scope 反查 `ResolveReadableCommunityForCompat`——见上 MUST-2**）；Task 6.2 移动端兼容回归冒烟 ✅
  - **R3 [SHOULD] proto 头注释**：080002 注释扩展「无发布权限 / 非帖作者」（评审 S1）、保留 080004（评审 SHOULD 2）→ Task 0.1 ✅
  - **R4 [SHOULD] Migration 003 一次性 RENAME 勿重跑**（评审 S2）→ Task 1.1/6.1 ✅
  - **R5 [SHOULD] publisher 档案查询接线缺口**（servicecontext 无 UserClient，评审 SHOULD 3）→ 新增 Task 1.9（config + yaml + servicecontext 注册 `userv1.UserServiceClient`）✅
  - 评审 INFO：I2 Update.status「action」vs ContentPost.status「state」注释 → Task 0.1；INFO 1 marquee 固定 notice 注释 → Task 0.1/1.15；INFO 3 ContentPostRole 预期破坏清单 → Task 0.1；I8 Rescanner 复用 Producer.Push → Task 1.19 ✅
  - 任务重编号：插入 Task 1.9（user-service 接线）后 1.9→1.10 ... 1.22→1.23（全部依赖引用已随重写对齐）✅
- **REVISION v2 修订（上一轮，对照设计评审 M1-M3 / S1-S2 / I4-I8，全部保留）**:
  - M1 字段号冲突 → Task 0.1 ContentPost 保留既有 1-12、新增 13/14/15、role 保持 5 + 唯一性复核 checkbox ✅
  - M2 跨层枚举错位 → Task 0.1/1.22/1.23 entry_status·status 改三侧同号 int32、删 ContentPostEntryStatus 枚举、去「透传」改显式映射 ✅
  - M3 D3 未覆盖 Update 路径 → Task 1.11 落 Update/submit 停 LPUSH + Task 1.20 RED 断言扩双路径 ✅
  - S1/I3 content_post_scope 偏离登记 + 增长估算 → design.md §数据模型 ✅
  - S2 NoticeMarqueeItem→ContentPostMarqueeItem → Task 0.1/1.15/1.22 ✅
  - I4 ResolveAdminDivision 多 division→080005 → Task 1.7（R1 后语义保留：多 distinct division）✅
  - I8 Rescanner 复用 Producer.Push → Task 1.19 ✅
  - SHOULD#4 property_admin 421↔427/428 不对称注明 → design.md §权限种子 + Task 3.2 ✅
- **占位符扫描**: 无 `<任务描述>`/TBD/TODO；全部精确到文件路径 ✅
- **TDD 覆盖**: 所有含逻辑任务（1.3-1.23、2.2-2.6、3.1、4.1）含 RED→GREEN（前端无任务）；Migration（1.1/2.5）与 Proto（0.1-0.3）与基建（0.4）与配置（1.9/1.17）无逻辑代码不写 TDD，但含验证门禁/编译门禁 ✅
- **依赖顺序**: Proto(0.1-0.3) → Kafka 基建(0.4) → 迁移/模型(1.1-1.6) → scope 基础设施(1.7-1.8) → 身份派生接线(1.9) → 写逻辑(1.10-1.12) → 读逻辑(1.13-1.16) → Kafka 推送(1.17-1.20) → 接口(1.21-1.23) → file(2) → permission(3) → moderation(4) → 运维验证(6) ✅
- **独立可测**: 每 Task 文件数 ≤3、单代码层级、可独立编译测试 ✅
- **记忆引用检查**: 相关记忆全部注入 design.md；高风险 Task（迁移 1.1/2.5、种子 3.2、Design Gate 3.1、Kafka 1.18/1.19、附件安全 2.2-2.4）已标注 `SEE:` 引用；`[[unique-index-migration-dup-precheck]]`/`[[notfound-cache-sentinel-vs-transient-error]]`/`[[redis-cache-soft-delete]]` 主动排除并记理由；`snake-camel-field-mismatch` 为用户 auto-memory 以 prose 引用避免 slug 校验 MISS ✅
