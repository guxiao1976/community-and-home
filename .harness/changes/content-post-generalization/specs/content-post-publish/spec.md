# Content Post Publish Capability Specification

## Purpose

定义通用图文发布的发布侧行为契约：`notices` 表通用化为 `content_posts`（保留 title/text/published_at/publisher_id/is_pinned/role/publisher 全字段，D1/Q1 + REVISION；新增 section_code/status/attachment_count；published_at 去 NOT NULL；弃 community_id 走 content_post_scope，D12），新增 content_post_scope 多小区关联表（D13）、content_post_attachments 附件级审核状态（D14），实现两阶段发布状态机（D9/Q9 + REVISION：draft 可编辑 → submitted 提交后不可编辑但可删 → 审核；存储映射 = status 枚举 draft=0/submitted=1/approved=2/rejected=3/withdrawn=4；draft→submitted 由 UpdateContentPost.status=submitted 触发，无独立 Submit RPC；本期无消费者 → submit 即隐式通过置 status=approved + published_at=NOW()）、多小区发布（grid_worker 多小区 / community_admin 选社区展开 / property_admin 本小区 / committee 本小区，D6/Q6）、附件绑定校验（复用 FileInfo 扩展载体，总量上限单源，REVISION）、Kafka content-review 推送（D3/Q3 停 Redis 只推 Kafka；契约单源于 REQ-CPM-2 且含可再生 file_url + version，D7/Q7；at-least-once 落库待推标记 + 定时重推，D20 REVISION）、attachment_count 审核完整性判定（D15）。涉及 community-hub-service（数据模型 + 迁移 + 发布 + draft/submitted 状态机 + Update/Delete + Kafka 推送 + 撤回）。

## Requirements

### Requirement: REQ-CPB-1 — content_posts 数据模型通用化（notices RENAME + 字段演化 + status 枚举 + published_at 去 NOT NULL）

The system SHALL rename the existing `notices` table to `content_posts` and SHALL retain the full field set `title`, `text` (renamed from `content`), `published_at`, `publisher_id` (D1/Q1 title-retention), and SHALL ALSO retain `is_pinned TINYINT DEFAULT 0` (marquee pinning carrier, REVISION REQ-CPR-3 dependency), `role VARCHAR(20) NOT NULL` (publisher publish-role, REVISION) and `publisher VARCHAR(100) NOT NULL` (display name, REVISION) — these three columns are physically preserved by the RENAME (migration/001_initial.sql:15-17) and are explicitly part of this change's model contract. The system SHALL add `section_code VARCHAR(30)` (board: `notice` is the first board, `repair` and future boards allowed), `status TINYINT` and `attachment_count INT` (attachment count used for review-completeness determination, D15). The `status` TINYINT SHALL encode the full publish lifecycle + review outcome with the enumerated values: **0 = draft（草稿，可编辑） / 1 = submitted（已提交，待审核，不可编辑可删） / 2 = approved（审核通过，可展示） / 3 = rejected（审核拒绝，不展示） / 4 = withdrawn（已撤回，不展示）** (REVISION — this redefines the notice-era 0=pending/1=approved/2=rejected/3=withdrawn enum; legacy rows are invisible this change so no conflict, but the new enumeration is the authoritative contract). The system SHALL make `published_at` nullable (`ALTER TABLE ... MODIFY published_at DATETIME DEFAULT NULL`), because per the approval-anchoring contract a content post in `draft` or `submitted`-pending state carries `published_at` NULL and it is set at approval; **this change (no review consumer, D16) SHALL set `published_at = NOW()` at submission time as the implicit-approval anchor** (REVISION — see REQ-CPB-4/REQ-CPB-7; this is the mechanism that keeps the marquee window and list ordering functional this phase). The system SHALL deprecate `community_id` (retain the column during the compatibility period but SHALL NOT write new values; the scope association single-source moves to `content_post_scope`, REUSE:notice-D1). **The system SHALL ALSO make `community_id` nullable in the migration (`ALTER TABLE content_posts MODIFY community_id BIGINT DEFAULT NULL`, REVISION-9/coverage MUST-1)**: the column is currently `BIGINT NOT NULL` (migration/001_initial.sql:12), and because the new publish path leaves `community_id` unset (NULL, scope lives in `content_post_scope`), the INSERT would otherwise violate the NOT NULL constraint and the publish main path would be unusable — same class as the `published_at` nullable migration, which this REVISION makes explicit for `community_id` too. The `moderation_status`/`moderation_time` legacy columns SHALL be retained during the compatibility period and progressively transitioned to `status` + per-attachment review (D12). **On INSERT the system SHALL set `status` explicitly** (draft=0 for a draft entry, approved=2 for an immediate-submit entry, REVISION — legacy rows whose status is NULL/0 are invisible because they lack content_post_scope rows, no impact). Legacy `notices` rows SHALL NOT be migrated (D2/Q2 legacy-data-migration) — only new publishes go through `content_post_scope`. The `role` value SHALL be derived from the publisher's actual held roles via the RBAC→publish-role mapping (REQ-CPP-1); the `publisher` display name SHALL be resolved from the authenticated user's real profile (the identity provider / user profile), NOT trusted from the request body (REVISION REQ-CPB-5 — closing the display-name spoofing vector present in createnoticelogic.go `Publisher: in.Publisher`).

