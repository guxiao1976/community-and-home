# Design Review — content-post-generalization（interface-proto 视角 v4）

**审查维度**: 接口契约 + Proto 破坏性 + 依赖顺序
**审查对象**: `.harness/changes/content-post-generalization/design.md`（REVISION v4）+ `tasks.md`（REVISION v4）
**对照**: proposal.md + 5 个 spec + `api-proto/api/community/v1/community.proto` + `api-proto/api/file/v1/file.proto` + `api-proto/api/permission/v1/permission.proto` + `api-proto/api/masterdata/v1/masterdata.proto` + `api-proto/api/user/v1/user.proto` + `services/community-hub-service/api/internal/types/types.go` + `services/file-service/rpc/internal/errx/errcode.go` + `web/mobile/src/api/community.ts` + `services/community-hub-service/api/internal/types/types.go` + `services/community-hub-service/rpc/internal/logic/lostfound/*.go`

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2

## 上一轮问题修复验证（v1 → v2 → v3 → v4）

| 轮次 | 问题 | 验证状态 |
|------|------|---------|
| v1 M1 | ContentPost 字段号冲突 | ✅ 已修复。实核 community.proto：ContentPost 既有 1-12（content=4/role=5/publisher=6/created_at=10/updated_at=11）无重复，新增 13/14/15 未占用；ContentPostAttachment 1-4+新 5/6/7；FileInfo 1-10+新 11/12 均无冲突 |
| v1 M2 | REST/Proto 枚举错位 | ✅ 已修复。entry_status/Update.status int32 三侧同号，ContentPostEntryStatus 枚举删除 |
| v1 M3 | D3 停 Redis 未覆盖 Update 路径 | ✅ 已修复。Task 1.11 整体移除 + Task 1.20 RED 扩双路径 |
| v2 M1 | 破坏面误评估 + wire 改名破坏 | ✅ 已修复（R2 落地）。v4 实核 wire 键逐一成立：REST `NoticeInfo` json 键（content/role/publisher/publisher_id/is_pinned/created_at/updated_at/community_id/attachments[file_name,file_url,file_size]）与移动端 `api/community.ts` Notice 接口完全一致；`getNoticeList` 读 `res.notices`、`getNoticeDetail` 读 `res.notice`（仅传 id，无 community_id）；`NoticeAttachmentInfo.FileSize json:"file_size,string"` 保留 |
| v2 S1 | 响应 base=1 未声明 | ✅ 已修复。Task 0.1 显式 base=1，与既有 ListNoticesResponse/GetNoticeResponse 一致 |
| v2 S3 | publisher 档案缺 UserClient | ✅ 已修复。user.proto `GetUsersByIds` 存在，Task 1.9 接线 |
| v3 MUST 1 | UpdateContentPost 授权模型自相矛盾（作者校验阻塞操作者置顶） | ✅ 主体已修复（V4-1 授权分流）。**遗留契约缺口见 MUST 1（字段 presence 未定义，置顶/取消置顶/清空附件分支判定不可确定）** |
| v3 MUST 2 | R2 详情回退仅支持单小区用户 | ✅ 已修复（V4-R2 scope 反查）。Task 1.14/1.23 弃 `ResolveSingleCommunityForCompat` 改 `ResolveReadableCommunityForCompat`（`FindCommunityIdsByPostId` → 逐小区 `FilterAllowed` 任一允许即放行），多小区用户迁移后详情不 080005；与现网 getnoticelogic.go 反查 LIMITED 语义一致 |
| v3 SHOULD 1 | 080004 标签漂移 | ✅ 已修复（V4-3）。实核 `CodeLostFoundMiss` types.go:19 = 80004「寻失记录不存在」，唯一使用方 lostfound（getlostfoundlogic.go:34 / resolvelostfoundlogic.go:33 / updatemoderationstatuslogic.go:39），contact 逻辑无引用——「寻失记录不存在（LostFoundService 仍用）」标签正确 |
| v3 SHOULD 2 | property_admin 427/428 不对称回归断言 | ✅ 已修复（V4-5）。Task 6.2 权限矩阵冒烟显式补「property_admin 调 DELETE/PUT /notices/:id → 403（fail-closed 护栏）」 |
| v3 INFO 1/2 | 详情统一 080001 / community_ids 非数字 080005 | ✅ 已修复（V4-6）。design §GetContentPost 注明统一 080001（含原 080006 拒绝路径）；Task 1.23 RED 补 community_ids 非数字 → 080005 用例 |

## 发现

### 🔴 MUST FIX

