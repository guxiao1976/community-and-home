## ADDED Requirements

### Requirement: Route-level permission guard
The router navigation guard SHALL check `to.meta.permission` before allowing navigation to protected routes. If the user lacks the required permission code, the guard SHALL redirect to the 403 error page. Routes without `meta.permission` SHALL be accessible to all authenticated users.

#### Scenario: User with permission accesses protected route
- **WHEN** a user with permission "user:list" navigates to `/users/list` which has `meta.permission = "user:list"`
- **THEN** the navigation proceeds normally

#### Scenario: User without permission is redirected to 403
- **WHEN** a user without permission "user:list" navigates to `/users/list`
- **THEN** the navigation is blocked and the user is redirected to the 403 error page

#### Scenario: Route without permission meta is accessible
- **WHEN** a user navigates to `/dashboard` which has no `meta.permission` defined
- **THEN** the navigation proceeds normally for any authenticated user

### Requirement: Dynamic sidebar menu filtering
The sidebar menu component SHALL filter menu items based on the current user's permissions. Each menu item SHALL have a `permissionCode` property. Menu items whose `permissionCode` is not in the user's permission list SHALL not be rendered. Parent menu items with no visible children SHALL also be hidden. The permission store SHALL be loaded on app initialization after authentication.

#### Scenario: Menu items visible based on user permissions
- **WHEN** a user with permissions `["user:list", "user:create", "dashboard:view"]` loads the app
- **THEN** the sidebar shows the "仪表板" and "用户管理 > 用户列表" menu items, but hides "主数据管理" if its permission code is not in the list

#### Scenario: Parent menu hidden when all children are hidden
- **WHEN** a user has no permissions for any items under "主数据管理"
- **THEN** the entire "主数据管理" sub-menu is not rendered in the sidebar

#### Scenario: Permissions loaded on app init
- **WHEN** the app initializes and the user is authenticated
- **THEN** the permission store loads the user's permissions before rendering the layout

### Requirement: v-permission directive for button-level control
The system SHALL provide a `v-permission` custom directive that accepts a permission code string or array of permission codes. When the current user does not have the specified permission(s), the directive SHALL hide the element by setting `display: none`. When permission codes are provided as an array, the element SHALL be visible if the user has ANY of the specified permissions (OR logic).

#### Scenario: Button visible with matching permission
- **WHEN** a user with permission "user:create" views a page with `<el-button v-permission="'user:create'">新建用户</el-button>`
- **THEN** the button is visible

#### Scenario: Button hidden without matching permission
- **WHEN** a user without permission "user:create" views a page with `<el-button v-permission="'user:create'">新建用户</el-button>`
- **THEN** the button is hidden (display: none)

#### Scenario: Button visible with array of permissions (OR logic)
- **WHEN** a user with permission "user:create" views a page with `<el-button v-permission="['user:create', 'user:delete']">操作</el-button>`
- **THEN** the button is visible because the user has at least one of the specified permissions

#### Scenario: Button hidden with array where no permission matches
- **WHEN** a user with only permission "user:list" views a page with `<el-button v-permission="['user:create', 'user:delete']">操作</el-button>`
- **THEN** the button is hidden because the user has none of the specified permissions

### Requirement: User detail page role and permission display
The user detail page SHALL display the user's assigned roles in a table showing role name, code, description, and system flag. The page SHALL provide a button to assign new roles via a dialog that lists available roles. Each assigned role SHALL have a remove button (disabled for system roles). The page SHALL also display the user's effective permissions in a tree structure.

#### Scenario: User detail shows assigned roles
- **WHEN** admin views user detail page for user 5 who has roles "community_manager" and "reviewer"
- **THEN** the page displays a "角色与权限" section with a table listing both roles with their name, code, description, and system flag

#### Scenario: Admin assigns role to user from detail page
- **WHEN** admin clicks "分配角色" and selects "property_manager" from the role picker dialog
- **THEN** the role is assigned to the user, the role table updates, and a success message is shown

#### Scenario: Admin removes custom role from user
- **WHEN** admin clicks remove on a non-system role assignment
- **THEN** the role is removed from the user, the role table updates, and a success message is shown

#### Scenario: System role cannot be removed from user
- **WHEN** admin views the remove button for a system role assignment
- **THEN** the remove button is disabled

#### Scenario: User detail shows effective permissions tree
- **WHEN** admin views user detail page and clicks "查看权限"
- **THEN** the page displays all effective permissions organized as a tree with menu and button permissions

### Requirement: Permission store initialization flow
The permission store SHALL be initialized during app startup after successful authentication. The store SHALL load both the full permission tree (for role management pages) and the current user's permissions (for menu and button filtering). The store SHALL expose a `hasPermission(code)` method for use by the v-permission directive and other components.

#### Scenario: Permissions loaded after login
- **WHEN** user successfully logs in
- **THEN** the permission store loads the user's permissions and the sidebar menu re-renders with filtered items

#### Scenario: Permissions cleared on logout
- **WHEN** user logs out
- **THEN** the permission store clears all cached permissions and the sidebar menu becomes empty
