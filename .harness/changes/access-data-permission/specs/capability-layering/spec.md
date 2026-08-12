# 能力分层（Capability Layering）Specification

## Purpose

修订 `rbac-design.md` §2.5 的鉴权规则，把「认证状态」从"能不能用"的开关改为"能力分层"的阶梯：`sys_permission.min_verf_level` 定义某权限需要多高的认证层级，`CheckPermission` 按权限的层级要求与授予角色（rel_user_role）的认证状态进行判定。核心效果：未认证业主/租户可执行 `min_verf_level=0` 的权限（如发布），但 `=2` 的高信任权限（如业委会选举）必须已认证；认证期间（待审）保持未认证能力，不中断正常发布。

## Requirements

### Requirement: 权限新增 min_verf_level 属性

permission-service SHALL maintain a `min_verf_level` attribute on each permission (`sys_permission`), taking one of the values `0` (role + data scope suffice) or `2` (verification required), defaulting to `0`.

#### Scenario: 发布类权限为层级 0
- **GIVEN** a publish permission such as `lostfound:create`
- **WHEN** its `min_verf_level` is resolved
- **THEN** the value is `0` (role + data scope suffice)

#### Scenario: 选举类权限为层级 2
- **GIVEN** a high-trust permission such as `committee:election:vote`
- **WHEN** its `min_verf_level` is resolved
- **THEN** the value is `2` (verification required)

### Requirement: CheckPermission 按层级判定认证状态

`CheckPermission` SHALL evaluate a matched permission's `min_verf_level` against the certification status of the rel_user_role records that grant it, using the following aggregate rule: a level-`0` permission is satisfied by any granting role at an active status (`0`=未认证, `1`=待审, or `2`=已认证); a level-`2` permission is satisfied only by a granting role at status `2` (已认证); a role at status `3` (已驳回) or `4` (已过期) satisfies neither level. When a permission is granted through multiple rel_user_role records, the check SHALL aggregate across all of them by taking the maximum satisfied level — the permission SHALL be granted if at least one granting role satisfies the required `min_verf_level`. This SHALL supersede the prior rule in `rbac-design.md` §2.5 where status 0/1 could not authenticate at all.

#### Scenario: 未认证业主可发布
- **GIVEN** an owner in community A whose rel_user_role status is `0` (未认证), holding the publish permission (`min_verf_level=0`)
- **WHEN** `CheckPermission` checks the publish permission
- **THEN** the request is allowed (role + data scope suffice)

#### Scenario: 未认证业主不可选举
- **GIVEN** the same unverified owner holding `committee:election:vote` (`min_verf_level=2`)
- **WHEN** `CheckPermission` checks the election permission
- **THEN** the request is denied (granting role not at status 2)

#### Scenario: 认证业主可选举
- **GIVEN** an owner whose rel_user_role status is `2` (已认证), holding `committee:election:vote`
- **WHEN** `CheckPermission` checks the election permission
- **THEN** the request is allowed

#### Scenario: 待审状态保留未认证能力
- **GIVEN** an owner whose rel_user_role status is `1` (待审), holding a publish permission (`min_verf_level=0`)
- **WHEN** `CheckPermission` checks the publish permission during the review window
- **THEN** the request is allowed (verification pending does not suspend publishing)

#### Scenario: 已过期角色不授予任何层级
- **GIVEN** a role whose rel_user_role status is `4` (已过期)
- **WHEN** `CheckPermission` checks any permission granted through that role
- **THEN** the request is denied for both level-0 and level-2 permissions

#### Scenario: 多角色叠加——未认证与已认证并存时取最高层级
- **GIVEN** a user holding `committee:election:vote` through two roles: an unverified owner role (status `0`) and a verified committee role (status `2`)
- **WHEN** `CheckPermission` checks the election permission
- **THEN** allowed (the verified committee role satisfies level-`2`, aggregated max wins)

#### Scenario: 多角色叠加——全部未认证则层级 2 拒绝
- **GIVEN** a user holding `committee:election:vote` through two roles, both at status `0` (unverified)
- **WHEN** `CheckPermission` checks the election permission
- **THEN** denied (no granting role satisfies level-`2`)

#### Scenario: registered_user（空数据范围角色）参与命中
- **GIVEN** a user holding browse permission via `registered_user` (status `2`, empty data scope) and publish permission via an unverified owner role (status `0`, community A)
- **WHEN** `CheckPermission` checks the publish permission
- **THEN** allowed by the owner role (level-`0`); the `registered_user` role contributes browse permission but no data scope, and does not block the owner role's grant

### Requirement: 能力分层能力矩阵

The system SHALL enforce the following capability matrix: browse is available to registered users, unverified residents, and verified residents; publishing (within quota) is available to unverified and verified residents but not registered users; participating in committee elections is available only to verified residents.

#### Scenario: 注册用户无发布能力
- **GIVEN** a registered_user with `empty` data scope
- **WHEN** the user attempts to publish
- **THEN** denied (no data scope; publishing requires a scope, not just a level-0 permission)

#### Scenario: 三层能力与层级一一对应
- **GIVEN** browse, publish, and election operations of min_verf_level 0, 0, and 2 respectively
- **WHEN** each is evaluated for registered / unverified / verified residents
- **THEN** browse passes for all three, publish passes only for unverified and verified, election passes only for verified

### Requirement: 认证状态变更即时生效

When a user's rel_user_role status changes (e.g., from `0`/`1` to `2`, or to `3`/`4`), permission and scope caches SHALL be invalidated so the new capability level takes effect immediately.

#### Scenario: 认证通过后立即解锁选举
- **GIVEN** an owner whose status changes from `0` to `2`
- **WHEN** `CheckPermission` is called for `committee:election:vote` immediately after the change
- **THEN** allowed (cache was invalidated; no stale "unverified" result is returned)

#### Scenario: 驳回后立即收回发布能力
- **GIVEN** an owner whose status changes from `1` to `3` (驳回)
- **WHEN** `CheckPermission` is called for a publish permission immediately after the change
- **THEN** denied (cache was invalidated)
