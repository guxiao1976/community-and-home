# Notice Publish Permission Capability Specification

## Purpose

定义「可发布判定」的后端单一契约（Q3 已拍板：新增 GetPublishPermission RPC）：后端返回可发布角色 + `can_publish` 标志驱动【我的】页发布入口显隐，前端不做权限判断。**可发布角色收敛为 {grid_worker, community_admin, committee}（D26）——property_admin 被剔除出移动端可发布角色集（can_publish=false + 种子回收 421）**。`can_publish` 判定对齐权限服务 level-2 语义（status=2 且 verified_at NOT NULL 且未过期，`expires_at==0 OR > now` 基于 RPC 输出），经 `GetUserRoles` RPC 获取角色状态（禁止直读 rel_user_role 表）。**写路径与入口判定共用同一角色状态门槛**（REVISION-1 闭环）：发布落库校验复用已有 `AssertPublishScope`（数据范围）+ 421 功能权限 `min_verf_level=2`（功能层强制已认证，方案 A）——未认证/已过期的发布者即使持有数据范围也无法落库，与 can_publish=false 一致。本 capability 还覆盖权限种子对齐（REQ-PP-4）：grid_worker 授 `community:notice:create-api` 并置 min_verf_level=2、**property_admin/owner/tenant 回收该权限（D26 + Q 决策）**，属行为回归。

## Requirements

### Requirement: REQ-PP-1 — GetPublishPermission 返回可发布角色 + can_publish（level-2 判定）

The system SHALL provide a GetPublishPermission RPC that, for the authenticated user, returns a boolean `can_publish` and the list of publishable roles. `can_publish` SHALL be true if and only if the user holds at least one publish-capable role with an active association matching the permission-service level-2 semantics: role status = 2 (已认证) AND `verified_at` NOT NULL AND not expired. **The expiry check SHALL be evaluated on the `GetUserRoles` RPC output (`expires_at == 0 OR expires_at > now`), NOT by writing SQL against `rel_user_role`** (the DB encodes NULL expiry as RPC `expires_at=0`; "0=永久"). **Publish-capable roles SHALL be: grid_worker, community_admin, committee (D26 — property_admin is excluded from the mobile publish-capable set, so it SHALL receive can_publish=false and its 421 is revoked per REQ-PP-4).** owner/tenant/merchant and any role not in the publish-capable set SHALL receive `can_publish=false`. **sys_admin SHALL also receive `can_publish=false` for the mobile entry (D16)**: although sys_admin holds 421 via the explicit seed binding for role 8 (init_permissions.sql grants sys_admin all permissions including 421), the mobile app is a community-member surface and sys_admin's publish surface is the admin console (PC), which is out of scope for this change; **the sys_admin write path is NOT additionally blocked in this change** (D16, per REQ-PP-3 the write gate is 421 + min_verf_level; sys_admin satisfying it passes — the mobile entry is hidden but the API remains callable; the admin console is the intended surface). The role-state data SHALL be obtained via the permission-service `GetUserRoles` RPC; community-hub-service SHALL NOT read `rel_user_role` directly (跨服务仅 gRPC).

#### Scenario: 网格员可发布（level-2 通过）
- **GIVEN** an authenticated grid_worker whose community-scoped role association is status=2 with a non-null verified_at and not expired
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has `can_publish=true` and the roles list contains grid_worker

#### Scenario: status=2 但 verified_at 为 NULL（level-2 不通过）
- **GIVEN** an authenticated grid_worker whose role association is status=2 but verified_at is NULL (registered_user-adjacent case)
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has `can_publish=false` (level-2 requires verified_at NOT NULL)

#### Scenario: 角色已过期
- **GIVEN** a user whose only publish-capable role association has expired (`expires_at` in the past) or status != 2
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has `can_publish=false`

#### Scenario: 业主/租户/商家只读
- **GIVEN** an authenticated owner (or tenant, or merchant) with only read-only roles
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has `can_publish=false` and roles does not include any publish-capable role

#### Scenario: property_admin 不在移动端可发布角色集（D26）
- **GIVEN** an authenticated property_admin (role 2) who previously held 421, revoked by REQ-PP-4 in this change
- **WHEN** the user calls GetPublishPermission from the mobile surface
- **THEN** the response has `can_publish=false` (property_admin excluded from the mobile publish-capable set, D26) and roles does not include property_admin; the【我的】page hides the publish entry (see REQ-NM-4)

