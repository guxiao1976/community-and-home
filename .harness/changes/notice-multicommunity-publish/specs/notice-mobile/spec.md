# Notice Mobile Capability Specification

## Purpose

定义移动端（web/mobile）通知模块的页面与组件行为（Q6 已拍板：仅移动端，PC 后续单独排期）：首页通栏跑马灯（NoticeMarquee，最近 15 天标题滚动，更多→浏览页；**15 天窗口以 published_at=审核通过日为唯一锚，D27/D30/D32**）、浏览页（NoticeList 按发布时间倒序）、详情页（NoticeDetail 标题/正文/时间/附件，携带当前小区上下文与后端 scope 规则一致）、【我的】页发布入口（can_publish 驱动）、发布表单（NoticePublisher：标题/正文/附件上传/范围选择，**提交中禁用防双击，D25**），并沉淀可复用组件。范围选择与后端发布范围语义一致（网格员多选小区 / 社区管理员选社区由后端展开 / **业委固定本小区；property_admin 无移动端发布入口，D26**），范围选项数据源为 permission GetDataScopes（division 选项经 masterdata division 树）。附件预校验含单文件白名单/大小 + **单通知总量（≤10 个 且 总 ≤50MB，D23）**。

## Requirements

### Requirement: REQ-NM-1 — 首页通栏跑马灯 NoticeMarquee

The home page SHALL render a full-width marquee bar (below the community switcher) that scrolls the titles of in-scope, moderation-passed notices within the last 15 days for the current community (置顶优先 + 倒序 + 封顶 10 条, boundary `published_at >= now - 15*24h` inclusive, consistent with REQ-NR-3; `published_at` is the approval-anchored value, D27/D30). The marquee SHALL consume the dedicated `GetMarqueeNotices` RPC data contract (D12) — the mobile marquee SHALL NOT re-derive the 15-day window / pin-ordering / 10-cap from a full ListNotices response. Tapping「更多 ›」SHALL navigate to the notice browse page; tapping a single title SHALL navigate to that notice's detail page.

#### Scenario: 切换小区后跑马灯刷新
- **GIVEN** the current community is C1 with in-scope notices, and the user switches to C2
- **WHEN** the community switch completes
- **THEN** the marquee re-fetches and displays C2's last-15-day notice titles; no stale C1 titles remain

#### Scenario: 更多进入浏览页 / 点击标题进详情
- **GIVEN** the marquee is showing notice titles
- **WHEN** the user taps「更多 ›」
- **THEN** the app navigates to the notice browse page; when the user taps a title, the app navigates to that notice's detail page

#### Scenario: 15 天边界与后端一致（含端点）
- **GIVEN** a notice published exactly 15 days before the current time
- **WHEN** the home page requests marquee data
- **THEN** the notice is included (`>= now - 15*24h`, boundary inclusive — matches REQ-NR-3, so the mobile marquee does not diverge from the server contract)

#### Scenario: 空数据占位
- **GIVEN** the current community has no moderation-passed notices in the last 15 days
- **WHEN** the home page loads
- **THEN** the marquee area renders an empty/placeholder state and does not error

### Requirement: REQ-NM-2 — 浏览页 NoticeList（发布时间倒序）

The notice browse page SHALL list the current community's moderation-passed notices in published_at DESC order (same notice_scope-filtered data as REQ-NR-1), with pagination.

#### Scenario: 倒序列表
- **GIVEN** notices N3(11:00), N1(10:00), N2(09:00) in the current community
- **WHEN** the user opens the browse page
- **THEN** the list renders N3, N1, N2 in that order

#### Scenario: 分页加载
- **GIVEN** more notices than the page size
- **WHEN** the user scrolls to the bottom of the browse page
- **THEN** the next page is loaded and appended; when all pages are exhausted the list stops loading

### Requirement: REQ-NM-3 — 详情页 NoticeDetail（携带当前小区上下文）

The notice detail page SHALL display the notice title, content body, publish time, and its attachment list (file_name / file_url / file_size / file_type). The page SHALL render attachments as a tappable list that opens/downloads the file. **The displayed publish time SHALL be the approval-anchored `published_at` (D27/D30) — the date the notice became visible, not its creation date.** The detail request SHALL carry the current community as the GetNotice request context (per REQ-NR-2); an out-of-scope or moderation-not-passed notice SHALL surface as a not-found/empty state, consistent with the backend's 080001 scope rule.

