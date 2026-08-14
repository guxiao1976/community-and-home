# Design: role-platforms-save — 角色管理 bug 修复 + platforms 允许登录端写链路补全

> 唯一权威基线：`proposal.md`（D1-D7 全部用户拍板）+ `specs/*/spec.md`（REQ-PLAT-1~8 / REQ-UPDATE-1~4 / REQ-LAYOUT-1~2）。
> 本设计按依赖序落地：proto → write-path（platforms 写链路）→ update-fix（500 修复 + 字段级策略 + ToError 审计）→ 前端列宽。

## 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| `CreateRoleRequest`/`UpdateRoleRequest` 增 `platforms` 字段 + 错误码 060008 + 生成 | api-proto | Proto 单一权威源（api-proto/CLAUDE.md §1），仅全局 Claude 可改 |
| platforms 写链路（API types + API logic + RPC logic + model） | permission-service | `sys_role.platforms` 数据属主；读链路 `splitPlatforms` 已就绪，补写链路 `joinPlatforms` |
| 系统角色字段级编辑策略（status 拦截、permission_ids 放行） | permission-service | `UpdateRole` RPC 拥有角色数据与系统角色保护策略 |
| 500 panic 修复 + ToError 类级审计（11 文件） | permission-service | `perm` 包全部 API logic 的类级缺陷，整类消除 |
| `sort_order` 写库修复 | permission-service | `SysRoleModel.Update` SQL 遗漏该列（潜伏缺陷，D6） |
| 列表列宽重排 | web/pc | 前端展示层 `List.vue` |

## 数据模型

**无新增表。** `sys_role.platforms` 列已由 `migration/002_add_role_platforms.sql` 建立（`VARCHAR(32) NOT NULL DEFAULT ''`，逗号分隔 pc/mobile，空=fail-open）；`sort_order` 列已存在。

变更点（唯一的数据层改动）：

```go
// services/permission-service/model/permission.go — defaultSysRoleModel.Update
// 现状（缺 sort_order）："update sys_role set role_name = ?, description = ?, status = ?, platforms = ?, updated_at = now() where id = ?"
// 目标（D6）：补 sort_order = ?
update sys_role set role_name = ?, description = ?, status = ?, platforms = ?, sort_order = ?, updated_at = now() where id = ?
```

- `platforms` 存储格式：逗号分隔字符串（`"pc,mobile"`），与 `splitPlatforms`/`joinPlatforms`（已存在，rpc helpers.go）互为逆操作
- 无唯一索引/外键变更，无数据迁移

## 接口设计

### permission.v1.CreateRole
- **输入**：`CreateRoleRequest` = `{ code(1), name(2), description(3), sort_order(4), permission_ids(5), platforms(6, 新增) }`
  - `platforms`：`repeated string`，值域 `pc`/`mobile`；空=fail-open（允许所有端）；非法值拒绝 60008；重复值落库前去重
- **输出**：`CreateRoleResponse` = `{ base, role }`
- **错误码**：`060006` 角色编码已存在；`060008`（新增）非法登录端

### permission.v1.UpdateRole
- **输入**：`UpdateRoleRequest` = `{ id(1), name(2), description(3), status(4), sort_order(5), permission_ids(6), platforms(7, 新增) }`
  - `platforms` 语义（D3）：**无条件覆盖**既有值；空列表=显式清空（fail-open），API 层始终透传（含空列表）
- **输出**：`UpdateRoleResponse` = `{ base, role }`
- **错误码**：
  - `060001` 角色不存在
  - `060004`（UpdateRole 路径语义收窄）：「系统角色状态不可修改」——由「系统角色不可修改」（整单拦截）收窄为「仅 status 拦截」（D1 字段级策略）；`DeleteRole` 路径 60004「系统角色不可删除 / 角色已被分配，无法删除」语义不改（同码异义问题记 backlog，见记忆）
  - `060008`（新增）非法登录端
- **校验顺序（钉死，REQ-UPDATE-4 优先于 REQ-PLAT-4）**：
  1. `FindOne` → 不存在 60001
  2. 系统角色（`is_system=1`）且 `status` 在场 → **60004**（先于任何字段应用、先于 platforms 校验；原子拒绝，无部分写入）
  3. `validatePlatforms(in.Platforms)` → 非法值 **60008**（原子拒绝，任何字段不落库）
  4. 应用字段：name/description/sort_order（在场才覆盖）、`existing.Platforms = joinPlatforms(deduped)`（无条件）、status（仅非系统角色）
  5. `Update`（含 sort_order）→ permission_ids 替换（原逻辑）→ `invalidateRoleCache`（platforms 变更同样触发）

### permission.v1.GetRole / ListRoles / GetUserRoles（读链路）
- `Role.platforms(10)`（已存在）→ `splitPlatforms` 已透出；本变更闭合 HTTP 读链路 `RoleInfo.Platforms`（D2）

### 错误码 060008（新增）
- 注册到 `permission.proto` 头注释错误码块：`060008 — 非法登录端`（既有约定，参照 2026-08-12 登记 060007）
- RPC 层用**命名常量** `CodeInvalidPlatform = 60008`（`rpc/internal/logic/permission/helpers.go`），禁止裸字面量（SEE [[error-code-literal-bypasses-qa-gate]]）

## 业务流程

### 编辑系统角色（Bug 1 修复后）
```
PUT /api/perm/roles/:id { name, description, platforms:["mobile"] }
  → API UpdateRole logic：ToError 前置 + 恒透传 Platforms
  → RPC UpdateRole：FindOne(60001) → 系统角色无 status 放行 → validatePlatforms(通过)
  → 应用 name/description/platforms → Update（status 保持）→ invalidateRoleCache → Role 返回
  → API 层返回 200，不再 500
```