| # | 位置 | 问题 | 修复建议 |
|---|------|------|---------|
| 1 | `design.md` §UpdateContentPost 授权分流（L192-202）+ `tasks.md` Task 0.1（UpdateContentPostRequest，L33）/ Task 1.11（L172-183）/ Task 1.22（UpdateContentPostReq，L277） | **UpdateContentPostRequest 的字段 presence 语义未定义，V4-1 授权分流的「仅 is_pinned 路径」判定与 draft 编辑的清空/全量替换操作不可确定实现。** proto3 非 `optional` 标量字段（title/text/section_code/is_pinned）不携带 presence——`title==""`、`is_pinned==false` 与「未携带」在二进制 gRPC 上不可区分；`repeated`（community_ids/attachment_ids）空数组与缺失在二进制 wire 上同样不可区分。后果：(a) 分支 (b)「请求只携带 is_pinned」只能靠 `is_pinned==true` 命中——**取消置顶（is_pinned=false，REQ-CPB-9(f) 隐含的双向置位）与空请求不可区分**，操作者/发布者无法确定性取消置顶；(b) draft 编辑以 attachment_ids 为「全量替换集」（Task 1.11 语义 DeleteByPostId+InsertBatch+重算 attachment_count）时，**清空全部附件（attachment_ids=[]）与「不改附件」不可区分**，attachment_count 不会归零，D19 不变量被绕开；(c) title/text 清空为合法空串与未变更不可区分（除非显式规定「编辑不得置空」，design 未声明）。而 V4-1 分支判定恰恰是本次 REVISION 的核心，当前设计把判定机制（presence 还是非空启发式）留给实现者自行决定，实现者将产出 design 未定义的两种行为。本仓已有 pointer-optional 先例（user-service `*string json:"nickname,optional"`、master-data `*int64 json:"...,optional"`、file-service `*int64/*string form:"...,optional"`）。 | Task 0.1 显式声明 UpdateContentPostRequest 各字段 presence 语义：title/text/section_code/is_pinned 用 proto3 `optional` 关键字（生成 `*string`/`*bool`，presence 可判）；community_ids/attachment_ids 为「全量替换集」且需区分「未携带=不改」与「空数组=清空」——因 repeated 无法用 optional，需显式机制（如 `google.protobuf.FieldMask`、包裹 message、或布尔 presence 标志 `has_attachment_change`），并同步 Task 1.22 REST 类型改 pointer 字段（`*string json:"title,optional"` 等，与仓内先例一致）、Task 1.23 代理按 pointer 判定转发、Task 1.11 分支判定/测试补「取消置顶」「清空全部附件」两用例（RED/GREEN）。 |

### 🟡 SHOULD FIX

| # | 位置 | 问题 | 建议 |
|---|------|------|---------|
| 1 | `tasks.md` Task 0.1 预期破坏清单（L40） | **expected-fail 清单遗漏 `content`(4)→`text`(4) 字段改名。** buf breaking-check 的 WIRE_JSON 类含 FIELD_SAME_NAME/FIELD_SAME_JSON_NAME——把字段 4 从 `content` 改名 `text` 会改变 JSON 名，`make ci` 会额外报 fail。当前清单只列了「ContentPostService 改名 / UpdateNoticeModerationStatus 移除 / role(5) FIELD_SAME_TYPE / enum NoticeRole→ContentPostRole」，实现者按清单人工核对时会遇到清单外 fail 项，误判为回归。 | Task 0.1 预期破坏清单补一行：`ContentPost.content(4) → text(4)` 字段改名（buf FIELD_SAME_NAME/FIELD_SAME_JSON_NAME 类，预期 fail；wire 兼容由 REST 层 `json:"content"` 承担，proto 层 JSON 名改 `text` 属有意）。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | **CreateContentPostRequest 载体内型不一致**：`community_ids`(5) 用 int64 JS_STRING，`attachment_ids`(6) 用 string，而两者同为 Snowflake ID、REST 均以 []string 承载、DB 均落 BIGINT。可行（各自在代理/逻辑层解析），但建议统一为 int64 JS_STRING（与 `ContentPostAttachment.file_id(6, int64 JS_STRING)`、`content_post_scope.community_id` 对齐），避免「同一 ID 两种载体内型」的长期契约漂移。 |
| 2 | **Task 0.1 未列出 GetContentPostResponse / UpdateContentPostResponse / DeleteContentPostResponse 的字段**（仅列请求消息）。建议补一行：`GetContentPostResponse { base=1; ContentPost content_post=2; }`（REST `GetContentPostResp` wire 键 `notice` 由 API 层映射，与移动端 `res.notice` 一致）；Delete/Update 响应 `{ base=1 }`（与既有 DeleteNoticeResponse/UpdateNoticeResponse 一致）。 |

