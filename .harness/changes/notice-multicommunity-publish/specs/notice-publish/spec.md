# Notice Publish Capability Specification

## Purpose

将通知发布从「单小区」升级为「多小区」：一次发布可覆盖多个小区（网格员多小区、社区管理员选社区展开快照、**业委本小区**；**property_admin 移动端发布角色本变更剔除（D26），收敛为 grid_worker/community_admin/committee**），范围关联走独立 notice_scope 表（单源），发布落库前经 AssertPublishScope 校验发布者数据范围，撤回全局生效（**附件行 + MinIO 对象全部保留，仅 notices 软删 + notice_scope 物理删，D28**）。本 capability 覆盖 community-hub-service 的发布侧数据模型、迁移与行为契约，并定义 CreateNoticeRequest 的字段演化与权限信任边界（role/publisher_id 以 JWT 实际身份为准）。**division_id 值域与展开契约已定义**（REVISION-2 闭环：int64，值域 `md_administrative_division.id`，经 masterdata GetResidentialAreasByDivision(community_div_id, status=1) 展开）；**单次发布目标数上限、division 载体角色限制、单事务原子性、附件绑定载体（扩展 FileInfo 经 GetFileUrl，D24）、单通知总量上限（≤10 个/≤50MB，D23）、后端不幂等（D25）、published_at 审核通过时设置（D27）+ published_at 去 NOT NULL 迁移（D30）、目标级解析失败统一 080006 错误码消歧（D31）、community/v1 头注释错误码对齐（D29）**均作为显式契约（REVISION 后新增决策 D13/D14/D23-D31）。

## Requirements

### Requirement: REQ-NP-1 — 多小区范围单源 + community_id 去 NOT NULL 迁移

The system SHALL treat `notice_scope` as the single source of truth for notice-to-community associations. For every newly created notice, the `notices.community_id` column SHALL NOT be populated or used for scope filtering; it SHALL be retained during the compatibility period only for schema stability and SHALL NOT be written. The system SHALL apply a schema migration that makes `notices.community_id` nullable (`ALTER TABLE notices MODIFY community_id BIGINT DEFAULT NULL`), because the column is currently `BIGINT NOT NULL` and unset writes would otherwise fail. **The system SHALL apply a second migration in the same schema change that makes `notices.published_at` nullable (`ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL`, D30)**: the column is currently `DATETIME NOT NULL` (migration/001_initial.sql:19), and per D27 the notice is created with `published_at` unset (NULL) and only set by the moderation-pass callback — without this migration the create-time INSERT (which must leave published_at NULL) violates the NOT NULL constraint and the publish main path is unusable. **The system SHALL apply a third migration in the same schema change that adds the `notice_attachments.file_type` column (`ALTER TABLE notice_attachments ADD COLUMN file_type VARCHAR(20) DEFAULT NULL`, REVISION-7)**: the column is currently absent from the DDL (migration/001_initial.sql:28-36), yet REQ-AS-5 requires persisting the whitelist-validated file type and REQ-NP-6 requires CreateNotice to record file_type from FileInfo into notice_attachments — without this migration the notice_attachments INSERT (which includes file_type) fails and the attachment-publish main path is unusable. **The system SHALL apply a fourth migration in the same schema change that adds the `notice_attachments.file_id` column (`ALTER TABLE notice_attachments ADD COLUMN file_id BIGINT DEFAULT 0`, REVISION-8/data-model v3 M1)**: the `file_id` is the authoritative carrier for regenerating short-lived presigned URLs on the read path (see REQ-NR-2/REQ-NM-3) — CreateNotice records the attachment_ids (file_id) into notice_attachments, and GetNotice regenerates the presigned URL via `GetFileUrl(file_id)`; without this column the detail page's attachment download of older notices fails once the stored presigned URL expires (3600s~7 days), violating REQ-NM-3. Legacy rows default file_id=0; GetNotice falls back to the stored file_url for file_id=0/NULL rows (compatibility, consistent with Q1). All four migrations SHALL be deployed before the new publish feature goes live. The notice row, its notice_scope rows and its notice_attachments rows SHALL be persisted atomically in a single transaction (all-or-nothing; no partial write on any failure).

