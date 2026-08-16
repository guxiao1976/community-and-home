# Design Review — content-post-generalization（data-model + interface-proto 视角）

**审查维度**: 数据模型 + 接口契约/Proto（data-model 与 interface-proto 双维度）
**审查对象**: `.harness/changes/content-post-generalization/design.md` + `tasks.md`（REVISION v5，对照 data-model v4 / interface-proto v4 修订后的最新版）
**审查者**: Reviewer（data-model 视角，含 interface-proto）
**对照**: proposal.md + 5 个 spec + `api-proto/api/community/v1/community.proto` + `api-proto/api/file/v1/file.proto` + `api-proto/api/permission/v1/permission.proto` + `api-proto/api/masterdata/v1/masterdata.proto` + `services/community-hub-service/migration/001_initial.sql`/`002_add_moderation_status.sql` + `services/community-hub-service/model/notice.go`/`notice_attachment.go` + `services/community-hub-service/rpc/internal/logic/notice/*.go` + `services/community-hub-service/rpc/internal/logic/scope/scope.go` + `services/permission-service/model/rel.go` + `services/permission-service/scripts/init_permissions.sql` + `web/mobile/src/api/community.ts` + `web/pc/src/`

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3

## 上一轮问题修复验证（v4 → v5）

| # | v4 问题 | 验证状态 |
|---|--------|:---:|
| data-model v4 M1（is_pinned-only 路径复用 `UpdateContent` 清空已发布帖正文） | ✅ **已修复（V5-2）**。Task 1.6 新增 `UpdateIsPinned(ctx,id,isPinned)`（独立列更新 `SET is_pinned=? WHERE id=? AND deleted_at IS NULL`，不碰 title/text/section_code）+「更新后正文不变」断言；Task 1.11 (b) 分支 draft 与 submitted/approved 统一走 `UpdateIsPinned`，显式禁止复用 `UpdateContent` 传空 title/text；Task 1.6 `UpdateContent` 已改为三列（title/text/section_code）不再写 is_pinned |
| data-model v4 S1 / interface v4 S1（兼容回退 080001/080005 冲突） | ✅ **已修复（V5-3）**。Task 1.14/1.23 + design.md §GetContentPost 统一「帖有 scope 但全部不可读 → 080001（不泄露）；帖无任何 scope（数据异常）→ 080005」 |
| data-model v4 S2（003 一次性迁移无部分失败恢复） | ✅ **已修复（V5-4）**。design.md §数据模型 + Task 6.1 补恢复指引：先 `RENAME TABLE content_posts TO notices`（或手动对齐表结构）回到可重入状态再重跑；中间态禁直接重跑完整脚本 |
| data-model v4 S3 / interface v4 MUST 1（UpdateContentPostRequest 字段 presence 未定义） | ✅ **已修复（V5-1）**。Task 0.1 权威定义 title/text/section_code/is_pinned 用 proto3 `optional`（Go `*string`/`*bool`）；community_ids/attachment_ids 为「全量替换集」+ `has_scope_change`(9)/`has_attachment_change`(7) bool 标志区分「未携带=不改」与「空数组=清空」；分支判定以 presence/标志位为准；Task 1.22 REST 改 pointer 字段、Task 1.23 代理按 pointer/标志转发；RED 补「取消置顶 *false / 清空全部附件 attachment_count→0 / scope 空集 080005 / title·text 空串 080005」用例 |
| interface v4 SHOULD 1（预期破坏清单漏 content→text 改名） | ✅ **已修复（V5-5）**。Task 0.1 补 `ContentPost.content(4)→text(4)`（buf FIELD_SAME_NAME/FIELD_SAME_JSON_NAME 预期 fail；wire 兼容由 REST `json:"content"` 承担） |
| interface v4 INFO 1/2（attachment_ids 载体内型 / 响应消息字段） | ✅ **已修复（V5-6）**。Task 0.1 attachment_ids 统一 int64 JS_STRING 对齐 file_id；显式列齐四个响应消息字段；Kafka 消费者按 post_id 去重注记；role 映射收敛 helper.go 单源；完整性子查询注明走 idx_notice |

## 本轮实测核实的 grounded 结论（无问题）

