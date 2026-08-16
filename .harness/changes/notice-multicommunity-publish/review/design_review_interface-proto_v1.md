# Design Review — notice-multicommunity-publish（接口契约+Proto 视角）

**审查模式**: 模式一.5 设计评审 · **视角**: interface-proto（接口契约 / Proto 破坏性 / 依赖顺序 / 鉴权-幂等-错误码）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`
**对照基准**: proposal.md + 6 个 spec + 实际 proto 文件（community/file/masterdata/permission）与既有代码核实
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 3 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3

## 已核实（设计正确、与磁盘真相一致的部分）
- Proto 变更全部为**兼容新增**（新字段号 / 新 RPC），`buf breaking-check` 可过：`CreateNoticeRequest.community_ids`(8)/`division_id`(9)、`GetNoticeRequest.community_id`(2)、`NoticeAttachment.file_type`(5)、`FileInfo.file_type`(11)/`confirmed`(12) 均落在当前空闲字段号上（community.proto 现状字段 1-7 已占用，8/9 空闲；NoticeAttachment 1-4 已占用，5 空闲；FileInfo 1-10 已占用，11/12 空闲）。
- `FileInfo.user_id`(2) 已存在并经 `GetFileUrl` 返回 —— D24「经 GetFileUrl 读扩展 FileInfo 校验 confirmed + user_id 归属 + file_type 回读」的载体**成立**（无需新增 RPC，契约自洽）。
- file.proto 头注释漂移判定准确：现状注释「070002 上传失败 / 070003 文件类型不支持 / 070004 文件大小超限 / 070005 bucket 不存在」与实际常量（070002 访问被拒绝 / 070003 操作失败）确实漂移；070004/070005 为新整数位，70003 不重编号（D11 正确）。
- 错误码消歧（D31：目标级解析失败统一 080006、080005 仅请求形状类；D29：community/v1 头注释对齐）语义自洽，与 spec REQ-NP-3/REQ-NP-4 一致。
- division 展开契约经代码核实：masterdata `GetResidentialAreasByDivision` 的 `CommunityDivId>0` 分支 + `status==1` 过滤 `SubmissionStatus!=2`，与 design「community_div_id=division_id, status=1 仅审核通过」一致（getresidentialareasbydivisionlogic.go:33-57）。
- 权限种子核对：421 现仅绑定 (2,420,421)/(3,420,421)/(6,420,421)，grid_worker(4) 确无 421，property_admin(2)/owner(1)/tenant(5) 回收 421 + 421 置 min_verf_level=2 的种子目标与 init_permissions.sql 现状（144/201/252-253 行）一致。
- REST 静态路径 `GET /api/community/notices/marquee` / `publish-permission` 先于 `:id` 注册的路由冲突风险已被 Task 1.17 显式处理。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | tasks.md Task 1.15（+design §CreateNotice 输入） | `CommunityIds []int64 \`json:"community_ids"\`` 无法解析移动端发送的 **string 形式** Snowflake ID。encoding/json 的 `,string` 选项**不支持 slice**（实测 `[]int64 json:",string"` 解 `["1001","1002"]` → `json: cannot unmarshal string into Go struct field ... of type int64`）。Task 4.1 明确移动端 `community_ids: string[]`（防 JS 精度丢失，硬性约束 #3），**CreateNotice 多小区发布主链路在 REST 层即解包失败**。 | REST 层 `CommunityIds` 改为 `[]string`（代理到 RPC 前 strconv 转 int64），或自定义 UnmarshalJSON；不能沿用 `[]int64`。同时校验 Task 1.15 头注释「int64 全 json:",string"」对 repeated 字段的误导性表述。 |
| 2 | tasks.md Task 1.15（+design §接口设计 CreateNotice） | API 层 `CreateNoticeRequest` **缺少 `AttachmentIds` 字段与透传**。经核实现有 `types.go`（30-35 行）与 api create 逻辑（createnoticelogic.go）完全丢弃 `attachment_ids`（proto 有字段 7，REST 不转发）。Task 1.15 只增 CommunityIds/DivisionId/FileType/GetNotice-CommunityId，**未补 AttachmentIds**——则 REQ-NP-6/REQ-AS-6/D23/D24 的附件绑定（本变更头条功能之一）在 web/mobile 发布表单选附件后永远到不了 RPC，静默丢附件。 | Task 1.15 增加 `AttachmentIds []string \`json:"attachment_ids"\`` 并在 api create 逻辑透传到 RPC `attachment_ids`；补对应 api 层测试。 |
| 3 | tasks.md Task 3.3（+design §服务归属 GetPublishPermission 鉴权） | `GET /api/perm/data-scopes` 端点的**鉴权绑定未钉死**且按现状措辞「绑定 422 同款读角色集」**排除发布者**。422 实际只绑 registered_user(9)/owner(1)/tenant(5)（init_permissions.sql 233/258-259 行），而 grid_worker(4)/community_admin(3)/committee(6) ——正是 REQ-NM-5「grid_worker 多选小区（getDataScopes()）」依赖此端点的发布者 —— 不在 422 绑定集内。Task 3.2/3.3 的「或经既有 read 权限放行」是悬空条件。 | Task 3.2 为 `GET:/api/perm/data-scopes` 钉死一个明确权限码（或与 423/424 同批绑全部移动端角色集），Task 3.3 去掉「或/同款」二义措辞，明确 PermMiddleware 归属；验收场景补「grid_worker 调 data-scopes 返回其 community scope_ids」。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 4 | design §Proto / Task 0.1 | `CreateNoticeRequest.role`(4)/`publisher`(5)/`publisher_id`(6) 明确「请求体不信任、服务端从 JWT 派生」，却**未像 community_id(1) 一样标记 deprecated**；且 api 层 createnoticelogic.go 仍转发 `req.Role/req.Publisher`。契约不自治，易误导未来消费方。 | 与 community_id(1) 一并标记 deprecated（机器可读指令，SEE [[go-deprecated-directive-not-test-comment]]）；api 层停止转发 req.Role/req.Publisher，只透传 publisher_id=JWT。 |
| 5 | tasks.md Task 1.6 / design §CreateNotice | `ExpandDivisionCommunities` 未对 `division_id<=0` 设防。masterdata `GetResidentialAreasByDivision` 在 county/street/community_div 全 ≤0 时走 **default 分支 FindAll（最多 1000 小区）而非空**（getresidentialareasbydivisionlogic.go:39-40）。proto3 int64 默认 0 无法区分「未传」与「0」，若空载体校验只判 `>0`，division_id=0 会落入 FindAll 过度展开（叠加 global 数据范围更危险）。 | 在 helper 内先判 `divisionID<=0 → 080005`，杜绝进入 masterdata default 分支；design §CreateNotice 校验顺序第 4 步显式补此 guard（fail-closed 一致）。 |
| 6 | tasks.md Task 0.4 / design §Proto 破坏性评估 | 仅声明「wire 兼容（新字段号）」，未登记**语义破坏**：GetNotice 新增必填 community_id（缺失 → 080005）、legacy CreateNotice 仅传 community_id(1) 不再接受（→ 080005）。已核实 web/pc 无 notice 消费方、web/mobile 在本变更内同步升级，破坏面收敛，但外部 API 消费方（若存在）会被静默破坏。 | Task 0.4 CHANGELOG 显式登记语义破坏项（不仅是 wire 兼容），标注兼容期与回退行为。 |
| 7 | design §UpdateNoticeModerationStatus / Task 1.13 | 编辑后重审（UpdateNotice 重置 moderation_status=0 → 重过审）会**重刷 published_at=NOW()**，把已发布通知重新顶到跑马灯/列表顶部。D27/D30 语义正确，但属可感知行为副作用，设计未在 ADR 显式记录。 | 在 ADR（或 REQ-NP-MOD-3 对应章节）补一句该副作用声明，避免执行/评审阶段再次争议。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 8 | `notice_scope` 仅 `created_at`、无 `updated_at`（时间字段规范 3.1）。纯关联表 + 物理删除场景可接受，建议注释说明以保持一致审计口径。 |
| 9 | file migration 002 `confirmed TINYINT NOT NULL DEFAULT 1` 使**存量文件免 magic-bytes 校验即 confirmed**。D24 信任边界下这是有意的迁移折中，建议在 design §安全考虑注明（存量无 file_type/未嗅探）。 |
| 10 | `FindListByCommunity` JOIN notice_scope 后 `ORDER BY is_pinned DESC, published_at DESC` 依赖 filesort（旧 idx_published 已 deprecated 不服务）。量级小可接受，design 已声明；留作索引兜底观察点。 |

## 问题跟踪表
| 状态 | 说明 |
|------|------|
| 待修复 | MUST 1/2/3 + SHOULD 4/5/6/7（本轮全部） |
| 已修复 / 已验证 | — |

---
VERDICT: REVISION（3 个 MUST FIX）
---

## 结论
设计总体自洽：Proto 变更全部为兼容新增、错误码消歧与头注释对齐判定准确、division 展开与种子变更经磁盘代码核实一致、D24 附件校验载体（扩展 FileInfo 经 GetFileUrl）契约成立。但接口契约层存在 3 个会破坏运行时/发布主链路的 MUST FIX：① REST 层 `CommunityIds []int64` 无法解析 TS string 形式 Snowflake ID（发布主链路解包失败）；② API 层缺 `AttachmentIds` 透传（附件功能经 REST 静默失效）；③ data-scopes 端点鉴权绑定未钉死且默认角色集排除发布者（网格员无法拉取范围选项）。修复后回传架构设计师修订。
