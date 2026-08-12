# Tasks: 数据权限核心（access-data-permission）

> **对执行 Agent 的指令**：每个 Task 独立可测，按 TDD 执行（先写测试→看失败→写实现→看通过）。精确到文件路径。
> 依赖顺序：**阶段0 Proto（Owner）→ 阶段1 permission + 阶段2 master-data（权威，可并行）→ 阶段3 user + 阶段4 community-hub（编排/消费，可并行）→ 阶段5 集成**。
> 权威契约：`docs/specs/access-control-design.md`（§1.4/§3/§5）+ `.harness/changes/access-data-permission/design.md`。

---

## 阶段 0 · 全局 / Proto（由全局 Owner 执行，**禁止分发子 Agent**，遵守 Proto 管理规范 + 硬约束 #1/#2）

> 阶段 4 由 Owner 统一执行本阶段。完成后 `make ci` 全绿，生成代码提交。

### Task 0.0: 修订 access-control-design.md §8 依赖方向（评审 S1 落地）
- **修改**: `docs/specs/access-control-design.md` §8 L273
- [ ] 将 `community-hub ──读配额/祖先链──▶ master-data` 改为 `community-hub ──读配额──▶ master-data`
- [ ] 加注：祖先链解析仅 permission-service 经 `ResolveScopeAncestors` 消费；community-hub 不得直连 master-data 做 scope 解析

### Task 0.1: permission.proto — min_verf_level 透出 + GetDataScopes 三态
- **修改**: `api-proto/api/permission/v1/permission.proto`
- [ ] `Permission` message 新增 `int32 min_verf_level = 12`（0=角色+数据范围即可, 2=需已认证）
- [ ] 新增 `enum DataScopeState { DATA_SCOPE_STATE_UNSPECIFIED=0; EMPTY=1; LIMITED=2; GLOBAL=3; }`
- [ ] `GetDataScopesResponse` 新增 `DataScopeState state = 3`（保留现网 `scope_ids = 2` 字段号不变，wire 兼容）
- [ ] 更新文件头注释（错误码 060006）

### Task 0.2: permission.proto — ScopeRef + AssertPublishScope RPC
- **修改**: `api-proto/api/permission/v1/permission.proto`
- [ ] 新增 `message ScopeRef { string scope_type=1; int64 scope_id=2 [jstype=JS_STRING]; }`
- [ ] 新增 `AssertPublishScopeRequest { int64 user_id=1 [jstype=JS_STRING]; repeated ScopeRef targets=2; }`
- [ ] 新增 `AssertPublishScopeResponse { BaseResp base=1; bool allowed=2; }`
- [ ] 新增 `rpc AssertPublishScope`（注释 `@auth: INTERNAL`, `@idempotent: true`, `@timeout: 500`）
- [ ] 在 `service PermissionService` 中按序注册

### Task 0.3: masterdata.proto — ResolveScopeAncestors RPC
- **修改**: `api-proto/api/masterdata/v1/masterdata.proto`
- [ ] 新增 `ResolveScopeAncestorsRequest { int64 node_id=1 [jstype=JS_STRING]; }`
- [ ] 新增 `ResolveScopeAncestorsResponse { BaseResp base=1; repeated int64 ancestor_ids=2 [jstype=JS_STRING]; bool found=3; }`
- [ ] 新增 `rpc ResolveScopeAncestors`（注释 `@auth: INTERNAL`, `@idempotent: true`）
- [ ] 在 `service MasterdataService` 中按序注册

### Task 0.4: user.proto — JoinCommunity ownership
- **修改**: `api-proto/api/user/v1/user.proto`
- [ ] 新增 `enum CommunityOwnership { COMMUNITY_OWNERSHIP_UNSPECIFIED=0; OWNED=1; RENTED=2; }`
- [ ] `JoinCommunityRequest` 新增 `CommunityOwnership ownership = 6`
- [ ] 更新文件头注释（新增错误码 10040 语义说明）

