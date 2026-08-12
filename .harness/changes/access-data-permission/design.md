# Design: 数据权限核心（access-data-permission）

> 阶段：3 架构设计。输入：proposal.md + 5 specs（已 3/3 APPROVED）+ 权威设计 `docs/specs/access-control-design.md`（§1.4/§3/§5/§8）。
> 本文件定稿 3 个 `[NEEDS CLARIFICATION]` 边界项，并落地评审遗留的 5 项 SHOULD FIX。
> Proto 变更（5+1 项）由阶段 4 全局 Owner 执行，设计只定契约，不实现。

---

## 0. 评审遗留处置（SHOULD FIX 落地总表）

| # | SHOULD FIX 来源 | 处置 | 落地位置 |
|---|----------------|------|---------|
| S1 | structure v2 #1 | **依赖方向**：community-hub 读祖先链一律 `community-hub → permission-service → master-data`，community-hub 不得直连 master-data 做 scope 解析。修订 access-control-design.md §8 L273（去掉「祖先链」） | §7、tasks 0.0 |
| S2 | clarity v2 #2 / coverage v2 关联 | **registered_user status=2 与「已认证」数值相撞**：聚合规则改为数据驱动——level-2 需 `status==2 AND verified_at IS NOT NULL`；registered_user 种子 `verified_at=NULL`、权限集恒 browse-only（level-0），从数据不变量上杜绝相撞，不引入角色名短路 | §3.1.4、§5.1.1、§9 |
| S3 | structure v2 #3 | **多角色 scope 合并优先级 REQ 化**：global 支配 → limited 并集 → empty；GetDataScopes（读）与 AssertPublishScope（写）共用同一判定函数 | §4.4、§5.2、REQ-A |
| S4 | coverage v2 #1 / structure v2 INFO #2 | **moderation 审核状态回调身份**：服务间回调（无用户 JWT）以预留**系统身份**（`system_user_id=0`，种子 sys_admin + global scope）执行数据权限校验；reverse-lookup 内容→community_id 校验存在，不用内容作者 scope | §4.9、§5.7 |
| S5 | structure v2 #2 | **.change.yaml `common_change_required=false` 与 proposal「common 可能」不一致**：定稿为 `false`——ScopeRef/AssertPublishScope 经 api-proto gen/go 生成代码消费，本变更**不触碰 common/**。同步修订 proposal 影响范围表 | §0.1、proposal.md 修订 |
| S6 | coverage v2 #2/#3、clarity v2 INFO | status=3 对 level-2 断言、scope 缓存认证变更触发、ancestor ≤6 截断、未知节点安全拒绝 → 并入接口契约与任务测试矩阵 | §4.7、§10、tasks 测试矩阵 |

### 0.1 proposal 一致性修订（common）
`common_change_required` 定稿 **false**。理由：`ScopeRef` 与 `AssertPublishScope`/`ResolveScopeAncestors` 的客户端桩均在 `api-proto`（permission/v1、masterdata/v1）由 `make generate` 产出，消费方（community-hub/user/master-data/permission）各用自己的 gen 包，无需 community-common 客户端包装。proposal 影响范围表「common 可能」行改为「本变更不触碰 common/」。

---

## 1. 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| scope 三态语义、祖先链统一判据、registered_user 角色定义、min_verf_level、权限/scope 缓存失效 | **permission-service（权威）** | 数据所有权：rel_user_role/sys_permission/sys_role 全在 permission 库 |
| 「scope 节点 → 祖先链」RPC、行政区划整树缓存 | **master-data-service** | md_administrative_division / md_residential_area 数据所有者 |
| CreateUser 自动分配 registered_user；Join/Leave 编排 Assign/RevokeRole（非权威） | **user-service** | 注册/成员资格流程编排方，判定仍委托 permission-service |
| 写接口挂 AssertPublishScope、publisher_id 取 JWT、读列表按 GetDataScopes 过滤、moderation 回调校验 | **community-hub-service（消费方）** | 内容数据所有者，校验由 permission-service 权威计算 |
| 全部 Proto 变更（permission/masterdata/user 三个 package） | **api-proto（阶段 4 Owner 执行）** | 硬约束 #1/#2，子 Agent 禁止修改 api-proto |

