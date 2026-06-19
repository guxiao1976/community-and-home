# Design: RBAC 管理界面

## 一、服务归属决策

| 功能模块 | 归属服务/目录 | 理由 |
|---------|-------------|------|
| Proto 变更 | `api-proto/api/permission/v1/` | 新增 gRPC 接口定义（GetRole 增强、ListPermissions） |
| 权限 gRPC 接口增强 | `permission-service/` | 拥有权限数据（sys_permission, rel_role_permission），提供查询接口 |
| 角色管理界面 | `web/pc/src/views/roles/` | 前端页面：List.vue（已存在，需增强），Permissions.vue（权限配置） |
| 权限树组件 | `web/pc/src/components/business/PermissionTree.vue` | 可复用的权限树组件（已存在） |
| 用户角色分配界面 | `web/pc/src/views/identity/` | 嵌入用户详情页面或独立页面 |
| 菜单权限控制 | `web/pc/src/router/` + `web/pc/src/directives/` | 路由守卫 + v-permission 指令 |
| 权限中间件（示例） | `services/user-service/internal/middleware/` | 示例集成，后续推广到其他服务 |

**前端涉及页面**：
- `/roles` - 角色列表（增强现有）
- `/roles/:id/permissions` - 角色权限配置（新建）
- `/users/:userId/roles` - 用户角色分配（新建）
- 路由守卫（全局）
- 403 页面（新建）

---

## 二、数据模型

### 2.1 数据库变更

**无表结构变更**。现有 4 张表已满足需求：
- `sys_role` - 角色定义（已有 `is_system` 字段）
- `sys_permission` - 权限定义（已有 `parent_id`, `type`, `path`）
- `rel_role_permission` - 角色-权限关联
- `rel_user_role` - 用户-角色关联（已有 `scope_type`, `scope_id`）

**需要数据初始化**（SQL 脚本）：
```sql
-- 1. 初始化 4 个系统角色（如果不存在）
INSERT IGNORE INTO sys_role (id, role_code, role_name, is_system, status, sort_order) VALUES
(1, 'owner', '业主', 1, 1, 10),
(2, 'property_admin', '物业管理员', 1, 1, 20),
(3, 'community_admin', '社区管理员', 1, 1, 30),
(4, 'grid_worker', '网格员', 1, 1, 40);

-- 2. 初始化权限树（菜单 → 按钮 → API）
-- 示例：用户管理模块
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, sort_order, status) VALUES
-- 菜单层
(100, 0, '用户管理', 'user:menu', 1, NULL, 10, 1),
(200, 0, '角色管理', 'role:menu', 1, NULL, 20, 1),
-- 按钮层
(110, 100, '查看用户', 'user:read', 2, NULL, 10, 1),
(120, 100, '创建用户', 'user:create', 2, NULL, 20, 1),
(210, 200, '查看角色', 'role:read', 2, NULL, 10, 1),
(220, 200, '配置权限', 'role:permission', 2, NULL, 20, 1),
-- API 层
(111, 110, 'GET /api/v1/user/users', 'user:read:api', 3, '/api/v1/user/users', 10, 1),
(121, 120, 'POST /api/v1/user/users', 'user:create:api', 3, '/api/v1/user/users', 10, 1);

-- 3. 系统角色自动关联所有权限
INSERT IGNORE INTO rel_role_permission (role_id, permission_id)
SELECT r.id, p.id FROM sys_role r, sys_permission p WHERE r.is_system = 1;
```

### 2.2 前端 TypeScript 类型

参考 specs 中的定义，核心类型：
- `Role` - 角色信息（含 `is_system` 标志）
- `Permission` - 权限信息（含 `parent_id`, `type`, `path`）
- `UserRole` - 用户角色关联（含 `scope_type`, `scope_id`）

**关键约束**：所有 ID 字段类型为 `string`（对应 Proto `[jstype = JS_STRING]`）

---

## 三、接口设计

### 3.1 Proto 新增/增强消息

**文件**：`api-proto/api/permission/v1/permission.proto`

