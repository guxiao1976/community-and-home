基于 RBAC 的角色与数据隔离引擎。不管理具体业务，只输出“可见范围”。

## 数据模型

- **MySQL (`db_permission`)**:
    - `sys_role`: `id`, `role_code(owner/property_admin)`, `role_name`
    - `rel_user_role`: `id`, `user_id`, `role_id`, `scope_type(community)`, `scope_id(小区ID)`

## 接口清单

- `AssignRole(AssignRoleReq) returns (BaseResp)`
- `RevokeRole(RevokeRoleReq) returns (BaseResp)`
- `CheckPermission(CheckPermReq) returns (CheckPermResp)`
- `GetDataScopes(GetDataScopesReq) returns (DataScopesResp)`

## 核心逻辑流

1. **AssignRole**: 校验角色和作用域是否存在，插入 `rel_user_role` 表。
2. **GetDataScopes**: 查询 `rel_user_role`，根据 `user_id` 和 `scope_type` 返回 `scope_ids` 列表。供其他服务 GORM 拦截器实现 `WHERE community_id IN (scope_ids)`。
3. **CheckPermission**: 组合 User 的 Roles，查询对应的 API 权限集，判断是否包含请求的 API Path。