---

## 2. 记忆注入报告（Step 1.5）

匹配 10 个，注入 8 个，不适用 2 个（[[qwen-3b-unsuitable-for-moderation]]、[[ai-model-template-variable-content]] 与权限域无关）。

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[is-system-no-permission-shortcut]] | §3.1.4 / §5.1 / §10 | registered_user 权限经 rel_role_permission 配置；sys_admin 全局访问建模为 `rel_user_role scope_type='global'` 授权，禁止角色名短路；moderation 系统身份同样走 grant 判定 |
| [[permission-seed-api-path-must-match-routes]] | §3.1.3 / §10 | 新增 registered_user browse 权限与发布权限的 seed `path` 必须与实际 REST 路由一致（`GET:/api/community/notices` 等），否则 CheckPermission 永久失败 |
| [[proto-jstype]] | §4 / §9 | 所有 int64 ID 字段 `[jstype = JS_STRING]`；JoinCommunity/ScopeRef/祖先链响应均遵守 |
| [[grpc-only-comms]] | §7 | community-hub → permission-service → master-data 全走 gRPC，禁止直连 DB |
| [[migration-must-execute]] | §3.1 / tasks | sys_permission.min_verf_level 列、rel_user_role 唯一索引迁移必须执行后再编码 |
| [[verify-api-before-calling]] | §5.5 / §10 | publisher_id/userId 取自 JWT 前先确认 claims 结构；新增 REST 路由先存在再挂权限 |
| [[grpc-timeout-layers]] | §4.6 | AssertPublishScope 内嵌 master-data ResolveScopeAncestors，gRPC 三层超时对齐（≤500ms） |
| [[redis-cache-soft-delete]] | §6 | scope/权限缓存失效与软删除联动，失效收敛到 grant 变更处理器内部 |

---

## 3. 数据模型

### 3.1 permission-service（permission 库）

#### 3.1.1 `sys_permission` 新增列
```sql
ALTER TABLE sys_permission ADD COLUMN min_verf_level TINYINT NOT NULL DEFAULT 0 COMMENT '能力层级: 0=持角色+数据范围即可, 2=需已认证(默认0)';
```
- 取值仅 `0` / `2`（`1` 为未来层级保留，本变更不使用）。
- 透出到 `Permission` proto 消息（列表/详情响应）。
- 种子数据标注：发布类权限（`lostfound:create`/`notice:create` 等）`min_verf_level=0`；选举类（`committee:election:vote`）`=2`。

#### 3.1.2 `rel_user_role` scope 三态物理表示（NEEDS CLARIFICATION #1 定稿）

**存储方案**：`scope_type` 增加两个保留取值，`scope_id=0` 约定为「非实体占位」：

| scope 状态 | scope_type | scope_id | 示例 |
|-----------|-----------|----------|------|
| `global`（全局放行） | `'global'` | `0` | 审核员 / sys_admin / moderation 系统身份 |
| `limited`（限定节点） | `'community'`/`'building'`/`'unit'`/`'grid'` | 对应实体 id（雪花 id，恒非 0） | 业主 owner@A → `'community', A` |
| `empty`（无数据范围） | `''`（空串） | `0` | 仅 registered_user 基角色 |

- **`''` ≠ `global`**：`FindScopesByUserId` / 判定逻辑必须显式区分。`empty` 行（`scope_type=''`）对任何 scopeType 查询**零贡献**；`global` 行对任何 scopeType 查询**全放行**。杜绝「空当 global」灾难（REQ-1.2）。
- `FindScopesByUserId` 需排除 `scope_id=0`（空/全局占位都不进 limited 并集）。

**唯一索引（幂等基础）**：
```sql
ALTER TABLE rel_user_role ADD UNIQUE KEY uk_user_role_scope (user_id, role_id, scope_type, scope_id);
```
- 支撑 registered_user 分配幂等（REQ-3.4）与 Join 自动授权幂等（REQ-4.2 并发只产生一条）。
- 开发阶段无存量数据，可直接加索引（决策 §12-3）。

