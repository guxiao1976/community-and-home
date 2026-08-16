# Design Review — content-post-generalization（data-model 视角 v3）

**审查维度**: 数据模型 + 接口契约/Proto
**审查对象**: `.harness/changes/content-post-generalization/design.md` + `tasks.md`（REVISION v3）
**审查者**: Reviewer（data-model 视角，含 interface-proto）
**对照**: proposal.md + 5 个 spec + api-proto（community.proto / file.proto / permission.proto / masterdata.proto）+ 相关服务代码（rel.go / apply_role_logic.go / scope.go / assertpublishscopelogic.go / getuserroleslogic.go / checkpermissionlogic.go / helper.go / migration 001-003 / init_permissions.sql）

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 5

## 上一轮问题修复验证（v2 → v3）

| # | v2 问题 | 验证状态 |
|---|--------|:---:|
| M1（data-model，R1 核心）Design Gate 前提未 grounded：community_div scope_type 不存在 | ✅ **已修复且代码库验证成立**。apply_role_logic.go:15,49-67 将 community_admin 绑定 `scope_type='community'`、`scope_id=communityId`；rel.go:77-82 scope_type 常量无 community_div；GetUserRoles 返回 scope_type/scope_id（permission.proto UserRoleInfo）支撑 ResolveAdminDivision；FindActiveRolesByUserId 返回 role_code 支撑 resolvePublishScope 角色感知；两层机制（community-hub 唯一 division 展开 + permission 角色感知展开）自洽，共享 blast radius 已登记并带回归场景（Task 3.1 场景 5） |
| S1（interface）080002 语义重载 | ✅ 已修复（R3）。proto 头注释扩展为「无发布权限 / 非帖作者」；080004 保留（ContactService 仍用，community.proto 现存） |
| S2（data-model）Migration 003 一次性 RENAME 勿重跑 | ✅ 已修复（R4）。Task 1.1/6.1 显式注明「003 为一次性 RENAME，勿重跑；重跑报错为预期」 |
| interface v2 M1（R2）REST wire 破坏面 | ✅ 已修复。web/mobile 消费方（pages.json tabbar / api/community.ts getNoticeList/getNoticeDetail / notice.vue / notice-browse.vue / notice-detail.vue）已核实并登记；wire 键保持 notices/notice/content + 详情 community_id 兼容回退（ResolveSingleCommunityForCompat）+ REST 路径保持 /notices |
| interface v2 SHOULD（base=1 / 080004 / user 客户端） | 🟡 部分修复。base=1 已落 List/GetPublishPermission/GetMarquee 响应；**GetContentPostResponse / Create/Update/DeleteContentPostResponse 仍未在 Task 0.1 显式枚举（见 INFO I1）**；080004 保留 ✅；Task 1.9 补 UserClient 接线 ✅ |

## 发现

### 🔴 MUST FIX

#### M1（data-model/interface）UpdateContentPost 作者校验与「授权操作者置顶 is_pinned」在任务内直接冲突 → REQ-CPB-9(f)/REQ-CPR-3 的置顶能力照字面实现即失效

- **位置**: `tasks.md` Task 1.11（首步「作者校验：publisher_id == JWT user_id（非发布者编辑 → 080002）」+ 同任务「is_pinned 置位：submitted/approved 帖由持发布角色（PublishRolesFrom 非空）且数据范围覆盖帖小区的授权操作者」）；`design.md` §UpdateContentPost（080002=非发布者编辑 + is_pinned 由授权操作者置位）；spec REQ-CPB-9(f) 场景「授权操作者置顶 approved 帖」
- **问题**: Task 1.11 把作者校验列为逻辑首步的**通用门禁**（任何非发布者 → 080002），但同一任务又要求「submitted/approved 帖由**授权操作者**（非作者，持发布角色且 scope 覆盖）置位 is_pinned」。两个判定条件对「非作者授权操作者置顶 approved 帖」这一 spec 场景**互斥**：作者校验先跑 → 操作者非作者 → 080002 → 置顶永不生效 → 跑马灯置顶排序（REQ-CPR-3）对授权操作者失效。任务未给出判定顺序消解，照字面实现即业务不可用。
- **修复**: Task 1.11 合并为明确判定顺序：(1) `FindOne(id)` → 080001；(2) **请求形状分流**——若请求仅携带 `is_pinned`（title/text/section_code/community_ids/attachment_ids 均未变更）→ 走 is_pinned 专用授权（draft：作者即可；submitted/approved：`PublishRolesFrom` 非空 且 `AssertCommunitiesScope`/scope 覆盖帖小区，满足即放行，非作者不 080002）；(3) 否则为内容/附件/scope 编辑路径 → 先作者校验（非发布者 → 080002），再按 draft/非 draft 走 080005。design.md §UpdateContentPost 错误码列表相应注明「080002 仅内容编辑路径的非作者；is_pinned-only 由操作者授权」。