#### Scenario: 创建草稿 published_at 置空 → 提交后隐式通过置 published_at=NOW() 与 status=approved
- **GIVEN** the `notices` table has been renamed to `content_posts` with `published_at` made nullable, `community_id` deprecated, and a publisher whose data scope covers communities C1 and C2
- **WHEN** the publisher creates a content post (status=draft entry) targeting C1 and C2 with a body text and one attachment, then submits it

#### Scenario: community_id 未迁移则创建被拒（迁移先于上线门禁，REVISION-9）
- **GIVEN** the new publish feature is enabled but the `ALTER TABLE content_posts MODIFY community_id BIGINT DEFAULT NULL` migration has not been applied (column still `BIGINT NOT NULL`)
- **WHEN** a publisher creates a content post (which per the scope-single-source contract must leave `community_id` unset/NULL)
- **THEN** the INSERT violates the `community_id` NOT NULL constraint and the publish fails; no partial content_post/scope/attachment row is persisted, and the failure surfaces a dependency/rollout error requiring the `community_id` nullable migration to be run before the feature ships
- **THEN** the draft `content_posts` row is created with `section_code` set, `text`/`title` populated, `publisher_id`/`role` derived from JWT/RBAC, `publisher` = the user's real profile name, `is_pinned`=0, `community_id` left unset (NULL), `published_at` left NULL, `status`=draft(0), `attachment_count`=1; two `content_post_scope` rows (post_id, C1) and (post_id, C2); upon submission the row transitions to `status`=approved(2) and `published_at`=NOW() (implicit approval this phase, D16 REVISION)

#### Scenario: published_at 未迁移则创建被拒（迁移先于上线门禁）
- **GIVEN** the new publish feature is enabled but the `ALTER TABLE ... MODIFY published_at DATETIME DEFAULT NULL` migration has not been applied (column still `NOT NULL`)
- **WHEN** a publisher creates a content post in draft state (which must leave published_at NULL while not yet approved)
- **THEN** the create-time INSERT violates the published_at NOT NULL constraint and the publish fails; no partial post or scope row is persisted, and the failure surfaces a dependency/rollout error requiring the migration to be run before the feature ships

#### Scenario: 存量 notices 行不迁移（Q2 兼容期，本期读路径不可见）
- **GIVEN** legacy `notices` rows created before this change with `community_id` set and no `content_post_scope` row
- **WHEN** a user lists content posts for that community
- **THEN** the legacy post is not returned (scope 查询走 content_post_scope，存量未迁移；迁移回填留后续迭代并挂 BACKLOG；「存量通知迁移后不可见」的业务影响已在 proposal 风险节登记)

#### Scenario: 板块非法值被拒（section_code 白名单）
- **GIVEN** a CreateContentPost request whose `section_code` is empty or not in the registered board set (e.g. an unknown board code)
- **WHEN** the system validates the request
- **THEN** the system rejects with a parameter-invalid error (080005 参数无效) and no content post is created

#### Scenario: 伪造展示名被纠正（publisher 取真实档案，禁请求体信任）
- **GIVEN** an authenticated grid_worker who submits a content post with a forged `publisher` display name in the request body (impersonating another user or an official account)
- **WHEN** the system processes the create
- **THEN** the persisted `publisher` display name is the authenticated user's real profile name, not the forged request-body value (the request body `publisher` SHALL be ignored)

### Requirement: REQ-CPB-2 — content_post_scope 多小区关联表（post_id + community_id，复用 notice_scope 模式）

The system SHALL maintain a `content_post_scope` table with `post_id BIGINT NOT NULL` (referencing `content_posts.id`) and `community_id BIGINT NOT NULL` (referencing `md_residential_area.id`, which represents a community or village) — both SHALL be declared NOT NULL — and SHALL enforce a composite primary key (post_id, community_id) serving as the uniqueness constraint. One content post MAY associate with N distinct communities. The table SHALL carry a read-path index whose leftmost column is `community_id` (e.g. `idx_scope_community (community_id, post_id)`), because the list and marquee read paths filter by `community_id` before joining `content_posts`. Legacy data SHALL NOT be back-filled into this table (D2/Q2) — only new publishes write scope rows.

