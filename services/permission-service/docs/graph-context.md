# 知识图谱上下文 — permission-service

> 自动生成于 2026-08-12 18:13:52 | 数据源: Neo4j 知识图谱 | 每次 `graph-sync.sh` 后刷新

## 服务标识

| 属性 | 值 |
|------|-----|
| 名称 | permission-service |
| 语言 | go |
| 端口 (gRPC) | 8084 |
| 端口 (API)  | 8883 |

## 服务依赖

| 依赖服务 | 依赖类型 |
|---------|---------|
| master-data-service | gRPC |

## 被依赖方

| 消费方 | 依赖类型 |
|---------|---------|
| community-hub-service | None |
| master-data-service | None |
| user-service | None |

## REST API 路由

| 方法 | 路径 |
|------|------|
| Get | /api/perm/permissions |
| Post | /api/perm/roles |
| Get | /api/perm/roles |
| Delete | /api/perm/roles/:id |
| Put | /api/perm/roles/:id |
| Get | /api/perm/roles/:id |
| Post | /api/perm/roles/:id/permissions |
| Get | /api/perm/roles/:id/permissions |
| Delete | /api/perm/user-roles |
| Post | /api/perm/user-roles |
| Get | /api/perm/users/:userId/permissions |
| Get | /api/perm/users/:userId/roles |

## gRPC 接口

| RPC 方法 | 输入消息 | 输出消息 |
|---------|---------|---------|
| AssertPublishScope | AssertPublishScopeRequest | AssertPublishScopeResponse |
| AssignRole | AssignRoleRequest | AssignRoleResponse |
| CheckPermission | CheckPermissionRequest | CheckPermissionResponse |
| CreateRole | CreateRoleRequest | CreateRoleResponse |
| DeleteRole | DeleteRoleRequest | DeleteRoleResponse |
| GetDataScopes | GetDataScopesRequest | GetDataScopesResponse |
| GetRole | GetRoleRequest | GetRoleResponse |
| GetRolesByIds | GetRolesByIdsRequest | GetRolesByIdsResponse |
| GetUserPermissions | GetUserPermissionsRequest | GetUserPermissionsResponse |
| GetUserRoles | GetUserRolesRequest | GetUserRolesResponse |
| InvalidateUserCache | InvalidateUserCacheRequest | InvalidateUserCacheResponse |
| ListPermissions | ListPermissionsRequest | ListPermissionsResponse |
| ListRoles | ListRolesRequest | ListRolesResponse |
| RevokeRole | RevokeRoleRequest | RevokeRoleResponse |
| UpdateRole | UpdateRoleRequest | UpdateRoleResponse |
| UpdateUserRoleStatus | UpdateUserRoleStatusRequest | UpdateUserRoleStatusResponse |

## 数据库表

| 表名 | 列 |
|------|-----|
| rel_role_permission | ur_status (bigint), expires_at (nullable), verified_at (nullable), status (bigint), created_at (datetime), description (varchar), role_status (bigint), is_system (bigint), role_name (varchar), role_code (varchar), scope_id (bigint), scope_type (varchar), user_id (bigint), created_time (datetime), permission_id (bigint) ... |
| sys_role | min_verf_level (bigint), deleted_at (nullable), updated_at (datetime), created_at (datetime), icon (nullable), path (nullable), type (bigint), code (varchar), name (varchar), parent_id (nullable), delete_time (nullable), updated_time (datetime), created_time (datetime), created_by (bigint), status (bigint) ... |

## 前端消费方

| 方法 | URL | 文件 |
|------|-----|------|
| DELETE | /api/perm/user-roles | web/pc/src/api/identity.ts |
| POST | /api/perm/user-roles | web/pc/src/api/identity.ts |
| GET | /api/perm/permissions | web/pc/src/api/identity.ts |
| POST | /api/perm/roles | web/pc/src/api/identity.ts |
| GET | /api/perm/roles | web/pc/src/api/identity.ts |

## 实体血缘（Proto → Go → DB）

| Proto 消息 | Go 结构体 | 数据库表 |
|-----------|----------|---------|
| UserRoleInfo | UserRoleInfo | - |

---
*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*
