# Spec: 角色管理 (Role Management)

## 一、能力概述

提供角色的增删改查功能，支持系统角色（4个核心角色）和自定义角色的管理。系统角色不可删除，自定义角色可自由创建删除。

---

## 二、功能需求

### 2.1 角色列表

**功能描述**：展示所有角色（系统角色 + 自定义角色），支持分页。

**界面元素**：
- 表格列：ID、角色名称、角色编码、描述、角色类型（系统/自定义）、状态（启用/禁用）、创建时间、操作
- 操作按钮：编辑、权限配置、删除（系统角色禁用删除按钮）
- 顶部按钮：新建角色（需要 `role:create` 权限）

**数据源**：`permission-service.ListRoles`

**前端路由**：`/roles`

**权限要求**：`role:read`（查看列表）

**交互行为**：
- 系统角色的"删除"按钮灰化且不可点击
- 点击"编辑"打开编辑对话框
- 点击"权限配置"跳转到权限配置页面（`/roles/:id/permissions`）
- 点击"删除"弹出二次确认对话框

**边界条件**：
- 空列表时显示 Empty 状态
- 删除失败时显示错误提示（如"角色已被分配，无法删除"）

---

### 2.2 创建角色

**功能描述**：创建自定义角色。

**表单字段**：
- 角色名称（必填，最多 50 字符）
- 角色编码（必填，英文字母、数字、下划线，最多 50 字符，唯一）
- 描述（可选，最多 200 字符）

**校验规则**：
- 角色编码正则：`/^[a-zA-Z0-9_]+$/`
- 角色编码唯一性：后端返回错误码 060006 时提示"角色编码已存在"

**数据提交**：`permission-service.CreateRole`

**权限要求**：`role:create`

**成功后行为**：关闭对话框，刷新列表，显示成功提示

---

### 2.3 编辑角色

**功能描述**：编辑自定义角色的基本信息（不含权限配置）。

**表单字段**：
- 角色名称（可修改）
- 角色编码（只读，不可修改）
- 描述（可修改）
- 状态（启用/禁用）

**数据提交**：`permission-service.UpdateRole`

**权限要求**：`role:update`

**交互行为**：
- 系统角色也可编辑基本信息（名称、描述），但不可修改编码和状态
- 编辑对话框标题显示"编辑角色 - {角色名称}"

---

### 2.4 删除角色

**功能描述**：软删除自定义角色。

**前置条件**：
- 角色不是系统角色（`is_system=0`）
- 角色未被分配给任何用户

**数据提交**：`permission-service.DeleteRole`

**权限要求**：`role:delete`

**交互行为**：
- 二次确认对话框："确定删除角色「{角色名称}」吗？此操作不可恢复。"
- 删除失败时显示具体错误（如"角色已被分配，无法删除"）

**错误处理**：
- 错误码 060004 → "角色已被分配，无法删除"
- 系统角色尝试删除 → 前端按钮禁用，不触发请求

---

### 2.5 角色详情（含权限列表）

**功能描述**：查询角色基本信息和已配置的权限ID列表。

**数据源**：`permission-service.GetRole`

**返回字段**：
- 角色基本信息：id, role_code, role_name, description, is_system, status
- 权限ID列表：`permission_ids: [1, 2, 5, 8]`

**使用场景**：
- 权限配置页面加载时，获取角色当前拥有的权限ID列表
- 编辑角色对话框回显数据

---

## 三、数据模型

### 3.1 前端 TypeScript 类型

```typescript
// src/types/permission.ts
export interface Role {
  id: string;                    // Snowflake ID（字符串）
  role_code: string;             // 角色编码（唯一）
  role_name: string;             // 角色名称
  description?: string;          // 描述
  is_system: number;             // 1=系统角色 0=自定义角色
  status: number;                // 1=启用 0=禁用
  sort_order: number;            // 排序
  created_at?: string;           // 创建时间
  updated_at?: string;           // 更新时间
}

export interface CreateRoleRequest {
  role_name: string;
  role_code: string;
  description?: string;
}

export interface UpdateRoleRequest {
  id: string;
  role_name: string;
  description?: string;
  status?: number;
}

export interface RoleDetailResponse {
  role: Role;
  permission_ids: string[];      // 该角色拥有的权限ID列表
}
```

### 3.2 后端 Proto 定义

