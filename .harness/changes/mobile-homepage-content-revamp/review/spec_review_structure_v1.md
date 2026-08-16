# Plan Review — mobile-homepage-content-revamp（结构合理性视角）

**审查维度**: 职责边界、一致性
**审查版本**: P1.3（fallback:r0:rc1，按磁盘最新内容独立审查）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3

## 审查对象核验
- 5 个 capability spec 与 proposal 7 个工作项一一对应：notice-time-window（通知窗口）、notice-detail-preview（详情+附件）、function-entries（4 入口）、contact-list-page（联络页+补表）、homepage-layout（邻里互助/寻失/广告）。
- change.yaml services（web/mobile、community-hub-service、api-proto）与各 spec 服务职责边界一致；file-service 为复用无契约变更，未列入 services 合理。
- 首页垂直区块链在各 spec 间自洽：通知(跑马灯+卡片) → 4 功能入口 → 邻里互助占位 → 寻失互助 → 广告位底部堆叠，无重叠无断层。
- 联络拨号网格职责拆分清晰：function-entries REQ-FE-2 移除首页内嵌网格，contact-list-page REQ-CLP-1 承接渲染——无重复。

## 发现

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| S1 | proposal.md 影响范围（community-hub-service 行）+ contact-list-page/spec.md REQ-CLP-2 | `community_contacts` 表 DDL 已存在于 `services/community-hub-service/migration/001_initial.sql`（L39-49，自初始提交 7df7827 起就在），列结构与 REQ-CLP-2 完全一致。proposal 声称「当前运行库缺表导致 ListContacts 报 Table doesn't exist」并新增 migration 004 重新建同一张表——若运行库缺表，根因应是迁移应用缺失/库漂移，而非缺 DDL；新增 004 造成同表 DDL 双份定义（001 与 004），且未解释缺表根因，004 可能同样因迁移未应用而不生效 | 在 spec 中补充缺表根因分析（确认运行库是否应用过 001、是否库漂移）；若确认 001 未被应用，应先在部署/迁移流程补跑 001 而非新增重复 DDL；或至少注明 004 是对漂移库的幂等补救，并统一 001/004 的 schema 单一权威来源 |
| S2 | notice-time-window/spec.md REQ-NTW-4 | 列表页卡片样式引用「see REQ-HL-2 相关卡片契约复用」——但 REQ-HL-2 是「寻失互助区块保持现状」需求（homepage-layout/spec.md），与通知卡片（role 色条/标签/时间）无关联，属跨 capability 引用错误；通知卡片样式契约实际由 REQ-NTW-1 场景隐含定义，无独立 REQ 锚点 | 将 REQ-NTW-4 的样式引用改为指向 REQ-NTW-1（首页通知卡片样式，role 色条/标签/时间）或在本 spec 内显式定义「通知卡片样式契约」REQ，移除对 REQ-HL-2 的错误引用 |
| S3 | notice-time-window/spec.md REQ-NTW-1 + REQ-NTW-2 | 首页「最多 3 条」封顶的强制边界未明确：REQ-NTW-2 规定 30 天窗口由后端强制、前端只传参数，但「≤3」由谁执行未说明——是前端传 page_size=3（复用既有分页参数）还是后端新增封顶契约？若走前端 page_size，则「≤3」非后端强制，与「时间口径由后端统一」的职责边界存在不一致表述 | 在 REQ-NTW-1/REQ-NTW-2 明确「≤3」的实现路径：明确前端以 page_size=3 传递（后端仅强窗口过滤，天然截断为 3），或明确后端新增封顶逻辑；保持与 D10「后端强时间口径、前端不实现业务逻辑」的边界一致 |
| S4 | notice-time-window/spec.md REQ-NTW-3 | 「同源 30 天数据」定义模糊：跑马灯走 GetMarqueeNotices/FindMarquee（MarqueeLimit=10、独立排序/完整性谓词），首页卡片走 ListContentPosts（封顶 3）；当 30 天窗口内通知 >3 条时，跑马灯（最多 10 条）与卡片（3 条）成员/顺序不一致，「同一 30 天数据集」的表述在实现层不成立，REQ-NTW-3 场景仅用 3 条覆盖了二者一致的情形，未覆盖分歧场景 | 将「同源 30 天数据」精确定义为「同一 30 天时间窗口过滤条件」（非同一成员集）；补充 >3 条时跑马灯与卡片成员不一致场景的预期，明确跑马灯不受卡片 3 条封顶约束 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | 变更目录无 `request.md`（仅有 proposal.md + .change.yaml + specs/），reviewer 上下文清单预期存在 request.md；proposal 决策日志 D1-D9 作为需求源头，建议确认决策包 stage1_clarify 是否已在 proposal 完整承载 |
| I2 | `.change.yaml` `revises` 将 `migration/004_add_community_contacts.sql` 列为修订文件，但该文件当前不存在（migration 目录仅 001/002/003）——属新建而非修订，元数据分类有误（不阻塞） |
| I3 | REQ-NTW-4 未提及列表页分页去留：proposal 称「翻页浏览→与首页一致的列表」，但 ListContentPosts 为分页接口，列表页是保留 page/page_size 分页还是展示 30 天内全量未定义；建议明确以消除实现歧义（非结构阻塞） |

## 问题跟踪表
- S1-S4 状态：待修复（本视角无 MUST FIX，SHOULD FIX 可进入阶段 3，遗留记入 BACKLOG）

---
VERDICT: APPROVED
---
