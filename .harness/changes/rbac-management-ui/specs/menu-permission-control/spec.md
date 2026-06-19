# Spec: 菜单权限控制 (Menu Permission Control)

## 一、能力概述

根据用户权限动态过滤前端路由和菜单，无权限的菜单不显示，访问受限路由自动跳转 403 页面。提供 `v-permission` 指令实现按钮级权限控制。

---

## 二、功能需求

### 2.1 动态路由过滤

**功能描述**：用户登录后，根据其权限动态过滤路由表，只展示有权限的菜单。

**实现位置**：Vue Router 导航守卫（`router/permission.ts`）

**核心流程**：
```
1. 用户登录 → JWT 包含 user_id
2. 前端调用 GetUserPermissions(user_id) → 获取权限码列表
3. 将权限码列表存入 Pinia Store（userStore.permissions）
4. 路由守卫：对比 route.meta.permissions 与 userStore.permissions
5. 有权限 → 放行；无权限 → 跳转 403
```

**权限码获取**：
- API：`permission-service.GetUserPermissions(user_id)`
- 返回示例：`["user:menu", "user:read", "role:menu", "role:permission"]`
- 存储位置：`Pinia userStore.permissions: string[]`

**路由元信息定义**：
```typescript
// src/router/index.ts
{
  path: '/users',
  name: 'Users',
  component: () => import('@/views/identity/UserList.vue'),
  meta: {
    title: '用户管理',
    icon: 'User',
    permissions: ['user:menu'],  // 需要的权限码（菜单级）
    requiresAuth: true
  }
}
```

**过滤逻辑**：
```typescript
// src/router/permission.ts
router.beforeEach((to, from, next) => {
  const userStore = useUserStore();
  
  // 1. 检查是否需要登录
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    return next({ path: '/login', query: { redirect: to.fullPath } });
  }
  
  // 2. 检查权限
  if (to.meta.permissions) {
    const hasPermission = (to.meta.permissions as string[]).some(
      permission => userStore.permissions.includes(permission)
    );
    if (!hasPermission) {
      return next({ path: '/403' });
    }
  }
  
  next();
});
```

---

### 2.2 侧边栏菜单过滤

**功能描述**：根据用户权限动态生成侧边栏菜单，无权限的菜单项不显示。

**实现位置**：`components/layout/Sidebar.vue`

**过滤逻辑**：
```typescript
// src/utils/permission.ts
export function filterMenus(routes: RouteRecordRaw[], permissions: string[]): RouteRecordRaw[] {
  return routes.filter(route => {
    // 1. 无权限要求 → 显示
    if (!route.meta?.permissions) {
      return true;
    }
    
    // 2. 检查权限码
    const hasPermission = (route.meta.permissions as string[]).some(
      permission => permissions.includes(permission)
    );
    
    // 3. 递归过滤子菜单
    if (hasPermission && route.children) {
      route.children = filterMenus(route.children, permissions);
    }
    
    return hasPermission;
  });
}
```

**使用示例**：
```vue
<!-- components/layout/Sidebar.vue -->
<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/stores/user';
import { filterMenus } from '@/utils/permission';

const router = useRouter();
const userStore = useUserStore();

const menus = computed(() => {
  const routes = router.getRoutes().filter(r => !r.meta?.hidden);
  return filterMenus(routes, userStore.permissions);
});
</script>

<template>
  <el-menu :default-active="activeMenu" mode="vertical">
    <sidebar-item v-for="route in menus" :key="route.path" :route="route" />
  </el-menu>
</template>
```

---

### 2.3 按钮级权限指令 `v-permission`

**功能描述**：自定义指令，控制按钮显示/隐藏。

**实现位置**：`directives/permission.ts`

**指令定义**：
```typescript
// src/directives/permission.ts
import type { Directive, DirectiveBinding } from 'vue';
import { useUserStore } from '@/stores/user';

export const permission: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding<string | string[]>) {
    const userStore = useUserStore();
    const requiredPermissions = Array.isArray(binding.value) ? binding.value : [binding.value];
    
    const hasPermission = requiredPermissions.some(
      permission => userStore.permissions.includes(permission)
    );
    
    if (!hasPermission) {
      el.parentNode?.removeChild(el);  // 移除 DOM 节点
    }
  }
};
```

**注册指令**：
```typescript
// src/main.ts
import { createApp } from 'vue';
import { permission } from '@/directives/permission';

const app = createApp(App);
app.directive('permission', permission);
```

