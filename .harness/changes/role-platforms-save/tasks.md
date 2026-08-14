# Tasks: role-platforms-save — 角色管理 bug 修复 + platforms 允许登录端写链路补全

> **对执行 Agent 的指令**：每个 Task 独立可测，按 TDD 执行（先写测试→看失败→写实现→看通过）。精确到文件路径。
> 依赖序：**Proto（Task 0.x）→ permission-service write-path（1.1-1.5, 1.7, 1.8, 1.11）→ update-fix（1.6, 1.9, 1.10 叠加，共享文件合并实现）→ 前端（2.1）**。
> 共享文件先改者：`types.go`/`helpers.go`（API）由 write-path 先改，update-fix 复用；`rpc/updaterolelogic.go` 两个 capability 同改同一函数，**在 Task 1.4 合并实现**。
> 每个含逻辑 Task 的 RED 证据必须含实际 FAIL 输出摘录（SEE [[tdd-red-evidence-requires-fail-excerpt]]）。
> 完成后必跑 `bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service`（硬性约束 #4）。

---

## 全局 / Proto（由全局 Claude 执行，不分发给子 Claude）

### Task 0.1: permission.proto 增 platforms 字段 + 错误码 060008
- **文件**: `api-proto/api/permission/v1/permission.proto`
- [ ] `CreateRoleRequest` 增字段 6：`repeated string platforms`，注释注明值域（pc/mobile）、空=fail-open、非法值 60008
- [ ] `UpdateRoleRequest` 增字段 7：`repeated string platforms`，注释同上
- [ ] 文件头错误码注释块增 `060008 — 非法登录端`（既有约定：新错误码登记头注释，见 2026-08-12 登记 060007 的 CHANGELOG；SEE [[error-code-collision-and-namespace-alignment]]——060008 语义必须唯一）
- [ ] `UpdateRole` RPC `@description` 注释补 platforms（「支持修改名称、描述、状态、权限列表」→ 追加「、允许登录端」）
- [ ] 确认字段号不与既有冲突（CreateRole 1-5 已占用→用 6；UpdateRole 1-6 已占用→用 7）
- [ ] TDD 不适用（纯契约声明）；验证由 Task 0.2 的 `make ci` 覆盖

### Task 0.2: make generate + lint + breaking-check + CHANGELOG
- **目录**: `api-proto/`
- [ ] `cd api-proto && make generate` — 重新生成 gen/go
- [ ] `make lint` → 确认 0 errors
- [ ] `make breaking-check` → 确认无破坏性变更（新增 repeated string 字段应兼容）
- [ ] `git diff gen/go/permission/v1/permission.pb.go` → 确认 `CreateRoleRequest`/`UpdateRoleRequest` 暴露 `Platforms []string`
- [ ] 更新 `api-proto/CHANGELOG.md`（顶部新增条目，参照 2026-08-13 access-control 格式：做了什么/为什么/影响）
- [ ] `make ci` 全绿

---

## permission-service

### Task 1.1: Model — `SysRoleModel.Update` 补 sort_order（REQ-PLAT-5 / D6）
- **修改**: `services/permission-service/model/permission.go`（`defaultSysRoleModel.Update`）
- [ ] Update SQL 由 `update sys_role set role_name = ?, description = ?, status = ?, platforms = ?, updated_at = now() where id = ?` 改为补 `sort_order = ?`
- [ ] 参数顺序：`role_name, description, status, platforms, sort_order, id`（参数数量与 SQL 占位符一致）
- [ ] TDD 说明：纯 SQL 字符串修改无分支，不写单测；由 Task 1.4 的 RPC 集成测试（Update 捕获 `SysRole.SortOrder==5`）兜底验证
- [ ] `go build ./...` 通过

### Task 1.2: RPC — `CodeInvalidPlatform` 常量 + platforms 校验/去重 helper（REQ-PLAT-4）
- **修改**: `services/permission-service/rpc/internal/logic/permission/helpers.go`
- **创建**: `services/permission-service/rpc/internal/logic/permission/helpers_test.go`（若不存在）
- [ ] 定义命名常量 `const CodeInvalidPlatform int32 = 60008`（防裸字面量，SEE [[error-code-literal-bypasses-qa-gate]]）
- [ ] 实现 `func validatePlatforms(ps []string) ([]string, error)`：
  - 值域 `{pc, mobile}`；任一非法值 → `errx.NewCodeError(int(CodeInvalidPlatform), "非法登录端: "+v)`（60008）
  - 空/nil → 通过（fail-open）
  - 重复值 → 去重且保持顺序（如 `["pc","pc","mobile"]` → `["pc","mobile"]`）