### 创建/编辑自定义角色（platforms 写链路，Bug 1b）
```
POST/PUT { ..., platforms:["pc","mobile"] }
  → API 透传 → RPC CreateRole/UpdateRole：validatePlatforms（非法→60008 原子拒绝）
  → joinPlatforms 落库 "pc,mobile" / 无条件覆盖
  → 读链路 splitPlatforms → RoleInfo.Platforms → 前端回显
```

### 权限独立分配（REQ-PLAT-8 一致性不变量）
```
POST /api/perm/roles/:id/permissions { permissionIds:[...] }
  → AssignRolePermissions API logic：先 GetRole（ToError base-check，60001→abort）
  → 读当前 platforms → UpdateRoleRequest{ Id, PermissionIds, Platforms: 当前值 }
  → UpdateRole → platforms 不被误清（D3 无条件覆盖语义下显式保留）
```

### 系统角色 status 拦截优先级（原子拒绝）
```
系统角色 + { status:0, name:"改名", platforms:["web"] }
  → 60004（status 门禁先于 60008 platforms 校验，先于任何字段应用）→ DB 零写入
```

## Proto 变更

| 文件 | 变更类型 | 说明 |
|------|:---:|------|
| `api-proto/api/permission/v1/permission.proto` | 兼容新增字段 | `CreateRoleRequest` 增 `platforms`(6)；`UpdateRoleRequest` 增 `platforms`(7)；头注释错误码块增 `060008 — 非法登录端`；`UpdateRole` RPC `@description` 补 platforms |
| `api-proto/CHANGELOG.md` | 文档 | 记录本次变更（参照 2026-08-13 access-control 条目格式） |
| `api-proto/gen/go/permission/v1`（生成） | 生成 | `make generate` 重新生成；`repeated string platforms` 兼容，`buf breaking` 应通过 |

**破坏性评估**：无。新增 `repeated string` 字段为向后兼容追加；旧客户端不传该字段时 `len==0`，fail-open 语义（REQ-PLAT-1）。

## 安全考虑

- **Base 类级检查（REQ-UPDATE-3）**：`perm` 包全部 API logic 在 RPC 调用后立即 `responsex.ToError(grpcResp.Base)`，非零返回 Go error，**禁止 deref 响应字段**。防「业务错误静默吞掉 + 空指针 panic」。SEE [[rpc-callback-must-check-response-base]]。
- **60008 原子拒绝（REQ-PLAT-4）**：platforms 校验先于任何字段应用，非法值整单拒绝，无部分写入。
- **60004 原子拒绝（REQ-UPDATE-4）**：系统角色 status 门禁先于 platforms 校验与字段应用。
- **AssignRolePermissions 不得误清 platforms（REQ-PLAT-8）**：先 GetRole 读当前值并 base-check；读取失败必须 abort，禁止带空 platforms 走 UpdateRole（否则 D3 会清掉端限制，解除安全回归）。
- **平台值域权威在后端**：前端仅选项/展示（`List.vue` 已含 `PLATFORM_OPTIONS`），非法端由 RPC 60008 拒绝。SEE [[frontend-business-rule-hardcode]]。
- **系统角色策略不改变权限语义**：字段级编辑放行 name/description/platforms/sort_order/permission_ids，**不**改变 is_system 的权限模型（无特权短路）。SEE [[is-system-no-permission-shortcut]]。

## 记忆引用（设计阶段预防性注入，Step 1.5 产出）

记忆注入报告：匹配 9 个，注入 9 个，不适用 3 个（`migration-must-execute` 本轮无新迁移；`grpc-only-comms` 无跨服务直连；`goctl-logic-stubs` 无新 stub）。

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[rpc-callback-must-check-response-base]] | 接口设计 / 安全 | ToError 类级规则推广 11 文件；listroleslogic 为正确样板 |
| [[error-code-literal-bypasses-qa-gate]] | 接口设计 | 60008 用命名常量 `CodeInvalidPlatform`，禁止裸数字 |
| [[error-code-collision-and-namespace-alignment]] | 接口设计 | 60004 语义收窄仅限 UpdateRole 路径；DeleteRole 路径 60004 语义不改；同码异义记 backlog |
| [[permission-cache-consistency]] | 业务流程 | platforms 变更后 `invalidateRoleCache`（现有模式复用，含 perm:user / perm:scopes 系列） |
| [[is-system-no-permission-shortcut]] | 安全考虑 | 字段级编辑策略不改变 is_system 权限模型 |
| [[proto-jstype]] | 接口设计 | platforms 为 `repeated string` 无需 jstype；时间戳字段沿用 `json:",string"` |
| [[edit-form-data-integrity]] | 接口设计 | `RoleInfo.Platforms` 读链路闭合，前端编辑回显完整（8 层链路检查） |
| [[verify-api-before-calling]] | 安全考虑 | AssignRolePermissions 先 GetRole（base-check）再 UpdateRole |
| [[tdd-red-evidence-requires-fail-excerpt]] | 任务 | 所有含逻辑 Task 必须带 RED FAIL 证据摘录 |

## 不做清单（Won't have，与 proposal 一致）

- 前端不新增 sort_order 编辑控件、不改状态开关
- 不做端准入权威判定（auth-service 负责）
- 不改 `init_permissions.sql` seed 数据
- 不重构存量错误码魔数（60001/60004/60006 命名常量化为独立 backlog）
- 不改 `Role.is_system` 既有误导注释（记 backlog）
- 不引入破坏性 proto 变更 / 不改 `Role` 既有字段