```protobuf
// api-proto/api/permission/v1/permission.proto

message RoleInfo {
  int64 id = 1 [jstype = JS_STRING];
  string role_code = 2;
  string role_name = 3;
  string description = 4;
  int64 is_system = 5;
  int64 status = 6;
  int64 sort_order = 7;
  string created_at = 8;
  string updated_at = 9;
}

message CreateRoleReq {
  string role_name = 1;
  string role_code = 2;
  string description = 3;
}

message CreateRoleResp {
  common.v1.BaseResp base_resp = 1;
  int64 role_id = 2 [jstype = JS_STRING];
}

message UpdateRoleReq {
  int64 role_id = 1 [jstype = JS_STRING];
  string role_name = 2;
  string description = 3;
  int64 status = 4;
  repeated int64 permission_ids = 5 [jstype = JS_STRING];  // 权限ID列表（全量替换）
}

message DeleteRoleReq {
  int64 role_id = 1 [jstype = JS_STRING];
}

message GetRoleReq {
  int64 role_id = 1 [jstype = JS_STRING];
}

message GetRoleResp {
  common.v1.BaseResp base_resp = 1;
  RoleInfo role = 2;
  repeated int64 permission_ids = 3 [jstype = JS_STRING];
}

message ListRolesReq {
  int64 page = 1;
  int64 page_size = 2;
  int64 status = 3;              // 可选：按状态筛选（1=启用 0=禁用）
}

message ListRolesResp {
  common.v1.BaseResp base_resp = 1;
  repeated RoleInfo roles = 2;
  int64 total = 3;
}
```

### 3.3 数据库表（现有）

**表名**：`sys_role`

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT PK | Snowflake ID |
| `role_code` | VARCHAR(50) UNIQUE | 角色编码 |
| `role_name` | VARCHAR(50) | 角色名称 |
| `description` | VARCHAR(200) | 描述 |
| `is_system` | TINYINT | 1=系统角色 0=自定义角色 |
| `status` | TINYINT | 1=启用 0=禁用 |
| `sort_order` | INT | 排序 |
| `delete_time` | DATETIME NULL | 软删除时间 |
| `created_at` | DATETIME | 创建时间 |
| `updated_at` | DATETIME | 更新时间 |

**索引**：
- PRIMARY KEY (`id`)
- UNIQUE KEY `uk_role_code` (`role_code`)

---

## 四、接口清单

### 4.1 前端 API 模块

**文件**：`src/api/permission.ts`（新建）

```typescript
import request from '@/utils/request';
import type { Role, CreateRoleRequest, UpdateRoleRequest, RoleDetailResponse } from '@/types/permission';

// 角色列表
export function getRoles(params: { page: number; page_size: number; status?: number }) {
  return request.get<{ roles: Role[]; total: number }>('/api/v1/permission/roles', { params });
}

// 创建角色
export function createRole(data: CreateRoleRequest) {
  return request.post<{ role_id: string }>('/api/v1/permission/roles', data);
}

// 更新角色
export function updateRole(id: string, data: UpdateRoleRequest) {
  return request.put(`/api/v1/permission/roles/${id}`, data);
}

// 删除角色
export function deleteRole(id: string) {
  return request.delete(`/api/v1/permission/roles/${id}`);
}

// 角色详情（含权限列表）
export function getRoleDetail(id: string) {
  return request.get<RoleDetailResponse>(`/api/v1/permission/roles/${id}`);
}
```

### 4.2 后端 gRPC 接口

**服务**：`permission-service`

| RPC | 请求 | 响应 | 说明 |
|-----|------|------|------|
| `CreateRole` | `CreateRoleReq` | `CreateRoleResp` | 创建角色（`is_system=0`，`status=1`） |
| `UpdateRole` | `UpdateRoleReq` | `BaseResp` | 更新角色基本信息 + 权限列表（全量替换） |
| `DeleteRole` | `DeleteRoleReq` | `BaseResp` | 软删除（检查 `is_system` 和分配状态） |
| `GetRole` | `GetRoleReq` | `GetRoleResp` | 角色详情 + 权限ID列表 |
| `ListRoles` | `ListRolesReq` | `ListRolesResp` | 角色列表（分页） |

**已存在接口**（无需修改）：
- `ListRoles`：已有，支持分页

**需要新增接口**：
- `GetRole`：返回角色详情 + `permission_ids`

---

## 五、业务规则

### 5.1 系统角色保护

| 规则 | 实现位置 |
|------|---------|
| 系统角色不可删除 | 前端：删除按钮 `disabled`；后端：`DeleteRole` 检查 `is_system=1` 返回错误 |
| 系统角色可编辑名称和描述 | 编辑对话框允许修改 `role_name` 和 `description` |
| 系统角色编码不可修改 | 编辑对话框 `role_code` 字段设为只读 |

### 5.2 角色编码唯一性

- 创建角色时，后端检查 `role_code` 是否已存在（包括已软删除的记录）
- 违反唯一性约束时，返回错误码 060006
- 前端捕获错误码，显示提示："角色编码已存在，请更换"

### 5.3 角色删除前置条件

- 检查 `rel_user_role` 表，若存在 `role_id` 关联记录，拒绝删除
- 返回错误码 060004："角色已被分配，无法删除"
- 前端显示错误提示，建议用户先撤销所有用户的该角色

