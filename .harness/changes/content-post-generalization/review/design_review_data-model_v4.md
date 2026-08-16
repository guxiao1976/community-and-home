# Design Review — content-post-generalization（data-model 视角 v4）

**审查维度**: 数据模型 + 接口契约/Proto
**审查对象**: `.harness/changes/content-post-generalization/design.md` + `tasks.md`（REVISION v4）
**审查者**: Reviewer（data-model 视角，含 interface-proto）
**对照**: proposal.md + 5 个 spec + api-proto（community.proto / file.proto / permission.proto / masterdata.proto）+ 相关服务代码（migration 001/002 / notice.go / notice_attachment.go / helpers.go / servicecontext.go / init_permissions.sql / web/mobile/src/api/community.ts）

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 4

## 上一轮问题修复验证（v3 → v4）

| # | v3 问题 | 验证状态 |
|---|--------|:---:|
| interface v3 MUST 1 / data-model v3 M1（UpdateContentPost 作者校验 vs is_pinned 操作者置顶互斥） | ✅ **已修复（V4-1）**。Task 1.11 + design.md §UpdateContentPost 均改为请求形状分流：(a) 内容/附件/scope 编辑路径先作者校验 080002；(b) 仅 is_pinned 路径跳过作者校验改验 `PublishRolesFrom` 非空 + `AssertCommunitiesScope`，非作者操作者不 080002、scope 不覆盖 080006；RED 增「非发布者操作者置顶 approved 帖成功 / 请求含 is_pinned+内容字段走 (a) 080002」用例。授权矛盾消除 |
| interface v3 MUST 2（R2 详情回退仅支持单小区用户） | ✅ **已修复（V4-R2）**。design.md §GetContentPost + Task 1.14/1.23 弃 grant 唯一假设，改 `ResolveReadableCommunityForCompat`（`FindCommunityIdsByPostId` scope 反查 + 逐小区 `FilterAllowed` 任一允许即放行），与现网 getnoticelogic.go 反查 LIMITED 语义一致；多小区用户（grid_worker/多房产业主）迁移后详情不 080005；RPC 层 community_id 严格必填不变，回退只落 REST 薄代理 |
| data-model v3 M2（ResolveAdminDivision 未过滤 grant 状态） | ✅ **已修复**。Task 1.7 增加 `URStatus==2（level-2 等价：status==2 且 verified_at NOT NULL 且未过期）` 过滤；实测 permission helpers.go:26 `grantSatisfiedLevel` 语义一致（status==2 && VerifiedAt.Valid && 未过期）；permission.proto `UserRoleInfo`（status=4/verified_at=5/expires_at=6）支撑该过滤；测试补「过期(4)/驳回(3) community_admin grant 不计入」 |
| data-model v3 S1（Kafka Producer.Push 触发点/顺序未落任务） | ✅ **已修复（V4-4）**。Task 1.10/1.11 各加「事务提交成功后调用 Task 1.18 `Producer.Push`（先提交后推送，提交失败不推送）」 |
| interface v3 SHOULD 1（080004 标签漂移） | ✅ **已修复（V4-3）**。Task 0.1 + design.md §Proto 改「080004 寻失记录不存在（LostFoundService 仍用，CodeLostFoundMiss types.go:19）」；080004 行保留不动 |
| interface v3 SHOULD 2（property_admin 427/428 403 回归断言） | ✅ **已修复（V4-5）**。Task 6.2 权限矩阵冒烟显式补「property_admin（platforms='pc' 不绑 427/428）调 DELETE/PUT /notices/:id → 403（fail-closed 护栏）」 |
| interface v3 INFO 1/2 | ✅ **已修复（V4-6）**。design.md §GetContentPost 注明 scope 外统一 080001（含原 080006 拒绝路径）；Task 1.23 补 community_ids 非数字 → 080005 用例 |

## 发现

### 🔴 MUST FIX

#### M1（data-model）is_pinned-only 路径（b 分支）的持久化机制缺失 → 复用 `UpdateContent` 将清空已发布帖 title/text/section_code（数据丢失）