- [ ] **RED**: table-driven tests（空→通过；`["web"]`→60008；`["pc","pc","mobile"]`→去重后 `["pc","mobile"]`；nil→通过）；`go test -run TestValidatePlatforms` → 看 FAIL
- [ ] **GREEN**: 实现 validatePlatforms → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 1.3: RPC — CreateRole 写 platforms（REQ-PLAT-3/4）
- **修改**: `services/permission-service/rpc/internal/logic/permission/createrolelogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/createrolelogic_test.go`
- [ ] CreateRole 入口：`platforms, err := validatePlatforms(in.Platforms)`；`err != nil` → 返回 `Base: responsex.NewBaseRespFromError(err)`（60008，**原子拒绝，不 Insert**）
- [ ] role struct 增 `Platforms: joinPlatforms(platforms)`
- [ ] **RED**: 补 `TestCreateRole_InvalidPlatform`（`["web"]`→Base.Code 60008，RoleModel.Insert 未被调用）/ `TestCreateRole_PlatformsPersisted`（`["pc","mobile"]`→Insert 捕获 `Platforms=="pc,mobile"`）/ `TestCreateRole_PlatformsDedup`（`["pc","pc"]`→`"pc"`）；`go test -run TestCreateRole` → 看 FAIL
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 1.4: RPC — UpdateRole 合并实现：platforms 无条件覆盖 + 校验 + sort_order + 字段级策略 + 缓存失效（REQ-PLAT-3/4/5/7 + REQ-UPDATE-4）
- **修改**: `services/permission-service/rpc/internal/logic/permission/updaterolelogic.go`
- **修改**: `services/permission-service/rpc/internal/logic/permission/updaterolelogic_test.go`（现有测试文件）
- **校验顺序（钉死）**：
  1. `FindOne` → err → `Base 60001`「角色不存在」
  2. `existing.IsSystem == 1 && in.Status != nil` → `Base 60004`「系统角色状态不可修改」——**先于任何字段应用、先于 platforms 校验**（60004 message 由「系统角色不可修改」收窄为「系统角色状态不可修改」，仅本路径）// SEE: [[is-system-no-permission-shortcut]]（字段级放行不改变权限模型）
  3. `platforms, err := validatePlatforms(in.Platforms)` → `err != nil` → `Base 60008`（原子拒绝）
  4. 应用字段：`in.Name != nil` → RoleName；`in.Description != nil` → Description；`in.SortOrder != nil` → SortOrder；**`existing.Platforms = joinPlatforms(platforms)`（无条件覆盖，空=fail-open）**；status 仅非系统角色应用
  5. `RoleModel.Update`（含 sort_order）→ err 返回 Go error
  6. permission_ids 替换逻辑保留原样
  7. `invalidateRoleCache(in.Id)`（platforms 变更同样触发，原模式复用）// SEE: [[permission-cache-consistency]]（角色变更必须失效持有者缓存）
- [ ] **RED**: 取消 `TestUpdateRole_SystemRoleCannotModify` 的 `t.Skip`，改断言字段级策略矩阵：
  - 系统角色 + name/platforms → 成功（Base.Code 0），`Update` 捕获的 role `Status` 保持原值
  - 系统角色 + status=0 + name → `Base.Code 60004`，且 role 未修改（原子，无部分写入）
  - 系统角色 + status=0 + platforms=["web"] → `Base.Code 60004`（status 门禁优先于 60008）
  - 系统角色 + platforms=["web"]（无 status）→ `Base.Code 60008`
  - 非系统角色 + status=0 → 成功（状态落库）
  - 传 sort_order=5 → `Update` 捕获 `SortOrder==5`
  - platforms=[] → `Update` 捕获 `Platforms==""`（fail-open 清空）
  - platforms 变更后持有者（user 1001/1002）`perm:user:*`/`perm:scopes:*` 被 DEL（miniredis 断言，仿 `TestUpdateRole_CacheInvalidatedForMultipleUsers`）
  - `go test -run TestUpdateRole` → 看 FAIL（skip 已解除的新用例）
- [ ] **GREEN**: 实现新逻辑 → 看 PASS（含既有 Success/RoleNotFound 用例仍绿）
- [ ] **REFACTOR**: 保持绿

### Task 1.5: API types — RoleInfo/CreateRoleReq/UpdateRoleReq 增 Platforms（REQ-PLAT-2/6）
- **修改**: `services/permission-service/api/internal/types/types.go`
- [ ] `RoleInfo` 增 `Platforms []string json:"platforms"`（**无 omitempty**，空必须序列化为 `[]` 非 null，供前端 `Array.isArray` 判定「全部」）
- [ ] `CreateRoleReq` 增 `Platforms []string json:"platforms,optional"`
- [ ] `UpdateRoleReq` 增 `Platforms []string json:"platforms,optional"`
- [ ] TDD 不适用（类型映射）；由 Task 1.6/1.7/1.8 的测试覆盖

