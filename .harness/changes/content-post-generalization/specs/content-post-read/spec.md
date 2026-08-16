# Content Post Read Capability Specification

## Purpose

定义通用图文发布的读路径行为契约：列表 / 详情 / 跑马灯在通用化模型下的行为。ListContentPosts 按当前小区经 content_post_scope 关联过滤可见帖，可加 `section_code` 板块筛选与可选 `role` 过滤（按发布者 role 列，notice 兼容语义，REVISION），响应保持单 community_id（取请求小区，由 scope 匹配行派生，不读弃用列）；GetContentPost 新增 community_id 请求上下文（必填，缺失 080005；scope 外/不存在/审核未完整 → 080001，不泄露）；GetMarqueeNotices 跑马灯数据（≤10 条、置顶优先 + 最近 15 天倒序、仅审核通过且审核完整，D5/Q5 本期新实现；is_pinned 为保留列且置位机制定义于 REQ-CPB-9，REVISION）。所有读路径统一应用审核完整性判定（正文 status=approved 且 已审附件数==attachment_count，REQ-CPB-8）与 published_at 锚定（本期 submit 即置 NOW()，REVISION — 排序/窗口锚点始终非空）。涉及 community-hub-service（读侧）。

## Requirements

### Requirement: REQ-CPR-1 — 列表按当前小区经 content_post_scope 过滤 + section_code 板块筛选 + role 过滤（REVISION）

The ListContentPosts RPC SHALL accept a single-value `community_id` (request context) and SHALL return only content posts whose `content_post_scope` includes that `community_id`, excluding soft-deleted posts (deleted_at set / status=withdrawn) and posts that do not pass the review-completeness gate (REQ-CPB-8: body status=approved AND count(attachments approved)==attachment_count). The RPC SHALL support an optional `section_code` filter (board filter; when present, only posts of that board are returned). The response `ContentPost` SHALL keep a single `community_id` field whose value SHALL be the request `community_id` (derived from the matching content_post_scope row) rather than the deprecated content_posts.community_id column (NULL for new posts). **The optional `role` filter SHALL be retained with notice-compatible semantics and is hereby defined: it filters the result set by the post's `role` column (the publisher publish-role recorded at create time, REQ-CPB-1), matching the NoticeRole enum; a caller passing role R SHALL receive only posts whose `role == R`.** Pagination and `published_at DESC` ordering SHALL be supported; **the ordering anchor is the submit-anchored `published_at` which is set to NOW() at submission this phase (REVISION — every readable post carries a non-NULL published_at because the completeness gate requires status=approved, which this phase only reaches via the submit implicit-approval path); as a defensive measure, any residual NULL `published_at` SHALL sort last (NULLS LAST) and SHALL NOT perturb the DESC order** (REVISION — replaces the prior ungrounded claim that NULL rows are pre-excluded by the gate).

#### Scenario: 多小区帖只出现在其发布的小区
- **GIVEN** a content post published to C1 and C2 (two content_post_scope rows) and review-complete, and no post published to C3
- **WHEN** a user lists content posts for community C3
- **THEN** the response contains zero rows for that post; listing for C1 or C2 each return the post

#### Scenario: section_code 板块筛选
- **GIVEN** a community with a notice-board post and a repair-board post, both review-complete
- **WHEN** a user lists content posts for the community with section_code="notice"
- **THEN** only the notice-board post is returned; the repair-board post is excluded

#### Scenario: 响应 community_id 取请求小区（不读弃用列）
- **GIVEN** a multi-community content post whose content_posts.community_id is NULL but whose content_post_scope matches C1 and C2
- **WHEN** a user lists content posts for community C2
- **THEN** the response post's community_id equals C2 (derived from the scope row), not a value read from the deprecated column