### Task 0.5: api-proto 生成 + CI + 变更通知
- **执行**: `cd api-proto`
- [ ] `make lint` → 0 errors；`make breaking-check` → 无破坏性变更；`make generate`
- [ ] 确认 `gen/go/permission/v1`、`gen/go/masterdata/v1`、`gen/go/user/v1` 已更新
- [ ] 更新 `api-proto/CHANGELOG.md`（P1-P6 兼容变更）
- [ ] 通知消费方：permission→(community-hub,user)；masterdata→(permission)；user→(auth,web/mobile)
- [ ] `go mod tidy` 校验生成代码同步

---

## 阶段 1 · permission-service（核心·权威）// SEE: [[is-system-no-permission-shortcut]], [[permission-seed-api-path-must-match-routes]], [[migration-must-execute]]

### Task 1.1: sys_permission.min_verf_level 列迁移 + registered_user 角色/权限种子
- **修改**: `services/permission-service/scripts/init_permissions.sql`（新增迁移段）
- [ ] 新增 `ALTER TABLE sys_permission ADD COLUMN min_verf_level TINYINT NOT NULL DEFAULT 0`
- [ ] 发布类权限（`lostfound:create`/`notice:create` 等）置 `min_verf_level=0`；选举类（`committee:election:vote`）置 `2`
- [ ] 新增 `sys_role` 行 `(id=9, 'registered_user', is_system=1, status=1)`
- [ ] 新增 `rel_role_permission` 种子：registered_user → browse 权限（`GET:/api/community/notices`、`GET:/api/community/lostfound`、`GET:/api/community/contacts`）——**path 必须与实际 REST 路由一致** // SEE: [[permission-seed-api-path-must-match-routes]]
- [ ] 预留系统审核身份：`rel_user_role` 种子 `(user_id=0, role_id=sys_admin, scope_type='global', scope_id=0, status=2)`
- [ ] 执行迁移到本地 DB 并验证列/行存在 // SEE: [[migration-must-execute]]

### Task 1.2: rel_user_role 模型层 — scope 三态语义 + 唯一索引 + 状态过滤
- **修改**: `services/permission-service/model/rel.go`
- **新增**: `services/permission-service/migration/001_scope_three_state.sql`
- [ ] 迁移：`ALTER TABLE rel_user_role ADD UNIQUE KEY uk_user_role_scope (user_id, role_id, scope_type, scope_id)`
- [ ] 定义常量：`ScopeTypeGlobal="global"`, `ScopeTypeEmpty=""`, `ScopeTypeCommunity="community"` 等
- [ ] `FindScopesByUserId`：过滤 `status IN (0,1,2)` 且 `scope_id != 0`
- [ ] 新增 `FindActiveRolesByUserId(userId)`：返回 `status IN (0,1,2)` 且未过期（`expires_at IS NULL OR >NOW()`）的 grants（含 scope_type/scope_id/verified_at/ur_status）
- [ ] **TDD**: `model/rel_test.go` — 状态过滤 / 空 scope 行不进结果 / 过期排除

### Task 1.3: 共享 scope 判定 helper — resolveUserScope 三态合并
- **新增**: `services/permission-service/rpc/internal/logic/permission/scope.go`
- **新增**: `services/permission-service/rpc/internal/logic/permission/scope_test.go`
- [ ] `resolveUserScope(ctx, userId, scopeType) (state DataScopeState, ids []int64)` 实现 REQ-A 合并优先级：global 支配 → limited 并集(排除 scope_id=0) → empty
- [ ] **RED**: 表驱动测试（仅 global / global+limited→global / 多 limited→并集 / 仅 empty→EMPTY / status=3,4 排除 / '' 行零贡献）
- [ ] **确认 RED**: `go test -run TestResolveUserScope` → FAIL
- [ ] **GREEN**: 实现后 → PASS
- [ ] **REFACTOR**: 清理，保持绿

