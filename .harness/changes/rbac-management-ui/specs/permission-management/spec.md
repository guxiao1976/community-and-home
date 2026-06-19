# Spec: 权限管理 (Permission Management)

## 一、能力概述

提供权限树的展示和角色-权限关系的配置功能。权限分为三层：菜单权限（控制侧边栏显示）、按钮权限（控制操作按钮显示）、API权限（控制接口调用）。

---

## 二、功能需求

### 2.1 权限树展示

**功能描述**：以树形结构展示所有权限，支持父子层级（菜单 → 按钮/API）。

**界面元素**：
- 树形控件：el-tree
- 节点显示：图标 + 权限名称 + 类型标签（菜单/按钮/API）
- 节点勾选：Checkbox（支持父子联动）

**数据源**：`permission-service.ListPermissions`

**树形结构**：
```
□ 用户管理（菜单）
  □ 查看用户列表（按钮）
    □ GET /api/v1/user/users（API）
  □ 创建用户（按钮）
    □ POST /api/v1/user/users（API）
  □ 编辑用户（按钮）
    □ PUT /api/v1/user/users/:id（API）
  □ 删除用户（按钮）
    □ DELETE /api/v1/user/users/:id（API）

□ 角色管理（菜单）
  □ 查看角色列表（按钮）
    □ GET /api/v1/permission/roles（API）
  □ 配置权限（按钮）
    □ PUT /api/v1/permission/roles/:id/permissions（API）
```

**交互行为**：
- 父节点勾选 → 自动勾选所有子节点
- 父节点取消 → 自动取消所有子节点
- 子节点全部勾选 → 父节点自动勾选
- 子节点部分勾选 → 父节点半选状态（indeterminate）

**性能优化**：
- 权限树数据缓存到 Vuex/Pinia（全局共享）
- 初次加载后，后续页面直接从 Store 读取

---

### 2.2 角色权限配置

**功能描述**：为指定角色配置权限（勾选权限树节点）。

**页面路由**：`/roles/:id/permissions`

**页面元素**：
- 页面标题："权限配置 - {角色名称}"
- 返回按钮：el-page-header
- 提示信息：Alert（说明权限变更立即生效）
- 权限树：PermissionTree 组件
- 底部按钮：取消、保存

**数据加载**：
1. 加载角色基本信息：`permission-service.GetRole` → 获取角色名称
2. 加载所有权限：`permission-service.ListPermissions` → 构建权限树
3. 加载角色当前权限：`GetRole.permission_ids` → 勾选对应节点

**数据提交**：
- 收集所有勾选节点的 ID（包括父节点和子节点）
- 调用 `permission-service.UpdateRole`，传递 `permission_ids` 数组（全量替换）

**权限要求**：`role:permission`

**成功后行为**：
- 显示成功提示："权限配置已保存"
- 返回角色列表页面

**缓存失效**：
- 后端 `UpdateRole` 完成后，自动失效所有相关用户的权限缓存（Redis `perm:user:*`）
- 前端无需额外处理，用户下次请求时自动重建缓存

---

### 2.3 权限树构建算法

**输入**：扁平的权限列表（`PermissionInfo[]`）

**输出**：树形结构（`PermissionTreeNode[]`）

**算法**：
```typescript
function buildPermissionTree(permissions: PermissionInfo[]): PermissionTreeNode[] {
  const tree: PermissionTreeNode[] = [];
  const map = new Map<string, PermissionTreeNode>();

  // 第一遍：构建 Map
  permissions.forEach(p => {
    map.set(p.id, { ...p, children: [] });
  });

  // 第二遍：建立父子关系
  permissions.forEach(p => {
    const node = map.get(p.id)!;
    if (p.parent_id && p.parent_id !== '0') {
      const parent = map.get(p.parent_id);
      if (parent) {
        parent.children!.push(node);
      }
    } else {
      tree.push(node);  // 根节点
    }
  });

  // 第三遍：按 sort_order 排序
  const sortTree = (nodes: PermissionTreeNode[]) => {
    nodes.sort((a, b) => a.sort_order - b.sort_order);
    nodes.forEach(n => {
      if (n.children && n.children.length > 0) {
        sortTree(n.children);
      }
    });
  };
  sortTree(tree);

  return tree;
}
```

---

## 三、数据模型

### 3.1 前端 TypeScript 类型

```typescript
// src/types/permission.ts

export interface Permission {
  id: string;                    // Snowflake ID（字符串）
  parent_id: string;             // 父权限ID（'0' 表示根节点）
  name: string;                  // 权限名称
  code: string;                  // 权限码（如 user:read）
  type: number;                  // 1=菜单 2=按钮 3=API
  path?: string;                 // API 路径（type=3 时有值）
  sort_order: number;            // 排序
  children?: Permission[];       // 子权限（树形结构）
}

export interface PermissionTreeNode extends Permission {
  children?: PermissionTreeNode[];
}

export interface UpdateRolePermissionsRequest {
  role_id: string;
  permission_ids: string[];      // 勾选的所有权限ID（全量替换）
}
```

