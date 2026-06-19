# Tasks: RBAC 管理界面

> 每个 Task 独立可测，按 TDD 执行。精确到文件路径。
> 遵循 [[testing-discipline]]：Static → Unit → Integration → Contract → E2E

---

## 全局 / Proto（由全局 Claude 执行）

### Task 0.1: Proto 定义 - 新增 ListPermissions

**文件**: `api-proto/api/permission/v1/permission.proto`

- [ ] 新增 message `ListPermissionsReq`（空消息）
- [ ] 新增 message `ListPermissionsResp`（含 `repeated PermissionInfo permissions`）
- [ ] 新增 message `PermissionInfo`（含 id, parent_id, name, code, type, path, sort_order）
- [ ] 为所有 `int64` ID 字段标注 `[jstype = JS_STRING]`（id, parent_id）
- [ ] 新增 RPC `rpc ListPermissions(ListPermissionsReq) returns (ListPermissionsResp);`

**记忆**: [[proto-jstype]] - int64 ID 必须加 `[jstype = JS_STRING]`

---

### Task 0.2: Proto 定义 - 增强 GetRoleResp

**文件**: `api-proto/api/permission/v1/permission.proto`

- [ ] 修改 message `GetRoleResp`，增加字段 `repeated int64 permission_ids = 3 [jstype = JS_STRING];`
- [ ] 确认字段编号不冲突

---

### Task 0.3: Proto 生成 + CI

**命令**: 

```bash
cd api-proto
make generate
make lint
make breaking-check
```

- [ ] `make generate` 成功生成 Go 代码（`api/permission/v1/*.pb.go`）
- [ ] `make generate` 成功生成 TS 代码（TypeScript 输出目录）
- [ ] `make lint` 通过（buf lint）
- [ ] `make breaking-check` 通过（无破坏性变更）

**记忆**: [[pre-commit-checks]] - 提交前必须通过机械化检查

---

## permission-service

### Task 1.1: ListPermissions - Model 查询方法

**文件**: `services/permission-service/internal/model/syspermissionmodel.go`

- [ ] 新增方法 `FindAllEnabled(ctx context.Context) ([]*SysPermission, error)`
- [ ] 查询条件：`WHERE status = 1 ORDER BY parent_id, sort_order`

**TDD**:

- [ ] **RED**: 编写测试 `services/permission-service/internal/model/syspermissionmodel_test.go`
  - 测试用例：查询所有启用权限，验证数量和排序
- [ ] **确认 RED**: `go test -run TestFindAllEnabled` → FAIL（方法未实现）
- [ ] **GREEN**: 实现 `FindAllEnabled` 方法
- [ ] **确认 GREEN**: `go test -run TestFindAllEnabled` → PASS
- [ ] **REFACTOR**: 清理代码，保持测试绿

---

### Task 1.2: ListPermissions - Logic 层

**创建**: `services/permission-service/rpc/internal/logic/listpermissionslogic.go`
**创建**: `services/permission-service/rpc/internal/logic/listpermissionslogic_test.go`

- [ ] **RED**: 编写 table-driven tests
  - Case 1: 正常查询，返回所有权限
  - Case 2: 空数据库，返回空列表
- [ ] **确认 RED**: `go test -run TestListPermissions` → FAIL
- [ ] **GREEN**: 实现 `ListPermissions` 方法
  - 调用 `PermissionModel.FindAllEnabled()`
  - 转换为 `pb.PermissionInfo` 列表
  - 返回 `ListPermissionsResp`
- [ ] **确认 GREEN**: `go test -run TestListPermissions` → PASS
- [ ] **REFACTOR**: 提取转换逻辑到 helper 函数

**记忆**: [[api-response-single-wrap]] - Logic 返回 Proto 响应类型（gRPC 直接返回）

---

### Task 1.3: GetRole - Logic 增强（含 permission_ids）

**修改**: `services/permission-service/rpc/internal/logic/getrolelogic.go`
**修改**: `services/permission-service/rpc/internal/logic/getrolelogic_test.go`

- [ ] 新增 Model 方法 `FindPermissionIdsByRoleId(ctx, roleId) ([]int64, error)`
  - 文件：`services/permission-service/internal/model/relrolepermissionmodel.go`
  - SQL: `SELECT permission_id FROM rel_role_permission WHERE role_id = ?`
