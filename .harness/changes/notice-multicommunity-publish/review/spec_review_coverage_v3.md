# Plan Review — notice-multicommunity-publish（覆盖完整性视角）

**审查维度**: 需求覆盖、场景完整性、边界识别
**审查版本**: P1.3（fallback:r2:rc1）— 按磁盘最新内容独立重审（specs 于 23:41-23:42 更新，晚于 v2 评审 23:36）
**审查基线**: `.change.yaml` + `proposal.md` + `specs/*/spec.md`（change 目录无 request.md，以 proposal 为原始需求基准）+ 实际代码/种子核验（init_permissions.sql、createnoticelogic.go、community.proto、masterdata.proto）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 4

总评：r2 版 spec 覆盖完整性已高度收敛。28 个 Requirement 覆盖了 proposal/design 全部决策点（D1-D22），绝大多数具备 ≥1 正向 + ≥1 异常 Scenario；v1/v2 两轮所有 MUST FIX/SHOULD/INFO 均已在磁盘上逐条落实（见下方「上轮问题修复校验」）。本轮未发现 MUST FIX 级覆盖空洞。剩余 3 个 SHOULD FIX 均为边界条件/新契约异常场景的补强，不阻塞阶段 3 设计。

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| S-1 | specs/notice-read/spec.md REQ-NR-3（跑马灯窗口）+ notice-mobile REQ-NM-1/REQ-NM-3 + notice-moderation REQ-NP-MOD-1（published_at 语义） | **`published_at` 与审核流程的时序语义未定义**。跑马灯「最近 15 天窗口」（REQ-NR-3）、列表倒序（REQ-NR-1）、详情「发布时间」（REQ-NM-3）都锚定 published_at，但 spec 未声明 published_at 是「创建时」还是「审核通过时」设置。核验现有实现：createnoticelogic.go:62 在 CreateNotice 即 `PublishedAt: time.Now()`，早于异步审核通过（audit 走 Redis 队列）。边界后果：一条通知若审核滞留 >15 天才通过，通过时 published_at 已超出跑马灯窗口 → 该通知通过审核却不进跑马灯（仅进浏览列表），与「最近 15 天」直觉（从可见日起算）可能相悖；且详情页显示的「发布时间」= 创建时间而非通过时间。此交互边界无任何 Scenario 覆盖。 | 补 1 个决策点 + 场景：显式声明 published_at 设置时机（保持现状=创建时，或改为审核通过时），并补「审核通过时 published_at 已超出 15 天窗口 → 不入跑马灯但入浏览列表」的边界场景，或在 REQ-NR-3 中显式锚定窗口起点语义。 |
| S-2 | specs/notice-read/spec.md REQ-NR-3（新 RPC GetMarqueeNotices 入参校验） | **新 RPC `GetMarqueeNotices(community_id)` 缺少 community_id 缺失/无效的异常场景**。REQ-NR-2 对 GetNotice 显式定义 community_id 必填（缺失 → 080005，D15）；ListNotices 有既有必填语义；而本变更新增的跑马灯 RPC 仅定义正常契约（窗口/排序/封顶/审核通过），未定义空/零 community_id 的处理与错误码。新契约未覆盖入参校验边界。 | 补 1 个异常场景：GetMarqueeNotices 收到空/缺失 community_id → 080005（与 REQ-NR-2/D15 一致），不静默回退默认小区。 |
| S-3 | specs/attachment-security/spec.md REQ-AS-4（全局基线不变式） | **REQ-AS-4 是 28 个 Requirement 中唯一无异常场景者**：3 个场景全为正向（基线默认、覆盖生效、存量不回归）。requirement 文本断言了两个不变式（10MB 硬上限不可放宽、禁止集不可弱化），但无任何异常场景断言「entity_type override 尝试放宽 10MB / 放行 exe → 被拒」。按覆盖视角「每个 Requirement ≥1 正向 + 1 异常」不达标，且不变式恰是未来最易被 override 破坏的点。 | 补 1 个异常场景：entity_type override 试图放宽 10MB 或放行禁止集（exe/sh/zip）→ 被基线拒绝（070004/070005），覆盖不弱化基线的断言。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I-1 | REQ-NP-4（division 展开仅含审核通过小区）与直接多选路径（grid_worker/property 的 community_ids）的「目标小区审核状态」要求不一致：division 路径显式只含审核通过小区，直接多选路径未声明目标须审核通过。若某网格员持有未审核通过小区的 scope grant，可直接发布到该小区而 division 路径不会。建议一句话确认直接路径是否同样要求目标小区已审核通过。 |
| I-2 | 附件总量边界未定义：单文件 ≤10MB（REQ-AS-2）有，但单通知附件总大小/附件数量上限无契约（恶意客户端一次挂大量附件）。建议在 REQ-NP-6 补数量/总量边界或显式声明不限。 |
| I-3 | 撤回（DeleteNotice）时 notice_attachments 的处置未定义：notice 软删 + notice_scope 物理删（REQ-NP-5），附件行/文件对象是否随删或保留未说明。建议一句话确认（文件对象在 file-service，多半保留；行级是否软删需声明）。 |
| I-4 | CreateNotice 幂等性未定义：网络重试/双击提交可能产生重复通知（无幂等键契约）。建议确认是否需幂等，或显式声明「不幂等、重复提交产生重复通知（既有语义）」。 |

## 上轮问题修复校验（v2 → r2）

| # | v2 问题 | 本轮状态 |
|---|---------|---------|
| MF-1 写路径角色状态门槛 | REQ-PP-3/REQ-PP-4 已按方案 A（421 置 min_verf_level=2）落地，含「未认证/已过期仍持数据范围 → 080002」异常场景；种子事实核验（init_permissions.sql §4.2 现状置 0、grid_worker(4) 无 421、owner(1)/tenant(5) 持 421）与 spec 断言一致 | ✅ 已修复 |
| S-1 跑马灯接口 | D12 → REQ-NR-3 新增专用 RPC GetMarqueeNotices（窗口/置顶/封顶/审核通过），REQ-NM-1 同步「不重算窗口」 | ✅ 已修复 |
| S-2 目标数上限 | D13 → REQ-NP-3 ≤100 / 超限 080003，含 division 展开后计数 | ✅ 已修复 |
| S-3 doc 魔数 | D18 → REQ-AS-3 doc=OLE2/CFB(D0CF11E0)、docx=ZIP+OOXML 分别识别，含真实 .doc 放行场景 | ✅ 已修复 |
| S-4 division 角色 | D14 → REQ-NP-3 非 community_admin 传 division_id → 080005，含场景 | ✅ 已修复 |
| I-1..I-5（community_id 缺失、下载侧外置、单事务、role 筛选保留、sys_admin） | 均已落实（REQ-NR-2 080005、out_of_scope 声明、REQ-NP-1/3 单事务、REQ-NR-1 role 保留、REQ-PP-1 sys_admin 场景） | ✅ 已修复 |

## 问题跟踪表

| # | 状态 |
|---|------|
| S-1 published_at 时序 | 新增（SHOULD FIX） |
| S-2 GetMarqueeNotices 入参校验 | 新增（SHOULD FIX） |
| S-3 REQ-AS-4 异常场景 | 新增（SHOULD FIX） |
| I-1..I-4 | 新增（INFO） |

---
VERDICT: APPROVED
---