### Task 1.6: API helpers — toRoleInfo nil 防御 + platforms 透传（REQ-PLAT-6 + REQ-UPDATE-2）
- **修改**: `services/permission-service/api/internal/logic/perm/helpers.go`
- **修改**: `services/permission-service/api/internal/logic/perm/helpers_test.go`
- [ ] `toRoleInfo(nil)` → 返回 `types.RoleInfo{}` 零值，不 panic（REQ-UPDATE-2 nil 防御）
- [ ] `toRoleInfo` 增 `info.Platforms = r.Platforms`；nil/空时归一为 `[]string{}`（空数组非 null，REQ-PLAT-6）
- [ ] **RED**: `TestToRoleInfo_NilRole`（nil → 零值不 panic）/ `TestToRoleInfo_PlatformsPassthrough`（`["mobile"]` → `info.Platforms == ["mobile"]`）/ `TestToRoleInfo_EmptyPlatformsIsEmptyArray`（空 → `len==0` 且非 nil）；`go test -run TestToRoleInfo` → 看 FAIL
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿（既有 TestToRoleInfo_FieldMapping 仍绿）

### Task 1.7: API CreateRole — platforms 透传 + ToError（REQ-PLAT-2 + REQ-UPDATE-3）
- **修改**: `services/permission-service/api/internal/logic/perm/createrolelogic.go`
- **创建**: `services/permission-service/api/internal/logic/perm/createrolelogic_test.go`
- [ ] `grpcReq` 增 `Platforms: req.Platforms`
- [ ] RPC 调用后、deref `grpcResp.Role` 之前：`if err := responsex.ToError(grpcResp.Base); err != nil { return nil, err }`（60006 不再触发空指针 panic）// SEE: [[rpc-callback-must-check-response-base]]
- [ ] **RED**: 复用/扩展 `listroleslogic_test.go` 的 `mockPermRpc`（增 `createRoleFn` 字段）——`TestCreateRole_PlatformsPassThrough`（捕获 grpcReq.Platforms）/ `TestCreateRole_BaseErrorNoPanic`（Base 60006 + Role=nil → 返回 Go error，不 panic）；`go test -run TestCreateRole` → 看 FAIL
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 1.8: API UpdateRole — platforms 恒透传 + ToError（REQ-PLAT-2 + REQ-UPDATE-1）
- **修改**: `services/permission-service/api/internal/logic/perm/updaterolelogic.go`
- **创建**: `services/permission-service/api/internal/logic/perm/updaterolelogic_test.go`
- [ ] `grpcReq` 增 `Platforms: req.Platforms`（**恒透传**：空列表也设置，落实 D3「API 层始终透传」，避免与「未传」混淆）
- [ ] RPC 调用后、deref `grpcResp.Role` 之前：`if err := responsex.ToError(grpcResp.Base); err != nil { return nil, err }`（60004 + Role=nil → Go error，不再 panic → HTTP 不再 500）// SEE: [[rpc-callback-must-check-response-base]]
- [ ] **RED**: 扩展 `mockPermRpc`（增 `updateRoleFn` 字段）——`TestUpdateRole_PlatformsAlwaysPassed`（req.Platforms=[] → 捕获 grpcReq.Platforms 已设置）/ `TestUpdateRole_Base60004NoPanic`（Base 60004 + Role=nil → 返回 Go error，不 panic）/ `TestUpdateRole_Success`；`go test -run TestUpdateRole` → 看 FAIL
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 1.9: API base-check 类级审计 — panic 类 5 文件（REQ-UPDATE-3）
> 类级规则（SEE [[rpc-callback-must-check-response-base]]）：每个 perm API logic 在 RPC 调用后立即 `responsex.ToError(grpcResp.Base)`，非零返回 Go error，禁止 deref 响应字段。样板：`listroleslogic.go:54`。
- **修改**（每文件在 RPC 调用后、deref 响应字段前插入 `if err := responsex.ToError(grpcResp.Base); err != nil { return nil, err }`）：
  - `services/permission-service/api/internal/logic/perm/getrolelogic.go`（deref `Role`）
  - `services/permission-service/api/internal/logic/perm/getrolepermissionslogic.go`（deref `Role.Permissions`）
  - `services/permission-service/api/internal/logic/perm/getuserpermissionslogic.go`（deref `PermissionCodes`）
  - `services/permission-service/api/internal/logic/perm/getuserroleslogic.go`（deref `Roles`）
  - `services/permission-service/api/internal/logic/perm/listpermissionslogic.go`（deref `Permissions`）