### 3.2 后端 Proto 定义

```protobuf
// api-proto/api/permission/v1/permission.proto

message PermissionInfo {
  int64 id = 1 [jstype = JS_STRING];
  int64 parent_id = 2 [jstype = JS_STRING];  // 0 表示根节点
  string name = 3;
  string code = 4;
  int64 type = 5;                            // 1=菜单 2=按钮 3=API
  string path = 6;                           // API 路径（type=3 时）
  int64 sort_order = 7;
}

message ListPermissionsReq {
  // 无参数，返回所有启用的权限（status=1）
}

message ListPermissionsResp {
  common.v1.BaseResp base_resp = 1;
  repeated PermissionInfo permissions = 2;   // 扁平列表，前端自行构建树
}

message UpdateRoleReq {
  int64 role_id = 1 [jstype = JS_STRING];
  string role_name = 2;
  string description = 3;
  int64 status = 4;
  repeated int64 permission_ids = 5 [jstype = JS_STRING];  // 全量替换
}
```

### 3.3 数据库表（现有）

**表名**：`sys_permission`

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT PK | Snowflake ID |
| `parent_id` | BIGINT NULL | 父权限ID（NULL 或 0 表示根节点） |
| `name` | VARCHAR(50) | 权限名称 |
| `code` | VARCHAR(100) UNIQUE | 权限码（如 user:read） |
| `type` | TINYINT | 1=菜单 2=按钮 3=API |
| `path` | VARCHAR(200) | API 路径（type=3 时） |
| `sort_order` | INT | 排序 |
| `status` | TINYINT | 1=启用 0=禁用 |
| `created_at` | DATETIME | 创建时间 |
| `updated_at` | DATETIME | 更新时间 |

**表名**：`rel_role_permission`

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT PK | 自增ID |
| `role_id` | BIGINT FK | 角色ID |
| `permission_id` | BIGINT FK | 权限ID |

**索引**：
- PRIMARY KEY (`id`)
- UNIQUE KEY `uk_role_perm` (`role_id`, `permission_id`)
- KEY `idx_role_id` (`role_id`)
- KEY `idx_permission_id` (`permission_id`)

---

## 四、接口清单

### 4.1 前端 API 模块

**文件**：`src/api/permission.ts`

```typescript
import request from '@/utils/request';
import type { Permission, UpdateRolePermissionsRequest } from '@/types/permission';

// 获取权限树（扁平列表，前端构建树）
export function getPermissions() {
  return request.get<{ permissions: Permission[] }>('/api/v1/permission/permissions');
}

// 更新角色权限
export function updateRolePermissions(roleId: string, permissionIds: string[]) {
  return request.put(`/api/v1/permission/roles/${roleId}/permissions`, {
    permission_ids: permissionIds
  });
}

// 获取角色权限ID列表（复用 getRoleDetail）
export function getRolePermissions(roleId: string) {
  return request.get<{ permission_ids: string[] }>(`/api/v1/permission/roles/${roleId}/permissions`);
}
```

### 4.2 后端 gRPC 接口

**服务**：`permission-service`

| RPC | 请求 | 响应 | 说明 |
|-----|------|------|------|
| `ListPermissions` | `ListPermissionsReq` | `ListPermissionsResp` | 返回所有启用的权限（扁平列表） |
| `UpdateRole` | `UpdateRoleReq` | `BaseResp` | 更新角色权限（`permission_ids` 全量替换） |
| `GetRole` | `GetRoleReq` | `GetRoleResp` | 获取角色详情 + 权限ID列表 |

**已存在接口**（需要增强）：
- `ListPermissions`：需要新增，返回所有权限的扁平列表
- `UpdateRole`：已存在，但需确认支持 `permission_ids` 字段（全量替换）