**使用示例**：
```vue
<template>
  <div>
    <!-- 单个权限 -->
    <el-button v-permission="'user:create'" type="primary" @click="handleCreate">
      创建用户
    </el-button>
    
    <!-- 多个权限（任一满足即显示） -->
    <el-button v-permission="['user:update', 'user:delete']" @click="handleEdit">
      编辑
    </el-button>
  </div>
</template>
```

---

### 2.4 权限缓存与刷新

**功能描述**：权限数据缓存到 LocalStorage，页面刷新时自动加载。

**存储策略**：
```typescript
// src/stores/user.ts
import { defineStore } from 'pinia';

export const useUserStore = defineStore('user', {
  state: () => ({
    permissions: [] as string[],
    token: '',
    userInfo: null as any
  }),
  
  actions: {
    // 登录后获取权限
    async fetchPermissions() {
      const res = await getUserPermissions(this.userInfo.id);
      this.permissions = res.permissions || [];
      localStorage.setItem('permissions', JSON.stringify(this.permissions));
    },
    
    // 页面刷新时恢复权限
    restorePermissions() {
      const cached = localStorage.getItem('permissions');
      if (cached) {
        this.permissions = JSON.parse(cached);
      }
    },
    
    // 登出时清除权限
    logout() {
      this.permissions = [];
      localStorage.removeItem('permissions');
      localStorage.removeItem('token');
    }
  }
});
```

**刷新策略**：
- 登录时：调用 `fetchPermissions()` 获取最新权限
- 页面刷新时：从 LocalStorage 恢复权限（避免闪烁）
- Token 刷新时：不重新获取权限（权限变更 5 分钟内生效）
- 登出时：清除所有权限缓存

---

## 三、数据模型

### 3.1 前端 TypeScript 类型

```typescript
// src/types/permission.ts

export interface UserPermissions {
  user_id: string;
  permissions: string[];         // 权限码列表
}

// 路由元信息扩展
declare module 'vue-router' {
  interface RouteMeta {
    title?: string;              // 页面标题
    icon?: string;               // 菜单图标
    permissions?: string[];      // 需要的权限码（菜单级）
    requiresAuth?: boolean;      // 是否需要登录
    hidden?: boolean;            // 是否隐藏菜单
  }
}
```

### 3.2 后端 Proto 定义

```protobuf
// api-proto/api/permission/v1/permission.proto

message GetUserPermissionsReq {
  int64 user_id = 1 [jstype = JS_STRING];
}

message GetUserPermissionsResp {
  common.v1.BaseResp base_resp = 1;
  repeated string permissions = 2;  // 权限码列表（如 ["user:menu", "user:read"]）
}
```

---

## 四、接口清单

### 4.1 前端 API 模块

**文件**：`src/api/permission.ts`

```typescript
import request from '@/utils/request';

// 获取用户权限码列表
export function getUserPermissions(userId: string) {
  return request.get<{ permissions: string[] }>(`/api/v1/permission/users/${userId}/permissions`);
}
```

### 4.2 后端 gRPC 接口

**服务**：`permission-service`

| RPC | 请求 | 响应 | 说明 |
|-----|------|------|------|
| `GetUserPermissions` | `GetUserPermissionsReq` | `GetUserPermissionsResp` | 已存在，返回用户所有权限码 |

**实现逻辑**：
```go
func (l *GetUserPermissionsLogic) GetUserPermissions(in *pb.GetUserPermissionsReq) (*pb.GetUserPermissionsResp, error) {
    // 1. 查询用户所有角色
    userRoles, err := l.svcCtx.UserRoleModel.FindActiveByUserId(ctx, in.UserId)
    if err != nil {
        return nil, err
    }
    
    // 2. 系统角色直接返回全部权限（或特殊标记）
    for _, ur := range userRoles {
        role, _ := l.svcCtx.RoleModel.FindOne(ctx, ur.RoleId)
        if role.IsSystem == 1 {
            // 系统角色拥有全部权限
            allPermissions, _ := l.svcCtx.PermissionModel.FindAll(ctx)
            codes := make([]string, len(allPermissions))
            for i, p := range allPermissions {
                codes[i] = p.Code
            }
            return &pb.GetUserPermissionsResp{
                BaseResp: &pb.BaseResp{Code: 0, Msg: "success"},
                Permissions: codes,
            }, nil
        }
    }
    
    // 3. 收集所有角色的权限码
    permissionSet := make(map[string]bool)
    for _, ur := range userRoles {
        perms, _ := l.svcCtx.RolePermissionModel.FindPermissionsByRoleId(ctx, ur.RoleId)
        for _, p := range perms {
            permissionSet[p.Code] = true
        }
    }
    
    // 4. 转为切片返回
    permissions := make([]string, 0, len(permissionSet))
    for code := range permissionSet {
        permissions = append(permissions, code)
    }
    
    return &pb.GetUserPermissionsResp{
        BaseResp: &pb.BaseResp{Code: 0, Msg: "success"},
        Permissions: permissions,
    }, nil
}
```

