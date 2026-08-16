# Notice Time Window Capability Specification

## Purpose

统一通知类内容的时间口径、排序与数据量行为：首页通知只展示最近 30 天内、最多 3 条、按发布时间倒序（最新在上）；首页跑马灯滚动内容取自同一 30 天数据（由首页通知列表派生，非独立 RPC）；通知列表页仅展示 30 天内、按置顶优先 + 发布时间倒序，与首页口径一致。时间窗口由后端服务强制（遵守「前端不定义业务逻辑、接口与 api-proto 一致」约束），前端只传窗口参数/消费过滤结果。涉及 web/mobile（首页通知区、列表页、跑马灯）与 community-hub-service（ListContentPosts 时间窗口过滤）与 api-proto（ListContentPostsRequest 可选时间窗口参数）。

> **REVISION 修订记录**：首轮修订解决评审 6 项 MUST/SHOULD——① 新增 REQ-NTW-5 显式定义通知卡片视觉契约（修复 REQ-NTW-4 对 REQ-HL-2 的错误交叉引用）；② REQ-NTW-4 补齐列表页数据量行为（page/page_size 复用 + 单请求截断语义）；③ REQ-NTW-1/2/4 定义 published_at NULL / 未来时间边界场景；④ REQ-NTW-3 改为「跑马灯由首页通知列表派生」，删除 GetMarqueeNotices 后端 15→30 天无效变更；⑤ 新增 REQ-NTW-6 索引/性能验收点。
>
> **第 2 轮 REVISION（本轮）**：r2-2（评审 coverage 同轮 MUST）——`since_days` 需经 REST 层透传：`GET /api/community/notices` 的 `since_days` query/form 参数由 `api/internal/types.ListContentPostsReq` 承载（新增 form 字段），并经 `api/internal/logic/notice/listcontentpostslogic.go` 透传至 RPC `ListContentPostsRequest.since_days`；仅列 RPC logic 与 proto 会致 REST 层丢弃该参数、移动端 30 天窗口静默失效（详见 REQ-NTW-2 服务职责边界）。r2-5（评审 clarity 同轮）——REQ-NTW-2 参数校验措辞去掉「non-numeric」：int32 wire 字段值恒为数字，非数字由 REST 网关解析层拒绝，服务端参数校验仅覆盖数值范围。

## Requirements

### Requirement: REQ-NTW-1 — 首页通知仅展示最近 30 天内、最多 3 条、发布时间倒序

The system SHALL display on the mobile homepage notice section at most 3 notices whose `published_at` falls within the last 30 days, ordered newest-first. The window predicate SHALL be `published_at >= (now - 30 days)` AND `published_at <= now`: notices with `published_at` earlier than 30 days ago SHALL NOT appear even if pinned; notices with `published_at` in the future (pre-scheduled, `published_at > now`) SHALL NOT appear; notices whose `published_at` is NULL SHALL NOT appear (NULL is not within the window, regardless of `status`). The "at most 3" cap SHALL be achieved by the frontend requesting `page_size = 3` (existing pagination parameter), while the 30-day window filter SHALL be enforced server-side (REQ-NTW-2). Within the window, pinned notices SHALL retain priority over non-pinned before the newest-first ordering truncates to 3 (existing `is_pinned DESC, published_at DESC` contract).

#### Scenario: 30 天窗口内最新 3 条倒序展示
- **GIVEN** a community with 5 review-complete notices, of which 4 are within the last 30 days (published 2h/6h/1d/5d ago) and 1 is 45 days old
- **WHEN** the homepage loads the notice section for that community
- **THEN** the homepage shows at most 3 notices, all within the 30-day window, ordered newest-first (2h → 6h → 1d); the 45-day-old notice is excluded

#### Scenario: 30 天窗口内超 3 条时截断为最新 3 条
- **GIVEN** a community with 6 review-complete notices all published within the last 30 days
- **WHEN** the homepage loads the notice section
- **THEN** only the 3 most recently published notices are shown (newest-first), and no notice older than those 3 appears

#### Scenario: 30 天前有置顶旧帖时不进首页
- **GIVEN** a community whose only pinned notice was published 40 days ago and 2 non-pinned notices published within the last 30 days
- **WHEN** the homepage loads the notice section
- **THEN** the 30-day window excludes the pinned 40-day-old notice; the homepage shows only the 2 in-window notices (pinned priority applies only within the window)

