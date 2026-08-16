# Notice Read Capability Specification

## Purpose

定义通知读路径（列表/详情/跑马灯数据）在多小区模型下的行为契约：ListNotices 入参保持单 community_id，按当前小区经 notice_scope 关联过滤可见通知，响应结构保持单 community_id（Q2 已拍板，取值 = 请求小区，由 notice_scope 匹配行派生，不读弃用列）；GetNotice 新增 community_id 请求上下文（**必填**，字段号 2，缺失 → 080005；scope 外/不存在/审核未过 → 080001，不泄露）；**跑马灯数据由新增专用 RPC `GetMarqueeNotices` 提供**（D12，返回 ≤10 条 title 摘要，置顶优先 + 最近 15 天倒序，仅审核通过，Q7 已拍板；community_id 缺失 → 080005）。**`published_at` 语义锚定审核通过时（D27/D30：创建时写 NULL、通过回调置 now）：列表倒序与跑马灯 15 天窗口均以审核通过日为基准（原创建时即设，行为登记；跑马灯窗口锚定唯一化见 D32）**。本 capability 覆盖 community-hub-service 的读侧行为。

## Requirements

### Requirement: REQ-NR-1 — 列表按当前小区经 notice_scope 过滤，响应 community_id 取请求小区

The ListNotices RPC SHALL keep `community_id` as a single-value request field (unchanged) and SHALL return only notices whose `notice_scope` includes that `community_id`, excluding soft-deleted and moderation-not-passed notices. **The existing optional `role` filter field (community.proto ListNoticesRequest field 2) SHALL be retained with its existing semantics** (filter by notice role) in the multi-community model. The response `Notice` SHALL keep a single `community_id` field (Q2: 响应保持单 community_id，不改为 repeated), and its value SHALL be the request `community_id` (derived from the matching notice_scope row) rather than the deprecated notices.community_id column, which is NULL for new notices. **The `published_at` used for DESC ordering SHALL be the approval-anchored value (D27/D30 — set when moderation passes, NULL at create, see REQ-NP-MOD-4), so ordering reflects the visible date, not the creation date; pending rows (published_at NULL) are excluded by the moderation-visibility gate before ordering (REQ-NP-MOD-1), so NULL never perturbs the DESC order.**

#### Scenario: 多小区通知只出现在其发布的小区
- **GIVEN** a notice published to C1 and C2 (two notice_scope rows), and no notice published to C3
- **WHEN** a user lists notices for community C3
- **THEN** the response contains zero rows for that notice; listing for C1 or C2 each return the notice

#### Scenario: 响应 community_id 取请求小区（不读弃用列）
- **GIVEN** a multi-community notice whose notices.community_id is NULL but whose notice_scope matches C1 and C2
- **WHEN** a user lists notices for community C2
- **THEN** each returned Notice carries community_id = C2 (the request community), not the NULL/deprecated column value

#### Scenario: 按角色筛选保留
- **GIVEN** a user listing notices for the current community with `role` filter set to a NoticeRole value
- **WHEN** the ListNotices request includes the role field
- **THEN** the response is further filtered to notices whose role matches the value, unchanged from the pre-change semantics

#### Scenario: 分页边界
- **GIVEN** more notices than `page_size`
- **WHEN** a user requests page 1 and page 2 with page_size=N
- **THEN** the response returns at most N notices per page, `total` reflects the full count for the current community, and the two pages are disjoint

#### Scenario: 按发布时间倒序
- **GIVEN** notices N1(published 10:00), N2(09:00), N3(11:00) in the same community
- **WHEN** a user lists notices for that community
- **THEN** the returned order is N3, N1, N2 (published_at DESC)

### Requirement: REQ-NR-2 — 详情返回标题/正文/时间/附件（scope 内才可见，community_id 必填）

The GetNotice RPC SHALL return the notice's title, content, published_at, and its attachments (each with file_name, file_url, file_size, file_type) for a single notice id. The GetNoticeRequest SHALL carry a `community_id` request-context field (new field, the requesting user's current community) alongside `id`; **`community_id` SHALL be required** — a request with it missing/empty SHALL be rejected with 080005 参数无效 (the scope context is mandatory for the not-found privacy rule). The system SHALL return a notice only if it is not soft-deleted, its moderation status is passed, AND its `notice_scope` includes the request `community_id`. If the notice exists but is outside the request context's scope (or is not moderation-passed), the system SHALL return 080001 通知不存在 (not-found, identical to the absent case) so the read path does not leak cross-community or unreviewed content.

#### Scenario: 正常详情
- **GIVEN** an existing notice published to the current community with two attachments (id, file_name, file_url, file_size, file_type)
- **WHEN** a user invokes GetNotice with the notice id and the current community_id
- **THEN** the response contains title, content, published_at, and the two attachments with all five attachment fields

#### Scenario: 通知不存在
- **GIVEN** no notice with the requested id, or the notice has been soft-deleted
- **WHEN** a user invokes GetNotice with that id
- **THEN** the system returns a not-found error (080001 通知不存在)

#### Scenario: community_id 缺失被拒
- **GIVEN** a request to GetNotice that omits `community_id` (or sends it empty)
- **WHEN** the request reaches the handler
- **THEN** the system rejects with 080005 参数无效 (community_id is a required request-context field, per D15); it does NOT silently fall back to any default community