**新增消息**：
```protobuf
// 权限树列表（扁平结构，前端构建树）
message ListPermissionsReq {}

message ListPermissionsResp {
  common.v1.BaseResp base_resp = 1;
  repeated PermissionInfo permissions = 2;
}

message PermissionInfo {
  int64 id = 1 [jstype = JS_STRING];
  int64 parent_id = 2 [jstype = JS_STRING];  // 0 表示根节点
  string name = 3;
  string code = 4;                           // 权限码（如 user:read）
  int64 type = 5;                            // 1=菜单 2=按钮 3=API
  string path = 6;                           // API 路径（type=3 时）
  int64 sort_order = 7;
}

// GetRole 响应增强（已有 RPC，增强返回字段）
message GetRoleResp {
  common.v1.BaseResp base_resp = 1;
  RoleInfo role = 2;
  repeated int64 permission_ids = 3 [jstype = JS_STRING];  // 该角色拥有的权限ID列表
}

// 用户权限码列表（已存在接口，确认返回格式）
message GetUserPermissionsResp {
  common.v1.BaseResp base_resp = 1;
  repeated string permissions = 2;  // 权限码列表（如 ["user:menu", "user:read"]）
}
```

**新增 RPC**：
```protobuf
rpc ListPermissions(ListPermissionsReq) returns (ListPermissionsResp);
```

**修改 RPC**：
- `GetRole` - 响应增加 `permission_ids` 字段
- 其他接口（AssignRole, RevokeRole, GetUserRoles, GetUserPermissions）已存在，无需修改

### 3.2 permission-service 新增 Logic

| Logic | 文件 | 职责 |
|-------|------|------|
| `ListPermissionsLogic` | `internal/logic/listpermissionslogic.go` | 查询所有启用权限（status=1），返回扁平列表 |
| `GetRoleLogic` 增强 | `internal/logic/getrolelogic.go` | 查询角色详情 + 该角色的 permission_ids |

### 3.3 前端 REST API 路由

**通过 API Gateway 转发到 permission-service**：
- `GET /api/v1/permission/permissions` → `ListPermissions`
- `GET /api/v1/permission/roles/:id` → `GetRole`（响应含 permission_ids）
- `PUT /api/v1/permission/roles/:id/permissions` → `UpdateRole`（传递 permission_ids）
- `GET /api/v1/permission/users/:userId/permissions` → `GetUserPermissions`
- 其他角色管理接口（CreateRole, UpdateRole, DeleteRole, ListRoles）已存在

### 3.4 错误码

复用 permission-service 现有错误码：
- `060001` - 角色不存在
- `060002` - 权限不存在
- `060003` - 访问被拒绝
- `060004` - 角色已被分配，不可删除
- `060005` - 不支持的数据范围类型
- `060006` - 角色编码已存在

新增错误码：
- `060007` - 用户至少需要保留一个角色（撤销角色前置检查）

---

## 四、业务流程

### 4.1 角色权限配置流程

```
1. 用户访问 /roles/:id/permissions
2. 前端加载数据：
   a. GetRole(role_id) → 获取角色名称 + permission_ids
   b. ListPermissions() → 获取所有权限（扁平列表）
3. 前端构建权限树（按 parent_id 和 sort_order）
4. 前端勾选已有权限（default-checked-keys = permission_ids）
5. 用户勾选/取消权限，点击"保存"
6. 前端收集所有勾选节点ID（el-tree.getCheckedKeys）
7. 调用 UpdateRole(role_id, permission_ids)
8. 后端：
   a. DELETE FROM rel_role_permission WHERE role_id = ?
   b. BatchInsert 新关联
   c. Redis DEL perm:user:* （失效所有用户权限缓存）
9. 返回成功，前端跳转回角色列表
```

### 4.2 菜单权限控制流程

