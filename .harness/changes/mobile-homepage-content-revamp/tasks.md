# Tasks: 移动端「社区家园」首页信息架构改造

> **对执行 Agent 的指令**: 每个 Task 独立可测，按 TDD 执行（先写测试→看失败→写实现→看通过）。精确到文件路径。**全局 / Proto 组由全局 Claude 执行**（本管线 Owner 处理），不分发子 Claude；**Owner 运维验证**组不走 Go 管线，由 Owner 在编码后单独执行。
>
> 依赖顺序：**全局/Proto（0.x）→ community-hub-service（1.x 模型→RPC→REST）→ web/mobile（2.x）→ Owner 运维验证（3.x）**。REST 透传（1.5）依赖 Proto since_days 字段存在（0.1/0.3 生成代码后 RPC 消费方才能引用新字段）。

## 全局 / Proto（由全局 Claude 执行）

### Task 0.1: 定义 ListContentPostsRequest.since_days 字段
- **修改**: `api-proto/api/community/v1/community.proto`
- [ ] `ListContentPostsRequest` 新增 `int32 since_days = 6;`（天，有效值 1..365，缺省 0=不过滤；字段号 6 为下一可用号，现 1-5 已占用）
- [ ] 字段注释：`// 可选时间窗口（天）：published_at >= now-since_days AND published_at <= now；缺省 0=不过滤（additive，PC 管理列表不传参行为不变）；非法值 ≤0 或 >365 → 080005`
- [ ] int32 非 ID 字段，**不加** jstype（`// SEE: [[proto-jstype]]` 仅约束 int64 ID）
- [ ] 同步 `ListContentPosts` RPC 方法注释（可选时间窗口说明）
- **TDD**: Proto 无逻辑代码，不写测试；由 Task 0.3 breaking-check 验证 additive

### Task 0.2: CHANGELOG 登记
- **修改**: `api-proto/CHANGELOG.md`
- [ ] 顶部新增 `## <date> — mobile-homepage-content-revamp：ListContentPosts 可选时间窗口` 条目
- [ ] 登记：`ListContentPostsRequest` 新增 `since_days`(6)（int32，天，1..365，缺省 0=不过滤）——**兼容新增，非破坏性**
- [ ] 登记影响：community-hub-service（RPC 窗口过滤 + REST 透传）、web/mobile（getNoticeList 传 since_days=30）；file-service/moderation-service 等不消费方无影响
- **TDD**: 无

### Task 0.3: 代码生成 + lint + breaking-check
- **运行**: `cd api-proto && make ci`
- [ ] `make generate` → 生成代码成功（gen/go/community/v1 更新）
- [ ] `make lint` → 0 errors
- [ ] `make breaking-check` → **无破坏性变更**（since_days 为新增可选字段，既有调用方兼容）
- [ ] `go mod tidy` 校验生成代码同步
- **TDD**: 生成代码编译即验证（`go build ./...` 于 community-hub-service 确认 RPC 侧引用新字段可编译）

## community-hub-service

### Task 1.1: Migration 004 — community_contacts 幂等补表
- **创建**: `services/community-hub-service/migration/004_add_community_contacts.sql`
- [ ] `CREATE TABLE IF NOT EXISTS community_contacts`（DDL 与 `migration/001_initial.sql` / `model/community_contact.go` 完全对齐）：`id BIGINT PK`（Snowflake）、`community_id BIGINT NOT NULL`、`category VARCHAR(30)`、`name VARCHAR(100)`、`phone VARCHAR(20)`、`sort_order INT DEFAULT 0`、`created_at`/`updated_at` DATETIME、`INDEX idx_community (community_id)`、InnoDB/utf8mb4
- [ ] **不预置种子数据**（REQ-CLP-2 场景 3，空态，D4）
- [ ] 文件头注释声明：001 为 schema 单源，004 为运行库缺表的幂等补救；表存在但结构漂移时 IF NOT EXISTS 不自动修复需人工订正（REQ-CLP-2 场景 5）
- **TDD**: 纯 DDL 无逻辑代码，不写测试（`// SEE: [[migration-must-execute]]` — 执行 + DESCRIBE 验证归 Task 3.1 Owner 运维验证）

