## Why

当前系统已完成用户/角色/权限的基础 CRUD 操作，但 RBAC 权限控制的核心执行层未实现——Casbin 中间件是空壳、用户权限查询返回空数组、前端路由和菜单无权限过滤。这意味着任何已登录用户都能访问所有接口和页面，存在严重的安全隐患。需要完成权限控制的完整闭环，使角色与权限真正生效。

## What Changes

- 实现后端 Casbin 权限拦截中间件，对 API 请求执行 RBAC 校验（基于 userId + domain + resource + action）
- 实现获取用户权限列表 API（通过 user_role → role_permission → permission 链路查询）
- 实现 RPC 层权限检查逻辑（供其他微服务网关调用）
- 新增用户-角色分配/移除 API（当前只能在角色侧分配权限，无法在用户侧管理角色）
- 实现前端路由级权限守卫（启用 guards.ts 中已注释的权限检查逻辑）
- 实现前端侧边栏菜单按用户权限动态过滤显示
- 实现前端按钮/操作级权限控制（v-permission 自定义指令）
- 前端用户详情页补充角色与权限展示及角色分配功能

## Capabilities

### New Capabilities
- `rbac-enforcement`: 后端 Casbin RBAC 权限拦截、用户权限查询、RPC 权限检查、用户-角色分配的完整后端实现
- `frontend-permission-control`: 前端路由权限守卫、菜单动态过滤、按钮级权限指令、用户角色管理 UI

### Modified Capabilities

## Impact

- **后端 Identity API**: permissionmiddleware.go、user logic、role logic、permission logic
- **后端 Identity RPC**: check_permission_logic.go、get_user_permissions_logic.go
- **前端路由**: guards.ts、router/index.ts（路由 meta 需补充 permission 字段）
- **前端组件**: AppSidebar.vue（菜单动态过滤）、users/Detail.vue（角色分配 UI）
- **前端工具**: permission.ts（v-permission 指令注册）
- **前端 Store**: permission.ts（权限加载与缓存逻辑）
- **数据库**: 无 schema 变更（auth_user_role、auth_role_permission、auth_permission 表已存在）
- **依赖**: Casbin 已引入，无需新增外部依赖
