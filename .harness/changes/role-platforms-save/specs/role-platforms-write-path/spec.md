# 允许登录端写链路补全（Role Platforms Write Path） Specification

## Purpose

补全允许登录端 `platforms`（值域 pc/mobile）的完整写链路：proto 请求字段 → `make generate` → API 类型与 API logic 透传 → RPC logic 经 `joinPlatforms` 写入 `sys_role.platforms`，含值域校验（非法端 60008 拒绝）、空值语义（空=显式清空，fail-open）、platforms 变更触发角色缓存失效；同时闭合 HTTP 读链路缺口（`RoleInfo.Platforms` 透传）并顺带修复 `sort_order` 未写库的潜伏缺陷（决策 D2/D3/D4/D6）。

## Requirements

### Requirement: REQ-PLAT-1 — Proto 请求必须新增 platforms 字段

The `CreateRoleRequest` and `UpdateRoleRequest` messages in `api/permission/v1/permission.proto` SHALL add a `repeated string platforms` field with a comment documenting the value domain（`pc`/`mobile`）、empty-is-fail-open semantics、and the rejection of illegal values（60008）. The proto file header error-code comment block SHALL also document `060008 — 非法登录端`（existing convention：code additions are recorded in the header block，see 2026-08-12 CHANGELOG for 060007）. After `make generate`, the generated Go code SHALL expose the field, and the change SHALL be backward-compatible（`buf breaking` passes）.

- **GIVEN** the proto file adds `repeated string platforms` to `CreateRoleRequest`（field 6）and `UpdateRoleRequest`（field 7），and the header error-code block adds `060008 — 非法登录端`
- **WHEN** `cd api-proto && make ci` executes
- **THEN** `buf lint`、`buf breaking`、`buf generate` all pass，gen/go exposes `Platforms []string`，and CHANGELOG.md records the change

- **GIVEN** an existing client sends `CreateRoleRequest`/`UpdateRoleRequest` without the new `platforms` field
- **WHEN** the request is received
- **THEN** the field is empty（`len==0`），interpreted as「no platforms declared」under the fail-open semantics（向后兼容，不破坏既有调用方）

### Requirement: REQ-PLAT-2 — API 类型与 logic 必须透传 platforms

The API types `CreateRoleReq`/`UpdateRoleReq` SHALL add a `Platforms []string` field（JSON tag `platforms`）. The `CreateRole` API logic SHALL set `grpcReq.Platforms = req.Platforms`; the `UpdateRole` API logic SHALL always set `grpcReq.Platforms = req.Platforms`（包括空列表，落实 D3「API 层始终透传」）.

- **GIVEN** an HTTP create-role request with `platforms: ["pc", "mobile"]`
- **WHEN** the API `CreateRole` logic builds the gRPC request
- **THEN** `grpcReq.Platforms == ["pc", "mobile"]`

- **GIVEN** an HTTP update-role request with `platforms: []`（显式清空）
- **WHEN** the API `UpdateRole` logic builds the gRPC request
- **THEN** `grpcReq.Platforms` is an empty slice（非 nil，且不等于「未传」的省略路径——始终透传）

- **GIVEN** an HTTP update-role request omits `platforms`
- **WHEN** the API `UpdateRole` logic builds the gRPC request
- **THEN** `grpcReq.Platforms` is still set（nil/empty 均透传），RPC 视为「空=清空」

### Requirement: REQ-PLAT-3 — RPC 层必须写入并无条件覆盖 platforms

The `CreateRole` RPC logic SHALL persist `joinPlatforms(in.Platforms)` into `sys_role.platforms` on insert. The `UpdateRole` RPC logic SHALL unconditionally overwrite `existing.Platforms = joinPlatforms(in.Platforms)`（D3：空串 = 允许所有端，fail-open）.

- **GIVEN** a new role is created with `platforms: ["pc", "mobile"]`
- **WHEN** the RPC `CreateRole` logic executes
- **THEN** `sys_role.platforms` is stored as `"pc,mobile"`

- **GIVEN** an existing role（platforms `"pc,mobile"`）is updated with `platforms: []`
- **WHEN** the RPC `UpdateRole` logic executes
- **THEN** `existing.Platforms` is overwritten to `""`（空串），read path returns empty slice, and the front-end displays「全部」

- **GIVEN** the DB update fails during `UpdateRole`
- **WHEN** the RPC logic executes
- **THEN** it returns a Go error（不做部分成功），and `Base` is not set as success

### Requirement: REQ-PLAT-4 — RPC 层必须校验 platforms 值域并拒绝非法值

The `CreateRole` and `UpdateRole` RPC logic SHALL validate every value in `in.Platforms` against the domain `{pc, mobile}`. Any value outside the domain MUST be rejected with business code 60008（「非法登录端」）and the rejection is **atomic**：the ENTIRE create/update MUST fail（`Base.Code=60008`），no field（name/description/status/sort_order/platforms/permission_ids）SHALL be persisted（no partial write）. Validation SHALL run before any field is applied. This requirement governs platforms-domain validation only; the system-role `status` gate and its precedence over 60008 are specified in REQ-UPDATE-4（`role-update-fix` capability, applied after this one per `.change.yaml` order）— dependency is one-directional（update-fix layers the field-level policy on top of this capability's platforms handling）. For `CreateRole`, no system-role check applies（new roles are non-system），so 60008 is the only front gate. An empty list or nil SHALL pass validation（fail-open）. Duplicate values SHALL be deduplicated before persistence. The 60008 code SHALL be expressed as a named constant（e.g. `CodeInvalidPlatform`）in the RPC layer rather than a raw literal（per [[error-code-literal-bypasses-qa-gate]]）, and SHALL be registered in `permission.proto`'s header error-code comment block as `060008 — 非法登录端`（per REQ-PLAT-1）.

