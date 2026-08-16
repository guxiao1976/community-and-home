# Design: 移动端「社区家园」首页信息架构改造

> 变更：`mobile-homepage-content-revamp`（P1 / L 级 / modify）
> 输入：`proposal.md` + `specs/{notice-time-window,notice-detail-preview,function-entries,contact-list-page,homepage-layout}/spec.md`
> 本设计覆盖全部 5 个能力 spec（19 个 REQ 点），追溯矩阵见 §需求追溯矩阵。

## 需求追溯矩阵（spec → design，防需求遗漏/设计蔓延）

| 需求 ID | 需求内容摘要 | 对应设计章节 | 覆盖状态 |
|---------|-------------|-------------|:---:|
| REQ-NTW-1 | 首页通知 ≤3 条 / 30 天窗口 / published_at 倒序（窗口内置顶优先） | §接口设计 ListContentPosts.since_days · §前端 notice.vue | ✅ |
| REQ-NTW-2 | 30 天窗口后端强制、缺省不过滤、非法参数 080005 | §接口设计 · §Proto 变更 · §业务流程 | ✅ |
| REQ-NTW-3 | 跑马灯由首页通知列表派生（同源 30 天） | §前端 notice.vue | ✅ |
| REQ-NTW-4 | 列表页仅 30 天内、置顶优先 + 倒序、单请求 page_size 截断 | §前端 notice-browse.vue | ✅ |
| REQ-NTW-5 | 通知卡片视觉契约（role 色条/标签/标题/时间） | §前端 notice-browse.vue + notice.vue | ✅ |
| REQ-NTW-6 | 30 天窗口过滤 + 倒序不走全表扫描（索引 + EXPLAIN 验收） | §数据模型 005 索引 · §非功能设计 | ✅ |
| REQ-NDP-1 | 详情完整展示标题/role/时间/内容/附件；无附件不渲染附件区；失败明确提示 | §前端 notice-detail.vue | ✅ |
| REQ-NDP-2 | 图片附件 `file_type ∈ {png,jpg,jpeg,gif}` 全屏 previewImage | §前端 notice-detail.vue · §接口设计 | ✅ |
| REQ-NDP-3 | 文档附件经响应重生 file_url 下载 + openDocument，不直连 file-service REST | §前端 notice-detail.vue | ✅ |
| REQ-NDP-4 | 前端 `NoticeAttachment` 类型扩展 `file_id`/`file_type` | §前端 api/community.ts | ✅ |
| REQ-FE-1 | 通知下方固定顺序 4 个功能图标入口 | §前端 notice.vue | ✅ |
| REQ-FE-2 | 便民联络入口做实跳联络列表页（移除首页内嵌网格） | §前端 notice.vue + contact-list.vue | ✅ |
| REQ-FE-3 | 物业报修/二手闲置/租房卖房占位 toast「功能开发中」不跳转 | §前端 notice.vue | ✅ |
| REQ-CLP-1 | 联络列表页经 ListContacts 渲染拨号网格 + 空态/失败态 | §前端 contact-list.vue · §接口设计 ListContacts 复用 | ✅ |
| REQ-CLP-2 | migration 004 幂等补 community_contacts（不预置种子，DDL 与 001/模型对齐） | §数据模型 004 · §业务流程 | ✅ |
| REQ-HL-1 | 首页邻里互助占位区块（无后端/无页面，点击不跳转） | §前端 notice.vue | ✅ |
| REQ-HL-2 | 寻失互助区块展示风格保持不变 | §前端 notice.vue（不改） | ✅ |
| REQ-HL-3 | 3 个广告位移除原分散位置、页面底部垂直堆叠集中展示 | §前端 notice.vue | ✅ |
| REQ-HL-4 | 首页区块垂直全序固定：通知→4入口→邻里互助→寻失→底部广告 | §前端 notice.vue | ✅ |

> 无 spec 依据的设计内容 = 设计蔓延：本设计无蔓延项（所有后端/前端改动均能回溯到对应 REQ）。

