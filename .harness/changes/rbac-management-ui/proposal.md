# Proposal: RBAC 管理界面

## 一、变更概述

### 1.1 背景

当前系统已有 `permission-service` 提供 RBAC 核心能力（角色管理、权限检查、数据范围），但管理后台缺少完整的角色权限管理界面。现有问题：

- 角色管理页面（`web/pc/src/views/roles/List.vue`）仅支持基础 CRUD，缺少角色类型（系统/自定义）区分
- 权限配置页面（`Permissions.vue`）存在但功能不完整，未区分菜单权限和功能点权限
- **缺失用户角色分配界面**，无法为用户分配角色并指定数据作用域
- **缺失菜单权限控制**，前端路由未根据用户权限动态过滤
- **缺失接口级权限校验**，后端 API 无统一权限拦截机制

### 1.2 目标

构建完整的 RBAC 管理界面和权限控制体系，实现：

1. **角色管理**：支持系统角色（4个核心角色不可删除）+ 自定义角色创建
2. **权限管理**：菜单+功能点两层权限树，角色-权限关系配置
3. **用户角色分配**：为用户分配角色时指定数据作用域（如"阳光小区"）
4. **菜单权限控制**：前端路由根据用户权限动态过滤，无权限的菜单不显示
5. **接口权限控制**：后端 API 统一拦截校验，防止前端绕过

### 1.3 范围

| 范围 | IN | OUT |
|------|:--:|:---:|
| 前端角色管理界面（CRUD+系统角色保护） | ✅ | |
| 前端权限树管理界面（菜单+功能点两层） | ✅ | |
| 前端用户角色分配界面（含数据作用域） | ✅ | |
| 前端路由动态过滤（菜单权限控制） | ✅ | |
| 前端按钮级权限指令 `v-permission` | ✅ | |
| 后端 REST API 权限中间件 | ✅ | |
| APISIX Gateway 统一权限拦截 | | ⚠️ 配置层面，超出本次范围 |
| 内部 gRPC 权限校验 | | ❌ 信任域内调用，不重复校验 |
| 权限数据初始化脚本 | ✅ | |
| 历史用户权限迁移 | | ⚠️ 视现有数据情况决定 |

### 1.4 非功能性需求

| 维度 | 要求 |
|------|------|
| **性能** | 权限检查 P99 < 100ms（利用 permission-service 现有 Redis 缓存） |
| **安全** | 系统角色不可删除、权限校验无法前端绕过、敏感操作审计日志 |
| **可用性** | 权限配置错误不导致系统锁死（owner 角色天然全权限） |
| **兼容性** | 不破坏现有认证体系（auth-service AT+RT 机制） |

---

## 二、核心决策

### 2.1 角色设计范围（决策点 1）

**决策**：混合模式（4个系统角色 + 自定义角色）

| 角色类型 | 角色编码 | 角色名称 | 特性 |
|---------|---------|---------|------|
| **系统角色** | `owner` | 业主 | 不可删除，`is_system=1`，天然全权限 |
| **系统角色** | `property_admin` | 物业管理员 | 不可删除，`is_system=1` |
| **系统角色** | `community_admin` | 社区管理员 | 不可删除，`is_system=1` |
| **系统角色** | `grid_worker` | 网格员 | 不可删除，`is_system=1` |
| **自定义角色** | 用户定义 | 用户定义 | 可创建/删除，`is_system=0`，需配置权限 |

**理由**：
- 系统角色对应现有业务实体（user-service 的 `user_membership_role` 表已存在）
- 自定义角色支持灵活扩展（如"保安"、"访客接待员"）
- `is_system` 标志位防止误删核心角色

### 2.2 权限粒度设计（决策点 2）

**决策**：菜单+功能点两层结构

```
菜单层（控制侧边栏显示）
  ├─ 用户管理（菜单）
  │   ├─ 查看用户列表（功能点：按钮+API）
  │   ├─ 创建用户（功能点：按钮+API）
  │   └─ 编辑用户（功能点：按钮+API）
  └─ 角色管理（菜单）
      ├─ 查看角色列表（功能点：按钮+API）
      └─ 配置权限（功能点：按钮+API）
```