- **创建**: `services/permission-service/api/internal/logic/perm/getrolelogic_test.go`、`getrolepermissionslogic_test.go`、`getuserpermissionslogic_test.go`、`getuserroleslogic_test.go`、`listpermissionslogic_test.go`（增补 Base 非零用例）
- [ ] **RED**: 代表性用例——GetRole 60001 → error 不 deref；GetRolePermissions 60001 → error 不 deref `Role.Permissions`；GetUserPermissions 60001 → error；GetUserRoles 60001 → error；ListPermissions 业务错误 → error；`go test ./api/internal/logic/perm/` → 看 FAIL
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 1.10: API base-check 类级审计 — 静默类 3 文件（REQ-UPDATE-3）
> 类级规则同上（SEE [[rpc-callback-must-check-response-base]]）：RPC 调用后立即 `responsex.ToError(grpcResp.Base)`，非零返回 Go error，不再静默成功。
- **修改**（每文件 RPC 调用后立即 `responsex.ToError(grpcResp.Base)`，非零 → 返回 Go error，不再静默成功）：
  - `services/permission-service/api/internal/logic/perm/deleterolelogic.go`（DeleteRole）
  - `services/permission-service/api/internal/logic/perm/assignuserrolelogic.go`（AssignRole）
  - `services/permission-service/api/internal/logic/perm/revokeuserrolelogic.go`（RevokeRole）
- **创建**: `services/permission-service/api/internal/logic/perm/deleterolelogic_test.go`、`assignuserrolelogic_test.go`、`revokeuserrolelogic_test.go`
- [ ] **RED**: `TestDeleteRole_Base60004`（系统角色不可删除 → Go error）/ `TestAssignUserRole_BaseError` / `TestRevokeUserRole_BaseError`；`go test -run TestDeleteRole|TestAssignUserRole|TestRevokeUserRole` → 看 FAIL
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 1.11: API AssignRolePermissions — platforms 保留 + GetRole base-check + 中止语义（REQ-PLAT-8）
> SEE [[rpc-callback-must-check-response-base]]（GetRole/UpdateRole 响应均须 base-check）+ [[verify-api-before-calling]]（先 GetRole 验证角色存在，再 UpdateRole）
- **修改**: `services/permission-service/api/internal/logic/perm/assignrolepermissionslogic.go`
- **创建**: `services/permission-service/api/internal/logic/perm/assignrolepermissionslogic_test.go`
- [ ] 先调 `PermissionRpc.GetRole`（`GetRoleRequest{Id: roleId}`）读角色当前 platforms
- [ ] `responsex.ToError(grpcResp.Base)` 非零（如 60001）→ **abort** 返回 Go error，**不调用 UpdateRole**（防 D3 清空端限制）
- [ ] GetRole 返回 Go err → abort
- [ ] 构造 `UpdateRoleRequest{ Id: roleId, PermissionIds: req.PermissionIds, Platforms: grpcResp.Role.Platforms }` → 调 UpdateRole（显式保留现有 platforms）
- [ ] UpdateRole 响应后 `responsex.ToError`（REQ-UPDATE-3）
- [ ] **RED**: 扩展 `mockPermRpc`（增 `getRoleFn`/`updateRoleFn` 字段）——`TestAssignRolePermissions_PreservesPlatforms`（mock GetRole 返回 `Platforms:["mobile"]`，捕获 UpdateRole 请求断言 `Platforms==["mobile"]`）/ `TestAssignRolePermissions_GetRoleBaseAbort`（GetRole Base 60001 → 返回 error 且 updateRoleFn 未被调用）；`go test -run TestAssignRolePermissions` → 看 FAIL
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

---

## 前端 web/pc

### Task 2.1: List.vue 列宽整体重排（REQ-LAYOUT-1/2 / D7）
- **修改**: `web/pc/src/views/roles/List.vue`
- [ ] ID 列 `width="200"` → `width="70"`（`sys_role.id` 为自增小整数，seed 1-2 位）
- [ ] 角色名称 / 角色编码 增 `min-width`（约 120px）自适应，不换行
- [ ] 描述列增 `min-width`（约 200px），保留 `show-overflow-tooltip`（超长省略号截断）
- [ ] 操作列 `width="380"` → `width="260"`（编辑/权限配置/查看用户/删除四按钮单行并排）
- [ ] 系统角色(100)/状态(100)/允许登录端(160)/创建时间(180) 保持固定宽度不变
- [ ] 功能不回退（REQ-LAYOUT-2）：编辑（系统角色保持可用）、权限配置、查看用户、删除（系统角色禁用）、分页 `loadRoles` 原样
- [ ] TDD 不适用（纯 CSS 宽度调整无逻辑）；验证：`cd web/pc && npm run lint`（vue-tsc type check）+ `npm run test:unit`（既有测试绿）
