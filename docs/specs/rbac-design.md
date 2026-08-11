# RBAC 权限管理设计方案

> 最后更新：2026-08-11

## 1. 概述

社区平台采用基于 RBAC（Role-Based Access Control）的权限体系。用户通过角色获得权限，权限精确控制 API 访问、菜单可见性和页面操作。

### 核心原则

- **权限由配置决定**：`is_system=1` 仅保护角色不被误删，不自动赋予全权限
- **Proto 是契约**：所有接口定义在 `api-proto/`，权限路径以 Proto HTTP 注解为准
- **前后端字段一致性**：全链路使用 snake_case，禁止前端自行转 camelCase

---

## 2. 数据模型

### 2.1 表结构

```
sys_role (角色)
├── id, role_code, role_name, description
├── is_system (1=系统角色, 不可删除)
├── status (1=启用), sort_order
└── created_at, updated_at, deleted_at

sys_permission (权限树)
├── id, parent_id (0=根节点)
├── name, code (全局唯一, 如 user:read)
├── type: 1=菜单 2=按钮 3=API
├── path (type=3 时存储 METHOD:/api/path, 如 GET:/api/users)
├── icon, sort_order, status
└── created_at, updated_at

rel_user_role (用户-角色)
├── user_id, role_id
├── scope_type, scope_id (数据范围, 预留)
└── created_at

rel_role_permission (角色-权限)
├── role_id, permission_id
└── created_at
```

### 2.2 系统角色

| role_code | 角色名 | 典型权限范围 |
|-----------|--------|-------------|
| `owner` | 业主 | 查看用户、查看社区（只读） |
| `property_admin` | 物业管理员 | 用户查看 + 社区查看 + 通知发布 + 审核查看 |
| `community_admin` | 社区管理员 | 用户/角色/社区/审核全面管理 |
| `grid_worker` | 网格员 | 用户查看 + 社区查看（最小权限） |

### 2.3 权限编码规范

```
{domain}:{action}          按钮层  user:read, role:create
{domain}:{action}:{scope}  API层   user:read:list-api, role:create:api
{domain}:menu              菜单层  user:menu, role:menu
```

### 2.4 权限路径格式

type=3（API 层）的 `path` 字段格式为 `{METHOD}:{URL}`，与 HTTP 请求的 Method + Path 对应：

```
GET:/api/users
POST:/api/users
GET:/api/users/:id
```

**:id 占位符匹配**：`PermMiddleware` 会将实际请求中的数字路径段自动替换为 `:id` 再进行匹配（`normalizePathParams`）。

---

## 3. 权限校验链路

### 3.1 完整请求流程

```
用户请求: GET /api/users/4542136688377323520
  │
  ▼
APISIX / Vite 代理 → Go REST API
  │
  ▼
rest.WithJwt (JWT 验证)
  │  从 Token 提取 user_id = 4544322234965024768
  │  注入 context: ctx.Value("user_id")
  │
  ▼
PermMiddleware (common/pkg/middleware/permission.go)
  │  normalizePathParams(): /api/users/4542136688377323520 → /api/users/:id
  │  调用 CheckPermission(userId=454..., Action="GET", ApiPath="/api/users/:id")
  │
  ▼
Permission RPC (checkpermissionlogic.go)
  │
  │  needle = "GET:/api/users/:id"
  │
  ├─ Step 1: 查 Redis (SIsMember)
  │    key = "perm:user:4544322234965024768"
  │    SIsMember(key, "GET:/api/users/:id")
  │    ├─ true  → ✅ Allowed=true (耗时 <1ms)
  │    └─ false → Step 2
  │
  └─ Step 2: 查 MySQL (3 次查询)
       ┌────────────────────────────────────────────────┐
       │ ① rel_user_role                                │
       │    WHERE user_id = 4544322234965024768         │
       │    → [role_id=3] (community_admin)              │
       │                                                │
       │ ② rel_role_permission (每个角色一次)            │
       │    WHERE role_id = 3                            │
       │    → [permission_id=100, 110, ..., 522]         │
       │                                                │
       │ ③ sys_permission                               │
       │    WHERE id IN (100, 110, ..., 522)             │
       │    → [{path:"GET:/api/users"}, ...]             │
       └────────────────────────────────────────────────┘
       │
       ▼
     回填 Redis
       SADD perm:user:454... "GET:/api/users"
       SADD perm:user:454... "GET:/api/users/:id"
       ... (该用户全部权限 path)
       EXPIRE 1800 秒
       │
       ▼
     匹配成功 → Allowed=true
     匹配失败 → Allowed=false → HTTP 200 + code=99401 "无权访问"
```

### 3.2 Redis 缓存策略 (Cache-Aside)

```
读: 先 Redis → 未命中 → MySQL → 回填 Redis → 返回
写: 角色/权限变更 → DEL perm:user:{userId} → 下次请求自动重建
```