### Task 1.4: GetDataScopes 三态重写 + 读穿缓存
- **修改**: `services/permission-service/rpc/internal/logic/permission/getdatascopeslogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/helpers.go`
- **新增**: `services/permission-service/rpc/internal/logic/permission/getdatascopeslogic_test.go`
- [ ] 调用 `resolveUserScope`；映射 `{state, ids}` 到 `GetDataScopesResponse{state, scope_ids}`
- [ ] 读穿缓存：`perm:scopes:{userId}:{scopeType}` JSON `{"state","ids"}`；HIT 解析返回，MISS 计算后 SET + EXPIRE 30min
- [ ] **TDD**: 三态响应 / 缓存 HIT 不查 DB / empty 返回空列表 / global 返回空 ids
- [ ] **确认**：`go test -run TestGetDataScopesLogic` 全绿

### Task 1.5: CheckPermission 能力分层（聚合规则 + 缓存 Hash 化）
- **修改**: `services/permission-service/rpc/internal/logic/permission/checkpermissionlogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/getuserpermissionslogic.go`（同用 status∈{0,1,2} 活跃 grant，保证未认证业主的发布权限码在列）
- **修改**: `services/permission-service/rpc/internal/logic/permission/helpers.go`
- **新增**: `services/permission-service/rpc/internal/logic/permission/checkpermissionlogic_test.go`
- [ ] 取 `FindActiveRolesByUserId`（status∈{0,1,2}），不再只取 status=2
- [ ] 聚合：对每个 granted path 计算 maxLevel（level-2 = status==2 AND verified_at NOT NULL；level-0 = status∈{0,1}；3/4 不计），与 `min_verf_level` 比较
- [ ] 权限定义缓存 `perm:def:{needle}`（String, min_verf_level, TTL 30min）
- [ ] 用户缓存改为 Hash `perm:user:{userId}` `{path: maxLevel}`
- [ ] **TDD**（对照 cap-layering 场景）：未认证业主发布✅ / 未认证业主选举❌ / 认证业主选举✅ / 待审发布✅ / 已过期❌ / 多角色叠加取最高✅ / registered_user 仅 browse✅且不满足 level-2✅
- [ ] **确认**：`go test -run TestCheckPermissionLogic` 全绿

### Task 1.6: AssignRole/RevokeRole/UpdateUserRoleStatus 幂等 + 缓存失效收敛
- **修改**: `services/permission-service/rpc/internal/logic/permission/assignrolelogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/revokerolelogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/updateuserrolestatuslogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/invalidateusercachelogic.go`
- **新增**: `services/permission-service/rpc/internal/logic/permission/invalidate_caches.go`
- [ ] 新增共享 `invalidateUserCaches(ctx, userId)`：DEL `perm:user:{userId}` + SCAN-DEL `perm:scopes:{userId}:*`；四处理器统一调用（失效收敛，不依赖调用方）
- [ ] AssignRole：唯一键冲突视为幂等成功（INSERT IGNORE 或捕获 duplicate）
- [ ] RevokeRole：唯一键级精确删除 + 失效
- [ ] **TDD**: 重复 Assign 只一条 / Revoke 后 GetDataScopes 立即 EMPTY（缓存 DEL 生效）// SEE: [[redis-cache-soft-delete]]

### Task 1.7: AssertPublishScope RPC
- **新增**: `services/permission-service/rpc/internal/logic/permission/assertpublishscopelogic.go`
- **新增**: `services/permission-service/rpc/internal/logic/permission/assertpublishscopelogic_test.go`
- **修改**: `services/permission-service/rpc/internal/svc/servicecontext.go`（挂 master-data client：`zrpc.MustNewClient(c.MasterDataRpc)` → `masterdatav1.NewMasterdataServiceClient`）
- **修改**: `services/permission-service/rpc/internal/server/permissionserviceserver.go`
- [ ] 逻辑：scope 解析 → GLOBAL 放行 / EMPTY 拒绝(060006) → 逐 target 校验 scope_type=='community' + `ResolveScopeAncestors` → 祖先链 ∩ ids ≠ ∅ → covered
- [ ] 未知/失效节点（`found=false`）→ 拒绝（安全拒绝）
- [ ] **TDD**: owner@A 发 A✅ / owner@A 发 B❌(060006) / global 审核员任意✅ / 多目标部分未覆盖整体拒绝 / empty 拒绝 / 未知节点拒绝
- [ ] 注册到 server；**确认** `go test -run TestAssertPublishScopeLogic` 全绿