#### 3.1.3 `sys_role` 新增 `registered_user` 基角色
```sql
-- init_permissions.sql 追加（is_system=1 仅保护，权限经 rel_role_permission 配置）
INSERT IGNORE INTO sys_role (id, role_code, role_name, description, is_system, status, sort_order, created_by)
VALUES (9, 'registered_user', '注册用户', '注册即自动分配的基角色：browse-only、空数据范围、永久有效', 1, 1, 5, 0);
```
- **rel_role_permission 种子**：仅 browse 类（读权限），seed path 与 `web/mobile` 实际路由一致：
  `GET:/api/community/notices`、`GET:/api/community/lostfound`、`GET:/api/community/contacts`。遵循 [[permission-seed-api-path-must-match-routes]]。
- **rel_user_role 种子形态**（CreateUser 时自动分配，非种子 SQL）：
  `(user_id, role_id=9, scope_type='', scope_id=0, status=2, verified_at=NULL, expires_at=NULL)`。

#### 3.1.4 registered_user status=2 与「已认证」区分（SHOULD FIX S2 定稿）
- `registered_user` 保留 `status=2`（per spec REQ-3.2「永久有效」），但聚合规则改为**数据驱动**（见 §5.1.1）：
  **level-2 满足条件 = `status==2 AND verified_at IS NOT NULL AND 未过期`**。
- registered_user 种子 `verified_at=NULL`（永不认证）→ 即使 status=2 也**只能满足 level-0**，未来误配任何 level-2 权限到 registered_user 都不会被放行——从数据不变量上杜绝数值相撞，不引入角色名/字段短路（[[is-system-no-permission-shortcut]]）。
- 权限集守卫：registered_user 的 rel_role_permission 恒为 browse-only（level-0）；新增 level-2 权限到基角色需显式设计评审。

#### 3.1.5 系统审核身份（moderation 回调 / 全局内部调用）
- 预留 `system_user_id = 0`（常量，config 可配），在 `rel_user_role` 种子一行：`(user_id=0, role_id=sys_admin, scope_type='global', scope_id=0, status=2)`。
- 校验走**同一条 grant 判定路径**（global → 放行），无代码级 `if userId==0 → allow` 短路（[[is-system-no-permission-shortcut]]）。
- 消费者按社区、不留存量数据。

### 3.2 user-service（无表结构变更）
- `JoinCommunityRequest` 新增 `ownership` 字段（§4.2）。membership 表不落权属（Leave 时对 owner+tenant 双调 RevokeRole 幂等覆盖）。

### 3.3 master-data-service / community-hub-service（无表结构变更）
- master-data 复用 `md_administrative_division`（parent_id/path）与 `md_residential_area`（community_div_id）表，新增**内存/Redis 整树缓存**服务于 `ResolveScopeAncestors`，不改表。
- community-hub 无 schema 变更。

---

## 4. 接口契约

### 4.1 空 scope 表示 → GetDataScopes 三态响应（NEEDS CLARIFICATION #1 落地）

`GetDataScopesResponse` 从「裸 scope_id 列表」升级为「状态 + 列表」：

```proto
// permission/v1
enum DataScopeState {
  DATA_SCOPE_STATE_UNSPECIFIED = 0;
  DATA_SCOPE_STATE_EMPTY   = 1;  // 空：无数据范围（仅 registered_user 等）
  DATA_SCOPE_STATE_LIMITED = 2;  // 限定：scope_ids 为并集
  DATA_SCOPE_STATE_GLOBAL  = 3;  // 全局：覆盖所有节点，scope_ids 忽略
}
message GetDataScopesResponse {
  common.v1.BaseResp base = 1;
  repeated int64 scope_ids = 2 [jstype = JS_STRING]; // 保留（现网字段号不变，wire 兼容）
  DataScopeState state = 3;                          // 新增三态
}
```
- 缓存 key `perm:scopes:{userId}:{scopeType}` 存 JSON `{"state":"empty|global|limited","ids":[int64]}`（§6）。

