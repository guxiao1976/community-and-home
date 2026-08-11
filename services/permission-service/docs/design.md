# Permission Service 设计方案

> **全局 RBAC 设计**：本文件是服务视角的实现细节。权限模型的完整方案（跨服务链路、前端对接、验收清单）见 [`docs/specs/rbac-design.md`](../../../docs/specs/rbac-design.md)。

## 一、定位

`permission-service` 是社区平台的 **RBAC 权限和数据范围引擎**。RPC 提供权限检查核心能力，API 网关（REST）提供管理后台操作入口。回答两个核心问题：

1. "这个用户能不能访问这个 API？"（`CheckPermission`）
2. "这个用户能看到哪些数据？"（`GetDataScopes`）

### 1.1 做什么

| 职责 | 说明 |
|------|------|
| 权限检查 | `CheckPermission` — 合并用户所有角色的权限集，匹配 `action:api_path`，Redis 缓存 30 分钟 |
| 数据范围 | `GetDataScopes` — 返回用户在指定 scope_type 下的所有 scope_id，供其他服务 GORM 拦截器做 WHERE IN 过滤 |
| 角色管理 | 创建/更新/删除/列表/详情，系统角色（is_system=1）不可删除，权限由配置决定 |
| 权限树管理 | `ListPermissions` — 按 parent_id 构建层级树，前端菜单渲染 |
| 角色分配/撤销 | `AssignRole` / `RevokeRole` — 用户-角色-数据范围三元组管理，Redis 缓存失效 |
| 用户角色/权限查询 | GetUserRoles, GetUserPermissions, GetRolesByIds |

### 1.2 不做什么

| 不负责 | 归属 |
|--------|------|
| 用户身份管理 | user-service |
| 登录/鉴权/Token | auth-service |
| API Gateway 路由配置 | API Gateway 配置文件 |
| 角色审批/认证 | user-service（user_membership_role.verf_status） |

### 1.3 核心设计决策

- **RBAC + 数据范围双层模型**：角色决定"能做什么"（权限码），scope_type+scope_id 决定"在哪做"（数据边界）
- **is_system 仅用于保护内置角色**：`is_system=1` 的角色不可删除、不可修改，权限由 `rel_role_permission` 配置决定，无特殊权限提升。如需全权限角色，通过管理后台创建自定义角色并勾选全部权限
- **权限树结构**：`sys_permission.parent_id` 自引用，前端按 sort_order 渲染层级菜单
- **Redis Set 缓存权限码**：key=`perm:user:{user_id}`，value=Set of `"METHOD:/api/path"`（直接存 `sys_permission.path`，不再拼 METHOD 前缀），TTL 30 分钟
- **角色变更精确失效缓存**：UpdateRole 查 FindByRoleId 拿到所有持有者，逐个 DEL `perm:user:{user_id}`（不用 KEYS 全扫）
- **用户禁用拦截**：CheckPermission 先查 `user:disabled:{user_id}` 标记，命中即拒绝；标记由 user-service 状态变更时写入
- **双入口**：权限检查/数据范围仅 gRPC 供其他服务调用；管理操作（角色/权限/用户角色）通过 REST API 网关，API 层挂 PermMiddleware 校验调用者权限

---

## 二、数据库设计（`permission` 库，4 张表）

### 2.1 `sys_role` — 角色定义表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int64 PK | 自增 |
| `role_code` | string UNIQUE | 全局唯一编码：owner, property_admin, community_admin, grid_worker |
| `role_name` | string | 显示名称 |
| `description` | string? | 描述 |
| `is_system` | int64 | 1=系统角色（不可删除，权限由配置决定，无自动继承） |
| `sort_order` | int64 | 排序 |
| `status` | int64 | 1=启用 0=禁用 |
| `deleted_at` | datetime? | 软删除时间 |

### 2.2 `sys_permission` — 权限定义表

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int64 PK | 自增 |
| `parent_id` | int64? | 父权限ID，NULL=根节点 |
| `name` | string | 权限名称 |
| `code` | string UNIQUE | 权限码，如 `user:read`, `user:write` |
| `type` | int64 | 1=菜单 2=按钮 3=API |
| `path` | string? | API 路径（type=3 时），格式 `METHOD:/api/path`，如 `GET:/api/users` |
| `sort_order` | int64 | 排序 |

权限不可删除（当前实现）。只读 `status=1` 的记录。

### 2.3 `rel_role_permission` — 角色-权限关联

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int64 PK | |
| `role_id` | int64 FK → sys_role | |
| `permission_id` | int64 FK → sys_permission | |

更新角色权限时：DELETE 旧关联 + BatchInsert 新关联（非事务性）。

### 2.4 `rel_user_role` — 用户-角色关联（含数据范围）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int64 PK | |
| `user_id` | int64 | |
| `role_id` | int64 FK → sys_role | |
| `scope_type` | string | community / building / unit / grid |
| `scope_id` | int64 | 数据范围实体 ID |

INSERT IGNORE 保证幂等。同一用户可拥有同一角色在不同 scope 的多条记录。

### ER 总览

```
sys_role (1) ──< (N) rel_role_permission (N) >── (1) sys_permission
sys_role (1) ──< (N) rel_user_role (N) >── (1) user (外部)
                                   └── scope_type, scope_id（数据边界）
```

---