- **位置**: `tasks.md` Task 1.6（模型方法清单）+ Task 1.11（(b) 分支）；`design.md` §UpdateContentPost（(b) 分支）
- **问题**: Task 1.11 (b)「仅 is_pinned 路径」对 **submitted/approved 帖**（REQ-CPB-9(f) 置顶能力、REQ-CPR-3 跑马灯排序依赖）只定义了**授权判定**（PublishRolesFrom 非空 且 AssertCommunitiesScope），**未定义持久化方式**。而 Task 1.6 模型层唯一的 is_pinned 写入方法是 `UpdateContent(ctx, id, title, text, sectionCode string, isPinned int32)` —— 它同时写 title/text/section_code 四列。is_pinned-only 请求（proto3 标量无 presence）携带的 title/text 为空串，若实现者按直觉复用它 → **把已 approved 帖的 title/text/section_code 全部清空**（发布后内容丢失，数据损坏）。同样问题波及 (b) 分支的 draft 帖（"draft 帖 → 发布者即可"）与 (a) 分支的「draft 编辑仅改附件/scope、不动正文」场景（UpdateContent 需 title/text 现值的 read-modify-write 或全量替换，任务未声明语义）。
- **修复**:
  1. Task 1.6 新增 `UpdateIsPinned(ctx, id int64, isPinned int32) error`（`UPDATE content_posts SET is_pinned=? WHERE id=? AND deleted_at IS NULL`，**仅写 is_pinned 列，不碰 title/text/section_code**），加 RED/GREEN 用例（更新 is_pinned 后 title/text 不变断言）。
  2. Task 1.11 (b) 分支持久化统一走 `UpdateIsPinned`（draft 与 submitted/approved 均如此），design.md §UpdateContentPost (b) 注明「is_pinned 持久化走独立列更新方法，禁止复用 UpdateContent 传空 title/text（会清空已发布帖正文）」。
  3. Task 1.11 (a) 分支 draft 编辑显式声明**内容字段全量替换语义**（title/text/section_code 请求须携带完整现值；REST 层用 pointer/omitempty 区分「未携带」与「置空」），或注明 read-modify-write 保留未携带字段——防止「仅改附件」时正文被空串覆盖。

### 🟡 SHOULD FIX

#### S1（interface-proto）design.md §GetContentPost 兼容回退的 080001/080005 错误码与 Task 1.14 语义冲突（同一场景两种描述）

- **位置**: `design.md` §GetContentPost（L224「无小区用户（scope 外/不存在/未过审）→ 080001，无任何可允许小区 → 080005」）vs `tasks.md` Task 1.14（「全部不读取 → 080001；帖无任何 scope 小区（数据异常）→ 080005」）
- **问题**: 对「帖存在且过审、有 scope 小区、但用户对任一小区均无读权限」这一场景，design 写 080005，Task 写 080001。两者对同一场景给了互斥错误码。080005（参数不合法）与「scope 外 → 080001（不泄露）」的 RPC 契约（§GetContentPost 流程）不自洽，且与 Task 的解析矛盾。
- **建议**: 统一为 Task 1.14 语义——帖有 scope 小区但全部不可读 → **080001**（与 RPC 层 scope 外统一 080001 一致，不泄露）；仅「帖无任何 scope 小区（数据异常）」→ 080005。design.md L224 措辞对齐修正。

#### S2（data-model）Migration 003 一次性 RENAME 无部分失败恢复/回滚说明

- **位置**: `tasks.md` Task 1.1 + `design.md` §数据模型（003 迁移）
- **问题**: 003 是**单脚本内串联** RENAME TABLE + 多次 ALTER + CREATE TABLE + MODIFY。若执行中某条 ALTER 失败（如并发 DDL 锁、语法报错），表已 RENAME 但后续语句未执行，处于半完成态；「一次性勿重跑」已注明重跑报错，但**未给出部分失败后的恢复指引**（Migration 专项检查要求配套回滚/恢复）。
- **建议**: Task 6.1 补一行恢复指引：若 003 中途失败，先 `RENAME TABLE content_posts TO notices`（或按失败语句前状态手动对齐）再修复重跑；注明「已 RENAME 但缺新列的中间态禁止直接重跑完整脚本」。

