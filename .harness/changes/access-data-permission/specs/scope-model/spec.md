# Scope 三态与祖先链统一规则 Specification

## Purpose

定义数据权限的统一模型：每个用户按 scope_type 持有「global / 限定 / 空」三态之一的授权节点集合，并建立唯一的祖先链命中判据（目标 t 被授权集合 S 覆盖 ⟺ A(t) ∩ S ≠ ∅）。该判据与授权规模/城市大小解耦（A(t) 固定 ≤6 节点），使"能发到哪"由后端权威计算、跨服务一致复用。覆盖树由 master-data 提供祖先链解析，授权来源可插拔（本次仅成员资格，商户广告模型兼容不实现）。

## Requirements

### Requirement: scope 三态

permission-service SHALL maintain, for each user, an authorized data-scope set per scope_type with exactly one of three states: `global` (all scope nodes), `limited` (a concrete set of scope nodes), or `empty` (no scope nodes).

> `[NEEDS CLARIFICATION] 待阶段3架构定稿`：`empty` 与 `global` 在 `rel_user_role` 上的具体存储表示（如 scope_type 取值 `none` vs `global`、scope_id=0）由架构设计确定；本 spec 仅约束行为——`empty` 必须表现为"空授权集合"且不得等价于 `global`（见 REQ-1.2）。

#### Scenario: 注册用户持有空数据范围
- **GIVEN** a user holding only the `registered_user` role with no community membership
- **WHEN** permission-service resolves that user's community data scope
- **THEN** the scope set is `empty` (no scope_id returned)

#### Scenario: 审核员持有全局数据范围
- **GIVEN** an auditor whose rel_user_role carries scope_type=`global`
- **WHEN** permission-service resolves that user's community data scope
- **THEN** the scope set is `global`, covering every scope node

#### Scenario: 业主持有限定数据范围
- **GIVEN** a resident with rel_user_role (owner, scope_type=`community`, scope_id=A)
- **WHEN** permission-service resolves the community scope
- **THEN** the scope set is `limited` to {A}

### Requirement: 空不等价于 global

permission-service MUST NOT treat an `empty` scope set as `global`. A user whose scope set is `empty` MUST be denied any data-scope-gated operation (such as publishing or reading community content), even though `empty` could be mistaken for "everything" by a naive implementation.

#### Scenario: 空数据范围不产生任何放行
- **GIVEN** a registered_user whose community scope set is `empty`
- **WHEN** any consumer asks whether that user may publish to community X
- **THEN** the request is denied; the empty set must never match any target

#### Scenario: 空与 global 是不同状态
- **GIVEN** two users, one with `empty` scope and one with `global` scope
- **WHEN** each attempts the same data-scope-gated operation on the same community
- **THEN** the `empty`-scope user is denied while the `global`-scope user is allowed

### Requirement: 祖先链命中统一判据

For a target scope node t, the target SHALL be considered covered by authorized set S if and only if A(t) ∩ S ≠ ∅, where A(t) is the ancestor chain {t, t.parent, ..., root} over the coverage tree (administrative-division tree with `residential_area` leaves), bounded at 6 nodes. A multi-target operation SHALL pass only if every target is covered.

#### Scenario: 目标在授权集合中 → 覆盖
- **GIVEN** a resident authorized to community A only (S={A})
- **WHEN** the ancestor chain A(A) = {A, div(A), street, county, city, province} is intersected with S
- **THEN** A(A) ∩ S = {A} ≠ ∅ → the target is covered

#### Scenario: 目标不在授权集合 → 拒绝
- **GIVEN** the same resident authorized to community A only (S={A})
- **WHEN** checking target community B where A is not an ancestor of B
- **THEN** A(B) ∩ S = ∅ → the target is not covered

#### Scenario: 多目标全部通过才放行
- **GIVEN** a publish targeting [A1, A2] where A1 is covered but A2 is not
- **WHEN** the operation is evaluated
- **THEN** the operation is denied because A2 fails, even though A1 passes

### Requirement: 覆盖树与祖先链解析（master-data）

