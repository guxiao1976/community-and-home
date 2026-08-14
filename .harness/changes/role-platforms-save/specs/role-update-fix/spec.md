# 角色更新修复（Role Update Fix） Specification

## Purpose

修复编辑角色时因 RPC 业务错误未被 API 层检查而触发的空指针 panic（HTTP 500），并将 `responsex.ToError(grpcResp.Base)` 基类检查推广为 perm 包全部 API logic 的类级规则（grep 核实 11 文件，见 REQ-UPDATE-3），为 `toRoleInfo` 增加 nil 防御，整类消除「业务错误静默吞掉 + 空指针 panic」隐患；同时落实「系统角色可编辑配置字段（name/description/platforms/sort_order/permission_ids）、status 仍拦截（原子拒绝）、权限走独立分配流程」的字段级编辑策略（决策 D1/D5）。

## Requirements

### Requirement: REQ-UPDATE-1 — API 层 UpdateRole 必须检查 RPC Base

The UpdateRole API logic SHALL, after invoking the `UpdateRole` RPC, check `grpcResp.Base` via `responsex.ToError`; when the business code is non-zero, it MUST return the resulting Go error and MUST NOT proceed to `toRoleInfo(grpcResp.Role)`.

- **GIVEN** the RPC returns `Base.Code=60004`（系统角色状态不可修改）and `Role=nil`
- **WHEN** the API `UpdateRole` logic processes the response
- **THEN** the logic returns a Go error carrying code 60004（经 `responsex.ToError`），and no panic occurs

- **GIVEN** the RPC returns a successful `Base`（code 0）with a populated `Role`
- **WHEN** the API `UpdateRole` logic processes the response
- **THEN** the logic returns `UpdateRoleResp` with `Role` converted via `toRoleInfo`

- **GIVEN** the RPC returns `Base.Code=60001`（角色不存在）
- **WHEN** the API `UpdateRole` logic processes the response
- **THEN** the logic returns a Go error carrying code 60001（HTTP 层呈现统一 `{code,msg}` 响应体，不再返回 500）

### Requirement: REQ-UPDATE-2 — toRoleInfo 必须对 nil 输入有防御

The `toRoleInfo` helper SHALL accept a nil proto `Role` and MUST return a zero-value `types.RoleInfo` without panicking.

- **GIVEN** `toRoleInfo` is called with `r == nil`
- **WHEN** the conversion executes
- **THEN** it returns `types.RoleInfo{}`（零值），and no panic occurs

- **GIVEN** `toRoleInfo` is called with a fully-populated proto `Role`
- **WHEN** the conversion executes
- **THEN** it returns a `types.RoleInfo` with all mapped fields（id/code/name/description/isSystem/status/sortOrder/permissions/timestamps）

### Requirement: REQ-UPDATE-3 — Base 检查类级规则推广至 perm 包全部 API logic

Every API logic method in the `perm` package that invokes a permission RPC SHALL check `grpcResp.Base` via `responsex.ToError` immediately after the RPC call, mirroring the existing `listroleslogic.go:54` pattern; on non-zero code it MUST return the resulting Go error and MUST NOT dereference any response field（`Role`/`Roles`/`PermissionCodes`/`Permissions`）or treat the RPC as success. This is a class-level rule; the verified-complete affected set（grep，2026-08-14）is:

- **空指针 panic 类（deref RPC 响应）**：`createrolelogic.go`（`CreateRole`）、`updaterolelogic.go`（`UpdateRole`）、`getrolelogic.go`（`GetRole`）、`getrolepermissionslogic.go`（`GetRole` → `Role.Permissions`）、`getuserpermissionslogic.go`（`GetUserPermissions` → `PermissionCodes`）、`getuserroleslogic.go`（`GetUserRoles` → `Roles`）、`listpermissionslogic.go`（`ListPermissions` → `Permissions`）
- **静默吞业务错误类（丢弃响应，仅查 Go err）**：`deleterolelogic.go`（`DeleteRole`）、`assignrolepermissionslogic.go`（`AssignRolePermissions`）、`assignuserrolelogic.go`（`AssignUserRole`）、`revokeuserrolelogic.go`（`RevokeUserRole`）
- **已有正确模式（不改）**：`listroleslogic.go`（`ListRoles`）

- **GIVEN** the `CreateRole` RPC returns `Base.Code=60006`（角色编码已存在）and `Role=nil`
- **WHEN** the API `CreateRole` logic processes the response
- **THEN** it returns a Go error carrying code 60006，and no panic occurs

