# Notice Moderation Capability Specification

## Purpose

定义通知审核流的复用契约：多小区发布后的通知仍走现有 moderation 审核流（发布→审核→通过才可见），不新增审核链路、不改变 `UpdateNoticeModerationStatus` 契约。多小区模型下，审核状态是通知级别（一条通知整体审核），不因关联小区数而拆分。**`published_at` 语义锚定审核通过时（D27/D30：创建时写 NULL、经 UpdateNoticeModerationStatus 通过回调时设置 published_at=now），跑马灯 15 天窗口/列表倒序/详情发布时间均以可见日为基准（原实现创建时即设，属行为登记）**。本 capability 还覆盖 `UpdateNotice`（编辑/置顶）在多小区模型下的处置边界：保持通知级语义，编辑后 moderation_status 重置为 0 的既有行为不变，scope 编辑不在本变更范围。涉及 community-hub-service（状态读取/回调/published_at 设置）、moderation-service（只读复用）、api-proto（复用现有 UpdateNoticeModerationStatus）。

## Requirements

### Requirement: REQ-NP-MOD-1 — 通知整体经审核通过才可见

The system SHALL expose a notice only after its moderation status indicates pass. A notice SHALL have a single moderation status at the notice level regardless of how many communities it is published to; a multi-community notice SHALL NOT become visible in any community before its overall moderation passes. This is the visibility gate for the list, detail, and marquee read paths (REQ-NR-1/REQ-NR-2/REQ-NR-3).

#### Scenario: 审核通过后各小区可见
- **GIVEN** a multi-community notice (C1, C2) created with pending moderation
- **WHEN** moderation passes for that notice
- **THEN** the notice becomes visible in both C1 and C2 lists simultaneously (single status, applied globally)

#### Scenario: 审核未通过不可见
- **GIVEN** a multi-community notice (C1, C2) with moderation pending or failed
- **WHEN** a user lists notices for C1
- **THEN** the notice is not returned (仅审核通过的可见)

#### Scenario: 审核失败对全部小区生效
- **GIVEN** a multi-community notice (C1, C2) with moderation failed
- **WHEN** a user lists notices for either C1 or C2
- **THEN** the notice is not returned in either community

### Requirement: REQ-NP-MOD-2 — 复用现有 UpdateNoticeModerationStatus 回调

The system SHALL keep using the existing `UpdateNoticeModerationStatus` RPC as the moderation-service → community-hub-service callback, with the existing moderation_status semantics (machine_pass / machine_fail / human_pass / human_fail). No new moderation API SHALL be introduced by this change.

#### Scenario: 回调更新通知审核状态
- **GIVEN** a notice with id and current moderation_status
- **WHEN** moderation-service invokes UpdateNoticeModerationStatus with status=human_pass
- **THEN** the notice's moderation_status is updated accordingly and the notice becomes listable

#### Scenario: 回调目标通知不存在
- **GIVEN** a callback referencing a non-existent or deleted notice id
- **WHEN** UpdateNoticeModerationStatus is invoked
- **THEN** the system returns a not-found response (080001) without corrupting other notices

### Requirement: REQ-NP-MOD-3 — UpdateNotice（编辑/置顶）处置边界：保持通知级语义

The system SHALL keep `UpdateNotice` as the notice-level edit operation (title, content, is_pinned) with its existing semantics: editing resets `moderation_status` to 0 (re-audit required), and the change is applied at notice level, not per-community. This change SHALL NOT introduce scope-editing (adding/removing communities after publish) via UpdateNotice — scope edits are out of scope for this change. `is_pinned` is set through this existing UpdateNotice path; this change adds no pin-writing UI (see REQ-NR-3).

#### Scenario: 编辑后重置审核状态（既有语义保持）
- **GIVEN** a moderation-passed multi-community notice (C1, C2)
- **WHEN** a user invokes UpdateNotice to change its content
- **THEN** the content is updated, `moderation_status` resets to 0 (pending) globally, and the notice becomes invisible in both C1 and C2 until re-approved (REQ-NP-MOD-1)

