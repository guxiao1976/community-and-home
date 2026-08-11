---
triggers: ["is_system", "系统角色", "permission", "权限", "CheckPermission", "短路", "RBAC"]
service: permission-service
type: decision
severity: must-follow
status: active
created: 2026-08-09
updated: 2026-08-09
last_applied: 2026-08-09
apply_count: 1
---

# is_system 字段仅用于保护角色不可删除/修改，不再授予全权限

## 为什么会有这条经验

原始设计中 `is_system=1` 做了两件事：(1) 角色不可删除 (2) `CheckPermission` 中直接放行（代码短路）。这导致 4 个系统角色（owner/property_admin/community_admin/grid_worker）拥有全部权限，权限系统形同虚设。

**改后语义**：`is_system=1` 仅表示角色不可删除+不可修改。权限检查走统一的 `rel_role_permission` 配置路径，无任何 flag 短路。如需超级管理员，通过管理后台新建自定义角色并全选权限。

## 如何应用

1. **CheckPermission**: 必须走 `rel_role_permission → sys_permission → needle 匹配` 全路径，禁止任何基于字段的短路
2. **GetUserPermissions**: 同上，通过 `rel_role_permission` 收集权限 ID，禁止 `is_system=1` 时直接 `FindAll()`
3. **DeleteRole**: 保留 `is_system=1` 返回错误 60004
4. **UpdateRole**: 新增 `is_system=1` 返回错误 60004（原实现漏掉了此检查）
5. **CreateRole**: 禁止设置 `is_system=1`（内置角色仅通过 SQL 初始化，`CreateRole` 不设置此字段）

## 相关改动

- `checkpermissionlogic.go`: 删除 L44-49 短路代码
- `getuserpermissionslogic.go`: 删除 L37-45 系统角色分支
- `updaterolelogic.go`: 新增 is_system 不可修改检查
- `deleterolelogic.go`: 保持不变
- `init_permissions.sql`: CROSS JOIN 全权限 → 按 4 角色精确赋权

## 相关记忆

- [[verify-api-before-calling]] — API 端点必须先确认存在才能调用