#### Scenario: sys_admin 不显示移动端发布入口（写路径不额外拦截）
- **GIVEN** an authenticated sys_admin who holds 421 via the explicit seed binding for role 8 (init_permissions.sql grants sys_admin all permissions including 421)
- **WHEN** the user calls GetPublishPermission from the mobile surface
- **THEN** the response has `can_publish=false` (sys_admin not in the mobile publish-capable set, D16); the sys_admin's publish surface is the admin console, out of scope for mobile; the sys_admin write path is NOT additionally blocked by this change (a direct CreateNotice satisfying 421+level-2 passes, consistent with REQ-PP-3)

#### Scenario: 未登录
- **GIVEN** an unauthenticated request to GetPublishPermission
- **WHEN** the call reaches the gRPC/API layer
- **THEN** the authentication middleware returns UNAUTHENTICATED and the handler is never reached; the frontend treats the auth failure as `can_publish=false` and hides the publish entry (single defined behavior, no dual branch)

### Requirement: REQ-PP-2 — 前端不判断权限（入口显隐由 can_publish 驱动）

The 【我的】page SHALL show the「发布通知」entry if and only if the `can_publish` flag returned by GetPublishPermission is true. The frontend SHALL NOT implement role-code-based permission logic to decide publish-entry visibility.

#### Scenario: 入口显隐完全由后端标志驱动
- **GIVEN** a grid_worker whose GetPublishPermission returns can_publish=true
- **WHEN** the user opens【我的】page
- **THEN** the「发布通知」entry is rendered; for an owner (can_publish=false) the entry is not rendered

#### Scenario: 隐藏入口但直接请求发布仍被拦截
- **GIVEN** an owner with can_publish=false (entry hidden) — and, per REQ-PP-4, the owner no longer holds the `community:notice:create-api` functional permission
- **WHEN** the user directly invokes CreateNotice (bypassing the UI)
- **THEN** the backend rejects the request with 080002 at the functional-permission layer (no publish role); frontend hiding is NOT a security boundary

### Requirement: REQ-PP-3 — 发布写路径的角色状态门槛与 AssertPublishScope（后端单一判定）

The CreateNotice path SHALL require the publisher to satisfy the same role-status gate as `can_publish`: the publisher SHALL hold a publish-capable role at level-2 (status=2 AND `verified_at` NOT NULL AND not expired) AND the target communities SHALL pass `AssertPublishScope(user_id, target_community_ids)` before persistence. `can_publish` is an entry-visibility hint; it SHALL NOT replace per-target scope validation or the role-status gate at write time. **The write-path role-status gate SHALL be enforced by setting `min_verf_level=2` on the `community:notice:create-api`(421) functional permission (REQ-PP-4, 方案 A — the single committed mechanism)**: permission-service `grantSatisfiedLevel` returns level-2 only for status=2 AND verified_at NOT NULL AND unexpired, so the function-permission layer (CheckPermission / PermMiddleware) already expresses the full level-2 gate; with 421 at min_verf_level=2, an unverified or expired publisher SHALL be rejected with 080002 before data-scope validation, making the write path consistent with `can_publish=false`. No separate write-path role-status re-check is required; the seed change (REQ-PP-4) is the enforcement point.

#### Scenario: 写路径不信任入口标志
- **GIVEN** a user with can_publish=true whose data scope covers only C1
- **WHEN** the user submits CreateNotice targeting C2 (outside scope)
- **THEN** AssertPublishScope denies and the write is rejected with 080006; a stale/forged can_publish flag cannot widen publish scope

#### Scenario: 角色未认证/已过期但仍持数据范围 → 写路径拒绝（REVISION-1 新增）
- **GIVEN** a grid_worker whose community-scope grant is status=0 (未认证) or expired, so `grantSatisfiedLevel` = 0 or -1 < 2, and whose data scope covers C1
- **WHEN** the grid_worker directly invokes CreateNotice targeting C1 (bypassing the hidden entry)
- **THEN** the function-permission layer rejects the request with 080002 (421 requires min_verf_level=2, the unverified grant satisfies only level-0); the write never reaches data-scope validation — consistent with GetPublishPermission returning can_publish=false; a hidden entry cannot substitute for write-path enforcement

#### Scenario: AssertPublishScope 服务不可用
- **GIVEN** permission-service (AssertPublishScope) is temporarily unavailable
- **WHEN** a user submits CreateNotice
- **THEN** the system fails closed: the write is rejected with a dependency error rather than being allowed without scope validation

### Requirement: REQ-PP-4 — 权限种子对齐（grid_worker 授 421 + 421 置 min_verf_level=2 / owner·tenant 回收 421）

