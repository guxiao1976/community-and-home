## ADDED Requirements

### Requirement: Casbin permission middleware enforcement
The permission middleware SHALL enforce RBAC policies via Casbin enforcer for all authenticated API requests. The middleware SHALL extract userId and scopeId from JWT context, map them to Casbin sub and dom, use the request path as obj and HTTP method as act, and call `enforcer.Enforce()`. Requests that fail enforcement SHALL return HTTP 403 with error message "无权限访问". Requests from users with a system role (is_system=1) SHALL bypass Casbin enforcement and be allowed through.

#### Scenario: Authorized request passes middleware
- **WHEN** an authenticated user with role "community_manager" (role:2) makes a GET request to `/api/identity/users` and Casbin has policy `(role:2, scope:100, GET:/api/identity/users, allow)`
- **THEN** the middleware calls `next()` and the request proceeds to the handler

#### Scenario: Unauthorized request is blocked
- **WHEN** an authenticated user with role "community_manager" (role:2) makes a DELETE request to `/api/identity/users/1` and Casbin has no matching allow policy
- **THEN** the middleware returns HTTP 403 with body `{ code: 403, message: "无权限访问" }`

#### Scenario: Super admin bypasses Casbin
- **WHEN** an authenticated user who has role "super_admin" (is_system=1) makes any request to any endpoint
- **THEN** the middleware skips Casbin enforcement and calls `next()` directly

#### Scenario: Unauthenticated request is rejected
- **WHEN** a request arrives without userId in context
- **THEN** the middleware returns HTTP 401 with error message "未授权"

### Requirement: Get user permissions API
The `GET /api/identity/users/:id/permissions` endpoint SHALL return the user's effective permission codes by querying through the auth_user_role → auth_role_permission → auth_permission chain. The response SHALL include a flat list of permission codes (strings) and a tree of menu-type permissions (type=1) for sidebar rendering. Only permissions with status=1 (active) SHALL be included.

#### Scenario: User with roles returns permission codes
- **WHEN** admin requests `GET /api/identity/users/5/permissions` and user 5 has role "community_manager" with permissions `[user:list, user:create, community:view]`
- **THEN** the response contains `{ permissions: ["user:list", "user:create", "community:view"], menus: [...] }`

#### Scenario: User with no roles returns empty permissions
- **WHEN** admin requests `GET /api/identity/users/5/permissions` and user 5 has no role assignments
- **THEN** the response contains `{ permissions: [], menus: [] }`

#### Scenario: User with multiple roles merges permissions
- **WHEN** admin requests `GET /api/identity/users/5/permissions` and user 5 has role "community_manager" with `[user:list]` and role "reviewer" with `[community:view, community:review]`
- **THEN** the response contains `{ permissions: ["user:list", "community:view", "community:review"], menus: [...] }` with duplicates removed

### Requirement: RPC CheckPermission implementation
The RPC `CheckPermission` method SHALL verify whether a given user has permission to perform a specific action on a resource within a domain. It SHALL query the user's roles, then check Casbin policies for each role. The response SHALL include a boolean `allowed` field.

#### Scenario: User has permission
- **WHEN** RPC receives CheckPermissionReq with userId=5, domain="100", resource="/api/identity/users", action="GET"
- **THEN** the response contains `{ allowed: true }`

#### Scenario: User lacks permission
- **WHEN** RPC receives CheckPermissionReq with userId=5, domain="100", resource="/api/identity/users", action="DELETE"
- **THEN** the response contains `{ allowed: false }`

### Requirement: RPC GetUserPermissions implementation
The RPC `GetUserPermissions` method SHALL return all permission codes for a given user by querying through user_role → role_permission → permission chain. The response SHALL include a list of permission code strings.

#### Scenario: Returns user permission codes
- **WHEN** RPC receives GetUserPermissionsReq with userId=5
- **THEN** the response contains `{ permissions: ["user:list", "user:create"] }`

### Requirement: User role assignment API
The system SHALL provide `POST /api/identity/users/:id/roles` endpoint to assign one or more roles to a user. The request body SHALL contain a list of role IDs. The system SHALL validate that all role IDs exist and are active (status=1). Duplicate assignments SHALL be ignored. The response SHALL indicate success.

#### Scenario: Assign roles to user
- **WHEN** admin sends `POST /api/identity/users/5/roles` with body `{ role_ids: [2, 3] }`
- **THEN** the system creates auth_user_role records linking user 5 to roles 2 and 3, and returns `{ success: true }`

#### Scenario: Assign non-existent role returns error
- **WHEN** admin sends `POST /api/identity/users/5/roles` with body `{ role_ids: [999] }`
- **THEN** the system returns error indicating role 999 does not exist

#### Scenario: Duplicate role assignment is idempotent
- **WHEN** admin sends `POST /api/identity/users/5/roles` with body `{ role_ids: [2] }` and user 5 already has role 2
- **THEN** the system does not create a duplicate record and returns `{ success: true }`

### Requirement: User role removal API
The system SHALL provide `DELETE /api/identity/users/:id/roles/:roleId` endpoint to remove a specific role from a user. The system SHALL validate that the role assignment exists before removal. System roles (is_system=1) SHALL NOT be removable from their assigned users through this endpoint.

#### Scenario: Remove role from user
- **WHEN** admin sends `DELETE /api/identity/users/5/roles/3` and user 5 has role 3 assigned
- **THEN** the system deletes the auth_user_role record and returns `{ success: true }`

#### Scenario: Remove non-assigned role returns error
- **WHEN** admin sends `DELETE /api/identity/users/5/roles/3` and user 5 does not have role 3
- **THEN** the system returns error indicating the role assignment does not exist

#### Scenario: Cannot remove system role
- **WHEN** admin sends `DELETE /api/identity/users/5/roles/1` and role 1 is a system role (is_system=1)
- **THEN** the system returns error indicating system roles cannot be removed

### Requirement: Get user roles API
The `GET /api/identity/users/:id/roles` endpoint SHALL return all roles assigned to a specific user, including role details (id, name, code, description, is_system, status).

#### Scenario: Returns user roles
- **WHEN** admin requests `GET /api/identity/users/5/roles`
- **THEN** the response contains a list of roles assigned to user 5 with full role details

### Requirement: Sync role permissions to Casbin on user-role change
When roles are assigned to or removed from a user, the system SHALL reload the affected user's role-policy mappings into the Casbin enforcer via `LoadPolicy()` to ensure policy consistency. This ensures that the permission middleware reflects the latest role assignments.

#### Scenario: Policy reload after role assignment
- **WHEN** admin assigns role 2 to user 5
- **THEN** the Casbin enforcer reloads policies and subsequent requests from user 5 reflect the new permissions