## 三、核心业务流程

### 3.1 CheckPermission 权限检查

```
0. 查 user:disabled:{user_id} 标记（用户禁用拦截）
   HIT → 返回 allowed=false
1. Redis SISMEMBER perm:user:{user_id} "{path}"（path 含 METHOD 前缀，如 "GET:/api/users"）
2. HIT → 返回 allowed=true（<1ms）
3. MISS → DB 查询:
   a. FindActiveByUserId → 获取用户所有活跃角色
   b. 收集所有 role_id → 查 rel_role_permission → 收集 permission_id
   c. 查 sys_permission → SADD 到 Redis Set → EXPIRE 30min
   d. 匹配 needle → 返回结果
```

> 全局设计见 `docs/specs/rbac-design.md` §3.1（完整请求流）、§3.2（缓存策略）

### 3.2 GetDataScopes 数据范围

```
1. FindScopesByUserId(user_id, scope_type) → DISTINCT scope_id 列表
2. SADD 到 Redis perm:scopes:{user_id}:{scope_type}
3. EXPIRE 30min
4. 返回 scope_id 列表 → 调用方做 WHERE x IN (scope_ids)
```

### 3.3 角色权限更新 → 缓存失效

```
UpdateRole:
  1. 更新 sys_role 字段
  2. DELETE + BatchInsert rel_role_permission（替换权限列表）
  3. FindByRoleId(roleId) → 查所有持有该角色的 user_id
  4. 逐个 DEL perm:user:{user_id} + perm:scopes:{user_id}:*（精确失效，不用 KEYS 全扫）
```

---

## 四、gRPC 接口

Proto 定义：`api-proto/api/permission/v1/permission.proto`，共 14 个 RPC：

| RPC | 鉴权 | 超时 | 说明 |
|-----|------|------|------|
| `CheckPermission` | INTERNAL | 500ms | 权限检查（最核心，高频调用，含用户禁用拦截） |
| `GetDataScopes` | INTERNAL | 500ms | 获取数据范围 |
| `GetUserPermissions` | INTERNAL | 500ms | 获取用户所有权限码 |
| `GetRolesByIds` | INTERNAL | 1s | 批量角色查询 |
| `AssignRole` | JWT | 3s | 分配角色（幂等 INSERT IGNORE） |
| `RevokeRole` | JWT | 3s | 撤销角色 |
| `GetUserRoles` | JWT | 1s | 用户角色列表 |
| `ListRoles` | JWT | 1s | 角色列表（分页） |
| `ListPermissions` | JWT | 1s | 权限树 |
| `CreateRole` | JWT | 3s | 创建角色 |
| `UpdateRole` | JWT | 3s | 更新角色（含权限替换） |
| `DeleteRole` | JWT | 3s | 软删除（系统角色/已分配角色不可删） |
| `GetRole` | JWT | 1s | 角色详情（含权限列表） |
| `InvalidateUserCache` | INTERNAL | 500ms | 失效用户权限缓存（用户禁用/状态变更时调用） |

---

## 五、缓存策略

| Key Pattern | 类型 | TTL | 读 | 写 | 失效 |
|-------------|------|:---:|------|------|------|
| `perm:user:{user_id}` | Set | 30min | CheckPermission MISS 时 SADD | CheckPermission MISS 时 | AssignRole/RevokeRole DEL；UpdateRole 精确 DEL 所有持有者；InvalidateUserCache |
| `perm:scopes:{user_id}:{scope_type}` | Set | 30min | ⚠️ 写后即返回（不从 Redis 读） | GetDataScopes | AssignRole/InvalidateUserCache DEL |
| `user:disabled:{user_id}` | String | 24h | CheckPermission 每次先查 | user-service 禁用时写 | user-service 启用时删 |

### 已知问题

- `GetDataScopes` 缓存为 write-only（只写不读），目前没起到加速作用

---

## 六、配置

```yaml
# rpc/etc/permissionservice.yaml
Name: permission.rpc
ListenOn: 0.0.0.0:8084
Etcd:
  Hosts: [127.0.0.1:2379]
  Key: permission.rpc
DataSource: root:root123456@tcp(localhost:3306)/permission?charset=utf8mb4&parseTime=true&loc=Local
Cache:
  - Host: localhost:6379
```

---

## 七、依赖

| 依赖 | 说明 |
|------|------|
| `api-proto` (permission/v1, common/v1) | Proto 定义 |
| `community-common/v2` (responsex) | 统一响应格式 |
| MySQL | permission 库（4 张表） |
| Redis | 权限/范围缓存 |
| etcd | 服务注册发现 |

**无出站 gRPC 调用** — 纯叶子节点。

---

## 八、错误码

| 错误码 | 含义 |
|--------|------|
| 060001 | 角色不存在 |
| 060002 | 权限不存在 |
| 060003 | 访问被拒绝 |
| 060004 | 角色已被分配，不可删除 |
| 060005 | 不支持的数据范围类型 |
| 060006 | 角色编码已存在 |

---

## 九、消费者

所有服务的 gRPC 拦截器 / API Gateway 调用：
- `CheckPermission` — 每次 API 请求的鉴权入口
- `GetDataScopes` — GORM 拦截器的数据隔离
- `GetRolesByIds` — 获取角色元数据
- `GetUserPermissions` — 用户权限码全集