- ✅ **Migration 003 与 001/002 现状逐列吻合**：notices（001: community_id NOT NULL / content TEXT NOT NULL / role VARCHAR(20) NOT NULL / publisher VARCHAR(100) NOT NULL / published_at DATETIME NOT NULL / is_pinned / deleted_at；002: moderation_status/moderation_time）→ 003 的 RENAME + `CHANGE COLUMN content \`text\` TEXT NOT NULL` + section_code/status/attachment_count + `MODIFY published_at DATETIME DEFAULT NULL` + `MODIFY community_id BIGINT DEFAULT NULL` + kafka_push_* 四列 + content_post_scope（双 NOT NULL 复合 PK + idx_scope_community）+ content_post_attachments（CHANGE notice_id→post_id + review_status/file_id/file_type + file_url 加宽 1024）。`idx_notice` 随 CHANGE COLUMN 自动改指 post_id（MySQL 语义，设计注释正确）。`content`→反引号 `text` 防保留字解析歧义。
- ✅ **Proto 字段号全量核对**：`community.proto` 实测 Notice：id=1/community_id=2/title=3/content=4/role=5/publisher=6/publisher_id=7/is_pinned=8/published_at=9/created_at=10/updated_at=11/attachments=12 → ContentPost 保留 1-12、新增 section_code=13/status=14/attachment_count=15、role 保持 5、publisher 占 6（无误标 6 冲突）；ContentPostAttachment 1-4+file_type=5/file_id=6/review_status=7；`file.proto` FileInfo 1-10 + file_type=11/confirmed=12（兼容新增）；int64 全部 `[jstype=JS_STRING]`。`UpdateModerationStatusRequest/Response` 保留（community.proto L229 LostFoundService.UpdateLostFoundModerationStatus 仍用），仅移除 UpdateNoticeModerationStatus RPC（D21）正确。
- ✅ **错误码语义实测**：80004 CodeLostFoundMiss「寻失记录不存在」唯一使用方 lostfound（getlostfoundlogic/resolvelostfoundlogic/updatemoderationstatuslogic），contact 无引用——「保留 080004 + 修正旧标签」正确；80006 CodePublishScopeDenied 已存在（scope.go:24），080002 跨端点语义（Create 功能层/Update·Delete 作者校验）与现网 deletelogic 复用一致；80003 CodeOverLimit 现无逻辑使用方，「超限（复用）」可承接「发布目标数量超限」语义。
- ✅ **权限种子现状**：init_permissions.sql:202 现 `UPDATE sys_permission SET min_verf_level = 0 WHERE code IN ('community:notice:create-api', 'community:lostfound:create-api')`——421 置 2（仅 notice）与 lostfound 保持 0 的区分正确；(1,421)/(5,421) 现绑定（L252-253）需显式 DELETE 保留 435/436——Task 3.2 措辞与 `[[insert-ignore-swallows-errors]]` 一致。
- ✅ **R1 grounded 全链路**：permission rel.go:77-82 scope_type 常量 = {global/""/community/building/unit/grid} 无 community_div（实测）；apply_role_logic 将 community_admin 绑 scope_type='community'；UserRoleInfo{scope_type=2/scope_id=3/status=4/verified_at=5/expires_at=6} 支撑 ResolveAdminDivision 的 URStatus==2 过滤；masterdata GetResidentialArea/GetResidentialAreasByDivision(community_div_id+status)/ResolveScopeAncestors 均存在；共享 blast radius（lostfound/contacts 复用 AssertPublishScope）已登记 + Task 3.1 门禁场景 5 回归。
- ✅ **R2 wire 兼容逐一成立**：web/mobile `api/community.ts` 实测 getNoticeList 读 `res.notices`、getNoticeDetail(id) 读 `res.notice`（不传 community_id）；notice-detail.vue/notice-browse.vue 为注册页、notice.vue 为 tabbar 主页面——REST 键保持 notices/notice/content + 详情 scope 反查兼容回退 + RPC community_id 严格必填（回退只落 REST 薄代理）成立。web/pc 无 notice 消费方（仅 moderation ReviewFilter.vue 标签引用）。
- ✅ **data-model 一致性**：IsReviewComplete 单一谓词（status=2 且 count(approved)==attachment_count）三读路径共用、不 mutate status、无附件帖恒完整、attachment_count 冻结/重算（D19）；status 枚举重定义 + 存量无 scope 不可见自洽；community_id 去 NOT NULL 迁移先于上线门禁；published_at 用 sql.NullTime 防零值（`[[restore-compensation-zero-time]]`）；content_post_scope 纯关联表偏离 §3.1 已显式登记理由（不可变行，撤回由主表软删表达）。
- ✅ **entry_status/Update.status int32 三侧同号** + 删 ContentPostEntryStatus 枚举 + 「action vs state」注释标注（评审 I2），消除跨层枚举错位。
- ✅ **Kafka 契约单源**（REQ-CPM-2）+ 可再生 file_url + version=1 + 空附件数组非 null + at-least-once（待推标记+定时重推+pending-push 可观测）+ 停 Redis 双路径（Task 1.20 断言扩双路径）。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