#### Scenario: 迁移后多小区发布写入成功（community_id 与 published_at 均置空）
- **GIVEN** a publisher holding a role whose data scope covers communities C1 and C2, and both the `notices.community_id` and `notices.published_at` columns have been made nullable by the migration
- **WHEN** the publisher creates a notice targeting C1 and C2
- **THEN** the system inserts one `notices` row with `community_id` left unset (NULL) and `published_at` left unset (NULL, set later by the approval callback per D27/D30), plus two `notice_scope` rows (notice_id, C1) and (notice_id, C2); no duplicate scope representation is written

#### Scenario: 迁移未执行则发布被拒（迁移先于上线门禁）
- **GIVEN** the new publish feature is enabled but the `ALTER TABLE notices MODIFY community_id BIGINT DEFAULT NULL` migration has not been applied (column still `NOT NULL`)
- **WHEN** a publisher creates a multi-community notice (which must leave community_id unset)
- **THEN** the insert violates the NOT NULL constraint and the publish fails; no partial notice or scope row is persisted, and the failure surfaces a dependency/rollout error that requires the migration to be run before the feature ships

#### Scenario: published_at 未迁移则创建被拒（迁移先于上线门禁，D30）
- **GIVEN** the new publish feature is enabled but the `ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL` migration has not been applied (column still `NOT NULL`)
- **WHEN** a publisher creates a notice (which per D27 must leave published_at NULL while pending)
- **THEN** the create-time INSERT violates the published_at NOT NULL constraint and the publish fails; no partial notice or scope row is persisted, and the failure surfaces a dependency/rollout error requiring the published_at migration to be run before the feature ships

#### Scenario: notice_attachments.file_type 未迁移则附件发布被拒（迁移先于上线门禁，REVISION-7）
- **GIVEN** the new publish feature is enabled but the `ALTER TABLE notice_attachments ADD COLUMN file_type` migration has not been applied (column still absent from the DDL)
- **WHEN** a publisher creates a notice with attachments (which per REQ-NP-6 must record file_type into notice_attachments)
- **THEN** the notice_attachments INSERT (including the file_type column) fails with an unknown-column error; no partial notice/scope/attachment row is persisted, and the failure surfaces a dependency/rollout error requiring the file_type migration to be run before the feature ships

#### Scenario: 存量旧数据不可见（Q1 已拍板接受）
- **GIVEN** a legacy `notices` row created before this change that has `community_id` set but no `notice_scope` row
- **WHEN** a user lists notices for that community
- **THEN** the legacy notice is not returned (scope 查询走 notice_scope，存量未迁移，旧通知暂不可见；迁移回填留后续迭代并挂 BACKLOG)

#### Scenario: 既有 community_id 索引的兼容处置（含 pending published_at=NULL 的排序表现）
- **GIVEN** the `notices` table has `idx_community (community_id, deleted_at)` and `idx_published (community_id, published_at DESC, deleted_at)` built on the deprecated column; pending notices carry `published_at = NULL` (D30)
- **WHEN** the schema migration runs and the new read path (notice_scope JOIN) becomes active
- **THEN** the two legacy indexes are retained during the compatibility period (to avoid breaking any mid-rollout code) and marked deprecated; the new read path is served by the `notice_scope` read index (see REQ-NP-2), and the deprecated indexes MAY be dropped in a later cleanup migration. **Pending rows (published_at=NULL) never surface in list/marquee DESC ordering or the 15-day window because they are filtered out by the moderation-visibility gate (REQ-NP-MOD-1) before ordering — only moderation-passed rows (which always carry a set published_at) are ordered, so NULL does not perturb the published_at DESC ordering.**

### Requirement: REQ-NP-2 — notice_scope 关联表唯一约束与读路径索引

