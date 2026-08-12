# 发布校验（AssertPublishScope）Specification

## Purpose

在 community-hub-service 的所有写接口落库前强制数据权限校验：内容的目标小区必须被发布者的授权集合覆盖（global 放行；否则逐目标祖先链 ∩ 授权集合 ≠ ∅），且 `publisher_id` 一律取自 JWT 认证身份、忽略客户端传值。由此杜绝跨小区发布（owner@A 发到 B）与发布者身份伪造（抓包改 publisher_id），使"能发到哪"由后端权威决定。

## Requirements

### Requirement: AssertPublishScope 数据权限判定

permission-service SHALL provide an RPC `AssertPublishScope(userId, targets []scopeRef)` that returns success if and only if every target is covered by the caller's authorized set: a `global` scope passes all targets; otherwise each target passes when its ancestor chain intersects the authorized set (per the unified rule); any uncovered target SHALL cause a denial with an error indicating no data permission (reference code `80006`).

> `[NEEDS CLARIFICATION] 待阶段3架构定稿`：`80006` 为设计文档引用的错误码，与 community-hub 现状错误码命名空间（`08xxxx`）不一致；最终错误码取值与归属由架构设计对齐，本 spec 只约束行为——拒绝且返回"无数据权限"类错误。

#### Scenario: owner 发布到自己小区 → 通过
- **GIVEN** an owner authorized to community A
- **WHEN** `AssertPublishScope(A)` is evaluated
- **THEN** the target A is covered → success

#### Scenario: owner 发布到未授权小区 → 拒绝
- **GIVEN** an owner authorized to community A only
- **WHEN** `AssertPublishScope(B)` is evaluated for community B outside the scope
- **THEN** B is uncovered → the request is denied with a no-data-permission error (80006)

#### Scenario: global 审核员发布任意小区 → 通过
- **GIVEN** an auditor holding a `global` scope
- **WHEN** `AssertPublishScope` is evaluated for any community target
- **THEN** all targets pass (global matches everything)

### Requirement: 多目标全部通过才放行

`AssertPublishScope` SHALL pass a multi-target request only when every target is covered; a single uncovered target SHALL fail the whole request.

#### Scenario: 全部目标覆盖则放行
- **GIVEN** targets [A, B] where both A and B are covered
- **WHEN** `AssertPublishScope([A, B])` is evaluated
- **THEN** the request passes

#### Scenario: 部分目标未覆盖则整体拒绝
- **GIVEN** targets [A, B] where A is covered and B is not
- **WHEN** `AssertPublishScope([A, B])` is evaluated
- **THEN** the request is denied (B fails), even though A passes

### Requirement: 所有改变数据状态的写操作挂载数据权限校验

community-hub-service SHALL enforce `AssertPublishScope` before persisting ANY state-changing write operation, using the operation's content `community_id` as the target. The set of state-changing write operations SHALL include at least: notice create (`CreateNotice`), notice update (`UpdateNotice`), notice delete (`DeleteNotice`), notice moderation-status update (`UpdateNoticeModerationStatus`); lostfound create (`CreateLostFound`), lostfound resolve (`ResolveLostFound`), lostfound moderation-status update (`UpdateLostFoundModerationStatus`); and contact upsert (`UpsertContacts`). No state change SHALL be persisted for an operation whose target community is not covered by the caller's data scope.

#### Scenario: 发布前先校验再落库
- **GIVEN** a user submitting a lostfound post to community A
- **WHEN** the create handler processes the request
- **THEN** `AssertPublishScope(A)` is evaluated before the record is persisted

#### Scenario: 通知更新同样受数据权限约束
- **GIVEN** a user updating a notice whose community_id is A
- **WHEN** the update handler processes the request
- **THEN** `AssertPublishScope(A)` is enforced before the update is persisted

#### Scenario: 删除通知超出数据范围被拒
- **GIVEN** a user attempting to delete a notice whose community_id is outside the user's data scope
- **WHEN** the delete handler processes the request
- **THEN** `AssertPublishScope` fails and the notice is not deleted (no unauthorized state change)

#### Scenario: 解决寻失超出数据范围被拒
- **GIVEN** a user attempting to resolve a lostfound item whose community_id is outside the user's data scope
- **WHEN** the resolve handler processes the request
- **THEN** the resolution is rejected before persistence

#### Scenario: 审核状态回调目标小区未覆盖则拒绝
- **GIVEN** a moderation-status update callback for content whose community_id is not covered by the caller's data scope
- **WHEN** the callback handler applies the update
- **THEN** the state change is rejected before persistence (the target community is validated, not assumed)

### Requirement: publisher_id 取自 JWT

community-hub-service MUST derive `publisher_id` from the authenticated JWT identity (the user_id carried by the request), NOT from the request body. Any client-supplied `publisher_id` SHALL be ignored and replaced by the JWT identity.

#### Scenario: 伪造 publisher_id 无效
- **GIVEN** an attacker submitting a lostfound post with a forged `publisher_id` equal to another user's id
- **WHEN** the create handler processes the request
- **THEN** the stored `publisher_id` is the attacker's own JWT user_id, not the forged value

#### Scenario: 合法发布者身份正确落库
- **GIVEN** a legitimate owner in community A publishing a post
- **WHEN** the create handler processes the request
- **THEN** the stored `publisher_id` equals the owner's user_id taken from the JWT

### Requirement: AssertPublishScope 校验身份取自 JWT

The user identity used for data-permission validation — the `userId` passed to `AssertPublishScope` — SHALL be derived exclusively from the authenticated JWT (the same identity used for `publisher_id`). A client-supplied `userId` SHALL be ignored and replaced by the JWT identity; the validation MUST NOT be performed on behalf of an identity chosen by the caller. This SHALL apply to every state-changing write operation covered by REQ-5.3.

#### Scenario: 伪造 admin userId 不能借用数据权限
- **GIVEN** an attacker submitting a write whose body includes a forged `userId` belonging to a `global`-scoped admin
- **WHEN** the handler builds `AssertPublishScope(userId, targets)`
- **THEN** the `userId` used is the attacker's own JWT identity, not the forged one; the check runs against the attacker's scope and is denied for an uncovered target

#### Scenario: 校验身份与 publisher_id 一致
- **GIVEN** a legitimate owner in community A submitting a write
- **WHEN** the handler validates data permission and persists
- **THEN** both the `AssertPublishScope` `userId` and the stored `publisher_id` equal the owner's JWT user_id

### Requirement: 校验链路顺序

For a community write, community-hub-service SHALL apply, in order: functional permission (existing `PermMiddleware` / `CheckPermission`) → data permission (`AssertPublishScope`) → persist. Quota and moderation checks are out of scope for this change and SHALL NOT be assumed by this requirement.

#### Scenario: 权限链路通过后落库
- **GIVEN** an unverified owner in community A who holds the publish permission (`min_verf_level=0`)
- **WHEN** the publish flow runs
- **THEN** functional permission passes, `AssertPublishScope(A)` passes, and the content is persisted

#### Scenario: 数据权限拦截在功能权限之后
- **GIVEN** a user who passes functional permission but whose target community is not in scope
- **WHEN** the publish flow runs
- **THEN** the flow stops at the data-permission stage and the content is not persisted