#### Scenario: 窗口内置顶帖与更早非置顶帖并存时置顶优先（REVISION 补：锁定 top-3 截断与置顶交互）
- **GIVEN** a community with 1 pinned notice published 20 days ago and 1 non-pinned notice published 1 hour ago, plus 3 older in-window notices
- **WHEN** the homepage loads the notice section
- **THEN** the pinned 20-day-old notice is shown first (pinned priority within window), then the newest non-pinned notices, truncated to 3 total; the 1-hour-old notice is shown before the 20-day-old non-pinned ones

#### Scenario: approved 但 published_at 为 NULL 的行不进首页（REVISION #3 边界）
- **GIVEN** a community whose only review-complete notice has `published_at = NULL` (draft-flow row that reached approved status but the publish time was never anchored)
- **WHEN** the homepage loads the notice section
- **THEN** the notice is excluded (NULL is not within the 30-day window); the homepage renders the empty state rather than substituting it

#### Scenario: 预排期（published_at > now）通知不进首页（REVISION #3 边界）
- **GIVEN** a community with a review-complete notice whose `published_at` is 2 days in the future (pre-scheduled) and no other in-window notice
- **WHEN** the homepage loads the notice section
- **THEN** the future-dated notice is excluded (window upper bound `published_at <= now`); the homepage does not show a not-yet-published notice

#### Scenario: 30 天窗口内无通知显示空态
- **GIVEN** a community with no review-complete notice published within the last 30 days (or none at all)
- **WHEN** the homepage loads the notice section
- **THEN** the homepage renders the empty state "暂无通知公告" and no stale (>30d) notice is substituted

### Requirement: REQ-NTW-2 — 30 天窗口由后端强制，前端不实现窗口业务逻辑

The system SHALL enforce the 30-day notice window server-side: the content-list read contract SHALL accept an optional server-side time-window parameter (`since_days`, int32, unit = days), and the service SHALL restrict results to posts with `published_at >= (now - since_days)` AND `published_at <= now` when the parameter is provided. The mobile frontend SHALL NOT compute the 30-day cutoff as business logic; it SHALL only pass the window parameter (or consume server-filtered results). When the parameter is absent, the read SHALL NOT apply a time filter (preserving existing non-mobile callers such as the PC notice management list). The valid value range SHALL be 1..365 inclusive; values ≤0 or >365 SHALL be rejected with the parameter-invalid error (REVISION r2-5: an int32 wire value is inherently numeric, so "non-numeric" is not a reachable service-side case — a malformed REST query value is rejected at the gateway parse layer, which likewise surfaces the parameter-invalid error; the service-side check covers the numeric range only).

#### Scenario: 前端传窗口参数、后端过滤
- **GIVEN** a mobile homepage/list requesting notices with the 30-day window parameter for a community that has both in-window and out-of-window review-complete notices
- **WHEN** the backend serves the list request
- **THEN** only in-window notices are returned (both lower bound `>= now - since_days` and upper bound `<= now` applied); out-of-window notices are excluded server-side, and the frontend does not re-filter

#### Scenario: 缺省不传参数时保留既有行为（PC 不受影响）
- **GIVEN** an existing caller (e.g. PC notice management list) invoking the content-list read without the time-window parameter
- **WHEN** the backend serves the request
- **THEN** no time filter is applied and the caller receives the full published set (behavioral backward-compatibility; additive optional parameter)

#### Scenario: 非法窗口参数不静默吞错
- **GIVEN** a caller passing an invalid time-window value (zero, negative, or >365)
- **WHEN** the backend validates the request
- **THEN** the backend rejects the request with a parameter-invalid error (community-hub 080005 参数无效) rather than silently applying a wrong window (REVISION r2-5: a malformed non-numeric REST query value is rejected at the gateway parse layer with the same parameter-invalid error; the service-side range check covers ≤0 / >365)

### Requirement: REQ-NTW-3 — 首页跑马灯由首页通知列表派生（同源 30 天数据），不依赖独立跑马灯 RPC