**实现逻辑（UpdateRole）**：
```go
func (l *UpdateRoleLogic) UpdateRole(in *pb.UpdateRoleReq) (*pb.BaseResp, error) {
    // 1. 更新角色基本信息
    err := l.svcCtx.RoleModel.Update(ctx, &model.SysRole{
        Id:          in.RoleId,
        RoleName:    in.RoleName,
        Description: in.Description,
        Status:      in.Status,
    })
    
    // 2. 全量替换权限关系
    if len(in.PermissionIds) > 0 {
        // DELETE 旧关联
        err = l.svcCtx.RolePermissionModel.DeleteByRoleId(ctx, in.RoleId)
        
        // BatchInsert 新关联
        relations := make([]*model.RelRolePermission, len(in.PermissionIds))
        for i, permId := range in.PermissionIds {
            relations[i] = &model.RelRolePermission{
                RoleId:       in.RoleId,
                PermissionId: permId,
            }
        }
        err = l.svcCtx.RolePermissionModel.BatchInsert(ctx, relations)
    }
    
    // 3. 失效缓存：删除所有 perm:user:* 和 perm:scopes:*
    keys, _ := l.svcCtx.Redis.Keys(ctx, "perm:user:*").Result()
    if len(keys) > 0 {
        l.svcCtx.Redis.Del(ctx, keys...)
    }
    keys, _ = l.svcCtx.Redis.Keys(ctx, "perm:scopes:*").Result()
    if len(keys) > 0 {
        l.svcCtx.Redis.Del(ctx, keys...)
    }
    
    return &pb.BaseResp{Code: 0, Msg: "success"}, nil
}
```

---

## 五、业务规则

### 5.1 权限树节点类型

| 类型 | type 值 | 图标 | 说明 |
|------|:------:|------|------|
| 菜单 | 1 | 📁 | 控制侧边栏菜单显示 |
| 按钮 | 2 | 🔘 | 控制页面内按钮显示 |
| API | 3 | 🔌 | 控制接口调用权限 |

### 5.2 父子联动规则

- **勾选父节点**：自动勾选所有子节点（递归）
- **取消父节点**：自动取消所有子节点（递归）
- **子节点全勾选**：父节点自动勾选
- **子节点部分勾选**：父节点半选状态（indeterminate）

**实现**：使用 el-tree 的 `check-strictly=false`（默认行为）

### 5.3 权限全量替换

- 提交时收集所有勾选节点（包括父节点和子节点）
- 后端 DELETE 旧的 `rel_role_permission` 记录
- 后端 BatchInsert 新的 `rel_role_permission` 记录
- **非事务性**：先删后插，中间状态可能导致权限校验失败（极短暂）

### 5.4 缓存失效策略

- `UpdateRole` 完成后，后端调用 Redis `KEYS perm:user:*` 批量删除
- 用户下次请求时，`CheckPermission` 缓存未命中，重新从 DB 查询并写入 Redis
- **性能影响**：批量删除 + 重建缓存，P99 < 100ms（依赖现有 permission-service 缓存机制）

### 5.5 系统角色权限继承

- 系统角色（`is_system=1`）在 `CheckPermission` 时直接返回 `allowed=true`，不查询权限表
- 系统角色的权限配置仅用于展示，不实际生效（后端逻辑优先级高于权限表）

---

## 六、界面交互

### 6.1 权限配置页面

**路径**：`/roles/:id/permissions`

**布局**：
- 顶部：el-page-header（返回按钮 + 标题）
- 提示信息（Alert）：
  > 勾选权限后，拥有该角色的用户将获得对应的菜单和按钮访问权限。权限变更将立即生效。
- 权限树（el-card 包裹）
- 底部按钮：取消、保存

**权限树配置**：
- `node-key="id"`：唯一标识
- `default-expand-all`：默认展开所有节点
- `show-checkbox`：显示复选框
- `check-strictly=false`：父子联动
- `default-checked-keys`：初始勾选的权限ID列表

**节点模板**：
```vue
<template #default="{ node, data }">
  <span class="permission-node">
    <span class="permission-icon">
      <el-icon v-if="data.type === 1"><Folder /></el-icon>
      <el-icon v-else-if="data.type === 2"><Operation /></el-icon>
      <el-icon v-else><Link /></el-icon>
    </span>
    <span class="permission-name">{{ data.name }}</span>
    <el-tag v-if="data.type === 1" type="info" size="small">菜单</el-tag>
    <el-tag v-else-if="data.type === 2" type="warning" size="small">按钮</el-tag>
    <el-tag v-else type="success" size="small">API</el-tag>
    <span v-if="data.path" class="permission-path">{{ data.path }}</span>
  </span>
</template>
```

### 6.2 加载状态

- 页面加载时，显示 `v-loading` 骨架屏
- 数据加载失败时，显示 el-empty（"加载失败，请刷新重试"）

### 6.3 保存交互

- 点击"保存"按钮，按钮进入 loading 状态
- 提交成功：显示 ElMessage.success("权限配置已保存")，1秒后返回角色列表
- 提交失败：显示 ElMessage.error（错误信息），停留在当前页面

---

## 七、权限树数据初始化

### 7.1 初始化脚本

**文件**：`services/permission-service/scripts/init_permissions.sql`

