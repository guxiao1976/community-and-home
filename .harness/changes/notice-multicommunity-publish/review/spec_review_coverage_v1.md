# Plan Review — notice-multicommunity-publish（覆盖完整性视角）

**审查维度**: 需求覆盖、场景完整性、边界识别
**审查版本**: P1.3（fallback:r0:rc3）— 按磁盘最新内容独立重审（specs 于 2026-08-16 00:09-00:13 更新，晚于 v3 评审 23:53；哈希与本轮历史 r0:rc1/r1:rc1/r2:rc1 均不同，spec 已更新，故独立重审）
**审查基线**: `.change.yaml` + `proposal.md` + `specs/*/spec.md`（change 目录无 request.md，以 proposal 为原始需求基准）+ 实际代码/种子核验（init_permissions.sql、community.proto、file.proto、masterdata.proto、notice_attachments DDL、createnoticelogic.go、section_quota.go）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 3

总评：r0:rc3 版 spec 覆盖完整性已高度收敛。32 个 Requirement（notice-publish 7 + notice-read 3 + publish-permission 4 + attachment-security 7 + notice-moderation 4 + notice-mobile 7）覆盖 proposal/design 全部 29 个决策点（D1-D29），绝大多数具备 ≥1 正向 + ≥1 异常 Scenario；v1/v2/v3 三轮全部 MUST FIX 与 SHOULD FIX 均已在磁盘上逐条落实（见「上轮问题修复校验」）。本轮未发现 MUST FIX 级覆盖空洞。剩余 2 个 SHOULD FIX 均为边界补强（附件 file_type 列迁移未显式成 Requirement；直接多选发布路径与 division 展开路径的目标小区审核状态要求不对称，v3 INFO I-1 残留），不阻塞阶段 3 设计。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| SF-1 | specs/attachment-security/spec.md REQ-AS-5 + specs/notice-publish/spec.md REQ-NP-6 | **`notice_attachments.file_type` 列的 schema 迁移未显式成 Requirement**。REQ-AS-5 要求「persist a file_type field on notice_attachments」，REQ-NP-6 要求 CreateNotice 时写入 file_type——但现状 DDL（services/community-hub-service/migration/001_initial.sql:28-36）notice_attachments 仅 (id, notice_id, file_name, file_url, file_size, created_at)，无 file_type 列。与 REQ-NP-1（notices.community_id 去 NOT NULL 的 ALTER + 「迁移先于上线」门禁 + 正向/异常场景）形成对照：REQ-AS-5/REQ-NP-6 均未定义加列迁移、未定义旧行 file_type 的可空/默认语义（REQ-AS-5 场景 2 隐含旧行 file_type 为空，但未落到 DDL 约束），也未挂部署门禁。遗漏该列迁移 → 带附件通知的 notice_attachments INSERT（含 file_type）直接 SQL 失败，附件发布主链路不可用。 | 在 REQ-AS-5（或并入 REQ-NP-1 迁移 Requirement）显式补 1 条迁移 Requirement：「ALTER TABLE notice_attachments ADD COLUMN file_type VARCHAR(20) NOT NULL DEFAULT ''」（或等价，旧行回读为空），定义旧数据可空/回读空语义，并挂「迁移先于功能上线」门禁 + 一个异常场景（列未迁移时带附件发布 INSERT 失败、无部分写入）。 |
| SF-2 | specs/notice-publish/spec.md REQ-NP-3/REQ-NP-4 + specs/notice-read/spec.md REQ-NR-1 | **直接多选发布路径与 division 展开路径的「目标小区审核状态」要求不对称（v3 INFO I-1 残留未闭合）**。division 路径显式只含审核通过小区（REQ-NP-4：「expansion SHALL include only residential areas whose submission status is approved, encoded by status=1」）；而直接多选路径（grid_worker 的 community_ids）只要求目标「exist in md_residential_area」（REQ-NP-3 网格员多选场景 GIVEN），未声明目标须审核通过。边界后果：网格员若持有未审核通过小区的 scope grant，可直接发布到该小区，而社区管理员走 division 展开永远不会覆盖该小区——两路径发布范围语义不一致，实现者可据此绕过审核通过过滤。 | 补 1 句产品决策（或场景）：显式声明直接多选路径的目标小区是否同样要求审核通过（若要求，则 AssertPublishScope 前补 approved 过滤；若不要求，则说明 division 的 status=1 是展开层过滤、直接多选信任 scope grant），闭合两条路径的边界不对称。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I-1 | REQ-NP-5（DeleteNotice）：无「通知不存在/已软删重复撤回」边界场景——发布者对已软删/不存在的通知再调 DeleteNotice 应返回什么（080001？）未定义。建议补一句。 |
| I-2 | REQ-NR-3/REQ-NM-1：跑马灯「置顶优先」在 15 天窗口内生效，但「置顶且超过 15 天的通知不入跑马灯」未以显式边界场景断言（文本「published_at within last 15 days — ordered by is_pinned first」已可推导，但建议一句话钉死，防实现者误解置顶可突破窗口）。 |
| I-3 | REQ-NP-3（D29 错误码对齐）：社区 API 层 types.go 仍声明 `CodeOverLimit = 80003 // 超限（复用）`（services/community-hub-service/api/internal/types/types.go:18，当前无引用）。spec 将 080003 语义对齐为「单次发布目标数超限」，但未说明该已声明常量在新语义下的处置——避免「080003 超限」在通知域与寻失域未来同时复用造成双语义。建议在 spec 一句话确认 CodeOverLimit 是否保留/退役。 |