The system SHALL derive the homepage marquee scrolling content from the same notice list that feeds the homepage notice cards (the server-filtered 30-day result of the homepage notice read, `getNoticeList` → ListContentPosts), NOT from a separate marquee RPC. The marquee SHALL scroll the titles of the notices returned by that same homepage list (which is capped at 3), in the order returned. The 30-day window on the marquee SHALL therefore be the same window enforced on the homepage notice list (REQ-NTW-1/2). When no notice is within the window, the marquee SHALL show the empty message "暂无通知公告". The backend marquee read (`GetMarqueeNotices`) is NOT consumed by the mobile homepage and SHALL NOT be modified by this change (REVISION #6: removed the ineffective 15→30-day constant change).

#### Scenario: 跑马灯标题集 = 首页通知列表标题（同源）
- **GIVEN** a community with 4 in-window notices, of which the homepage list returns the newest 3 (titles T1/T2/T3, published 1h/2h/3h ago) and 1 older in-window notice (5d ago) not in the top-3
- **WHEN** the homepage renders the marquee
- **THEN** the marquee scrolls only T1/T2/T3 (the same titles as the homepage cards, in the same order); the 5d-old notice's title does not appear in the marquee (marquee inherits the homepage list cap of 3)

#### Scenario: 30 天窗口内无通知时跑马灯空态
- **GIVEN** a community with no review-complete notice in the 30-day window
- **WHEN** the homepage renders the marquee
- **THEN** the marquee displays "暂无通知公告" and does not fabricate content

### Requirement: REQ-NTW-4 — 通知列表页仅展示 30 天内、置顶优先 + 发布时间倒序，与首页口径一致

The notice list page (`pages/notice-browse/notice-browse`) SHALL display only notices published within the last 30 days (by `published_at`, D2), ordered by the same contract as the homepage: pinned notices first (`is_pinned DESC`), then `published_at DESC` within the window. The list SHALL render in the notice card style defined by REQ-NTW-5 (role color bar / role tag / title / time, REVISION #1: fixed cross-reference); tapping a notice card SHALL navigate to the notice detail page. The 30-day filter SHALL be server-enforced (REQ-NTW-2), replacing the previous client-side cutoff (`notice-browse.vue` currently requests 50 and filters 3 months client-side). The window predicate SHALL be the same as the homepage: `published_at >= (now - 30 days)` AND `published_at <= now`; NULL `published_at` rows and future-dated rows SHALL be excluded (REVISION #3).

#### Scenario: 列表页展示 30 天内通知并倒序
- **GIVEN** a community with in-window notices (published 1h/1d/20d ago, non-pinned) and one 40-day-old notice
- **WHEN** a user opens the notice list page
- **THEN** only the 3 in-window notices are listed newest-first (1h → 1d → 20d); the 40-day-old notice is excluded

#### Scenario: 列表页置顶帖优先于更早的非置顶帖（REVISION 补：排序契约与首页一致）
- **GIVEN** a community with 1 pinned notice published 10 days ago and 1 non-pinned notice published 1 hour ago
- **WHEN** a user opens the notice list page
- **THEN** the pinned 10-day-old notice is listed first, then the 1-hour-old non-pinned notice (pinned priority before recency, matching REQ-NTW-1)

#### Scenario: 列表页超过一屏的数据量行为——单请求、固定 page_size、窗口内截断（REVISION #2）
- **GIVEN** a community whose 30-day window contains 120 notices, and the list page requests the window with `page_size = 50`
- **WHEN** a user opens the notice list page
- **THEN** the system returns the newest 50 in-window notices (newest-first), the response `total` reflects the full in-window count (120), and the list renders those 50 in a single scrollable list without client-side pagination/infinite-scroll this iteration; the window filter and page/page_size parameters are the existing ListContentPosts contract (additive optional window parameter)

#### Scenario: 点击通知卡片进入详情
- **GIVEN** a rendered notice card on the list page
- **WHEN** the user taps the card
- **THEN** the system navigates to the notice detail page for that notice id

#### Scenario: 列表加载失败明确提示（不静默）
- **GIVEN** the notice list API fails (network error / backend unavailable)
- **WHEN** a user opens the list page
- **THEN** the page shows an explicit load-failure message (禁止静默吞错) and does not render partial/stale data as fresh

#### Scenario: 列表 30 天窗口内无通知显示空态
- **GIVEN** a community with no review-complete notice in the 30-day window
- **WHEN** a user opens the list page
- **THEN** the page renders the empty state "暂无通知公告"

### Requirement: REQ-NTW-5 — 通知卡片视觉契约（role 色条 / role 标签 / 标题 / 时间）显式锚点

The notice card (shared by the homepage notice section and the notice list page) SHALL render each notice with: (a) a role color bar on the card's leading edge whose color is derived from the notice's `role`; (b) a role tag (pill) showing the role name, colored by the same role color; (c) the notice title; (d) the publish time formatted from `published_at` (fallback `created_at` when `published_at` is NULL). REQ-NTW-4 SHALL reference this contract as the list style source. (REVISION #1: this REQ provides the explicit anchor for the "homepage-style list" the list page must match; the previous cross-reference to REQ-HL-2 was erroneous — REQ-HL-2 is the lost-and-found section contract and is unrelated.)

#### Scenario: 首页与列表页卡片按同一视觉契约渲染
- **GIVEN** a notice with a non-default role
- **WHEN** the notice is rendered both on the homepage and on the list page
- **THEN** both render the same role color bar, role tag, title, and formatted publish time (from `published_at`, or `created_at` when `published_at` is NULL)

#### Scenario: 时间格式与回退一致
- **GIVEN** a notice with `published_at` set and a notice with `published_at` NULL but `created_at` set
- **WHEN** the cards are rendered
- **THEN** the first shows the time formatted from `published_at`; the second shows the time formatted from `created_at` (no blank time), consistent with the current homepage card behavior

### Requirement: REQ-NTW-6 — 30 天窗口过滤 + 倒序不走全表扫描（性能验收点）

The 30-day window filtered, newest-first ordered list read (ListContentPosts with `since_days`) SHALL be index-servable and SHALL NOT require a full table scan of `content_posts` for a community of typical size (existing scope-driven query joins `content_post_scope` on `community_id` and orders by `is_pinned DESC, published_at DESC`). The design SHALL add an index that serves the scope-filtered `published_at` predicate and `published_at DESC` ordering (e.g. on `content_posts (status, published_at)` or a scope-JOIN-usable `published_at` index), and the implementation SHALL verify with a query plan (e.g. EXPLAIN) that the windowed read does not degrade to a full table scan. (REVISION #5: the existing `idx_published(community_id, published_at DESC, deleted_at)` cannot serve this query because `community_id` on `content_posts` is a deprecated NULL column; scope filtering happens via the join table.)

#### Scenario: 窗口查询走索引而非全表扫描
- **GIVEN** a community whose `content_posts` contains a large number of rows across many communities, and a windowed list request (`since_days = 30`)
- **WHEN** the backend executes the list read
- **THEN** the query plan uses the added index (no full table scan on `content_posts`), and the read completes within the service's normal latency budget

#### Scenario: 索引变更不破坏既有非窗口调用
- **GIVEN** the index is added and an existing caller (no `since_days`) executes the same read
- **WHEN** the read runs
- **THEN** results are unchanged (index is additive; absence of `since_days` preserves the unfiltered behavior of REQ-NTW-2)

## 服务职责边界

- **web/mobile**: 首页通知区/跑马灯/列表页消费后端 30 天窗口过滤结果；首页传 `since_days` + `page_size=3`（不实现窗口业务逻辑）；列表页传 `since_days` + 固定 `page_size` 并渲染为与首页一致的卡片列表（REQ-NTW-5 视觉契约、置顶优先 + published_at 倒序）；点击卡片进详情；加载失败明确提示
- **community-hub-service**: ListContentPosts 读路径支持可选时间窗口参数 `since_days`（`published_at >= now - since_days AND published_at <= now`，服务端强制，缺省不过滤，NULL/未来行排除）；新增支持窗口过滤 + 倒序的索引（REQ-NTW-6）；非法窗口参数返回 080005；**GetMarqueeNotices 不变**（移动端跑马灯由首页通知列表派生，本能力不修改 GetMarqueeNotices，REVISION #6）。**REST 层透传（REVISION r2-2）**：`since_days` 必须贯通完整链路——`api/internal/types.ListContentPostsReq` 新增 `since_days` form 字段（`api/internal/types/types.go`），`api/internal/logic/notice/listcontentpostslogic.go` 将其透传至 RPC `ListContentPostsRequest.since_days`，RPC `rpc/internal/logic/notice/listcontentpostslogic.go` 应用窗口谓词；REST 层任一环丢失该参数都会使移动端 30 天窗口静默失效
- **api-proto**: `ListContentPostsRequest` 新增可选时间窗口参数 `since_days`（int32，天，有效值 1..365，additive 字段 6，缺省行为不变）——proto 变更由本管线 Owner 走全局 Claude 流程
