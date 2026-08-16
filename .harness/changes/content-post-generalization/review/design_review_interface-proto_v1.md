# Design Review — content-post-generalization（interface-proto + data-model 视角）

**审查维度**: 接口契约 + Proto + 数据模型
**审查对象**: `.harness/changes/content-post-generalization/design.md` + `tasks.md`
**对照**: proposal.md + 5 个 spec + api-proto/api/community/v1/community.proto + api-proto/api/file/v1/file.proto + permission.proto + 相关服务代码（scope.go / updatenoticelogic.go / task_handler.go / migration 001-002 / errcode.go）

## 摘要
- 🔴 MUST FIX: 3 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 4

## 发现

### 🔴 MUST FIX
| # | 位置 | 问题 | 修复建议 |
|---|------|------|---------|
| 1 | `tasks.md` Task 0.1 + `design.md` §Proto（`ContentPost` 消息） | **ContentPost 消息字段号与保留字段冲突**。Task 0.1 定义 `content(4)→text`、新增 `section_code(3)`、`status(10)`、`attachment_count(11)`、`role(6) 改 ContentPostRole`；但既有 `Notice` 消息保留字段为 `title=3`、`created_at=10`、`updated_at=11`、`role=5`（`publisher=6`）。即 `section_code(3)↔title(3)`、`status(10)↔created_at(10)`、`attachment_count(11)↔updated_at(11)` 三处重复字段号，`role` 也错标为 6（实际 5）。按此定义消息无法通过 buf（同消息重复字段号），实现必挂。 | 新字段改用未占用号：`section_code(13)`、`status(14)`、`attachment_count(15)`，`role` 保持 5；或在 design.md 显式声明整消息重建编号并逐一列出全部字段（title/community_id/created_at/updated_at 也需分配），两法取一并在 design.md §Proto 表补全字段号清单。 |
| 2 | `tasks.md` Task 0.1 vs Task 1.21/1.22 | **REST 层状态值与 Proto 枚举值语义不一致 + 直传导致静默错误**。Proto `ContentPostEntryStatus` = `UNSPECIFIED=0/DRAFT=1/SUBMITTED=2`；REST types.go（Task 1.21）= `entry_status: 0=draft 默认/1=submitted`、`Update status 仅接受 1=submitted`；Task 1.22 声明 `entry_status/status ... 透传`。透传后 REST submitted(1) → proto 值 1 = DRAFT：Create 的「立即提交」被静默当成 draft；Update 的 submit 被 080005「仅接受 SUBMITTED」拒绝（或按 DRAFT 视为非提交）。 | 统一值语义并明确 API 层翻译：推荐 proto 枚举改为 `DRAFT=0/SUBMITTED=1` 与 REST 对齐（UNSPECIFIED 用 0 兼容默认 draft 语义需另行定义），或在 Task 1.22 指定 REST 1 → proto SUBMITTED(2) 的显式映射、删除「透传」措辞。两者必须二选一并写入 Task 0.1/1.21。 |
| 3 | `tasks.md` Task 1.10/1.19 + `design.md` §Kafka 推送（D3） | **D3「content_posts 只推 Kafka、不再 LPUSH Redis」只覆盖 Create 路径，Update/submit 路径仍会 LPUSH**。`updatenoticelogic.go`（Task 1.10 rename 为 updatecontentpostlogic.go）第 ~97 行仍有 `CreateAuditLog + RedisClient.LpushCtx("moderation:task:queue", ...)`。Task 1.19 仅移除 CreateContentPost 的 LPUSH，Task 1.10 未提及移除 Update 路径的 LPUSH。submit 经 Update 触发时会「既推 Kafka 又 LPUSH Redis」，违反 D3/REQ-CPM-3；且 Task 4.1 消费者跳过 source_type="notice" 后这些成为死任务/噪声。 | Task 1.10 须一并移除 update/submit 路径的 `CreateAuditLog + LPUSH`（updatenoticelogic.go 对应块）；Task 1.19 的 RED 断言扩展为「Create 与 Update-submit 后 `moderation:task:queue` 均无新元素」，并在 Task 1.10 落一句「不再 LPUSH（D3）」。 |