## 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| 通知 30 天窗口过滤（`since_days` 读契约） | community-hub-service（读路径） + api-proto（契约） | content_posts 数据所有权在 community-hub；既有 ListContentPosts 读路径扩展，不新建服务 |
| `since_days` REST 层透传 | community-hub-service（api 层） | `ListContentPostsReq` form 字段 + api logic 透传 RPC（REVISION r2-2 强制链路贯通） |
| content_posts 窗口索引 + community_contacts 补表 | community-hub-service | 数据层迁移，DDL 单源在 migration/ |
| 首页信息架构重排 / 4 功能入口 / 邻里互助占位 / 广告集中 / 跑马灯同源 | web/mobile | 纯展示与交互；窗口等业务规则由后端强制，前端只传参/消费结果 |
| 联络列表页（拨号网格） | web/mobile（页面） + community-hub-service（ListContacts 数据） | 数据在 community_contacts（community-hub 拥有）；页面消费既有 ListContacts API，无新增接口 |
| 附件预览（图片/文档分发） | web/mobile + community-hub（GetContentPost 重生 file_url，复用） + file-service（GetFileUrl，复用） | 详情响应 file_url 已由 GetContentPost 服务端重生；前端不直连 file-service REST；无新增后端变更 |
| 邻里互助后端数据源 / 列表页 / 详情页 / 发布入口 | 不做（D8/D9） | 本期前端占位，无模型/表/接口/契约 |
| 后端 GetMarqueeNotices | 不做（REVISION #6） | 移动端跑马灯由首页通知列表派生，不消费该 RPC |

归属存疑项：无（本变更归属清晰，无灰色地带需评审裁决）。

## 数据模型

### 新增表：community_contacts（migration 004，幂等补救）

运行库缺表导致 ListContacts 报 `Table doesn't exist`。001 为 schema 单源（`migration/001_initial.sql` 已声明同 DDL），004 是对「001 已应用后运行库仍缺表」的幂等补救，DDL 与 001 / `model/community_contact.go` 完全对齐，不预置种子（空态，运营后续维护）。

```sql
-- migration/004_add_community_contacts.sql
USE community_hub_db;

CREATE TABLE IF NOT EXISTS community_contacts (
    id          BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
    community_id BIGINT NOT NULL COMMENT '小区ID',
    category    VARCHAR(30) NOT NULL COMMENT '类别：water/electricity/gas/unicom/mobile/telecom/police',
    name        VARCHAR(100) NOT NULL COMMENT '显示名称',
    phone       VARCHAR(20) NOT NULL COMMENT '电话号码',
    sort_order  INT DEFAULT 0 COMMENT '排序',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_community (community_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='便民联络（001 单源，004 幂等补救运行库缺失，不预置种子）';
```

**字段约束补充**：
- `id` BIGINT PK + Snowflake 生成（`common/pkg/snowflake`），Go 端 `json:",string"`（编码规范 §5 / memory `[[proto-jstype]]`）。
- 表引擎 InnoDB / utf8mb4 / utf8mb4_unicode_ci，与 001 一致。
- 硬删除语义（Upsert：先删后插），无 deleted_at（遵循 design.md §设计决策 4 既有约定）。
- 不预置种子数据（REQ-CLP-2 场景 3；D4 空态）。
- 边界：运行库表存在但结构漂移时 `CREATE TABLE IF NOT EXISTS` 不自动修复，需人工订正（REQ-CLP-2 场景 5，不静默掩盖）。
- `created_at`/`updated_at` 满足编码规范 §3.1 时间字段（community_contacts 无 deleted_at 为既有显式偏离，见 design.md §设计决策 4）。

### 索引设计：content_posts 窗口过滤索引（migration 005）

30 天窗口谓词落在 content_posts：`WHERE scope.community_id=? AND status=approved AND published_at >= ? AND published_at <= ? ... ORDER BY is_pinned DESC, published_at DESC`。既有 `idx_published(community_id, published_at DESC, deleted_at)` 因 community_id 为弃用 NULL 列无法服务 scope JOIN 后的过滤/排序（REVISION #5）。新增索引：