### Task 1.2: Migration 005 — content_posts 窗口过滤索引
- **创建**: `services/community-hub-service/migration/005_content_posts_window_index.sql`
- [ ] 幂等守卫（MySQL 8.0 无 `ADD INDEX IF NOT EXISTS`）：`information_schema.statistics` 检查 `content_posts` 上 `idx_status_pinned_published` 是否存在，不存在才 `ALTER TABLE content_posts ADD INDEX idx_status_pinned_published (status, is_pinned, published_at)`
- [ ] 索引列序 `(status, is_pinned, published_at)`：等值 status + 等值 is_pinned + 范围 published_at，覆盖 `ORDER BY is_pinned DESC, published_at DESC`（REQ-NTW-6 / D13 / ADR-3）
- [ ] 文件头注释声明：additive 纯增量索引，不影响缺省（无 since_days）非窗口调用
- **TDD**: 纯 DDL 无逻辑代码，不写测试（执行 + EXPLAIN 验证归 Task 3.2 Owner 运维验证）

### Task 1.3: Model 层 FindListByCommunity 可选时间窗口谓词
- **修改**: `services/community-hub-service/model/content_post.go`
- **修改**: `services/community-hub-service/model/content_post_test.go`
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/notice_helpers_test.go`（fakeContentPostModel 签名随接口变参同步）
- [ ] 新增 `ContentPostListOption` 选项类型 + `WithTimeWindow(since time.Time)` 构造器（内部 `*contentPostListParams.since *time.Time`，nil=不过滤，保持 additive）
- [ ] `FindListByCommunity` 签名改为变参：`(ctx, communityId int64, sectionCode, role string, offset, limit int64, opts ...ContentPostListOption) ([]*ContentPost, int64, error)` — 变参使既有调用方/测试零改动即可编译
- [ ] impl 在 `opts` 提供 since 时，count 与 list 两条 SQL 追加 `and content_posts.published_at >= ? and content_posts.published_at <= ?`（上界 `time.Now()`，params：since、now）；`published_at` NULL 恒不匹配下界、未来行被上界排除（D12，参数化查询防注入）
- [ ] **RED**: `model/content_post_test.go` 新增 `TestContentPostModel_FindListByCommunity_WithWindow`（`sqlmock.AnyArg()` 匹配 since/now 两个 time 参数；场景：窗口内行返回、NULL/未来行被排除、无窗口参数时 SQL 不含窗口谓词）。既有 `TestContentPostModel_FindListByCommunity*` 因变参省略 `opts` 而无需改动，仍应通过（缺省无窗口谓词）
- [ ] **确认 RED**: `cd services/community-hub-service && go test ./model/ -run TestContentPostModel_FindListByCommunity -count=1` → 看到 FAIL（新签名/谓词未实现）
- [ ] **GREEN**: 实现变参 + 窗口谓词 → `go test ./model/ -run TestContentPostModel_FindListByCommunity -count=1` → PASS
- [ ] **REFACTOR**: fakeContentPostModel 签名同步变参（`opts ...model.ContentPostListOption`）；清理重复，保持测试绿

### Task 1.4: RPC ListContentPosts — since_days 校验 + 窗口传参
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/listcontentpostslogic.go`
- **创建**: `services/community-hub-service/rpc/internal/logic/notice/listcontentpostslogic_test.go`
- [ ] `since_days < 0 || since_days > 365` → 返回 `&communityv1.ListContentPostsResponse{Base: responsex.NewBaseRespWithError(scope.CodeInvalidParam, "since_days 超出有效范围 1..365")}`（REQ-NTW-2 场景 3；r2-5：int32 wire 恒数字，非数字由 REST 网关解析层拒绝，服务端仅校验数值范围）
- [ ] `since_days == 0`（缺省）→ 不传窗口选项，行为不变（REQ-NTW-2 场景 2，PC 管理列表兼容）
- [ ] `since_days > 0` → `model.WithTimeWindow(time.Now().AddDate(0, 0, -int(since_days)))` 传入 `FindListByCommunity`（窗口谓词下界，D10）
- [ ] **RED**: `listcontentpostslogic_test.go` 写 table-driven tests（≤5 用例）：`valid since_days=30 → fake 捕获窗口选项、since≈now-30d` / `since_days=0 → 无窗口选项` / `since_days=-1 → Base code 080005` / `since_days=366 → Base code 080005` / `since_days=365 → 边界合法`；fake 增加 `listOpts` 捕获字段（`// SEE: [[error-code-collision-and-namespace-alignment]]` — 复用 080005 不新增码）
- [ ] **确认 RED**: `go test ./rpc/internal/logic/notice/ -run TestListContentPosts_SinceDays -count=1` → 看到 FAIL
- [ ] **GREEN**: 实现校验 + 窗口传参 → `go test ./rpc/internal/logic/notice/ -run TestListContentPosts -count=1` → PASS
- [ ] **REFACTOR**: 清理命名，保持测试绿；回归 `go test ./rpc/... ./model/...`

