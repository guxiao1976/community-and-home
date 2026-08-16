# Plan Review — mobile-homepage-content-revamp（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL 唯一解释、Scenario 落地性、术语一致）
**审查版本**: P1.3 · fallback:r1:rc1（本轮按磁盘最新内容独立审查；r0→r1 已更新：REVISION #1/#4/#6 重构附件路径、新增 REQ-NTW-5/6、REQ-CLP-2 补根因与漂移边界）

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 4

## 上轮（v1）问题修复验证
| v1 项 | 结论 |
|------|------|
| SHOULD #1（004 与 001 重复建表无唯一解释） | ✅ 已修复：REQ-CLP-2 显式声明根因（001 已应用后运行库仍缺表的补救）、DDL 与 001/CommunityContactModel 对齐、幂等、漂移边界（IF NOT EXISTS 不修复漂移需人工订正） |
| SHOULD #2（图片附件 file_id 缺失行为未定义） | ✅ 已修复：REVISION #4 重构主路径为「消费详情响应重生 file_url」，REQ-NDP-2/3 对称定义 file_url 不可用/失败 → 明确报错，不静默 |
| SHOULD #3（REQ-NTW-4 误引 REQ-HL-2） | ✅ 已修复：新增 REQ-NTW-5 通知卡片视觉契约锚点，REQ-NTW-4 改引 REQ-NTW-5；REQ-HL-2 回归寻失互助语义 |
| INFO #2（since_days 字段名未钉死） | ✅ 已修复：REQ-NTW-2 显式钉死 `since_days`（int32，天，1..365） |

## 交叉校验（Spec ↔ 磁盘真实契约，逐条核过）
- `file_type` 实际落库值 = **嗅探的规范扩展名**（`guard/magic.go:35` 返回 `"jpg"` 等，`extMatch` 按 jpg/jpeg 等价比较），**非 MIME**。
- 既有排序契约 `order by content_posts.is_pinned desc, content_posts.published_at desc`（model/content_post.go:191）与 REQ-NTW-1/4 一致。
- `idx_published (community_id, published_at DESC, deleted_at)`（001_initial.sql:24）+ 003 迁移确认 community_id/published_at 为弃用 NULL 列，REQ-NTW-6 前提成立。
- `ListContentPostsRequest` 现无 since_days（字段 1-5），additive 加字段 6 可行；page_size 上限 100（maxPageSize），REQ-NTW-4 场景 page_size=50 合法。
- file-service REST `GET /api/files/:id` 强制所有权（getfileurllogic.go:41 `UserId != userId → 拒绝`）；GetContentPost 已服务端经 RPC GetFileUrl 重生 file_url（getcontentpostlogic.go:37,74-100）——REQ-NDP-2/3 主路径前提成立。
- 移动端跑马灯已由首页通知列表派生（notice.vue:280-283 marqueeText 取自 notices，非 GetMarqueeNotices）——REVISION #6 成立。
- `NoticeAttachment` 现仅 id/file_name/file_url/file_size（community.ts:10-15），缺 file_id/file_type——REQ-NDP-4 成立。
- 首页 = `pages/notice/notice`（pages.json TabBar 首位）；`pages/notice-browse`、`pages/notice-detail` 均已注册——页面路径引用一致。
- 广告点击预留不跳转已是现状（notice.vue:307 onAdClick 注释）——REQ-HL-3 与现状一致。
- 空态文案一致：暂无通知公告 / 暂无联络信息 / 功能开发中。

## 发现

### 🔴 MUST FIX
| # | 文件:章节 | 问题 | 修复建议 |
|---|------|------|------|
| 1 | specs/notice-detail-preview/spec.md → REQ-NDP-2 (line 35) vs REQ-NDP-4 (line 68) | **图片/文档分发谓词自相矛盾且与 wire 不符**：REQ-NDP-2 场景将图片附件定义为 `file_type ∈ image/*`（MIME 通配符读法），REQ-NDP-4 场景却用 `file_type: "jpg"`；而 file-service 实际落库 `file_type` 为**嗅探扩展名**（"jpg"/"png"/"pdf"，guard/magic.go:35 返回规范扩展名，非 MIME）。实现者若按 REQ-NDP-2 的 `image/*` 前缀匹配实现 → `"jpg"` 永不命中 → 全部图片附件走文档分支（openDocument），REQ-NDP-2 全屏预览功能**业务不可用**。同一字段两种格式，SHALL 无唯一解释。 | 统一分发谓词为**扩展名白名单**（如 `file_type ∈ {jpg, jpeg, png, gif, webp}` 与 file-service 白名单一致），在 REQ-NDP-2 显式注明「wire file_type 为嗅探扩展名、非 MIME」；REQ-NDP-2/3/4 三处图片/文档判定口径同步为同一谓词。 |