#### S1（interface-proto）空 UpdateContentPostRequest（无任何字段、status=0）落不到 (a)/(b) 任一授权分支——行为未定义

- **位置**: `design.md` §UpdateContentPost（授权分流 (a)/(b)）+ `tasks.md` Task 1.11
- **问题**: (a) 分支条件 = `Title!=nil || Text!=nil || SectionCode!=nil || HasScopeChange==true || HasAttachmentChange==true || status==1`；(b) 分支条件 = `IsPinned!=nil && Title/Text/SectionCode 均 nil && HasScopeChange/HasAttachmentChange 均 false && status==0`。一个请求若全部字段缺失（所有 pointer nil、两标志 false、status=0）**同时不满足 (a) 与 (b)**——例如客户端误发空 PUT，或未来前端渲染异常发出仅含 id 的更新。当前设计未声明该形状的行为（静默 no-op？080005？）。
- **建议**: 在 Task 1.11 显式补一行：请求不含任何变更字段（(a)/(b) 均不命中）→ 视为空操作返回成功（或 080005），并加一个 RED 用例锁定，避免实现者自行选择。

#### S2（data-model/interface）两阶段状态机的 draft 无任何读路径——草稿仅能靠 Create 返回的 id 访问，丢失即不可恢复

- **位置**: `design.md` §ListContentPosts（status=2 谓词）/ §GetContentPost（FindOneReviewComplete）+ §UpdateContentPost/§DeleteContentPost（按 id）
- **问题**: draft（status=0）在全部读 RPC 中不可见：ListContentPosts 谓词 `status=2` 排除草稿；GetContentPost 经 `FindOneReviewComplete`（status=2）→ 080001。草稿的唯一访问路径是 Update/Delete 按 id——id 丢失（客户端未保留 Create 响应）后草稿永久不可达（数据留存但业务不可用）。draft 状态机本质是「只写不读」，与本变更核心卖点「draft 草稿可编辑（可删）」的 round-trip 承诺存在缺口。本期无前端（D10）不阻塞，但未来前端接线「我的草稿」必然需要草稿列表/详情 RPC。
- **建议**: design §本期范围/接口设计补一句显式声明：draft 仅支持按 id 编辑/删除，无「我的草稿」枚举/读取路径，属本期后端契约边界（D10 前端未接线），未来前端接线时另立「my drafts」读 RPC；BACKLOG 登记。若产品要求草稿可恢复，则需在 ListContentPosts 增加「仅作者 + status=draft」查询模式。

#### S3（interface-proto）content-review 契约未含 `title`——未来审核消费者只能审正文，标题（同为 UGC）被排除；契约已冻结（version=1）则需 version bump 才能补

- **位置**: `design.md` §Kafka 推送契约（REQ-CPM-2）+ `tasks.md` Task 1.18
- **问题**: 契约顶层字段 = version/post_id/section_code/text/publisher_id/attachments[]，无 title。现网 Redis 流（createnoticelogic.go）的审核消息是 `title + "\n" + content` 全量文本；新契约只带 `text`（正文）。标题同样是用户生成内容、属内容审核对象，缺 title 意味着未来 D18 消费者无法对标题做关键字/大模型审核。契约被设计为「本期冻结，消费者不改推送方；变更走 version bump」——标题缺失后补将强制 bump 版本，属可避免的演进成本。
- **建议**: 发布前将 `title` 并入契约顶层（version 仍为 1，本期契约未上线消费者，无破坏）；或至少在 design §Kafka 契约注明「title 不在审核范围」是有意决策（若产品确认标题免审）。二选一，不留默认。

#### S4（design↔tasks 漂移）CreateContentPostRequest 无 is_pinned 字段，但 Task 1.22 REST `CreateContentPostReq` 含 `IsPinned bool json:"is_pinned,optional"`——REST 死字段被静默丢弃

- **位置**: `tasks.md` Task 1.22（CreateContentPostReq IsPinned）vs Task 0.1（CreateContentPostRequest 无 is_pinned）vs design.md §CreateContentPost（落库 `is_pinned=0` 恒定）
- **问题**: Proto/RPC/落库三处均不支持创建时置顶（恒 0），但 REST 请求体声明了 `is_pinned` 键——代理层不转发（RPC 无此字段）即静默丢弃。客户端若相信该字段会被处理会产生「创建即置顶」假象；也违反「分支判定以 presence/标志位为准，禁止 value 非空启发式」的同一精神（bool 无 presence，`is_pinned=false` 与缺失不可分，且此处无任何语义）。
- **建议**: Task 1.22 从 `CreateContentPostReq` 移除 `IsPinned`（与 Proto/落库对齐）；若有意保留扩展位，显式注明「创建不支持置顶，is_pinned 恒 0，字段不转发」。

