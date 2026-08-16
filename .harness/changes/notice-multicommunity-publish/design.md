# Design: 通知模块多小区发布 + 通栏跑马灯 + 附件安全

> 变更名: `notice-multicommunity-publish` · 规模 L · 优先级 P1 · 类型 modify
> 输入: proposal.md + 6 个 spec（notice-publish / notice-read / notice-moderation / publish-permission / attachment-security / notice-mobile）
> 本设计对应需求追溯矩阵（Step 8 交付 traceability 的唯一依据：spec REQ → 设计章节）。

---

## 需求追溯矩阵（spec → design，防需求遗漏/设计蔓延）

| 需求 ID | 需求内容摘要 | 对应设计章节 | 覆盖状态 |
|---------|-------------|-------------|:---:|
| REQ-NP-1 | notice_scope 单源 + community_id/published_at 去 NOT NULL + notice_attachments.file_type 三迁移 + 单事务原子 | §数据模型 / §CreateNotice | ✅ |
| REQ-NP-2 | notice_scope 双列 NOT NULL + uk_notice_community + community_id 读索引 | §数据模型 | ✅ |
| REQ-NP-3 | CreateNotice 多小区契约（community_ids/division_id 互斥、≤100、division 仅 admin、AssertPublishScope、080002/080003/080005/080006 消歧、JWT 为准） | §CreateNotice / §Proto | ✅ |
| REQ-NP-4 | 各角色发布范围 + RBAC→NoticeRole 映射 + design gate（division 授权落位） | §服务归属 / §Design Gate / §CreateNotice | ✅ |
| REQ-NP-5 | 撤回复用 DeleteNotice（仅发布者本人、全局生效、scope 物理删、附件保留 D28） | §DeleteNotice | ✅ |
| REQ-NP-6 | 附件绑定校验（GetFileUrl 扩展 FileInfo：confirmed+user_id+file_type；≤10/≤50MB） | §CreateNotice / §附件绑定 | ✅ |
| REQ-NP-7 | CreateNotice 后端不幂等（D25） | §CreateNotice（非功能） | ✅ |
| REQ-NR-1 | 列表按 scope 过滤 + 响应单 community_id（取请求小区）+ role 筛选保留 + 分页 + 倒序 | §ListNotices | ✅ |
| REQ-NR-2 | GetNotice community_id 必填（080005）+ scope 外/不存在/未过审→080001 + 附件含 file_type | §GetNotice | ✅ |
| REQ-NR-3 | GetMarqueeNotices 专用 RPC（≤10 条、置顶优先、15 天含端点、仅审核通过、080005） | §GetMarqueeNotices | ✅ |
| REQ-NP-MOD-1 | 通知整体审核通过才可见（各小区一致） | §可见性门禁 | ✅ |
| REQ-NP-MOD-2 | 复用 UpdateNoticeModerationStatus 回调（不新增审核 API） | §UpdateNoticeModerationStatus | ✅ |
| REQ-NP-MOD-3 | UpdateNotice 保持通知级语义（编辑重置审核态；不做 scope 编辑） | §UpdateNotice | ✅ |
| REQ-NP-MOD-4 | published_at 锚定审核通过时（D27/D30，创建写 NULL、pass 回调置 now、重过审更新） | §UpdateNoticeModerationStatus / §数据模型 | ✅ |
| REQ-PP-1 | GetPublishPermission 返回 can_publish + 可发布角色（level-2 判定，经 GetUserRoles） | §GetPublishPermission | ✅ |
| REQ-PP-2 | 前端不判权限，入口由 can_publish 驱动 | §web/mobile / §GetPublishPermission | ✅ |
| REQ-PP-3 | 写路径角色状态门槛 = 421 min_verf_level=2（功能层强制）+ AssertPublishScope 落库前 | §CreateNotice（鉴权）/ §权限种子 | ✅ |
| REQ-PP-4 | 种子对齐（grid_worker 授 421、421 置 min_verf_level=2、收 property_admin/owner/tenant 421） | §权限种子 | ✅ |
| REQ-AS-1 | 白名单 {png,jpg,jpeg,gif,pdf,doc,docx} + 禁止集 + zip/rar 拒绝 + 070004 登记 | §附件安全 | ✅ |
| REQ-AS-2 | 单文件 ≤10MB 硬上限 + 070005 登记 | §附件安全 | ✅ |
| REQ-AS-3 | 两层校验（GetUploadUrl 快速拒绝 + ConfirmUpload magic-bytes 回读，doc OLE2/CFB+WordDocument、docx ZIP+OOXML word/document.xml） | §附件安全 | ✅ |
| REQ-AS-4 | 全局基线 + 按 entity_type 可扩展（10MB 硬上限不可放宽、禁止集不可弱化） | §附件安全 | ✅ |
| REQ-AS-5 | notice_attachments.file_type 记录（载体 = 扩展 FileInfo 回读，非客户端） | §附件绑定 / §数据模型 | ✅ |
| REQ-AS-6 | 单通知总量 ≤10 个 且 ≤50MB（CreateNotice 绑定校验，080005） | §CreateNotice / §附件绑定 | ✅ |
| REQ-AS-7 | FileInfo 扩展 file_type(11)+confirmed(12)（非破坏，经 GetFileUrl/ListFiles 返回） | §Proto / §FileInfo 扩展 | ✅ |
| REQ-NM-1 | 首页通栏 NoticeMarquee（消费 GetMarqueeNotices，15 天/置顶/封顶 10/更多→浏览/点击→详情/空态） | §web/mobile | ✅ |
| REQ-NM-2 | 浏览页 NoticeList（published_at DESC + 分页） | §web/mobile | ✅ |
| REQ-NM-3 | 详情页 NoticeDetail（标题/正文/published_at/附件含 file_type，携带当前小区上下文，080001→空态） | §web/mobile | ✅ |
| REQ-NM-4 | 【我的】页发布入口（can_publish 驱动，property_admin 无入口 D26） | §web/mobile | ✅ |
| REQ-NM-5 | 发布表单 NoticePublisher（范围选择：grid 多选/community_admin 选 division/committee 固定；提交中禁用 D25） | §web/mobile | ✅ |
| REQ-NM-6 | 附件前端一致预校验（白名单/10MB/≤10个/≤50MB） | §web/mobile | ✅ |
| REQ-NM-7 | 组件化沉淀（NoticeMarquee/NoticePublisher/NoticeList/NoticeDetail + 共享校验器） | §web/mobile | ✅ |

> 设计蔓延声明：§GetMarqueeNotices 的 `FilterAllowed` 读范围过滤（GLOBAL/LIMITED/EMPTY 语义）与现有 ListNotices 读路径一致，属「与既有读过滤语义对齐」而非新增需求——标注供 reviewer 确认非遗漏。

---

## 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| notices 去 community_id/published_at NOT NULL + notice_scope 表 + notice_attachments.file_type 迁移 | community-hub-service | 数据归属（notices 域） |
| CreateNotice 多小区/division 展开/附件绑定/单事务落库 | community-hub-service | 数据归属 + 业务域 |
| ListNotices / GetNotice / GetMarqueeNotices（scope 过滤读） | community-hub-service | 读路径数据归属 |
| GetPublishPermission 判定（经 permission GetUserRoles） | community-hub-service（判定逻辑）+ permission-service（角色状态权威） | 入口显隐属业务面；角色状态数据在 permission-service |
| published_at 审核通过回调设置 | community-hub-service（UpdateNoticeModerationStatus 处理） | 回调落库 |
| DeleteNotice 撤回收窄（仅发布者本人 + scope 物理删 + 附件保留） | community-hub-service | 数据归属 |
| **AssertPublishScope 判据扩展 community_div（design gate，REV-17 已定案：需改判据）** | permission-service | 授权集解析权威（证据见 §Design Gate） |
| 权限种子（grid_worker 授 421 / 421 min_verf_level=2 / 收 property_admin·owner·tenant 421 / 新读接口 423/424/425/426 绑定 + 422 扩展全部移动端角色 + 写接口 427/428 绑定——读/写权限矩阵见 §权限种子，评审 interface-proto v4 MUST 1/2） | permission-service | 种子数据归属 |
| 附件白名单/10MB/magic-bytes 两层校验 + FileInfo 扩展 + 070004/070005 | file-service | 文件数据归属 |
| division→小区展开 / 祖先链解析 | master-data-service | 只读复用（GetResidentialAreasByDivision / ResolveScopeAncestors 已存在） |
| moderation 审核流 | moderation-service | 只读复用（本变更不改 moderation-service 代码） |
| 移动端页面/组件 | web/mobile | 前端（PC 不做，Q6） |