---

## 五、业务规则

### 5.1 权限检查逻辑

**菜单级权限**（`route.meta.permissions`）：
- 权限码格式：`{module}:menu`（如 `user:menu`）
- 检查逻辑：`permissions.includes('user:menu')`
- 无权限行为：菜单不显示，访问路由跳转 403

**按钮级权限**（`v-permission`）：
- 权限码格式：`{module}:{action}`（如 `user:create`）
- 检查逻辑：`permissions.includes('user:create')`
- 无权限行为：按钮不显示（DOM 移除）

**API 级权限**（后端拦截，见 `api-permission-control` spec）：
- 权限码格式：`{module}:{action}:api`（如 `user:create:api`）
- 检查逻辑：后端 `CheckPermission` RPC
- 无权限行为：返回 403 Forbidden

### 5.2 系统角色特权

- 系统角色（`is_system=1`）在 `GetUserPermissions` 中返回所有权限码
- 系统角色用户可见所有菜单和按钮
- 系统角色的 API 请求在 `CheckPermission` 时直接放行

### 5.3 权限码继承

**场景**：用户拥有 `user:menu` 权限时，是否自动拥有 `user:read` 权限？

**规则**：不自动继承。权限是扁平的，不存在父子继承关系。

**原因**：避免权限配置复杂度，权限树仅用于组织展示，不影响校验逻辑。

### 5.4 多角色权限合并

- 用户拥有多个角色时，权限码取**并集**
- 示例：角色 A 有 `["user:read"]`，角色 B 有 `["role:read"]` → 用户有 `["user:read", "role:read"]`

---

## 六、界面交互

### 6.1 登录后权限加载

**流程**：
```
1. 用户输入账号密码 → 登录
2. auth-service 验证 → 返回 JWT（含 user_id）
3. 前端解析 JWT，提取 user_id
4. 调用 GetUserPermissions(user_id) → 获取权限码列表
5. 存入 Pinia Store 和 LocalStorage
6. 跳转首页，侧边栏菜单动态过滤
```

**加载提示**：
- 登录成功后，显示全屏 Loading："加载权限中..."
- 权限加载完成后，Loading 消失，展示首页

### 6.2 无权限路由访问

**场景**：用户手动输入 URL（如 `/users`），但无 `user:menu` 权限。

**行为**：
1. 路由守卫检测到无权限
2. 自动跳转 `/403` 页面
3. 403 页面显示："您没有权限访问此页面"
4. 提供"返回首页"按钮

**403 页面示例**：
```vue
<!-- views/error/403.vue -->
<template>
  <div class="error-page">
    <el-result icon="warning" title="403" sub-title="您没有权限访问此页面">
      <template #extra>
        <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
      </template>
    </el-result>
  </div>
</template>
```

### 6.3 权限变更实时生效

**场景**：管理员修改用户角色权限后，用户何时生效？

**策略**：
- **前端**：权限缓存在 LocalStorage，刷新页面后生效
- **后端**：权限缓存在 Redis（TTL 30min），修改后立即失效，下次请求重建

**用户提示**：
- 管理员修改权限后，前端显示提示："权限配置已更新，请通知用户刷新页面"
- 用户侧：如果 API 请求返回 403，提示"权限已变更，请刷新页面"

---

## 七、路由配置示例

### 7.1 模块化路由配置

**文件结构**：
```
src/config/modules/
  ├─ dashboard.config.ts       # 首页模块
  ├─ identity.config.ts        # 用户&角色模块
  ├─ masterdata.config.ts      # 主数据模块
  └─ ...
```