**权限表结构**（`sys_permission`）：
- `type`: 1=菜单, 2=按钮, 3=API
- `parent_id`: 父权限ID（菜单的 parent_id 为 NULL 或其他菜单，按钮/API 的 parent_id 为菜单）
- `code`: 权限码（如 `user:read`, `user:write`）
- `path`: API 路径（仅 type=3 时有值）

**理由**：
- 菜单层简化侧边栏渲染逻辑
- 功能点层精细控制按钮显示和 API 调用
- 三层扁平化避免过度嵌套

### 2.3 数据权限作用域（决策点 3）

**决策**：分配角色时指定作用域

在 `rel_user_role` 表中记录：
- `user_id`: 用户ID
- `role_id`: 角色ID
- `scope_type`: 作用域类型（community / building / unit / grid）
- `scope_id`: 作用域实体ID（如小区ID）

**示例**：
- 用户 A 在"阳光小区"（ID=101）是 `property_admin` → `{user_id: A, role_id: 2, scope_type: community, scope_id: 101}`
- 用户 A 在"和谐花园"（ID=102）也是 `property_admin` → `{user_id: A, role_id: 2, scope_type: community, scope_id: 102}`

**数据隔离实现**：
- 前端：请求列表时自动注入 `community_id` 过滤
- 后端：调用 `permission-service.GetDataScopes(user_id, scope_type='community')` 获取 `[101, 102]`，SQL 加 `WHERE community_id IN (101, 102)`

### 2.4 接口级权限控制（决策点 4）

**决策**：外部请求统一拦截 + 内部 REST 二次校验

| 层 | 机制 | 职责 |
|----|------|------|
| **APISIX Gateway** | 统一拦截外部请求 | JWT 校验 + 调用 `permission-service.CheckPermission` |
| **服务 REST 中间件** | 二次校验 | 防止内部直连绕过 Gateway（开发环境/内网调试） |
| **内部 gRPC** | 不校验 | 信任域内调用，避免性能损耗 |

**REST 中间件伪代码**：
```go
func PermissionMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("user_id")
        apiPath := r.URL.Path
        resp, err := permissionClient.CheckPermission(ctx, &pb.CheckPermReq{
            UserId: userID,
            Action: r.Method,
            Resource: apiPath,
        })
        if !resp.Allowed {
            http.Error(w, "403 Forbidden", 403)
            return
        }
        next(w, r)
    }
}
```

---

## 三、架构影响

### 3.1 涉及服务

| 服务 | 变更类型 | 说明 |
|------|---------|------|
| **permission-service** | 轻微增强 | 新增接口：`GetRole`（角色详情含权限列表）、`ListPermissions`（权限树） |
| **web/pc** | 重度变更 | 新增3个页面、权限指令、路由守卫、API 模块 |
| 其他服务（user/auth/moderation） | 中间件集成 | 引入权限校验中间件（可选，阶段性推进） |

### 3.2 Proto 变更

新增 Proto 消息（`api-proto/api/permission/v1/permission.proto`）：

```protobuf
// 角色详情（含权限列表）
message GetRoleReq {
  int64 role_id = 1 [jstype = JS_STRING];
}
message GetRoleResp {
  RoleInfo role = 1;
  repeated int64 permission_ids = 2 [jstype = JS_STRING];  // 该角色拥有的权限ID列表
}

// 权限树
message ListPermissionsReq {}
message ListPermissionsResp {
  repeated PermissionInfo permissions = 1;
}
message PermissionInfo {
  int64 id = 1 [jstype = JS_STRING];
  int64 parent_id = 2 [jstype = JS_STRING];
  string name = 3;
  string code = 4;
  int64 type = 5;         // 1=菜单 2=按钮 3=API
  string path = 6;        // API路径（type=3时）
  int64 sort_order = 7;
}

// 用户角色分配（含作用域）
message AssignRoleWithScopeReq {
  int64 user_id = 1 [jstype = JS_STRING];
  int64 role_id = 2 [jstype = JS_STRING];
  string scope_type = 3;  // community / building / unit / grid
  int64 scope_id = 4 [jstype = JS_STRING];
}
```

