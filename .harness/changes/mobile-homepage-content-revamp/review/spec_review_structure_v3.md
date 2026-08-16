# Plan Review — mobile-homepage-content-revamp（结构合理性视角）

**审查维度**: 职责边界、一致性
**审查版本**: P1.3（fallback:r2:rc1，按磁盘最新内容独立重新审查，未沿用 r0/r1 结论）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 5

## 审查对象核验（磁盘最新 + 代码验证）
- 5 个 capability spec 与 proposal 7 个工作项一一对应：notice-time-window（首页通知/跑马灯/列表/后端窗口过滤）、notice-detail-preview（详情+附件预览）、function-entries（4 入口）、contact-list-page（联络页+补表）、homepage-layout（邻里互助占位/寻失保持/广告集中/垂直全序）。
- change.yaml services（web/mobile、community-hub-service、api-proto）与各 spec 服务职责边界一致；file-service 为复用（RPC GetFileUrl 无所有权限制、REST GET /api/files/:id 强制所有权），未列入 services 合理。
- 跨 capability 引用全部正确：FE-2→CLP-1、NTW-4→NTW-5、NTW-4→NTW-1、HL-4→NTW-1/3/5、HL purpose→NTW-3；首页垂直全序（REQ-HL-4）与 proposal 验收标准一致；联络拨号网格职责拆分清晰（FE-2 移除首页内嵌网格 → CLP-1 承接渲染，无重复）。
- 附件预览「无新增后端变更」边界成立：`services/community-hub-service/rpc/internal/logic/notice/getcontentpostlogic.go:100` 已调用 FileClient.GetFileUrl 重生 file_url；file-service REST 端点 `api/internal/logic/file/getfileurllogic.go:41-42` 强制 `rpcResp.File.UserId != userId → permission denied`，RPC GetFileUrl 无所有权校验。REQ-NDP-2/3「前端不直连 REST、消费详情响应重生 file_url」与真实 wire 一致。
- r2-2 REST 层透传边界完整：`api/internal/types/types.go:43-49` ListContentPostsReq 当前无 since_days（待新增 form 字段），`api/internal/logic/notice/listcontentpostslogic.go` 为薄代理（仅组 RPC 请求转发），透传可行；spec 边界已列全三环（types 字段 → REST logic 透传 → RPC logic 应用窗口谓词），.change.yaml revises 已补入 REST 两文件。
- 跑马灯「由首页通知列表派生、不改 GetMarqueeNotices」成立：notice.vue L280-282 marqueeText 为 computed，由 notices（getNoticeList→ListContentPosts）派生；移动端不消费 GetMarqueeNotices RPC；proto GetMarqueeNoticesRequest 保持不动。
- REQ-NTW-4 基线准确：notice-browse.vue L110-113 `getNoticeList(cid,1,50)` + 客户端 3 个月过滤——「原行为→新行为」diff 与代码一致。
- REQ-NTW-6 索引前提成立：`migration/001_initial.sql:24` idx_published(community_id, published_at DESC, deleted_at) 的 community_id 为弃用 NULL 列（003 L23），列表查询按 content_post_scope JOIN 过滤（model/content_post.go:154-158 FindListByCommunity），旧索引无法服务窗口过滤/排序；窗口谓词可落入 FindListByCommunity，职责边界（RPC 应用、REST 透传）无重叠。
- REQ-CLP-2 根因声明成立：community_contacts DDL 已存在于 001_initial.sql:39-49，与 CommunityContactModel 对齐；004 为幂等补救，001 为 schema 单一权威——v1 S1 已解决。
- api-proto since_days 为 additive：ListContentPostsRequest 现用字段 1-5（community_id=1/role=2/section_code=3/page=4/page_size=5），新增字段 6 不破坏兼容；缺省不过滤保留 PC 行为。
- REQ-NDP-4 与 wire 一致：proto ContentPostAttachment 已有 file_type=5、file_id=6（L87-95），前端类型扩展 file_id/file_type 与 wire 匹配。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| S1 | notice-time-window/spec.md REQ-NTW-5（L121、L128-131 Scenario「时间格式与回退一致」） | **v2 S1 遗留未修复**：REQ-NTW-5 规定卡片时间「formatted from published_at（fallback created_at when published_at is NULL）」并带测试场景断言 NULL→created_at 回退；但同 capability 的 REQ-NTW-1/REQ-NTW-4 明确「NULL published_at 行不进窗口」——首页与列表页两处卡片消费场景均被窗口排除，该回退分支成为不可达死规，spec 内自相矛盾。更关键：唯一 NULL published_at 可达的消费点是**详情页**（REQ-NDP-1 显示发布时间），而 REQ-NDP-1 未定义 NULL 回退行为——回退规则放在不可达处、缺失于可达处，职责分配错位 | 将 NULL→created_at 回退规则移入 REQ-NDP-1（详情页时间展示，含空值场景测试）；REQ-NTW-5 移除回退子句或显式标注「防御性保留、窗口排除后不可达」；统一两处口径 |
| S2 | .change.yaml revises 清单 | **新文件遗漏**：核心新交付 `web/mobile/src/pages/contact-list/contact-list.vue`（REQ-CLP-1 拨号网格页）未列入 revises；仅 pages.json（注册）在列。revises 已用注释区分「新建文件（非修订）」并收录 migration 004，contact-list.vue 应同模式列入，否则变更完整性追踪漏掉主交付物 | 在 revises 补列 `web/mobile/src/pages/contact-list/contact-list.vue`（标注新建） |
| S3 | notice-time-window/spec.md REQ-NTW-6（L135） | **v2 S2 遗留未修复**：示例索引 `(status, published_at)` 无法直接服务 ORDER BY `is_pinned DESC, published_at DESC`（is_pinned 夹在 status 与 published_at 之间），窗口内需文件排序 is_pinned。spec 已把索引选择委托设计阶段并以 EXPLAIN 为验收门，故非阻塞，但示例会误导设计选型 | 设计阶段考虑 `(status, is_pinned, published_at)` 或经 EXPLAIN 确认 (status, published_at) 满足延迟预算；将落地索引写回 REQ-NTW-6 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | REQ-NDP-2 图片白名单 `{png, jpg, jpeg, gif}` 含 `jpeg`，但 file-service `guard/magic.go:30-50` SniffType 只产出 {png, jpg, gif, pdf, doc, docx}（JPEG 一律归一为 jpg）——`jpeg` 为无害死条目，且「与 file-service 白名单对齐」表述略不精确；可保留（防御 legacy 值）但建议注明 |
| I2 | api-proto community.proto L54-55 GetMarqueeNotices 注释仍写「最近 15 天置顶优先倒序 ≤10 条」，本变更后移动端不再消费该 RPC；保持不动正确，建议记入 BACKLOG 核验调用方后同步注释/废弃 |
| I3 | .change.yaml revises 将 migration 004（新建）列于 revises 下并注释「新建文件（非修订）」——元数据可接受，但与新文件应入新建区的一般惯例略有出入（不阻塞） |
| I4 | ListContentPosts 校验语义将分叉：page/pageSize 非法时静默取默认值（rpc logic L59-61），since_days 非法时拒绝 080005。新参数更严可接受，建议设计阶段统一注释说明两套语义 |
| I5 | 变更目录无 request.md（仅有 .change.yaml + proposal.md + specs/），决策包 stage1_clarify（D1-D16）已由 proposal 决策日志承载；建议补 request.md 或确认 proposal 为需求单源（v1/v2 均提示） |

## 问题跟踪表
- S1（v2 遗留，REQ-NTW-5 NULL 回退死规/职责错位）：待修复——本轮重新定位至详情页可达性
- S2（v2 遗留，REQ-NTW-6 示例索引）：待修复——设计阶段定索引后回写
- S3（本轮新增，contact-list.vue 未列入 revises）：待修复
- v1 S1（004 双份 DDL 根因）：已修复并代码验证（001 权威 + 004 幂等补救）
- v1 S2（REQ-NTW-4 误引 REQ-HL-2）：已修复（REQ-NTW-5 锚点）
- v1 S3（「≤3」强制边界）：已修复（page_size=3 明示）
- v1 S4（跑马灯「同源」歧义）：已修复（REQ-NTW-3 明示继承 3 条封顶）

---
VERDICT: APPROVED
---
