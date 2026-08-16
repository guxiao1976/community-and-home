# Plan Review — notice-multicommunity-publish（结构合理性视角）

**审查版本**: P1.3（fallback:r0:rc3）
**审查维度**: 职责边界、一致性（proposal 影响范围 ↔ specs 职责边界、capability 职责是否清晰无重叠、跨服务契约载体）
**审查方式**: 本轮 spec 全部 6 文件与 proposal/.change.yaml 均在 08-16 00:09-00:13 更新（晚于 structure_v3 评审 08-15 23:51）——按磁盘最新内容独立重新审查，不沿用旧轮结论。

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2

## 上轮问题修复验证（structure_v3 r2:rc1：0 MUST / 2 SHOULD / 2 INFO）

| 上轮# | 级别 | 结论 | 证据（磁盘） |
|:---:|:---:|:---:|------|
| 1 | SHOULD | ✅ 已修复 | FileInfo 契约载体已钉死：REQ-AS-7（attachment-security）显式「FileInfo 新增 file_type(11) + confirmed(12)，经既有 GetFileUrl 返回」，D24 落 proposal；已核实 FileInfo 现字段 1-10，11/12 空闲。REQ-NP-6 的服务职责边界已补 file-service 为「附件引用校验的数据提供方」。 |
| 2 | SHOULD | ✅ 已修复 | D26 显式剔除 property_admin 移动端可发布角色：REQ-PP-1/REQ-PP-4/REQ-NP-4/REQ-NM-4/REQ-NM-5 全文一致（can_publish=false + 421 回收）。已核实种子 platforms（init_permissions.sql:277 property_admin='pc'，grid_worker/committee='mobile'）——D26 的移动端剔除与平台口径自洽；421 回收对 PC 的影响登记为行为回归且 PC 本变更不做（Q6）。 |
| 1 | INFO | ✅ 已修复 | file.proto 头注释语义迁移已登记 REQ-AS-1（D11）+ community 头注释对齐登记 D29 + 均要求 CHANGELOG 登记。 |
| 2 | INFO | ✅ 已修复 | REQ-PP-1 与 REQ-NP-4 均含 merchant 只读；RBAC→NoticeRole 映射已定义；proposal 影响范围已补 file-service 附件校验契约。 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/notice-read/spec.md` L109（服务职责边界 api-proto 行） | 边界摘要与正文契约自相矛盾：服务职责边界写「新增 GetMarqueeNotices RPC 契约（**响应保持单 community_id**）」；但同文件 REQ-NR-3（L69）明确响应为专用 `NoticeMarqueeItem`（notice id + title 摘要，无 community_id）——「响应保持单 community_id」是 ListNotices 行的残留短语误挂到 GetMarqueeNotices 行。实现者若按边界摘要而非正文需求实现，会误给跑马灯项携带 community_id。 | 删除该行「（响应保持单 community_id）」或改为「（返回 NoticeMarqueeItem：notice id + title，见 REQ-NR-3）」，与正文契约一致。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | `specs/notice-mobile/spec.md` REQ-NM-5（L77）division 选项数据源 | 明确「SHALL 源自 master-data division 树（e.g., GetDivisionTree / 按用户县区过滤的行政区划列表）」——`GetDivisionTree` 已核实存在，但「过滤到用户县区」的取值契约未钉死，且 community_admin 可见哪些 division 的判定规则未定义。当前以「e.g.」+ design gate（D17）显式留待设计评审，结构上诚实；建议在设计评审任务中把「division 选项树的过滤/可见规则」列为必验收项，避免实现阶段自造规则。 |
| 2 | `specs/notice-publish/spec.md` REQ-NP-4（L113）+ `services/permission-service/rpc/internal/logic/permission/assertpublishscopelogic.go` | 已核实 permission 数据模型 scope_type 仅 global/community/building/unit/grid（rel.go:77-82），AssertPublishScope 仅 resolveUserScope(ScopeTypeCommunity)（assertpublishscopelogic.go:38）——`community_div` scope_type 当前不存在。REQ-NP-4 的「division→community 授权落位」已如实承载于 design gate REV-17（"判据逻辑是否变更待设计评审定夺"+ 验收场景），结构上不预先断言机制，OK。建议设计评审硬性校验：division grant 在 rel_user_role 的表示方式 + AssertPublishScope 展开判据，未验证不得进入编码（proposal 风险表中已列，勿丢）。 |

## 职责边界总评
- **capability 职责无重叠**：notice-publish（写）/ notice-read（读）/ publish-permission（入口 RPC + 种子对齐）/ attachment-security（file-service 上传安全 + FileInfo 扩展）/ notice-moderation（moderation 复用 + published_at 锚定）/ notice-mobile（前端）分工清晰；交叉引用闭环（REQ-NP-3→REQ-PP-3/4、REQ-NP-6→REQ-AS-5/6、REQ-NR-3→REQ-NP-MOD-4、REQ-NM-5→REQ-NP-4、REQ-NM-6→REQ-AS-6）。
- **proposal 影响范围 ↔ specs 职责边界一致**：7 个服务（api-proto/community-hub/file/permission/master-data/moderation/web-mobile）在 proposal 影响范围表与 6 个 spec 的服务职责边界全覆盖且无冲突；master-data/moderation 均为「只读复用」，职责边界如实标注。
- **跨服务契约载体全部经磁盘验证**：CreateNoticeRequest 字段 8/9 空闲（community_ids/division_id）、GetNoticeRequest 字段 2 空闲（community_id 必填）、NoticeAttachment 字段 5 空闲（file_type）、FileInfo 字段 11/12 空闲（file_type/confirmed）——全部走兼容新增字段号，无 wire 破坏；GetUserRoles 返回 status/verified_at/expires_at 可支撑 level-2 判定；GetDataScopes 仅 community/building/unit/grid（REQ-NM-5 的「无 community_div」断言准确）；GetResidentialAreasByDivision/ResolveScopeAncestors/GetDivisionTree 均存在；UpdateNoticeModerationStatus RPC 与 REQ-NP-MOD-2 一致。
- **错误码分层契约正确**：file 70001-70003 常量与 D11 断言一致（70002=访问被拒绝/70003=操作失败，非头注释旧语义）；community 80001-80007 与 D29 对齐断言一致（80003=CodeOverLimit 通用超限、80007=CodeSectionQuotaExceeded 寻失配额实际码）；080002（功能层）/080006（数据层，80006 已新增）分层与 scope.go 一致。
- **行为回归登记属实**：DeleteNotice 现用 CheckPublishScope（数据范围删），收窄为仅发布者本人已登记；CreateNotice 现创建时即设 PublishedAt（createnoticelogic.go:62），D27 锚定审核通过已登记。
- **种子变更与 REQ-PP-4 全符合**：grid_worker（role 4）现无 421 需授、property_admin（2）/owner（1）/tenant（5）现持 421 需回收、community_admin（3）/committee（6）保留 421、421 min_verf_level 现 0 置 2（init_permissions.sql:144/152/171/201-202/251-253）。
- **无 CRITICAL**（无架构违反/安全漏洞/数据丢失/业务不可用一票否决项）；1 条 SHOULD 为 capability 边界摘要与正文契约的措辞不一致，属实现前易消除项；2 条 INFO 均为已显式 design gate 承载的开放契约提示。

---
VERDICT: APPROVED
---