### 5.4 默认状态

- 创建角色时，默认 `status=1`（启用）
- 创建角色时，默认 `is_system=0`（自定义角色）
- 创建角色时，默认 `sort_order=0`（可后续调整）

---

## 六、界面交互

### 6.1 角色列表页面

**路径**：`/roles`

**布局**：
- 页面标题："角色管理"
- 右上角："新建角色"按钮（`v-permission="'role:create'"`）
- 表格：展示所有角色
- 分页器：底部居中

**表格列定义**：
- ID（宽度 80px）
- 角色名称
- 角色编码
- 描述（超长省略）
- 角色类型（Tag：系统=warning，自定义=info）
- 状态（Tag：启用=success，禁用=danger）
- 创建时间（宽度 180px）
- 操作（固定右侧，宽度 240px）

**操作列按钮**：
- 编辑（`v-permission="'role:update'"`）
- 权限配置（`v-permission="'role:permission'"`）
- 删除（`v-permission="'role:delete'"`，系统角色 `disabled`）

### 6.2 创建/编辑角色对话框

**对话框宽度**：600px

**表单布局**：
- Label 宽度：100px
- 字段：
  - 角色名称（Input，必填，最多 50 字符）
  - 角色编码（Input，必填，创建时可编辑，编辑时只读）
  - 描述（Textarea，3 行，最多 200 字符）
  - 状态（仅编辑时显示，Switch）

**底部按钮**：
- 取消（关闭对话框）
- 确定（提交表单，loading 状态）

**校验提示**：
- 角色名称为空："请输入角色名称"
- 角色编码为空："请输入角色编码"
- 角色编码格式错误："只能包含英文字母、数字、下划线"
- 角色编码已存在（后端返回）："角色编码已存在，请更换"

### 6.3 删除二次确认

**对话框类型**：MessageBox.confirm

**提示内容**：
```
确定删除角色「{role_name}」吗？

此操作不可恢复。
```

**按钮**：
- 取消（默认聚焦）
- 确定（红色危险按钮）

---

## 七、错误处理

### 7.1 错误码映射

| 后端错误码 | 前端提示 |
|-----------|---------|
| 060001 | 角色不存在 |
| 060004 | 角色已被分配，无法删除 |
| 060006 | 角色编码已存在 |
| 060003 | 无权限操作 |

### 7.2 通用错误处理

- 网络错误：显示"网络请求失败，请稍后重试"
- 超时错误：显示"请求超时，请检查网络连接"
- 权限错误（403）：跳转到 `/403` 页面

---

## 八、测试场景

### 8.1 单元测试

**前端**：
- 角色列表加载正确
- 创建角色表单校验生效
- 系统角色删除按钮禁用

**后端**：
- `CreateRole` 创建成功，返回 `role_id`
- `DeleteRole` 系统角色拒绝删除，返回错误码 060004
- `DeleteRole` 已分配角色拒绝删除，返回错误码 060004
- `GetRole` 返回角色详情 + 权限ID列表

### 8.2 集成测试

- E2E 流程：登录 → 进入角色管理 → 创建角色 → 编辑角色 → 删除角色
- 权限控制：无 `role:create` 权限时，"新建角色"按钮不显示
- 系统角色保护：尝试删除 `owner` 角色，按钮禁用

### 8.3 边界条件

- 角色编码包含特殊字符，校验拒绝
- 角色名称超过 50 字符，校验拒绝
- 删除已分配角色，显示错误提示
- 空列表时，显示 Empty 状态

---

## 九、依赖与约束

### 9.1 依赖

| 依赖 | 说明 |
|------|------|
| `permission-service` | 提供 gRPC 接口 |
| `auth-service` | 提供 JWT 校验（用户身份） |
| `v-permission` 指令 | 按钮级权限控制（需在 menu-permission-control 中实现） |

### 9.2 约束

- 角色编码一旦创建不可修改（防止权限配置混乱）
- 系统角色不可删除（保护核心业务实体）
- 角色删除为软删除（`delete_time` 不为 NULL），便于审计

---

## 十、追溯

### 10.1 需求来源

- 用户原话："在后台web pc 管理系统中，目前还没有完整的角色管理、权限管理等页面"
- Brainstorming 决策点 1：角色设计范围 → 选择混合模式（4个系统角色 + 自定义角色）

### 10.2 关联 Spec

- `permission-management/spec.md`：角色权限配置依赖本 Spec 的角色详情接口
- `user-role-assignment/spec.md`：用户角色分配依赖本 Spec 的角色列表接口
- `menu-permission-control/spec.md`：菜单权限控制依赖 `v-permission` 指令

### 10.3 记忆触发

- [[proto-jstype]]：所有 `int64` ID 字段必须添加 `[jstype = JS_STRING]`
- [[pre-commit-checks]]：提交前运行 `harness-checks.sh`