#### M2（data-model/interface）ResolveAdminDivision 未过滤角色 grant 状态 → 过期/已驳回的 community_admin grant 仍可驱动 division 展开（越权）

- **位置**: `tasks.md` Task 1.7 `ResolveAdminDivision`；`design.md` §Design Gate 第 1 层
- **问题**: `GetUserRoles` 返回**全部状态**角色（`FindAllByUserId`，getuserroleslogic.go:28「返回全部状态（前端展示认证进度）」），Task 1.7 过滤条件仅为 `role_code=community_admin 且 scope_type='community' 且 scope_id!=0`——**未过滤 URStatus**。而 421 `min_verf_level=2` 写路径门禁可被用户**另一** level-2 发布角色满足（committee(6) 保留 421 / grid_worker(4) 本变更新授 421），此时用户持有一个**已过期(4)/已驳回(3)**的 community_admin grant（scope_id 仍非 0）→ ResolveAdminDivision 仍将其 scope_id 经 GetResidentialArea 展开为 division → 用户获得超出其当前有效权限的整个 division 发布范围。grantSatisfiedLevel（helpers.go:21）只对 status==2 且 verified_at NOT NULL 计 level-2，但 ResolveAdminDivision 未复用它 → 422/唯一 division 守卫均不拦，**安全边界被绕过（权限提升）**。
- **修复**: Task 1.7 ResolveAdminDivision 对 community_admin grant 增加状态过滤：仅取 `URStatus==2`（level-2 等价，与写路径 421 门槛同语义：status==2 且 verified_at NOT NULL 且未过期）的 grant；与 permission-side `resolvePublishScope`（FindActiveRolesByUserId + grantActive）语义对齐。测试补：「过期 community_admin grant + 另一有效 level-2 角色 → 不展开（080005 或按有效 grant 直传）」。

### 🟡 SHOULD FIX

#### S1（data-model）Create/Update 的 Kafka Producer.Push 触发点与「事务提交后再推」顺序未显式指派任务

- **位置**: `tasks.md` Task 1.10 / 1.11 / 1.18
- **问题**: Task 1.10/1.11 只落 `kafka_push_status=1`（待推标记），Task 1.18 只定义 `Producer.Push`（含 UpdateKafkaPushStatus(2)）；「提交后调用 Producer.Push」的实际触发点（CreateContentPostLogic / UpdateContentPostLogic）与「先提交事务、成功后推送」的顺序约束未显式写入任一任务。照字面实现可能：漏接推送（仅靠 Rescanner 补投，延迟）或事务提交前推送（孤儿消息 / push 后事务回滚）。design §CreateContentPost 的「同事务置 1，随后生产者推」意图清晰，但任务未落点。
- **建议**: Task 1.10/1.11 各加一步「**事务提交成功后**调用 Task 1.18 `Producer.Push`（submitted 入口 / submit 动作）」，并注明「先提交后推送，提交失败不推送」。

### 🔵 INFO

#### I1（interface）Task 0.1 响应消息枚举不完整
- `GetContentPostResponse` / `CreateContentPostResponse` / `UpdateContentPostResponse` / `DeleteContentPostResponse` 未在 Task 0.1 显式列出（base=1 + 字段号）。既有 proto 全部 base=1，直接改名应继承，但为防实现遗漏建议补全（评审 SHOULD 1 仅覆盖了 List/GetPublishPermission/GetMarquee）。

#### I2（data-model）role 筛选的 enum→DB 字符串映射需与 PublishRoleToString 单一来源
- 现有 `roleToString`（helper.go:11）产出 community/committee/property/grid_officer，与 Task 1.8 `PublishRoleToString` 产出一致（已核对）；建议把 ListContentPosts role 筛选与写入侧映射收敛为单一来源，防两份映射漂移（Task 1.13「保留 notice 兼容语义」应注明复用 helper.go 映射）。