```sql
-- migration/005_content_posts_window_index.sql（幂等：仅当索引不存在时创建，MySQL 8.0 无 ADD INDEX IF NOT EXISTS）
SET @db = DATABASE();
SET @idx_exists = (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = @db AND table_name = 'content_posts' AND index_name = 'idx_status_pinned_published');
SET @sql = IF(@idx_exists = 0,
    'ALTER TABLE content_posts ADD INDEX idx_status_pinned_published (status, is_pinned, published_at)',
    'SELECT ''idx_status_pinned_published already exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
```

- `idx_status_pinned_published (status, is_pinned, published_at)`：等值 status=approved + 等值 is_pinned（低基数）+ 范围 published_at，同时覆盖 `ORDER BY is_pinned DESC, published_at DESC`（等值列在前、范围列在后，读取天然有序，减少 filesort）。相比 spec 示例 `(status, published_at)`，多带 is_pinned 以覆盖排序前导列（见 §ADR ADR-3）。
- 纯增量索引，不影响缺省（无 since_days）非窗口调用（REQ-NTW-6 场景 2）。
- 执行需真实 DB 实例 → 标注「Owner 运维验证」（编码规范规则 12：DDL 执行不走 Go 管线）。

### 表关系

- community_contacts 独立表，与既有表无外键（外键约束在业务层，遵循既有约定）。
- content_posts 索引为既有表增量，无结构变更。

## 接口设计

### community-hub-service.ListContentPosts（RPC，支持可选时间窗口）

- **输入**：`ListContentPostsRequest` 新增 `int32 since_days = 6`（天，1..365；缺省 0=不过滤）。既有字段不变：community_id / role / section_code / page / page_size。
- **输出**：`ListContentPostsResponse` 不变（posts / total）。
- **窗口谓词**（D10/D12）：`since_days > 0` 时 `published_at >= (now - since_days) AND published_at <= now`；`published_at` NULL 恒不匹配下界、未来行被上界排除（均不进窗口）。
- **鉴权**：沿用既有 scope 读过滤（`scope.FilterAllowed`，越权 → 空列表不泄露），无新增鉴权。
- **幂等**：只读接口，天然幂等。
- **参数校验**：`since_days < 0 || since_days > 365` → `080005 参数无效`（REQ-NTW-2 场景 3；r2-5：int32 wire 恒为数字，非数字由 REST 网关解析层拒绝，服务端仅校验数值范围）。`since_days == 0`（缺省）→ 不过滤，PC 管理列表行为不变（REQ-NTW-2 场景 2）。
- **性能约束**：窗口查询走 `idx_status_pinned_published`，EXPLAIN 验收不走 content_posts 全表扫描（REQ-NTW-6）。
- **错误码**：`080005`（since_days 超范围），复用既有 `CodeInvalidParam`（types.go `=80005`），不新增错误码（memory `[[error-code-collision-and-namespace-alignment]]`：一码一义，080005 语义「参数无效」未冲突）。

### community-hub-service ListContentPosts（REST 层透传，REVISION r2-2）

- **输入**：`GET /api/community/notices` query `since_days`（form，optional）。`api/internal/types.ListContentPostsReq` 新增 `SinceDays int32 form:"since_days,optional"`。
- **透传**：`api/internal/logic/notice/listcontentpostslogic.go` 将 `req.SinceDays` 填入 RPC `ListContentPostsRequest.SinceDays`。
- **Base 错误上抛**：该 api logic 目前仅检查 gRPC err，未检查 `resp.GetBase()`；RPC 侧 080005 以 Base 错误返回（非 gRPC err），REST 层须补 `responsex.ToError(resp.GetBase())` 检查使非法窗口参数在 REST 面同样返回 080005（不静默吞错，与 getcontentpostlogic.go 等既有模式一致）。

### GetContentPost（复用，无新增变更）