### 4.2 JoinCommunity 权属 API 形状（NEEDS CLARIFICATION #2 定稿）

`JoinCommunityRequest` 新增 `ownership` 枚举；**不与房屋注册合并**（BindResidence 保持独立）：

```proto
// user/v1
enum CommunityOwnership {
  COMMUNITY_OWNERSHIP_UNSPECIFIED = 0; // 非法值，后端拒绝（10040 参数校验失败）
  COMMUNITY_OWNERSHIP_OWNED   = 1;     // 自有 → owner 角色
  COMMUNITY_OWNERSHIP_RENTED  = 2;     // 租住 → tenant 角色
}
message JoinCommunityRequest {
  int64 user_id = 1 [jstype = JS_STRING];
  int64 community_id = 2 [jstype = JS_STRING];
  int32 building = 3;
  int32 unit = 4;
  int32 room = 5;
  CommunityOwnership ownership = 6; // 新增，必填
}
```
- 后端逻辑：`ownership==UNSPECIFIED` → 400/10040 拒绝；`OWNED` → `AssignRole(owner, community, A, status=0)`；`RENTED` → `AssignRole(tenant, community, A, status=0)`。
- role_code→role_id 复用 user-service 既有 `roleMapper`（ListRoles 缓存 5min，见 `rpc/internal/logic/user/role_mapper.go`），无需新增 RPC。

### 4.3 错误码对齐（NEEDS CLARIFICATION #3 定稿）

| 场景 | 错误码 | 归属 | 说明 |
|------|:---:|------|------|
| 数据权限拒绝（API 面） | **080006** | community-hub | 「目标小区超出发布者数据范围，无发布权限」。沿用 08xxxx 命名空间（设计文档 `80006` 即 `08 0006`） |
| 数据权限拒绝（RPC 面） | **060007** | permission-service | AssertPublishScope 内部拒绝；community-hub 映射 `060007 → 080006`（060006 已登记「角色编码已存在」，避免契约冲突） |
| 参数校验（ownership 缺失） | 10040 | user-service | 复用现有 |
| 无发布权限（功能权限） | 080002 | community-hub | 复用现有 |

### 4.4 GetDataScopes（改造）
- 输入不变：`(user_id, scope_type)`。
- 判定合并优先级（REQ-A，见 §5.2）：`global` 支配 → `limited` 并集 → `empty`。
- 读穿缓存：HIT 直接返回；MISS 计算后写 `perm:scopes:{userId}:{scopeType}` JSON + EXPIRE 30min。

### 4.5 ScopeRef（新增）
```proto
// permission/v1
message ScopeRef {
  string scope_type = 1; // 本变更仅 "community"
  int64 scope_id = 2 [jstype = JS_STRING];
}
```

### 4.6 AssertPublishScope（新增 RPC）
```proto
// permission/v1 — @auth: INTERNAL, @idempotent: true, @timeout: 500
rpc AssertPublishScope(AssertPublishScopeRequest) returns (AssertPublishScopeResponse);
message AssertPublishScopeRequest {
  int64 user_id = 1 [jstype = JS_STRING];
  repeated ScopeRef targets = 2;
}
message AssertPublishScopeResponse {
  common.v1.BaseResp base = 1;
  bool allowed = 2; // false = 无数据权限（060007）
}
```
**逻辑**（统一判据封装，消费方不再重复实现祖先链规则）：
1. 解析用户 community scope → `{state, ids}`（复用 §4.4 判定函数）。
2. `state==GLOBAL` → allowed。
3. `state==EMPTY` → denied（060007）。
4. `LIMITED`：逐 target——
   - `target.scope_type != "community"` → denied（060005 不支持的类型）；
   - 调 master-data `ResolveScopeAncestors(target.scope_id)`，`found==false` → denied（**安全拒绝**未知节点）；
   - 祖先链 ∩ ids ≠ ∅ → 该 target covered，否则 denied。
5. 全部 covered → allowed。
- 依赖：permission-service ServiceContext 需挂 master-data gRPC client（`c.MasterDataRpc` 已存在，`servicecontext.go` sysconfig 已引用 masterdatav1）。三层超时对齐 [[grpc-timeout-layers]]。

