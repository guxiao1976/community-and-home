# Proposal: 移动端「社区家园」首页信息架构改造

> **优先级**: P1 · **改动规模**: 大（L） · **影响风险**: 中
> **核心风险点**: 通知时间窗口过滤涉及后端 ListContentPosts 契约（需 api-proto 加参数，proto 变更仅全局 Claude 执行，本管线 Owner 处理）；`since_days` 必须贯通 REST 层透传（`api/internal/types.ListContentPostsReq` → `api/internal/logic/notice/listcontentpostslogic.go` → RPC，REVISION r2-2——仅列 RPC logic 会致 REST 层丢弃参数、移动端 30 天窗口静默失效）；社区家首页作为 TabBar 首个页面，信息架构重排需保证空态/异常态不回归；30 天窗口过滤的索引/性能需在设计中评估（REVISION #5）；附件图片/文档分发谓词必须以 file_type 扩展名白名单判定（REVISION r2-1/r2-3/r2-4，wire file_type 为嗅探扩展名、非 MIME）。
> **变更类型**: modify（改造既有「移动端首页/通知列表/通知详情」+ 新增联络列表页 + 后端 migration + api-proto 参数扩展；原行为→新行为 diff 见 §影响范围）
> **设计文档**: 用户已确认决策包 `stage1_clarify`（D1-D14，映射见 §决策日志）。需求分析阶段澄清由用户拍板完成，本变更直接形式化；首轮 REVISION 的 6 项评审 MUST/SHOULD 已逐条解决（见 §决策日志 REVISION 标注与各 spec 修订记录）。**第 2 轮 REVISION（r2-1~r2-6，本轮）** 已修订：r2-1/r2-3/r2-4 = 附件图片/文档分发谓词由 `file_type ∈ image/*`（MIME 式，wire 无此值）统一为扩展名白名单 `{png, jpg, jpeg, gif}`；r2-2 = `since_days` REST 层透传文件补入 revises 清单 + REQ-NTW-2 职责边界；r2-5 = REQ-NTW-2 参数校验措辞去掉 non-numeric；r2-6 = 附件重生整单失败边界归入 REQ-NDP-1 详情加载失败态（见 §决策日志 D15/D16 与各 spec 修订记录）。

## 为什么做

移动端「社区家园」首页目前信息架构散乱：通知无时间范围约束（>30 天的置顶旧帖也可能上首页）、3 个广告位分散在页面中部打断内容流、便民联络直接内嵌拨号网格占据首页空间、缺少「邻里互助」与 4 个常用功能入口。用户希望首页聚焦「通知 + 4 个功能入口 + 邻里互助占位 + 寻失互助 + 底部广告区」的清晰结构，并让通知类内容（首页/列表/详情）的时间口径、排序、附件预览体验统一。本轮让首页信息架构一次性理顺，同时为便民联络等模块补齐缺失的数据基建。

## 做什么