### Task 1.5: REST 层 since_days 透传 + Base 错误上抛（REVISION r2-2）
- **修改**: `services/community-hub-service/api/internal/types/types.go`
- **修改**: `services/community-hub-service/api/internal/logic/notice/listcontentpostslogic.go`
- **修改**: `services/community-hub-service/api/internal/logic/notice/api_proxy_test.go`
- [ ] `api/internal/types/types.go` `ListContentPostsReq` 新增 `SinceDays int32 form:"since_days,optional"`（缺省 0 → RPC 缺省不过滤）
- [ ] `api/internal/logic/notice/listcontentpostslogic.go` RPC 请求补 `SinceDays: req.SinceDays`（**必须贯通**，仅列 RPC logic 会致 REST 层丢弃参数、移动端 30 天窗口静默失效）
- [ ] 该 logic 补 `responsex.ToError(resp.GetBase())` 检查（RPC 侧 080005 以 Base 错误返回而非 gRPC err，REST 层须上抛，禁止静默吞错；与 `getcontentpostlogic.go` 等既有模式一致）
- [ ] **RED**: `api_proxy_test.go` fakeContentPostRpc 补 `ListContentPosts` override（捕获 `listReq` + 返回 `listResp/listErr`）；新增用例：`req.SinceDays=30 → RPC 收到 SinceDays=30` / `RPC Base code=080005 → api 返回 error` / `RPC nil Base → 正常映射 res.notices + total`（`// SEE: [[verify-api-before-calling]]` — 路由 `GET /api/community/notices` 已在 graph-context 确认）
- [ ] **确认 RED**: `go test ./api/internal/logic/notice/ -run TestListContentPosts -count=1` → 看到 FAIL
- [ ] **GREEN**: 实现透传 + Base 检查 → `go test ./api/internal/logic/notice/ -run TestListContentPosts -count=1` → PASS
- [ ] **REFACTOR**: 清理，保持测试绿；`go build ./... && go test ./...` 全绿

## web/mobile

### Task 2.1: api/community.ts — 类型扩展 + since_days 参数 + 图片白名单
- **修改**: `web/mobile/src/api/community.ts`
- [ ] `NoticeAttachment` 扩展 `file_id: string` + `file_type: string`（snake_case 对齐 wire；`// SEE: [[snake-camel-field-mismatch]]`；缺失时字段为 undefined 不崩溃 — REQ-NDP-4 场景 2）
- [ ] `getNoticeList` 新增 `sinceDays?: number = 0` 参数 → `params.since_days`（缺省不传，PC/其他调用方行为不变）
- [ ] 新增图片白名单常量 `IMAGE_FILE_TYPES: string[] = ['png','jpg','jpeg','gif']` + `isImageAttachment(fileType?: string): boolean` 辅助函数（注释声明：`file_type` 为 file-service magic-bytes 嗅探落库的规范小写扩展名、非 MIME；白名单须与 `services/file-service` `guard/magic.go` 对齐 — `// SEE: [[frontend-business-rule-hardcode]]`）
- [ ] **测试（可选）**: 若新建 `web/mobile/src/api/community.spec.ts`，覆盖 `isImageAttachment('png')===true` / `isImageAttachment('pdf')===false` / `isImageAttachment(undefined)===false`；否则在 Task 2.2 组件测试中覆盖
- **TDD**: 纯类型/常量扩展，核心可测点为 `isImageAttachment`（分支逻辑，建议 vitest 单测）

### Task 2.2: notice.vue — 通知窗口参数 + 4 功能入口（移除内嵌联络网格）
- **修改**: `web/mobile/src/pages/notice/notice.vue`
- **修改**: `web/mobile/src/pages/notice/notice.spec.ts`
- [ ] 通知区改调 `getNoticeList(cid, 1, 3, 30)`（`since_days=30` + `page_size=3`，倒序、窗口内置顶优先；窗口业务逻辑后端强制 — REQ-NTW-1/2）
- [ ] 移除首页内嵌联络拨号网格：删除 `fetchContacts`/`contacts`/`contactGroups`/`getContacts` 调用与模板网格（REQ-FE-2 场景 2）
- [ ] 通知区块下方渲染 4 个功能图标入口（固定顺序）：便民联络 → `uni.navigateTo('/pages/contact-list/contact-list')`（做实跳页）；物业报修 / 二手闲置 / 租房卖房 → `uni.showToast('功能开发中')` 不跳转（REQ-FE-1/2/3；占位重复点击仍仅提示）
- [ ] 未加入小区（`hasCommunities==false`）时不渲染入口区（REQ-FE-1 场景 2，沿用既有 no-community hint）
- [ ] **RED**: `notice.spec.ts` 更新 mock（移除 `getContacts`，保留 `getNoticeList`/`getLostFoundList`）；新增用例：`4 个入口按固定顺序渲染` / `点击便民联络 → uni.navigateTo 到 contact-list` / `点击物业报修/二手闲置/租房卖房 → showToast('功能开发中') 且不 navigateTo` / `getNoticeList 以 since_days=30&page_size=3 调用`；移除依赖联络网格的旧断言
- [ ] **确认 RED**: `cd web/mobile && npm run test:unit -- src/pages/notice/notice.spec.ts` → 看到 FAIL
- [ ] **GREEN**: 实现模板 + 逻辑 → 同命令 PASS
- [ ] **REFACTOR**: 清理，保持测试绿