**归属存疑项（已裁决）**：
- **division grant → community 授权集解析**：`AssertPublishScope` 判据逻辑**必须变更**（详见 §Design Gate）。推荐方案：新增 `resolvePublishScope` 变体收集 `community` + `community_div` 双 scope_type 授权并集；`GetDataScopes` 读路径保持 `community` 单 scope_type 不变（division 选项前端经 masterdata 树取，REQ-NM-5）。

---

## 数据模型

### 迁移文件：community-hub-service `migration/003_multi_community_notice.sql`

```sql
-- 003_multi_community_notice.sql — 多小区发布 + published_at 审核锚定 + 附件 file_type
USE community_hub_db;

-- 1. notices.community_id 去 NOT NULL（D1：兼容期保留列，不再写入；范围关联单源 notice_scope）
ALTER TABLE notices MODIFY community_id BIGINT DEFAULT NULL COMMENT '弃用：范围关联走 notice_scope（兼容期保留列）';

-- 2. notices.published_at 去 NOT NULL（D30：创建写 NULL，审核通过回调置 now，D27）
ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL COMMENT '审核通过时设置（D27/D30），创建时 NULL（待审不可见）';

-- 3. notice_scope 关联表（REQ-NP-2：单源、双列 NOT NULL、读索引 community_id 左）
CREATE TABLE IF NOT EXISTS notice_scope (
    notice_id    BIGINT NOT NULL COMMENT '通知ID（notices.id）',
    community_id BIGINT NOT NULL COMMENT '目标小区（md_residential_area.id，代表小区或村）',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (notice_id, community_id),              -- 即 REQ-NP-2 的 uk_notice_community 唯一约束（同一小区只一条）
    KEY idx_scope_community (community_id, notice_id)   -- 读路径：ListNotices/跑马灯按 community_id 先过滤
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知-小区范围关联（多小区发布单源；纯关联表，仅 created_at，无 updated_at/deleted_at，物理删除语义，符合编码规范 §3.1）';

-- 4. notice_attachments.file_type 列（REQ-NP-1/REQ-AS-5：白名单校验通过的文件类型，D24 自 FileInfo 回读）
ALTER TABLE notice_attachments ADD COLUMN file_type VARCHAR(20) DEFAULT NULL COMMENT '白名单校验通过的文件类型（扩展名，D24 自 FileInfo 回读）';
-- 5. notice_attachments.file_id 列（设计评审 data-model v3 M1：附件预签名 URL 重生载体，S4 闭环）
--    CreateNotice 落库时把 attachment_ids（file_id）写入；GetNotice 按 file_id 重生预签名 URL（file_url 为短期 3600s~7d）
ALTER TABLE notice_attachments ADD COLUMN file_id BIGINT DEFAULT 0 COMMENT 'file-service 文件ID（重生预签名 URL 的权威载体，D24/S4）；兼容期存量行 0/NULL，读路径回退 stored file_url';

-- 6. notice_attachments.file_url 加宽（data-model v4 S4：消除预签名 URL 快照截断风险）
--    新行 file_url 写占位空串 ''（file_id 为权威重生载体，不落 URL 快照）；本行仅防存量/边界长 URL 触发 VARCHAR(500) ERROR 1406
ALTER TABLE notice_attachments MODIFY file_url VARCHAR(1024) NOT NULL COMMENT '存量回退用 stored URL（新行占位空串，file_id 为权威重生载体）';
```

**字段约束补充**：
- `notices.community_id`/`published_at`：去 NOT NULL 后 `DEFAULT NULL`；`idx_community`/`idx_published` **兼容期保留并标记 deprecated**（避免 mid-rollout 代码破坏；新读路径走 notice_scope 索引，后续清理迁移可再删）。`published_at=NULL` 的 pending 行**不参与 idx_published 排序/跑马灯窗口**——可见性门禁（REQ-NP-MOD-1）先过滤，仅审核通过行（恒有 published_at）进入排序，NULL 不扰动 DESC（REQ-NP-1 场景）。
- `notice_scope`：复合 PK = 唯一约束（无需 Snowflake 主键，纯关联表）；`idx_scope_community` 以 `community_id` 为最左列服务列表/跑马灯读路径（REQ-NP-2 场景「读路径按 community_id 走索引」）。
- `notice_attachments.file_type`：VARCHAR(20) 存扩展名小写（如 `pdf`/`doc`/`docx`），兼容期存量行允许 NULL（读路径不因缺省而坏，REQ-AS-5 场景「附件元数据缺失」）。
- `notice_attachments.file_id`：BIGINT，file-service 文件 ID（`attachment_ids` 传入的 file_id），**预签名 URL 重生的权威载体（设计评审 data-model v3 M1/S4 闭环）**；CreateNotice 落库时写，GetNotice 按它调 `GetFileUrl(file_id)` 重生 URL；兼容期存量行默认 0，GetNotice 遇 file_id=0/NULL 时回退返回 stored file_url（尽力而为，与 Q1 存量兼容期一致）。
- `notices.moderation_status`：**存审核回调原始值（1=machine_pass / 2=machine_fail / 3=human_pass / 4=human_fail，int32 注释见 community.proto:316）**。回调直接写库（现 updatemoderationstatuslogic 存 `in.ModerationStatus`），status=3 会真实落库。**因此读路径可见性谓词必须与回调 pass 集合一致——通过态 = `{1, 3}`，而非仅 1**（见 §可见性门禁，评审 M1 修订）。
- 表引擎/字符集：InnoDB/utf8mb4_unicode_ci（与现有表一致）；软删除策略：`notices` 保持 `deleted_at` 逻辑删除，`notice_scope` 为**物理删除**（撤回时删行，REQ-NP-5）。
- 数据增长预估：notice_scope 行数 ≈ 通知数 × 平均目标小区数（≤100/条），量级小，索引兜底无归档需求。

### 索引设计

| 索引 | 表 | 用途 |
|------|-----|------|
| `uk_notice_community (notice_id, community_id)` = 复合 PK | notice_scope | 唯一约束 + 按通知查范围 |
| `idx_scope_community (community_id, notice_id)` | notice_scope | 列表/跑马灯按小区读路径（左列 community_id） |
| `idx_community` / `idx_published`（保留，deprecated） | notices | 兼容期旧代码；新读路径不走 |
| `idx_notice (notice_id)`（既有） | notice_attachments | 详情附件查询 |

---

## 可见性门禁（REQ-NP-MOD-1，评审 M1 修订）

**单一通过态谓词 `model.IsModerationPassed(status)`**（`status ∈ {1=machine_pass, 3=human_pass}`）为**列表 / 详情 / 跑马灯读查询与审核通过回调共用**的唯一判定，杜绝「读路径过滤 1、回调置 3」的两处漂移（谓词命名统一为 `IsModerationPassed`，与 Task 1.4 导出的函数一致，评审 S1 收敛二义）：