### 🟡 SHOULD FIX
| # | 位置 | 问题 | 建议 |
|---|------|------|------|
| 4 | `design.md` §权限种子 | **property_admin 绑 421（create）但不绑 427/428（update/delete，均「全部移动端角色」），移动端面出现「可建不可改/删」不对称**；D6 保留 property_admin 发布权，但 PC 本期不接线，property_admin 的编辑/撤回在移动端面无对应权限码闭环（仅 080002 作者校验兜底创建外的操作，而创建权限已授予）。 | 显式确认该不对称为有意（property_admin 走 PC 后续接线），并在 rbac-design.md §6.5 注明；或 427/428 补绑 property_admin（080002 作者校验不变）。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 5 | `NoticeMarqueeItem` 消息命名沿用 Notice 前缀，与 ContentPostService 命名不一致——新消息、风格问题，且 REST 路径保持 /notices，可接受；建议设计注释说明理由。 |
| 6 | `CreateContentPostRequest`/`UpdateContentPostRequest`/`ListContentPostsRequest` 均采用全新字段编号（未沿旧请求消息字段号增量）——因整服务改名已属破坏性，可接受；但建议 design.md 显式声明「新契约全新编号」而非「改名保留字段号」，避免实现歧义。 |
| 7 | migration `CHANGE COLUMN content text TEXT NOT NULL` 中 `text` 虽为 MySQL 非保留关键字可作列名，建议反引号包裹（`\`text\``）防解析歧义（Task 6.1 已含三态验证，风险低）。 |
| 8 | Task 1.18 Rescanner 重推时建议显式注明复用 Task 1.17 `Producer.Push`（内含 GetFileUrl 重生 file_url），避免重推携带过期预签名 URL；契约已声明消费者可经 file_id 再生，非阻塞，故仅提示。 |

## 架构一致性 / 契约核对（已核实，无问题）
- ✅ Proto 破坏性影响评估属实：web/pc 无 notice API 消费方（仅 ReviewFilter.vue 一个 radio 文案）、moderation-service 是 `UpdateNoticeModerationStatus` 唯一调用方（task_handler.go:170）且 Task 4.1 同步移除接线；`UpdateModerationStatusRequest/Response` 消息保留供 `LostFoundService`（community.proto:229）正确。
- ✅ file.proto `FileInfo` +`file_type(11)`/`confirmed(12)` 为未占用字段号，兼容新增。
- ✅ 070004/070005 为新增整数位：file-service errcode.go 现仅 70001-70003，Task 0.2/2.1 新增 70004/70005 无重编号冲突；头注释漂移修正方向与代码一致。
- ✅ Design Gate 证据核实：`scope.go:46` `g.ScopeType != scopeType → continue` 只收集单 scope_type，community_div grant 不会被 `scopeType='community'` 收集——division→community 授权落位（`resolvePublishScope` 收集双 scope_type）证据充分，Task 3.1 含共享调用方回归（lostfound/contacts）。
- ✅ `GetUserRoles` 输出（UserRoleInfo: role.code/status/verified_at/expires_at）支撑 GetPublishPermission level-2 判定与 ResolveAdminDivision。
- ✅ 数据模型：content_post_scope 复合 PK + idx_scope_community、content_post_attachments post_id/review_status/file_id/file_type、status 0-4 枚举、published_at/community_id 去 NOT NULL（迁移先于上线门禁）、软删 + withdrawn 语义均与 spec REQ-CPB-1/2/3/10 一致；int64 + `[jstype=JS_STRING]`、时间字段规范无违反。

## 问题跟踪表
| # | 状态 | 说明 |
|---|------|------|
| 1 | 待修复 | ContentPost 消息字段号冲突 |
| 2 | 待修复 | REST/Proto 状态值语义不一致（透传） |
| 3 | 待修复 | D3 停 Redis 推送未覆盖 Update/submit 路径 |
| 4 | 待确认 | property_admin 移动端 create/update/delete 不对称 |

---
VERDICT: REVISION
---

summary: 接口契约+Proto 视角设计方向正确、跨服务破坏面与 Design Gate 证据均核实无误，但存在 3 个 MUST FIX（ContentPost 消息字段号冲突、REST/Proto 状态值语义不一致且「透传」致静默错误、D3 停 Redis 推送未覆盖 Update/submit 路径），需修订后复审。
