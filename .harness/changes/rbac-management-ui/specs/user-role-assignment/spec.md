# Spec: 用户角色分配 (User Role Assignment)

## 一、能力概述

为用户分配角色时指定数据作用域（如"阳光小区"），支持一个用户在多个作用域拥有同一角色。提供用户-角色关联的增删查功能。

---

## 二、功能需求

### 2.1 用户角色列表

**功能描述**：查看指定用户拥有的所有角色（含数据作用域）。

**界面位置**：
- 独立页面：`/users/:userId/roles`
- 或嵌入用户详情页面作为 Tab

**界面元素**：
- 表格列：角色名称、角色编码、作用域类型、作用域名称、分配时间、操作
- 顶部按钮：分配角色（`v-permission="'user:assign-role'"`）

**数据源**：`permission-service.GetUserRoles(user_id)`

**返回示例**：
```json
{
  "roles": [
    {
      "id": "1234567890",
      "role_code": "property_admin",
      "role_name": "物业管理员",
      "scope_type": "community",
      "scope_id": "101",
      "scope_name": "阳光小区",
      "assigned_at": "2026-06-15 10:30:00"
    },
    {
      "id": "1234567891",
      "role_code": "property_admin",
      "role_name": "物业管理员",
      "scope_type": "community",
      "scope_id": "102",
      "scope_name": "和谐花园",
      "assigned_at": "2026-06-16 14:20:00"
    }
  ]
}
```

**交互行为**：
- 点击"分配角色"打开分配对话框
- 点击"撤销"弹出二次确认对话框

---

### 2.2 分配角色

**功能描述**：为用户分配角色并指定数据作用域。

**表单字段**：
- 用户信息（只读显示）：用户ID、用户名、手机号
- 角色选择（必填）：下拉框，选项来自 `permission-service.ListRoles`
- 作用域类型（必填）：下拉框，选项：community（小区）、building（楼栋）、unit（单元）、grid（网格）
- 作用域实体（必填）：级联选择器，根据作用域类型动态加载
  - `community` → 从 `master-data-service` 获取小区列表
  - `building` → 先选小区，再选楼栋
  - `unit` → 先选小区，再选楼栋，再选单元
  - `grid` → 从 `master-data-service` 获取网格列表

**数据提交**：`permission-service.AssignRole`

**请求示例**：
```json
{
  "user_id": "1234567890",
  "role_id": "2",
  "scope_type": "community",
  "scope_id": "101"
}
```

**权限要求**：`user:assign-role`

**成功后行为**：
- 关闭对话框
- 刷新用户角色列表
- 显示成功提示："角色分配成功"

**幂等性**：
- 后端使用 `INSERT IGNORE` 保证幂等
- 重复分配相同角色+作用域，不返回错误

---

### 2.3 撤销角色

**功能描述**：撤销用户在指定作用域的角色。

**前置条件**：
- 用户在该作用域拥有该角色
- 撤销后用户至少保留一个角色（防止锁死）

**数据提交**：`permission-service.RevokeRole`

**请求示例**：
```json
{
  "user_id": "1234567890",
  "role_id": "2",
  "scope_type": "community",
  "scope_id": "101"
}
```

**权限要求**：`user:revoke-role`

**交互行为**：
- 二次确认对话框：
  > 确定撤销用户「{userName}」在「{scopeName}」的「{roleName}」角色吗？
- 确认后撤销

**错误处理**：
- 撤销最后一个角色：前端提示"用户至少需要保留一个角色"
- 角色不存在：显示"该用户未拥有此角色"

---

### 2.4 批量分配角色

**功能描述**：为多个用户批量分配同一角色+作用域。

**界面位置**：用户列表页面（`/users`）

**操作流程**：
1. 勾选多个用户（表格多选）
2. 点击"批量分配角色"按钮
3. 弹出对话框，选择角色+作用域
4. 提交，后端循环调用 `AssignRole`

**表单字段**：
- 已选用户（只读显示）：用户列表（最多显示 10 个，超出显示"等 N 人"）
- 角色选择（必填）
- 作用域类型（必填）
- 作用域实体（必填）