### Task 1.8: 门禁
- [ ] `go build ./...` + `go test ./...`
- [ ] `bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service` → PASS // SEE: [[pre-commit-checks]]

---

## 阶段 2 · master-data-service（辅助：祖先链解析）

### Task 2.1: ResolveScopeAncestors — 整树缓存 + 自包含祖先链
- **新增**: `services/master-data-service/rpc/internal/logic/scoperesolve/resolvescopeancestorslogic.go`
- **新增**: `services/master-data-service/rpc/internal/logic/scoperesolve/resolvescopeancestorslogic_test.go`
- **新增**: `services/master-data-service/rpc/internal/svc/scopeancestorcache.go`
- [ ] `scopeancestorcache`：全量 division 树（parent_id 链）+ residential_area→community_div_id 映射，内存/Redis TTL 30min
- [ ] 逻辑：node 命中 residential_area → 经 community_div_id 入树；命中 division → parent 链；自包含 self-first，截断 ≤6（root 优先保留）；未知/删除 → `found=false`
- [ ] **TDD**: 小区完整链(R→div→street→county→city→province) / city 级链(D,province) / 拓扑重挂父节点后反映新拓扑 / 未知节点 found=false / 深度>6 截断
- [ ] 注册到 `rpc/internal/server/`；**确认** `go test -run TestResolveScopeAncestors` 全绿

### Task 2.2: 整树缓存失效（拓扑变更）
- **修改**: `services/master-data-service/rpc/internal/logic/division/`（create/update/delete）、`residentialarea/`（create/update/delete/review 落库）
- **新增**: `services/master-data-service/rpc/internal/logic/scoperesolve/invalidate_test.go`
- [ ] division/residential_area 实际变更点调用 `scopeancestorcache.Invalidate()`
- [ ] **TDD**: 变更后 ResolveScopeAncestors 立即反映新拓扑（缓存失效非 stale）

### Task 2.3: 门禁
- [ ] `go build ./...` + `go test ./...`
- [ ] `bash .harness/skills/qa/scripts/harness-checks.sh --service master-data-service` → PASS

---

## 阶段 3 · user-service（编排，非权威）

### Task 3.1: CreateUser 自动分配 registered_user
- **修改**: `services/user-service/rpc/internal/logic/user/create_user_logic.go`
- **新增**: `services/user-service/rpc/internal/logic/user/create_user_logic_test.go`
- [ ] DB 落库成功后（moderation 回调前）同步 `AssignRole(userId, roleIDByCode('registered_user'), scope_type='', scope_id=0, status=2)`
- [ ] role_id 经既有 `roleMapper` 解析；失败仅告警不阻塞注册
- [ ] **TDD**: 注册成功→出现 registered_user grant / 重复分配幂等（无重复行）/ 手机号已注册失败→不分配
- [ ] 参照既有 `apply_role_logic_test.go` 构造 fake PermissionClient

### Task 3.2: JoinCommunity ownership + 自动授权
- **修改**: `services/user-service/rpc/internal/logic/user/join_community_logic.go`
- **新增**: `services/user-service/rpc/internal/logic/user/join_community_ownership_test.go`
- [ ] 校验 `ownership ∈ {OWNED, RENTED}`，UNSPECIFIED → 10040
- [ ] membership 落库后同步 `AssignRole(user_id, roleIDByCode(owner|tenant), 'community', community_id, status=0)`；失败 → 补偿恢复 membership 并返回失败
- [ ] **TDD**: 自有→owner grant / 租住→tenant grant / 授权失败→join 失败且无「有成员无 scope」/ ownership 缺失→10040 / 重复加入幂等
- [ ] 同步更新生成代码引用的 `userv1.JoinCommunityRequest` 字段