### Task 2.3: notice.vue — 区块垂直全序重排（邻里互助占位 + 广告集中）
- **修改**: `web/mobile/src/pages/notice/notice.vue`
- [ ] 按 REQ-HL-4 固定全序渲染：通知（跑马灯+卡片）→ 4 功能入口（Task 2.2）→ 邻里互助占位 → 寻失互助 → 底部广告位
- [ ] 新增「邻里互助」占位区块：占位文案（如「互助功能开发中」/空态）、点击不导航、不伪造需求数据（REQ-HL-1，D8/D9）
- [ ] 移除 3 个广告位的分散位置（原 2 个在联络下方、1 个在寻失下方），集中到页面底部垂直堆叠、保留原硬编码内容（REQ-HL-3 场景 1/2）
- [ ] 广告点击保持预留不跳转（`onAdClick` 空实现，REQ-HL-3 场景 3 / D7）；寻失互助区块样式与数据不动（REQ-HL-2）
- [ ] 各区块空态保底：通知空态「暂无通知公告」、寻失空态保持现状，全序不被空态破坏（REQ-HL-4 场景 2）
- [ ] **测试（可选）**: notice.spec.ts 补「页面含邻里互助占位、广告位渲染在寻失之后」断言；如不可行，依赖 Task 3.2 Owner 端到端验收
- **TDD**: 布局调整，组件测试可选；关键验收点（区块全序 + 空态）在 Task 3.2 端到端验收兜底

### Task 2.4: notice-browse.vue — 从单条翻页浏览改为 30 天卡片列表
- **修改**: `web/mobile/src/pages/notice-browse/notice-browse.vue`
- [ ] 改为与首页一致的卡片列表渲染（REQ-NTW-5 视觉契约：role 色条 + role 标签 + 标题 + 时间，复用首页卡片样式逻辑）
- [ ] 数据：`getNoticeList(cid, 1, 50, 30)`（`since_days=30` + 固定 `page_size=50` 单请求，置顶优先 + published_at 倒序；`total` 反映窗口内全量 — REQ-NTW-4/REQ-NTW-2）
- [ ] **移除客户端 3 个月过滤**：删除 `threeMonthsAgo` 计算与 `filter(...)`（窗口由后端强制，前端不实现窗口业务逻辑 — REQ-NTW-2 / `// SEE: [[frontend-business-rule-hardcode]]`）
- [ ] 点卡片 → `uni.navigateTo('/pages/notice-detail/notice-detail?id=...')`（REQ-NTW-4 场景 4）
- [ ] 空态「暂无通知公告」+ 加载失败明确提示（不静默 — REQ-NTW-4 场景 5/6，`// SEE: [[verify-api-before-calling]]`）
- [ ] 移除翻页按钮/currentIndex 逻辑；`formatTime` 保持与首页一致格式
- **TDD**: 页面重写，组件测试可选；关键验收（卡片契约 + 窗口数据）在 Task 3.2 端到端验收兜底

### Task 2.5: notice-detail.vue — 附件预览改造（file_type 白名单分发）
- **修改**: `web/mobile/src/pages/notice-detail/notice-detail.vue`
- [ ] 附件点击按 `isImageAttachment(att.file_type)` 分发：图片 → `uni.previewImage({ urls: [att.file_url] })` 全屏预览；非图片（pdf/doc/docx 或缺失/无法识别）→ `uni.downloadFile({ url: att.file_url })` 成功 `uni.openDocument({ filePath })`（REQ-NDP-2/3，`// SEE: [[frontend-business-rule-hardcode]]` 白名单来源）
- [ ] **移除前端直连 file-service REST**：`file_url` 直接用详情响应已重生值，不调用 `GET /api/files/:id`（REVISION #4，该端点强制所有权，查看者非上传者被拒）
- [ ] 图片 `file_url` 空或预览失败 → `uni.showToast('预览失败')`，**不**降级到文档打开器（REQ-NDP-2 场景 2）
- [ ] 文档 `file_url` 空或下载/打开失败 → `uni.showToast('附件打开失败')`（REQ-NDP-3 场景 2）
- [ ] 详情加载失败（API 失败/不存在/详情读整单失败 r2-6）→ 明确失败态文案「加载失败」/「通知不存在」，不静默空白页（REQ-NDP-1 场景 3）
- [ ] 无附件（`attachment_count==0` 或空数组）不渲染附件区（REQ-NDP-1 场景 2）
- **TDD**: 页面交互改造，组件测试可选；分发分支（图片/文档/空 file_url）与失败态在 Task 3.2 端到端验收兜底