#### Scenario: 通知存在但不在请求者小区 scope 内（拒绝、不泄露）
- **GIVEN** a notice published only to C1 (notice_scope includes C1), and a user in current community C2 requests it with community_id=C2
- **WHEN** the user invokes GetNotice with the notice id and community_id=C2
- **THEN** the system returns 080001 通知不存在 (scope-outside treated as not-found); the notice title/content is not disclosed to the C2 user

#### Scenario: 审核未通过不可见
- **GIVEN** a notice with moderation status pending or failed
- **WHEN** a user invokes GetNotice with the notice id and a matching community_id
- **THEN** the system returns 080001 通知不存在 (仅审核通过的可见，与 REQ-NP-MOD-1 一致)

### Requirement: REQ-NR-3 — 首页通栏跑马灯数据（专用 RPC GetMarqueeNotices，最近 15 天，含端点）

The system SHALL provide marquee data for the current community (selected via community switch) through a **dedicated read RPC `GetMarqueeNotices(community_id)`** (D12) returning a list of notice title summaries (notice id + title) for the current community: notices whose `published_at` is within the last 15 days — i.e., `published_at >= now - 15*24h` (boundary inclusive: a notice published exactly 15 days ago SHALL be included) — ordered by is_pinned first then published_at DESC, capped at 10 entries, showing each notice's title. **The request `community_id` SHALL be required — a missing/empty/zero value SHALL be rejected with 080005 参数无效 (consistent with REQ-NR-2/D15); the RPC SHALL NOT fall back to a default community. `published_at` SHALL be set at moderation approval (D27/D30, see REQ-NP-MOD-4), so the 15-day window is measured from the visible date (审核通过日), not the creation date; the creation-to-approval gap is never the window anchor (D32).** The marquee data source SHALL be derived from the same notice_scope-filtered read path and SHALL only include moderation-passed notices (REQ-NP-MOD-1). `is_pinned` values SHALL be read from the notices table as written by the existing `UpdateNotice` mechanism; this change does not introduce a pin-writing entry (out of scope). The response message shape (e.g., a dedicated `NoticeMarqueeItem` with notice id + title) is determined at design stage; the behavioral contract is one entry per notice carrying its id and title.

#### Scenario: 置顶优先 + 时间倒序 + 封顶 10 条
- **GIVEN** a community with 12 in-scope moderation-passed notices within the last 15 days, of which 2 are pinned
- **WHEN** the home page requests marquee data via GetMarqueeNotices
- **THEN** the two pinned notices appear first, remaining notices follow in published_at DESC order, and the list is truncated to 10 entries; each entry carries the notice id and title

#### Scenario: 15 天边界含端点
- **GIVEN** a moderation-passed notice published exactly 15*24h before the current time, and another published 15*24h minus 1 second ago, both in scope
- **WHEN** the home page requests marquee data
- **THEN** both notices are included (`published_at >= now - 15*24h` boundary inclusive); only notices strictly older than the boundary are excluded

#### Scenario: 超过 15 天的通知不入跑马灯（以 published_at 为准）
- **GIVEN** a notice whose published_at (approval day) is 20 days before now and a notice whose published_at is 5 days before now, both in scope and moderation-passed
- **WHEN** the home page requests marquee data
- **THEN** only the published_at-5-days-ago notice is included; the published_at-20-days-ago notice is not (published_at is the sole window anchor, D32)

#### Scenario: 未审核通过的通知不入跑马灯
- **GIVEN** a moderation-pending notice within the last 15 days in scope
- **WHEN** the home page requests marquee data
- **THEN** the pending notice is not included (仅审核通过的可见)

#### Scenario: 审核通过日即为窗口锚（created 20 天前、今天通过 → 入跑马灯，D27/D32）
- **GIVEN** a notice created 20 days ago whose moderation approval lands today (published_at = approval time = now, per D27/D30)
- **WHEN** the home page requests marquee data
- **THEN** the notice IS included in the marquee (published_at = approval day, within the 15-day window); the creation-to-approval gap (审核滞留) is NOT the window anchor — only published_at (approval day) anchors the 15-day window (D32)

#### Scenario: published_at 距今 >15 天不入跑马灯但入浏览列表（D32）
- **GIVEN** a notice whose published_at (approval time) is 20 days before the current time (e.g., approved 20 days ago, whether or not it was created even earlier), in scope and moderation-passed
- **WHEN** the home page requests marquee data, and a user opens the browse list
- **THEN** the notice is excluded from the marquee (published_at 距今 >15 天) but remains in the browse list; published_at (approval day) is the sole 15-day window anchor (D27/D32)

#### Scenario: community_id 缺失被拒（080005）
- **GIVEN** a request to GetMarqueeNotices that omits `community_id` (or sends zero)
- **WHEN** the request reaches the handler
- **THEN** the system rejects with 080005 参数无效 (community_id is required, consistent with REQ-NR-2/D15) and does not fall back to any default community

#### Scenario: 无通知时空状态
- **GIVEN** a community with no moderation-passed notices within the last 15 days
- **WHEN** the home page requests marquee data
- **THEN** the system returns an empty list and the frontend renders an empty/placeholder marquee state without error

## 服务职责边界

- **community-hub-service**: ListNotices / GetNotice / GetMarqueeNotices（经 notice_scope 过滤，仅审核通过）；GetNotice 的 community_id 上下文校验（缺失 080005，scope 外→080001）
- **api-proto**: ListNoticesRequest/Response、GetNoticeRequest（新增 community_id 必填）、新增 GetMarqueeNotices RPC 契约（响应保持单 community_id）
- **web/mobile**: 浏览页/详情页/首页跑马灯消费（见 notice-mobile capability）