1. **通知时间窗口统一（web/mobile + community-hub-service + api-proto）**：首页通知只展示最近 30 天内（按 `published_at`）、最多 3 条、发布时间倒序（最新在上，窗口内置顶优先）；跑马灯滚动内容取自同一 30 天数据（由首页通知列表派生，非独立跑马灯 RPC）；通知列表页同样只展示 30 天内、置顶优先 + 发布时间倒序，与首页口径一致。时间窗口由后端强制（前端不实现窗口业务逻辑），`ListContentPosts` 契约新增可选时间窗口参数 `since_days`（api-proto 变更，Owner 处理），并**贯通 REST 层透传**：`api/internal/types.ListContentPostsReq` 新增 `since_days` form 字段、`api/internal/logic/notice/listcontentpostslogic.go` 透传至 RPC（REVISION r2-2）。窗口谓词为 `published_at >= now - since_days AND published_at <= now`：`published_at` 为 NULL 的行与未来预排期行均不进窗口（REVISION #3 边界）。
2. **通知详情页 + 附件预览（web/mobile）**：详情完整展示标题、发布单位（role）、发布时间、内容、附件；附件点击统一走 file-service：图片 `uni.previewImage` 全屏预览，文档经下载详情响应中已重生的 file-service 签名 URL + `openDocument`。图片/文档分发按 `file_type` **扩展名白名单**判定：`file_type ∈ {png, jpg, jpeg, gif}` → 图片，其余（pdf/doc/docx 或缺失/无法识别）→ 文档（REVISION r2-1/r2-3/r2-4：wire `file_type` 为 file-service 嗅探落库的规范小写扩展名、**非 MIME**，`image/*` 值在 wire 不存在，统一 REQ-NDP-2/3/4 谓词）。**主路径直接消费详情响应 `file_url`**（community-hub GetContentPost 服务端经 RPC GetFileUrl 重生，无所有权限制），前端不直连 file-service REST `GET /api/files/:id`（该端点强制文件所有权，查看者非附件上传者会被拒，REVISION #4）。前端附件类型扩展 `file_id`/`file_type`（REQ-NDP-4）。附件重生整单失败（file-service 不可用/文件已删）→ 详情读整单失败 → 按详情加载失败态处理（REVISION r2-6）。
3. **4 个功能图标入口（web/mobile）**：通知下方放 4 个图标入口——便民联络（做实：跳转新建联络列表页承载拨号网格）、物业报修/二手闲置/租房卖房（仅入口占位，点击提示「功能开发中」）。首页原内嵌拨号网格移除，联络数据迁至联络列表页渲染。
4. **便民联络数据基建（community-hub-service）**：补 migration 004 幂等建 `community_contacts` 表（当前运行库缺表，ListContacts 报 Table doesn't exist；001 已声明同表 DDL，004 为对漂移库的幂等补救，DDL 与 001/模型对齐），不预置种子数据（空态，运营后续维护）。
5. **邻里互助占位区块（web/mobile）**：首页新增「邻里互助」占位区块；本期无后端数据源、无列表页/详情页、无用户发布入口（D8/D9 拍板）。
6. **寻失互助保持（web/mobile）**：首页寻失互助横滑卡片展示风格保持不变。
7. **广告位集中（web/mobile）**：3 个广告位从分散位置统一移到页面下部垂直堆叠集中展示，内容仍为前端硬编码，点击保持预留不跳转。

## 决策日志（Step 2，Step 8 追溯唯一依据）