## 已核实无问题（架构一致性 / 契约核对）

- ✅ **字段号全量唯一**：ContentPost 1-15（content=4→text、role=5→ContentPostRole、publisher=6 不变）、ContentPostAttachment 1-7（file_id=6 int64 JS_STRING、review_status=7 int32）、FileInfo 1-12（file_type=11 string、confirmed=12 bool）逐一与现有 proto 核对无冲突。
- ✅ **响应 base=1 全对齐**：ListContentPosts/GetPublishPermission/GetMarqueeNotices 响应显式 base=1；既有响应全部 base=1。
- ✅ **UpdateModerationStatusRequest/Response 保留正确**：community.proto L229 LostFoundService.UpdateLostFoundModerationStatus 仍用（消息不删）；仅移除 UpdateNoticeModerationStatus RPC。
- ✅ **moderation-service 为唯一外部 NoticeService 消费方**：Task 4.1 移除接线成立（v3 已核实）；同仓同步无运行期破坏。
- ✅ **file.proto 头注释修正 grounded**：实测 errcode.go 常量 70001=文件不存在/70002=文件访问被拒绝/70003=文件操作失败——当前头注释（070002 上传失败/070003 文件类型不支持/070004 文件大小超限/070005 bucket 不存在）确实漂移，Task 0.2 修正与常量一致；70004/70005 新整数位不重编号。
- ✅ **R2 wire 兼容键逐一成立**（v4 复核）：ListContentPostsResp.notices / GetContentPostResp.notice / 帖体 content / role(int32) / publisher / publisher_id / is_pinned / published_at / created_at / updated_at / community_id / attachments[id,file_name,file_url,file_size] 与移动端 Notice 接口 + notice.vue/browse/detail 读取字段一致；新增键 section_code/status/attachment_count/file_type/file_id/review_status additive。
- ✅ **R1 grounded + division 展开契约**：permission.proto UserRoleInfo{scope_type=2/scope_id=3/status=4/verified_at=5/expires_at=6} 支撑 level-2 判定与 ResolveAdminDivision（community_admin + scope_type='community' + scope_id!=0 + URStatus==2）；masterdata GetResidentialArea / GetResidentialAreasByDivision(community_div_id+status) / ResolveScopeAncestors 均存在；permission-service AssertPublishScope 为内部判据扩展（无 proto 变更），共享 blast radius 已在 Task 3.1 门禁场景 5 回归。
- ✅ **entry_status/Update.status 三侧同号**：REST 0/1 ↔ proto 0/1 ↔ DB（draft=0/submitted 隐式 approved=2），消除跨层枚举错位。
- ✅ **Update/Delete 错误码语义**：080002 跨端点复用（Create 功能层/Update·Delete 作者校验）头注释扩展；080003 超限/080005 请求形状/080006 目标级解析或越权，与 spec 一致。
- ✅ **Kafka 契约单源**：REQ-CPM-2 唯一权威，version=1 + file_url 可再生 + attachments 空数组非 null；非事务性 outbox（落库待推 + 定时重推）与 at-least-once 一致。
- ✅ **REST 路由顺序**：marquee/publish-permission 静态路径先于 :id（Task 1.23）；权限码 422-428 path 与 /notices 路由一致。
- ✅ **依赖顺序**：Proto(0.1-0.3) → Kafka 基建(0.4) → 模型/写/读/Kafka/接口(1.x) → file(2) → permission(3) → moderation(4) → 运维(6)；003 一次性 RENAME 勿重跑（R4）；task 单服务/单代码层级/独立可测。

## 问题跟踪表

| # | 状态 | 说明 |
|---|------|------|
| 1 | 待修复（本轮新增） | UpdateContentPostRequest 字段 presence 未定义 → V4-1 分支判定（仅 is_pinned / 取消置顶 / 清空全部附件 / 清空字段）不可确定实现 |
| 2 | 待修复（本轮新增） | Task 0.1 expected-fail 清单遗漏 content(4)→text(4) 改名（buf FIELD_SAME_NAME） |
| 3 | 已修复（v4 验证） | v3 MUST 1（授权分流）/ MUST 2（scope 反查）/ SHOULD 1（080004）/ SHOULD 2（427/428 回归）/ INFO 1-2 全部落地 |
| 4 | 已修复（v2/v1 验证） | 字段号/枚举错位/D3 双路径/破坏面 wire 兼容/base=1/UserClient 接线 |

---
VERDICT: REVISION
---
