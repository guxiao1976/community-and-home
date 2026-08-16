# Plan Review — mobile-homepage-content-revamp（结构合理性视角）

**审查维度**: 职责边界、一致性
**审查版本**: P1.3（fallback:r1:rc1，按磁盘最新内容独立重新审查，未沿用 r0 结论）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 3

## 审查对象核验（磁盘最新）
- 5 个 capability spec 与 proposal 7 个工作项一一对应：notice-time-window（通知窗口/列表/跑马灯）、notice-detail-preview（详情+附件）、function-entries（4 入口）、contact-list-page（联络页+补表）、homepage-layout（邻里互助占位/寻失保持/广告集中/垂直全序）。
- change.yaml services（web/mobile、community-hub-service、api-proto）与各 spec 服务职责边界一致；file-service 为复用（RPC GetFileUrl 无所有权限制、REST GET /api/files/:id 强制所有权），不列入 services 合理。
- 首页垂直区块链在 proposal 验收与 homepage-layout REQ-HL-4 间自洽：通知(跑马灯+卡片) → 4 功能入口 → 邻里互助占位 → 寻失互助 → 底部广告位；联络拨号网格职责拆分清晰（function-entries REQ-FE-2 移除首页内嵌网格 → contact-list-page REQ-CLP-1 承接，无重复）。
- 跨 capability 引用全部正确：FE-2→CLP-1、NTW-4→NTW-5（卡片视觉契约锚点）、NTW-4→NTW-1、HL-4→NTW/HL；v1 遗留的错误 REQ-HL-2 交叉引用已消除。

## 结构断言代码验证（关键职责边界的准确性）
- **附件预览「无新增后端变更」边界成立**：`services/community-hub-service/rpc/internal/logic/notice/getcontentpostlogic.go:100` 已调用 `FileClient.GetFileUrl` 重生 file_url（有测试断言 read_write_logic_test.go:152）；file-service RPC `GetFileUrl`（rpc/internal/logic/file/getfileurllogic.go）无所有权校验，而 REST 端点（api/internal/logic/file/getfileurllogic.go:41）强制 `rpcResp.File.UserId != userId → permission denied`。REQ-NDP-2/3「前端不直连 REST、消费详情响应重生 file_url」职责边界与真实 wire 一致。
- **跑马灯「由首页通知列表派生、不改 GetMarqueeNotices」成立**：web/mobile notice.vue marqueeText 为 computed（L280-282），由 `notices`（getNoticeList→ListContentPosts）派生，移动端不消费 GetMarqueeNotices RPC；proto GetMarqueeNoticesRequest 保持不动，REVISION #6 一致。
- **REQ-NTW-6 索引前提成立**：`migration/001_initial.sql:24` `idx_published(community_id, published_at DESC, deleted_at)` 的 community_id 在 003 中被标记为弃用 NULL 列（003 L23），列表查询按 `content_post_scope.community_id` JOIN 过滤（model/content_post.go:158-197），旧索引确实无法服务窗口过滤/排序——补索引必要。
- **REQ-NTW-4 基线准确**：notice-browse.vue L110 `getNoticeList(cid,1,50)` + L113-114 客户端 3 个月过滤；首页 notice.vue L336 `getNoticeList(cid)` 无时间约束——「原行为→新行为」diff 与代码一致。
- **migration 004 根因声明成立**：`community_contacts` DDL 已存在于 `001_initial.sql:39-48`（id/community_id/category/name/phone/sort_order/created_at/updated_at + idx_community），与 CommunityContactModel 对齐；REQ-CLP-2 已声明缺表根因（迁移应用缺失/库漂移）并明确 001 为 schema 单一权威、004 为幂等补救，v1 S1 已解决。
- **api-proto since_days 为 additive**：ListContentPostsRequest 现用字段 1-5，新增字段 6 不破坏兼容；缺省不过滤保留 PC 行为（REQ-NTW-2 场景明确）。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| S1 | notice-time-window/spec.md REQ-NTW-5（含 Scenario「时间格式与回退一致」） | REQ-NTW-5 规定卡片时间「formatted from published_at（fallback created_at when published_at is NULL）」并显式测试 NULL→created_at 回退；但同 capability 的 REQ-NTW-1/REQ-NTW-4 明确规定 NULL `published_at` 行不进窗口（「NULL is not within the window, regardless of status」）——首页与列表页两处卡片使用场景均被窗口排除，该回退分支成为不可达死规（遗留自「当前首页卡片行为」，改造后永不触发）。spec 内自相矛盾：一边断言 NULL 行永不出现，一边要求卡片处理 NULL published_at。实现者会困惑到底哪种语义为准。 | 从 REQ-NTW-5 移除 NULL→created_at 回退子句及其 Scenario（或显式标注「防御性保留、窗口排除后不可达」），使卡片时间契约与窗口语义自洽；如确需保留回退（防未来卡片复用场景），在 REQ 中说明其适用边界 |
| S2 | notice-time-window/spec.md REQ-NTW-6 | 示例索引 `(status, published_at)` 无法直接服务 ORDER BY `is_pinned DESC, published_at DESC`（is_pinned 位于 status 与 published_at 之间），只能范围扫 published_at 后文件排序 is_pinned；需设计阶段确认（spec 已将索引选择委托给设计阶段并以 EXPLAIN 为验收门，故非阻塞） | 设计阶段考虑 `(status, is_pinned, published_at)` 或经 EXPLAIN 确认 (status, published_at) 窗口内文件排序满足延迟预算；将所选索引写入 REQ-NTW-6 落地项 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | api-proto 中 `GetMarqueeNotices` 在本变更后确认无移动端消费方（跑马灯由列表派生）。变更正确保持其不动（删除 RPC 需先核对调用方），建议记入 BACKLOG 追踪后续调用方核验/废弃 |
| I2 | ListContentPosts 校验风格将出现分叉：page/pageSize 非法时静默取默认值，而 since_days 非法时拒绝 080005。新参数更严可接受，但建议设计阶段统一注释说明两套校验语义，避免后续维护误解 |
| I3 | REQ-NTW-2 描述非法参数含「non-numeric」，但 proto 字段为 int32，非数字在 proto 反序列化层即被拒，到不了后端校验——表述可精简为「≤0 或 >365」（不阻塞） |

## 问题跟踪表（v1 遗留项修复验证）
- S1（community_contacts 004 根因/双份 DDL）：已修复——REQ-CLP-2 显式声明根因 + 001 单一权威 + 004 幂等补救；已代码验证 001 DDL 存在且与模型一致
- S2（REQ-NTW-4 错误引用 REQ-HL-2）：已修复——新增 REQ-NTW-5 显式卡片视觉契约锚点，NTW-4 改引 NTW-5
- S3（「≤3」强制边界）：已修复——REQ-NTW-1 明确前端传 page_size=3、后端只强窗口过滤
- S4（跑马灯「同源」歧义）：已修复——REQ-NTW-3 明确跑马灯 = 同一首页列表（≤3，继承封顶），代码验证 marquee 由 notices 派生

---
VERDICT: APPROVED
---
