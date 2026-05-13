## 1. Backend: Model Layer Enhancement

- [x] 1.1 Add `DeleteByUserIdAndRoleId` method to AuthUserRoleModel for removing a specific user-role assignment
- [x] 1.2 Add `FindByUserIdAndRoleId` method to AuthUserRoleModel for checking if a specific assignment exists
- [x] 1.3 Add `BatchInsertUserRoles` method to AuthUserRoleModel for assigning multiple roles at once (with duplicate handling via INSERT IGNORE)
- [x] 1.4 Add `FindActiveByUserId` method to AuthUserRoleModel that joins auth_role to only return active roles (status=1)

## 2. Backend: User Permission Query Implementation

- [x] 2.1 Implement `GetUserPermissionsLogic` in `api/internal/logic/user/get_user_permissions_logic.go` — query user_role → role_permission → permission chain, return permission codes and menu tree
- [x] 2.2 Enhance `GetUserPermissionsResp` type in `identity.api` to include `menus` field (tree of menu-type permissions) and regenerate types

## 3. Backend: User-Role Assignment APIs

- [x] 3.1 Add `AssignUserRolesReq`/`AssignUserRolesResp`, `RemoveUserRoleReq`/`RemoveUserRoleResp`, `GetUserRolesReq`/`GetUserRolesResp` types to `identity.api`
- [x] 3.2 Add three new endpoints to `identity.api`: `POST /users/:id/roles`, `DELETE /users/:id/roles/:roleId`, `GET /users/:id/roles`
- [x] 3.3 Run `goctl` to regenerate handler and types from updated `.api` file
- [x] 3.4 Implement `assign_user_roles_logic.go` — validate role IDs exist and are active, batch insert with duplicate handling
- [x] 3.5 Implement `remove_user_role_logic.go` — validate assignment exists, prevent removal of system roles
- [x] 3.6 Implement `get_user_roles_logic.go` — return all active roles for a user with full details

## 4. Backend: Casbin Permission Middleware

- [x] 4.1 Implement Casbin enforce call in `permissionmiddleware.go` — map userId to role via user_role table, determine domain from scopeId, check `enforcer.Enforce(role:roleId, domain, method:path, allow)`
- [x] 4.2 Add system role bypass logic — query user's roles, if any has is_system=1, skip Casbin and allow
- [ ] 4.3 Add user-role cache to avoid DB query on every request — cache user's role list in Redis with short TTL (60s) or use go-zero BuiltInCache
- [x] 4.4 Wire permission middleware into the API router for JWT-protected routes (excluding public auth endpoints)

## 5. Backend: RPC Permission Implementation

- [x] 5.1 Implement `CheckPermissionLogic` in `rpc/internal/logic/check_permission_logic.go` — query user roles, check Casbin policies, return allowed boolean
- [x] 5.2 Implement `GetUserPermissionsLogic` in `rpc/internal/logic/get_user_permissions_logic.go` — query through user_role → role_permission → permission chain, return permission codes

## 6. Frontend: API Service Layer

- [x] 6.1 Add `assignUserRoles`, `removeUserRole`, `getUserRoles` API functions to `web/pc/src/api/identity.ts`
- [x] 6.2 Update `getUserPermissions` return type to include `menus` field in `web/common/types/identity.d.ts`

## 7. Frontend: Permission Store Enhancement

- [x] 7.1 Update `web/pc/src/stores/permission.ts` — add `loadUserPermissionsAndMenus` action that loads permissions on login and stores both codes and menu tree
- [x] 7.2 Add `userRoles` state and `loadUserRoles` action to permission store
- [x] 7.3 Integrate permission loading into auth flow — call `loadUserPermissionsAndMenus` after successful login in `stores/auth.ts` and on app init in `App.vue`

## 8. Frontend: Route Permission Guard

- [x] 8.1 Uncomment and activate the permission check logic in `web/pc/src/router/guards.ts` — check `to.meta.permission` against permission store
- [x] 8.2 Add `permission` meta field to all protected routes in `web/pc/src/router/index.ts` (user routes, role routes, verification routes, etc.)

## 9. Frontend: Dynamic Sidebar Menu

- [x] 9.1 Add `permissionCode` property to each menu item in `AppSidebar.vue`
- [x] 9.2 Replace static `menuItems` array with a computed property that filters based on permission store
- [x] 9.3 Hide parent menu items when all children are filtered out

## 10. Frontend: v-permission Directive

- [x] 10.1 Create `v-permission` custom directive in `web/pc/src/directives/permission.ts` — accept string or string[], check against permission store, set `display: none` if no match
- [x] 10.2 Register the directive globally in `web/pc/src/main.ts`
- [x] 10.3 Apply `v-permission` directive to action buttons in user management pages (create, edit, disable, enable, delete)
- [x] 10.4 Apply `v-permission` directive to action buttons in role management pages (create, edit, delete, assign permissions)
- [x] 10.5 Apply `v-permission` directive to action buttons in verification pages (approve, reject)

## 11. Frontend: User Detail Page Role & Permission UI

- [x] 11.1 Replace the "Phase 7" placeholder in `web/pc/src/views/users/Detail.vue` with a role assignment table
- [x] 11.2 Add role assignment dialog — select roles from available role list, call `assignUserRoles` API
- [x] 11.3 Add role removal action — call `removeUserRole` API, disable for system roles
- [x] 11.4 Add effective permissions tree display — use `PermissionTree` component to show user's permissions

## 12. Verification

- [ ] 12.1 Backend: Test Casbin middleware with a non-admin user — verify 403 on unauthorized endpoints
- [ ] 12.2 Backend: Test super admin bypass — verify all endpoints accessible
- [ ] 12.3 Backend: Test user-role assignment and removal APIs
- [ ] 12.4 Backend: Test getUserPermissions returns correct codes and menu tree
- [ ] 12.5 Frontend: Verify sidebar menu filters based on logged-in user's permissions
- [ ] 12.6 Frontend: Verify v-permission directive hides buttons for unauthorized users
- [ ] 12.7 Frontend: Verify user detail page shows roles and supports assignment/removal
- [ ] 12.8 Frontend: Verify route guard redirects to 403 for unauthorized page access