The system SHALL maintain a `notice_scope` table with `notice_id BIGINT NOT NULL` and `community_id BIGINT NOT NULL` (referencing `md_residential_area.id`, which represents a community or village) — **both columns SHALL be declared NOT NULL** (empty or zero-valued rows are rejected at the schema level) — and SHALL enforce a UNIQUE constraint `uk_notice_community(notice_id, community_id)`. One notice MAY associate with N distinct communities. The table SHALL also carry a read-path index whose leftmost column is `community_id` (e.g., `idx_scope_community (community_id, notice_id)`), because ListNotices and the marquee read path filter by `community_id` before joining `notices`; a unique index that is leftmost on `notice_id` does not serve those queries.

#### Scenario: 同一小区只写入一条关联
- **GIVEN** a CreateNotice request whose `community_ids` contains C1 twice
- **WHEN** the system persists the notice scopes
- **THEN** the system either deduplicates the target list before insert or relies on `uk_notice_community` to prevent duplicate rows; the notice ends up with exactly one scope row for C1

#### Scenario: notice_id 缺失被拒
- **GIVEN** a notice_scope row being persisted with a valid community_id but a zero/unset notice_id
- **WHEN** the system writes the notice_scope row
- **THEN** the insert is rejected by the NOT NULL constraint on notice_id; no orphan scope row is persisted

#### Scenario: community_id 缺失被拒
- **GIVEN** a notice_scope row being persisted with a valid notice_id but a zero/unset community_id
- **WHEN** the system writes the notice_scope row
- **THEN** the insert is rejected by the NOT NULL constraint on community_id; no scope row without a target community is persisted

#### Scenario: 读路径按 community_id 走索引（列表/跑马灯）
- **GIVEN** a community with a large number of notice_scope rows, and `idx_scope_community (community_id, notice_id)` present
- **WHEN** ListNotices or the marquee data query filters scopes by `community_id = ?`
- **THEN** the query is served by the `community_id`-leftmost index rather than scanning the whole notice_scope table; the read is not dependent on the deprecated notices.community_id index

### Requirement: REQ-NP-3 — CreateNotice 多小区发布契约（字段演化 + 权限校验 + 信任边界 + 边界约束）