| ID | 决策内容（结论） | 依据 |
|----|-----------------|------|
| D1 | **保留跑马灯 + 3 条通知卡片并存**：跑马灯滚动内容取自同一 30 天数据（由首页通知列表 `getNoticeList`→ListContentPosts 派生，即首页通知卡片标题集，≤3 条）；**不修改后端 GetMarqueeNotices**（移动端未消费该 RPC，原「15 天→30 天常量对齐」变更删除，REVISION #6 已解决） | 用户拍板 `notice-marquee` + REVISION #6 |
| D2 | **列表页同样过滤**：通知列表仅展示 30 天内（按 published_at）、置顶优先 + 发布时间倒序，与首页口径一致 | 用户拍板 `time-window-scope` + validity SF-2 |
| D3 | **功能入口**：便民联络做实（复用 ListContacts API + 补 community_contacts 表 + 新建联络列表页承载拨号网格）；物业报修/二手闲置/租房卖房 3 个仅入口占位（点击提示「功能开发中」） | 用户拍板 `function-entry` |
| D4 | **联络种子**：仅补 migration 004 幂等建 community_contacts 表（DDL 与 001/CommunityContactModel 对齐，001 为 schema 单源、004 为运行库缺失的幂等补救），不预置种子数据（空态，运营后续维护） | 用户拍板 `contact-seed` + structure S1 / clarity SHOULD #1 |
| D5 | **附件预览统一走 file-service**：主路径直接消费详情响应中已由 community-hub GetContentPost 服务端重生（RPC GetFileUrl，无所有权限制）的 `file_url`——图片 `uni.previewImage` 全屏预览，文档下载后 `openDocument`；`file_id` 为服务端重生键，**前端不直连 file-service REST `GET /api/files/:id`**（该端点强制文件所有权，查看者非附件上传者会被拒）。**REVISION #4 已解决**：原「file_id 权威、file_url 回退」机制与既有 wire 相悖（详情响应 file_url 即权威新鲜签名 URL；新帖落库 file_url 为占位空串，回退路径无 URL），主路径改为消费重生 file_url | 用户拍板 `attachment-preview`（意图：统一走 file-service）+ REVISION #4 |
| D6 | **广告位集中**：3 个广告卡片原样垂直堆叠到页面底部集中展示，内容仍为前端硬编码 | 用户拍板 `ad-consolidation` |
| D7 | **广告点击**：保持预留不跳转，本期仅完成位置集中 | 用户拍板 `ad-click` |
| D8 | **邻里互助页面**：本期无列表页/详情页，仅首页占位展示 | 用户拍板 `mutual-help-pages` |
| D9 | **邻里互助数据源**：本期不开发——邻里互助仅前端首页占位区块，无后端数据源、无用户发布入口 | 用户拍板 `mutual-help-source` |
| D10 | **时间窗口由后端强制**：30 天窗口过滤在服务端执行（ListContentPosts 契约新增可选时间窗口参数 `since_days`，int32，有效值 1..365），前端只传窗口参数/消费过滤结果，不在前端实现窗口业务逻辑；首页「最多 3 条」经现有分页参数 `page_size=3` 达成，窗口过滤仍后端强制 | 项目硬性约束派生（CLAUDE.md/项目编码规范：禁止前端定义业务逻辑、接口必须与 api-proto 一致）+ structure S3 |
| D11 | **列表页数据量行为（REVISION #2 已解决）**：列表页以 `since_days` + 固定 `page_size`（如 50）单请求拉取，服务端返回窗口内最新 page_size 条（置顶优先 + published_at 倒序），`total` 反映窗口内全量；本期不做客户端分页/触底加载，超一屏即截断为最新 page_size 条 | 评审 coverage #2 / structure I3 |
| D12 | **published_at 边界（REVISION #3 已解决）**：窗口谓词含上界 `published_at <= now`（排除未来预排期行）并排除 `published_at` 为 NULL 的行（NULL 恒不匹配窗口，即使 status=approved 也不显示）；首页/列表同口径 | 评审 coverage #3 |
| D13 | **索引/性能（REVISION #5 已解决）**：30 天窗口过滤 + `published_at` 倒序不走全表扫描；设计阶段为 content_posts 补充可服务 (status, published_at) 的索引（现有 idx_published 的 community_id 为弃用 NULL 列，无法服务 scope JOIN 后的 published_at 过滤/排序），并验证查询计划（EXPLAIN） | 评审 validity #2 |
| D14 | **通知卡片视觉契约锚点（REVISION #1 已解决）**：新增 REQ-NTW-5 显式定义通知卡片视觉契约（role 色条 / role 标签 / 标题 / 时间），列表页样式引用 REQ-NTW-5；原 REQ-NTW-4 对 REQ-HL-2（寻失互助）的交叉引用删除 | 评审 coverage #1 / clarity SHOULD #3 / structure S2 |
| D15 | **附件分发谓词 = file_type 扩展名白名单（REVISION r2-1/r2-3/r2-4 已解决）**：wire `file_type` 为 file-service magic-bytes 嗅探落库的规范小写扩展名（图片 {png, jpg, jpeg, gif}、文档 {pdf, doc, docx}），**非 MIME**，`image/*` 值在 wire 不存在；图片/文档分发统一按 `file_type ∈ {png, jpg, jpeg, gif}` 判定（图片），其余视为文档（REQ-NDP-3 回退），REQ-NDP-2/3/4 三处口径一致——删除原 `image/*` 表述（字面 `startsWith('image/')` 实现会致全部图片附件走文档分支、图片预览不可用） | 评审 coverage/clarity/validity 同轮 MUST（REVISION r2-1/r2-3/r2-4） |
| D16 | **REST 层透传 since_days + 附件重生整单失败边界（REVISION r2-2/r2-6 已解决）**：`since_days` 必须贯通 REST 层——`api/internal/types.ListContentPostsReq` 加 form 字段、`api/internal/logic/notice/listcontentpostslogic.go` 透传 RPC（补入 revises）；附件 `file_url` 服务端重生失败（file-service 不可用/文件已删）时 GetContentPost 读**整单失败** → REST 透错 → 详情页按 REQ-NDP-1 加载失败态处理，REQ-NDP-2/3 的「file_url 为空」逐附件降级仅限响应已返回但 file_url 为空（legacy 无重生可能） | 评审 coverage（REVISION r2-2）+ validity（REVISION r2-6） |

