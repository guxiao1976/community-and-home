# Plan Review — mobile-homepage-content-revamp（覆盖完整性视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别
**审查版本**: P1.3 fallback:r0:rc1（首轮，无历史 review 目录）
**审查对象**: proposal.md + specs/{notice-time-window, notice-detail-preview, function-entries, contact-list-page, homepage-layout}/spec.md（对照 .change.yaml 决策包 D1-D10、验收标准、out_of_scope）

## 摘要
- 🔴 MUST FIX: 3 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 1

## 覆盖矩阵核验（决策点 → REQ）

| 决策点 | 覆盖 REQ | 结论 |
|--------|---------|------|
| D1 跑马灯 + 3 条卡片同源 30 天 | REQ-NTW-1 / REQ-NTW-3 | ✅ |
| D2 列表页同样 30 天过滤 | REQ-NTW-4 | ✅（边界缺失见 MUST FIX #2） |
| D3 4 功能入口（便民做实 + 3 占位） | REQ-FE-1/2/3 + REQ-CLP-1 | ✅ |
| D4 不预置联络种子 | REQ-CLP-2 | ✅ |
| D5 附件预览统一 file-service | REQ-NDP-2/3 | ✅（wire 含 file_id/file_name/file_size/file_type，已核实 api-proto ContentPostAttachment） |
| D6/D7 广告集中 + 点击预留 | REQ-HL-3 | ✅ |
| D8/D9 邻里互助占位无后端无页面 | REQ-HL-1 | ✅ |
| D10 时间窗口后端强制 | REQ-NTW-2 | ✅ |

验收标准 8 条 → 全部映射到 REQ。out_of_scope 8 条 → 全部显式对齐（HL-1 邻里互助、HL-3 广告、CLP-2 空态、HL-2 寻失保持）。已核实事实：migration 001 已定义 community_contacts（列与 CommunityContactModel 完全一致，无 deleted_at）；file-service REST 为 `GET /api/files/:id`；ListContacts 路由 `GET /api/community/contacts` 存在；当前 detail 页确为 `onDownload(att.file_url, …)` 直链（本次改造目标）。

## 发现

### 🔴 MUST FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | specs/notice-time-window/spec.md · REQ-NTW-4（line 68） | 交叉引用错误：「render in a list style consistent with the homepage notice cards … see REQ-HL-2 相关卡片契约复用」。REQ-HL-2 是 homepage-layout 的「寻失互助区块保持现状」，与通知卡片无关；全 spec 没有任何 REQ 定义首页通知卡片的视觉契约（role 色条/标签/时间）。「与首页卡片同风格」的决策点引用了不存在/错误的契约源。 | 修正引用为 REQ-NTW-1（首页通知区）或在 homepage-layout 新增一条 REQ 明确通知卡片视觉契约（色条按 role、标签、时间格式），并让 REQ-NTW-4 指向它。 |
| 2 | specs/notice-time-window/spec.md · REQ-NTW-4（line 66-68） | 列表页分页/条数上限未定义。proposal §影响范围明确「由翻页浏览改为与首页一致的列表」（现有 notice-browse.vue 有翻页按钮），但 REQ-NTW-4 只写「仅展示 30 天内、倒序」，未指定：30 天窗口内超过一屏（如 50 条）时是固定上限、触底加载还是全量返回；原分页参数 page/page_size 是否沿用。实现者将产出不同行为。 | 在 REQ-NTW-4 补充列表页数据量行为：明确条数上限（或分页/触底加载策略）及超限时的截断语义；说明与 ListContentPosts page/page_size 的关系。 |
| 3 | specs/notice-time-window/spec.md · REQ-NTW-1/2/4 | published_at 为 NULL / 未来时间的边界未定义。migration 003 已把 published_at 改为 `DEFAULT NULL`（审核锚定，submit 即置 NOW()），窗口谓词 `published_at >= now - 30 天`：NULL 行在 SQL 中恒不匹配 → 会从首页/列表被静默剔除（即使 status=approved）；published_at > now 的预排期行则会被包含进窗口（未来通知提前上首页）。两种行在窗口语义下均无明确行为。 | 在 REQ-NTW-1 增补边界 Scenario：a) approved 但 published_at 为 NULL 的行在窗口内的处理（排除还是显示）；b) published_at > now（预排期）的行是否被 30 天窗口包含（建议排除 published_at > now）。 |

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 4 | specs/notice-time-window/spec.md · REQ-NTW-3 | 跑马灯 15→30 天窗口对齐后，排序（置顶优先倒序）与条数上限（既有 GetMarqueeNotices 契约为「置顶优先倒序 ≤10 条」，api-proto line 54 注释可证）未重述；「同源 30 天数据集」未说明跑马灯是取全部窗口内标题还是继承 ≤10 上限，也未说明与首页 3 卡片的基数差异。 | REQ-NTW-3 明确：跑马灯在 30 天窗口内沿用置顶优先倒序与 ≤10 条上限（或显式改为全量/与 3 卡片对齐），并注明与首页卡片基数的关系。 |
| 5 | specs/notice-time-window/spec.md · REQ-NTW-2（line 47-50） | 非法窗口参数仅定义了 zero/negative/非数值 → 080005；未定义上界（如 since_days=9999 是否合法）。 | 补充参数取值范围（如 1~365）或说明超上界的处理。 |
| 6 | specs/homepage-layout/spec.md · 全局 | 首页整页纵向区块顺序（通知 → 4 入口 → 邻里互助 → 寻失互助 → 底部广告）仅在 proposal §做什么 与各能力 Purpose 中隐式描述，未在任一 REQ 固定完整顺序。 | 在 REQ-HL 新增/明确一条首页区块全序（含跑马灯位置），作为前端布局唯一依据。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 7 | migration 004 与 migration 001 的 community_contacts 定义重复（001 已建表），004 以 CREATE TABLE IF NOT EXISTS 补齐运行库属合理补救，建议 REQ-CLP-2 注明「001 已定义、004 为运行库缺失补救」，避免后续误删 001 定义。 |

## 问题跟踪表
| # | 状态 |
|---|------|
| 1（交叉引用） | 待修复 |
| 2（列表分页） | 待修复 |
| 3（published_at 边界） | 待修复 |
| 4/5/6/7 | 待评估（SHOULD/INFO） |

---
VERDICT: REVISION
---