The CreateNotice RPC SHALL accept a `repeated community_ids` field (one or more) and SHALL retain the existing `attachment_ids` field. The `community_ids` elements SHALL be `int64` with `[jstype=JS_STRING]` (community ids are Snowflake integers; without jstype the TS client loses precision — hard constraint #3). For division-based publishing, the RPC SHALL also accept a `division_id` field (NEW field number 9, `int64` with `[jstype=JS_STRING]`) whose **value domain SHALL be `md_administrative_division.id`** — the same source as `md_residential_area.community_div_id` (REVISION-2). **The system SHALL expand a division by invoking master-data-service `GetResidentialAreasByDivision` with `community_div_id = division_id` AND `status = 1` (approved only)** (existing RPC contract, `community_div_id > 0` branch) and use the returned residential-area ids as the target community set for `AssertPublishScope` and for the notice_scope snapshot. `division_id` and `community_ids` are mutually exclusive (at most one SHALL be set). Field evolution SHALL be non-breaking: the new `repeated int64 community_ids` SHALL use a NEW field number (8), while the legacy single `community_id` field (field 1) SHALL be kept and marked deprecated — reusing field 1 with a changed type would break the wire format and fail `buf breaking-check`. **The single publish SHALL NOT exceed 100 target communities** (the length of `community_ids`, or the count of communities after division expansion); an exceeding request SHALL be rejected with 080003 超限. **`division_id` SHALL be usable only by community_admin** (D14); any other role submitting `division_id` SHALL be rejected with 080005 参数无效. The server SHALL treat `community_ids`/`division_id` as authoritative and SHALL reject (080005 参数无效) a request where both are empty or both are set; it SHALL NOT fall back to the deprecated `community_id`. **The community/v1 proto header error-code block SHALL be aligned to these actual semantics (D29): 080003 = 单次发布目标数超限, 080005 = 参数无效（含小区ID无效）; the stale "080003 寻失发布次数已达上限" header comment SHALL be removed (that semantics is actually 080007, CodeSectionQuotaExceeded) and the block registered in the community CHANGELOG — mirroring the file/v1 header alignment in REQ-AS-1 (D11).** The system SHALL validate every target community id via `AssertPublishScope(user_id, targets)` before any persistence; if any target is outside the publisher's data scope, the entire request SHALL be rejected with 080006 (数据权限) and no partial write. **Error-code distinction for frontend handling (single authoritative mapping, D31): 目标级解析失败（target-level resolution failure）统一 080006 数据权限 — a `community_ids` element that does not exist or cannot be resolved in `md_residential_area` (e.g., community_ids=[99999]) is treated exactly like an out-of-scope target (fail-closed security-reject of unknown nodes; identical to the REQ-NP-3 目标小区不存在 scenario and consistent with REQ-NP-4's post-expansion out-of-scope target → 080006). 080005 参数无效 is reserved for request-shape errors only: empty scope carrier, both `community_ids` and `division_id` set (dual-carrier), non-community_admin submitting `division_id`, an invalid `division_id` value (a `division_id` that does not parse/resolve as a legal `md_administrative_division.id` — this is a request-shape error, distinct from the target-resolution 080006 for a nonexistent community), and attachment reference invalid/over-limit. 080002 功能权限 = no publish role at the function layer (before scope validation).** The notice, scope rows and attachment rows SHALL be persisted atomically in a single transaction. The `role` and `publisher_id` fields SHALL NOT be trusted from the request body: `publisher_id` SHALL be derived from the authenticated JWT identity, and `role` SHALL be derived from the publisher's actual held roles via the RBAC→NoticeRole mapping (see REQ-NP-4). The write-path role-status gate (publisher must be level-2 verified) SHALL be enforced as specified in REQ-PP-3/REQ-PP-4. **CreateNotice SHALL be non-idempotent (D25): no idempotency key is introduced; a duplicate submission (network retry / double click) SHALL produce a duplicate notice; duplicate suppression is the responsibility of the client (submit-disabled while in flight, see REQ-NM-5), not the server.**

#### Scenario: 网格员多小区发布成功
- **GIVEN** a grid_worker whose data scope covers C1 and C2, both communities exist in `md_residential_area`, and the user is level-2 verified (seed 已授 421 + min_verf_level=2, see REQ-PP-4)
- **WHEN** the grid_worker submits CreateNotice with community_ids=[C1, C2], title, content, and attachment_ids
- **THEN** the notice is persisted with `community_id` unset, `notice_scope` records C1 and C2 (single transaction), and the response returns the new notice id

#### Scenario: 空范围请求被拒绝
- **GIVEN** an authenticated publisher
- **WHEN** the publisher submits CreateNotice with an empty `community_ids` list
- **THEN** the system rejects the request with a parameter-invalid error (080005 小区ID无效), and no notice is created (no implicit fallback to the deprecated single `community_id`)

#### Scenario: division_id 与 community_ids 互斥
- **GIVEN** an authenticated publisher who submits CreateNotice with both `community_ids=[C1]` and `division_id=D1` set
- **WHEN** the system validates the request
- **THEN** the system rejects the request with 080005 参数无效 (both scope carriers set is ambiguous) and no notice is created

#### Scenario: 目标小区数量超限（080003）
- **GIVEN** an authenticated publisher submitting community_ids with more than 100 entries, or a division whose expansion yields more than 100 communities
- **WHEN** the system validates the request
- **THEN** the system rejects with 080003 超限 (target count exceeds the 100-cap) and no notice is created

#### Scenario: 非 community_admin 提交 division_id 被拒
- **GIVEN** an authenticated grid_worker or committee submitting CreateNotice with `division_id` set (division carrier is reserved for community_admin, D14)
- **WHEN** the system validates the request
- **THEN** the system rejects with 080005 参数无效 (division_id only usable by community_admin) and no notice is created

#### Scenario: 无发布角色被拒绝（功能权限 080002）
- **GIVEN** an authenticated user holding no publish-capable role in an active state (e.g., owner/tenant after the 421 revocation in REQ-PP-4, or an unverified publisher failing the 421 min_verf_level=2 gate)
- **WHEN** the user invokes CreateNotice (bypassing the UI)
- **THEN** the system rejects the request at the functional-permission layer with 080002 无发布权限 (PermMiddleware / role check); the request never reaches data-scope validation

#### Scenario: 任一目标小区越权导致整体拒绝（数据权限 080006）
- **GIVEN** a committee (or grid_worker) whose data scope covers only C1
- **WHEN** the user submits CreateNotice with community_ids=[C1, C2] where C2 is outside scope
- **THEN** `AssertPublishScope` returns denied and the entire request is rejected with 080006 目标小区超出发布者数据范围 (data-permission denial, mapped from permission 060007); no notice or scope row is written for C1 either (all-or-nothing)

#### Scenario: 目标小区不存在（安全拒绝未知节点，目标级解析失败统一 080006，D31）
- **GIVEN** a publisher submitting community_ids=[99999] where 99999 does not exist in `md_residential_area`
- **WHEN** the system resolves the target's scope ancestors
- **THEN** the unknown node is treated as a target-level resolution failure and the request is rejected with 080006 (fail-closed, identical to an out-of-scope target; 安全拒绝未知节点，不静默创建无效小区通知) — this is the authoritative mapping for nonexistent community targets, distinct from 080005 which covers only request-shape errors (see REQ-NP-3 error-code distinction, D31)

#### Scenario: 客户端伪造 role / publisher_id 被纠正（JWT 为准）
- **GIVEN** an authenticated grid_worker who submits CreateNotice with `role` set to a community_admin NoticeRole value and a `publisher_id` of another user
- **WHEN** the system processes the create
- **THEN** the persisted notice's `publisher_id` is the authenticated JWT user id, and `role` is derived from the user's actual held roles (NOTICE_ROLE_GRID_OFFICER), not from the forged request body

#### Scenario: 旧 community_id 字段不再生效
- **GIVEN** a legacy client that still sends the deprecated single `community_id` (field 1) without the new `community_ids`
- **WHEN** the server receives the request
- **THEN** the server rejects the request with 080005 (community_ids empty, no fallback to the deprecated field); the deprecated field does not create or scope any notice

### Requirement: REQ-NP-4 — 各角色发布范围与 RBAC→NoticeRole 映射

The system SHALL enforce per-role publish scope as follows: grid_worker MAY publish to multiple communities all within their `community` scope; community_admin MAY select a division (`community_div` scope) which the backend expands into concrete community ids snapshot at publish time via master-data `GetResidentialAreasByDivision(community_div_id=division_id, status=1)`; **the expansion SHALL include only residential areas whose submission status is approved (审核通过, encoded by the `status=1` parameter)**; committee SHALL publish only to its own single community; **property_admin is NOT a mobile publish role (D26) — the mobile publish-capable set is converged to {grid_worker, community_admin, committee}**; property_admin SHALL NOT receive a publish entry (`can_publish=false`) and SHALL NOT pass the CreateNotice write path (its 421 is revoked, see REQ-PP-4); owner/tenant/merchant SHALL be read-only and MUST NOT have a publish entry. The scope unit SHALL always be `md_residential_area.id`; `community_type` (小区 vs 村) SHALL NOT alter scope representation. **Whether the division grant (community_div scope) falls into the community authorization set used by AssertPublishScope requires permission-service judging-logic change is a design-gated decision (D17/REV-17): the behavioral contract is that a division-selected publish expands to concrete community ids and each target SHALL pass AssertPublishScope; how a division-scope grant authorizes community targets SHALL be verified by permission-service unit/integration tests in the design review before coding.** The `role` recorded on a notice SHALL be derived from the publisher's actual held role via this mapping for the mobile publish-capable roles: grid_worker→`NOTICE_ROLE_GRID_OFFICER`, community_admin→`NOTICE_ROLE_COMMUNITY`, committee→`NOTICE_ROLE_COMMITTEE`; **property_admin→`NOTICE_ROLE_PROPERTY` is NOT an active mobile mapping in this change (D26) — the enum value is retained for schema/legacy stability but is not reachable from the mobile publish path**; the request body `role` SHALL NOT be authoritative (see REQ-NP-3).

#### Scenario: 社区管理员选社区展开为小区快照（division→community 授权经 design gate 验证）
- **GIVEN** a community_admin whose data scope is a division grant over D1 (containing communities C1, C2, all approved), and the division→community authorization resolution (design-gated, see REV-17) yields C1, C2 as the admin's publishable community set
- **WHEN** the community_admin submits CreateNotice with a division selection referencing D1 (`division_id=D1`)
- **THEN** the backend expands D1 to its concrete approved communities via GetResidentialAreasByDivision, expands the targets to [C1, C2], validates each via `AssertPublishScope` (targets of scope_type=community), and writes `notice_scope` rows for C1 and C2 as a fixed snapshot (snapshot fixed at publish time; later division membership changes do not affect the published notice)

#### Scenario: 展开后无小区则拒绝（division 载体合法但展开为空 → 080005，与 D31 边界一致）
- **GIVEN** a community_admin selecting a division that currently contains zero residential areas (a legal `md_administrative_division.id` whose expansion yields no approved communities)
- **WHEN** the community_admin submits the notice
- **THEN** the system rejects the request with 080005 参数无效 (no valid targets after division expansion — a request-shape/degenerate-expansion error on a legal division carrier, per D31; distinct from a nonexistent `community_ids` element, which is a target-level resolution failure → 080006) and creates nothing

#### Scenario: 展开仅含审核通过小区
- **GIVEN** a division D1 whose residential areas include C1 (approved) and C2 (pending submission), and a community_admin selecting D1
- **WHEN** the community_admin submits the notice
- **THEN** the expansion includes only approved C1 as a target; C2 (not yet approved) is not published to

#### Scenario: 展开后目标小区越权导致整体拒绝
- **GIVEN** a community_admin whose division grant covers D1 (C1, C2) and the division selection resolves to targets including a community C3 outside the resolved authorization set
- **WHEN** the community_admin submits the notice
- **THEN** `AssertPublishScope` denies the out-of-scope target and the entire request is rejected with 080006; no partial snapshot is written

#### Scenario: 业委仅本小区（D26 后 property_admin 不再发布）
- **GIVEN** a committee whose community is C1
- **WHEN** the committee opens the publish form and submits
- **THEN** the scope is fixed to C1 (`community_ids=[C1]`); the selection is not editable and a CreateNotice targeting C2 is rejected with 080006

#### Scenario: property_admin 无移动端发布能力（D26）
- **GIVEN** an authenticated property_admin (role 2, whose 421 has been revoked per REQ-PP-4)
- **WHEN** the property_admin calls GetPublishPermission or directly invokes CreateNotice on the mobile surface
- **THEN** GetPublishPermission returns can_publish=false (property_admin not in the mobile publish-capable set), the mobile「发布通知」entry is hidden, and a direct CreateNotice attempt SHALL be rejected with 080002 (no publish role at the function layer)

#### Scenario: 业主无发布能力
- **GIVEN** an owner/tenant/merchant authenticated user
- **WHEN** the user invokes CreateNotice or checks publish entry
- **THEN** the system SHALL NOT expose a publish entry (`can_publish=false`, see publish-permission capability) and a direct CreateNotice attempt SHALL be rejected with 080002 for lack of publish role (function-permission layer)

### Requirement: REQ-NP-5 — 撤回复用 DeleteNotice（仅发布者本人，全局生效）

The system SHALL treat `DeleteNotice` as the withdraw operation. Deleting a notice SHALL take effect globally across all its associated communities in one operation; the notice's `notice_scope` rows SHALL be physically removed together with the notice soft delete (per design doc: 撤回 = 删 notices 行 + notice_scope 按 notice_id 删). **On withdraw, all attachments SHALL be retained (D28): the `notice_attachments` rows are NOT deleted, and the file objects in file-service (MinIO) are NOT deleted — only the notices row is soft-deleted and its notice_scope rows physically removed.** Only the notice publisher (the user whose JWT id equals `publisher_id`) MAY withdraw it; there is no separate "authorized role" deletion privilege in this change. **This author-only narrowing is a registered behavior regression (validity REVISION-6 SHOULD-3): the existing `DeleteNotice` uses CheckPublishScope (data-scope — a community_admin/property_admin could delete notices within their jurisdiction); this change narrows deletion to the publisher only, so in-jurisdiction admins lose the ability to delete others' notices.**

#### Scenario: 发布者撤回全局生效（附件保留，D28）
- **GIVEN** a notice associated with C1 and C2 with two attachment rows (notice_attachments) and two file objects, published by user U
- **WHEN** U invokes DeleteNotice for that notice
- **THEN** the notice is soft-deleted, its `notice_scope` rows are physically removed, the notice no longer appears in either C1 or C2's notice list, and its `notice_attachments` rows and file-service objects remain (not deleted)

#### Scenario: 非发布者撤回被拒绝
- **GIVEN** a notice published by U, and another user V who is not the publisher (even if V is an in-jurisdiction community_admin)
- **WHEN** V invokes DeleteNotice for that notice
- **THEN** the system rejects the request with 080002 无权限 (author-identity check; the data-scope-based delete privilege is revoked by this change) and the notice (and its scopes) remain unchanged

### Requirement: REQ-NP-6 — 附件绑定校验（attachment_ids 引用已确认文件）

The system SHALL only bind attachment ids that reference files confirmed in file-service and owned by the publishing user. On CreateNotice, for each `attachment_ids` entry the system SHALL verify the referenced file exists, is confirmed, and belongs to the authenticated user; otherwise the request SHALL be rejected with **080005 参数无效（附件引用无效）** and no notice created. **The verification carrier SHALL be the extended `FileInfo` (D24): community-hub-service SHALL invoke file-service `GetFileUrl(file_id)` per attachment id (no new file-service RPC) and confirm the returned FileInfo has `confirmed == true` and `user_id == authenticated user`; the file_type is read back from `FileInfo.file_type` (see REQ-AS-5).** The system SHALL also enforce the aggregate caps (D23): `attachment_ids` count ≤10 AND the sum of the bound files' `FileInfo.file_size` ≤50MB; exceeding either SHALL be rejected with 080005 参数无效（附件超限）and no notice created (see REQ-AS-6). The `notice_attachments` rows SHALL be written by community-hub-service at CreateNotice time (recording `file_type` from the validated FileInfo and `file_id` from the attachment_ids / file_id carrier, see REQ-AS-5 and REQ-NP-1 migration-4), within the same transaction as the notice and scopes.

#### Scenario: 有效附件绑定成功
- **GIVEN** a user who confirmed two files (a1, a2) via the file-service upload flow, whose extended FileInfo returns confirmed=true, user_id=the user, file_type="pdf"/"png"
- **WHEN** the user submits CreateNotice with attachment_ids=[a1, a2]
- **THEN** the notice and two `notice_attachments` rows are created with file_type read back from the FileInfo contract (same transaction)

#### Scenario: 引用未确认 / 他人文件被拒
- **GIVEN** a user submitting CreateNotice with an attachment id whose FileInfo has confirmed=false, or whose user_id does not match the authenticated user
- **WHEN** the system validates the attachment references via GetFileUrl
- **THEN** the request is rejected with 080005 参数无效 (attachment reference invalid) and no notice is created; the stale/foreign file is not bound

#### Scenario: 单通知附件总量超限被拒（D23）
- **GIVEN** a user submitting CreateNotice with attachment_ids whose count exceeds 10, or whose bound files' total size exceeds 50MB (verified from FileInfo.file_size)
- **WHEN** the system validates the attachment references
- **THEN** the request is rejected with 080005 参数无效（附件超限，见 REQ-AS-6）and no notice is created

### Requirement: REQ-NP-7 — CreateNotice 后端不幂等（D25）

The CreateNotice RPC SHALL be non-idempotent: the system SHALL NOT implement an idempotency key or deduplication on create, and a repeated submission of the same create payload (e.g., a network retry, or a double tap that reaches the server) SHALL produce a distinct duplicate notice. Duplicate suppression is a client responsibility (the publish form disables submission while the request is in flight, see REQ-NM-5); the backend SHALL NOT compensate for client double-submission in this change.

#### Scenario: 双击/重试产生重复通知（后端不幂等）
- **GIVEN** a user whose client submit is disabled while in flight, but a network retry nevertheless delivers the same CreateNotice payload twice
- **WHEN** the server receives both requests
- **THEN** two distinct notices are created (no deduplication, non-idempotent per D25); the client-side submit-disable reduces (but does not eliminate) this risk

#### Scenario: 提交中禁用（前端防重，D25）
- **GIVEN** a user who tapped the publish submit button and the request is in flight
- **WHEN** the user taps the button again before the first response returns
- **THEN** the client ignores the second tap (button disabled while in flight) and sends only one CreateNotice request (see REQ-NM-5)

## 服务职责边界

- **community-hub-service**: notices 写入/查询、notice_scope 维护、notice_attachments 写入、通知撤回（仅发布者本人，附件保留 D28）、发布范围展开（社区管理员 division → 具体小区，经 masterdata GetResidentialAreasByDivision，status=1）；`publisher_id`/`role` 从 JWT 实际身份派生；单事务原子落库；目标数上限校验（≤100，080003）；**附件绑定校验经 file GetFileUrl 读扩展 FileInfo（confirmed + user_id 归属 + file_type 回读，D24）+ 单通知总量上限（≤10 个/≤50MB，080005，D23）**；**published_at 审核通过时设置（D27）**
- **file-service**: 附件引用校验的数据提供方（GetFileUrl 返回扩展 FileInfo：confirmed/user_id/file_type/file_size，D24）；单通知总量上限的尺寸数据源（FileInfo.file_size）；上传白名单/大小两层校验（见 attachment-security capability）
- **permission-service**: `AssertPublishScope` 统一判据（越权→060007，消费方映射 080006）、`GetUserRoles` 供 GetPublishPermission 角色状态查询、`GetDataScopes` 供范围选项；本变更改种子（见 REQ-PP-4，授 421 + 置 min_verf_level=2 + 收 property_admin/owner/tenant 421，D26）；**division→community 授权集解析是否需要改判据逻辑由 design gate（REV-17）定夺**
- **master-data-service**: `md_residential_area` 提供范围单位；`GetResidentialAreasByDivision`（community_div_id>0 分支，status=1 仅审核通过，向下展开）供社区管理员 division 展开；`ResolveScopeAncestors`（向上祖先链）供 AssertPublishScope 解析（只读复用）
- **api-proto**: `community/v1` CreateNoticeRequest 新增 `community_ids`(8，`[jstype=JS_STRING]`)/`division_id`(9，`[jstype=JS_STRING]`)、`community_id`(1) deprecated、`role` 语义由服务端派生；NoticeAttachment 加 file_type；**community/v1 头注释错误码块对齐实际语义（080003=目标数超限、080005=参数无效，剔除陈旧 080003 寻失注释，D29）**；file/v1 FileInfo 扩展 file_type/confirmed（D24）
- **design gate**: division 授权如何落入 community 授权集（division grant 解析契约为 permission-service 单测/集成验收覆盖）须经设计评审验证后方可编码（D17/REV-17）