### 🔵 INFO

#### I1（interface-proto）`enum NoticeRole` → `ContentPostRole` 的枚举值命名是否改名未在预期破坏清单显式登记
Task 0.1 写「值不变：UNSPECIFIED/COMMUNITY/COMMITTEE/PROPERTY/GRID_OFFICER」——若含义是枚举**值名**也去掉 `NOTICE_ROLE_` 前缀（如 `CONTENT_POST_ROLE_COMMUNITY`），则 proto-JSON 序列化值名改变（枚举值经名称序列化），属 JSON wire 破坏，建议像 content→text 一样在 expected-fail 清单（Task 0.1）补一行；若保留旧值名则无须。REST 层 `Role int32`（数值 1-4 不变）不受影响。

#### I2（data-model）List 路径附件 URL：`file_url` 占位空串不能流入列表响应
新行 `content_post_attachments.file_url=''`，详情由 GetContentPost 重生（GetFileUrl）。ListContentPosts 若把 ContentPost.attachments（proto 字段 12）带进响应而**不**重生 URL，移动端列表会收到空 file_url 附件。建议 Task 1.13 注明列表响应不填充 attachments（保持空数组）或对列表附件同样重生，杜绝占位空串泄漏到 wire。

#### I3（tasks）Task 1.11 测试场景数超「单任务 1~10 个」刚性线（约 14 个 table-driven 行）
属同一 handler（UpdateContentPostLogic）的 table-driven 行，未跨服务/未混合数据模型变更，仍属一个逻辑单元；接受但请知悉该任务是全变更最重任务，若执行期拆分，注意 (a)/(b) 授权分流用例须与 S1 判空用例一同保留。

## 已核实无问题（契约/数据模型交叉核对）

- ✅ ContentPost 字段号 1-15 全唯一（含 role=5 保持、新增 13/14/15）；ContentPostAttachment 1-7；FileInfo 1-12。
- ✅ UpdateContentPostRequest 字段号 1-10 全唯一（id/title/text/section_code/community_ids/attachment_ids/has_attachment_change/is_pinned/has_scope_change/status）；与 design §UpdateContentPost 完全一致。
- ✅ `optional` presence 方案有仓内先例（user-service `*string`/master-data `*int64`/file-service `*int64`）；repeated 用 bool 标志区分「未携带=不改」与「空数组=清空」的机制自洽（REST 显式 false == 未携带 == 不改，无行为歧义；is_pinned 用 *bool 双向置位可确定）。
- ✅ IsReviewComplete 谓词与附件行 review_status 默认 approved 使本期 submit 后谓词立即成立；attachment_count 冻结/重算同事务；并发重复 submit 双推由 at-least-once + 消费者 post_id 幂等容忍。
- ✅ 停 Redis 队列双路径（Create 主路径 Task 1.10 移除 + Update/submit 路径 Task 1.11 整体移除 + Task 1.20 双路径 RED 断言）；moderation-service 精确跳过 source_type="notice"。
- ✅ R1 division 派生两层 grounded（community-hub ResolveAdminDivision + permission resolvePublishScope 角色感知展开）与共享 blast radius 回归（Task 3.1 场景 5）闭环。
- ✅ 任务粒度：单任务单服务、无「数据模型变更+业务逻辑+前端」混排、Proto/Migration 独立成任务（0.x / 1.1 / 2.5）；依赖顺序 Proto→Kafka→迁移/模型→scope 基础设施→身份派生→写→读→Kafka 推送→接口→file→permission→moderation→运维。

## 问题跟踪表（本轮）

| # | 级别 | 状态 | 说明 |
|---|------|------|------|
| S1 | SHOULD | 待确认 | 空 UpdateContentPostRequest 落不到 (a)/(b) 分支，行为未定义 |
| S2 | SHOULD | 待确认 | draft 无读路径，仅 id 可达；需显式声明边界 + BACKLOG 登记 |
| S3 | SHOULD | 待确认 | Kafka 契约缺 title（UGC 审核对象）；冻结后补需 version bump |
| S4 | SHOULD | 待确认 | CreateContentPostReq REST is_pinned 死字段被静默丢弃 |
| I1-I3 | INFO | 待确认 | 枚举值名 expected-fail / List 附件空 URL / Task 1.11 粒度 |

---
VERDICT: APPROVED
---