- **GIVEN** a request with `platforms: ["web"]`
- **WHEN** the RPC `CreateRole`/`UpdateRole` logic executes
- **THEN** it returns `Base.Code=60008`, and the role is not created / the platforms not updated

- **GIVEN** a request with `platforms: []` or `platforms` omitted
- **WHEN** the RPC `CreateRole`/`UpdateRole` logic executes
- **THEN** validation passes（fail-open semantics）

- **GIVEN** a request with `platforms: ["pc", "pc", "mobile"]`
- **WHEN** the RPC logic executes
- **THEN** the persisted value is deduplicated to `"pc,mobile"`

- **GIVEN** an update request carrying `name="改名"` **together with** `platforms: ["web"]`
- **WHEN** the RPC `UpdateRole` logic executes
- **THEN** it returns `Base.Code=60008` and **no field is persisted**（`name` 不落库，原子拒绝，无部分写入）

### Requirement: REQ-PLAT-5 — sys_role Update SQL 必须写回 sort_order

The `SysRoleModel.Update` SQL SHALL include `sort_order = ?`（alongside role_name/description/status/platforms）so that a `sort_order` update is persisted. The `UpdateRole` RPC logic SHALL set `existing.SortOrder` only when `in.SortOrder` is present.

- **GIVEN** an `UpdateRoleRequest` with `sort_order=5`
- **WHEN** the RPC logic and the model `Update` execute
- **THEN** `sys_role.sort_order` is updated to `5`

- **GIVEN** an `UpdateRoleRequest` omitting `sort_order`
- **WHEN** the RPC logic executes
- **THEN** `existing.SortOrder` is left unchanged and the update preserves the original value

### Requirement: REQ-PLAT-6 — HTTP 读链路必须透出 platforms

The API type `RoleInfo` SHALL add `Platforms []string`（JSON tag `platforms`）. The `toRoleInfo` helper SHALL pass through `r.Platforms`. `UserRoleInfo.Role`（reusing `RoleInfo`）SHALL therefore also expose `platforms`. An empty platforms SHALL serialize as `[]`（non-nil empty slice）, not `null`, so the front-end's `Array.isArray(row.platforms)` / `.length > 0` checks see an empty array and render「全部」（the nil-defensive zero value from REQ-UPDATE-2 MUST be normalized to `[]` for the `platforms` field）.

- **GIVEN** `GetRole`/`ListRoles`/`GetUserRoles` RPC returns a `Role` with `Platforms: ["mobile"]`
- **WHEN** the API converts it via `toRoleInfo`
- **THEN** the HTTP response includes `"platforms": ["mobile"]`

- **GIVEN** a role whose platforms is empty（`splitPlatforms` returns `[]`）
- **WHEN** `toRoleInfo` converts it
- **THEN** the HTTP response includes `"platforms": []`（non-nil empty array, not `null`；前端据此展示「全部」）

### Requirement: REQ-PLAT-7 — platforms 变更必须触发角色缓存失效

The `UpdateRole` RPC logic SHALL invalidate the permission/data-scope caches of every user holding the modified role（reuse the existing `invalidateRoleCache` pattern）when a role update（including a platforms change）succeeds.

- **GIVEN** a role with two holders（user 1001、1002）has its `platforms` updated
- **WHEN** the `UpdateRole` RPC logic completes
- **THEN** `perm:user:*` and `perm:scopes:*` cache keys for both users are deleted from Redis

- **GIVEN** a role has no holders
- **WHEN** cache invalidation runs
- **THEN** it is a no-op and does not return an error

### Requirement: REQ-PLAT-8 — 权限独立更新路径不得误清 platforms

The permission-assignment flow（`AssignRolePermissions` API → `UpdateRole` RPC with `PermissionIds`）SHALL preserve the role's existing `platforms`; a permission-only update MUST NOT clear `sys_role.platforms` to empty as a side effect（consistency invariant of D3's unconditional-overwrite semantics）. Implementation: the `AssignRolePermissions` API logic SHALL fetch the role's current `platforms`（via `GetRole` RPC）and include them in the `UpdateRoleRequest` it constructs（alongside `PermissionIds`）. The `GetRole` response SHALL be Base-checked（`responsex.ToError`）before use. **Failure semantics**: if reading the role's current platforms fails（`GetRole` returns a non-zero Base, e.g. 60001 角色不存在, or a Go error）, the flow MUST abort and return the error—it MUST NOT proceed to `UpdateRole` with empty/omitted platforms（which D3 would persist as a clear, un-doing the role's end restriction）.

- **GIVEN** a role with `platforms = "mobile"` and an admin re-configures its permissions via the permission page（sends only `PermissionIds`）
- **WHEN** `AssignRolePermissions` invokes `UpdateRole`
- **THEN** `sys_role.platforms` remains `"mobile"`（not cleared to `""`）

- **GIVEN** the `GetRole` call inside `AssignRolePermissions` returns `Base.Code=60001`（角色不存在）
- **WHEN** the flow processes the response
- **THEN** it aborts with a Go error carrying code 60001 and does NOT call `UpdateRole`（no risk of clearing platforms）

- **GIVEN** a role with `platforms = "mobile"` and an admin re-configures its permissions（a generic custom or system role; the platforms-preservation invariant holds independent of `is_system`）
- **WHEN** `AssignRolePermissions` invokes `UpdateRole`
- **THEN** `sys_role.platforms` remains `"mobile"`（whether the role is system or custom is governed by REQ-UPDATE-4 in the `role-update-fix` capability, applied after this capability per `.change.yaml` order）

- **GIVEN** a permission-only update on a role with empty platforms
- **WHEN** the flow executes
- **THEN** platforms remains empty（fail-open）and no error occurs