- **存储结构**：Redis SET，key=`perm:user:{userId}`
- **存储内容**：直接使用 `sys_permission.path` 原始值（如 `GET:/api/users`），不拼接 Method 前缀
- **TTL**：默认 1800 秒，可通过 `permission.cache.ttl_seconds` 配置
- **失效触发**：`assignRole`/`revokeRole`/`updateRole` 时主动 `DEL`
- **用户禁用**：user-service 调 `InvalidateUserCache` RPC 删缓存 + 写 `user:disabled:{userId}` 标记（TTL 24h），CheckPermission 每次先查该标记，命中即拒绝

### 3.2.1 缓存失效/变更方案（2026-08-11 实施）

**策略：Cache-Aside + 删除失效**（低频写、高频读场景，删比更新简单且一致）

| 变更场景 | 操作 | 实现 |
|----------|------|------|
| 给用户加角色 | `DEL perm:user:{userId}` | assignrolelogic `invalidateCache` |
| 给用户删角色 | `DEL perm:user:{userId}` | revokerolelogic |
| 角色权限配置变更 | `DEL perm:user:{所有持有者}` | updaterolelogic `invalidateRoleCache` |
| 用户禁用 (status=2) | `DEL perm:user:{userId}` + 写 `user:disabled:{userId}` | user-service 调 `InvalidateUserCache` RPC + 写标记 |
| 用户启用 (status=1) | 删 `user:disabled:{userId}` | user-service |

**新增 RPC**：`InvalidateUserCache(userId)` → 删 `perm:user` + `perm:scopes:*` 系列
**禁用拦截**：`CheckPermission` 开头查 `user:disabled:{userId}`，命中即 `Allowed=false`
**验收**：禁用用户访问任何 API → 99401；启用后恢复 → code=0

### 3.3 路径参数归一化

`common/pkg/middleware/permission.go:normalizePathParams()`

**为什么需要**：真实请求路径含数字 ID，但 `sys_permission.path` 存的是模式 `:id`。直接比较永远不匹配。

```
请求路径:  /api/users/4542136688377323520
DB 存储:  GET:/api/users/:id

直接比较: 4542136688377323520 ≠ :id → ❌

归一化后: 正则 /\d+/ → :id
/api/users/4542136688377323520   →   /api/users/:id
/api/perm/roles/3/permissions    →   /api/perm/roles/:id/permissions
/api/users                        →   /api/users (无变化)

比较: GET:/api/users/:id == GET:/api/users/:id → ✅
```

### 3.4 PermMiddleware 挂载

| 服务 | 状态 |
|------|------|
| user-service API | ✅ 已挂载 |
| permission-service API | ✅ 已挂载 |
| 其他服务 API | 待评估 |

---

## 4. 前端权限

### 4.1 菜单权限

侧边栏菜单通过 `permission` 字段控制可见性，权限码对应后端 type=2 按钮权限：

```ts
{ path: '/users/list', title: '用户管理', permission: 'user:read' }
{ path: '/roles',      title: '角色管理', permission: 'role:read' }
```

无 `permission` 字段的菜单项始终可见。

### 4.2 按钮权限

使用 `v-permission` 指令控制按钮显隐：

```html
<el-button v-permission="'user:create'" @click="handleCreate">创建用户</el-button>
<el-button v-permission="'role:delete'" @click="handleDelete">删除</el-button>
```

### 4.3 路由守卫

`router/permission.ts` 通过路由 `meta.permissions` 控制页面访问：

```ts
meta: { permissions: ['user:read'] }
```

### 4.4 权限加载流程

```
登录成功 → authStore.login()
  └── permissionStore.loadUserPermissionsAndMenus(userId)
        └── GET /api/perm/users/:userId/permissions
              └── 返回 { permissionCodes: [...] }
                    └── 存入 userPermissions[]
                          ├── 菜单过滤：hasPermission(code)
                          └── 按钮显隐：v-permission 指令
```

---

## 5. 前端字段契约

### 5.1 API 字段命名

全链路使用 snake_case，与后端 `json` tag 一致：

| 层 | 字段名 | 禁止 |
|----|--------|------|
| Go json tag | `created_at` | — |
| TS interface | `created_at: string` | `createdAt` |
| Vue 模板 | `row.created_at` | `row.createdAt` |
| 请求参数 | `{ permissionIds: [...] }` | `{ permission_ids: [...] }` |

### 5.2 ID 列宽

Snowflake 19 位 ID，`el-table-column` 统一 `width="200"`。

### 5.3 时间格式

- 后端返回 Unix 时间戳（秒）
- 前端 `formatTime()`：`new Date(ts * 1000).toLocaleString('zh-CN')`

---

## 6. 功能清单与验收

### 6.1 用户管理