The permission-service seed (scripts/init_permissions.sql) SHALL be updated in this change so the publish-capability matrix matches the product rules: (1) grant grid_worker (role_id=4) the `community:notice:create-api`(421) functional permission — currently grid_worker has no 421, so per rbac-design §6.5 its notice publish is blocked (✅ 拦截) at the function layer before AssertPublishScope; (2) **set `min_verf_level=2` on 421** — currently §4.2 explicitly sets `min_verf_level=0` for `community:notice:create-api` (未认证即可发布); the change flips it to 2 so the function layer enforces the level-2 role-status gate (REVISION-1, 方案 A, consistent with REQ-PP-1/REQ-PP-3); the 435 lostfound permission SHALL remain at min_verf_level=0 (out of scope); (3) **revoke the `community:notice:create-api`(421) binding from property_admin (role_id=2, D26)** — currently seed §3 grants `(2, 421)` (物业发布通知); property_admin is excluded from the mobile publish-capable set, so its 421 SHALL be revoked to make the write path consistent with can_publish=false (per REQ-PP-3, the function layer is the enforcement point); (4) revoke the `community:notice:create-api`(421) binding from owner (role_id=1) and tenant (role_id=5) — currently seed §4.8 deliberately grants owner/tenant 421 (comment: 未认证业主/租户即可发布). This is a registered behavior regression: property_admins who could previously publish notices (per §3) and owners/tenants who could per §4.8 become read-only for notices, and unverified role holders can no longer publish notices (previously possible at level-0). The lostfound create (435) and contacts upsert (436) bindings for owner/tenant, and any other property_admin permissions, are OUT OF SCOPE for this change and remain unchanged.

#### Scenario: 网格员功能权限放行（不再被 080002 拦截）
- **GIVEN** an authenticated grid_worker with a verified (status=2, verified_at NOT NULL) community data scope
- **WHEN** the grid_worker invokes CreateNotice
- **THEN** the functional-permission layer passes (grid_worker holds 421 with min_verf_level=2 satisfied), and the request proceeds to `AssertPublishScope` data-scope validation (越权仍 080006)

#### Scenario: 未认证网格员发布被拒（min_verf_level=2 生效）
- **GIVEN** a grid_worker whose 421-bearing role association is status=0 (未认证) or status=2 with verified_at NULL, but who holds a community data scope
- **WHEN** the grid_worker invokes CreateNotice
- **THEN** the functional-permission layer rejects with 080002 (421 now requires min_verf_level=2; the grant satisfies only level-0) — the write-path gate matches can_publish=false from GetPublishPermission

#### Scenario: owner/tenant 回收 421 后直接发布被拒
- **GIVEN** an authenticated owner whose only community permission binding is read/browse (421 revoked per this change)
- **WHEN** the owner directly invokes CreateNotice
- **THEN** the functional-permission layer rejects with 080002 (no publish role); the owner cannot reach data-scope validation

#### Scenario: property_admin 回收 421 后直接发布被拒（D26）
- **GIVEN** an authenticated property_admin (role 2) whose 421 was revoked by this change
- **WHEN** the property_admin directly invokes CreateNotice on the mobile surface
- **THEN** the functional-permission layer rejects with 080002 (no publish role); the property_admin cannot reach data-scope validation, consistent with can_publish=false from REQ-PP-1

#### Scenario: 种子变更的幂等与回归登记
- **GIVEN** the permission-service seed script is re-run against a database that already has the old bindings (property_admin/owner/tenant holding 421, 421 at min_verf_level=0)
- **WHEN** the seed migration executes
- **THEN** grid_worker's 421 binding is added, 421's `min_verf_level` becomes 2, and property_admin/owner/tenant's 421 binding is removed idempotently; the change is recorded as a behavior regression (原 §3「物业发布通知」/§4.8「未认证业主/租户即可发布通知」→ 物业/业主/租户只读 + 发布需已认证) in the change notes and reflected in rbac-design.md §6.5 acceptance matrix

## 服务职责边界

- **community-hub-service**: GetPublishPermission 实现（判定可发布角色 + 角色状态经 permission `GetUserRoles` RPC）、CreateNotice 内 AssertPublishScope 消费 + 写路径角色状态门槛（经 421 min_verf_level=2 功能层强制）
- **permission-service**: `GetUserRoles`（角色状态权威，供 level-2 判定）、`AssertPublishScope`（数据范围权威）、`GetDataScopes`（范围选项）；本变更修改种子（REQ-PP-4：授 421 + 置 min_verf_level=2 + **收 property_admin/owner/tenant 421（D26）**）；**AssertPublishScope 判据逻辑是否变更（如 division→community 授权集解析）由 design gate（REV-17）定夺，本 capability 不预先断言「不改判据逻辑」**
- **api-proto**: `community/v1` 新增 GetPublishPermission RPC 及请求/响应消息
- **web/mobile**: 消费 can_publish 驱动入口（不实现权限逻辑）