### 3.3 数据库变更

**无表结构变更**。现有 4 张表已满足需求：
- `sys_role`（角色定义，已有 `is_system` 字段）
- `sys_permission`（权限定义，已有 `parent_id`、`type`、`path`）
- `rel_role_permission`（角色-权限关联）
- `rel_user_role`（用户-角色关联，已有 `scope_type`、`scope_id`）

**需要数据初始化**：
- 初始化 4 个系统角色（`is_system=1`）
- 初始化权限树（菜单+功能点，覆盖现有所有 REST API）
- 系统角色自动关联全部权限

---

## 四、实施策略

### 4.1 分阶段交付

| 阶段 | 交付物 | 验收标准 |
|------|--------|---------|
| **P0** | Proto 变更 + permission-service 接口增强 | `make ci` 通过，新接口可调用 |
| **P1** | 前端角色管理界面（系统角色保护） | 可创建自定义角色，系统角色不可删除 |
| **P2** | 前端权限树管理 + 角色权限配置 | 权限树正确渲染，角色权限保存生效 |
| **P3** | 前端用户角色分配界面（含数据作用域） | 可为用户分配角色并指定小区 |
| **P4** | 前端菜单权限控制（路由守卫） | 无权限菜单不显示，访问受限路由跳转403 |
| **P5** | 后端权限中间件（示例服务） | user-service REST API 权限校验生效 |

### 4.2 回滚策略

- 前端：Feature Flag 控制新界面入口，问题回滚至旧版
- 后端：中间件可配置开关（`EnablePermissionCheck: false` 时跳过）
- Proto：仅新增消息，不破坏现有接口

### 4.3 测试策略

| 类型 | 覆盖 |
|------|------|
| **单元测试** | permission-service 新增 Logic 层 |
| **集成测试** | 前端 E2E：登录 → 角色管理 → 权限配置 → 用户分配 → 菜单过滤 |
| **性能测试** | 权限检查 P99 < 100ms（100 并发） |
| **安全测试** | 尝试绕过前端访问无权限 API，验证后端拦截 |

---

## 五、风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 权限配置错误导致管理员无法登录 | 🔴 高 | owner 角色天然全权限，确保至少一个 owner 账号 |
| 权限树初始化数据不完整 | 🟡 中 | 提供 SQL 脚本 + 手动补录界面 |
| 历史用户无角色数据 | 🟡 中 | 数据迁移脚本：user_membership_role → rel_user_role |
| 性能回退（每次请求查权限） | 🟡 中 | 依赖 permission-service 现有 Redis 缓存（30min TTL） |
| 前端路由守卫逻辑复杂 | 🟢 低 | 参考 vue-element-admin 成熟方案 |

---

## 六、相关资源

- 现有设计：`services/permission-service/docs/design.md`
- 业务流程：`.harness/knowledge/business-flows.md` § RBAC 权限检查
- Proto 规范：`.harness/rules/Proto管理规范.md`
- 前端现有角色页面：`web/pc/src/views/roles/List.vue`

---

## 七、术语表

| 术语 | 定义 |
|------|------|
| **系统角色** | `is_system=1` 的角色，不可删除，对应业务实体 |
| **自定义角色** | `is_system=0` 的角色，用户可创建/删除 |
| **菜单权限** | `type=1` 的权限，控制侧边栏菜单显示 |
| **功能点权限** | `type=2/3` 的权限，控制按钮显示和 API 调用 |
| **数据作用域** | `scope_type` + `scope_id`，用户在哪个范围拥有该角色 |
| **权限码** | `code` 字段，如 `user:read`，用于前端 `v-permission` 指令 |