- [ ] **RED**: 编写测试用例
  - Case 1: 角色有权限，返回 permission_ids
  - Case 2: 角色无权限，返回空列表
- [ ] **确认 RED**: `go test -run TestGetRole` → FAIL
- [ ] **GREEN**: 修改 `GetRole` 方法
  - 查询角色基本信息（已有逻辑）
  - 调用 `FindPermissionIdsByRoleId(role.Id)`
  - 填充 `resp.PermissionIds`
- [ ] **确认 GREEN**: `go test -run TestGetRole` → PASS
- [ ] **REFACTOR**: 清理代码

---

### Task 1.4: 数据初始化脚本

**创建**: `services/permission-service/scripts/init_rbac_data.sql`

- [ ] 初始化 4 个系统角色（`INSERT IGNORE`）
  - owner, property_admin, community_admin, grid_worker
- [ ] 初始化权限树示例（菜单 → 按钮 → API）
  - 用户管理模块（user:menu, user:read, user:create, ...）
  - 角色管理模块（role:menu, role:read, role:permission, ...）
- [ ] 系统角色自动关联所有权限
  - `INSERT IGNORE INTO rel_role_permission SELECT r.id, p.id FROM sys_role r, sys_permission p WHERE r.is_system = 1`
- [ ] 手动执行脚本验证
  - `mysql -u root -p permission < scripts/init_rbac_data.sql`
  - 查询验证：`SELECT COUNT(*) FROM sys_permission;`

---

### Task 1.5: Static Analysis + Integration Tests

- [ ] **Static**: `cd services/permission-service && go vet ./...` → PASS
- [ ] **Integration**: 启动 permission-service，curl 测试
  - `grpcurl -plaintext -d '{}' localhost:8084 permission.v1.PermissionService/ListPermissions`
  - 验证返回权限列表
  - `grpcurl -plaintext -d '{"role_id":"1"}' localhost:8084 permission.v1.PermissionService/GetRole`
  - 验证返回 permission_ids 字段

---

### Task 1.6: Harness 检查

**命令**: 

```bash
bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service
```

- [ ] 15 项检查全部 PASS 或仅有已知 WARN
- [ ] 有任何 FAIL → 修复后重新检查

**记忆**: [[pre-commit-checks]] - 提交前必须通过机械化检查

---

## web/pc 前端

### Task 2.1: TypeScript 类型定义

**创建**: `web/pc/src/types/permission.ts`

- [ ] 定义 `Permission` 接口（id, parent_id, name, code, type, path, sort_order, children?）
- [ ] 定义 `Role` 接口（复用或增强现有类型）
- [ ] 定义 `UserRole` 接口（id, user_id, role_id, scope_type, scope_id, ...）
- [ ] 所有 ID 字段类型为 `string`（对应 Proto `[jstype = JS_STRING]`）

**记忆**: [[proto-jstype]] - 前端 ID 类型为 string

---

### Task 2.2: API 模块 - 权限相关接口

**修改**: `web/pc/src/api/permission.ts`

- [ ] 新增 `getPermissions()` 函数 → `GET /api/v1/permission/permissions`
- [ ] 修改 `getRoleDetail(id)` 函数，确认返回类型含 `permission_ids: string[]`
- [ ] 新增 `updateRolePermissions(roleId, permissionIds)` 函数 → `PUT /api/v1/permission/roles/:id/permissions`
- [ ] 新增 `getUserPermissions(userId)` 函数 → `GET /api/v1/permission/users/:userId/permissions`

---

### Task 2.3: 权限树工具函数

**创建**: `web/pc/src/utils/permission.ts`

- [ ] 函数 `buildPermissionTree(permissions: Permission[]): Permission[]`
  - 算法：扁平列表 → 树形结构（按 parent_id）
  - 排序：按 sort_order
- [ ] 函数 `filterMenus(routes: RouteRecordRaw[], permissions: string[]): RouteRecordRaw[]`
  - 过滤路由：检查 `route.meta.permissions`
  - 递归过滤子路由
- [ ] **测试**: 编写 Vitest 单元测试
  - `web/pc/src/utils/permission.spec.ts`
  - Case 1: 扁平列表正确构建树
  - Case 2: 菜单过滤正确

---

