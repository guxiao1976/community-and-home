# Plan Review — mobile-homepage-content-revamp（覆盖完整性视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别
**审查版本**: P1.3 fallback:r1:rc1（与上轮 r0:rc1 哈希不同，按磁盘最新内容独立重新审查，未沿用旧轮结论）
**审查对象**: proposal.md + specs/{notice-time-window, notice-detail-preview, function-entries, contact-list-page, homepage-layout}/spec.md（对照 .change.yaml 决策包 D1-D14、验收标准、out_of_scope；request.md 不存在，以 .change.yaml + proposal 为原始需求对照）
**事实核验**: 已对照磁盘代码核验 spec 依赖的全部事实锚点（见下文）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2
- 上轮 MUST FIX ×3 / SHOULD ×3 / INFO ×1 → 本轮全部修复并核验

## 事实锚点核验（spec 依赖的既有行为/契约均属实）

| 锚点 | 核验结果 |
|------|---------|
| migration 001 已定义 community_contacts（id/community_id/category/name/phone/sort_order/created_at/updated_at + idx_community，无 deleted_at） | ✅ 001_initial.sql:39-49 与 CommunityContactModel 完全一致 |
| 移动端跑马灯由首页通知列表派生（非 GetMarqueeNotices） | ✅ notice.vue:280-283 marqueeText 由 notices 派生；community.ts 无 getMarqueeNotices 引用（REVISION #6 前提成立） |
| notice-browse.vue 现请求 page_size=50 + 客户端 3 个月过滤 | ✅ notice-browse.vue:110-114 |
| 首页现状：内嵌联络拨号网格 + 广告位 ×2（联络下方）+ 广告位 ×1（寻失下方） | ✅ notice.vue:92-138, 198-200（REQ-HL-3 前提成立） |
| GetContentPost 服务端重生 file_url（GetFileUrl RPC，file_id=0 回退 stored URL） | ✅ getcontentpostlogic.go:74,98-111 |
| file-service REST GET /api/files/:id 强制所有权（非上传者 permission denied） | ✅ getfileurllogic.go:41-43（REVISION #4 前提成立） |
| ContentPostAttachment wire 含 file_id/file_name/file_url/file_size/file_type | ✅ community.proto:87-95（REQ-NDP-4 前提成立） |
| ListContentPosts 现有排序契约 is_pinned DESC, published_at DESC（NULLS LAST） | ✅ listcontentpostslogic.go:30 |
| ListContacts 路由 GET /api/community/contacts 存在 | ✅ routes.go:57 |
| api REST 层 ListContentPostsReq 现无 since_days 字段 | ✅ types.go:43-49（见 SHOULD #2） |
| file-service 白名单扩展名：图片 png/jpg/jpeg/gif；文档 pdf/doc/docx | ✅ guard/whitelist.go（见 SHOULD #1） |

## 覆盖矩阵核验（决策点 → REQ）

| 决策点 | 覆盖 REQ | 结论 |
|--------|---------|------|
| D1 跑马灯 + 3 条卡片同源 30 天（REVISION #6） | REQ-NTW-1/3 | ✅ |
| D2 列表页同样 30 天过滤 | REQ-NTW-4 | ✅ |
| D3 4 功能入口（便民做实 + 3 占位） | REQ-FE-1/2/3 + REQ-CLP-1 | ✅ |
| D4 不预置联络种子 | REQ-CLP-2 | ✅ |
| D5 附件预览统一 file-service（REVISION #4） | REQ-NDP-2/3/4 | ✅ |
| D6/D7 广告集中 + 点击预留 | REQ-HL-3 | ✅ |
| D8/D9 邻里互助占位无后端无页面 | REQ-HL-1 | ✅ |
| D10 时间窗口后端强制 | REQ-NTW-2 | ✅ |
| D11 列表页数据量行为（REVISION #2） | REQ-NTW-4 | ✅ |
| D12 published_at NULL/未来边界（REVISION #3） | REQ-NTW-1/4 | ✅ |
| D13 索引/性能（REVISION #5） | REQ-NTW-6 | ✅ |
| D14 通知卡片视觉契约（REVISION #1） | REQ-NTW-5 | ✅ |

验收标准 10 条 → 全部映射到 REQ；out_of_scope 10 条 → 全部显式对齐（HL-1 邻里互助、HL-3 广告、CLP-2 空态、HL-2 寻失保持、GetMarqueeNotices 不变、file-service REST 不直连等）。

## 上轮问题修复核验（问题跟踪表）