#### S3（interface-proto）is_pinned-only 分支检测依赖「未携带字段」可判别，但 proto3/REST 层未定义 presence 语义

- **位置**: `tasks.md` Task 1.22（UpdateContentPostReq 字段为值类型）+ Task 1.11（(b) 分支请求形状判别）
- **问题**: (b) 分支依赖「请求只携带 is_pinned，其余内容字段均未变更」的判别。proto3 标量字段无 presence，REST 层若用值类型 + omitempty，无法区分「title 未传」与「title 传空串」；且无法区分「is_pinned=false 未传」与「is_pinned=false 显式传」。判别规则若只靠「Title=="" && len(CommunityIds)==0...」，恶意/异常客户端可构造歧义请求。
- **建议**: Task 1.22 将 UpdateContentPostReq 的 Title/Text/SectionCode 改为 `*string`（pointer，`json:"...,omitempty"`）、IsPinned 改为 `*bool`（或 REST 层以 `json.RawMessage`/存在性 map 判别），使「未携带」与「置空」可区分；并在 Task 1.11 注明判别以「字段存在性」为准（pointer != nil），非「值非空」。

### 🔵 INFO

#### I1（interface）Task 0.1 响应消息枚举仍不完整
`CreateContentPostResponse` / `GetContentPostResponse` / `UpdateContentPostResponse` / `DeleteContentPostResponse` 未显式列出（base=1 + 字段号）。直接改名应继承既有 `CreateNoticeResponse{base=1,id=2}` 等（已实测 community.proto:99-103），但建议 Task 0.1 补全显式字段号，防实现遗漏。

#### I2（data-model）ListContentPosts role 筛选的 enum→DB 字符串映射未收敛单一来源
Task 1.13「role 筛选保留 notice 兼容语义」未注明复用哪份映射；写入侧 `PublishRoleToString`（Task 1.8，RBAC code→DB role 列）与读侧 `ContentPostRole` enum→DB role 列映射是两套产生同一字符串集合的映射，建议统一为单一来源（如 helper.go 收敛），防后续板块扩展时两份映射漂移（承 v3 INFO I2，本轮仍未闭合）。

#### I3（interface）并发重复 submit 可能产生重复 Kafka 消息
两并发 `UpdateContentPost(status=submitted)` 同时读到 draft → 都置 approved → 都 Push content-review。at-least-once 语义容忍重复投递，但建议 design §Kafka 契约注明「后续消费者须按 post_id 幂等去重」，并把并发 submit 的 RED 场景补进 Task 1.11（防双推）。

#### I4（data-model）`FindListByCommunity` 附件完整性子查询为关联标量子查询
`(SELECT count(*) FROM content_post_attachments a WHERE a.post_id=p.id AND a.review_status=1) = p.attachment_count` 每行一次索引回查（idx_notice 兜底）。量级（百级帖 × ≤100 小区）可接受，无需预聚合；建议注释注明走 idx_notice(post_id)，勿退化为全表扫描。

## 已核实无问题（数据模型 / 契约核对）