```
1. 用户登录，auth-service 返回 JWT（含 user_id）
2. 前端调用 GetUserPermissions(user_id) → 获取权限码列表
3. 存入 Pinia Store + LocalStorage
4. 路由守卫检查 route.meta.permissions
5. 有权限 → 放行；无权限 → 跳转 /403
6. 侧边栏渲染时，过滤无权限菜单（filterMenus）
7. 页面内按钮使用 v-permission 指令控制显示
```

### 4.3 接口权限控制流程

```
1. 外部请求 → REST API Gateway（可选，本次不实现）
2. 服务 REST 层 → JWT 中间件（提取 user_id）
3. 权限中间件 → CheckPermission(user_id, method, path)
4. permission-service：
   a. 查 Redis perm:user:{user_id} → HIT 返回结果
   b. MISS → 查 DB（合并所有角色权限）→ 写入 Redis
   c. 匹配权限码 → 返回 allowed=true/false
5. allowed=true → 继续；allowed=false → 返回 403
```

### 4.4 数据范围隔离流程

```
1. 用户分配角色时指定 scope_type + scope_id
   （如：user_id=123, role_id=2, scope_type=community, scope_id=101）
2. 用户查询数据时，服务调用 GetDataScopes(user_id, scope_type='community')
3. permission-service 返回 [101, 102]（该用户在多个小区的角色）
4. 服务 SQL 自动加 WHERE community_id IN (101, 102)
```

---

## 五、Proto 变更清单

| 文件 | 变更类型 | 内容 |
|------|---------|------|
| `api-proto/api/permission/v1/permission.proto` | 新增 message | `ListPermissionsReq`, `ListPermissionsResp`, `PermissionInfo` |
| `api-proto/api/permission/v1/permission.proto` | 修改 message | `GetRoleResp` 增加 `repeated int64 permission_ids` 字段 |
| `api-proto/api/permission/v1/permission.proto` | 新增 RPC | `rpc ListPermissions(ListPermissionsReq) returns (ListPermissionsResp)` |

**破坏性检查**：
- `GetRoleResp` 增加字段（向后兼容，非破坏性）
- 新增 RPC（非破坏性）

**代码生成**：
```bash
cd api-proto
make generate   # 生成 Go/TS 代码
make lint       # 检查规范
make breaking-check  # 检查破坏性变更
```

---

## 六、安全考虑

### 6.1 系统角色保护

- 系统角色（`is_system=1`）不可删除（前端按钮禁用 + 后端校验）
- 系统角色在 `CheckPermission` 中直接返回 `allowed=true`（天然全权限）
- 至少保留一个 owner 角色账号（防止锁死）

### 6.2 权限校验防绕过

- **前端控制**：菜单/按钮隐藏（可被浏览器开发者工具绕过）
- **后端兜底**：REST API 权限中间件（强制拦截）
- **缓存失效**：角色变更后立即失效 Redis `perm:user:*`

### 6.3 最后一个角色保护

- `RevokeRole` 前检查用户角色数量
- 若只剩一个角色，拒绝撤销（错误码 060007）
- 例外：owner 角色可撤销（系统管理员特权）

### 6.4 敏感操作审计

- 角色权限变更记录日志（user_id, action, role_id, timestamp）
- 权限校验失败记录日志（user_id, path, method, timestamp, ip）

---

## 七、记忆引用

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[proto-jstype]] | § 三、接口设计 | 所有 `int64` ID 字段添加 `[jstype = JS_STRING]`，前端类型为 `string` |
| [[grpc-only-comms]] | § 一、服务归属决策 | 服务间通信仅通过 gRPC，权限中间件调用 permission-service RPC |
| [[pre-commit-checks]] | § 八、实施策略（未写入，补充） | 提交前运行 `harness-checks.sh --service permission-service` |
| [[testing-discipline]] | § 八、实施策略（未写入，补充） | 5 层测试：Static → Unit → Integration → Contract → E2E |
| [[api-response-single-wrap]] | § 三.3、前端 REST API | Logic 返回纯业务数据，Handler 统一用 `response.Success(w, data)` 包装 |
| [[change-verification-checklist]] | § 八、实施策略（未写入，补充） | 改完代码后执行验证清单（编译 → 测试 → 前后端对齐 → 部署验证） |