#### Scenario: 同一小区只写入一条关联
- **GIVEN** a CreateContentPost request whose target list contains C1 twice
- **WHEN** the system persists the post scopes
- **THEN** the system either deduplicates the target list before insert or relies on the composite primary key to prevent duplicate rows; the post ends up with exactly one scope row for C1

#### Scenario: post_id 或 community_id 缺失被拒
- **GIVEN** a content_post_scope row being persisted with a valid post_id but a zero/unset community_id (or vice versa)
- **WHEN** the system writes the content_post_scope row
- **THEN** the insert is rejected by the NOT NULL constraint; no orphan scope row is persisted

#### Scenario: 读路径按 community_id 走索引
- **GIVEN** a community with a large number of content_post_scope rows and `idx_scope_community (community_id, post_id)` present
- **WHEN** the list or marquee read path filters scopes by `community_id = ?`
- **THEN** the query is served by the `community_id`-leftmost index rather than scanning the whole content_post_scope table

### Requirement: REQ-CPB-3 — content_post_attachments 附件级审核状态与 file_id 载体

The system SHALL evolve `notice_attachments` to `content_post_attachments` and SHALL add the columns: `post_id BIGINT NOT NULL` (the attachment's parent content post id — the association column is `post_id` throughout, no `main_id` alias), `review_status TINYINT` (per-attachment review state: 0=pending / 1=approved / 2=rejected; this change defaults new attachments to `approved`, D14), `file_id BIGINT` (the file-service file id, the authoritative carrier for regenerating short-lived presigned URLs, REUSE:notice-D24/S4) and `file_type VARCHAR(20)` (whitelist-validated extension, e.g. `pdf`/`png`, recorded from the FileInfo contract rather than the client). The existing `file_name`/`file_url`/`file_size` columns SHALL be retained; `file_url` is the legacy fallback (new rows MAY store an empty placeholder with `file_id` as the authoritative carrier, REUSE:notice S4). Legacy attachment rows SHALL retain their existing values (review_status/file_id not back-filled; D2/Q2 compatibility).

#### Scenario: 附件落库携带 post_id + review_status + file_id + file_type
- **GIVEN** a CreateContentPost request with attachment_ids=[a1] where a1 is a confirmed file owned by the publisher with FileInfo.file_type="pdf"
- **WHEN** the post and its attachments are persisted
- **THEN** one `content_post_attachments` row is created with `post_id` = the new content post id, `review_status` = approved (default, this change), `file_id` = a1, `file_type` = "pdf" (read back from FileInfo), and the content post's `attachment_count` = 1

#### Scenario: 附件行缺 file_id 的读回退为防御性行为（存量帖本期不可达）
- **GIVEN** a reachable `content_post_attachments` row (in a scope query for a new content post) whose `file_id` is 0/NULL — this can only occur via a future migration/back-fill or a data anomaly, because legacy rows are not reachable this change (no content_post_scope rows, D2)
- **WHEN** a user reads the content post detail
- **THEN** the read path treats the attachment as review-complete for completeness counting consistent with its review_status and falls back to the stored `file_url` for download (defensive behavior, no crash); this scenario is a reserved defensive path, not part of this change's observable behavior for legacy posts

### Requirement: REQ-CPB-4 — 两阶段发布状态机（draft → submitted，D9/Q9 REVISION）

The system SHALL implement a two-phase publish state machine with the storage mapping defined in REQ-CPB-1 (`status`: draft=0 / submitted=1 / approved=2 / rejected=3 / withdrawn=4). The state transitions and edit/delete rules SHALL be: (a) **draft (0)**: created via CreateContentPost with entry `status=draft` (the default entry); editable by the publisher (title/text/attachments/scope/section_code/is_pinned MAY be updated, see REQ-CPB-9); deletable by the publisher; NO Kafka message is pushed while in draft. (b) **draft → submitted**: the ONLY client-triggerable transition, carried by `UpdateContentPostRequest.status = submitted` (the submit action — there is NO separate Submit RPC; REVISION). (c) **submitted (1)**: NOT editable; deletable by the publisher. Submission SHALL be the trigger for the Kafka content-review push (REQ-CPB-7). **In this change, because no review consumer is implemented (D16), submission SHALL be treated as implicit approval: the submit action SHALL set `status = approved(2)` AND `published_at = NOW()` atomically** (REVISION — the `submitted(1)` value is defined for the later consumer phase where a submitted post awaits review, but is not a stable terminal stored state this change). (d) **approved (2)**: displayable, NOT editable, deletable by the publisher (withdraw). (e) **rejected (3)** / **withdrawn (4)**: not displayable; written by the review flow / the delete flow respectively — the review consumer that writes `rejected` is out of scope this change (D18). CreateContentPost SHALL NOT be idempotent (REUSE:notice-D25 — no idempotency key; duplicate submission produces a duplicate post; client-side submit-disable is the duplicate guard). A post with `status != draft` SHALL NOT accept any content edit; the delete (withdraw) capability is available to the publisher for draft/submitted/approved states (REQ-CPB-10).

#### Scenario: draft 草稿可编辑（不推送）
- **GIVEN** a publisher who created a content post in `draft` state
- **WHEN** the publisher updates the draft (text/attachments/scope)
- **THEN** the update is accepted and the draft reflects the new values; `status` remains draft(0); no Kafka review message is pushed while the post remains in `draft`

#### Scenario: submitted 提交后不可编辑但可删（提交即隐式通过）
- **GIVEN** a content post in `draft` state
- **WHEN** (a) the publisher submits it (UpdateContentPostRequest.status=submitted), then (b) attempts to edit it, then (c) deletes it
- **THEN** (a) the submit transitions the post to `status`=approved(2) and `published_at`=NOW() (implicit approval, D16), pushes the Kafka content-review message (REQ-CPB-7); (b) any subsequent edit is rejected with 080005 参数无效（仅 draft 可编辑）; (c) the delete succeeds (withdraw, REQ-CPB-10: soft-delete + status=withdrawn, global across all scope rows, attachments retained)

#### Scenario: draft 状态可删除（非提交状态也可撤回）
- **GIVEN** a content post in `draft` state created by the publisher
- **WHEN** the publisher deletes the draft
- **THEN** the delete succeeds (soft-delete + status=withdrawn); no Kafka message was pushed for this draft and none is pushed on delete

#### Scenario: draft 状态重复提交产生重复帖（后端不幂等）
- **GIVEN** a client whose submit is disabled while in flight, but a network retry delivers the same CreateContentPost payload twice
- **WHEN** the server receives both requests
- **THEN** two distinct content posts are created (non-idempotent per REUSE:notice-D25); the client-side submit-disable reduces (but does not eliminate) this risk

#### Scenario: 非发布者删除被拒
- **GIVEN** a content post published by user U, and another user V who is not the publisher
- **WHEN** V invokes DeleteContentPost for that post
- **THEN** the system rejects the request with 080002 无权限 (author-identity check; deletion is publisher-only, REUSE:notice-D19) and the post (and its scopes) remain unchanged

### Requirement: REQ-CPB-5 — CreateContentPost 多小区发布契约（section_code + scope + 入口状态 + 权限信任边界）

The CreateContentPost RPC SHALL accept a `section_code` (registered board), a body `text`, `title`, an entry `status` field (allowed values: `draft` (default) / `submitted` — a `submitted` entry performs create+submit atomically: persist, push Kafka, set status=approved + published_at=NOW(); REVISION), multi-community scope carrier (`repeated community_ids` int64 JS_STRING; **NO division_id request field — REVISION-10/A2: a community_admin governs exactly one division and the backend auto-expands that single governing division via master-data `GetResidentialAreasByDivision(community_div_id, status=1)` at publish time; the frontend does not select a division, overturning the notice-era `division_id` carrier REUSE:notice-D13/D14**), and `repeated attachment_ids` (file ids). The scope carrier contract, division expansion via master-data `GetResidentialAreasByDivision(community_div_id, status=1)`, the `AssertPublishScope` validation (any out-of-scope or unresolvable target → 080006, fail-closed, REUSE:notice-D31), the target-count cap and the request-shape errors (080005 参数无效) SHALL follow the notice multi-community contract. **The ≤100 target cap SHALL count the expanded snapshot (the length of `community_ids`, or the number of communities after division expansion), not the raw input element count** (REVISION); exceeding it SHALL be rejected with 080003 超限 (REUSE:notice D13). `division_id` submitted by a non-community_admin SHALL be rejected with 080005 参数无效 (REUSE:notice D14). The `publisher_id` SHALL be derived from the authenticated JWT identity; `role` SHALL be derived from the publisher's actual held roles via the RBAC→publish-role mapping (see REQ-CPP-1); `publisher` display name SHALL be resolved from the authenticated user's real profile — **none of publisher_id/role/publisher SHALL be trusted from the request body** (REVISION, REQ-CPB-1). The post, scope rows and attachment rows SHALL be persisted atomically in a single transaction. The write-path role-status gate (min_verf_level=2 on 421, REQ-CPP-3) SHALL be enforced before scope validation (080002 on failure).

#### Scenario: 网格员多小区发布成功（draft 入口后提交）
- **GIVEN** a grid_worker whose data scope covers C1 and C2, both communities exist, and the user passes the write-path role-state gate (REQ-CPP-2/REQ-CPP-3)
- **WHEN** the grid_worker submits CreateContentPost with section_code=notice, entry status=draft, community_ids=[C1, C2], text, title, and attachment_ids, then submits it
- **THEN** the draft is persisted with `community_id` unset, `content_post_scope` records C1 and C2 (single transaction), `section_code`=notice; after submit the status=approved, published_at=NOW(); the response returns the new post id

#### Scenario: 立即提交入口（status=submitted）一步完成发布
- **GIVEN** a grid_worker whose data scope covers C1 and the user passes the write-path gate
- **WHEN** the user submits CreateContentPost with entry status=submitted, community_ids=[C1], text, title
- **THEN** the post is created and immediately transitions to approved (2) with published_at=NOW() in the same transaction, the Kafka content-review message is pushed, and the response returns the new post id (no separate submit call needed)

#### Scenario: 任一目标小区越权导致整体拒绝（080006）
- **GIVEN** a committee whose data scope covers only C1
- **WHEN** the user submits CreateContentPost with community_ids=[C1, C2] where C2 is outside scope
- **THEN** `AssertPublishScope` returns denied and the entire request is rejected with 080006 (data-permission denial); no post or scope row is written for C1 either (all-or-nothing)

#### Scenario: 目标小区不存在（安全拒绝未知节点，080006）
- **GIVEN** a publisher submitting community_ids=[99999] where 99999 does not exist in `md_residential_area`
- **WHEN** the system resolves the target's scope ancestors
- **THEN** the unknown node is treated as a target-level resolution failure and the request is rejected with 080006 (fail-closed, identical to an out-of-scope target; 不静默创建无效小区帖)

#### Scenario: 目标数超限被拒（080003，按展开后快照计量）
- **GIVEN** a publisher submitting community_ids with more than 100 entries, OR a community_admin whose single governing division expands (via GetResidentialAreasByDivision(community_div_id, status=1)) into more than 100 approved communities
- **WHEN** the system validates the target set
- **THEN** the request is rejected with 080003 超限 (the cap counts the expanded snapshot) and nothing is created

#### Scenario: 社区管理员管辖社区展开为空被拒（080005，A2）
- **GIVEN** a community_admin whose single governing division (derived from scope) expands into zero approved communities
- **WHEN** the system auto-expands the admin's governing division
- **THEN** the request is rejected with 080005 参数无效 (degenerate expansion of the admin's governing division; distinct from a nonexistent `community_ids` element → 080006) and nothing is created

#### Scenario: 客户端伪造 role / publisher_id / publisher 被纠正（JWT 真实身份为准）
- **GIVEN** an authenticated grid_worker who submits CreateContentPost with a forged role, a `publisher_id` of another user, and a forged display name
- **WHEN** the system processes the create
- **THEN** the persisted post's `publisher_id` is the authenticated JWT user id, `role` is derived from the user's actual held roles, and `publisher` is the user's real profile name — none taken from the forged request body

### Requirement: REQ-CPB-6 — 附件绑定校验（attachment_ids 引用已确认文件 + 总量上限，单源）

The system SHALL only bind attachment ids that reference files confirmed in file-service and owned by the publishing user. On CreateContentPost, for each `attachment_ids` entry the system SHALL invoke file-service `GetFileUrl(file_id)` (no new file-service RPC) and confirm the returned extended FileInfo has `confirmed == true` and `user_id == authenticated user`; the `file_type` is read back from `FileInfo.file_type` (REUSE:notice-D24). **This Requirement SHALL be the single source for the aggregate caps (REVISION — REQ-CAS-5 references it rather than re-declaring them)**: `attachment_ids` count ≤10 AND the sum of the bound files' `FileInfo.file_size` ≤50MB (REUSE:notice-D23); exceeding either SHALL be rejected with 080005 参数无效（附件超限）and no post created. The `attachment_count` on the content post SHALL equal the number of bound attachments. The same binding validation SHALL be re-run on any draft attachment-set change (REQ-CPB-9).

#### Scenario: 有效附件绑定成功
- **GIVEN** a user who confirmed two files (a1, a2) whose FileInfo returns confirmed=true, user_id=the user, file_type="pdf"/"png"
- **WHEN** the user submits CreateContentPost with attachment_ids=[a1, a2]
- **THEN** the post and two `content_post_attachments` rows are created with file_type read back from the FileInfo contract (same transaction) and `attachment_count`=2

#### Scenario: 引用未确认 / 他人文件被拒
- **GIVEN** a user submitting CreateContentPost with an attachment id whose FileInfo has confirmed=false, or whose user_id does not match the authenticated user
- **WHEN** the system validates the attachment references via GetFileUrl
- **THEN** the request is rejected with 080005 参数无效 (attachment reference invalid) and no post is created

#### Scenario: 单帖附件总量超限被拒（≤10 个/≤50MB）
- **GIVEN** a user submitting CreateContentPost with attachment_ids whose count exceeds 10, or whose bound files' total size exceeds 50MB (verified from FileInfo.file_size)
- **WHEN** the system validates the attachment references
- **THEN** the request is rejected with 080005 参数无效（附件超限）and no post is created

### Requirement: REQ-CPB-7 — Kafka content-review 推送（停 Redis 只推 Kafka，契约单源，at-least-once REVISION）

When a content post is submitted (the submit action, REQ-CPB-4/REQ-CPB-9), the system SHALL package a JSON message and push it to the Kafka `content-review` topic. **The message payload SHALL follow the single, authoritative contract defined in REQ-CPM-2 (REVISION — this Requirement does not re-enumerate the payload fields; it references REQ-CPM-2, which is the single source).** The system SHALL NOT push content_posts to the Redis `moderation:task:queue` (D3/Q3 redis-kafka-dual-write — content_posts moves to Kafka; lostfound/user and other sources SHALL continue using Redis). **The push SHALL be at-least-once (REVISION — replacing the prior best-effort MAY-retry): within the submit transaction the system SHALL record the post's push state (a pending-push marker), and the system SHALL implement a reconciliation path (scheduled rescan) that retries delivery for posts recorded as pending-push until acknowledged or quarantined; a permanently-failed push SHALL be observable (pending-push / quarantine metric + log).** A Kafka push failure SHALL NOT roll back the published post (the post remains visible because status is approved this phase, D16), but the failure SHALL NOT be silently dropped — it SHALL leave the post in pending-push for reconciliation. The business risk of a never-delivered review message (a post that is never reviewed and remains visible by default approval once a consumer exists) SHALL be registered explicitly in the proposal risk section with the pending-push observable.

#### Scenario: submitted 推送 Kafka 契约消息（契约遵循 REQ-CPM-2）
- **GIVEN** a content post in draft state with one approved attachment, and the publisher submits it
- **WHEN** the submission completes
- **THEN** the system records the push state and pushes to the `content-review` topic a JSON message conforming to REQ-CPM-2 (post_id, section_code, text, publisher_id, version, attachments[0] = {file_id, file_type, review_status, file_url}); no message is LPUSHed to `moderation:task:queue` for this content post; on successful push the pending-push marker is cleared

#### Scenario: Kafka 不可用时发布不阻塞但进入待推（at-least-once 重推）
- **GIVEN** the Kafka broker is unavailable at submission time
- **WHEN** the post is submitted
- **THEN** the post is still persisted and visible (status=approved, D16); the post is recorded as pending-push; the reconciliation scan retries delivery to `content-review` after the broker recovers until acknowledged or quarantined; the pending-push count is observable via metric/log

#### Scenario: draft 状态不推送
- **GIVEN** a content post in `draft` state being saved/updated
- **WHEN** the draft is updated
- **THEN** no Kafka review message is pushed for the draft (review push only happens on submission to `submitted`, REQ-CPB-4)

### Requirement: REQ-CPB-8 — attachment_count 审核完整性判定（D15，读路径单源谓词）

The system SHALL determine content-post review completeness using `attachment_count`: a content post SHALL be considered review-complete for display when `count(content_post_attachments WHERE review_status=approved) == content_posts.attachment_count` AND the post's `status == approved` (body passed). If any attachment has `review_status == rejected`, the post SHALL NOT be displayed. **This completeness predicate SHALL be the single source used by all read paths (list/detail/marquee, see REQ-CPR). The predicate SHALL NOT mutate `content_posts.status` — a rejected attachment SHALL be hidden by the predicate; the status value `rejected(3)` SHALL be written only by the review flow (the later consumer, D18), not by any read path** (REVISION — resolves the design-doc §3.1 ambiguity; see the synced design doc). `attachment_count` SHALL be frozen at submission (REVISION — it is recalculated on every draft attachment-set change per REQ-CPB-9, then frozen on submit so the predicate is stable). This change defaults status and review_status to approved, so a post with all-approved attachments and approved body satisfies the predicate immediately (no consumer required, D16). As a defensive fallback, a read path MAY reconcile the stored `attachment_count` against the live count when the two diverge, but the stored-count predicate SHALL be the primary behavior and any divergence SHALL be logged as a data-consistency anomaly (REVISION).

#### Scenario: 正文通过 + 附件审核完整 → 展示
- **GIVEN** a content post with status=approved, attachment_count=2, and both attachments review_status=approved
- **WHEN** a user reads the post or lists its section
- **THEN** the post is returned (review-complete: 已审附件数 2 == attachment_count 2 且 正文 approved)

#### Scenario: 任一附件 rejected → 不展示（谓词隐藏，status 不写入）
- **GIVEN** a content post with status=approved, attachment_count=2, attachment A approved and attachment B rejected (review_status=2)
- **WHEN** a user reads the post or lists its section
- **THEN** the post is NOT returned (任一附件 rejected → 整体不展示); the post's `content_posts.status` is NOT mutated by this read (it stays approved; the review flow later writes status=rejected); the post remains visible to the review flow but not to end readers

#### Scenario: 附件审核数小于计数（审核不完整）→ 不展示
- **GIVEN** a content post with status=approved but attachment_count=3 and only 2 attachments approved (1 still pending)
- **WHEN** a user lists the section
- **THEN** the post is NOT returned (审核不完整：已审附件数 2 != attachment_count 3)

#### Scenario: 无附件帖审核完整性恒成立
- **GIVEN** a content post with attachment_count=0 (no attachments) and status=approved
- **WHEN** a user lists the section
- **THEN** the post is returned (count(approved attachments)=0 == attachment_count 0 且 正文 approved；无附件恒完整)

### Requirement: REQ-CPB-9 — UpdateContentPost（draft 编辑 + submit 动作 + is_pinned 置位，REVISION）

The UpdateContentPost RPC SHALL define the draft-edit and submit contracts. (a) **Draft edit**: a post in `draft` state MAY have its `title`, `text`, `section_code`, `attachment_ids` (attachment-set change), `community_ids`/`division_id` (scope change) and `is_pinned` updated by the publisher. (b) **attachment_count recalculation invariant (REVISION, REQ-CPB-10/REQ-CPB-8 dependency)**: any attachment-set change SHALL recalculate `attachment_count` to the new bound-attachment count within the SAME transaction, and SHALL re-run the full binding validation (REQ-CPB-6: confirmed + user_id ownership + file_type read-back + ≤10 count / ≤50MB total) and the completeness predicate inputs; new attachments SHALL default `review_status=approved` (this change, D14). (c) **Scope change**: a draft scope change SHALL re-run `AssertPublishScope` over the new target set (any out-of-scope → 080006, all-or-nothing) and rewrite the content_post_scope rows accordingly. (d) **Submit action**: `UpdateContentPostRequest.status = submitted` on a draft SHALL perform submission — validate the final state, push the Kafka content-review message (REQ-CPB-7), and set `status=approved(2)` + `published_at=NOW()` atomically (implicit approval, D16); the `status` field SHALL accept ONLY the submit transition value (`submitted`); any other status value SHALL be rejected with 080005 (review outcomes are written only by the review flow). (e) **Non-draft immutability**: any content edit on a post with `status != draft` SHALL be rejected with 080005 参数无效（仅 draft 可编辑）. (f) **is_pinned**: SHALL be settable on a draft by its publisher, and on a submitted/approved post by a user holding a publish-capable role (REQ-CPP-1) whose data scope covers the post's community (the marquee pin operator, REVISION REQ-CPR-3); `is_pinned` defaults to 0 and is honored by the marquee ordering.

#### Scenario: draft 编辑附件并重算 attachment_count
- **GIVEN** a draft post with attachment_count=2 (attachments a1, a2)
- **WHEN** the publisher updates the draft removing a1 (leaving a2) in the same transaction
- **THEN** `content_post_attachments` reflects a2 only, `attachment_count` is recalculated to 1 in the same transaction, the completeness predicate on the later-submitted post compares against the new count (no stale count makes the post permanently hidden)

#### Scenario: draft 编辑新增附件复跑绑定校验（超限被拒）
- **GIVEN** a draft post whose attachment-set change would bind files with a combined size exceeding 50MB (or count exceeding 10)
- **WHEN** the publisher updates the draft adding those attachments
- **THEN** the update is rejected with 080005 参数无效（附件超限）; `attachment_count` and the attachment rows are unchanged (all-or-nothing)

#### Scenario: draft scope 变更越权被拒（重跑 AssertPublishScope）
- **GIVEN** a draft post scoped to C1, and the publisher whose scope covers only C1
- **WHEN** the publisher updates the draft adding community C2 (outside scope)
- **THEN** the update is rejected with 080006 (AssertPublishScope re-run on the new target set) and the scope rows are unchanged

#### Scenario: submit 动作触发推送 + 隐式通过
- **GIVEN** a draft post with final content validated
- **WHEN** the publisher calls UpdateContentPost with status=submitted
- **THEN** the post transitions to status=approved(2) with published_at=NOW() (implicit approval, D16), the Kafka content-review message is pushed (REQ-CPB-7), and the post becomes immutable to content edits

#### Scenario: submitted/approved 帖子编辑被拒
- **GIVEN** a content post in status=approved (already submitted)
- **WHEN** the publisher attempts to update its text or attachments
- **THEN** the update is rejected with 080005 参数无效（仅 draft 可编辑）; no content, attachment_count, or scope change occurs

#### Scenario: 授权操作者置顶 approved 帖（is_pinned 置位可执行）
- **GIVEN** an approved content post in community C1, and an operator holding a publish-capable role (e.g. committee) whose data scope covers C1
- **WHEN** the operator calls UpdateContentPost with is_pinned=true on the approved post
- **THEN** the post's `is_pinned` becomes 1 and the marquee ordering (REQ-CPR-3 `order by is_pinned desc`) places it first

### Requirement: REQ-CPB-10 — DeleteContentPost（撤回，仅发布者本人，REVISION）

The system SHALL provide a DeleteContentPost RPC that withdraws a content post. Deletion SHALL be restricted to the publisher (author-identity check; another user SHALL be rejected with 080002 无权限, REUSE:notice-D19). Deletion SHALL be available for posts in `draft`/`submitted`/`approved` states (REVISION — draft delete is explicitly allowed; submitted 不可编辑但可删). The delete SHALL be global across all scope rows: the post row SHALL be soft-deleted (`deleted_at` set) and `status` SHALL be set to `withdrawn(4)` (audit trail); the `content_post_scope` rows SHALL be retained (not physically deleted — the post's withdrawal is expressed by the soft-deleted post row, and all scope rows share that single post row); `content_post_attachments` rows SHALL be retained. A soft-deleted/withdrawn post SHALL NOT be returned by any read path (list/detail/marquee, REQ-CPR). No Kafka message SHALL be pushed on delete (withdraw is not a review submission).

#### Scenario: 发布者撤回已发布帖（全局生效）
- **GIVEN** an approved content post scoped to C1 and C2, published by user U
- **WHEN** U calls DeleteContentPost for that post
- **THEN** the post row is soft-deleted (deleted_at set) and status=withdrawn(4); the (post,C1) and (post,C2) scope rows and all attachment rows are retained; the post disappears from list/detail/marquee for both C1 and C2; no Kafka message is pushed

#### Scenario: 非发布者删除被拒（080002）
- **GIVEN** a content post published by user U, and another user V
- **WHEN** V invokes DeleteContentPost for that post
- **THEN** the system rejects with 080002 无权限 and the post (and its scopes/attachments) remain unchanged

#### Scenario: 已撤回帖不再返回（任何读路径）
- **GIVEN** a content post withdrawn via DeleteContentPost (soft-deleted + status=withdrawn)
- **WHEN** a user lists content posts for its former community or reads it by id
- **THEN** the post is not returned (soft-deleted/withdrawn excluded, consistent with REQ-CPR-1/REQ-CPR-2/REQ-CPR-3)

## 服务职责边界

- **community-hub-service**: content_posts 写入/查询、content_post_scope 维护、content_post_attachments 写入、发布（多小区 scope + 附件绑定 + 单事务 + section_code + 入口状态 draft/submitted）、两阶段 draft/submitted 状态机（draft 可编辑 / submitted 不可编辑可删 / submit 隐式通过）、UpdateContentPost（draft 编辑 + attachment_count 重算 + scope 复校验 + is_pinned + submit 动作）、DeleteContentPost（撤回，软删 + withdrawn）、Kafka content-review 推送（at-least-once，落库待推标记 + 定时重推）、attachment_count 审核完整性判定（读路径单源谓词，不 mutate status）；`publisher_id`/`role`/`publisher` 从 JWT 实际身份/真实档案派生；published_at submit 即置 NOW()；`community_id` 弃用（不写入）
- **file-service**: 附件引用校验的数据提供方（GetFileUrl 返回扩展 FileInfo：confirmed/user_id/file_type/file_size，REUSE:notice-D24）；总量上限的尺寸数据源；上传白名单/大小两层校验（见 content-post-attachment-security capability）
- **permission-service**: `AssertPublishScope` 统一判据（越权→060007，消费方映射 080006）、`GetUserRoles`、`GetDataScopes`；种子（property_admin 保留 421、grid_worker 授 421、撤销 owner/tenant 421、421 置 min_verf_level=2，D6/D22 REVISION）
- **master-data-service**: `GetResidentialAreasByDivision`（division→小区展开，status=1）+ `ResolveScopeAncestors`（只读复用）
- **moderation-service**: Kafka content-review 消费者后期开发（D18）；Redis 消费者对 content_posts 不再回调 NoticeService（D4）
- **api-proto**: community/v1 CreateContentPost/UpdateContentPost/DeleteContentPost 系列直接改名（D4）+ section_code/status 入口字段/ContentPostAttachment（review_status/file_id/file_type）+ UpdateNoticeModerationStatus 移除（D21）
- **docker-compose**: Kafka 单节点 KRaft + 数据卷持久化 + content-review topic（D8/D17）