### Task 2.4: Pinia Store - 用户权限

**修改**: `web/pc/src/stores/user.ts`

- [ ] 增加 state `permissions: string[]`
- [ ] 增加 action `fetchPermissions()` - 调用 `getUserPermissions()`
- [ ] 增加 action `restorePermissions()` - 从 LocalStorage 恢复
- [ ] 修改 action `logout()` - 清除 permissions 缓存

---

### Task 2.5: v-permission 指令

**创建**: `web/pc/src/directives/permission.ts`

- [ ] 定义 `permission` 指令
  - mounted: 检查用户权限，无权限则移除 DOM 节点
  - 支持单个权限：`v-permission="'user:create'"`
  - 支持多个权限（任一满足）：`v-permission="['user:create', 'user:update']"`
- [ ] **注册**: `web/pc/src/main.ts`
  - `app.directive('permission', permission)`

---

### Task 2.6: 路由守卫

**创建**: `web/pc/src/router/permission.ts`

- [ ] 白名单：`['/login', '/register', '/404', '/403']`
- [ ] `router.beforeEach` 逻辑
  - 白名单直接放行
  - 检查登录状态（token）
  - 权限未加载 → 调用 `fetchPermissions()`
  - 检查路由权限（`route.meta.permissions`）
  - 无权限 → 跳转 `/403`
- [ ] **注册**: `web/pc/src/main.ts` 引入 `./router/permission`

---

### Task 2.7: 403 页面

**创建**: `web/pc/src/views/error/403.vue`

- [ ] el-result 组件：图标 warning，标题 "403"，副标题 "您没有权限访问此页面"
- [ ] "返回首页" 按钮 → `$router.push('/')`
- [ ] **注册路由**: `web/pc/src/router/index.ts`
  - `{ path: '/403', name: '403', component: () => import('@/views/error/403.vue') }`

---

### Task 2.8: 角色权限配置页面

**创建**: `web/pc/src/views/roles/Permissions.vue`

- [ ] 页面布局
  - el-page-header（返回 + 标题"权限配置 - {角色名}"）
  - el-alert 提示信息
  - el-card 包裹权限树
  - 底部按钮：取消、保存
- [ ] 数据加载
  - `onMounted`: 调用 `getRoleDetail(roleId)` + `getPermissions()`
  - 构建权限树：`buildPermissionTree(permissions)`
  - 设置勾选：`treeRef.setCheckedKeys(permission_ids)`
- [ ] el-tree 配置
  - `node-key="id"`
  - `default-expand-all`
  - `show-checkbox`
  - `check-strictly=false`（父子联动）
  - 自定义节点模板（图标 + 名称 + 类型标签 + path）
- [ ] 保存逻辑
  - 收集勾选节点：`treeRef.getCheckedKeys()`
  - 调用 `updateRolePermissions(roleId, checkedKeys)`
  - 成功 → ElMessage.success + 跳转 `/roles`
- [ ] **路由注册**: `web/pc/src/router/modules/identity.ts`
  - `{ path: '/roles/:id/permissions', name: 'RolePermissions', meta: { permissions: ['role:permission'] } }`

**记忆**: [[change-verification-checklist]] - 前端改动后验证：编译 → 视觉检查 → 实际操作

---

### Task 2.9: 角色列表增强（系统角色保护）

**修改**: `web/pc/src/views/roles/List.vue`

- [ ] 表格列增加"角色类型"（Tag: is_system=1 → 系统，否则 → 自定义）
- [ ] 操作列"删除"按钮
  - `:disabled="row.is_system === 1"`（系统角色禁用）
  - Tooltip 提示："系统角色不可删除"
- [ ] 操作列"权限配置"按钮
  - `@click="$router.push(\`/roles/\${row.id}/permissions\`)"`
  - `v-permission="'role:permission'"`

---

### Task 2.10: 用户角色分配页面

**创建**: `web/pc/src/views/identity/UserRoles.vue`

- [ ] 页面布局
  - 顶部用户信息卡片（用户名、手机号、状态）
  - "分配角色"按钮（`v-permission="'user:assign-role'"`）
  - el-table 展示用户角色列表
  - 分页器