**数据提交**：
```typescript
// 前端循环调用
for (const userId of selectedUserIds) {
  await assignRole({ user_id: userId, role_id, scope_type, scope_id });
}
```

**权限要求**：`user:assign-role`

**成功后行为**：
- 显示成功数量："成功为 {N} 个用户分配角色"
- 失败时显示部分成功："成功 {M}/{N}，失败用户：{userNames}"

---

## 三、数据模型

### 3.1 前端 TypeScript 类型

```typescript
// src/types/permission.ts

export interface UserRole {
  id: string;                    // rel_user_role 表的 ID
  user_id: string;               // 用户ID
  role_id: string;               // 角色ID
  role_code: string;             // 角色编码
  role_name: string;             // 角色名称
  scope_type: string;            // community / building / unit / grid
  scope_id: string;              // 作用域实体ID
  scope_name: string;            // 作用域名称（前端关联查询）
  assigned_at: string;           // 分配时间
}

export interface AssignRoleRequest {
  user_id: string;
  role_id: string;
  scope_type: string;
  scope_id: string;
}

export interface RevokeRoleRequest {
  user_id: string;
  role_id: string;
  scope_type: string;
  scope_id: string;
}

export interface ScopeOption {
  value: string;                 // 作用域ID
  label: string;                 // 作用域名称
  children?: ScopeOption[];      // 级联子选项
}
```

### 3.2 后端 Proto 定义

```protobuf
// api-proto/api/permission/v1/permission.proto

message UserRoleInfo {
  int64 id = 1 [jstype = JS_STRING];           // rel_user_role.id
  int64 user_id = 2 [jstype = JS_STRING];
  int64 role_id = 3 [jstype = JS_STRING];
  string role_code = 4;
  string role_name = 5;
  string scope_type = 6;                       // community / building / unit / grid
  int64 scope_id = 7 [jstype = JS_STRING];
  string assigned_at = 8;
}

message GetUserRolesReq {
  int64 user_id = 1 [jstype = JS_STRING];
}

message GetUserRolesResp {
  common.v1.BaseResp base_resp = 1;
  repeated UserRoleInfo roles = 2;
}

message AssignRoleReq {
  int64 user_id = 1 [jstype = JS_STRING];
  int64 role_id = 2 [jstype = JS_STRING];
  string scope_type = 3;                       // community / building / unit / grid
  int64 scope_id = 4 [jstype = JS_STRING];
}

message RevokeRoleReq {
  int64 user_id = 1 [jstype = JS_STRING];
  int64 role_id = 2 [jstype = JS_STRING];
  string scope_type = 3;
  int64 scope_id = 4 [jstype = JS_STRING];
}
```

### 3.3 数据库表（现有）

**表名**：`rel_user_role`

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT PK | Snowflake ID |
| `user_id` | BIGINT | 用户ID |
| `role_id` | BIGINT FK | 角色ID（→ sys_role） |
| `scope_type` | VARCHAR(20) | community / building / unit / grid |
| `scope_id` | BIGINT | 作用域实体ID |
| `created_at` | DATETIME | 分配时间 |
| `updated_at` | DATETIME | 更新时间 |

**索引**：
- PRIMARY KEY (`id`)
- UNIQUE KEY `uk_user_role_scope` (`user_id`, `role_id`, `scope_type`, `scope_id`)
- KEY `idx_user_id` (`user_id`)
- KEY `idx_role_id` (`role_id`)

---

## 四、接口清单

### 4.1 前端 API 模块

**文件**：`src/api/permission.ts`

```typescript
import request from '@/utils/request';
import type { UserRole, AssignRoleRequest, RevokeRoleRequest } from '@/types/permission';

// 获取用户角色列表
export function getUserRoles(userId: string) {
  return request.get<{ roles: UserRole[] }>(`/api/v1/permission/users/${userId}/roles`);
}

// 分配角色
export function assignRole(data: AssignRoleRequest) {
  return request.post('/api/v1/permission/users/roles', data);
}

// 撤销角色
export function revokeRole(data: RevokeRoleRequest) {
  return request.delete('/api/v1/permission/users/roles', { data });
}
```

