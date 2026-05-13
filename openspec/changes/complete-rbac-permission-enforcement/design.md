## Context

当前 Identity Service 已建立完整的 RBAC 数据模型（auth_user、auth_role、auth_permission、auth_user_role、auth_role_permission 五张表），Casbin enforcer 已在 ServiceContext 中初始化，rbac_model.conf 已配置 RBAC with domains 模型，角色-权限分配的前后端 CRUD 均已实现。

但权限控制的核心执行层全部为占位实现：
- `permissionmiddleware.go`：Casbin enforce 调用被注释为 TODO，所有认证用户放行
- `get_user_permissions_logic.go`：返回空数组
- `check_permission_logic.go`（RPC）：空函数体
- 前端 `guards.ts`：权限检查代码被注释
- `AppSidebar.vue`：菜单为静态硬编码，无权限过滤
- 用户详情页 `Detail.vue`：角色权限区域显示"Phase 7 实现"

现有 Casbin 模型使用 RBAC with domains：`g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == r.obj && r.act == r.act`，其中 sub=role:N, dom=scopeId, obj=API path, act=HTTP method。

## Goals / Non-Goals

**Goals:**

- 后端 Casbin 中间件真正执行权限拦截，未授权请求返回 403
- 获取用户权限 API 返回完整的权限 code 列表（通过 user_role → role_permission → permission 链路）
- RPC CheckPermission 供网关/其他服务调用验证权限
- 新增用户-角色分配/移除 API 端点
- 前端路由守卫按 meta.permission 字段检查权限
- 侧边栏菜单根据用户权限动态过滤
- 提供 v-permission 指令控制按钮级显示
- 用户详情页展示角色列表并支持角色分配

**Non-Goals:**

- 不修改数据库表结构（现有表已满足需求）
- 不实现 Redis pub/sub 策略热更新（已有框架，本次聚焦核心拦截逻辑）
- 不实现数据行级权限（scope 过滤由各业务 logic 自行处理）
- 不实现前端权限缓存过期策略（刷新页面重新加载即可）
- 不修改 Casbin 模型配置（现有 rbac_model.conf 满足需求）

## Decisions

### D1: Casbin 策略加载方式 — 数据库驱动，而非文件

**选择**: 使用已配置的 PolicyAdapter（数据库适配器），不引入文件策略。

**原因**: Casbin enforcer 已在 ServiceContext 中通过 `c.Casbin.PolicyAdapter` 初始化，策略存储在数据库 casbin_rule 表中。角色分配权限时通过 `SyncRolePolicies` 同步到 casbin_rule 表，无需额外文件。

**替代方案**: 使用文件策略 + 自动加载 — 放弃，因为策略已通过 DB 适配器管理。

### D2: 权限中间件资源映射 — 使用 API path + HTTP method 作为 obj/act

**选择**: `obj = r.URL.Path`，`act = r.Method`。Casbin 策略中的 obj 为 API 路径（如 `/api/identity/users`），act 为 HTTP 方法（如 `GET`、`POST`）。

**原因**: 与现有 rbac_model.conf matchers 一致，无需修改模型。每个 permission 记录的 `code` 字段存储格式为 `METHOD:PATH`（如 `GET:/api/identity/users`），便于在分配权限时将 permission code 转换为 Casbin 策略。

**替代方案**: 使用 permission.code 直接匹配 — 放弃，因为 code 是语义化的（如 `user:list`），不如 path+method 精确。

### D3: 超级管理员豁免逻辑 — 在中间件层跳过 Casbin 检查

**选择**: 中间件检查用户角色，若包含 `super_admin`（is_system=1 的角色），直接放行，不经过 Casbin enforce。

**原因**: 规范要求 Super Administrator 拥有所有权限且不可删除。在 Casbin 中为超级管理员维护全量策略不实际（每次新增权限都要同步），在中间件层豁免更简洁。

**实现**: 通过查询 auth_user_role + auth_role 表判断用户是否拥有系统角色，使用带缓存的查询避免每次请求都查库。

### D4: 用户权限查询 — 直接 SQL 查询而非通过 Casbin

**选择**: `get_user_permissions_logic.go` 通过 `user_role → role → role_permission → permission` SQL 链路查询，返回 permission.code 列表和 type=1 的菜单权限树。

**原因**: Casbin 只存储 role → permission 策略（role:N, code, allow），不包含 permission 的完整信息（name、type、path、children 等）。前端需要完整的权限树来渲染菜单和判断按钮权限，直接查数据库更高效。

### D5: 用户-角色分配 — 新增独立 API 端点而非复用 UpdateUser

**选择**: 新增 `POST /api/identity/users/:id/roles` 和 `DELETE /api/identity/users/:id/roles/:roleId` 端点。

**原因**: 角色分配是独立操作，与用户基本信息更新（昵称、头像等）职责不同。独立端点更清晰，也便于审计日志记录。

### D6: 前端菜单权限 — 按菜单权限 code 匹配路由 path

**选择**: 在 MenuItem 中新增 `permissionCode` 字段，与用户权限 code 列表比对，无权限的菜单项不渲染。

**原因**: 菜单项与后端 permission 记录（type=1, menu）对应，每个菜单 permission 有唯一 code。比对 code 列表即可实现动态过滤。

### D7: v-permission 指令 — 基于 permission code 的元素显隐控制

**选择**: 实现 `v-permission="'user:create'"` 指令，使用 `el-remove` 或 `display:none` 控制元素可见性。

**原因**: Element Plus 生态中 v-permission 是常见模式，实现简单。使用 `display:none` 而非 `v-if` 可保持 DOM 稳定性，避免 v-for 中 key 冲突。

## Risks / Trade-offs

- **[Casbin 策略与数据库不同步]** → SyncRolePolicies 在角色权限变更时同步。若直接操作数据库导致不一致，需提供 `LoadPolicy()` 手动重载接口。本次不实现自动检测机制，依赖正确的业务调用链路。
- **[超级管理员豁免绕过 Casbin]** → 这是可接受的简化。若未来需要更细粒度的超级管理员控制（如仅部分权限），需改为策略全量同步方案。
- **[菜单 permissionCode 硬编码]** → 前端菜单的 permissionCode 需与后端 permission 表数据一致。若后端新增菜单权限但前端未更新 permissionCode，菜单不会显示。通过在开发文档中约定来降低此风险。
- **[v-permission 使用 display:none]** → 隐藏的元素仍在 DOM 中，敏感操作需后端权限校验兜底。后端 Casbin 中间件已提供此保障。