### 🟡 SHOULD FIX
| # | 文件:章节 | 问题 | 建议 |
|---|------|------|------|
| 1 | specs/notice-time-window/spec.md → REQ-NTW-1 场景「窗口内置顶帖与更早非置顶帖并存时置顶优先」(line 30-33) | THEN 子句「the 1-hour-old notice is shown before the **20-day-old non-pinned** ones」自相矛盾：GIVEN 中唯一 20 天前的帖子是**置顶**帖，不存在「20-day-old 非置顶帖」，读者无法确定意图（应为 1 小时前非置顶在其余较旧非置顶之前）。 | 改写为「the 1-hour-old non-pinned notice is shown before the older in-window non-pinned notices」，与 SHALL 的排序规则（置顶优先 → published_at 倒序）逐字对齐。 |
| 2 | specs/notice-time-window/spec.md → REQ-NTW-5 (line 119,126-129) vs REQ-NTW-1/4 & specs/notice-detail-preview/spec.md → REQ-NDP-1 | 跨能力一致性格差：REQ-NTW-5 定义卡片时间「published_at 缺省回退 created_at」，但 REQ-NTW-1/4 已把 published_at 为 NULL 的行**排除在首页/列表窗口外**，该回退在其卡片使用页永不触发（死规则）；而唯一可能展示 NULL published_at 行的是**详情页**（按 id 取、不过窗口），REQ-NDP-1 对 published_at 为 NULL 时的时间展示**未定义**。实现者对「详情页 NULL 时间显示什么」无唯一解释。 | 明确回退适用范围：要么声明 REQ-NTW-5 回退仅属共享卡片组件（首页/列表不触发，属防御），并在 REQ-NDP-1 补一条「published_at 为 NULL 时详情页时间显示回退 created_at」与卡片口径一致；要么显式声明详情页对 NULL 时间的处理。 |
| 3 | specs/notice-time-window/spec.md → REQ-NTW-2 (line 52) | 「values ≤0、>365、或 **non-numeric**」中 non-numeric 对 int32 proto 字段是不成立的表述（proto wire 不可能传非数字，REST 网关在解析层即拒），会让实现者疑惑「非数字怎么进到参数校验」。 | 精简为「values ≤0 or >365 SHALL be rejected with the parameter-invalid error」；如需覆盖网关层，单独注明 REST 解析失败返回 080005。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | specs/notice-time-window/spec.md → REQ-NTW-6：content_posts 窗口索引的迁移编号未钉死（「设计阶段定迁移编号」），而 REQ-CLP-2 / .change.yaml 已固定 `004_add_community_contacts.sql`。设计阶段须确保索引迁移编号 ≠ 004（建议 005），避免同变更内两个 community-hub 迁移撞号。 |
| 2 | specs/notice-time-window/spec.md → REQ-NTW-4 (line 85)：SHALL 写「固定 page_size」但具体值只出现在场景（50）与 D11（「如 50」）。建议在 SHALL 直接钉死 page_size=50（与现状 notice-browse 请求 50 一致），避免实现者取值分歧。 |
| 3 | specs/homepage-layout/spec.md → REQ-HL-1 (line 11)：占位区块「点击不导航」已定义，但未定义点击是否弹 toast（REQ-FE-3 对 3 个入口占位定义了「功能开发中」toast）。建议为邻里互助占位明确「点击弹『互助功能开发中』toast + 不跳转」，与 FE-3 交互范式对齐，避免实现者自行猜测。 |
| 4 | specs/notice-detail-preview/spec.md → REQ-NDP-2/3：图片/文档两分支均已定义「失败明确报错」，但失败提示文案未统一（NDP-2 无文案示例、NDP-3 有「附件打开失败」）。建议两分支共用同一失败文案常量，保持一致性。 |

---
VERDICT: REVISION（清晰可执行视角存在 1 项 MUST FIX：REQ-NDP-2 图片分发谓词与 wire/NDP-4 自相矛盾，按字面实现即图片预览不可用）
---
