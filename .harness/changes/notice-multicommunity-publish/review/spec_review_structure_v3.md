# Plan Review — notice-multicommunity-publish（结构合理性视角）

**审查版本**: P1.3（fallback:r2:rc1，修订轮）
**审查维度**: 职责边界、一致性（proposal 影响范围 ↔ specs 职责边界、capability 职责是否清晰无重叠、跨服务契约载体）
**审查方式**: spec 在 r1:rc1 之后已更新（specs/* 与 proposal.md 均于 23:41-23:42 修改，晚于 r1:rc1 评审 23:35-23:36）——按磁盘最新内容独立重新审查，不沿用旧轮结论。

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2

## 上轮问题修复验证（structure_v2：0 MUST / 2 SHOULD / 2 INFO）

| 上轮# | 级别 | 结论 | 证据（磁盘） |
|:---:|:---:|:---:|------|
| 1 | SHOULD | ⚠️ 部分修复 | REQ-NP-6 仍要求 community-hub 校验「已确认 + 归属当前用户」并记录 file_type，但契约载体仍未钉死：file-service 现有 `GetFileUrl(file_id)` 返回 `FileInfo`（含 `user_id`），可支撑存在性+归属校验；但 `FileInfo` 无 `file_type`、无显式 confirmed 状态字段，REQ-AS-5 的「validated extension 回读」无 proto 载体。见本轮 SHOULD #1。 |
| 2 | SHOULD | ❌ 未修复 | REQ-PP-1 仍将 property_admin 列为移动端可发布角色（can_publish=true），REQ-NM-5/REQ-NP-4 定义其移动端发布表单行为；但权限种子明确 property_admin=`pc`（init_permissions.sql:277）且端准入在登录期强制（auth-service platform.go roleAllows），PC 本变更 Q6 不做。sys_admin 已获 D16 显式排除，property_admin 无同类处理。见本轮 SHOULD #2。 |
| 1 | INFO | ✅ 已修复 | file.proto 头注释错误码漂移已识别并纳入 REQ-AS-1 对齐修正（070002 上传失败/070003 文件类型不支持 → 070001 不存在/070002 访问被拒绝/070003 操作失败/070004 类型不支持/070005 大小超限），与磁盘实际常量（70001/70002/70003）一致。 |
| 2 | INFO | ✅ 已修复 | REQ-PP-1 与 REQ-NP-4 均含 merchant 只读；RBAC→NoticeRole 映射已定义。 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 修复建议 |
|---|-------------|------|---------|
| 1 | `specs/notice-publish/spec.md` REQ-NP-6（L159-171）+ `specs/attachment-security/spec.md` REQ-AS-5（L100-112） | 附件校验的 file-service 读取契约载体仍未钉死：REQ-NP-6 要求 community-hub 校验「文件存在 + 已确认 + 归属当前用户」，REQ-AS-5 要求记录「validated extension」到 notice_attachments.file_type。磁盘现状：① `GetFileUrl(file_id)`（file.proto:164-178）可返回 FileInfo（含 user_id），支撑存在性+归属；但 `FileInfo`（file.proto:71-91）无 `file_type` 字段、无显式 confirmed 状态字段——「按 id 校验已确认」与「回读已确认类型的 file_type」无 proto 载体；② notice-publish 服务职责边界（L173-179）仍不列 file-service；③ proposal 影响范围 file-service 行仍只覆盖上传白名单/大小/错误码，无「附件引用校验 + file_type 回读」契约。实现若照此推进，将被迫在 design 阶段自造契约（新增 FileInfo 字段/新 RPC），或退化为从 FileInfo.file_name 扩展名派生（非魔数嗅探结果，弱化 REQ-AS-3 安全意图）。 | 明确 REQ-NP-6/REQ-AS-5 的契约载体：复用 `GetFileUrl(file_id)` 校验存在+归属，并新增 `file_type` 到 FileInfo（或新增按 id 返回含确认状态+file_type 的读取 RPC）；写明 file_type 从该契约回读（非客户端回传）。file-service 加入 REQ-NP-6 服务职责边界清单；proposal 影响范围 file-service 行补「附件引用校验 + file_type 回读」契约。 |
| 2 | `specs/publish-permission/spec.md` REQ-PP-1（L11）+ `specs/notice-mobile/spec.md` REQ-NM-5（L77）+ `specs/notice-publish/spec.md` REQ-NP-4（L113）↔ `services/permission-service/scripts/init_permissions.sql:277` | property_admin 平台覆盖不一致未修复：REQ-PP-1 将 property_admin 列为移动端可发布角色，REQ-NM-5/REQ-NP-4 定义其「范围固定本小区」的移动端发布表单/发布路径；但权限种子明确 property_admin=`pc`（init_permissions.sql:277），端准入登录期强制（auth-service platform.go），PC 端本变更 Q6 不做——property_admin 在移动端无法登录、PC 端不在本变更，其发布路径在本变更内实际不可达。sys_admin 已获 D16 显式排除（can_publish=false），property_admin 无同类处理，spec 角色覆盖与种子平台口径不一致。 | 仿照 D16 为 property_admin 增加显式处置：注明「property_admin 发布面为 PC 管理台，随 Q6 一并顺延；移动端 can_publish=false」，或本变更内将 property_admin 加入 mobile platforms 并同步种子；否则从种子取得的平台口径与 spec 角色覆盖不一致，设计评审将无法判定 421 绑定是否落地移动端。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | `specs/notice-publish/spec.md` REQ-NP-4（L113）+ 服务职责边界（L177） | division 展开经 community-hub 直连 master-data `GetResidentialAreasByDivision` 与既有依赖规则不冲突：`scope.go`「community-hub → permission-service → master-data；不直连 master-data 做 scope 解析」专指 authz 祖先链解析（ResolveScopeAncestors 经 permission 消费），而 GetResidentialAreasByDivision 是向下数据展开（非 authz 解析）；community-hub 已有直连 master-data GetSectionQuota/GetConfig 的先例（servicecontext.go）。建议在边界注释中明确这一区分，避免实现阶段误判为依赖方向冲突。 |
| 2 | `specs/attachment-security/spec.md` REQ-AS-1（L11） | file.proto 头注释对齐为「070001 不存在/070002 访问被拒绝/070003 操作失败/070004 类型不支持/070005 大小超限」后，旧头注释中的「070003 文件类型不支持」标签将语义迁移到 070004——属纯注释修正，无 int32 冲突；建议在 CHANGELOG 中登记该「头注释语义迁移」以提醒依赖该注释的历史文档。 |

## 职责边界总评
- capability 职责无重叠：notice-publish（写）/ notice-read（读）/ publish-permission（入口 RPC + 种子）/ attachment-security（file-service）/ notice-moderation（moderation 复用）/ notice-mobile（前端）分工清晰，交叉引用到位（REQ-NP-3→REQ-PP-3/4、REQ-NP-6→REQ-AS-5、REQ-NR-3→REQ-NP-MOD-1、REQ-NM-5→REQ-NP-4）。
- 跨服务契约关键事实经磁盘验证：GetUserRoles 返回 UserRoleInfo{status, verified_at, expires_at}（permission.proto:353-366）可支撑 level-2；GetDataScopes/GetResidentialAreasByDivision（含 status=1 approved 过滤）/ResolveScopeAncestors 均存在；CreateNoticeRequest 字段 8/9 空闲、GetNoticeRequest 字段 2 空闲、NoticeAttachment 字段 5 空闲，全部走兼容新增，无 wire 破坏；UpdateNoticeModerationStatus 名称与 REQ-NP-MOD-2 一致。
- 种子现状与 REQ-PP-4 吻合：grid_worker（role 4）现无 421 需授，property_admin/community_admin/committee 现持 421，owner/tenant 持 421 需回收，421 min_verf_level 现为 0 需置 2（init_permissions.sql:144/152/171/201-202/252-253）。
- 错误码分层契约正确：080002（功能权限，PermMiddleware 产出）与 080006（数据权限，AssertPublishScope 产出，permission 060007 映射）分层与 scope.go 磁盘实现一致；080003（超限）/080005（参数无效）均已在 types.go 存在（CodeOverLimit=80003/CodeInvalidParam=80005）。
- division→community 授权集落库由 design gate（REV-17）显式承载，spec 如实表述「判据逻辑是否变更待设计评审定夺」，不构成此轮 MUST。
- 无 CRITICAL（无架构违反/安全漏洞/数据丢失/业务不可用一票否决项）；两条 SHOULD 均为契约载体/角色覆盖一致性问题，属设计评审可闭合项。

---
VERDICT: APPROVED
---