| 功能 | API | 权限码 | 验收 |
|------|-----|--------|------|
| 用户列表 | GET /api/users | user:read | ✅ |
| 搜索（手机号精确/昵称模糊） | GET /api/users?phone=/keyword= | user:read | ✅ |
| 创建用户 | POST /api/users | user:create | ✅ |
| 编辑用户 | PUT /api/users/:id | user:update | ✅ |
| 查看用户详情 | GET /api/users/:id | user:read | ✅ |
| 分配角色 | 按钮入口 | user:assign-role | ✅ |

### 6.2 角色管理

| 功能 | API | 权限码 | 验收 |
|------|-----|--------|------|
| 角色列表 | GET /api/perm/roles | role:read | ✅ |
| 创建角色 | POST /api/perm/roles | role:create | ✅ |
| 编辑角色 | PUT /api/perm/roles/:id | role:update | ✅ |
| 删除角色 | DELETE /api/perm/roles/:id | role:delete | ✅ |
| 查看角色用户 | GET /api/perm/roles/:id/users | — | ✅ |
| 权限配置（进入） | 按钮入口 | role:permission | ✅ |
| 权限配置（保存） | POST /api/perm/roles/:id/permissions | role:permission | ✅ |

### 6.3 权限资源

| 功能 | API | 权限码 | 验收 |
|------|-----|--------|------|
| 查看权限树（只读） | GET /api/perm/permissions | role:permission | ✅ |
| 自动发现 | POST /api/perm/permissions/auto-discover | — | ✅ |

### 6.4 角色验收矩阵

| 测试场景 | owner | property_admin | community_admin |
|----------|-------|---------------|-----------------|
| 查看用户列表 | ✅ | ✅ | ✅ |
| 创建用户 | ❌ | ❌ | ✅ |
| 查看角色列表 | ❌ | ❌ | ✅ |
| 创建/删除角色 | ❌ | ❌ | ✅ |
| 配置角色权限 | ❌ | ❌ | ✅ |
| 为用户分配角色 | ❌ | ❌ | ✅ |

---

## 6.5 角色验收结果（2026-08-11）

| 测试场景 | owner | property_admin | community_admin | grid_worker |
|----------|-------|---------------|-----------------|-------------|
| 查看用户列表 | ✅ | ✅ | ✅ | ✅ |
| 创建用户 | ✅ 拦截 | ✅ 拦截 | ✅ | ✅ 拦截 |
| 查看角色列表 | ✅ 拦截 | ✅ 拦截 | ✅ | ✅ 拦截 |
| 创建角色 | ✅ 拦截 | ✅ 拦截 | ✅ | ✅ 拦截 |
| 查看权限列表 | ✅ 拦截 | ✅ 拦截 | ✅ | ✅ 拦截 |
| 发布通知 | — | ✅ 权限通过* | ✅ | ✅ 拦截 |
| 分配角色 | — | — | ✅ | — |

\* property_admin 发布通知返回 99500 为业务错误（缺参数），权限校验已通过。

**发现的问题**：`community:read` 权限对应的 `GET /api/community/communities` 路由在 community-hub-service 中不存在（仅有 notices/contact/lostfound）。该权限暂时无对应 API。

---

## 7. 已知差距

| # | 问题 | 影响 | 优先级 |
|---|------|------|--------|
| 1 | permission-service API 无 PermMiddleware | 任何登录用户可操作角色/权限 | ✅ 已修复 |
| 2 | 前端 `getRoleById` 返回 `{role:{...}}` 与类型定义不一致 | 角色名显示为空 | ✅ 已修复 |
| 3 | permission-service 未加入 RouteRegistry | 自动发现只能扫描权限服务自身路由 | ✅ 已修复 |
| 4 | 用户角色分配缺少数据范围实现 | scope 用 hardcode 的 global/0 | P2（需社区 API） |
| 5 | `v-permission` 指令和路由守卫的权限码未完全对齐 | 部分按钮显隐不准确 | ✅ 已修复 |
| 6 | 自动发现新增权限自动归入 `parent_id=0` | 直到 specialParent 映射补充前是孤儿节点 | P2（已部分修复） |

---

## 8. 相关文件索引

| 类别 | 文件 |
|------|------|
| 编码规范 | `.harness/rules/项目编码规范.md` §5.1, §5.2 |
| Proto 定义 | `api-proto/api/permission/v1/permission.proto` |
| 权限中间件 | `common/pkg/middleware/permission.go` |
| 缓存逻辑 | `services/permission-service/rpc/.../checkpermissionlogic.go` |
| 初始化脚本 | `services/permission-service/scripts/init_permissions.sql` |
| 前端权限 Store | `web/pc/src/stores/permission.ts` |
| 前端菜单配置 | `web/pc/src/config/modules/user-permission.config.ts` |
| 字段对齐 QA | `.harness/skills/qa/scripts/check-api-field-align.sh` |
| 交付前验证 | `.claude/skills/verify-before-deliver/SKILL.md` |