---

## 八、实施策略

### 8.1 分阶段交付

| 阶段 | 交付物 | 验收标准 | 预估工作量 |
|------|--------|---------|----------|
| **P0** | Proto 变更 + 代码生成 | `make ci` 通过 | 0.5h |
| **P1** | permission-service 接口增强 | 新 RPC 可调用，单元测试通过 | 2h |
| **P2** | 前端角色管理界面增强 | 可创建/编辑/删除自定义角色，系统角色保护生效 | 3h |
| **P3** | 前端权限树管理 + 角色权限配置 | 权限树渲染正确，保存权限生效 | 4h |
| **P4** | 前端用户角色分配界面 | 可为用户分配角色并指定作用域 | 3h |
| **P5** | 前端菜单权限控制 | 路由守卫生效，无权限菜单不显示 | 3h |
| **P6** | 前端按钮权限指令 | `v-permission` 指令生效 | 1h |
| **P7** | 后端权限中间件（示例） | user-service 集成权限中间件，API 校验生效 | 2h |
| **P8** | 数据初始化脚本 | 4 个系统角色 + 权限树初始化完成 | 1h |

**总计**：约 19.5 小时

### 8.2 回滚策略

- **Proto 变更**：仅新增，不破坏现有接口
- **前端**：Feature Flag 控制新界面入口（环境变量 `VITE_ENABLE_RBAC_UI`）
- **后端中间件**：配置开关（`EnablePermissionCheck: false` 时跳过）

### 8.3 测试策略

按 [[testing-discipline]] 5 层测试：

**Static Analysis**：
- Go: `go vet ./...`（permission-service）
- TS: `npx vue-tsc --noEmit`（web/pc）
- Proto: `cd api-proto && make lint`

**Unit Tests**：
- `ListPermissionsLogic` 单元测试（查询 + 排序）
- `GetRoleLogic` 增强单元测试（含 permission_ids）
- 前端权限树构建算法测试（扁平 → 树形）
- 前端 `v-permission` 指令测试

**Integration Tests**：
- permission-service RPC 集成测试（真实 DB + Redis）
- 前端 E2E：登录 → 角色列表 → 权限配置 → 保存 → 验证生效

**Contract Tests**：
- `buf breaking-check`（Proto 向后兼容）
- `harness-checks.sh`（Snowflake ID 序列化、响应格式）

**E2E Tests**：
- 完整流程：登录 → 创建角色 → 配置权限 → 分配用户 → 用户登录 → 菜单过滤 → API 权限校验

---

## 九、依赖与风险

### 9.1 依赖

| 依赖 | 说明 |
|------|------|
| `permission-service` 现有表 | sys_role, sys_permission, rel_role_permission, rel_user_role |
| `auth-service` JWT | 提供 user_id |
| `master-data-service` | 提供作用域实体数据（小区/楼栋/单元/网格） |
| Element Plus | el-tree 组件 |
| Pinia | 前端状态管理 |

### 9.2 风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 权限配置错误导致管理员无法登录 | 🔴 高 | owner 角色天然全权限，确保至少一个 owner 账号 |
| 权限树初始化数据不完整 | 🟡 中 | 提供 SQL 脚本 + 手动补录界面 |
| Redis 批量失效性能影响 | 🟡 中 | 依赖现有缓存机制（TTL 30min），后续优化为 SCAN |
| 前端路由守卫逻辑复杂 | 🟢 低 | 参考 vue-element-admin 成熟方案 |

---

## 十、相关资源

- 需求规格：`.harness/changes/rbac-management-ui/proposal.md`
- 详细规格：`.harness/changes/rbac-management-ui/specs/*.md`
- 现有设计：`services/permission-service/docs/design.md`
- Proto 规范：`.harness/rules/Proto管理规范.md`
- 前端现有角色页面：`web/pc/src/views/roles/List.vue`
- 前端现有权限树组件：`web/pc/src/components/business/PermissionTree.vue`