**内容示例**：
```sql
-- 菜单权限（parent_id = 0）
INSERT INTO sys_permission (id, parent_id, name, code, type, path, sort_order, status) VALUES
(1, 0, '用户管理', 'user:menu', 1, NULL, 10, 1),
(2, 0, '角色管理', 'role:menu', 1, NULL, 20, 1),
(3, 0, '主数据管理', 'masterdata:menu', 1, NULL, 30, 1);

-- 按钮权限（parent_id = 菜单ID）
INSERT INTO sys_permission (id, parent_id, name, code, type, path, sort_order, status) VALUES
(11, 1, '查看用户列表', 'user:read', 2, NULL, 10, 1),
(12, 1, '创建用户', 'user:create', 2, NULL, 20, 1),
(13, 1, '编辑用户', 'user:update', 2, NULL, 30, 1),
(14, 1, '删除用户', 'user:delete', 2, NULL, 40, 1);

-- API 权限（parent_id = 按钮ID）
INSERT INTO sys_permission (id, parent_id, name, code, type, path, sort_order, status) VALUES
(111, 11, 'GET /api/v1/user/users', 'user:read:api', 3, '/api/v1/user/users', 10, 1),
(121, 12, 'POST /api/v1/user/users', 'user:create:api', 3, '/api/v1/user/users', 10, 1),
(131, 13, 'PUT /api/v1/user/users/:id', 'user:update:api', 3, '/api/v1/user/users/:id', 10, 1),
(141, 14, 'DELETE /api/v1/user/users/:id', 'user:delete:api', 3, '/api/v1/user/users/:id', 10, 1);

-- 系统角色自动关联所有权限
INSERT INTO rel_role_permission (role_id, permission_id)
SELECT r.id, p.id
FROM sys_role r, sys_permission p
WHERE r.is_system = 1;
```

### 7.2 权限码命名规范

**格式**：`{module}:{action}[:{type}]`

**示例**：
- `user:menu`（用户管理菜单）
- `user:read`（查看用户按钮）
- `user:read:api`（查看用户 API）
- `role:permission`（配置权限按钮）

**规则**：
- 菜单权限：`{module}:menu`
- 按钮权限：`{module}:{action}`
- API 权限：`{module}:{action}:api`

---

## 八、错误处理

### 8.1 错误码映射

| 后端错误码 | 前端提示 |
|-----------|---------|
| 060001 | 角色不存在 |
| 060002 | 权限不存在 |
| 060003 | 无权限操作 |

### 8.2 通用错误处理

- 网络错误：显示"网络请求失败，请稍后重试"
- 数据加载失败：显示 el-empty + "刷新"按钮
- 权限保存失败：显示错误提示，停留在当前页面

---

## 九、测试场景

### 9.1 单元测试

**前端**：
- 权限树构建算法正确（扁平 → 树形）
- 父子联动逻辑正确（勾选/取消）
- 权限ID收集正确（包括父节点和子节点）

**后端**：
- `ListPermissions` 返回所有启用权限（`status=1`）
- `UpdateRole` 全量替换权限关系成功
- `UpdateRole` 完成后缓存失效（`perm:user:*` 被删除）

### 9.2 集成测试

- E2E 流程：登录 → 角色列表 → 点击"权限配置" → 勾选权限 → 保存 → 返回列表
- 权限生效：保存后，新分配用户立即生效（缓存重建）
- 系统角色：配置系统角色权限，保存成功但不实际限制权限（`is_system=1` 优先级高）

### 9.3 性能测试

- 加载 1000+ 权限节点的树，渲染时间 < 500ms
- 勾选 500 个权限节点，提交时间 < 1s
- 权限变更后，用户下次请求权限检查 P99 < 100ms

---

## 十、依赖与约束

### 10.1 依赖

| 依赖 | 说明 |
|------|------|
| `permission-service` | 提供 gRPC 接口 |
| `role-management` | 依赖角色详情接口（`GetRole`） |
| Element Plus | el-tree 组件 |

### 10.2 约束

- 权限树深度最多 3 层（菜单 → 按钮 → API）
- 权限节点总数建议 < 2000（性能考虑）
- 权限变更非事务性（先删后插，极短暂中间状态）

---

## 十一、追溯

### 11.1 需求来源

- 用户原话："权限管理要实现接口级的管理，不能仅仅是前端是否展示相关按钮或者内容"
- Brainstorming 决策点 2：权限粒度设计 → 选择菜单+功能点两层结构

### 11.2 关联 Spec

- `role-management/spec.md`：依赖角色列表和角色详情接口
- `menu-permission-control/spec.md`：菜单权限控制依赖权限树数据
- `api-permission-control/spec.md`：API 权限控制依赖权限表的 `path` 字段

### 11.3 记忆触发

- [[proto-jstype]]：所有 `int64` ID 字段必须添加 `[jstype = JS_STRING]`
- [[grpc-only-comms]]：服务间通信仅通过 gRPC