### Task 3.3: LeaveCommunity 撤销授权
- **修改**: `services/user-service/rpc/internal/logic/user/leave_community_logic.go`
- **新增**: `services/user-service/rpc/internal/logic/user/leave_community_revoke_test.go`
- [ ] membership 置 left 后同步 `RevokeRole(user_id, owner_role_id, 'community', community_id)` + `RevokeRole(user_id, tenant_role_id, 'community', community_id)`（双调幂等）；失败 → 恢复 bind_status=active 并返回失败
- [ ] **TDD**: 退出撤销该小区 owner/tenant / 其他小区保留 / 撤销失败→leave 失败并恢复 / 重复 leave 幂等

### Task 3.4: 门禁
- [ ] `go build ./...` + `go test ./...`
- [ ] `bash .harness/skills/qa/scripts/harness-checks.sh --service user-service` → PASS

---

## 阶段 4 · community-hub-service（消费方）

### Task 4.0: 前置 — 确认写接口 JWT 身份注入通道
- **修改**: `services/community-hub-service/api/internal/svc/servicecontext.go`
- **修改**: `services/community-hub-service/api/internal/config/config.go`
- [ ] 确认 REST API 层 `rest.WithJwt` 后能从 ctx 提取 user_id（新增 helper `jwtUserID(ctx) (int64, error)`）// SEE: [[verify-api-before-calling]]
- [ ] 确认 `PermMiddleware`（功能权限）先于数据权限执行

### Task 4.1: publisher_id 取 JWT（写接口身份规范化）
- **修改**: `services/community-hub-service/api/internal/handler/*.go`（notice/lostfound 写接口 handler）
- **修改**: `services/community-hub-service/api/internal/logic/lostfound/createlostfoundlogic.go`
- **修改**: `services/community-hub-service/api/internal/logic/notice/createnoticelogic.go`
- [ ] 构建 gRPC 请求时用 JWT user_id 覆盖 `publisher_id`，忽略客户端 body 值
- [ ] **TDD**: 伪造 publisher_id → 落库为 JWT 身份 / 合法发布者身份正确落库

### Task 4.2: AssertPublishScope 挂载 — lostfound 写接口
- **修改**: `services/community-hub-service/api/internal/logic/lostfound/createlostfoundlogic.go`
- **修改**: `services/community-hub-service/api/internal/logic/lostfound/resolvelostfoundlogic.go`（或对应 RPC 逻辑）
- **新增**: `services/community-hub-service/rpc/internal/logic/lostfound/publishscope_test.go`
- [ ] 落库前 `AssertPublishScope(jwtUserID, [{scope_type:'community', scope_id: community_id}])`；`allowed=false` → 080006
- [ ] 校验顺序：功能权限 → 数据权限 → 落库
- [ ] **TDD**: owner@A 发 A✅ / owner@A 发 B❌(080006) / global 审核员✅ / 伪造 userId 借用 admin scope → 按 JWT 自身 scope 拒绝
- [ ] 错误码映射 `060006 → 080006`

### Task 4.3: AssertPublishScope 挂载 — notice 写接口
- **修改**: `services/community-hub-service/api/internal/logic/notice/createnoticelogic.go` / `updatenoticelogic.go` / `deletenoticelogic.go`
- [ ] Create/Update/DeleteNotice 落库前 AssertPublishScope(目标 community_id)
- [ ] **TDD**: 更新超范围拒绝 / 删除超范围不删 / 正常通过

### Task 4.4: AssertPublishScope 挂载 — UpsertContacts
- **修改**: `services/community-hub-service/api/internal/logic/contact/upsertcontactlogic.go`
- [ ] 落库前 AssertPublishScope(community_id)
- [ ] **TDD**: 目标小区超范围拒绝 / 正常通过

