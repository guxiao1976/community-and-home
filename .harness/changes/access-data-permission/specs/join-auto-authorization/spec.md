# 加入自动授权 / 退出撤销（Join Auto-Authorization）Specification

## Purpose

保证「成员资格」与「数据范围」天然一致、无需人工审批：用户加入小区即自动获得对应业主/租户角色 + 该小区 scope（未认证状态，即可在配额内发布），退出小区即自动撤销该小区角色与 scope。授权由 user-service 编排、permission-service 执行，Join/Leave 成功即触发 Assign/RevokeRole，并即时失效缓存，杜绝"加入后无权限"或"退出后仍可发布"的漂移。

## Requirements

### Requirement: 加入小区即自动授权

On a successful `JoinCommunity`, user-service SHALL synchronously call permission-service `AssignRole` with role `owner` (自有) or `tenant` (租住) — determined by the housing-ownership choice captured in the join/house-registration flow — scope_type=`community`, scope_id=the community id, and role status `0` (未认证).

> `[NEEDS CLARIFICATION] 待阶段3架构定稿`：触发自动授权的确切 API 形状——`JoinCommunity` 是否携带「自有/租住」权属、是否与房屋注册步骤合并——由架构设计确定；本 spec 只约束行为契约：加入成功后 owner/tenant 角色 + 该小区 scope 必须同步出现。

#### Scenario: 自购房屋加入 → owner 角色
- **GIVEN** a user joins community A and registers the housing as owned (自有)
- **WHEN** the join flow succeeds
- **THEN** a rel_user_role (owner, community, A, status=0) is created

#### Scenario: 租住加入 → tenant 角色
- **GIVEN** a user joins community A and registers the housing as rented (租住)
- **WHEN** the join flow succeeds
- **THEN** a rel_user_role (tenant, community, A, status=0) is created

#### Scenario: 授权失败则加入不成功
- **GIVEN** a join where the membership record is created but the synchronous `AssignRole` fails
- **WHEN** the join flow completes
- **THEN** the join is reported as failed (membership creation rolled back or otherwise kept consistent with the missing grant), leaving no member-without-scope state

### Requirement: 自动授权幂等

`AssignRole` SHALL be idempotent: when an identical (user, role, scope) grant already exists, the call SHALL leave it untouched and MUST NOT create a duplicate.

#### Scenario: 重复加入不产生重复授权
- **GIVEN** an existing (owner, community A) grant for the user
- **WHEN** the join flow is retried for the same community and ownership
- **THEN** no duplicate role row is created and the existing grant is unchanged

#### Scenario: 并发重复加入只产生一条授权
- **GIVEN** two concurrent join requests for the same user, community A, and ownership choice
- **WHEN** both reach `AssignRole`
- **THEN** exactly one (owner/tenant, community A) grant exists and no partial state is observable

### Requirement: 退出小区即撤销授权

On a successful `LeaveCommunity`, user-service SHALL synchronously call permission-service `RevokeRole` to remove the community's owner/tenant role and scope.

#### Scenario: 退出后撤销该小区角色与 scope
- **GIVEN** a user holding (owner, community A) who leaves community A
- **WHEN** the leave flow succeeds
- **THEN** the owner/tenant role scoped to A is revoked

#### Scenario: 仅撤销目标小区，不影响其他小区
- **GIVEN** a user holding owner roles in both A and B, who leaves only A
- **WHEN** the leave flow succeeds
- **THEN** the owner role scoped to B remains intact

#### Scenario: 撤销失败则退出不成功
- **GIVEN** a leave where the membership is deactivated but the synchronous `RevokeRole` fails
- **WHEN** the leave flow completes
- **THEN** the leave is reported as failed or the grant is retried, leaving no active-scope-without-membership state

### Requirement: 授权变更即时失效缓存

Join and leave SHALL invalidate the affected user's scope and permission caches (`perm:scopes:{userId}:{scopeType}`, `perm:user:{userId}`) so the new authorization takes effect immediately.

#### Scenario: 退出后立即不能发布
- **GIVEN** a user who just left community B
- **WHEN** the user immediately attempts to publish in B
- **THEN** denied (scope cache was deleted, not reused)

#### Scenario: 加入后立即可发布
- **GIVEN** a user who just joined community A
- **WHEN** the user immediately attempts to publish in A
- **THEN** allowed (scope grant and its cache are in effect)