### 4.7 ResolveScopeAncestors（master-data 新增 RPC）
```proto
// masterdata/v1 — @auth: INTERNAL
rpc ResolveScopeAncestors(ResolveScopeAncestorsRequest) returns (ResolveScopeAncestorsResponse);
message ResolveScopeAncestorsRequest {
  int64 node_id = 1 [jstype = JS_STRING]; // md_administrative_division.id 或 md_residential_area.id
}
message ResolveScopeAncestorsResponse {
  common.v1.BaseResp base = 1;
  repeated int64 ancestor_ids = 2 [jstype = JS_STRING]; // 自包含、self-first（t,parent,...,root），≤6
  bool found = 3; // 节点不存在/已删除 → false（消费方安全拒绝）
}
```
**逻辑**：
- `node_id` 命中 `md_residential_area` → 经 `community_div_id` 入树；命中 `md_administrative_division` → 直接走 `parent_id` 链。
- 祖先链自包含、**截断 ≤6 节点**（超过深度时返回最接近 root 的 6 个，root 优先保留，见 §10 边界）。
- 数据源：全量行政区划树 + 小区→community_div_id 映射的**整树缓存**（内存/Redis，TTL 30min），低频变更。
- 拓扑变更（division/residential_area 增删改/审批落库）→ 显式失效缓存（§6.4）。
- 复用组合：不另起双套解析，`ResolveScopeAncestors` 是唯一被 permission-service 消费的祖先链入口（Info #5）。

### 4.8 AssignRole / RevokeRole（复用，强化幂等）
- 复用现有 RPC（无需新增）。`AssignRoleRequest` 已含 `scope_type/scope_id/status/verified_at/expires_at`。
- 幂等：`uk_user_role_scope` 唯一索引 + INSERT IGNORE（或捕获唯一键冲突视为成功）；并发重复加入只产生一条（REQ-4.2）。
- 缓存失效收敛：AssignRole/RevokeRole/UpdateUserRoleStatus 内部统一调用 `invalidateUserCaches(userId)`（§6.3）。

### 4.9 moderation 回调身份语义（SHOULD FIX S4 定稿）
- `UpdateNoticeModerationStatus` / `UpdateLostFoundModerationStatus` 由 **moderation-service 服务间回调**（`rpc/internal/consumer/task_handler.go` 发起，无用户 JWT）。
- 处理流程（community-hub）：
  1. reverse-lookup 内容 `id → community_id`（查不到 → 拒绝，符合「目标小区被校验而非假设」）。
  2. 以**系统身份**（`system_user_id=0`，global scope）调 `AssertPublishScope` → 放行（不按内容作者 scope 判定）。
- 语义：审核系统是全局权威，服务身份回调一律 global 放行；用户身份（未来如有）按用户 scope 校验。**不使用**内容作者/审核员的个人 scope。

---

## 5. 业务流程

### 5.1 CheckPermission 能力分层（修订 rbac-design §2.5）

#### 5.1.1 聚合规则（数据驱动，SHOULD FIX S2 落地）
```
权限 P 的 min_verf_level = L（0 或 2）
对每个授予 P 的角色 grant g：
  满足层级(g) = 2  若 status==2 AND verified_at IS NOT NULL AND 未过期
               = 0  若 status∈{0,1} AND 未过期
               = none 若 status∈{3,4}
P 放行 ⟺ max(满足层级(g)  for g in grants(P)) ≥ L
```
- 与 REQ-2.2「max satisfied level / 任一满足即放行」一致。
- registered_user（status=2, verified_at=NULL）恒满足 level-0；绝不满足 level-2。
- **现有 `FindActiveByUserId`（`ur.status=2`）语义必须改造**：CheckPermission 需取 status∈{0,1,2} 的全部 grant 再分层判定，不能只取 status=2。