- ✅ 数据模型与 REQ-CPB-1/2/3/4/8/9/10 一致：notices 现状（001：community_id/published_at NOT NULL、content/role/publisher/is_pinned 均存在）与 003 迁移（RENAME + content→text + section_code/status 0-4/attachment_count + 去 NOT NULL + content_post_scope 复合 PK + idx_scope_community + content_post_attachments post_id/review_status/file_id/file_type + file_url 加宽）逐列吻合；status 重定义与存量不可见（无 scope 行）自洽。
- ✅ Migration 003 与 002 现状核对：moderation_status/moderation_time 由 002 添加且经 RENAME 物理保留；idx_notice 随 CHANGE COLUMN 自动改指 post_id；一次性 RENAME 勿重跑（R4）已登记。
- ✅ Proto ContentPost 保留 1-12（实测 community.proto Notice：id=1/community_id=2/title=3/content=4/role=5/publisher=6/publisher_id=7/is_pinned=8/published_at=9/created_at=10/updated_at=11/attachments=12）+ 新增 13/14/15、role 保持 5；ContentPostAttachment 1-4 + 5/6/7；int64 均 JS_STRING；file.proto FileInfo 1-10 + file_type(11)/confirmed(12) 兼容新增。
- ✅ entry_status / Update.status int32 三侧同号（0=draft/1=submitted），删 ContentPostEntryStatus 枚举，消除跨层枚举错位（REST 1=submitted ↔ proto 1=DRAFT 根因）。
- ✅ R1 grounded 全链路实测：UserRoleInfo（scope_type/scope_id/status/verified_at/expires_at）支撑 ResolveAdminDivision 的 URStatus==2 过滤；`ResidentialArea.community_div_id=4`（GetResidentialArea）支撑 community_admin 经既有 community grant 派生唯一 division；`GetResidentialAreasByDivisionReq.community_div_id(3)+status(4)` 支撑 approved 子树展开；md_residential_area 模型含 community_div_id 列。shared blast radius（lostfound/contacts 复用 AssertPublishScope）已登记 + Task 3.1 场景 5 回归。
- ✅ servicecontext.go 实测无 UserClient（仅 Moderation/Permission/MasterData/SysConfig/Redis）→ Task 1.9 补 UserRpc 接线成立；web/mobile `getNoticeDetail(id)`（api/community.ts:133，不传 community_id）确认 R2 兼容回退必要性。
- ✅ REST wire 键保持（notices/notice/content）+ REST 路径保持 /notices + 详情兼容回退只落 REST 薄代理 + RPC community_id 严格必填；`CommunityIds []string`（encoding/json `,string` 不支持 slice）已注明。
- ✅ 权限种子：property_admin 保留 421、grid_worker 授 421、owner/tenant 显式 DELETE (1,421)/(5,421) 保留 435/436、421 min_verf_level 0→2（init_permissions.sql:201-202 现为 0）、422 扩展、423/424/426/427/428 新增——与 init_permissions.sql 现状（(2,421)/(3,421)/(6,421)/(1,421)/(5,421)/(9,422)/(1,422)/(5,422)）核对一致。
- ✅ IsReviewComplete 单一谓词（status==2 且 count(approved)==attachment_count）三读路径共用、不 mutate status、无附件帖恒完整、attachment_count 冻结/重算（D19）；attachment_count 与现写链路（Insert 显式 + UpdateContent + UpdateStatusAndPublish 同事务 kafka_push_status=1）一致。
- ✅ Kafka 契约单源（REQ-CPM-2，REQ-CPB-7 引用）；at-least-once（待推标记 + 定时重推 + pending-push 可观测）；停 Redis 双路径（Task 1.20 断言扩双路径）；UpdateNoticeModerationStatus RPC 移除、UpdateModerationStatusRequest/Response 保留（LostFound 复用，community.proto:229）。

## 问题跟踪表（v3 → v4）

| # | 状态 | 说明 |
|---|------|------|
| interface v3 MUST 1 | 已修复 | UpdateContentPost 授权分流（(a)/(b)） |
| interface v3 MUST 2 | 已修复 | R2 详情回退改 scope 反查（ResolveReadableCommunityForCompat） |
| data-model v3 M2 | 已修复 | ResolveAdminDivision URStatus==2 过滤 |
| data-model v3 S1 | 已修复 | Producer.Push 触发点/顺序落任务 |
| interface v3 SHOULD 1/2 | 已修复 | 080004 标签 / property_admin 427/428 403 断言 |
| M1（本轮新增） | 待修复 | is_pinned-only 路径持久化机制缺失 → UpdateContent 复用清空已发布帖内容 |
| S1（本轮新增） | 待修复 | design vs task 的兼容回退 080001/080005 冲突 |
| S2（本轮新增） | 待修复 | 003 一次性迁移部分失败恢复指引缺失 |
| S3（本轮新增） | 待修复 | is_pinned-only 判别依赖 presence，REST 层未定义 pointer 语义 |
| I1-I4（本轮新增） | 待确认 | 响应消息枚举 / role 映射单一来源 / 并发双推去重 / 关联子查询索引 |

---
VERDICT: REVISION
---