> 补充解释（D3 落地口径）：首页「便民联络」由内嵌拨号网格改为 4 功能入口之一，入口跳转新建联络列表页，拨号网格整体迁移至联络列表页承载。这是「通知下方放 4 个功能图标入口 + 新建联络列表页承载拨号网格」两句拍板决策的自然合并，若评审认为应保留首页内嵌网格，请在下游评审提出。

## 影响范围（原行为 → 新行为）

| 服务/前端 | 变更类型 | 说明 |
|------|:---:|------|
| web/mobile | 修改 | 首页 `pages/notice/notice.vue` 信息架构重排（固定垂直全序 REQ-HL-4）：通知区传 `since_days=30` + `page_size=3`（倒序、窗口内置顶优先）；跑马灯由首页通知列表派生（REQ-NTW-3）；原内嵌联络网格 → 4 功能图标入口（便民联络跳新页、3 个占位 toast）；新增邻里互助占位区块；寻失互助区块样式不变；3 个广告位从分散位置移除、在页面底部垂直堆叠集中展示（点击预留）。通知列表页 `pages/notice-browse/notice-browse.vue` 由「单条翻页浏览 + 客户端 3 个月过滤」改为「与首页一致的卡片列表」（REQ-NTW-5 视觉契约，`since_days=30` + 固定 page_size 单请求、置顶优先 + published_at 倒序、点卡片进详情，移除客户端时间过滤）。通知详情页 `pages/notice-detail/notice-detail.vue` 保持完整展示 + 附件预览改为主路径消费详情响应重生 `file_url`（图片 previewImage / 文档下载 + openDocument；**移除前端直连 file-service REST 与 stored file_url 直链**，REVISION #4）。**新增** `pages/contact-list/contact-list.vue`（联络拨号网格）+ `pages.json` 注册 + api/community.ts 增加 `since_days` 参数、`NoticeAttachment` 类型扩展 `file_id`/`file_type`（REQ-NDP-4） |
| community-hub-service | 修改 | **migration 004**：幂等建 `community_contacts` 表（CREATE TABLE IF NOT EXISTS，DDL 与 001/CommunityContactModel 对齐，补齐运行库缺失表，不预置种子）。**ListContentPosts 读路径支持可选时间窗口过滤**（`since_days`：`published_at >= now - since_days AND published_at <= now`，D10/D12；契约参数见 api-proto）——**REST 层透传 `since_days`**：`api/internal/types/types.go` ListContentPostsReq 新增 `since_days` form 字段、`api/internal/logic/notice/listcontentpostslogic.go` 透传至 RPC（REVISION r2-2，两文件补入 revises）。**content_posts 索引补充**（设计阶段：可服务 (status, published_at) 或 scope JOIN 后 published_at 排序，REQ-NTW-6 / D13）。**GetMarqueeNotices 不变**（REVISION #6：移动端跑马灯不消费该 RPC，删除原 15→30 天常量变更）。Contacts 读（ListContacts）模型已存在，仅补表即可用 |
| api-proto | 修改 | community/v1 `ListContentPostsRequest` 新增可选时间窗口参数 `since_days`（int32，天，有效值 1..365，additive，缺省不过滤；D10）；CHANGELOG 登记。**GetMarqueeNoticesRequest 不变**。**proto 变更仅全局 Claude 执行，本管线 Owner 处理** |
| file-service | 复用 | 附件预览主路径消费 community-hub 详情响应中已重生的 file-service 预签名 URL（GetContentPost 服务端 RPC GetFileUrl，无所有权限制）；**前端不调用 REST `GET /api/files/:id`**（该端点强制文件所有权，REVISION #4）；无契约变更 |
| 邻里互助后端 | 不做 | D9 拍板：本期无后端数据源、无模型/表/接口、无用户发布入口，不涉及 api-proto 邻里互助契约 |

## 风险评估