#### 5.1.2 判定流程
```
0. user:disabled 标记 → 拒绝（保留）
1. 查权限定义 min_verf_level：perm:def:{needle}（Redis String TTL 30min，MISS 查 sys_permission）
2. 查用户满足层级：HGET perm:user:{userId} {needle} → maxLevel
   HIT → allowed = maxLevel >= L
   MISS → DB：FindActiveRolesByUserId(status∈{0,1,2}, 未过期) →
          角色→rel_role_permission→sys_permission →
          逐 granted path 计算 maxLevel → HSET perm:user:{userId}（Hash {path: maxLevel}）→
          命中 needle 比较返回
```
- 缓存从 Set 改为 **Hash**（field=path, value=满足层级）。开发阶段无兼容负担。

### 5.2 scope 合并优先级（REQ-A，SHOULD FIX S3 REQ 化）
> 供 GetDataScopes（读）与 AssertPublishScope（写）共用的唯一判定函数 `resolveUserScope(userId, scopeType)`。

```
active_grants = rel_user_role WHERE user_id=? AND status∈{0,1,2} AND (expires_at IS NULL OR expires_at>NOW())
              JOIN sys_role ON role_id=id AND sys_role.status=1 AND deleted_at IS NULL
1. 若 ∃ grant.scope_type=='global'              → state=GLOBAL   （global 支配）
2. 否则 limited_ids = ∪ {scope_id : grant.scope_type==scopeType AND scope_id!=0}
   若 limited_ids≠∅                            → state=LIMITED  （并集）
3. 否则                                       → state=EMPTY    （空；''/0 占位零贡献）
```
- 三态互斥；**EMPTY 永不等于 GLOBAL**（REQ-1.2）。
- 场景：审核员兼业主（global+limited）→ GLOBAL 生效；多小区业主 → LIMITED 并集；仅 registered_user → EMPTY。

### 5.3 注册自动分配 registered_user
```
user-service CreateUser 成功（DB 落库后）→ 同步 AssignRole(userId, role_id(registered_user=9), scope_type='', scope_id=0, status=2)
- 幂等：uk_user_role_scope（user_id,9,'',0）
- 失败：不阻塞注册（记录告警，后续可补）；重复注册不产生重复行（REQ-3.4）
```

### 5.4 加入=自动授权 / 退出=撤销（NEEDS CLARIFICATION #2 配套）
```
JoinCommunity(user_id, community_id, building/unit/room, ownership):
  1. 校验 ownership ∈ {OWNED, RENTED}，否则 10040
  2. 复用现有 join 校验（用户存在/小区上限/地址唯一/频次限制）
  3. 写 membership（active）
  4. role_code = OWNED→'owner' | RENTED→'tenant'；role_id = roleMapper 解析
  5. 同步 AssignRole(user_id, role_id, 'community', community_id, status=0)
     - 失败 → 补偿：恢复 membership bind_status 或删除 → 返回失败（不留「有成员无 scope」状态，REQ-4.1）
  6. 成功 → 返回（授权即时生效）

LeaveCommunity(user_id, community_id):
  1. 校验 membership active
  2. 置 membership left
  3. 同步 RevokeRole(user_id, owner_role_id, 'community', community_id)
        + RevokeRole(user_id, tenant_role_id, 'community', community_id)  // 双调幂等，只撤销存在的
     - 失败 → 补偿：恢复 bind_status=active → 返回失败（REQ-4.3）
  4. 清 default_community_id 逻辑保留
```

### 5.5 发布校验链路 + publisher_id 取 JWT
```
community-hub 写接口（8 个）：功能权限(PermMiddleware) → 数据权限(AssertPublishScope) → 落库
- targets = [{scope_type:'community', scope_id: 内容 community_id}]
- userId / publisher_id：一律取 JWT 认证身份（REST API 层从 ctx 提取 user_id，
  覆盖写入 gRPC 请求的 publisher_id 与 AssertPublishScope.userId，忽略客户端 body 传值）
- 校验顺序错误码：功能失败 080002；数据失败 060007→080006
```
- 涉及写接口（REQ-5.3 全量）：`CreateNotice`/`UpdateNotice`/`DeleteNotice`/`UpdateNoticeModerationStatus`/`CreateLostFound`/`ResolveLostFound`/`UpdateLostFoundModerationStatus`/`UpsertContacts`。

