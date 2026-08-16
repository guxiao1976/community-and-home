# Content Post Permission Capability Specification

## Purpose

定义通用图文发布的发布权限契约：GetPublishPermission（本期新实现，D5/Q5——代码中本不存在）返回 can_publish + 可发布角色，前端不做权限判断；**物业管理员 property_admin 保留本小区发布权（D6/Q6，推翻原 notice 设计剔除 property_admin 的 D26）**——发布角色集 = {grid_worker 多小区 / community_admin 选社区展开 / property_admin 本小区 / committee 本小区}，业主/租户/商家只读；can_publish 判定对齐权限服务 level-2 语义（status=2 且 verified_at NOT NULL 且未过期，经 GetUserRoles RPC，禁止直读 rel_user_role）；权限种子变更（grid_worker 授 421 + **421 置 min_verf_level 由现有 0 提升到 2** + **撤销 owner/tenant 的 421 绑定 (1,421)/(5,421)，保留 435/436** + property_admin 保留 421，REVISION——与 notice 的 D26 回收相反，且补齐 notice 种子未覆盖的业主只读闭环）；division→community 授权集解析行为结论化（REVISION，fail-closed）。涉及 community-hub-service（判定 + 入口显隐）、permission-service（种子 + 角色状态权威）。

## Requirements

### Requirement: REQ-CPP-1 — GetPublishPermission 返回 can_publish + 可发布角色（含 property_admin，D6）