- **proto 契约扩展风险**：`ListContentPostsRequest` 加可选参数 `since_days` 为 additive，不破坏既有调用（PC 通知管理列表不传参 → 保持无时间过滤行为）。可能性：低；影响：低；缓解：参数 optional、缺省不改变现行为，由 Owner 统一走全局 proto 流程。
- **首页重排回归风险**：首页为 TabBar 首屏，涉及跑马灯/通知/入口/邻里互助/寻失/广告多区块重排。可能性：中；影响：中；缓解：各区块空态/异常态独立保底（通知空态「暂无通知公告」、联络空态、寻失原空态、广告位纯静态），首页垂直全序由 REQ-HL-4 固定，QA 门禁 + 阶段评审覆盖。
- **30 天窗口过滤性能风险（REVISION #5）**：新增 `published_at` 窗口谓词落在 content_posts 上，既有 idx_published 因 community_id 弃用 NULL 无法服务 scope JOIN 后的过滤/排序。可能性：中；影响：中；缓解：设计阶段补充 content_posts 索引（status, published_at 等），以 EXPLAIN 验证窗口查询不走全表扫描，REQ-NTW-6 作为性能验收点。
- **community_contacts 缺表依赖**：ListContacts 当前因缺表报错；migration 004 幂等补表后可用；运行库表存在但结构漂移时 IF NOT EXISTS 不自动修复，需人工订正（REQ-CLP-2 边界）。可能性：低；影响：低；缓解：CREATE TABLE IF NOT EXISTS 幂等 + DDL 与 001/模型对齐 + 不预置种子（空态可接受）。

## 不做清单（MoSCoW 的 Won't have — 本轮明确不实现）

- 不做邻里互助后端数据源/模型/表/接口（D9）——「展示 7 天内需求、首页 3 条」行为依赖后端，本期无法实现，仅前端占位
- 不做邻里互助列表页/详情页/用户发布入口（D8）
- 不做物业报修/二手闲置/租房卖房对应功能落地页（D3，仅入口占位，点击提示「功能开发中」）
- 不做广告位内容动态化/运营配置/点击跳转（D6/D7，内容仍前端硬编码、点击预留不跳转）
- 不预置 community_contacts 种子数据（D4，空态由运营后续维护）
- 不改动寻失互助首页展示风格（保持现状）
- **不改动后端 GetMarqueeNotices**（REVISION #6：移动端跑马灯由首页通知列表派生，非该 RPC）
- 不做前端直连 file-service REST `GET /api/files/:id` 的附件读（REVISION #4：该端点强制所有权，查看者非上传者；改为消费详情响应重生 file_url）
- 不做 PC 端 / 管理后台联动改造
- 不做通知发布侧改造（发布流程不属于本变更）

## 验收标准

- 首页通知仅显示 30 天内 ≤3 条、发布时间倒序（窗口内置顶优先）；`published_at` 为 NULL 或未来预排期的行不进首页；跑马灯标题取自同一首页通知列表数据
- 通知列表页与首页同风格（REQ-NTW-5：role 色条/标签/标题/时间）、置顶优先 + published_at 倒序、30 天窗口，单请求 page_size 截断（`total` 反映窗口内全量），点卡片进详情；移除客户端时间过滤
- 通知详情完整展示标题/role/发布时间/内容/附件；图片附件点开全屏预览，文档附件经详情响应重生 file_url 下载打开（前端不直连 file-service REST，附件类型扩展 file_id/file_type；图片/文档分发按 file_type 扩展名白名单 `{png, jpg, jpeg, gif}` 判定，REVISION r2-1/r2-3/r2-4）
- 首页按 REQ-HL-4 固定垂直全序渲染：通知（跑马灯+卡片）→ 4 功能入口 → 邻里互助占位 → 寻失互助 → 底部广告位
- 首页通知下方渲染 4 个功能图标入口：便民联络跳联络列表页（拨号网格），其余 3 个点击提示「功能开发中」
- 联络列表页经 ListContacts API 渲染拨号网格；community_contacts 表 migration 004 幂等建表后不再报缺表（DDL 与 001/模型对齐，不预置种子）
- 首页新增邻里互助占位区块；寻失互助展示风格不变
- 3 个广告位集中到页面底部垂直堆叠，点击不跳转
- **性能验收（REVISION #5）**：`since_days=30` 窗口过滤 + published_at 倒序查询经 EXPLAIN 验证不走全表扫描（content_posts 索引已补充）
- **REST 透传验收（REVISION r2-2）**：移动端经 `GET /api/community/notices?since_days=30` 调用，`ListContentPostsReq.since_days` 成功透传至 RPC 并生效（30 天窗口在移动端路径不静默失效），PC 不传参时行为不变
- 提交前 `bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service` PASS；api-proto 变更经全局 proto 流程（lint + breaking-check + generate）