#### Scenario: 正常详情渲染
- **GIVEN** a notice with title, content, published_at, and two attachments
- **WHEN** the user opens the detail page for a notice in the current community
- **THEN** the page shows title, content, formatted publish time, and the two attachment rows (with file_type)

#### Scenario: 详情不存在 / 越权 / 未过审 → 空态
- **GIVEN** a notice id that does not exist, has been deleted, is outside the current community scope, or is not moderation-passed
- **WHEN** the user opens the detail page
- **THEN** the backend returns 080001 and the page shows a not-found/empty state (no crash, no cross-community content disclosed)

### Requirement: REQ-NM-4 — 【我的】页发布入口（can_publish 驱动）

The【我的】page SHALL show a「发布通知」entry if and only if the `can_publish` flag from GetPublishPermission is true. The frontend SHALL NOT evaluate role codes to decide this (per REQ-PP-2).

#### Scenario: 可发布角色显示入口
- **GIVEN** a grid_worker whose GetPublishPermission returns can_publish=true
- **WHEN** the user opens【我的】page
- **THEN** the「发布通知」entry is visible

#### Scenario: 只读角色隐藏入口（含 property_admin，D26）
- **GIVEN** an owner, or a property_admin whose 421 was revoked (D26), both with can_publish=false
- **WHEN** the user opens【我的】page
- **THEN** the「发布通知」entry is not rendered for either

### Requirement: REQ-NM-5 — 发布表单 NoticePublisher（范围选择与数据源）