- **读路径**：`FindListByCommunity` / `FindOnePublished` / `FindMarquee` 一律 `moderation_status IN (1,3)`（或 `WHERE IsModerationPassed(moderation_status)`），**不再恒过滤 `moderation_status=1`**。
- **回调 pass 判定**：`UpdateNoticeModerationStatus` 的 pass 分支用同一 `IsModerationPassed(status)`（1 / 3 均置 `published_at=NOW()`，D27/D30）。
- **谓词落点**：通过态谓词实现在 `model/notice.go`（常量 `ModerationStatusMachinePass=1` / `ModerationStatusHumanPass=3` + **导出函数 `func IsModerationPassed(status int64) bool`**，评审 S1——rpc 层跨包复用必须导出），供 rpc 层（Task 1.5/1.13）与读查询共用；DB 存回调原始值（1/2/3/4），谓词是查询/回调共用常量，不是表结构变更。
- **一致性后果**：一条 `human_pass(3)` 通知被回调置上 `published_at` 后，必须同时出现在列表/详情/跑马灯——这是本变更的契约；现 moderation-service 回调仅发 1/2（task_handler.go 只映射 machine_pass/machine_fail），status=3 属潜在路径，设计以谓词统一封死（评审 M1）。

---

## 接口设计

> 所有 `int64` ID 字段 `[jstype=JS_STRING]`（硬性约束 #3，SEE [[proto-jstype]]）；错误码 5 位 `XXYYY`。

### community-hub-service.CreateNotice（重写）

- **输入**：`title`、`content`、`community_ids`（repeated int64 JS_STRING，≥1）、`division_id`（int64 JS_STRING，可选，仅 community_admin）、`attachment_ids`（repeated string = file_id）。**`publisher`（展示名）取请求体展示字符串（非安全字段，仅展示；`notices.publisher` 为 NOT NULL 列，来源必须显式，评审 S1）**；`role`/`publisher_id` 请求体**不信任**（服务端从 JWT 派生）。
- **REST/JSON 绑定**（评审 MUST 1/2 修订）：API 层 `CreateNoticeReq.CommunityIds` 为 **`[]string`**（移动端发 string 形式 Snowflake ID，防 JS 精度丢失，硬性约束 #3；encoding/json 的 `,string` 选项**不支持 slice**，`[]int64` 解 `["1001",...]` 会解包失败），代理到 RPC 前逐个 `strconv.ParseInt` 转 int64；`AttachmentIds []string` **新增并透传**到 RPC `attachment_ids`（现 api create 逻辑丢弃该字段，评审 MUST 2）；`DivisionId` 标量 `int64 json:"division_id,string"`（标量支持 `,string`）。
- **校验顺序**（任一步失败整体拒绝，无部分写）：
  1. title/content 非空 → 否则 080005；**`publisher`（展示名）非空 → 否则 080005**（`notices.publisher` 为 NOT NULL 列，缺省将 500，评审 I3）。
  2. scope 载体：`community_ids` 与 `division_id` 恰一设置（双载/双空 → 080005）；不回退旧 `community_id`(1)。
  3. `community_ids`：去重；>100 → 080003；全去重后空 → 080005。
  4. `division_id`：非 community_admin → 080005；**`division_id<=0` → 080005（fail-closed，杜绝进入 masterdata 默认 FindAll 分支过度展开，评审 SHOULD 5）**；masterdata `GetResidentialAreasByDivision(community_div_id=division_id, status=1)` 展开；空展开 → 080005；展开数 >100 → 080003；仅含审核通过小区。
  5. 附件绑定（REQ-NP-6/REQ-AS-6）：逐 `attachment_id` 调 file `GetFileUrl(file_id)` → `FileInfo.confirmed==true` 且 `user_id==JWT`；数量 ≤10 且 `Σ file_size ≤ 50MB` → 否则 080005（附件引用无效/超限）。
  6. 数据权限：**单次批量** `AssertPublishScope(user, targets)`（一次携带全部 target ScopeRef，`AssertPublishScopeRequest` 本身 `repeated targets`；任一目标越权/不可解析（目标不存在）→ **060007→080006 一次映射整体拒绝**，D31，fail-closed；避免逐目标 N+1 RPC，评审 SHOULD 3）。