**文件**：`src/api/masterdata.ts`（现有，增强）

```typescript
// 获取小区列表（用于作用域选择）
export function getCommunities() {
  return request.get<{ communities: Community[] }>('/api/v1/masterdata/communities');
}

// 获取楼栋列表（按小区筛选）
export function getBuildings(communityId: string) {
  return request.get<{ buildings: Building[] }>(`/api/v1/masterdata/communities/${communityId}/buildings`);
}

// 获取单元列表（按楼栋筛选）
export function getUnits(buildingId: string) {
  return request.get<{ units: Unit[] }>(`/api/v1/masterdata/buildings/${buildingId}/units`);
}

// 获取网格列表
export function getGrids() {
  return request.get<{ grids: Grid[] }>('/api/v1/masterdata/grids');
}
```

### 4.2 后端 gRPC 接口

**服务**：`permission-service`

| RPC | 请求 | 响应 | 说明 |
|-----|------|------|------|
| `GetUserRoles` | `GetUserRolesReq` | `GetUserRolesResp` | 已存在，返回用户角色列表 |
| `AssignRole` | `AssignRoleReq` | `BaseResp` | 已存在，`INSERT IGNORE` 幂等 |
| `RevokeRole` | `RevokeRoleReq` | `BaseResp` | 已存在，`DELETE WHERE` 精确删除 |

**实现逻辑（AssignRole）**：
```go
func (l *AssignRoleLogic) AssignRole(in *pb.AssignRoleReq) (*pb.BaseResp, error) {
    // 1. 校验角色存在
    role, err := l.svcCtx.RoleModel.FindOne(ctx, in.RoleId)
    if err != nil {
        return nil, errorx.NewCodeError(060001, "角色不存在")
    }
    
    // 2. 校验作用域类型
    validScopes := []string{"community", "building", "unit", "grid"}
    if !contains(validScopes, in.ScopeType) {
        return nil, errorx.NewCodeError(060005, "不支持的数据范围类型")
    }
    
    // 3. INSERT IGNORE（幂等）
    err = l.svcCtx.UserRoleModel.InsertIgnore(ctx, &model.RelUserRole{
        UserId:    in.UserId,
        RoleId:    in.RoleId,
        ScopeType: in.ScopeType,
        ScopeId:   in.ScopeId,
    })
    
    // 4. 失效缓存
    l.svcCtx.Redis.Del(ctx, fmt.Sprintf("perm:user:%d", in.UserId))
    l.svcCtx.Redis.Del(ctx, fmt.Sprintf("perm:scopes:%d:%s", in.UserId, in.ScopeType))
    
    return &pb.BaseResp{Code: 0, Msg: "success"}, nil
}
```

**实现逻辑（RevokeRole）**：
```go
func (l *RevokeRoleLogic) RevokeRole(in *pb.RevokeRoleReq) (*pb.BaseResp, error) {
    // 1. 检查用户是否只剩一个角色
    count, err := l.svcCtx.UserRoleModel.CountByUserId(ctx, in.UserId)
    if err != nil || count <= 1 {
        return nil, errorx.NewCodeError(060007, "用户至少需要保留一个角色")
    }
    
    // 2. DELETE WHERE 精确删除
    err = l.svcCtx.UserRoleModel.DeleteByCondition(ctx, &model.RelUserRole{
        UserId:    in.UserId,
        RoleId:    in.RoleId,
        ScopeType: in.ScopeType,
        ScopeId:   in.ScopeId,
    })
    if err != nil {
        return nil, err
    }
    
    // 3. 失效缓存
    l.svcCtx.Redis.Del(ctx, fmt.Sprintf("perm:user:%d", in.UserId))
    l.svcCtx.Redis.Del(ctx, fmt.Sprintf("perm:scopes:%d:%s", in.UserId, in.ScopeType))
    
    return &pb.BaseResp{Code: 0, Msg: "success"}, nil
}
```