#### Scenario: 编辑不改变发布范围
- **GIVEN** a notice published to C1 and C2
- **WHEN** a user invokes UpdateNotice to change title/content/is_pinned
- **THEN** the notice_scope rows (C1, C2) are unchanged; UpdateNotice does not add/remove scope associations (scope edits are out of scope)

### Requirement: REQ-NP-MOD-4 — published_at 锚定审核通过时（D27/D30，创建写 NULL）

The system SHALL set (or update) a notice's `published_at` to the current time when its moderation passes — i.e., in the community-hub-service handler of the `UpdateNoticeModerationStatus` pass callback (human_pass / machine_pass). **D30 pins the create-time schema contract: the notice SHALL be created with `published_at` NULL (pending), not with a create-time placeholder; this requires the `ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL` migration (see REQ-NP-1), because the current column is `DATETIME NOT NULL` (001_initial.sql:19) and the current insert writes published_at at create time (createnoticelogic.go:62 / model/notice.go:52-56). Without the migration, the create-time INSERT violates NOT NULL and the publish main path is unusable (registered as an exception gate in REQ-NP-1).** Until first approval the notice is invisible (REQ-NP-MOD-1), so published_at=NULL before approval is not user-visible and does not participate in idx_published ordering or the marquee window; on approval the callback SHALL set published_at=now. After an edit that resets `moderation_status` to 0 (REQ-NP-MOD-3), a subsequent re-approval SHALL update `published_at` to the new approval time (while re-pending, published_at keeps the previous approval value or NULL if never approved — either way the row is invisible via REQ-NP-MOD-1). This SHALL be the single anchor for the marquee 15-day window, list DESC ordering, and detail publish time (REQ-NR-1 / REQ-NR-3 / REQ-NM-3). This is a behavior change from the current implementation (createnoticelogic.go sets PublishedAt at create time) and is registered as a behavior change.

#### Scenario: 首次审核通过时设置 published_at（创建时 published_at=NULL，D30）
- **GIVEN** a notice created at time T0 with moderation pending and `published_at` persisted as NULL (create-time insert writes NULL per D30, enabled by the REQ-NP-1 migration)
- **WHEN** moderation passes at time T1 (UpdateNoticeModerationStatus with pass status)
- **THEN** the notice's published_at is set to T1 (the approval time); the marquee 15-day window and list ordering anchor on T1, not T0; the pending row's NULL published_at never appeared in any ordering or window (invisible via REQ-NP-MOD-1)

#### Scenario: 编辑后重新审核通过 → published_at 更新
- **GIVEN** a moderation-passed notice with published_at=T1, edited via UpdateNotice (moderation_status resets to 0, published_at unchanged while pending)
- **WHEN** moderation passes again at time T2
- **THEN** published_at is updated to T2; the 15-day window restarts from the re-approval time

#### Scenario: 审核失败/待审不设置 published_at
- **GIVEN** a notice with moderation status pending or failed (published_at NULL if never approved, or the previous approval value if re-pending after edit)
- **WHEN** the notice is in that state
- **THEN** published_at is not (re)set by the callback; the notice remains invisible regardless of published_at (REQ-NP-MOD-1)

#### Scenario: published_at 未迁移则创建失败（异常门禁，D30，与 REQ-NP-1 一致）
- **GIVEN** the publish feature is enabled but the `ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL` migration has not been applied (column still `DATETIME NOT NULL`)
- **WHEN** a publisher creates a notice (which per D27/D30 must persist published_at as NULL)
- **THEN** the create-time INSERT violates the NOT NULL constraint and the publish fails (dependency/rollout error); the migration must be applied before the feature ships

## 服务职责边界

- **community-hub-service**: 通知审核状态存储、可见性过滤、UpdateNoticeModerationStatus 回调处理（**通过时设置/更新 published_at，D27；创建时 published_at 写 NULL，D30**）、UpdateNotice 既有语义保持（编辑重置审核状态）
- **moderation-service**: 审核执行（只读复用，本变更不修改）
- **api-proto**: 复用现有 UpdateModerationStatusRequest/Response（不新增）