### Task 2.6: contact-list.vue 新建 + pages.json 注册
- **创建**: `web/mobile/src/pages/contact-list/contact-list.vue`
- **修改**: `web/mobile/src/pages.json`
- [ ] `pages.json` `pages` 数组新增 `{ "path": "pages/contact-list/contact-list", "style": { "navigationBarTitleText": "便民联络" } }`（Uni-app 页面必须注册，否则编译器不识别）
- [ ] 新建 `contact-list.vue`：`onLoad` 读当前小区 → `getContacts(cid)` 渲染拨号网格（类别图标 `getContactCategoryIcon` + 类别名 `getContactCategoryName` + 电话），点击 → `uni.makePhoneCall({ phoneNumber })`（REQ-CLP-1，样式沿用首页原联络网格）
- [ ] 空数据 → 空态「暂无联络信息」；API 失败 → 明确加载失败提示（REQ-CLP-1 场景 3/4，`// SEE: [[verify-api-before-calling]]`，路由 `GET /api/community/contacts` 已在 graph-context 确认）
- **TDD**: 新页面，组件测试可选；空态/失败态 + 拨号在 Task 3.2 端到端验收兜底

## Owner 运维验证（编码后 Owner 执行，不走 Go 管线派发）

> 规则 12：需要真实 DB 实例的验证（DDL 执行、EXPLAIN、端到端验收）不属于编码任务，编码完成后由 Owner 单独执行。

### Task 3.1: migration 004/005 执行 + 结构验证
- **运行**: 真实 MySQL 实例（docker compose mysql）
- [ ] 执行 `migration/004_add_community_contacts.sql` → `DESCRIBE community_contacts` 确认 8 列 + idx_community 存在（`// SEE: [[migration-must-execute]]`）
- [ ] 重跑 004 → 无报错且数据不变（幂等，REQ-CLP-2 场景 2）
- [ ] 执行 `migration/005_content_posts_window_index.sql` → `SHOW INDEX FROM content_posts WHERE Key_name='idx_status_pinned_published'` 确认存在
- [ ] 重跑 005 → 无报错（幂等守卫生效）
- [ ] `bash scripts/init-databases.sh --check-only` 确认库齐备

### Task 3.2: 端到端验收（EXPLAIN + since_days 透传 + 门禁）
- **运行**: 全栈启动 + 真实 DB
- [ ] **EXPLAIN 验收（REVISION #5 / REQ-NTW-6）**：对窗口查询 `EXPLAIN SELECT ... FROM content_posts JOIN content_post_scope ... WHERE scope.community_id=? AND status=2 AND published_at >= ? AND published_at <= ? ORDER BY is_pinned DESC, published_at DESC` → 确认走 `idx_status_pinned_published`/`idx_scope_community`，`content_posts` 无全表扫描
- [ ] **REST 透传验收（REVISION r2-2）**：`curl "http://localhost:8887/api/community/notices?community_id=<cid>&page=1&page_size=3&since_days=30"` → 仅返回 30 天内置顶优先倒序 ≤3 条；不传 `since_days` → 行为与改造前一致（PC 兼容）；`since_days=400` → 080005 参数无效
- [ ] 首页 UI 验收：区块全序 = 通知（跑马灯+卡片）→ 4 功能入口 → 邻里互助占位 → 寻失互助 → 底部广告位；通知空态「暂无通知公告」；4 入口点击行为正确（REQ-HL-4 / REQ-FE-1/2/3）
- [ ] 列表页验收：30 天卡片列表 + 置顶优先 + 倒序 + 点卡片进详情（REQ-NTW-4/5）
- [ ] 详情页验收：附件图片全屏预览 / 文档打开 / 失败提示（REQ-NDP-2/3）
- [ ] 联络列表页验收：拨号网格渲染 + 空态 + 点卡片拨号（REQ-CLP-1）
- [ ] 门禁：`bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service` → PASS；`cd api-proto && make ci` → PASS（`// SEE: [[pre-commit-checks]]`）