## 上轮问题修复校验（v3 → r0:rc3）

| # | v3 问题 | 本轮状态 |
|---|---------|---------|
| S-1 published_at 时序 | D27 → REQ-NP-MOD-4（经 UpdateNoticeModerationStatus 通过回调设置 published_at）+ REQ-NR-3/REQ-NM-1/REQ-NM-3 锚定可见日；行为登记（原 createnoticelogic.go:62 创建时即设，已核验） | ✅ 已修复 |
| S-2 GetMarqueeNotices 入参校验 | REQ-NR-3 补「community_id 缺失/空/零 → 080005，不回退默认小区」异常场景 | ✅ 已修复 |
| S-3 REQ-AS-4 异常场景 | REQ-AS-4 补「override 试图放宽 10MB / 放行禁止集 → 070004/070005 拒绝」异常场景 | ✅ 已修复 |
| I-1 直接多选 vs division 目标审核状态 | **未闭合**（见 SF-2） | ⏳ 残留为 SHOULD FIX |
| I-2 附件总量边界 / I-3 撤回附件处置 / I-4 幂等 | D23 → REQ-AS-6/REQ-NP-6（≤10 个/≤50MB，080005）；D28 → REQ-NP-5（附件行 + MinIO 对象保留）；D25 → REQ-NP-7（后端不幂等 + 前端防重） | ✅ 已修复 |

v1（r0:rc1）MF-1（community_id 去 NOT NULL 迁移）、MF-2（GetNotice community_id 上下文）、v2（r1:rc1）MF-1（写路径角色状态门槛 min_verf_level=2）均已随 v3 判定为已修复，本轮复检仍保持已修复状态（REQ-NP-1 迁移 + REQ-NR-2 080005 + REQ-PP-3/REQ-PP-4 种子变更，均已落到磁盘）。

## 事实核验（本轮对照实际代码，全部与 spec 断言一致）

- init_permissions.sql：grid_worker(4) 当前无 421（仅 100/110/111/112/400/410/411）✅；property_admin(2)/owner(1)/tenant(5) 当前持 421（REQ-PP-4 回收属行为回归）✅；421 当前 min_verf_level=0（§4.2）→ 置 2 ✅
- community.proto：CreateNoticeRequest 现有字段 1-7，community_ids(8)/division_id(9) 字段号空闲 ✅；GetNoticeRequest 仅 id=1，community_id(2) 字段号空闲 ✅；NoticeRole 枚举含 GRID_OFFICER=4/COMMUNITY=1/COMMITTEE=2/PROPERTY=3，映射一致 ✅
- file.proto：FileInfo 现有字段 1-10，file_type(11)/confirmed(12) 字段号空闲 ✅；GetFileUrl 存在（返回 FileInfo 含 user_id/file_size）✅；file errcode 现仅 70001-70003，70004/70005 无冲突 ✅
- masterdata.proto：GetResidentialAreasByDivision(community_div_id, status: 0=all/1=approved only) 存在 ✅；GetDivisionTree 存在（REQ-NM-5 division 选项数据源）✅
- section_quota.go:13 `CodeSectionQuotaExceeded = 80007`，寻失板块配额实际码为 080007 → D29「剔除陈旧 080003 寻失注释」断言正确 ✅
- notice_attachments DDL 无 file_type 列 → SF-1 迁移缺口成立 ✅

## 问题跟踪表

| # | 状态 |
|---|------|
| SF-1 notice_attachments.file_type 迁移 | 新增（SHOULD FIX） |
| SF-2 直接多选 vs division 目标审核状态 | 残留（v3 INFO I-1 → SHOULD FIX） |
| I-1 DeleteNotice 已删/不存在边界 | 新增（INFO） |
| I-2 跑马灯置顶超窗口边界 | 新增（INFO） |
| I-3 CodeOverLimit=80003 常量处置 | 新增（INFO） |

---
VERDICT: APPROVED
---
