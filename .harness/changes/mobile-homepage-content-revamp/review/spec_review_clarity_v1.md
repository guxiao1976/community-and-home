# Plan Review — mobile-homepage-content-revamp（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL 唯一解释、Scenario 落地性、术语一致）
**审查版本**: P1.3 · fallback:r0:rc1（本轮按磁盘最新内容独立审查，无旧轮缓存）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 4

## 交叉校验（Spec ↔ 磁盘真实契约）
以下断言逐一与代码交叉核对，均成立，Spec 的术语与既有契约一致：
- `published_at` 为社区 proto 真实字段（`community.proto:77`），REST 响应亦有 `json:"published_at"`（types.go:182）；符合规则 3.1 术语一致。
- 既有排序契约 `order by is_pinned desc, published_at desc`（listcontentpostslogic.go:30）与 REQ-NTW-1「置顶优先 + 倒序」一致，无歧义。
- GetMarqueeNotices 现窗口 `MarqueeWindowDays = 15`（getmarqueenoticeslogic.go:14），与 REQ-NTW-3「15 天 → 30 天对齐」前提一致。
- 附件字段 `file_id / file_type / file_url / file_size` 均在 proto（`ContentPostAttachment`）与 REST types.go:195-198 暴露，REQ-NDP-2/3 可落地。
- 错误码 080005 = 参数无效（community-hub CLAUDE.md 错误码表 + contentcompat 实际使用），REQ-NTW-2 引用正确。
- file-service 下载路由确为 `GET /api/files/:id`（routes.go:36 + prefix `/api/files`），REQ-NDP-3 引用正确。
- ListContacts REST 路径 `GET /api/community/contacts` 存在（community.ts:144），REQ-CLP-1 引用正确。
- 详情页现已渲染 title/role/时间/publisher/content/附件（notice-detail.vue），REQ-NDP-1「保持完整展示」与现状一致。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX
| # | 文件:章节 | 问题 | 建议 |
|---|------|------|------|
| 1 | specs/contact-list-page/spec.md → REQ-CLP-2 | Spec 将 `community_contacts` 描述为「新增 migration 004 补表」，但 `migration/001_initial.sql:39-49` **已存在完全相同的 `CREATE TABLE IF NOT EXISTS community_contacts`（列名/索引逐字一致）**。Spec 未acknowledge 001 已声明该表、也未说明运行库为何缺表（合理推因：001 在已应用后又被回溯追加该表 DDL——迁移反模式）。实现者面对「004 与 001 重复」无法唯一判断：是照抄 001 DDL、还是应先核查运行库真实状态（缺表 vs 表存在但结构漂移——`IF NOT EXISTS` 对漂移不生效）。 | 在 REQ-CLP-2 显式声明根因（001 已应用后才补入该 DDL，故需 004 独立补表），声明 004 DDL 必须与 CommunityContactModel（id/community_id/category/name/phone/sort_order/created_at/updated_at + community_id 索引）完全对齐，并补一条「运行库表存在但结构漂移 → 需人工订正」的边界说明，避免错误假设漂移也能被 IF NOT EXISTS 修复。 |
| 2 | specs/notice-detail-preview/spec.md → REQ-NDP-2 | 图片附件当 `file_id` 缺失（0/空）时的行为**未定义**，与 REQ-NDP-3（文档附件显式定义 `file_id` 缺失/解析失败 → 回退 `file_url`）不一致。实现者可能实现「file_id 缺失 → 报预览失败」（使全部遗留图片附件无法预览，违背 proposal D5「file_id 权威、file_url 回退」通用原则）或「回退 file_url」，两者行为分歧。 | 在 REQ-NDP-2 补一条与 REQ-NDP-3 对称的语句：图片附件 `file_id` 缺失/解析失败时回退 stored `file_url` 预览；两路径均失败才报明确错误。统一两种附件类型的回退语义。 |
| 3 | specs/notice-time-window/spec.md → REQ-NTW-4 | 交叉引用错误：REQ-NTW-4 写「list style consistent with the homepage notice cards (role color bar / role tag / publish time, **see REQ-HL-2** 相关卡片契约复用)」，但 **REQ-HL-2 是「寻失互助区块保持当前展示风格」**，与通知卡片契约无关。同时 REQ-NTW-4 仅写「倒序」未复述置顶优先，与首页口径（REQ-NTW-1 置顶优先）是否一致未显式声明。 | 交叉引用改为 REQ-NTW-1（首页通知卡片契约）或在 REQ-NTW-4 内联定义卡片契约；并显式声明列表页沿用同一排序契约（置顶优先 + published_at 倒序），与首页口径完全一致。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | REQ-NTW-1 无「窗口内置顶帖与非置顶新帖并存」的 Scenario 来锁定 top-3 截断与置顶优先的交互（如 3 条置顶 + 1 条新帖时仅显示 3 条置顶）。需求文本已唯一，但补一条 Scenario 可固化实现。 |
| 2 | 新 proto 参数名仅在 proposal 以「如 since_days」出现，Spec 未钉死字段名/类型；前端传参与后端过滤需在全局 Claude 生成 proto 后对齐。建议在 REQ-NTW-2 记录建议字段名 `since_days`（int32，天），避免前端实现等待。 |
| 3 | REQ-NTW-3「跑马灯与卡片同源 30 天数据」表述偏松：GetMarqueeNotices 有 ≤10 置顶优先封顶，>10 条窗口内通知时跑马灯集合是卡片 top-3 的超集。可接受，但补一条「窗口内 >10 条」Scenario 可消除歧义。 |
| 4 | 变更目录无 `request.md`，需求唯一来源为 proposal 决策日志 D1-D10（用户已拍板）。过程性提示，非 Spec 缺陷。 |

---
VERDICT: APPROVED（清晰可执行视角无 MUST FIX；3 项 SHOULD FIX 建议进入 BACKLOG 技术债跟踪）
---