| # | 上轮问题 | 状态 |
|---|---------|------|
| 1（MUST）REQ-NTW-4 错误引用 REQ-HL-2 | 已修复——新增 REQ-NTW-5 显式定义卡片视觉契约，REQ-NTW-4 改引用 REQ-NTW-5（REVISION #1） |
| 2（MUST）列表页分页/条数上限未定义 | 已修复——REQ-NTW-4 新增「单请求、固定 page_size=50、窗口内截断、total 反映窗口全量」场景（REVISION #2） |
| 3（MUST）published_at NULL/未来边界未定义 | 已修复——REQ-NTW-1 新增 NULL 排除 + 预排期排除两边界场景（REVISION #3） |
| 4（SHOULD）跑马灯排序/条数上限未重述 | 已修复——REQ-NTW-3 明确由首页列表派生、继承 ≤3 上限与顺序（REVISION #6） |
| 5（SHOULD）窗口参数上界未定义 | 已修复——REQ-NTW-2 明确有效值 1..365，>365/≤0/非数值 → 080005 |
| 6（SHOULD）首页区块全序未固定 | 已修复——新增 REQ-HL-4 固定垂直全序 |
| 7（INFO）001/004 定义重复未说明 | 已修复——REQ-CLP-2 显式声明「001 为 schema 单源、004 为运行库漂移的幂等补救，不得删除/偏离 001 定义」 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | specs/notice-detail-preview/spec.md · REQ-NDP-2（Scenario line 36） | 图片/文档分发的判定谓词不一致且与真实 wire 不符：REQ-NDP-2 场景写「`file_type` ∈ image/*」（MIME 前缀式），REQ-NDP-4 场景却用 `file_type: "jpg"` → previewImage（扩展名式）。已核验 file-service 白名单，wire 的 file_type 是规范小写扩展名（图片 png/jpg/jpeg/gif、文档 pdf/doc/docx），不存在 `image/*` 值——按 REQ-NDP-2 字面实现 `startsWith('image/')` 将导致所有图片附件被当文档处理，图片预览功能整体不可用；两 REQ 并存使实现者无单一确定谓词。 | 在 REQ-NDP-2 显式固定图片类型集合：`file_type` ∈ {png, jpg, jpeg, gif}（与 file-service 白名单对齐）→ 图片；其余 → 文档（REQ-NDP-3 回退）。删除 `image/*` 表述，统一 REQ-NDP-2/4 谓词。 |
| 2 | .change.yaml revises 列表 + proposal §影响范围 | `since_days` 的 REST API 层透传文件未列入变更清单：已核验 mobile 经 `GET /api/community/notices`（REST）→ api/internal/types.go ListContentPostsReq（现无 since_days）→ api/internal/logic/notice/listcontentpostslogic.go → RPC。revises 只列了 `rpc/internal/logic/notice/listcontentpostslogic.go` 与 community.proto，未列 REST 层 types.go 与 api logic——若任务清单仅从 revises 派生，since_days 将在 REST 层被丢弃，30 天窗口在移动端路径上静默失效（前端看不到错误，但窗口不生效）。 | 将 `services/community-hub-service/api/internal/types/types.go`（ListContentPostsReq 加 since_days form 字段）与 `api/internal/logic/notice/listcontentpostslogic.go`（透传 RPC）补入 revises 清单，或在 REQ-NTW-2 服务职责边界显式点明「REST 层透传 since_days」。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 3 | REQ-NTW-5 的卡片时间 `created_at` 回退在首页/列表实际不可达（NULL-published_at 行已被窗口排除，永不渲染卡片），而唯一可能展示 NULL-published_at 行的详情页（REQ-NDP-1）反而未定义发布时间回退（formatTime 对空值返回 ''）。建议在 REQ-NDP-1 明确详情页发布时间对 NULL 的处理（回退 created_at 或显示「-」），并顺手说明 REQ-NTW-5 回退仅为防御性定义。 |
| 4 | 首页跑马灯条「更多 →」入口（现状 notice.vue:49-54 点击进 notice-browse）在 REQ-HL-4 的「notice section (marquee bar + up-to-3 notice cards)」中未声明是否保留。若该入口被移除且卡片只进详情，列表页将从首页失去可达路径。建议 REQ-HL-4/REQ-NTW-3 明确「更多 →」保留与否。 |

## 问题跟踪表
| # | 状态 |
|---|------|
| 1（图片谓词） | 待评估（SHOULD） |
| 2（REST 透传清单） | 待评估（SHOULD） |
| 3/4 | 待评估（INFO） |

---
VERDICT: APPROVED
---