### Task 4.5: moderation 回调身份 + 校验（评审 S4 落地）
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/updatemoderationstatuslogic.go`
- **修改**: `services/community-hub-service/rpc/internal/logic/lostfound/updatemoderationstatuslogic.go`
- **新增**: `services/community-hub-service/rpc/internal/logic/notice/updatemoderationstatus_test.go`
- [ ] reverse-lookup 内容 `id → community_id`；查不到 → 拒绝
- [ ] 以系统身份（`system_user_id=0` 常量）调 `AssertPublishScope`（服务身份回调 global 放行，不按作者 scope）
- [ ] **TDD**: 服务回调存在内容→放行 / 内容不存在→拒绝

### Task 4.6: 读列表按 GetDataScopes 过滤
- **修改**: `services/community-hub-service/rpc/internal/logic/notice/listnoticeslogic.go`
- **修改**: `services/community-hub-service/rpc/internal/logic/lostfound/listlostfoundlogic.go`
- **修改**: `services/community-hub-service/rpc/internal/logic/contact/listcontactslogic.go`
- **新增**: `services/community-hub-service/rpc/internal/logic/scope/filter.go` + `_test.go`
- [ ] `filterByScope(ctx, jwtUserID)` → GetDataScopes('community')：GLOBAL 不过滤 / LIMITED `IN(ids)` / EMPTY 空列表
- [ ] **TDD**: 业主只见所属小区 / 空范围空列表 / global 跨小区可见

### Task 4.7: 错误码 080006 注册
- **修改**: `services/community-hub-service/api/internal/types/types.go`（或既有 errx 常量文件）
- **修改**: `services/community-hub-service/api/internal/svc/route_registry.go`（如权限自动发现）
- [ ] 注册 `080006`「目标小区超出发布者数据范围」；确认 080002/080006 分层语义

### Task 4.8: 门禁
- [ ] `go build ./...` + `go test ./...`
- [ ] `bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service` → PASS

---

## 阶段 5 · 移动端（web/mobile）

### Task 5.1: 加入小区流程携带 ownership
- **修改**: `web/mobile/src/api/user.ts` — `joinCommunity(communityId, building, unit, room, ownership)` 传 `{community_id, building, unit, room, ownership}`
- **修改**: `web/mobile/src/pages/join-community/join-community.vue` — 加入前收集「自有/租住」选择（必填）+ 楼/单元/房号输入
- [ ] 补全表单并调用更新后的 joinCommunity
- [ ] `npm run type-check` 通过；`npm run test:unit` 通过

---

## 阶段 6 · 集成验收

### Task 6.1: 跨服务集成测试（端到端验收矩阵 §11）
- **执行**: `scripts/start.sh`（全栈）+ 构造验收脚本
- [ ] 注册用户→无小区发布 → ❌；加入小区(自有)→owner+scope 出现→发布 ✅；未认证选举 → ❌；认证后选举 → ✅
- [ ] owner@A 发 B → ❌ 080006；抓包改 publisher_id → ❌（JWT 生效）；审核员（global）→ ✅
- [ ] 退出 B 后立刻在 B 发布 → ❌（缓存 DEL 生效）
- [ ] 读列表按 scope 过滤；注册用户读不到小区内部内容
- [ ] moderation 服务间回调 → 放行；内容不存在 → 拒绝

### Task 6.2: 收尾
- [ ] 更新 `docs/devlog/`（三层体系）
- [ ] 各服务 harness-checks 全绿；`git status` 无遗漏文件
- [ ] 复核 design.md 决策与实现一致（空 scope / 权属 / 错误码 / 依赖方向）

---

## Self-Review 结论（Step 5）
- **占位符**：0（全部精确到文件路径）
- **TDD 覆盖**：含逻辑任务均含 RED→GREEN→REFACTOR（1.3/1.4/1.5/1.7/2.1/3.1/3.2/3.3/4.2-4.6/5.1）
- **依赖顺序**：阶段0 Proto → 阶段1+2（权威并行）→ 阶段3+4（编排/消费并行）→ 阶段5 前端 → 阶段6 集成 ✅
- **记忆引用**：must-follow 已标注（is-system-no-permission-shortcut、permission-seed-api-path-must-match-routes、migration-must-execute、proto-jstype、grpc-only-comms、verify-api-before-calling、grpc-timeout-layers、redis-cache-soft-delete、pre-commit-checks）
- **Proto 变更**：阶段0 全部标记 Owner 执行，不分发子 Agent