#### Scenario: role 过滤按发布者角色列筛选
- **GIVEN** a community with a post published by a committee (role=committee) and a post published by a grid_worker (role=grid_officer), both review-complete
- **WHEN** a user lists content posts for the community with role filter = committee
- **THEN** only the committee-published post is returned; the grid_officer post is excluded

#### Scenario: 未过完整性门禁的帖不返回
- **GIVEN** a content post with attachment_count=2 but only 1 attachment approved (review incomplete, REQ-CPB-8)
- **WHEN** a user lists content posts for a scope community
- **THEN** the post is not returned (越权/未完整不泄露)

#### Scenario: 排序锚点非空（published_at submit 即置值，NULL 排最后）
- **GIVEN** a community with two review-complete posts P1 (published_at = now-1h) and P2 (published_at = now-10min), both approved this phase
- **WHEN** a user lists content posts for the community ordered by published_at DESC
- **THEN** P2 precedes P1 (deterministic DESC ordering; no NULL perturbation because every readable post carries a submit-anchored published_at)

### Requirement: REQ-CPR-2 — 详情按请求小区 scope 校验 + 审核完整性门禁

The GetContentPost RPC SHALL accept `id` AND a required `community_id` request context (missing/empty → 080005 参数无效). The system SHALL return the content post with its attachments only when: the post exists (not soft-deleted), a `content_post_scope` row matches (id, community_id), the post passes the review-completeness gate (REQ-CPB-8), and the caller has read access to that community. Any of scope-outside / not-exists / not-review-complete SHALL map to 080001 (not exposed, indistinguishable from not-exists). The response `ContentPost` SHALL include `attachments[]` with `review_status`/`file_type`/`file_id`; the attachment `file_url` SHALL be regenerated per attachment from `file_id` via `GetFileUrl(file_id)` (presigned URL regeneration, REUSE:notice S4), falling back to the stored `file_url` for rows with file_id=0/NULL (defensive — such rows are not reachable this phase since legacy posts are not returned, REVISION).

#### Scenario: 详情返回完整（含附件审核状态）
- **GIVEN** a review-complete content post in community C1 with one approved attachment, and the caller has read access to C1
- **WHEN** the user calls GetContentPost(id=C1-community, community_id=C1)
- **THEN** the response contains the post (text/title/published_at/section_code/status/attachment_count) and attachments[] with review_status=approved, file_type, file_id, and a regenerated file_url

#### Scenario: scope 外 / 不存在 / 未完整统一 080001
- **GIVEN** a content post that exists but is (a) published only to C2 while the caller requests community C1, or (b) not review-complete (a rejected attachment), or (c) soft-deleted
- **WHEN** the user calls GetContentPost for that post
- **THEN** the system returns 080001 (不存在/scope 外/未完整不泄露，与不存在同，不区分)

#### Scenario: community_id 缺失被拒（080005）
- **GIVEN** a caller invoking GetContentPost with only `id` and no `community_id` request context
- **WHEN** the system validates the request
- **THEN** the system rejects with 080005 参数无效 (community_id 必填) and returns no post

#### Scenario: 附件 file_id=0 回退 stored file_url（防御性读路径）
- **GIVEN** a reachable attachment row with file_id=0/NULL (only conceivable via a future migration/data anomaly; legacy rows are not reachable this phase, D2)
- **WHEN** a user reads the content post detail
- **THEN** the read path falls back to the stored `file_url` for that attachment (download still works, no crash); this is a reserved defensive behavior, not part of this change's observable legacy-post path

### Requirement: REQ-CPR-3 — GetMarqueeNotices 跑马灯数据（新 RPC，D5/Q5）