---

## 五、业务规则

### 5.1 作用域类型约束

| scope_type | 含义 | 实体来源 |
|-----------|------|---------|
| `community` | 小区 | master-data-service (residential_area) |
| `building` | 楼栋 | master-data-service (building) |
| `unit` | 单元 | master-data-service (unit) |
| `grid` | 网格 | master-data-service (grid) |

**校验规则**：
- `scope_type` 必须是上述 4 种之一
- `scope_id` 必须在对应实体表中存在（可选校验，避免跨服务依赖）

### 5.2 多作用域角色

**场景**：用户 A 是"阳光小区"的物业管理员，同时也是"和谐花园"的物业管理员。

**数据表示**：
```sql
INSERT INTO rel_user_role (user_id, role_id, scope_type, scope_id) VALUES
(123, 2, 'community', 101),  -- 阳光小区
(123, 2, 'community', 102);  -- 和谐花园
```

**权限检查**：
- `CheckPermission(user_id=123)` → 合并所有角色的权限码，返回 `allowed=true`
- `GetDataScopes(user_id=123, scope_type='community')` → 返回 `[101, 102]`

**数据隔离**：
- 用户 A 查询通知列表 → SQL 自动加 `WHERE community_id IN (101, 102)`

### 5.3 角色分配幂等性

- 使用 `INSERT IGNORE` 或 `ON DUPLICATE KEY UPDATE`
- 重复分配相同角色+作用域，不返回错误
- 前端无需额外处理幂等性

### 5.4 最后一个角色保护

- 撤销角色前，检查用户是否只剩一个角色
- 若是，拒绝撤销，返回错误码 060007
- 前端提示："用户至少需要保留一个角色"

**例外**：owner 角色可以撤销自己的最后一个角色（系统管理员特权）

---

## 六、界面交互

### 6.1 用户角色列表页面

**路径**：`/users/:userId/roles`

**布局**：
- 顶部：用户信息卡片（用户名、手机号、状态）
- 右上角："分配角色"按钮（`v-permission="'user:assign-role'"`）
- 表格：用户角色列表
- 分页器（如果角色数量较多）

**表格列定义**：
- 角色名称
- 角色编码
- 作用域类型（Tag：community=primary，building=success，unit=warning，grid=info）
- 作用域名称（需前端关联查询 master-data-service）
- 分配时间（宽度 180px）
- 操作（宽度 100px）：撤销按钮

### 6.2 分配角色对话框

**对话框宽度**：600px

**表单布局**：
- Label 宽度：120px
- 字段：
  - 用户信息（只读显示）：{userName} ({phone})
  - 角色选择（必填）：el-select，选项来自 `ListRoles`
  - 作用域类型（必填）：el-select，选项固定 4 种
  - 作用域实体（必填）：el-cascader（级联选择器）

**级联选择器配置**：
```typescript
// scope_type = 'community'
scopeOptions = [
  { value: '101', label: '阳光小区' },
  { value: '102', label: '和谐花园' }
];

// scope_type = 'building'
scopeOptions = [
  {
    value: '101',
    label: '阳光小区',
    children: [
      { value: '1001', label: '1号楼' },
      { value: '1002', label: '2号楼' }
    ]
  }
];

// scope_type = 'unit'
scopeOptions = [
  {
    value: '101',
    label: '阳光小区',
    children: [
      {
        value: '1001',
        label: '1号楼',
        children: [
          { value: '10001', label: '1单元' },
          { value: '10002', label: '2单元' }
        ]
      }
    ]
  }
];
```

**底部按钮**：
- 取消（关闭对话框）
- 确定（提交表单，loading 状态）

**校验提示**：
- 角色未选择："请选择角色"
- 作用域类型未选择："请选择作用域类型"
- 作用域实体未选择："请选择作用域实体"

### 6.3 撤销二次确认

**对话框类型**：MessageBox.confirm

**提示内容**：
```
确定撤销用户「{userName}」在「{scopeName}」的「{roleName}」角色吗？

撤销后该用户将无法访问此作用域的数据。
```