- [ ] 表格列：角色名称、角色编码、作用域类型（Tag）、作用域名称、分配时间、操作（撤销）
- [ ] 分配角色对话框
  - 表单字段：用户信息（只读）、角色选择（el-select）、作用域类型（el-select）、作用域实体（el-cascader）
  - 作用域级联选择器：根据 scope_type 动态加载（调用 master-data-service）
  - 提交：`assignRole(userId, roleId, scopeType, scopeId)`
- [ ] 撤销角色
  - 二次确认对话框
  - 提交：`revokeRole(userId, roleId, scopeType, scopeId)`
- [ ] **路由注册**: `{ path: '/users/:userId/roles', name: 'UserRoles', meta: { permissions: ['user:assign-role'] } }`

---

### Task 2.11: 侧边栏菜单过滤

**修改**: `web/pc/src/components/layout/AppSidebar.vue`

- [ ] 计算属性 `menus`
  - 获取路由列表：`router.getRoutes().filter(r => !r.meta?.hidden)`
  - 过滤：`filterMenus(routes, userStore.permissions)`
- [ ] 渲染：遍历过滤后的 `menus`

---

### Task 2.12: 前端 E2E 测试

- [ ] **登录 → 权限加载**
  - 登录成功，LocalStorage 存储 permissions
  - 侧边栏菜单过滤（无权限菜单不显示）
- [ ] **角色管理 → 权限配置**
  - 访问 `/roles`，点击"权限配置"
  - 权限树正确渲染，勾选保存生效
  - 返回列表页
- [ ] **用户角色分配**
  - 访问 `/users/:userId/roles`
  - 分配角色并指定作用域
  - 撤销角色
- [ ] **菜单权限控制**
  - 手动输入无权限路由 → 自动跳转 `/403`
  - 按钮级权限：无权限按钮不显示

---

### Task 2.13: 前端静态检查 + 构建

- [ ] **Static**: `cd web/pc && npx vue-tsc --noEmit` → PASS
- [ ] **Build**: `cd web/pc && npm run build` → SUCCESS
- [ ] **Vitest**: `cd web/pc && npm run test:unit` → PASS

---

## user-service（权限中间件示例）

### Task 3.1: 权限中间件实现

**创建**: `services/user-service/api/internal/middleware/permissionmiddleware.go`

- [ ] 中间件函数 `PermissionMiddleware(permClient permissionclient.Client) rest.Middleware`
  - 提取 user_id（从 JWT context）
  - 白名单检查（`/api/v1/health`, `/api/v1/auth/login`）
  - 调用 `permClient.CheckPermission(userId, method, path)`
  - allowed=true → next()
  - allowed=false → http.Error(403)
- [ ] 日志记录：权限拒绝时记录 user_id, path, method

**记忆**: [[grpc-only-comms]] - 调用 permission-service RPC

---

### Task 3.2: ServiceContext 注入 PermissionClient

**修改**: `services/user-service/api/internal/svc/servicecontext.go`

- [ ] 增加字段 `PermissionClient permissionclient.Client`
- [ ] NewServiceContext 中初始化
  - `permConn := zrpc.MustNewClient(c.PermissionRpc)`
  - `PermissionClient: permissionclient.NewClient(permConn)`

---

### Task 3.3: 配置文件增加 PermissionRpc

**修改**: `services/user-service/api/etc/userservice.yaml`

- [ ] 增加配置段
  
  ```yaml
  PermissionRpc:
    Etcd:
      Hosts:
        - 127.0.0.1:2379
      Key: permission.rpc
  ```

---

### Task 3.4: 注册权限中间件

**修改**: `services/user-service/api/internal/handler/routes.go`

- [ ] 在 `RegisterHandlers` 函数中
  - JWT 中间件之后，注册权限中间件
  - `server.Use(middleware.PermissionMiddleware(serverCtx.PermissionClient))`

---

### Task 3.5: user-service 集成测试

- [ ] 启动 user-service（确保 permission-service 已启动）
- [ ] curl 测试无权限 API → 返回 403
  - `curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/user/users`
  - 验证：用户无 `user:read:api` 权限 → 403
- [ ] curl 测试有权限 API → 返回 200
  - 用 owner 角色 token → 所有 API 返回 200

---

### Task 3.6: user-service Harness 检查

**命令**: 

```bash
bash .harness/skills/qa/scripts/harness-checks.sh --service user-service
```

