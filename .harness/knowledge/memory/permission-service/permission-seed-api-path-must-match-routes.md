---
triggers: ["seed", "init_permissions", "API path", "权限路径", "permission", "CheckPermission", "path 前缀", "路由匹配"]
service: permission-service
type: pitfall
severity: must-follow
status: active
created: 2026-08-09
updated: 2026-08-09
last_applied: 2026-08-09
apply_count: 1
---

# 种子数据 API 权限的 path 必须与实际 REST 路由一致

## 为什么会有这条经验

`CheckPermission` 的匹配逻辑是 `needle = "{Action}:{ApiPath}"` 比对 `sys_permission.path`。原始种子数据中全部 API 权限的 path 使用 `/api/v1/` 前缀（如 `GET:/api/v1/permission/roles`），但实际 REST 网关使用 `/api/perm/`、`/api/users/` 等前缀。**前缀不一致导致所有 type=3 的权限匹配永久失败**。

## 如何应用

1. 新增或修改 `sys_permission` 的 `path` 值时，必须逐条对照实际 REST 路由确认路径一致
2. 建议在 CI 中增加自动化检查脚本：`SELECT path FROM sys_permission WHERE type=3` → 比对 `routes.go` 和 `api-proto` 中的实际路径
3. path 格式统一为 `{METHOD}:{/api/prefix/path}`，如 `GET:/api/perm/roles`
4. 仅 gRPC（无 REST 代理）的端点不应有 type=3 权限记录

## 本次修正清单

| 原路径（错误）| 修正后 |
|-------------|--------|
| `GET:/api/v1/permission/roles` | `GET:/api/perm/roles` |
| `POST:/api/v1/user/users` | `POST:/api/users` |
| `POST:/api/v1/permission/assign-role` | `POST:/api/perm/user-roles`（新增端点后） |
| ... | (共修正 16 条 API path) |

## 相关记忆

- [[is-system-no-permission-shortcut]] — is_system 不再授予全权限，权限检查完全由 path 匹配决定