### 5.6 读过滤（REQ-1.6）
- 列表接口 `ListNotices`/`ListLostFound`/`ListContacts`：调 `GetDataScopes(userId,'community')` →
  - GLOBAL → 不追加过滤；
  - LIMITED → `WHERE community_id IN (ids)`（若请求的 community_id ∉ ids → 空列表）；
  - EMPTY → 空列表（不泄露小区内部内容）。

### 5.7 moderation 回调（SHOULD FIX S4 落地）
见 §4.9。回调不入用户 JWT 校验；以系统身份 global 放行；内容不存在 → 拒绝。

---

## 6. 缓存设计

| Key | 类型 | 内容 | TTL | 失效 |
|-----|------|------|-----|------|
| `perm:user:{userId}` | Hash | `{path: maxLevel}`（能力分层后） | 30min | Assign/Revoke/UpdateUserRoleStatus/InvalidateUserCache |
| `perm:def:{needle}` | String | `{min_verf_level}` | 30min | 权限定义变更（TTL 兜底） |
| `perm:scopes:{userId}:{scopeType}` | String | JSON `{"state","ids"}` | 30min | Assign/Revoke/UpdateUserRoleStatus/InvalidateUserCache |
| `user:disabled:{userId}` | String | 标记 | 24h | user-service 状态变更（保留） |
| master-data 整树缓存 | 内存/Redis | division 树 + community_div_id 映射 | 30min | division/residential_area 增删改/审批落库 |

### 6.3 失效收敛（Info #8 落地）
- 新增共享 helper `invalidateUserCaches(ctx, userId)`：DEL `perm:user:{userId}` + SCAN-DEL `perm:scopes:{userId}:*`。
- 在 `assignrolelogic.go` / `revokerolelogic.go` / `updateuserrolestatuslogic.go` / `invalidateusercachelogic.go` 内部统一调用——不依赖调用方（user-service/community-hub）记得失效。
- SCAN 范围限定单用户前缀，安全（[[redis-cache-soft-delete]] 联动）。

---

## 7. 依赖方向（SHOULD FIX S1 定稿，修订 access-control-design.md §8）

```
mobile ──加入/退出──▶ user-service ──Assign/RevokeRole──▶ permission-service
mobile ──发布──▶ community-hub（功能权限 → 数据权限 → 落库）
community-hub ──AssertPublishScope──▶ permission-service ──ResolveScopeAncestors──▶ master-data
community-hub ──读配额(未来§7)──▶ master-data     ← 仅配额；祖先链解析不再直连
community-hub ──GetDataScopes──▶ permission-service（读过滤）
user-service ──Assign/Revoke/CheckPermission──▶ permission-service（已有）
```
- **修订 action**：`docs/specs/access-control-design.md` §8 L273 `community-hub ──读配额/祖先链──▶ master-data` → `community-hub ──读配额──▶ master-data`，并加注「祖先链解析仅 permission-service 消费」。tasks 0.0 落此修订。

---

## 8. 错误码清单（对齐后）

| 错误码 | 含义 | 归属 | 状态 |
|:---:|------|------|:---:|
| 060005 | 不支持的 scope_type（AssertPublishScope target） | permission-service | 复用 |
| 060007 | 无数据范围权限（AssertPublishScope 拒绝） | permission-service | **新增** |
| 080006 | 目标小区超出发布者数据范围（API 面） | community-hub | **新增** |
| 080002 / 080003 / 080001 / 080005 | 无发布权限 / 超限 / 不存在 / 参数无效 | community-hub | 复用 |
| 10040 / 10001 / 10006 / 10007 / 10011-10013 | Join 校验 | user-service | 复用 |

---

## 9. Proto 变更清单（阶段 4 Owner 执行，5 项核心 + 1 项配套）