- [ ] 15 项检查全部 PASS 或仅有已知 WARN

---

## 全局验证（Contract + E2E）

### Task 4.1: Contract Tests

- [ ] **Proto 向后兼容**: `cd api-proto && make breaking-check` → PASS
- [ ] **Snowflake ID 序列化**: harness-checks.sh 检测 proto `[jstype = JS_STRING]`
- [ ] **API 响应格式**: curl 任意 permission API，验证单层 `{code, msg, data}` 包装

**记忆**: [[api-response-single-wrap]] - Logic 返回纯数据，Handler 包装

---

### Task 4.2: E2E 核心流程

**流程 1: 角色权限配置 → 用户权限生效**

1. 创建自定义角色"测试角色"
2. 配置权限（仅 `user:menu`, `user:read`）
3. 创建测试用户，分配"测试角色"
4. 测试用户登录
5. 验证：侧边栏只显示"用户管理"菜单，"创建用户"按钮不显示
6. curl 调用创建用户 API → 返回 403

**流程 2: 系统角色全权限**

1. owner 角色用户登录
2. 验证：所有菜单显示，所有按钮可见
3. curl 调用任意 API → 返回 200

**流程 3: 权限变更实时生效**

1. 为"测试角色"增加 `user:create` 权限
2. 测试用户刷新页面
3. 验证："创建用户"按钮显示
4. curl 调用创建用户 API → 返回 200

---

### Task 4.3: 性能测试

- [ ] 权限检查 P99 < 100ms
  - 工具：ab 或 wrk，100 并发
  - 命令示例：`ab -n 1000 -c 100 -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/user/users`
- [ ] 权限树加载时间 < 500ms（前端）
  - Chrome DevTools Network 面板查看

---

## 自检清单（Self-Review）

### 占位符检查

- [ ] 搜索 tasks.md 中是否有 `<任务描述>` / `TBD` / `TODO`
- [ ] 搜索 tasks.md 中是否有 `...` / `示例` / `伪代码`
- [ ] 所有文件路径是否精确（无占位符）

### TDD 覆盖

- [ ] 所有含逻辑的 Task 是否包含 RED → GREEN → REFACTOR 步骤
- [ ] permission-service Logic 层是否有单元测试
- [ ] 前端工具函数是否有 Vitest 单元测试

### 依赖顺序

- [ ] Proto 变更 → permission-service → 前端（顺序正确）
- [ ] 数据初始化 → 集成测试（顺序正确）
- [ ] 基础设施（指令/守卫）→ 页面组件（顺序正确）

### 独立可测

- [ ] 每个 Task 是否能独立完成（无循环依赖）
- [ ] 每个 Task 的验收标准是否明确

### 记忆引用

- [ ] [[proto-jstype]] - Task 0.1, 0.2, 2.1 已引用
- [ ] [[grpc-only-comms]] - Task 3.1 已引用
- [ ] [[pre-commit-checks]] - Task 0.3, 1.6, 3.6 已引用
- [ ] [[testing-discipline]] - 文件头部已声明
- [ ] [[api-response-single-wrap]] - Task 4.1 已引用
- [ ] [[change-verification-checklist]] - Task 2.8 已引用

**高风险 Task 记忆引用检查**：

- Task 0.1（Proto 变更）→ [[proto-jstype]] ✅
- Task 1.6（Harness 检查）→ [[pre-commit-checks]] ✅
- Task 3.1（权限中间件）→ [[grpc-only-comms]] ✅
- Task 4.1（Contract Tests）→ [[api-response-single-wrap]] ✅

---

## 总结

**Task 统计**：

- 全局 / Proto 组：3 个 Task
- permission-service 组：6 个 Task
- web/pc 组：13 个 Task
- user-service 组：6 个 Task
- 全局验证组：3 个 Task
- **总计**：31 个 Task

**预估工作量**：约 19.5 小时（见 design.md § 八.1）

**记忆注入报告**：匹配 6 个 must-follow 记忆，注入 6 个，不适用 0 个

**自检结果**：

- 占位符：0 个
- TDD 覆盖：5 层测试全覆盖
- 依赖顺序：Proto → 后端 → 前端，正确
- 独立可测：每个 Task 验收标准明确
- 记忆引用：高风险 Task 全部引用相关记忆