- **GIVEN** the `GetRole` RPC returns `Base.Code=60001`（角色不存在）with `Role=nil`
- **WHEN** the API `GetRolePermissions` logic processes the response
- **THEN** it returns a Go error carrying code 60001，and does not dereference `Role.Permissions`

- **GIVEN** the `DeleteRole` RPC returns `Base.Code=60004`（系统角色不可删除）
- **WHEN** the API `DeleteRole` logic processes the response
- **THEN** it returns a Go error carrying code 60004（而非静默返回成功）

- **GIVEN** the `CreateRole` RPC returns a successful `Base`（code 0）with a populated `Role`
- **WHEN** the API `CreateRole` logic processes the response
- **THEN** it returns `CreateRoleResp` with `Role` converted via `toRoleInfo`，no error（正向路径，避免 ToError 过度拦截）

### Requirement: REQ-UPDATE-4 — 系统角色字段级编辑策略

The `UpdateRole` RPC logic SHALL allow updating `name`/`description`/`platforms`/`sort_order` **and `permission_ids`** for system roles（`is_system=1`）, and MUST reject a `status` change for system roles with business code 60004. This change narrows 60004's `UpdateRole` semantics from「系统角色不可修改」(whole-request block) to「系统角色状态不可修改」(status-only block), so the message for this path is updated accordingly（the `DeleteRole` path's 60004 message「系统角色不可删除」is unchanged; the pre-existing same-code-different-meaning issue is tracked separately）. The `status` rejection is **atomic**: when a system-role request contains `status`（regardless of other fields）, the RPC MUST return 60004 **before applying any field** and **before platforms validation**, so no partial write（name/platforms/sort_order/permission_ids）is persisted. **Precedence**: the system-role `status` check SHALL be evaluated before `platforms` value validation（REQ-PLAT-4）; if a request hits both（system role + `status` + invalid platform value）, 60004 wins and is returned（structural protection first）. **Permission_ids boundary (D1 alignment)**: D1's「权限仍走独立分配流程」refers to the front-end interaction path（permission page → `AssignRolePermissions`）; the backend `UpdateRole` RPC's `permission_ids` replacement is the transport for that flow, and the HTTP `UpdateRole` API's `permissionIds` field remains exposed and unchanged（no new front-end path added）. For system roles, `permission_ids` SHALL NOT be blocked by `is_system`（so `AssignRolePermissions` works on system roles）, consistent with REQ-PLAT-8's platforms-preservation invariant.

- **GIVEN** a system role（`is_system=1`, e.g. owner）and an `UpdateRoleRequest` containing only `name` and `platforms`
- **WHEN** the RPC `UpdateRole` logic executes
- **THEN** the update succeeds（`Base.Code=0`），name 与 platforms 落库，且 DB `status` 保持不变

- **GIVEN** a system role（`is_system=1`）and an `UpdateRoleRequest` containing `status=0` **together with** editable fields（e.g. `name="改名"`, `platforms=["mobile"]`）
- **WHEN** the RPC `UpdateRole` logic executes
- **THEN** it returns `Base.Code=60004`（系统角色状态不可修改）**before applying any field**，and the DB row is not modified（`name`/`platforms`/`sort_order`/`permission_ids` 均不落库，无部分写入）

- **GIVEN** a system role（`is_system=1`）and an `UpdateRoleRequest` containing `status=0` **together with** an invalid platform value（`platforms: ["web"]`）
- **WHEN** the RPC `UpdateRole` logic executes
- **THEN** it returns `Base.Code=60004`（the system-role status gate is the first front gate and takes precedence over the 60008 platforms-validation; per the validation-order clause）and no field is persisted

- **GIVEN** a system role（`is_system=1`）and an `AssignRolePermissions` request carrying only `permission_ids`
- **WHEN** the flow invokes `UpdateRole` RPC
- **THEN** the permission replacement succeeds（`Base.Code=0`），and the role's existing `platforms` are preserved（REQ-PLAT-8）

- **GIVEN** a non-system role（`is_system=0`）and an `UpdateRoleRequest` containing `status=0`
- **WHEN** the RPC `UpdateRole` logic executes
- **THEN** the status update succeeds（`Base.Code=0`），状态落库

- **GIVEN** the web front-end edits a system role via the enabled「编辑」button（sends name/description/platforms only）
- **WHEN** the request reaches the backend
- **THEN** the update succeeds（不再 500），front-end button remains enabled（决策 D1 明确前端编辑按钮保持可用）