- 详情响应 `ContentPost.Attachments[].file_url` 已由服务端经 RPC `GetFileUrl` 重生为 file-service 预签名 URL（无所有权限制）；`file_id` 为权威重生载体、兼容期 0 回退 stored URL（REVISION #4，r2-6）。
- 任一附件重生失败 → 读整单失败 → REST 透错 → 详情页按 REQ-NDP-1 加载失败态处理（r2-6）。
- 前端**不直连** file-service REST `GET /api/files/:id`（该端点强制文件所有权，查看者非上传者被拒，REVISION #4）。

### ListContacts（复用，无新增变更）

- `GET /api/community/contacts?community_id=` 返回 contacts（category/name/phone/sort_order），联络列表页消费；migration 004 补表后不再报缺表。

## 业务流程

**正常路径（首页通知 + 跑马灯同源）**：
1. `notice.vue` onLoad → `getNoticeList(cid, page=1, page_size=3, since_days=30)` → REST `GET /api/community/notices?community_id&page=1&page_size=3&since_days=30`
2. REST api logic 透传 since_days → RPC `ListContentPosts`（since_days 校验 + 窗口谓词 + `idx_status_pinned_published` 索引路径）→ 返回窗口内置顶优先 + published_at 倒序 ≤3 条
3. 跑马灯标题集 = 同一返回集 titles（顺序一致，≤3 条；空 → 「暂无通知公告」，REQ-NTW-3）

**正常路径（列表页）**：`notice-browse.vue` → `getNoticeList(cid, 1, 50, 30)` 单请求；服务端返回窗口内最新 page_size 条（置顶优先 + 倒序），`total` 反映窗口内全量；渲染 REQ-NTW-5 卡片列表；点卡片 → notice-detail。

**正常路径（联络列表页）**：首页「便民联络」入口 → `navigateTo /pages/contact-list/contact-list` → `getContacts(cid)` → 渲染拨号网格（类别图标/名称/电话）→ 点击 `uni.makePhoneCall`。

**正常路径（附件预览）**：详情响应 attachment → `file_type ∈ {png,jpg,jpeg,gif}` → `uni.previewImage([file_url])`；其余 → `uni.downloadFile(file_url)` + `uni.openDocument`。

**异常/失败路径**：
- since_days 非法（<0 / >365）→ RPC 080005 → REST `ToError` 上抛 → 前端明确提示（不静默）。
- 列表/详情/联络 API 失败 → 前端明确错误提示（REQ-NTW-4/REQ-NDP-1/REQ-CLP-1，禁止静默吞错）。
- 详情附件重生整单失败（file-service 不可用/文件已删）→ GetContentPost 读整单失败 → 详情页按加载失败态处理（REQ-NDP-1，r2-6）；响应已返回但单附件 file_url 为空（legacy）→ 逐附件「附件打开失败」提示（REQ-NDP-2/3 降级分支）。
- 广告点击 → 预留不跳转（D7）；占位入口点击 → toast「功能开发中」不跳转（REQ-FE-3）；邻里互助占位点击 → 不导航（REQ-HL-1）。

**跨服务一致性**：本变更无跨服务数据写入（附件重生为读侧 RPC 调用；community_contacts 为单服务数据）。GetFileUrl 重生失败 → 读整单失败，无半态一致性风险。服务间仅经 gRPC（memory `[[grpc-only-comms]]`）。

## Proto 变更

| 文件 | 变更类型 | 破坏性(是/否) | 说明 |
|------|:---:|:---:|------|
| `api-proto/api/community/v1/community.proto` | 新增字段 | 否（additive） | `ListContentPostsRequest` 新增 `int32 since_days = 6`（天，1..365，缺省 0=不过滤）；字段号 6 为下一可用号（现 1-5 已占用）；int32 无需 jstype |
| `api-proto/CHANGELOG.md` | 登记 | 否 | 登记 since_days 新增（兼容新增，无破坏性） |

