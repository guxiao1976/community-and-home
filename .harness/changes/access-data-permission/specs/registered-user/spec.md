# 注册用户基角色（registered_user）Specification

## Purpose

在角色体系顶层引入正式基角色 `registered_user`：任何用户注册成功后即自动获得该角色（永久有效、空数据范围、browse 类权限），作为「尚未加入任何小区」状态的身份载体。它保证权限模型对每个用户都有一个确定的基座：注册用户无数据范围、不可发布、不可见小区内部内容；一旦加入小区获得社区角色，能力 = registered_user 权限 ∪ 社区角色权限。

## Requirements

### Requirement: registered_user 角色定义

permission-service SHALL define `registered_user` as a formal system role (`is_system=1`, `status=1`) whose granted permissions are browse-only, configured through `rel_role_permission` (no permission shortcut; per [[is-system-no-permission-shortcut]]).

#### Scenario: registered_user 为正式系统角色
- **GIVEN** the role catalog in permission-service
- **WHEN** `registered_user` is queried
- **THEN** it exists as a system role with browse-only permissions wired via `rel_role_permission`

#### Scenario: browse 权限经配置生效
- **GIVEN** a registered_user
- **WHEN** `CheckPermission` checks a browse permission
- **THEN** allowed through the normal role→permission path (no field-based shortcut)

### Requirement: 注册自动分配

When a user is created via user-service `CreateUser`, the system SHALL automatically assign `registered_user` with role status `2` (permanently valid) and an `empty` data scope.

#### Scenario: 新注册用户自动获得基角色
- **GIVEN** a new phone number registering for the first time
- **WHEN** user creation succeeds
- **THEN** a rel_user_role record for `registered_user` exists with status `2` and no community scope

#### Scenario: 注册失败不产生角色分配
- **GIVEN** a registration attempt that fails (e.g., phone already registered)
- **WHEN** the create-user flow aborts
- **THEN** no `registered_user` assignment occurs for a non-existent account

### Requirement: 空数据范围约束

A `registered_user` SHALL have an `empty` data scope: SHALL NOT publish to any community and SHALL NOT see community-internal content. The user's effective permission set SHALL be the union of `registered_user` browse permissions and any community-role permissions later acquired.

#### Scenario: 注册用户不可发布
- **GIVEN** a registered_user with no community membership
- **WHEN** the user attempts to publish to any community
- **THEN** denied (empty data scope)

#### Scenario: 权限集合叠加
- **GIVEN** a registered_user who later joins community A and gains the owner role
- **WHEN** the effective permission set is evaluated
- **THEN** it contains browse permissions (from registered_user) and publish permissions (from owner)

### Requirement: 分配幂等

Assigning `registered_user` SHALL be idempotent: an existing identical (user, role, scope) assignment MUST NOT be duplicated, so a retried registration does not create duplicate role rows.

#### Scenario: 重复分配不产生重复
- **GIVEN** a user who already holds the `registered_user` assignment
- **WHEN** the registration/assignment flow runs again
- **THEN** no duplicate `registered_user` row is created and the existing one is unchanged

#### Scenario: 幂等重分配后缓存不残留旧授权
- **GIVEN** a user whose `registered_user` assignment is re-applied while permission/scope caches are warm
- **WHEN** the assignment completes
- **THEN** the caches are invalidated so no stale or doubled grant effect is observed on subsequent checks