master-data-service SHALL provide an RPC that resolves any scope node (an `md_administrative_division` id or an `md_residential_area` id) to its ancestor chain including itself (bounded at 6 nodes). The division tree SHALL be cached in full (division data changes infrequently) to serve this RPC efficiently. A residential_area SHALL be linked into the tree via its `community_div_id`.

The consumption direction for this RPC SHALL be fixed as `community-hub → permission-service → master-data`: `AssertPublishScope` (permission-service) resolves target ancestor chains through this RPC; community-hub-service SHALL NOT call master-data directly for scope resolution.

#### Scenario: 小区解析出完整祖先链
- **GIVEN** a residential_area R whose community_div_id → street → county → city → province
- **WHEN** master-data resolves the ancestor chain of R
- **THEN** the chain returned is {R, div(R), street, county, city, province}

#### Scenario: 行政区划解析自身及祖先
- **GIVEN** a division D at city level (parent = province)
- **WHEN** master-data resolves the ancestor chain of D
- **THEN** the chain returned is {D, province}

#### Scenario: 覆盖树拓扑变更后解析一致
- **GIVEN** a division node is re-parented under a new parent
- **WHEN** master-data resolves a descendant's ancestor chain after the change
- **THEN** the chain reflects the new topology (tree cache refreshed, not stale)

### Requirement: 授权来源可插拔

The authorized-set model SHALL be source-agnostic: an authorization source (membership for residents now; merchant ad-orders as a future source) feeds scope nodes into the same set S, and the coverage rule in REQ-1.3 SHALL NOT change when a new source is added. Merchant advertising SHALL NOT be implemented in this change, but the model MUST NOT preclude a later ad-order source (scope + validity period).

#### Scenario: 成员资格作为当前唯一来源
- **GIVEN** a resident who joined community A
- **WHEN** their authorized set S is computed
- **THEN** S contains A, contributed by the membership-derived rel_user_role scope

#### Scenario: 未来商户来源可接入同一模型
- **GIVEN** a future merchant ad-order authorizing scope nodes with a validity period
- **WHEN** that source is plugged into the authorized-set model
- **THEN** the coverage rule (ancestor-chain hit) remains unchanged; no new coverage logic is required

### Requirement: 读操作按数据范围过滤

Community read and list endpoints SHALL filter results to the caller's visible communities using the scope returned by `GetDataScopes(user_id, "community")`. A user whose scope set is `empty` SHALL receive no community content.

#### Scenario: 业主只读到所属小区内容
- **GIVEN** an owner of community A only
- **WHEN** the user lists community content
- **THEN** only content whose community_id is in {A} is returned

#### Scenario: 空数据范围读到空列表
- **GIVEN** a registered_user with `empty` community scope
- **WHEN** the user lists community content
- **THEN** an empty list is returned (no community-internal content leaked)

### Requirement: global 例外

Data-scope-gated access SHALL pass for a user who holds a `global` scope or the `sys_admin` role, enabling cross-community viewing and auditing.

#### Scenario: global 审核员跨小区查看
- **GIVEN** an auditor holding scope_type=`global`
- **WHEN** the auditor views content in a community outside any limited scope
- **THEN** access is allowed

#### Scenario: sys_admin 跨小区操作放行
- **GIVEN** a user holding the `sys_admin` role
- **WHEN** the user accesses data across communities
- **THEN** data permission passes (functional permission still applies)

### Requirement: scope 缓存与失效

Scope caches (e.g., `perm:scopes:{userId}:{scopeType}`) SHALL be invalidated on any role/scope change — join, leave, or verification-status change — so that the cached authorized set is never older than the underlying grants. The coverage-tree cache SHALL be invalidated when division/community topology changes.

#### Scenario: 退出后立即失效
- **GIVEN** a user who just left community B
- **WHEN** the user's scope is resolved immediately after leaving
- **THEN** B is absent from the returned scope (the stale cache was deleted, not reused)

#### Scenario: 加入后立即生效
- **GIVEN** a user who just joined community A
- **WHEN** the user's scope is resolved immediately after joining
- **THEN** A is present in the returned scope