**示例**（`identity.config.ts`）：
```typescript
import type { RouteRecordRaw } from 'vue-router';

export const identityModule = {
  name: 'identity',
  routes: [
    {
      path: '/users',
      name: 'Users',
      component: () => import('@/views/identity/UserList.vue'),
      meta: {
        title: '用户管理',
        icon: 'User',
        permissions: ['user:menu'],
        requiresAuth: true
      }
    },
    {
      path: '/roles',
      name: 'Roles',
      component: () => import('@/views/roles/List.vue'),
      meta: {
        title: '角色管理',
        icon: 'Avatar',
        permissions: ['role:menu'],
        requiresAuth: true
      }
    },
    {
      path: '/roles/:id/permissions',
      name: 'RolePermissions',
      component: () => import('@/views/roles/Permissions.vue'),
      meta: {
        title: '权限配置',
        hidden: true,  // 不在菜单中显示
        permissions: ['role:permission'],
        requiresAuth: true
      }
    }
  ] as RouteRecordRaw[]
};
```

### 7.2 路由守卫注册

**文件**：`src/router/permission.ts`

```typescript
import router from './index';
import { useUserStore } from '@/stores/user';
import { ElMessage } from 'element-plus';

// 白名单（无需登录）
const whiteList = ['/login', '/register', '/404', '/403'];

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();
  
  // 1. 白名单直接放行
  if (whiteList.includes(to.path)) {
    return next();
  }
  
  // 2. 检查登录状态
  if (!userStore.token) {
    ElMessage.warning('请先登录');
    return next({ path: '/login', query: { redirect: to.fullPath } });
  }
  
  // 3. 权限未加载，先加载
  if (userStore.permissions.length === 0) {
    try {
      await userStore.fetchPermissions();
    } catch (error) {
      ElMessage.error('权限加载失败，请重新登录');
      userStore.logout();
      return next('/login');
    }
  }
  
  // 4. 检查路由权限
  if (to.meta.permissions && to.meta.permissions.length > 0) {
    const hasPermission = (to.meta.permissions as string[]).some(
      permission => userStore.permissions.includes(permission)
    );
    if (!hasPermission) {
      return next('/403');
    }
  }
  
  next();
});
```

---

## 八、错误处理

### 8.1 权限加载失败

**场景**：`GetUserPermissions` 接口调用失败（网络错误、服务异常）。

**处理**：
1. 显示 ElMessage.error("权限加载失败，请重新登录")
2. 清除 token 和权限缓存
3. 跳转登录页面

### 8.2 权限缓存不一致

**场景**：LocalStorage 缓存的权限与后端实际权限不一致。

**处理**：
1. 前端每次刷新页面时，后台异步调用 `GetUserPermissions` 更新缓存
2. 如果 API 请求返回 403，提示"权限已变更，请刷新页面"
3. 用户点击"刷新"按钮，调用 `fetchPermissions()` 强制更新

---

## 九、测试场景

### 9.1 单元测试

**前端**：
- 路由守卫权限检查逻辑正确
- 侧边栏菜单过滤正确（有权限显示，无权限隐藏）
- `v-permission` 指令移除无权限按钮

**后端**：
- `GetUserPermissions` 返回正确权限码列表
- 系统角色返回全部权限
- 多角色权限合并正确（并集）

### 9.2 集成测试

- E2E 流程：登录 → 权限加载 → 侧边栏菜单过滤 → 访问无权限路由 → 跳转 403
- 权限变更：管理员修改权限 → 用户刷新页面 → 菜单动态更新
- 多角色用户：拥有多个角色，权限码合并正确

### 9.3 性能测试

- 权限加载时间 < 500ms
- 路由守卫检查耗时 < 10ms
- 侧边栏菜单过滤耗时 < 50ms（100+ 菜单项）

---

## 十、依赖与约束

### 10.1 依赖

| 依赖 | 说明 |
|------|------|
| `permission-service` | 提供 `GetUserPermissions` 接口 |
| `auth-service` | 提供 JWT（含 user_id） |
| Pinia | 状态管理（权限存储） |
| Vue Router | 路由守卫、动态路由 |

### 10.2 约束

- 权限码必须在权限树中预定义（不支持动态权限码）
- 权限变更 5 分钟内生效（依赖 Redis 缓存 TTL）
- 前端权限控制可被浏览器开发者工具绕过（需后端 API 权限拦截兜底）

---

## 十一、追溯

### 11.1 需求来源

- 用户原话："不同角色的人能够看到不同的菜单、前端能够看到不同的内容"
- Brainstorming 决策点 2：权限粒度设计 → 菜单+功能点两层结构

### 11.2 关联 Spec

- `permission-management/spec.md`：依赖权限树数据
- `api-permission-control/spec.md`：后端 API 权限拦截（兜底）
- `role-management/spec.md`：依赖角色列表接口

### 11.3 记忆触发

- [[proto-jstype]]：所有 `int64` ID 字段必须添加 `[jstype = JS_STRING]`