> **影响评估（additive 非破坏）**：既有调用方（PC 通知管理列表等）不传 since_days → 0 → 不过滤，行为不变。`make ci`（lint + breaking-check + generate）应 PASS breaking-check。proto 变更仅全局 Claude 执行（硬约束 #2），归「全局 / Proto」组，由本管线 Owner 走全局 proto 流程。
> **GetMarqueeNoticesRequest 不变**（REVISION #6）。
> int32 非 ID 字段无需 `jstype=JS_STRING`（memory `[[proto-jstype]]` 仅约束 int64 ID）。

## 安全考虑

- **窗口参数校验**：since_days 服务端强制范围校验（1..365），防止超范围窗口导致的越权/异常查询；缺省不过滤保持 PC 行为（REQ-NTW-2）。
- **附件权限边界**：前端不直连 file-service REST `GET /api/files/:id`（强制所有权，查看者非上传者被拒，REVISION #4）；改消费 community-hub 详情响应已重生的预签名 file_url。file_url 为临时预签名 URL，过期由服务端重生，无长期密钥泄露面。
- **文件分发谓词**：图片/文档分发按 `file_type` 扩展名白名单 `{png,jpg,jpeg,gif}` 判定（REVISION r2-1/r2-3/r2-4）；wire `file_type` 为 file-service 嗅探落库的规范小写扩展名（非 MIME，wire 无 `image/*`），白名单与 `file-service` `guard/magic.go` 对齐；缺失/无法识别一律按文档走下载 + openDocument，不做绕过 file-service 的原始页内跳转（REQ-NDP-3 场景 3）。
- **SQL 注入**：窗口谓词经参数化查询（`published_at >= ? AND published_at <= ?`），无字符串拼接。

## 记忆引用（设计阶段预防性注入，Step 1.5 产出）

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| `[[proto-jstype]]` | 数据模型 | community_contacts.id BIGINT + Snowflake + Go `json:",string"`；since_days 为 int32 无需 jstype |
| `[[migration-must-execute]]` | 数据模型 | migration 004/005 执行 + DESCRIBE 验证归入「Owner 运维验证」，编码后由 Owner 执行 |
| `[[verify-api-before-calling]]` | 接口设计 | 前端各 API 调用路由已在 graph-context 确认（/api/community/notices、/contacts、/notices/:id）；失败明确提示不静默吞错 |
| `[[grpc-only-comms]]` | 业务流程 | 附件重生经 GetFileUrl RPC；无跨服务直连数据库 |
| `[[restore-compensation-zero-time]]` | 数据模型 | published_at 用 sql.NullTime（模型既有）；窗口谓词对 NULL 由范围比较天然排除，无需零值写入 |
| `[[frontend-business-rule-hardcode]]` | 安全考虑 | file_type 图片白名单 {png,jpg,jpeg,gif} 为前端分发谓词，必须与 file-service 白名单对齐并注释来源，防止漂移；30 天窗口不落前端（后端强制） |
| `[[error-code-collision-and-namespace-alignment]]` | 接口设计 | since_days 非法复用 080005（CodeInvalidParam），不新增错误码 |
| `[[snake-camel-field-mismatch]]` | 前端 | NoticeAttachment 扩展字段 file_id/file_type 用 snake_case 对齐 wire |
| `[[api-response-single-wrap]]` | 接口设计 | 前端消费 res.notices / res.contacts / res.notice 单层包装，不改响应结构 |

**不适用记忆**（主动排除，供 reviewer 确认非遗漏）：
- `[[unique-index-migration-dup-precheck]]` — 004 为 CREATE TABLE IF NOT EXISTS（无对既有数据加唯一索引）；005 为非唯一普通索引，均无需查重预检。
- `[[phone-encryption]]` — community_contacts.phone 为联络展示电话（公开信息，非用户敏感手机号），无加密落库。
- `[[redis-cache-soft-delete]]` / `[[notfound-cache-sentinel-vs-transient-error]]` — 本变更不引入 Redis 缓存（design.md §缓存策略：当前无缓存）。

## 非功能设计（精简 checklist）

