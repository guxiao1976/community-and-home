# Plan Review — notice-multicommunity-publish（覆盖完整性视角）

**审查维度**: 需求覆盖、场景完整性、边界识别
**审查版本**: P1.3（fallback:r1:rc3）— 哈希与历史轮次（r0:rc3 / r1:rc1 / r2:rc1）不同，按磁盘最新内容独立重审，勿沿用旧轮结论
**审查基线**: `.change.yaml` + `proposal.md` + `specs/*/spec.md`（change 目录无 request.md，以 proposal + 设计文档为原始需求基准）+ 实际代码/种子核验（001_initial.sql、init_permissions.sql、community.proto、file.proto、masterdata.proto、permission.proto、model/notice.go、createnoticelogic.go）

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2

总评：r1:rc3 版 spec 覆盖完整性高度收敛。32 个 Requirement（notice-publish 7 + notice-read 3 + publish-permission 4 + attachment-security 7 + notice-moderation 4 + notice-mobile 7）覆盖 proposal/design 全部 32 个决策点（D1-D32），绝大多数具备 ≥1 正向 + ≥1 异常 Scenario；v1/v2/v3 三轮全部 MUST FIX 与绝大部分 SHOULD/INFO 已逐条落实（published_at 迁移 D30、错误码消歧 D31、迁移门禁异常场景、跑马灯 published_at 锚 D32 均已闭合）。本轮保留 1 个 MUST FIX（v1 SF-1 遗留未闭合）：**`notice_attachments.file_type` 列的 schema 迁移未显式成 Requirement**——与 REQ-NP-1 中 community_id/published_at 迁移同 class，缺 ADD COLUMN 迁移契约、缺「列未迁移→附件绑定 INSERT 失败」异常门禁、缺部署先于上线门禁，遗漏则该列缺失导致带附件发布主链路不可用。1 个 SHOULD FIX（v1 SF-2 遗留）：直接多选路径与 division 展开路径的目标小区审核通过要求不对称。

## 发现

### 🔴 MUST FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| MF-1 | specs/attachment-security/spec.md REQ-AS-5 + specs/notice-publish/spec.md REQ-NP-1 | **`notice_attachments.file_type` 列迁移未显式成 Requirement（v1 SF-1 遗留未闭合）**。REQ-AS-5 要求「persist a file_type field on notice_attachments」，REQ-NP-6 要求 CreateNotice 时写 notice_attachments 行并 recording file_type from FileInfo——但现状 DDL（services/community-hub-service/migration/001_initial.sql:28-36）notice_attachments 仅 (id, notice_id, file_name, file_url, file_size, created_at)，**无 file_type 列**；而 REQ-NP-1 的迁移 Requirement 只枚举了两个 ALTER（community_id 去 NOT NULL、published_at 去 NOT NULL），未包含 notice_attachments 加列迁移，也未挂「迁移先于功能上线」门禁 + 异常场景（published_at 迁移在 REQ-NP-1/REQ-NP-MOD-4 有完整正向+异常场景，file_type 加列完全没有）。遗漏该列迁移 → 带附件通知的 notice_attachments INSERT（含 file_type）直接 SQL 失败，附件发布主链路不可用，与 published_at 未迁移导致的发布主链路不可用（D30 已判 MUST FIX）同 class。 | 在 REQ-AS-5（或并入 REQ-NP-1 迁移 Requirement）显式补 1 条迁移契约：「`ALTER TABLE notice_attachments ADD COLUMN file_type VARCHAR(20) NOT NULL DEFAULT ''`（或等价，旧行回读为空，与 REQ-AS-5 场景 2 兼容）」，并挂「迁移先于功能上线」门禁 + 1 个异常场景（列未迁移时带附件发布 INSERT 失败、无部分写入），与 REQ-NP-1 迁移门禁场景同构。 |

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| SF-1 | specs/notice-publish/spec.md REQ-NP-3 网格员多选场景 + REQ-NP-4 division 展开场景 | **直接多选路径与 division 展开路径的「目标小区审核通过」要求不对称（v1 SF-2 遗留未闭合）**。division 路径显式只含审核通过小区（REQ-NP-4：「expansion SHALL include only residential areas whose submission status is approved, encoded by status=1」）；而直接多选路径（grid_worker 的 community_ids，REQ-NP-3 网格员场景 GIVEN：「both communities exist in md_residential_area」）未声明目标小区须审核通过。边界后果：网格员若持有未审核通过小区的 scope grant，可直接发布到该小区，而社区管理员走 division 展开永远不会覆盖该小区——两路径发布范围语义不一致，实现者可据此绕过审核通过过滤。 | 补 1 句产品决策（或场景）：显式声明直接多选路径的目标小区是否同样要求审核通过（若要求则在 AssertPublishScope 前补 approved 过滤；若不要求则说明 division 的 status=1 是展开层过滤、直接多选信任 scope grant），闭合两条路径的边界不对称。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I-1 | REQ-NR-1（ListNotices）：未显式覆盖「软删通知从列表排除」的异常场景（REQ-NP-MOD-1 覆盖未过审排除，软删排除仅隐含在 requirement 文本），建议补 1 句或场景钉死，防实现者遗漏 deleted_at 过滤。 |
| I-2 | REQ-NP-6（附件绑定）：`attachment_ids` 中引用同一文件 id 重复出现的边界未定义（重复 id 是否去重/是否重复写 notice_attachments 行），建议一句话钉死或补场景。 |

## 上轮问题修复校验（r0:rc3 → r1:rc3）

- ✅ v1 MF-1（published_at 迁移去 NOT NULL）：已由 D30 显式化并落入 REQ-NP-1 场景 3 + REQ-NP-MOD-4 场景 5（含异常门禁）。
- ✅ v1 SF-1（notice_attachments.file_type 列迁移）：**未闭合**，升级为本轮 MF-1。
- ✅ v1 SF-2（直接多选 vs division 审核通过不对称）：**未闭合**，保留为本轮 SF-1。
- ✅ v3 S-1（published_at 时序语义）：D27/D30/D32 已锚定审核通过时设置，REQ-NP-MOD-4 + REQ-NR-3 + REQ-NM-1 均以 published_at 为唯一锚。
- ✅ v3 S-2（GetMarqueeNotices community_id 缺失）：REQ-NR-3 已补「community_id 缺失被拒（080005）」场景。
- ✅ v3 S-3（REQ-AS-4 无异常场景）：已补「override 试图放宽 10MB/放行禁止集被拒（不变式不弱化）」异常场景。
- ✅ v3 I-1/I-2（附件总量边界、直接多选审核状态）：附件总量已由 D23/REQ-AS-6 显式化；直接多选审核状态保留为本轮 SF-1。

## 问题跟踪表

| 编号 | 状态 |
|---|---|
| MF-1 | 待修复 |
| SF-1 | 待修复 |

---
VERDICT: REVISION
---
