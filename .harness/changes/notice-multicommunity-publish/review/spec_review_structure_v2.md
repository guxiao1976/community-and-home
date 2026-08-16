# Plan Review — notice-multicommunity-publish（结构合理性视角）

**审查版本**: P1.3（fallback:r1:rc3）
**审查维度**: 职责边界、一致性（proposal 影响范围 ↔ specs 职责边界、capability 职责是否清晰无重叠、跨服务契约载体）
**审查方式**: 本轮 proposal.md（00:27）与 specs/notice-publish·notice-read·notice-mobile·notice-moderation（00:28）在 r0:rc3 评审（00:17-00:23）之后更新——哈希与历史轮次不同，按磁盘最新内容独立重新审查，不沿用旧轮结论。磁盘事实均逐一核验（community.proto/file.proto/permission.proto/masterdata.proto、file errcode、community types.go、001_initial.sql）。

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2

## 上轮问题修复验证（structure r0:rc3：0 MUST / 1 SHOULD / 2 INFO）

| 上轮# | 级别 | 结论 | 证据（磁盘） |
|:---:|:---:|:---:|------|
| 1 | SHOULD | ❌ 未修复 | `specs/notice-read/spec.md` L114 api-proto 服务职责边界行仍为「…新增 GetMarqueeNotices RPC 契约（响应保持单 community_id）」——「响应保持单 community_id」是 ListNotices 残留短语，仍误挂在 GetMarqueeNotices 行，与 REQ-NR-3 正文（L69：响应为 NoticeMarqueeItem，notice id + title，无 community_id）矛盾。见本轮 SHOULD #1。 |
| 1 | INFO | ⚠️ 未闭合（显式留待） | REQ-NM-5（L77）division 选项数据源仍以「e.g., GetDivisionTree / 按用户县区过滤」+ design gate（D17/REV-17）显式留待设计评审，「community_admin 可见哪些 division」的过滤/可见规则未钉死——结构上诚实，不构成 MUST/SHOULD。 |
| 2 | INFO | ✅ 保留 | REQ-NP-4（L118）division→community 授权落位如实承载于 design gate（「判据逻辑是否变更待设计评审定夺」+ 验收场景），未预先断言机制。 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/notice-read/spec.md` L114（服务职责边界 api-proto 行） | GetMarqueeNotices 响应形状与正文契约自相矛盾（上轮 r0:rc3 SHOULD #1 未修复，本轮重申）：边界摘要写「新增 GetMarqueeNotices RPC 契约（**响应保持单 community_id**）」，但同文件 REQ-NR-3（L69）明确响应为专用 `NoticeMarqueeItem`（notice id + title 摘要，无 community_id）；「响应保持单 community_id」是 ListNotices 行的语义残留误挂到 GetMarqueeNotices 行。实现者若按边界摘要而非正文需求实现，会误给跑马灯项携带 community_id 或误造响应形状。 | 将该行改为「ListNotices 响应保持单 community_id；新增 GetMarqueeNotices RPC 契约（返回 NoticeMarqueeItem：notice id + title，见 REQ-NR-3）」——把「响应保持单 community_id」挂回 ListNotices 句，GetMarqueeNotices 句改为与正文一致。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | `specs/notice-mobile/spec.md` REQ-NM-5（L77）division 选项数据源 | 「按用户县区过滤的行政区划列表」取值契约与 community_admin 可见 division 的判定规则未钉死，当前以「e.g.」+ design gate 显式留待设计评审。建议设计评审任务将「division 选项树的过滤/可见规则」列为必验收项，避免实现阶段自造规则。 |
| 2 | `specs/notice-publish/spec.md` REQ-NP-4（L118）division→community 授权 | 已核实 permission 数据模型 scope_type 仅 global/community/building/unit/grid（无 community_div），division grant 落位须经 permission-service 单测/集成验收验证（design gate REV-17），spec 未预先断言机制。建议设计评审硬性校验：division grant 在 rel_user_role 的表示方式 + AssertPublishScope 展开判据，未验证不得进入编码（proposal 风险表已列）。 |

## 职责边界总评
- **proposal 影响范围 ↔ specs 职责边界一致**：7 个服务（api-proto/community-hub/file/permission/master-data/moderation/web-mobile）在 proposal 影响范围表与 6 个 spec 的服务职责边界全覆盖且无冲突；master-data/moderation 均标注「只读复用」，web/pc 标注「不做」（Q6），与 .change.yaml services 列表吻合。
- **capability 职责无重叠**：notice-publish（写侧：数据模型/迁移/CreateNotice/DeleteNotice 撤回/附件绑定）、notice-read（读侧：ListNotices/GetNotice/GetMarqueeNotices）、publish-permission（入口 RPC + 种子对齐）、attachment-security（file-service 上传安全 + FileInfo 扩展 + 总量上限契约）、notice-moderation（moderation 复用 + published_at 锚定 + UpdateNotice 边界）、notice-mobile（前端）分工清晰；单一契约源交叉引用闭环（附件总量上限 REQ-AS-6 单一源、published_at 锚定 REQ-NP-MOD-4 单一源、can_publish REQ-PP-1 单一源、division 展开 REQ-NP-3/4 单一源），无重复定义/职责漂移。
- **跨服务契约载体全部经磁盘验证**：CreateNoticeRequest 字段 8/9 空闲（community_ids/division_id）、GetNoticeRequest 字段 2 空闲（community_id 必填）、NoticeAttachment 字段 5 空闲（file_type）、FileInfo 字段 11/12 空闲（file_type/confirmed）——全部走兼容新增字段号，无 wire 破坏；GetUserRoles 返回 UserRoleInfo{status=4/verified_at=5/expires_at=6} 可支撑 level-2 判定；GetResidentialAreasByDivisionReq（community_div_id=3, status=4，0=all/1=approved）与「status=1 仅审核通过」断言一致；UpdateNoticeModerationStatus RPC 存在；file errcode 70001-70003（不存在/访问被拒绝/操作失败）与 REQ-AS-1 断言一致；community types.go 80001-80006 与 080002（功能层）/080006（数据层）分层一致。
- **行为回归登记属实**：DeleteNotice 现用 CheckPublishScope（数据范围删）→ 收窄为仅发布者本人已登记；CreateNotice 现创建时即设 PublishedAt → D27/D30 锚定审核通过已登记；property_admin/owner/tenant 回收 421 + 421 置 min_verf_level=2 已登记（REQ-PP-4 + proposal 风险表）。
- **无 CRITICAL**（无架构违反/安全漏洞/数据丢失/业务不可用一票否决项）；1 条 SHOULD 为 capability 边界摘要与正文契约的措辞不一致（上轮未修复，本轮重申），属实现/设计前易消除项；2 条 INFO 均为已显式 design gate 承载的开放契约提示。

---
VERDICT: APPROVED
---

**summary**: 结构视角 APPROVED——proposal 影响范围与 6 spec 职责边界一致、capability 无重叠、跨服务契约载体全部经磁盘验证无 wire 破坏；唯一遗留 SHOULD（notice-read L114 GetMarqueeNotices 边界行误挂「响应保持单 community_id」）为摘要措辞与正文矛盾，非阻塞，建议在设计评审前消除并同步 BACKLOG。