#### I3（interface）ResolveSingleCommunityForCompat 多 community grant → 080005 对多小区用户详情为行为回退
- 多小区用户（如 owner+C2 tenant）调详情（不传 community_id）→ 080005。fail-closed 已文档化、可接受；建议移动端未来显式传 community_id，或在详情错误提示引导。

#### I4（interface）GetContentPost 附件 file_url 重生时单附件 GetFileUrl 失败语义未定义
- 建议 best-effort：单附件失败跳过或回退 stored file_url，勿整单 080xxx；与「防御性 file_id=0/NULL 回退」一致。

#### I5（data-model）R1 闭环确认
- permission-side resolvePublishScope 多 division 并集展开（Task 3.1 场景 6）与 community-hub ResolveAdminDivision 唯一 division 守卫自洽，无授权覆盖不一致；两处 GetResidentialAreasByDivision 均 status=1（approved only），展开目标与授权集一致。

## 已核实无问题（数据模型 / 契约核对）

- ✅ 数据模型与 REQ-CPB-1/2/3/4/8/9/10 一致：content_posts RENAME+字段演化（section_code/status 0-4/attachment_count/Kafka 四件套）、published_at/community_id 去 NOT NULL（先于上线门禁）、content_post_scope 复合 PK+idx_scope_community、content_post_attachments post_id/review_status/file_id/file_type、status 枚举重定义与存量不可见自洽、attachment_count 冻结/重算（D19）。
- ✅ Migration 003 与 001/002 现状核对一致：moderation_status/moderation_time 由 002 添加且被 RENAME 保留；`content`→`text` 反引号包裹；idx_notice 随 CHANGE COLUMN 自动改指 post_id；一次性 RENAME 勿重跑已登记。
- ✅ Proto ContentPost 保留 1-12 + 新增 13/14/15、role 保持 5；ContentPostAttachment 1-4 + 5/6/7；新请求/响应消息全号唯一；int64 均 JS_STRING；file.proto FileInfo 1-10 + file_type(11)/confirmed(12) 兼容新增；GetFileUrl 返回 `FileInfo file=3` 支撑绑定校验与 URL 重生。
- ✅ entry_status / Update.status int32 三侧同号（0=draft/1=submitted），删 ContentPostEntryStatus 枚举，消除跨层枚举错位。
- ✅ REST wire 键保持（notices/notice/content）+ 详情 community_id 兼容回退 + REST 路径保持 /notices + 权限码 422-428 path 一致；go-zero GET form 标签绑定注意事项已落 Task 1.22。
- ✅ 权限种子：property_admin 保留 421、grid_worker 授 421、owner/tenant 显式 DELETE (1,421)/(5,421) 保留 435/436、421 min_verf_level 0→2、422 扩展、423/424/426/427/428 新增——与 init_permissions.sql 现状（(2,421)/(3,421)/(6,421)/(1,421)/(5,421)/(1,422)/(5,422)/(9,422)）核对一致；426 现无码（fail-closed 下现详情 403）需新增成立。
- ✅ IsReviewComplete 单一谓词（status==2 且 count(approved)==attachment_count）三读路径共用、不 mutate status；附件 rejected 隐藏；无附件帖恒完整；防御性 file_id=0/NULL 回退 stored file_url。
- ✅ moderation Task 4.1 移除 NoticeServiceClient + `case "notice"` 跳过，与 proto 移除 UpdateNoticeModerationStatus 一致（UpdateModerationStatusRequest/Response 被 LostFound 复用保留）。

## 问题跟踪表（v1 → v2 → v3）

| # | 状态 | 说明 |
|---|------|------|
| M1 (v1) | 已修复 | ContentPost 字段号唯一 + role 保持 5 |
| M2 (v1) | 已修复 | entry_status/Update.status 三侧同号，删枚举 |
| M1 (v2) | 已修复 | Design Gate 改经既有 community grant 派生（R1 grounded，代码库验证成立） |
| S1/S2 (v2) | 已修复 | 080002 注释扩展 / 003 一次性勿重跑 |
| interface v2 M1 | 已修复 | REST wire 兼容 + 破坏面修正（R2） |
| M1 (v3 新增) | 待修复 | is_pinned 授权操作者 vs 作者校验冲突（REQ-CPB-9(f)） |
| M2 (v3 新增) | 待修复 | ResolveAdminDivision 未过滤 grant 状态 → 过期 admin 越权展开 |
| S1 (v3 新增) | 待修复 | Kafka Producer.Push 触发点/顺序未落任务 |

---
VERDICT: REVISION
---