- **鉴权**：功能权限 `community:notice:create-api`(421) + `min_verf_level=2` 由 **REST PermMiddleware** 强制（写路径角色状态门槛，REQ-PP-3 方案 A）；gRPC 面经 API 网关代理，不重复校验。
- **幂等**：**不幂等**（D25，无幂等键；重复提交产生重复通知；防重由前端提交中禁用承担）。
- **落库**（单事务，all-or-nothing）：notices（`community_id=NULL`、`published_at=NULL`、**`publisher`=请求体展示字符串**、`role`=实际角色→NoticeRole、`publisher_id`=JWT）+ notice_scope 批量插入 + notice_attachments 批量插入（file_name/file_size/**file_type 自 FileInfo 回读**、**`file_id` = `attachment_ids` 传入的 file_id（S4 闭环，预签名 URL 重生的权威载体）**；**`file_url` 新行写占位空串 `''`——file_id 为权威重生载体，不再落短期预签名 URL 快照，消除 `file_url VARCHAR(500) NOT NULL` 对长 URL（含 X-Amz-Signature 参数）的截断风险（ERROR 1406 → 整事务回滚，data-model v4 S4）；file_url 仅保留给 file_id=0 存量行回退**，migration 003 同时把 file_url 扩为 `VARCHAR(1024)` 兼容存量）。随后异步入 moderation 队列（既有流程）。
- **role 派生（评审 S3 修订）**：JWT 仅含 user_id，**实际角色必须显式调 permission `GetUserRoles(user_id)` 解析**（禁止从请求体读取）→ RBAC→NoticeRole 映射（grid_worker→GRID_OFFICER、community_admin→COMMUNITY、committee→COMMITTEE，复用 §GetPublishPermission 的判定/映射 helper）；多角色时取授权 421 的发布角色优先（选择顺序 grid_worker > community_admin > committee）——`notices.role` 为 NOT NULL 列，落空将 INSERT 失败。
- **错误码**：080002（功能权限，功能层先于 scope）、080003（目标数超限）、080005（请求形状：空范围/双载/division 非法值/非 admin 传 division/附件引用无效或超限）、080006（目标级解析失败或越权，D31）。
- **输出**：`id`。

### community-hub-service.ListNotices（改造）

- **输入**：`community_id`（单值，不变）、`role`（可选筛选，语义不变）、`page`/`page_size`（≤100，默认 10）。
- **流程**：`FilterAllowed(userID, community_id)`（GLOBAL 放行 / LIMITED IN / EMPTY 空列表）→ `FindListByCommunity` JOIN notice_scope（`scope.community_id=?` + **通过态 `moderation_status IN (1,3)`（IsModerationPassed，评审 M1）** + deleted_at IS NULL + 可选 role）→ `order by is_pinned desc, published_at desc`（审核锚定，D27/D32）→ 分页。**JOIN 投影（data-model v4 S3）**：显式 `notices.*, notice_scope.community_id`（右表限定/别名），模型 `CommunityId` = scope 派生值，防 `select *` 双 community_id 列按列序取到弃用 NULL。
- **输出**：`Notice[]`（**community_id = 请求小区**，由 notice_scope 匹配行派生，不读弃用列）+ `total`。
- **鉴权**：REST 权限码 `422 GET:/api/community/notices`（**扩展绑定全部移动端角色**，评审 MUST 1——发布角色 grid_worker/community_admin/committee 经 marquee「更多→浏览」进入浏览页不得 403）由 PermMiddleware 强制；越权读返回空列表，不泄露。
- **错误码**：无（越权读返回空列表，不泄露）。

### community-hub-service.GetNotice（重写）

- **输入**：`id` + **`community_id`（必填，字段号 2；缺失/空 → 080005，D15）**。
- **REST 绑定（评审 MUST 1 修订）**：`GetNoticeReq.CommunityId` 用 **`form:"community_id"`** 标签（GET query 绑定，仓库惯例同 `ListNoticesReq form:"community_id"`）；`id` 用 `path:"id"`。**禁用 `json` 标签**——go-zero httpx.Parse 对 GET 走 ParseForm（formUnmarshaler，key="form"），`unmarshaler.go:571 usingDifferentKeys("form", field)` 对无 form 标签的字段直接 skip，`json` 标签在 query 绑定上不生效 → `CommunityId` 恒 0 → RPC 判缺失 → 080005，详情页整体不可用。community_id 缺失/空由 api 逻辑显式判空 → 080005（可选 `form:"community_id,optional"` + 显式判空）。
- **流程**：`FindOnePublished(id)`（**通过态 1/3，IsModerationPassed，评审 M1**）未找到 → 080001；`notice_scope` 匹配 `(id, community_id)` 不存在 → **080001**（scope 外/未过审不泄露，与不存在同）；`FilterAllowed(userID, community_id)` false → 080001。
- **输出**：`Notice`（community_id=请求小区）+ `attachments[]`（id/file_name/file_url/file_size/**file_type**）。
- **共享转换器同步（data-model v4 S1）**：`toProtoNotice`（`rpc/internal/logic/notice/helper.go`，getnoticelogic/listnoticeslogic 共用）须同步适配模型类型变更——`CommunityId` 为 `*int64`（由请求/scope 注入派生值，非直接解引用弃用列）、`PublishedAt` 用 `sql.NullTime` null 感知取值（NULL 不落 0 时间戳，SEE [[restore-compensation-zero-time]]）；否则 Task 1.3 改类型后 converter 编译即断（`n.PublishedAt.Unix()` 对 sql.NullTime 无此方法）。
- **鉴权**：REST 权限码 `426 GET:/api/community/notices/:id`（**新增，绑定全部移动端角色**，评审 MUST 1——详情端点现无任何码，fail-closed 下全体 403）由 PermMiddleware 强制；scope 级校验（scope 外/未过审 → 080001）在逻辑层。
- **附件 file_url 语义（评审 S4 修订 + data-model v3 M1 闭环）**：落库的 `file_url` 为发布时经 GetFileUrl 返回的**短期预签名 URL**（默认 3600s，上限 7 天），非永久链接；详情读路径对每个附件**按 `notice_attachments.file_id` 重生**预签名 URL（`GetFileUrl(file_id)` 再取，file_type/file_name/file_size 同源回读），保证数周/数月后的旧通知附件仍可下载（REQ-NM-3）。**兼容期 file_id=0/NULL 的存量附件行**：GetNotice 回退返回 stored file_url（尽力而为，与 Q1 存量兼容期一致；后续回填迁移需同时回填 file_id 或声明旧附件不可恢复）。CHANGELOG 登记该语义，防执行方误当永久链接。
- **错误码**：080001（不存在/scope 外/未过审）、080005（community_id 缺失）。

### community-hub-service.GetMarqueeNotices（新建 RPC，D12）

- **输入**：`community_id`（必填；缺失/空/0 → 080005，不回退默认小区，D15）。
- **REST 绑定（评审 MUST 1 修订）**：`GetMarqueeNoticesReq.CommunityId` 用 **`form:"community_id"`** 标签（GET query，与 GetNotice 同——勿用 json/path）；缺失/空 → 080005。
- **流程**：`FilterAllowed`（与列表读路径一致）→ `FindMarquee(community_id, since=now-15*24h, limit=10)`：JOIN notice_scope、**通过态 `moderation_status IN (1,3)`（IsModerationPassed，评审 M1）**、`published_at >= since`（**含端点**，D32）、`order by is_pinned desc, published_at desc`、`limit 10`。
- **输出**：`NoticeMarqueeItem[]{id, title}` ≤10 条（跑马灯数据，不承载正文）。
- **鉴权**：REST 权限码 `423 GET:/api/community/notices/marquee`（绑定全部移动端角色）由 PermMiddleware 强制。
- **错误码**：080005（community_id 缺失）；空态返回空列表。

### community-hub-service.GetPublishPermission（新建 RPC，D3）

- **输入**：空（身份经 JWT metadata，`UserIDFromCtx`）。
- **流程**：permission `GetUserRoles(user_id)`；对每个 role，level-2 判定：`role_code ∈ {grid_worker, community_admin, committee}` 且 `status==2` 且 `verified_at>0`（NOT NULL）且 `expires_at==0 OR expires_at>now`（基于 RPC 输出，禁止直读 rel_user_role）。任一满足 → `can_publish=true` 并追加映射后的 NoticeRole。
- **输出**：`can_publish bool` + `publishable_roles []NoticeRole`（映射：grid_worker→NOTICE_ROLE_GRID_OFFICER、community_admin→NOTICE_ROLE_COMMUNITY、committee→NOTICE_ROLE_COMMITTEE；property_admin/sys_admin/owner/tenant/merchant → can_publish=false，D16/D26）。
- **鉴权**：任意已认证用户（读接口，REST 权限码 `424 GET:/api/community/notices/publish-permission` 绑定全部移动端角色，§权限种子）。
- **错误码**：无（未登录由认证中间件 UNAUTHENTICATED → 前端视为 can_publish=false，REQ-PP-1 场景）。

### permission-service GET /api/perm/data-scopes（REST，评审 MUST 3 钉死鉴权）

- **输入**：`?scope_type=community`（仅接受 community，不返回 community_div——division 选项经 masterdata division 树，D17）。
- **流程**：解析 JWT user_id → 代理 RPC `GetDataScopes` → 返回调用者自身 `scope_ids` + `state`（self-scope 读，返回本人数据范围选项，供网格员发布表单多选小区，REQ-NM-5）。
- **鉴权（钉死）**：**专用权限码 `425 GET:/api/perm/data-scopes`（code `community:data-scopes:read-api`，**parent_id=310**（permission:read 按钮，评审 SHOULD 3——路由在 /api/perm/*，显式指定不依赖 AutoDiscover），绑定全部移动端社区角色**（registered_user 9 / owner 1 / tenant 5 / grid_worker 4 / committee 6 / merchant 7 / community_admin 3 / sys_admin 8，与 423/424 同批种子，评审 MUST 3，见 §权限种子）——**不复用 422**（422 现仅绑 registered_user/owner/tenant，grid_worker 等发布者不在其中，复用会排除发布者）；端点归 **permission-service `PermMiddleware`**（/api/perm 全部路由统一走该中间件）强制。
- **验收**：grid_worker（已认证 + community 数据范围）调 `GET /api/perm/data-scopes?scope_type=community` → 返回其 community `scope_ids`；scope_type 非法 → 错误；未登录 → 认证中间件拒绝。

### community-hub-service.DeleteNotice（撤回收窄，REQ-NP-5）

- **输入**：`id`。
- **流程**：`FindOne(id)` 未找到 → 080001；**作者校验**：`publisher_id == JWT user_id`，否则 **080002**（收窄：原 CheckPublishScope 数据范围判定 → 仅发布者本人，行为回归登记）；**单事务（data-model v4 S2 钉死）**：逻辑层用 `conn.Transact(func(session sqlx.Session) error)` 传共享 session 给两个 model 方法（`SoftDelete` + `NoticeScopeModel.DeleteByNoticeId` 签名改为接受 session/executor），任一失败整体回滚、无孤儿 scope 行；**notice_attachments 行与 MinIO 对象全部保留**（D28）。
- **鉴权**：REST 权限码 `427 DELETE:/api/community/notices/:id`（**新增，绑定全部移动端角色**，评审 MUST 2——现无任何码，撤回对全体用户 403、080002 作者校验永远走不到）由 PermMiddleware 强制；**真正越权判定交 080002 作者校验**。
- **错误码**：080001、080002。

### community-hub-service.UpdateNoticeModerationStatus（回调，published_at 锚定）

- **输入**：`id` + `moderation_status`（复用现有消息，不新增，REQ-NP-MOD-2）。
- **流程**：`FindOne(id)` → 080001；系统身份 scope 校验（`CheckSystemPublishScope`：legacy 行用 `notice.community_id`，新行用 notice_scope 首个小区 id 作为 target；无 scope 行则跳过——系统身份 global 放行，回调可信）；`UpdateModerationStatusPass`：**pass 判定复用 §可见性门禁的 `IsModerationPassed(status)`（1=machine_pass / 3=human_pass）→ 同时置 `published_at=NOW()`**（D27/D30，与读路径同一谓词，评审 M1），非 pass 仅更新 moderation_status + moderation_time。DB 存回调原始值，`IsModerationPassed` 是回调与读查询共用常量，杜绝「回调置 3、读路径只认 1」的漂移。
- **错误码**：080001。
- **幂等**：重复回调幂等（覆盖写）。

### community-hub-service.UpdateNotice（保持，REQ-NP-MOD-3）

- 既有通知级语义（title/content/is_pinned）不变；编辑后 `moderation_status` 重置 0（重审）既有行为不变；**不引入 scope 编辑**（不动 notice_scope）。**逻辑无代码变更，仅回归测试确认**——但**权限码需补齐（评审 MUST 2）**：REST 权限码 `428 PUT:/api/community/notices/:id`（绑定全部移动端角色）由 PermMiddleware 强制，否则编辑端点在中间件层被 fail-closed 拦截在前（现无任何码 → 全体 403），回归测试无法触达。

### file-service.GetUploadUrl（新增快速拒绝层）

- **输入**：`user_id`/`entity_type`/`file_name`/`file_size`/`mime_type`（既有）。
- **校验（L1）**：`file_name` 扩展名 ∈ 白名单 {png,jpg,jpeg,gif,pdf,doc,docx}（大小写不敏感），禁止集 {exe,bat,sh,cmd,com,msi,apk,js,vbs,ps1,py,pl,php} + 全部 zip/rar → 070004；`file_size` > 10MB → 070005（=10MB 放行）；无扩展名/点文件 → 070004。
- **entity_type 覆盖**：全局基线 + 按 entity_type 精细化（可更严不可放宽；10MB 硬上限 + 禁止集不可弱化，REQ-AS-4 不变量）。
- **错误码**：070004、070005（新增登记）。

### file-service.ConfirmUpload（新增 L2 magic-bytes 回读校验）

- **输入**：既有字段 + 回读 MinIO 实际对象。
- **校验（L2）**：按**内容魔数**嗅探真实类型（非扩展名/非客户端 Content-Type）：png `89 50 4E 47`、jpg `FF D8 FF`、gif `47 49 46 38`、pdf `%PDF`、**doc = OLE2/CFB（D0 CF 11 E0 A1 B1 1A E1）且含 `WordDocument` 流**、**docx = ZIP（PK）+ 含 `word/document.xml` 部件**；其他 OLE2 子类型（msi/xls/ppt）与 OOXML 子类型（xlsx/pptx）与通用 zip/rar → **070004**。嗅探类型映射到白名单扩展名，与声明扩展名一致才放行。
- **落库**：`file_type`（嗅探映射的规范扩展名）+ `confirmed=true` 写入 File 记录。
- **语义注记**（评审 INFO 9）：migration 002 对存量行 `confirmed` 置 `DEFAULT 1`，即**存量文件免 magic-bytes 校验即 confirmed**——D24 信任边界下这是有意的迁移折中（存量无 file_type、未嗅探），CreateNotice 的 confirmed 检查对存量文件退化为「文件存在 + 归属本人」；仅新上传走嗅探。**注记**（评审 INFO 1）：`uploaded_file` 行仅在 ConfirmUpload 创建，故对任何存在的行 `confirmed` 恒为 true——GetFileUrl 未确认文件返回 not-found 而非 confirmed=false，CreateNotice 的 confirmed 检查实际是「文件存在 + 归属本人」。
- **错误码**：070004（类型主码）、070005（大小次码）。

### 跨服务一致性

- **CreateNotice 单事务**：notices + notice_scope + notice_attachments 在同一 MySQL 事务内 all-or-nothing（本地事务，无跨服务写）。
- **moderation 队列**：异步（Redis LPUSH），审核回调 UpdateNoticeModerationStatus 回写（既有最终一致性链路；published_at 通过回调置 now，D27）。
- **附件校验**：经 file `GetFileUrl` 只读（无跨服务写），FileInfo 为校验载体（D24）。
- **division 展开**：masterdata `GetResidentialAreasByDivision` 只读快照（发布时固定，后续 division 成员变化不影响已发布通知，REQ-NP-4 场景）。

---

## 权限种子（读/写权限矩阵，评审 interface-proto v4 MUST 1/2 修订）

> **fail-closed 鉴权语义（磁盘核实 checkpermissionlogic.go:55-59）**：`permissionDefMinLevel(needle)` 找不到权限 def（sys_permission 无匹配 type=3 行）→ `Allowed=false` → 中间件 403。**因此移动端每个通知 REST 端点必须有对应 sys_permission 权限码并绑定角色**，否则所有用户一律 403、RPC 契约（GetNotice 080001/080005、DeleteNotice 080001/080002 等）永远不可达。本变更补齐通知**读路径**（详情 426 / 浏览列表 422 扩展）与**写路径**（撤回 427 / 编辑 428）权限矩阵。

### 写路径权限

| 权限码 | 端点 | code | parent_id | min_verf_level | 角色绑定 |
|--------|------|------|:---:|:---:|------|
| 421 | POST:/api/community/notices（创建） | community:notice:create-api | 420（community:notice 按钮） | 0→**2**（REQ-PP-3 写路径角色状态门槛） | **grid_worker(4) 新增授**（现无 421）；**property_admin(2)/owner(1)/tenant(5) 回收**（D26，行为回归：物业/业主/租户原可发布 → 只读）；community_admin(3)/committee(6) 保留 |
| 427 | DELETE:/api/community/notices/:id（撤回，REQ-NP-5） | community:notice:delete-api | 420（community:notice 按钮） | 0 | **全部移动端角色** {owner 1 / community_admin 3 / grid_worker 4 / tenant 5 / committee 6 / merchant 7 / sys_admin 8 / registered_user 9}（评审 MUST 2——现无任何码，fail-closed 下撤回对全体用户 403；**真正越权判定交 DeleteNotice 080002 作者校验**） |
| 428 | PUT:/api/community/notices/:id（编辑，REQ-NP-MOD-3） | community:notice:update-api | 420（community:notice 按钮） | 0 | **全部移动端角色**（同上，评审 MUST 2——现无任何码，编辑端点在中间件层被拦截在前；UpdateNotice 保持既有通知级语义 title/content/is_pinned，无新增越权判定） |

### 读路径权限

| 权限码 | 端点 | code | parent_id | min_verf_level | 角色绑定 |
|--------|------|------|:---:|:---:|------|
| 422 | GET:/api/community/notices（浏览列表） | community:notice:read-list-api | 410（community:read 按钮） | 0 | 由现 (9,1,5) **扩展为全部移动端角色** {1,3,4,5,6,7,8,9}（评审 MUST 1——本变更新授 421 的 grid_worker(4)/community_admin(3)/committee(6) 现无 422，marquee「更多→浏览」引向浏览页时 403；与 423/424/425 同批扩展，避免「425 因复用 422 排除发布者而新建码、422 自身却仍排除发布者」的自相矛盾） |
| 423 | GET:/api/community/notices/marquee（跑马灯） | community:notice:read-marquee-api | **410**（评审 SHOULD 3，同 422） | 0 | 全部移动端角色 {1,3,4,5,6,7,8,9} |
| 424 | GET:/api/community/notices/publish-permission | community:notice:publish-permission-api | **410**（评审 SHOULD 3，同 422） | 0 | 全部移动端角色 {1,3,4,5,6,7,8,9} |
| 426 | GET:/api/community/notices/:id（详情，REQ-NR-2） | community:notice:read-detail-api | **410**（评审 SHOULD 3，同 422） | 0 | **全部移动端角色** {1,3,4,5,6,7,8,9}（评审 MUST 1——现无任何码，对照 user 112 / role 212 均有显式详情读码，通知详情端点零匹配 → 全体 403，GetNotice 080001/080005 契约不可达、REQ-NM-1 点击标题→详情不可用） |
| 425 | GET:/api/perm/data-scopes（permission-service，范围选项） | community:data-scopes:read-api | **310**（permission:read 按钮，评审 SHOULD 3——路由在 /api/perm/*，AutoDiscover 对 /api/perm/data-scopes 会 miss 落到 /data-scopes 菜单，本变更在 init_permissions.sql 显式指定 parent_id 不依赖自动发现） | 0 | 全部移动端角色 {1,3,4,5,6,7,8,9} |

> **「全部移动端角色」集合**：8 个社区角色 {owner 1 / community_admin 3 / grid_worker 4 / tenant 5 / committee 6 / merchant 7 / sys_admin 8 / registered_user 9}，与 423/424/425 同批（评审 MUST 1 指引）。property_admin(2) `platforms='pc'`（非移动端），本变更不绑（Q6/D26）。
> **读码 min_verf_level=0**：持角色即可读，数据范围/过审过滤在 community-hub `FilterAllowed`/scope 校验（与现有 422 语义一致）；写码 421 置 2（REQ-PP-3），427/428 置 0（真正授权交业务层作者校验/通知级语义）。
> **INFO 5（幻影 435）**：`community:lostfound:create-api` 在种子中**无 sys_permission 行**（仅 line 202 `UPDATE ... WHERE code IN` 引用，rel_role_permission 却绑 (1,435)/(5,435) 指向幻影 id）——本变更不新增 435 行，措辞统一为「lostfound:create-api 权限无 sys_permission 行，保持现状不动」，核 Task 5.1 断言不与幻影 435 冲突。
> **INFO 6**：本变更将 421/422/423/424/425/426/427/428 权限码写入 `docs/specs/rbac-design.md` §6.5 角色验收矩阵（.change.yaml 已列该文件为修订对象）。

---

## Design Gate（D17/REV-17）—— division→community 授权落位

**证据（读 assertpublishscopelogic.go + scope.go）**：
- `AssertPublishScopeLogic.AssertPublishScope` 调 `resolveUserScope(..., scopeType='community')`，其实现 `g.ScopeType != scopeType → continue`，**只收集 `community` scope_type 的 grant**。
- 社区管理员的 division grant 是 `scope_type='community_div', scope_id=D1`，**不会被收集**进授权 id 集。
- `targetCovered(nodeID, ids)` 解析目标小区祖先链（`ResolveScopeAncestors`，self-first ≤6）∩ ids——由于 D1 不在 ids，division 授权**无法覆盖**目标小区 → division 发布会被误拒。

**裁决（design gate 结论）**：permission-service `AssertPublishScope` 判据逻辑**必须变更**。推荐方案：在 AssertPublishScope 内新增 `resolvePublishScope`（收集 `community` ∪ `community_div` 双 scope_type 授权并集，供 targetCovered 使用）；`GetDataScopes` 读路径保持 `community` 单 scope_type 不变。

**共享 blast radius 声明（评审 SHOULD 2）**：`AssertPublishScope` 是**共享 RPC**，除通知创建（新 AssertCommunitiesScope）外还被 lostfound 创建（createlostfoundlogic.go）、contacts upsert（upsertcontactslogic.go）调用。`resolvePublishScope` 生效后，community_admin 的 `community_div` grant 将**连带放行** lostfound/contacts 写到 division 下小区——这大概率是修复（division 授权本就该生效，符合数据所有权），但属**跨服务共享语义变更**，必须显式声明并经回归验证（见验证门禁 4），不得仅按「通知发布」框定。

**验证门禁（编码前必须通过）**：permission-service 单测/集成验收覆盖——
1. community_admin 持 `community_div=D1` grant，可发布到 C1（`community_div_id=D1`，祖先链含 D1）→ allowed；
2. 发布到 D1 外的小区 C2 → 060007（denied）；
3. 目标小区不存在（ResolveScopeAncestors found=false）→ denied（安全拒绝未知节点，060007 面）；
4. **共享调用方回归（评审 SHOULD 2）**：community_admin 持 `community_div=D1` 时（a）lostfound 发布到 C1（祖先链含 D1）→ allowed、（b）contacts upsert 到 C1 → allowed（如产品语义要求隔离则显式声明并在 resolvePublishScope 内按调用方隔离）。

> 该 Task 标注 `SEE: [[grpc-timeout-layers]]`（AssertPublishScope 内嵌 ResolveScopeAncestors 三层超时对齐）。

---

## Proto 变更

### community/v1/community.proto

| 变更 | 类型 | 破坏性 | 说明 |
|------|:---:|:---:|------|
| 头注释错误码块对齐（D29） | 注释 | 否 | 080003=单次发布目标数超限、080005=参数无效（含小区ID无效）、080006=目标小区超出发布者数据范围；剔除陈旧「080003 寻失发布次数已达上限」（实际码 080007） |
| `CreateNoticeRequest` +`community_ids`(8, repeated int64 JS_STRING) | 新增字段 | 否 | 多小区目标；`community_id`(1) deprecated 保留 |
| `CreateNoticeRequest` +`division_id`(9, int64 JS_STRING) | 新增字段 | 否 | 仅 community_admin，值域 `md_administrative_division.id` |
| `CreateNoticeRequest.role`(4)/`publisher_id`(6) 标记 deprecated | 注释/机器可读指令 | 否 | 服务端从 JWT 派生，请求体不信任（评审 SHOULD 4）；`publisher`(5) **不 deprecated**（请求体展示字符串，评审 S1） |
| `GetNoticeRequest` +`community_id`(2, int64 JS_STRING) | 新增字段 | 否 | 必填请求上下文（缺失 080005，D15） |
| `NoticeAttachment` +`file_type`(5, string) | 新增字段 | 否 | 白名单校验通过的文件类型（D24） |
| 新增 RPC `GetPublishPermission` + `GetPublishPermissionRequest/Response` | 新增 RPC | 否 | can_publish + publishable_roles（D3） |
| 新增 RPC `GetMarqueeNotices` + `GetMarqueeNoticesRequest/Response` + `NoticeMarqueeItem` | 新增 RPC | 否 | 跑马灯数据 ≤10 条（D12） |

### file/v1/file.proto

| 变更 | 类型 | 破坏性 | 说明 |
|------|:---:|:---:|------|
| 头注释错误码块对齐（D11） | 注释 | 否 | 070001 文件不存在 / 070002 文件访问被拒绝 / 070003 文件操作失败 / **070004 文件类型不支持** / **070005 文件大小超限**；修正漂移的「070002 上传失败/070003 文件类型不支持」 |
| `FileInfo` +`file_type`(11, string) | 新增字段 | 否 | 白名单规范类型（ConfirmUpload magic-bytes 层产出，D24） |
| `FileInfo` +`confirmed`(12, bool) | 新增字段 | 否 | 上传流程完成标志（REQ-AS-7） |

> 破坏性影响评估：wire 上全部为**兼容新增**（新字段号/新 RPC），`buf breaking-check` 通过；生成后调用方（community-hub / web-mobile / file-service）同步更新。**语义破坏项（评审 SHOULD 6，CHANGELOG 显式登记 + 兼容期）**：① `GetNotice` 新增必填 `community_id`（缺失 → 080005）——未升级消费方（仅传 id）将全部 080005，属行为回归；已核实 web/pc 无 notice 读消费方、web/mobile 在本变更内同步升级，破坏面收敛，但外部消费方若存在会被静默破坏；② legacy `CreateNotice` 仅传 `community_id`(1) 不再接受（→ 080005，不回退），兼容期=本变更上线即生效，回退行为=回滚 proto 变更。

---

## 安全考虑

- **附件上传两层校验**（Q5/D5）：L1 GetUploadUrl 扩展名+大小快速拒绝；L2 ConfirmUpload **magic-bytes 内容嗅探**回读实际对象，非扩展名/非客户端元数据。doc 按 OLE2/CFB + `WordDocument` 流、docx 按 ZIP + `word/document.xml` 部件显式放行；`msi/xls/ppt 改 .doc`、`xlsx/pptx 改 .docx` 一律 070004（D18，封堵同容器子类型绕过）。
- **附件引用信任边界**（D24）：community-hub 经 `GetFileUrl(file_id)` 读扩展 FileInfo（`confirmed` + `user_id` 归属 + `file_type` 回读），不信任客户端回传类型；未确认/他人文件 → 080005。
- **权限信任边界**：`role`/`publisher_id` 以 JWT 实际身份为准，请求体伪造被纠正（REQ-NP-3 场景）。
- **读路径隐私**：GetNotice scope 外/不存在/未过审统一 080001（不泄露跨小区/未审内容）；越权 List 返回空列表；`GetMarqueeNotices` 仅审核通过。
- **可见性谓词一致性**（评审 M1）：通过态 `IN (1,3)` 由 `IsModerationPassed` 单一谓词承载，读查询与回调共用；DB 存回调原始值，status=3（human_pass）落库后必须在列表/详情/跑马灯可见，不得出现「置了 published_at 却读不到」的漂移。
- **目标级解析失败 fail-closed**（D31）：不存在的 community_id 与越权同处理 080006，不静默创建无效小区通知。
- **权限回收回归**（REQ-PP-4/D26）：property_admin/owner/tenant 回收 421 → 写路径 080002；`min_verf_level=2` 堵未认证发布（SEE [[auto-grant-unverified-grant-confers-scope-level0]]——未认证 grant 立即生效的既有语义，现经 min_verf_level 收窄）。
- **权限矩阵完整性（fail-closed，评审 interface-proto v4 MUST 1/2）**：移动端每个通知 REST 端点必须有 sys_permission type=3 权限码并绑定角色（缺 def → 403，RPC 契约不可达）。读路径：详情 426 + 列表 422（扩展全部移动端角色）+ 跑马灯 423 + publish-permission 424；写路径：创建 421 + 撤回 427 + 编辑 428（均绑定全部移动端角色，真正越权判定交业务层：DeleteNotice 080002 作者校验）。新权限码 path 与实际 REST 路由逐一对照（SEE [[permission-seed-api-path-must-match-routes]]），parent_id 显式指定（423/424/426→410、427/428→420、425→310）。
- **RPC 身份**：`UserIDFromCtx` 盲信入站 metadata，前提是 RPC 回环绑定（既有）；新 RPC 同约束（SEE [[rpc-identity-spoofing-loopback-isolation]]）。

---

## 记忆引用（Step 1.5 注入，slug 均已自验存在）

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[proto-jstype]] | Proto | community_ids/division_id/GetNoticeRequest.community_id 均 int64 + `[jstype=JS_STRING]` |
| [[grpc-only-comms]] | 接口设计 | 附件校验/角色查询/scope 校验全部经 gRPC，禁止直读 rel_user_role/uploaded_file 表 |
| [[migration-must-execute]] | 数据模型 | 003/002 迁移文件提交后必须执行；去 NOT NULL 迁移先于功能上线（REQ-NP-1 门禁） |
| [[error-code-collision-and-namespace-alignment]] | Proto/错误码 | 070004/070005 全新整数位不重编号 70003；080005/080006 消歧（D11/D31） |
| [[error-code-literal-bypasses-qa-gate]] | 错误码 | 070004/070005/080003/080005/080006 用 errx 命名常量登记，禁裸数字 |
| [[permission-seed-api-path-must-match-routes]] | 权限种子 | 新读/写接口 423/424/426/427/428 path 与 REST 路由逐一对照（`GET:/api/community/notices/:id`/`DELETE:/.../:id`/`PUT:/.../:id` 等）；421 path 保持 POST:/api/community/notices |
| [[is-system-no-permission-shortcut]] | 权限种子/Design Gate | sys_admin 全权限经 rel_role_permission 配置；AssertPublishScope 无字段短路 |
| [[auto-grant-unverified-grant-confers-scope-level0]] | 权限种子 | 421 置 min_verf_level=2 收窄未认证发布；与 GetPublishPermission level-2 判定一致 |
| [[insert-ignore-swallows-errors]] | 权限种子 | 回收 421 必须显式 DELETE（INSERT IGNORE 无法撤销）；grid_worker 授 421 用 INSERT IGNORE |
| [[grpc-timeout-layers]] | Design Gate/GetPublishPermission | AssertPublishScope(≤500ms) 内嵌 ResolveScopeAncestors；GetUserRoles 超时对齐 |
| [[rpc-identity-spoofing-loopback-isolation]] | 安全 | UserIDFromCtx 盲信 metadata 的回环前提；新 RPC 同约束 |
| [[rpc-callback-must-check-response-base]] | 回调 | UpdateNoticeModerationStatus 消费方检查响应 Base，不只看 gRPC err |
| [[async-submit-double-guard]] | 前端 | 发布表单提交中禁用（D25，后端不幂等） |
| [[snake-camel-field-mismatch]] | 前端 | TS 字段名与 Go snake_case 对齐；file_type/file_size/community_ids 等 |
| [[api-required-field-marked-optional]] | 接口设计 | GetNotice.community_id / GetMarqueeNotices.community_id 必填，禁止误标 optional |
| [[go-deprecated-directive-not-test-comment]] | Proto | community_id(1) 的 deprecated 是机器可读指令，不当作自测/占位注释 |
| [[tdd-red-evidence-requires-fail-excerpt]] | 任务 TDD | 含逻辑任务 RED 证据须含实际 FAIL 摘录 |
| [[cross-service-seed-deployment-order]] | 部署 | community-hub 003 迁移 + permission 种子 + file 002 迁移纳入部署编排（迁移先于上线） |
| [[moderation-status-write-without-read-gating]] | 可见性门禁 | 读路径（list/detail/marquee）恒过滤**通过态 `IN (1,3)`**（`IsModerationPassed` 单一谓词，评审 M1 修订：原「恒过滤=1」与回调 pass={1,3} 矛盾，已统一） |
| [[restore-compensation-zero-time]] | 数据模型 | published_at 用 `sql.NullTime` 处理 NULL，严禁 `time.Time{}` 零值写 DATETIME |

**不适用记忆（主动排除，供 reviewer 确认非遗漏）**：
- [[unique-index-migration-dup-precheck]] — notice_scope 为**新表**无存量重复；唯一约束用于写路径去重兜底（业务层先 dedupe），排除。
- [[notfound-cache-sentinel-vs-transient-error]] — 本变更无 Redis 缓存新增（不做缓存层，Q 决策），排除。
- [[redis-cache-soft-delete]] — 本变更不动 Redis 索引缓存，排除。

---

## 非功能设计

- **可靠性**：CreateNotice 单事务 all-or-nothing；AssertPublishScope/GetUserRoles 依赖失败 fail-closed（080002/080006 或传输错误）；moderation 异步队列不阻塞发布（既有）；DeleteNotice 撤回全局生效（单事务）。UpdateNoticeModerationStatus 幂等（覆盖写）。
- **性能**：列表/跑马灯走 `idx_scope_community` 索引兜底（无缓存，沿用现有策略）；跑马灯 limit 10 + 15 天窗口；附件校验逐条 GetFileUrl（≤10 条 × 1 RPC）；GetUserRoles 只读。单次发布目标 ≤100 限制（080003）。
- **可观测性**：CreateNotice 日志（id/目标数/角色）、GetPublishPermission/GetMarqueeNotices 关键路径 Infof；070004/070005 拒绝计数可经日志聚合；错误码全量 errx 常量。
- **显式「无」项**：无新增缓存层；无幂等键；无批量操作；无 CDN/病毒扫描（magic-bytes 仅类型判定）。

---

## 关键设计决策与权衡（ADR，轻量）

| 决策点 | 备选方案 | 最终选型 | 取舍理由 | 未采用方案原因 |
|-------|---------|---------|---------|--------------|
| 范围关联单源 | 复用 community_id 列（repeated 不可行）/ 独立 notice_scope 表 | notice_scope 表 | 一通知多小区需 1:N；单源避免双写漂移（D1） | community_id 单值无法表达多目标 |
| 响应 community_id | 改 repeated community_ids / 保持单值 | 保持单值（取请求小区） | 最小改动，前端无感（Q2/D2） | repeated 波及全部消费方 |
| division→community 授权 | 不改判据（误拒）/ 收集双 scope_type | **收集 community ∪ community_div**（resolvePublishScope 变体） | 证据证明现判据漏收集 division grant（§Design Gate） | 不改判据则社区管理员 division 发布全被 080006 误拒 |
| published_at 锚点 | 创建时即设（现状）/ 审核通过时设 | 审核通过回调置 now（D27/D30） | 跑马灯/列表从可见日起算，语义正确 | 创建时即设使待审内容占用窗口 |
| 附件类型判定 | 扩展名/MIME 元数据 / magic-bytes 回读 | 两层（L1 快速 + L2 魔数） | 客户端直传下元数据可伪造；魔数防绕过（Q5/D5） | 仅 L1 可改名 png 传 exe 绕过 |
| 附件校验载体 | 新 RPC 查询 / 扩展 FileInfo 复用 GetFileUrl | 扩展 FileInfo + GetFileUrl（D24） | 零新 RPC，校验/读回一次完成 | 新 RPC 增加契约面 |
| property_admin 移动端发布 | 保留 / 剔除 | 剔除（D26） | 收敛发布角色为 grid_worker/community_admin/committee | 保留与「物业只读」产品方向冲突 |
| CreateNotice 幂等 | 幂等键 / 不幂等 + 前端防重 | 不幂等（D25） | 简化后端；防重由提交中禁用承担 | 幂等键增加状态面，收益低 |
| 撤回附件处置 | 删附件行+对象 / 全保留 | 全保留（D28） | 撤回不销毁证据；删除走独立审核 | 删除会丢审计链 |
| 错误码冲突 | 复用 70003 / 新 70004/70005 | 新整数位（D11） | 消除同整数双语义 | 复用 70003 破坏既有 ErrCodeFileOperationFailed |
| 可见性谓词 | 恒过滤 status=1（现状）/ 通过态 IN(1,3) 单一谓词 | IN(1,3) + `IsModerationPassed` | 回调 pass={1,3} 与读路径一致，杜绝 human_pass 置 published_at 却读不到（评审 M1） | 仅过滤 1 则 status=3 行失联，违反 REQ-NP-MOD-1 |
| REST community_ids 绑定 | `[]int64 json:",string"` / `[]string`+strconv | `[]string`（代理前转换） | encoding/json `,string` 不支持 slice；移动端发 string 形式 Snowflake ID 主链路解包失败（评审 MUST 1） | `[]int64` 在 REST 层直接解包报错 |
| 重审 published_at 副作用 | 编辑重审不刷新 published_at / 刷新 | 刷新（D27/D30 既有语义） | 重过审重新锚定可见日，跑马灯/列表重排顶部，属可感知行为副作用（评审 SHOULD 7 显式声明） | 不刷新则窗口锚定创建日，与 D27/D32 矛盾 |
| 附件 file_url 持久化 | 存永久 URL / file_id 权威 + 读时重生预签名 | file_id 权威 + 详情读时重生预签名 URL | 预签名 URL 有 3600s~7 天有效期，旧通知附件直接落库会过期死链（评审 S4） | 存永久 URL 与预签名模式冲突，需自建签名中心 |
| AssertPublishScope 共享 blast radius | 仅按通知发布框定 / 显式声明 + lostfound·contacts 回归 | 显式声明 + 回归验证（评审 SHOULD 2） | 共享 RPC 判据变更会连带 lostfound/contacts 写路径，须回归确认 division 授权放行属修复 | 不声明则跨服务语义漂移未被识别，评审放行时遗漏 |
| 通知读端点权限矩阵 | 详情/浏览无权限码（fail-closed 全体 403）/ 补齐 426+422 扩展 | 补齐 426 GET:/api/community/notices/:id + 422 扩展全部移动端角色（评审 interface-proto v4 MUST 1） | 详情/浏览是移动端核心读交付，缺 def 即 403 击穿端到端验收 | 「发布角色不浏览+隐藏入口」二选一亦可，但产品要求全角色可浏览（REQ-NM-1/2/3），补齐更贴合 |
| 通知写端点权限矩阵 | 撤回/编辑无权限码（fail-closed 全体 403）/ 新增 427/428 绑定全角色 | 新增 427 DELETE + 428 PUT 绑定全部移动端角色（评审 interface-proto v4 MUST 2） | 撤回是本变更核心交付（REQ-NP-5）；真正越权判定交 DeleteNotice 080002 作者校验，与 design 语义一致 | 「撤回/编辑不在移动端开放+隐藏入口」会砍掉核心交付，不可取 |
| 新权限码 parent_id | AutoDiscover 推断（/api/perm/data-scopes miss）/ 种子显式指定 | 种子显式指定：423/424/426→410、427/428→420、425→310（评审 SHOULD 3） | 避免 425 落到 /data-scopes 孤儿菜单；与既有 422（→410）、421（→420）树一致 | 依赖自动发现会生成菜单树不一致的孤儿节点 |
| division_id 空串绑定 | 非 admin 发 `division_id:""`（`,`string` 解空串 4xx）/ 不适用时省略字段 | 移动端不适用时**省略** division_id 字段，不发送空串（评审 SHOULD 4）；RPC 侧 double-empty 判 080005 兜底 | `encoding/json ,string` 解空串报错会在 REST 层 4xx，进不到 RPC 080005 | 发送空串会在 REST 层报错，语义混乱 |
| 附件 file_url 存储 | 存预签名 URL 快照（VARCHAR(500) 截断风险）/ **新行写占位空串 + file_id 权威** + file_url 加宽 1024 | 新行 file_url=''（file_id 权威重生载体）+ migration 003 加宽 VARCHAR(1024)（评审 S4） | 预签名 URL 含 X-Amz-Signature 可超 500 字符 → ERROR 1406 → 整事务回滚；file_id 已为重生载体，快照冗余 | 仅加宽仍存冗余快照；仅 file_id 不加宽则存量长 URL 仍可能超限 |
| toProtoNotice 转换器 | 模型类型变更后 converter 不更新（编译断）/ Task 同步适配 | Task 1.3/1.8 同步更新 `toProtoNotice`（*int64 + sql.NullTime null 感知 + community_id 由 scope 注入）（评审 S1） | 共享 converter 被 getnotice/listnotices 共用，改模型不改 converter 编译即断 | 忽略则 build 反复回退 |