The system SHALL provide a GetPublishPermission RPC (newly implemented this change, D5/Q5) that, for the authenticated user, returns a boolean `can_publish` and the list of publishable roles. `can_publish` SHALL be true if and only if the user holds at least one publish-capable role with an active association matching the permission-service level-2 semantics: role status = 2 (已认证) AND `verified_at` NOT NULL AND not expired (`expires_at == 0 OR expires_at > now`, evaluated on the `GetUserRoles` RPC output, NOT by writing SQL against `rel_user_role`). **Publish-capable roles SHALL be: grid_worker, community_admin, property_admin, committee (D6/Q6 property-admin-exclusion — property_admin SHALL retain its own-community publish right via the generic component, overturning the prior notice design's D26 removal).** owner/tenant/merchant and any role not in the publish-capable set SHALL receive `can_publish=false`. The role-state data SHALL be obtained via the permission-service `GetUserRoles` RPC; community-hub-service SHALL NOT read `rel_user_role` directly (跨服务仅 gRPC). The write-path gate SHALL be the same role-status level (min_verf_level=2 on 421, REQ-CPP-3), so a stale/forged `can_publish` cannot widen publish scope.

#### Scenario: 网格员可发布（level-2 通过）
- **GIVEN** an authenticated grid_worker whose community-scoped role association is status=2 with a non-null verified_at and not expired
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has can_publish=true and the roles list contains grid_worker

#### Scenario: property_admin 可发布本小区（D6）
- **GIVEN** an authenticated property_admin (role 2) with a level-2 active association (this change retains its 421, D6, unlike the notice D26 removal)
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has can_publish=true and the roles list contains property_admin (own-community publish right retained via the generic component)

#### Scenario: committee 可发布本小区（正向角色场景）
- **GIVEN** an authenticated committee member whose community-scoped role association is level-2 (status=2, verified_at NOT NULL, not expired)
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has can_publish=true and the roles list contains committee

#### Scenario: status=2 但 verified_at 为 NULL（level-2 不通过）
- **GIVEN** an authenticated grid_worker whose role association is status=2 but verified_at is NULL
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has can_publish=false (level-2 requires verified_at NOT NULL)

#### Scenario: 角色已过期
- **GIVEN** a user whose only publish-capable role association has expired (expires_at in the past) or status != 2
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has can_publish=false

#### Scenario: 业主/租户/商家只读
- **GIVEN** an authenticated owner (or tenant, or merchant) with only read-only roles
- **WHEN** the user calls GetPublishPermission
- **THEN** the response has can_publish=false and roles does not include any publish-capable role

### Requirement: REQ-CPP-2 — 各角色发布范围（property_admin 本小区 / 业主只读 / division 展开结论化 REVISION）

The system SHALL enforce per-role publish scope for content posts: grid_worker MAY publish to multiple communities all within their `community` scope; **community_admin SHALL govern exactly ONE division (community_div scope), and its publish target SHALL be ALL approved communities under that single governing division, expanded automatically by the backend at publish time via master-data `GetResidentialAreasByDivision(community_div_id, status=1)` (REVISION-10/A2 — a community_admin manages one community/division containing multiple residential areas; the frontend does NOT select a division and CreateContentPost does NOT take a division_id — the backend derives the admin's single governing division from scope and expands it; this overturns the earlier "community_admin MAY select a division / CreateContentPost.division_id" design of REUSE:notice-D6);** **property_admin SHALL publish only to its own single community (D6)**; committee SHALL publish only to its own single community; owner/tenant/merchant SHALL be read-only and MUST NOT have a publish entry. The scope unit SHALL always be `md_residential_area.id`; `community_type` (小区 vs 村) SHALL NOT alter scope representation. **The division→community authorization behavior SHALL be concluded as follows (REVISION — promoted from design-gate to spec contract): a community_admin's division grant (`community_div` scope) SHALL resolve into the community authorization set used by `AssertPublishScope` at publish time (each expanded approved community is a publishable target); if the resolution fails to authorize any expanded target, the request SHALL fail closed with 080006 (no partial snapshot). The detailed judging-logic correctness SHALL still be verified by permission-service unit/integration tests during design (REUSE:notice REV-17), but the behavioral contract — resolution outcome and failure semantics — SHALL NOT block implementation (REVISION).**

#### Scenario: 物业管理员发布本小区成功（D6）
- **GIVEN** an authenticated property_admin whose own community is C1
- **WHEN** the property_admin submits CreateContentPost targeting C1
- **THEN** the publish is allowed (property_admin retains own-community publish right) and the post is persisted with content_post_scope (post, C1)

#### Scenario: 物业管理员发布其他小区被拒
- **GIVEN** an authenticated property_admin whose own community is C1
- **WHEN** the property_admin submits CreateContentPost targeting C2 (outside own community)
- **THEN** the publish is rejected (080006 data-permission denial, out-of-scope target) and no post is created

#### Scenario: 社区管理员自动展开其管辖社区下所有小区（A2，REVISION-10）
- **GIVEN** a community_admin governing exactly one division D1 (containing communities C1, C2, all approved), and the division→community authorization resolution (this change's contract) yields C1, C2 as the admin's publishable community set
- **WHEN** the community_admin submits CreateContentPost (no division_id in the request — the backend derives the admin's single governing division D1 from scope and expands it)
- **THEN** the backend expands D1 to its concrete approved communities via GetResidentialAreasByDivision, validates each via AssertPublishScope, and writes content_post_scope rows for C1 and C2 as a fixed snapshot

#### Scenario: division 展开目标越权 → 整体拒绝（fail-closed，080006）
- **GIVEN** a community_admin whose division grant resolves to an authorization set that does NOT include a community C3 which the division expansion yields
- **WHEN** the community_admin submits CreateContentPost with division_id that expands to [C1, C2, C3] where C3 is outside the resolved authorization set
- **THEN** `AssertPublishScope` denies C3 and the entire request is rejected with 080006; no partial snapshot is written (REVISION — defined failure semantics for the resolved set)

#### Scenario: 业主无发布能力
- **GIVEN** an authenticated owner/tenant/merchant user
- **WHEN** the user invokes CreateContentPost or checks publish entry
- **THEN** the system SHALL NOT expose a publish entry (can_publish=false) and a direct CreateContentPost attempt SHALL be rejected with 080002 (no publish role at the function-permission layer)

### Requirement: REQ-CPP-3 — 权限种子对齐（property_admin 保留 421 + grid_worker 授 421 + 撤销 owner/tenant 421 + min_verf_level 0→2，REVISION）

The system SHALL align the permission seeds for the content-post write path (REVISION — the change set below is complete and includes the two bindings the prior seed change omitted): **(1) property_admin SHALL retain the `community:notice:create-api` (421) binding (D6 — this change DOES NOT revoke property_admin 421, overturning the prior notice D26 revocation); (2) grid_worker SHALL be granted 421 (it previously lacked it); (3) the write-path role-state gate SHALL be enforced by raising `min_verf_level` on 421 from the existing value 0 to 2** — this is a behavior change to `UPDATE sys_permission SET min_verf_level = 0 WHERE code IN ('community:notice:create-api', ...)` at init_permissions.sql:201-202 (REVISION, item 3); **(4) owner(1) and tenant(5) SHALL be REVOKED from 421: delete the `(1, 421)` and `(5, 421)` bindings from `rel_role_permission` (init_permissions.sql:252-253, keep the `(1,435)/(1,436)/(5,435)/(5,436)` bindings)** (REVISION, item 3) — so owner/tenant SHALL NOT hold the create permission (read-only) and a level-2 owner can no longer pass the functional-permission layer; (5) the read-path permission codes (list/detail/marquee/publish-permission/data-scopes) SHALL be bound to the full mobile role set per the notice read-path matrix (REUSE:notice REQ-PP-4 — 422/423/424/426/427/428). All seed changes SHALL follow the fail-closed semantics: a mobile endpoint without a sys_permission def + role binding returns 403, so every content-post REST endpoint needs its code bound.

#### Scenario: property_admin 保留 421 后仍可发布（D6 种子生效）
- **GIVEN** the seed retains property_admin's 421 binding and sets min_verf_level=2 on the create code
- **WHEN** an authenticated, level-2-verified property_admin invokes the create endpoint for its own community
- **THEN** the request passes the functional-permission layer (421 present, level-2 gate passed) and proceeds to scope validation (own community → allowed)

#### Scenario: grid_worker 授 421 后可发布（此前 080002）
- **GIVEN** the seed grants grid_worker the 421 create permission and the user is level-2 verified
- **WHEN** the grid_worker invokes the create endpoint for an in-scope community
- **THEN** the request passes the functional-permission layer (previously it was rejected with 080002 for missing 421)

#### Scenario: 撤销后业主/租户直调创建被拒（(1,421)/(5,421) 移除生效）
- **GIVEN** the seed change has removed the `(1,421)`/`(5,421)` bindings (owner/tenant 421 revoked, 435/436 retained) and 421's min_verf_level raised to 2
- **WHEN** an authenticated owner/tenant (even a level-2-verified one) invokes the create endpoint directly (bypassing the UI)
- **THEN** the request is rejected with 080002 (no publish role at the function layer — the revoked 421 binding + elevated min_verf_level closes the prior gap where a level-2 owner could pass the functional layer); the request never reaches data-scope validation

#### Scenario: 未认证/已过期发布者被拒（min_verf_level=2 提升生效）
- **GIVEN** a publisher who holds 421 but is not level-2 verified (unverified or expired association)
- **WHEN** the publisher invokes the create endpoint
- **THEN** the request is rejected at the functional-permission layer (min_verf_level=2 gate — a behavior change from the prior level-0; consistent with can_publish=false in REQ-CPP-1)

#### Scenario: 业主直调创建被拒（功能权限 080002）
- **GIVEN** an authenticated owner/tenant who does not hold 421 (post-revocation)
- **WHEN** the user invokes the create endpoint directly (bypassing the UI)
- **THEN** the request is rejected with 080002 (no publish role at the function layer); the request never reaches data-scope validation

## 服务职责边界

- **community-hub-service**: GetPublishPermission 实现（经 permission GetUserRoles RPC，level-2 判定，可发布角色集含 property_admin，D6）；发布入口显隐的 can_publish 载体；CreateContentPost 写路径（功能权限 421 + scope 校验）
- **permission-service**: GetUserRoles 角色状态权威；AssertPublishScope 统一判据（越权 060007 → 消费方 080006）；GetDataScopes 范围选项；种子（property_admin 保留 421、grid_worker 授 421、**撤销 (1,421)/(5,421) 保留 435/436**、**421 min_verf_level 0→2**、读码绑定全移动端角色，REVISION）；division→community 授权集解析（行为结论化 + design 阶段单测验证）
- **api-proto**: GetPublishPermission RPC 契约（can_publish + publishable_roles）