| # | 文件 | 变更 | 类型 |
|---|------|------|:---:|
| P1 | `api-proto/api/permission/v1/permission.proto` | `Permission` message 新增 `int32 min_verf_level`（透出列表/详情响应） | 兼容（加字段） |
| P2 | `api-proto/api/permission/v1/permission.proto` | 新增 `enum DataScopeState`；`GetDataScopesResponse` 新增 `state` 字段（保留 `scope_ids`） | 兼容（加字段） |
| P3 | `api-proto/api/permission/v1/permission.proto` | 新增 `message ScopeRef` | 兼容（新增 message） |
| P4 | `api-proto/api/permission/v1/permission.proto` | 新增 `rpc AssertPublishScope` + Request/Response（`@auth: INTERNAL`, `@idempotent: true`） | 兼容（新增方法） |
| P5 | `api-proto/api/masterdata/v1/masterdata.proto` | 新增 `rpc ResolveScopeAncestors` + Request/Response（`@auth: INTERNAL`） | 兼容（新增方法） |
| P6 | `api-proto/api/user/v1/user.proto` | 新增 `enum CommunityOwnership`；`JoinCommunityRequest` 新增 `ownership`（NEEDS CLARIFICATION #2 配套） | 兼容（加字段） |

- 全部为**兼容性变更**（`make breaking-check` 应无破坏）。int64 ID 字段一律 `[jstype = JS_STRING]`（[[proto-jstype]]）。
- 变更后更新 `api-proto/CHANGELOG.md`，通知消费方：permission→(community-hub,user)；masterdata→(permission)；user→(auth,web/mobile)。

---

## 10. 安全考虑

| # | 风险 | 缓解 |
|---|------|------|
| 1 | **空当 global 灾难** | `''` 与 `global` 物理分离；`resolveUserScope` 三态互斥；`scope_id=0` 不进并集；EMPTY 对任何 target 拒绝（REQ-1.2） |
| 2 | `publisher_id`/`userId` 伪造 | 一律取 JWT，忽略 body；AssertPublishScope 不替调用方选定身份校验（REQ-5.4/5.5） |
| 3 | 越权跨小区发布 | 逐 target 祖先链 ∩ 并集；多目标全过才放行；未知节点安全拒绝 |
| 4 | 退出后仍可发布（缓存漂移） | 失效收敛到 grant 变更处理器内部；验收矩阵 §11.4「退出 B 后立刻在 B 发布」 |
| 5 | 能力分层误放行 | 聚合规则数据驱动（`verified_at` 区分）；level-2 恒需认证；status 3/4 不满足任何层级 |
| 6 | 审核回调身份冒用 | 服务间回调走系统身份（global grant）；reverse-lookup 校验存在 |
| 7 | 角色名短路 | sys_admin/registered_user/系统身份全部建模为 grant，无字段短路（[[is-system-no-permission-shortcut]]） |
| 8 | 覆盖树深度/未知节点 | ≤6 截断（root 优先保留）；`found=false` → 拒绝 |
| 9 | 读泄露 | EMPTY → 空列表；LIMITED → IN 过滤 |

---

## 11. 边界与后续（不实现）

- 同屋互见（§5.7）→ 变更②；端限制（§4）→ auth-service 独立变更；当前小区（§6）→ user-service 独立变更；发布配额（§7）→ 独立变更；商户广告（§5.4 未来授权源）→ 模型兼容（REQ-1.5），不实现。

---

## 12. 记忆引用汇总（Step 1.5 产出）

见 §2 表。must-follow 记忆均落实到对应章节；tasks 中高风险任务（Migration/Proto/安全）显式标注 `// SEE: [[...]]`。

---

## 附：评审 SHALL 一致行为契约（供集成测试对照）

1. 注册用户发布 → ❌（无数据范围）；未认证业主发布（配额内）→ ✅；未认证业主选举 → ❌（level-2）；认证业主选举 → ✅；待审发布 → ✅。
2. 加入 → rel_user_role 出现 owner/tenant + community scope(status=0)；退出 → 撤销；重复加入幂等。
3. owner@A 发 B → ❌ 080006；伪造 publisher_id → 无效（JWT）；审核员（global）→ ✅；退出 B 后立刻在 B 发 → ❌。
4. 读列表按 scope 过滤；注册用户读不到小区内部内容。
5. 审核状态回调（服务身份）→ ✅ 放行；内容不存在 → 拒绝。
