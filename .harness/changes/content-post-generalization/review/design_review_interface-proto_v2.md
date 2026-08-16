# Design Review — content-post-generalization（接口契约 + Proto 视角）

**审查对象**: `.harness/changes/content-post-generalization/design.md`（REVISION v5）+ `tasks.md`
**审查者**: Design Reviewer（interface-proto 视角，模式一.5）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 3

## 磁盘核实结论（正向确认，均通过）

| # | 核实项 | 结论 |
|---|--------|------|
| 1 | **Proto 破坏面**（D4 直接改名 + D21 移除回调） | 属实。gRPC 消费方仅 moderation-service（`task_handler.go:38/44` + `servicecontext.go:147-152` 的 `communityv1.NoticeServiceClient`，Task 4.1 同步移除）+ community-hub API 层（Task 1.22/1.23 同步改名）；web/pc 无 notice 消费方（grep 为空，仅 `ReviewFilter.vue` 一个 radio 标签）；web/mobile 走 REST（wire 键保持，不破坏）。设计断言「不是无消费方、而是 wire 兼容」与代码现状一致 |
| 2 | **ContentPost 字段号** | 正确。现有字段 1-12（content=4/role=5/publisher=6/…/attachments=12），新增 13/14/15（section_code/status/attachment_count）与既有无冲突；role 保持 5、publisher 占 6（勿标 6）——评审 M1 冲突已修复 |
| 3 | **NoticeAttachment / FileInfo 字段号** | 正确。NoticeAttachment 1-4 + 新增 5/6/7（file_type/file_id/review_status）；FileInfo 现为 1-10（file.proto:71-90），新增 11/12（file_type/confirmed）为兼容新增 |
| 4 | **错误码 080003/080004 语义** | 安全。080003 当前无任何 RPC 产出（`CodeOverLimit` 仅 types.go:18 常量定义；lostfound 配额已迁 080007 `CodeSectionQuotaExceeded`，createlostfoundlogic.go:53-56），重定义「发布目标数量超限」无冲突；080004 仍由 lostfound 使用（`CodeLostFoundMiss`，getlostfoundlogic.go:29-34/resolvelostfoundlogic.go:33），保留不动正确 |
| 5 | **UpdateModerationStatusRequest 消息保留** | 正确。task_handler.go:154 与 LostFoundService `UpdateLostFoundModerationStatus` 共用该消息；移除 `UpdateNoticeModerationStatus` RPC 而消息保留不破坏 lostfound |
| 6 | **R2 wire 兼容断言** | 属实。`api/community.ts:123-135` `getNoticeList` 读 `res.notices` / `getNoticeDetail` 读 `res.notice`（不传 community_id）；`notice-browse.vue:44,113` / `notice-detail.vue:103-104,22-23` 读 `content`/`role`(int)/`publisher`。REST wire 键保持 `notices`/`notice`/`content` + 详情 community_id 兼容回退是破坏面最小且自洽的裁决 |
| 7 | **proto3 optional 仓内先例** | 属实。user-service types.go `*string json:"...,optional"`（L39-43）、`*int32 form:"status,optional"`（L12）——presence 指针模式有先例，V5 方案 grounded |
| 8 | **三态 status 枚举同号** | entry_status / Update.status 用 int32（0=draft/1=submitted），删除 `ContentPostEntryStatus` 枚举，消除「REST 1=submitted ↔ proto 1=DRAFT」枚举偏移根因；DB 落库映射在 Task 1.10/1.6 显式（submitted→2 隐式通过），见 SHOULD-2 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 章节 | 问题 | 修复建议 |
|---|------|------|---------|
| S1 | design.md §UpdateContentPost (V5) / §CreateContentPost | **draft 编辑 `section_code` 无白名单/非空校验（契约缺口）**：创建路径 `section_code ∈ 注册板块集 → 否则 080005`，但编辑路径只声明 title/text 非空不变量；`SectionCode` presence 携带空串会被接受并清空板块列（DB `NOT NULL DEFAULT 'notice'` 接受空串），产生「approved 但 section_code='' 无板块」帖——任何 section_code 筛选/notice 板块跑马灯均不可见，发布者以为成功实则丢失可见性 | draft 编辑时若 `SectionCode` presence 且值为空串或不在注册板块集 → 080005（与创建同规则，REQ-CPB-5 一致性）；Task 1.11 RED 补「section_code 携带空串 → 080005」用例 |
| S2 | design.md §接口设计 前言（三侧同号声明） | **「REST / Proto / DB 三侧同号」表述不精确**：DB 落库对 submitted 入口写 `status=2`（隐式通过 D16，见 §CreateContentPost 落库与 Task 1.10「status=入口（draft=0 / submitted=2）」），非 1。该强声明若被实现者字面误读为「DB 也写 1」→ 帖停留在 submitted(1)、读路径谓词（status=2）永不成立 → 发布不可见。行为正确但表述有误导风险 | 将声明改为「REST/Proto 同号（0=draft / 1=submit action）；DB 落库按 D16 隐式通过写 status=2，submitted(1) 为未来消费者阶段保留非本期稳定终态」；Task 0.1 entry_status/status 注释同步措辞 |
| S3 | tasks.md 依赖顺序 / design.md §Design Gate | **跨服务耦合依赖未显式落位**：Task 1.10「community_admin 展开成功」RED 需 permission `AssertPublishScope` 对 division 子树授权，该能力在 Task 3.1 `resolvePublishScope` 角色感知展开才实现；Task 3.1 排在 community-hub(1.x) 之后。若 Task 1.10 用真实 permission 联调会失败（target 不在 admin 精确 grant → 080006）。同理 Task 1.10/1.11 前向依赖 Task 1.18 `Producer.Push` | Task 1.10/1.11 显式声明「community_admin 授权经 mock AssertPublishScope 允许结果验证（真实授权由 Task 3.1 + Task 6.2 端到端兜底）」+「Producer 经 mock 接口（writer 可 mock），真实投递 Task 6.2」；或将 Task 3.1 提前于 Task 1.10 排程 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | **Task 1.10/1.11 RED 测试用例数超模式 1.5 刚性上限（~14/~15 > 10）**：虽为单文件 table-driven、拆分反损逻辑原子性，interface-proto 视角不作为阻塞，提请 Owner 裁决是否接受或拆子任务 |
| I2 | **CreateContentPost community_admin 携带 community_ids 的语义未显式**：设计只描述 division 展开路径，未说明 admin 请求带 community_ids 时是忽略还是合并。建议显式声明（建议：admin 时 community_ids 忽略，目标集以 division 展开为准；非 admin 时 community_ids 必填） |
| I3 | **REST 兼容回退多可读小区取首个**：`ResolveReadableCommunityForCompat` 对多 scope 小区返回首个可读小区（响应 community_id=该小区）。Task 1.14 已含多小区用例，建议注释补「多可读小区时取首个、行为由测试固定」 |

## 校验摘要（模式 1.5 焦点维度）

- **接口契约**（gRPC/Proto 自洽）：✅ ContentPostService 5 个 RPC 改名 + GetPublishPermission/GetMarqueeNotices 新增 + UpdateNoticeModerationStatus 移除，消息/字段号/响应 base 布局自洽；UpdateContentPost presence 语义（optional + has_scope_change/has_attachment_change 标志）完整闭合「未携带=不改 vs 空数组=清空」
- **Proto 破坏性**（向后兼容 + 评估）：✅ D4 一次性破坏 + D21 移除回调，破坏面经磁盘核实（仅同仓 moderation + 本仓 API 层），REST wire 兼容使移动端运行期不破坏；file.proto 全兼容新增；预期破坏清单含 content(4)→text(4) 的 buf WIRE_JSON 项
- **依赖顺序**：✅ Proto→Kafka 基建→迁移/模型→scope→身份→写→读→推送→接口→种子→moderation→运维验证；仅见 S3 跨服务耦合落位缺口（SHOULD）

---
VERDICT: APPROVED
---