**按钮**：
- 取消（默认聚焦）
- 确定（红色危险按钮）

---

## 七、作用域名称关联查询

### 7.1 问题

- `rel_user_role` 表只存储 `scope_id`，不存储 `scope_name`
- 前端展示时需要关联查询 master-data-service 获取名称

### 7.2 解决方案

**方案 A（推荐）**：前端批量查询

```typescript
// 1. 获取用户角色列表
const { roles } = await getUserRoles(userId);

// 2. 提取所有 scope_id，按 scope_type 分组
const communityIds = roles.filter(r => r.scope_type === 'community').map(r => r.scope_id);
const buildingIds = roles.filter(r => r.scope_type === 'building').map(r => r.scope_id);

// 3. 批量查询名称
const communities = await getCommunities({ ids: communityIds });
const buildings = await getBuildings({ ids: buildingIds });

// 4. 关联映射
const communityMap = new Map(communities.map(c => [c.id, c.name]));
roles.forEach(r => {
  if (r.scope_type === 'community') {
    r.scope_name = communityMap.get(r.scope_id) || '未知小区';
  }
});
```

**方案 B**：后端关联查询（跨服务 gRPC 调用）

- `permission-service.GetUserRoles` 内部调用 `master-data-service.GetCommunityByIds`
- 返回时直接包含 `scope_name`
- **缺点**：增加服务间耦合，性能略差

**选择**：方案 A（前端批量查询，解耦服务）

---

## 八、错误处理

### 8.1 错误码映射

| 后端错误码 | 前端提示 |
|-----------|---------|
| 060001 | 角色不存在 |
| 060005 | 不支持的数据范围类型 |
| 060007 | 用户至少需要保留一个角色 |
| 060003 | 无权限操作 |

### 8.2 通用错误处理

- 网络错误：显示"网络请求失败，请稍后重试"
- 角色分配失败：显示具体错误信息
- 批量分配部分失败：显示"成功 {M}/{N}，失败用户：{userNames}"

---

## 九、测试场景

### 9.1 单元测试

**前端**：
- 级联选择器数据加载正确
- 作用域名称关联查询正确
- 批量分配循环调用正确

**后端**：
- `AssignRole` 幂等性正确（重复分配不报错）
- `RevokeRole` 最后一个角色保护生效（拒绝撤销）
- 缓存失效正确（`perm:user:*` 被删除）

### 9.2 集成测试

- E2E 流程：登录 → 用户列表 → 用户详情 → 角色列表 → 分配角色 → 撤销角色
- 多作用域角色：为同一用户分配相同角色+不同作用域，数据隔离生效
- 权限生效：分配角色后，用户立即获得对应权限（缓存重建）

### 9.3 边界条件

- 作用域实体不存在（ID 无效），显示"未知作用域"
- 撤销最后一个角色，显示错误提示
- 批量分配 100 个用户，进度提示正确

---

## 十、依赖与约束

### 10.1 依赖

| 依赖 | 说明 |
|------|------|
| `permission-service` | 提供 gRPC 接口 |
| `master-data-service` | 提供作用域实体数据（小区/楼栋/单元/网格） |
| `role-management` | 依赖角色列表接口 |

### 10.2 约束

- 作用域类型固定 4 种（community / building / unit / grid）
- 用户至少保留一个角色（防止锁死）
- 作用域实体ID 不强制校验存在性（避免跨服务依赖）

---

## 十一、追溯

### 11.1 需求来源

- 用户原话："数据权限一并考虑，只能看到授权了的数据"
- Brainstorming 决策点 3：数据权限作用域 → 选择"分配角色时指定作用域"

### 11.2 关联 Spec

- `role-management/spec.md`：依赖角色列表接口
- `api-permission-control/spec.md`：API 权限控制依赖数据范围隔离

### 11.3 记忆触发

- [[proto-jstype]]：所有 `int64` ID 字段必须添加 `[jstype = JS_STRING]`
- [[grpc-only-comms]]：服务间通信仅通过 gRPC