The publish form SHALL collect title, content, attachments, and a scope selection, and SHALL submit via CreateNotice with `community_ids` (multi) — or with a division selection for community_admin, which the backend expands (REQ-NP-4). **The form SHALL disable the submit button while the CreateNotice request is in flight (D25): a second tap before the response returns is ignored, so a double-click/double-tap sends only one request; the backend is non-idempotent (REQ-NP-7), so the client-side submit-disable is the only anti-duplicate mechanism.** The form's scope options SHALL be sourced from the publishable-scope data consistent with the user's role: grid_worker SHALL multi-select communities within scope (options from permission-service `GetDataScopes(scope_type=community)`, as today); community_admin SHALL select a division (submitted as division selection, expanded server-side); committee SHALL be fixed to its own single community (selection not editable); **property_admin SHALL NOT reach this form on mobile (D26 — no publish entry, see REQ-NM-4/REQ-PP-1).** **For community_admin's division options: `GetDataScopes` currently only returns `community/building/unit/grid` scope types (no `community_div`), so the division option list SHALL be sourced from master-data-service's division tree (e.g., `GetDivisionTree` / administrative-division listing filtered to the user's county region) — NOT from `GetDataScopes`; whether to later extend `GetDataScopes` with a `community_div` scope_type is a design-gated decision (D17/REV-17) recorded for design review, not assumed by this spec (REVISION-4 clarity SHOULD-3 钉死 division 选项数据源).**

#### Scenario: 网格员多选小区发布
- **GIVEN** a grid_worker whose GetDataScopes(scope_type=community) returns C1, C2, C3
- **WHEN** the user fills title/content, selects C1 and C3, adds an attachment, and submits
- **THEN** the form calls CreateNotice with community_ids=[C1, C3]; on success the app navigates away with a success toast

#### Scenario: 社区管理员选社区（提交 division，后端展开）
- **GIVEN** a community_admin whose publishable scope is a division grant (permission GetDataScopes), and the publish form renders the selectable division
- **WHEN** the user selects a division and submits the form
- **THEN** the form submits the division selection via CreateNotice; the backend expands it to concrete communities, validates scope, and writes the notice_scope snapshot (REQ-NP-4); the mobile form does not pre-resolve the division itself

#### Scenario: 业委范围固定（property_admin 无发布入口，D26）
- **GIVEN** a committee whose community is C1
- **WHEN** the committee opens the publish form
- **THEN** the scope selector shows C1 as fixed/unchangeable, and CreateNotice is submitted with community_ids=[C1]; a property_admin (D26) never reaches this form — no publish entry is rendered for it

#### Scenario: 提交中禁用防双击（D25）
- **GIVEN** a user who tapped the submit button and the CreateNotice request is in flight
- **WHEN** the user taps the submit button again
- **THEN** the second tap is ignored (button disabled while in flight) and only one CreateNotice request is sent; after a successful response the app navigates away (and, if a retry were to reach the server, the backend would create a duplicate — non-idempotent per REQ-NP-7)

#### Scenario: 空标题/正文校验
- **GIVEN** the publish form with an empty title
- **WHEN** the user submits
- **THEN** the form blocks submission and shows a validation message (title required)

### Requirement: REQ-NM-6 — 附件前端一致预校验

The frontend upload flow SHALL apply a client-side pre-check consistent with the backend baseline (whitelist extensions + 10MB single-file size + **single-notice aggregate caps: ≤10 attachments and ≤50MB total, D23**) before requesting an upload URL, to fail fast on obviously invalid files. This pre-check SHALL be a shared reusable validator shared with any other file-upload entry.

#### Scenario: 前端预校验拒绝明显非法文件
- **GIVEN** a user selects "install.exe" in the publish form's attachment picker
- **WHEN** the file is selected
- **THEN** the client rejects it immediately with a message (不可上传该类型文件) and does not call the upload API

#### Scenario: 前端预校验拒绝超量/超大附件（D23）
- **GIVEN** a user who already selected 10 attachments, or whose selected attachments' total size exceeds 50MB, and then selects another file
- **WHEN** the file is selected / the user submits
- **THEN** the client blocks adding/submitting with a message (单通知最多 10 个附件且总大小不超过 50MB) and does not call CreateNotice with the over-limit set (final acceptance still enforced by the backend REQ-AS-6)

#### Scenario: 前端预校验通过但后端仍可能拒绝
- **GIVEN** a client-side-validated 8MB PDF
- **WHEN** the file is uploaded
- **THEN** the client-side pre-check passes; final acceptance is still enforced by the backend two-layer validation (frontend pre-check is UX only, not a security boundary)

### Requirement: REQ-NM-7 — 组件化沉淀

The mobile codebase SHALL implement the notice feature as reusable components: `NoticeMarquee`, `NoticePublisher`, `NoticeList`, `NoticeDetail`, plus the shared attachment validator, so that other aggregation pages (or future PC work) can reuse them without duplicating logic. Each reusable component SHALL encapsulate its own data-fetch contract once; consuming pages SHALL NOT re-implement the notice-entity data fetch.

#### Scenario: 组件复用不重复实现数据拉取
- **GIVEN** the four components implemented under the components directory, each exposing the notice-entity props/events and owning its data-fetch
- **WHEN** any page needs to render a marquee, list, detail, or publisher
- **THEN** it imports the component rather than re-implementing the behavior; a code-level check finds no duplicated notice-entity data-fetch logic outside the components

#### Scenario: 组件加载失败降级
- **GIVEN** the notice API returns an error (e.g., network failure) while a reusable component is loading data
- **WHEN** the component mounts
- **THEN** the component renders a fallback/empty state with an optional retry and does not break the hosting page

## 服务职责边界

- **web/mobile**: 页面（home / notice-browse / notice-detail / mine / publish 表单）与组件（NoticeMarquee / NoticePublisher / NoticeList / NoticeDetail + 附件校验器）；范围选项经 permission `GetDataScopes`（社区列表）与 masterdata division 树（division 选项，REQ-NM-5）；**发布表单提交中禁用（D25）+ 附件总量预校验（D23）+ property_admin 无发布入口（D26）**
- **community-hub-service**: 提供 CreateNotice（含 division 展开）/ ListNotices / GetNotice（带 community_id 上下文）/ GetPublishPermission / GetMarqueeNotices 数据契约（消费方）；published_at 审核通过时设置、创建时写 NULL（D27/D30）
- **permission-service**: `GetDataScopes` 提供社区列表范围选项；**不返回 community_div scope_type**（division 选项经 masterdata division 树，D17 design gate 记录）
- **web/pc**: 本变更不做（Q6）