- [x] 可靠性：列表/详情/联络读失败明确提示；附件重生整单失败 → 详情加载失败态；占位/广告点击不产生导航。无写路径，无需补偿/对账。
- [x] 性能：`since_days` 窗口查询走新增 `idx_status_pinned_published`；EXPLAIN 验收（Owner 运维验证）确认不走全表扫描（REQ-NTW-6 / D13）；列表页单请求 page_size=50 截断（D11，本期无客户端分页）。
- [x] 可观测性：RPC 侧 080005 参数无效经 REST 透传，前端 toast 明确提示；API 失败 console.error 已遵循既有模式；无新增指标要求（标注：仅复用既有链路日志）。
- [ ] 无显式要求则标注：无

## 关键设计决策与权衡（ADR，轻量）

| 决策点 | 备选方案 | 最终选型 | 取舍理由 | 未采用方案原因 |
|-------|---------|---------|---------|--------------|
| ADR-1 窗口谓词实现位置 | RPC 层 / REST 层 / 前端 | RPC 层（模型方法） | 单一强制点服务所有调用方（REST + 未来 gRPC 消费者）；前端不实现业务逻辑（REQ-NTW-2） | REST 层过滤会使 gRPC 直连消费方失效；前端过滤违反业务规则约束且与 total/分页漂移 |
| ADR-2 since_days 校验位置 | 仅 RPC / REST+RPC 双份 | 仅 RPC（REST 透传 + ToError 上抛） | REST 是薄代理，RPC 为契约权威；避免两处校验规则漂移（r2-5：int32 wire 非数字由网关解析层拒绝） | REST 重复校验会与 RPC 规则漂移，且难以覆盖 gRPC 消费方 |
| ADR-3 窗口索引形状 | `(status, published_at)` / `(status, is_pinned, published_at)` | `(status, is_pinned, published_at)` | 等值 status + 等值 is_pinned + 范围 published_at，同时覆盖 `ORDER BY is_pinned DESC, published_at DESC` 前导列，减少 filesort | `(status, published_at)` 无法消除排序前导 is_pinned 的 filesort；EXPLAIN 验收一致为「不走全表扫描」 |
| ADR-4 migration 编号 | 004 合并 contacts+索引 / 004 与 005 分离 | 004=contacts 幂等建表，005=窗口索引 | 单一关注点、幂等风格统一；004 对齐 .change.yaml 声明的 `004_add_community_contacts.sql` | 合并会混合「建表补救」与「性能索引」两类变更，评审与回滚边界不清 |
| ADR-5 附件分发谓词 | `file_type` 扩展名白名单 / MIME `image/*` | 扩展名白名单 `{png,jpg,jpeg,gif}` | wire `file_type` 为 file-service 嗅探落库的规范小写扩展名，非 MIME；白名单与 file-service 对齐（REVISION r2-1/r2-3/r2-4） | `image/*` 前缀判断按字面实现会致全部图片附件走文档分支（wire 无该值） |
| ADR-6 附件 URL 主路径 | 消费详情响应重生 file_url / 前端以 file_id 直连 file-service REST | 消费详情响应重生 file_url | 详情响应 file_url 即权威新鲜签名 URL；file-service REST `GET /api/files/:id` 强制所有权，查看者非上传者被拒（REVISION #4） | file_id 直连 REST 必然被拒（查看者非上传者）；新帖落库 stored file_url 为占位空串，回退无 URL |

## 设计 Self-Review（Step 3.5）

- [x] 追溯矩阵 100% 覆盖 19 个 REQ，无设计蔓延。
- [x] 服务归属符合数据所有权（content_posts / community_contacts → community-hub；页面 → web/mobile）；无存疑项。
- [x] 数据模型符合编码规范：Snowflake BIGINT + `json:",string"`、created_at/updated_at（community_contacts 无 deleted_at 为既有硬删除显式偏离）、软删除遵循 service design.md。
- [x] 非功能（可靠性/性能/可观测性）完整。
- [x] 破坏性变更：Proto 为 additive 非破坏 + 影响评估；DB 为幂等补救/增量索引。
- [x] 记忆引用已注入 + slug 自验（文件均在 `.harness/knowledge/memory/`）；不适用项已记录排除理由。