The system SHALL provide a `GetMarqueeNotices` RPC (newly implemented this change, D5/Q5 — the code does not exist yet) that returns up to 10 notice-board title summaries for a required `community_id`: `section_code` fixed to `notice`, window `published_at >= now-15*24h` (inclusive endpoint, REUSE:notice-D32), `order by is_pinned desc, published_at desc`, limit 10, restricted to review-complete posts (REQ-CPB-8) and to the caller's allowed communities (scope filter). **`is_pinned` is a retained column (REQ-CPB-1) defaulting to 0 and is set via UpdateContentPost by an authorized operator (REQ-CPB-9) — this is the executable pin mechanism** (REVISION). **The window and ordering anchors are `published_at`, which is set to NOW() at submission this phase (implicit approval, REVISION), so the 15-day window is functional for newly published posts** (this phase every submitted post carries a non-NULL published_at; defensive NULLS LAST applies). Missing/empty `community_id` SHALL be rejected with 080005; an empty result SHALL return an empty list. The response items SHALL carry `id` and `title` only (marquee payload, no body).

#### Scenario: 跑马灯返回 ≤10 条置顶优先 + 15 天窗口
- **GIVEN** a community with 12 review-complete notice posts in the last 15 days (2 pinned by an authorized operator via UpdateContentPost is_pinned=true) and 1 post older than 15 days
- **WHEN** a caller invokes GetMarqueeNotices(community_id)
- **THEN** the response returns ≤10 items, the 2 pinned posts first, then remaining posts by published_at DESC, all within the 15-day window (the >15-day post is excluded); each item carries id + title

#### Scenario: 本期新帖入跑马灯（published_at submit 即置值）
- **GIVEN** a notice post submitted this change (status=approved, published_at=NOW()) and a second submitted one hour earlier, both review-complete
- **WHEN** a caller invokes GetMarqueeNotices for a scope community
- **THEN** both posts fall in the 15-day window (non-NULL published_at) and are returned ordered by is_pinned desc, published_at desc (newest first); the marquee is non-empty this phase (REVISION — the prior NULL-published_at contradiction is eliminated)

#### Scenario: community_id 缺失被拒（080005）
- **GIVEN** a caller invoking GetMarqueeNotices without a community_id (or with an empty one)
- **WHEN** the system validates the request
- **THEN** the system rejects with 080005 参数无效 and returns no marquee items

#### Scenario: 空态返回空列表
- **GIVEN** a community with no review-complete notice posts in the 15-day window
- **WHEN** a caller invokes GetMarqueeNotices(community_id)
- **THEN** the response returns an empty list (空态，不报错)

#### Scenario: 未过完整性门禁的帖不入跑马灯
- **GIVEN** a notice post whose attachment was rejected (review incomplete, REQ-CPB-8)
- **WHEN** a caller invokes GetMarqueeNotices for a scope community
- **THEN** the rejected post is excluded from the marquee (仅审核通过且审核完整才展示)

#### Scenario: 置顶帖的置位来源可执行（GIVEN 可复现）
- **GIVEN** an approved notice post in community C1 whose `is_pinned` was set to 1 by an authorized operator (committee holding a publish-capable role scoped to C1) via UpdateContentPost (REQ-CPB-9)
- **WHEN** a caller invokes GetMarqueeNotices(community_id=C1)
- **THEN** the pinned post appears before all non-pinned posts within the 15-day window (order by is_pinned desc honored)

## 服务职责边界

- **community-hub-service**: ListContentPosts / GetContentPost / GetMarqueeNotices 读实现（content_post_scope JOIN + section_code 筛选 + role 过滤 + 审核完整性谓词 REQ-CPB-8 + published_at 锚定倒序（NULLS LAST 防御）+ 附件 file_id 重生 + is_pinned 置顶排序）；community_id 必填校验（080005）；scope 外/未完整 → 080001
- **file-service**: GetFileUrl(file_id) 提供附件预签名 URL 重生（REUSE:notice-D24）
- **api-proto**: community/v1 ListContentPosts/GetContentPost/GetMarqueeNotices 契约（ContentPost/ContentPostAttachment 含 review_status/file_type/file_id）
- **permission-service**: 读路径数据范围过滤（FilterAllowed 语义，复用 notice 读路径）
