-- ============================================================================
-- 权限系统初始化脚本
-- 用途：初始化 4 个系统角色 + 权限树（菜单+按钮+API）+ 角色-权限关联
-- 执行：mysql -h<host> -u<user> -p<password> <database> < init_permissions.sql
-- ============================================================================

-- 1. 初始化 4 个系统角色（如果不存在）
-- is_system=1 表示系统角色，不可删除，天然拥有所有权限
INSERT IGNORE INTO sys_role (id, role_code, role_name, description, is_system, status, sort_order, created_at, updated_at)
VALUES
(1, 'owner', '业主', '小区业主，拥有查看和管理自己单元的权限', 1, 1, 10, NOW(), NOW()),
(2, 'property_admin', '物业管理员', '物业公司管理人员，负责小区日常管理', 1, 1, 20, NOW(), NOW()),
(3, 'community_admin', '社区管理员', '社区居委会管理人员，负责社区事务', 1, 1, 30, NOW(), NOW()),
(4, 'grid_worker', '网格员', '社区网格员，负责网格内的巡查和服务', 1, 1, 40, NOW(), NOW());

-- 2. 初始化权限树（菜单 → 按钮 → API 三层）
-- type: 1=菜单(menu), 2=按钮(button), 3=API(api)
-- parent_id: 0 表示根节点

-- ============================================================================
-- 用户管理模块
-- ============================================================================
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
-- 菜单层
(100, 0, '用户管理', 'user:menu', 1, '/users', 'user', 10, 1, NOW(), NOW()),

-- 按钮层
(110, 100, '查看用户', 'user:read', 2, NULL, NULL, 10, 1, NOW(), NOW()),
(120, 100, '创建用户', 'user:create', 2, NULL, NULL, 20, 1, NOW(), NOW()),
(130, 100, '编辑用户', 'user:update', 2, NULL, NULL, 30, 1, NOW(), NOW()),
(140, 100, '删除用户', 'user:delete', 2, NULL, NULL, 40, 1, NOW(), NOW()),
(150, 100, '分配角色', 'user:assign-role', 2, NULL, NULL, 50, 1, NOW(), NOW()),

-- API 层
(111, 110, 'GET /api/v1/user/users', 'user:read:list-api', 3, 'GET:/api/v1/user/users', NULL, 10, 1, NOW(), NOW()),
(112, 110, 'GET /api/v1/user/users/:id', 'user:read:detail-api', 3, 'GET:/api/v1/user/users/:id', NULL, 20, 1, NOW(), NOW()),
(121, 120, 'POST /api/v1/user/users', 'user:create:api', 3, 'POST:/api/v1/user/users', NULL, 10, 1, NOW(), NOW()),
(131, 130, 'PUT /api/v1/user/users/:id', 'user:update:api', 3, 'PUT:/api/v1/user/users/:id', NULL, 10, 1, NOW(), NOW()),
(141, 140, 'DELETE /api/v1/user/users/:id', 'user:delete:api', 3, 'DELETE:/api/v1/user/users/:id', NULL, 10, 1, NOW(), NOW()),
(151, 150, 'POST /api/v1/permission/assign-role', 'user:assign-role:api', 3, 'POST:/api/v1/permission/assign-role', NULL, 10, 1, NOW(), NOW());

-- ============================================================================
-- 角色管理模块
-- ============================================================================
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
-- 菜单层
(200, 0, '角色管理', 'role:menu', 1, '/roles', 'role', 20, 1, NOW(), NOW()),

-- 按钮层
(210, 200, '查看角色', 'role:read', 2, NULL, NULL, 10, 1, NOW(), NOW()),
(220, 200, '创建角色', 'role:create', 2, NULL, NULL, 20, 1, NOW(), NOW()),
(230, 200, '编辑角色', 'role:update', 2, NULL, NULL, 30, 1, NOW(), NOW()),
(240, 200, '删除角色', 'role:delete', 2, NULL, NULL, 40, 1, NOW(), NOW()),
(250, 200, '配置权限', 'role:permission', 2, NULL, NULL, 50, 1, NOW(), NOW()),

-- API 层
(211, 210, 'GET /api/v1/permission/roles', 'role:read:list-api', 3, 'GET:/api/v1/permission/roles', NULL, 10, 1, NOW(), NOW()),
(212, 210, 'GET /api/v1/permission/roles/:id', 'role:read:detail-api', 3, 'GET:/api/v1/permission/roles/:id', NULL, 20, 1, NOW(), NOW()),
(221, 220, 'POST /api/v1/permission/roles', 'role:create:api', 3, 'POST:/api/v1/permission/roles', NULL, 10, 1, NOW(), NOW()),
(231, 230, 'PUT /api/v1/permission/roles/:id', 'role:update:api', 3, 'PUT:/api/v1/permission/roles/:id', NULL, 10, 1, NOW(), NOW()),
(241, 240, 'DELETE /api/v1/permission/roles/:id', 'role:delete:api', 3, 'DELETE:/api/v1/permission/roles/:id', NULL, 10, 1, NOW(), NOW()),
(251, 250, 'GET /api/v1/permission/permissions', 'role:permission:list-api', 3, 'GET:/api/v1/permission/permissions', NULL, 10, 1, NOW(), NOW()),
(252, 250, 'PUT /api/v1/permission/roles/:id/permissions', 'role:permission:update-api', 3, 'PUT:/api/v1/permission/roles/:id/permissions', NULL, 20, 1, NOW(), NOW());

-- ============================================================================
-- 权限管理模块
-- ============================================================================
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
-- 菜单层
(300, 0, '权限管理', 'permission:menu', 1, '/permissions', 'permission', 30, 1, NOW(), NOW()),

-- 按钮层
(310, 300, '查看权限', 'permission:read', 2, NULL, NULL, 10, 1, NOW(), NOW()),

-- API 层
(311, 310, 'GET /api/v1/permission/permissions', 'permission:read:list-api', 3, 'GET:/api/v1/permission/permissions', NULL, 10, 1, NOW(), NOW());

-- ============================================================================
-- 社区管理模块
-- ============================================================================
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
-- 菜单层
(400, 0, '社区管理', 'community:menu', 1, '/community', 'community', 40, 1, NOW(), NOW()),

-- 按钮层
(410, 400, '查看社区', 'community:read', 2, NULL, NULL, 10, 1, NOW(), NOW()),
(420, 400, '发布通知', 'community:notice', 2, NULL, NULL, 20, 1, NOW(), NOW()),
(430, 400, '活动管理', 'community:activity', 2, NULL, NULL, 30, 1, NOW(), NOW()),

-- API 层
(411, 410, 'GET /api/v1/community/communities', 'community:read:list-api', 3, 'GET:/api/v1/community/communities', NULL, 10, 1, NOW(), NOW()),
(421, 420, 'POST /api/v1/community/notices', 'community:notice:create-api', 3, 'POST:/api/v1/community/notices', NULL, 10, 1, NOW(), NOW()),
(431, 430, 'GET /api/v1/community/activities', 'community:activity:list-api', 3, 'GET:/api/v1/community/activities', NULL, 10, 1, NOW(), NOW()),
(432, 430, 'POST /api/v1/community/activities', 'community:activity:create-api', 3, 'POST:/api/v1/community/activities', NULL, 20, 1, NOW(), NOW());

-- ============================================================================
-- 审核管理模块
-- ============================================================================
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
-- 菜单层
(500, 0, '审核管理', 'moderation:menu', 1, '/moderation', 'moderation', 50, 1, NOW(), NOW()),

-- 按钮层
(510, 500, '查看审核', 'moderation:read', 2, NULL, NULL, 10, 1, NOW(), NOW()),
(520, 500, '审核操作', 'moderation:review', 2, NULL, NULL, 20, 1, NOW(), NOW()),

-- API 层
(511, 510, 'GET /api/v1/moderation/reviews', 'moderation:read:list-api', 3, 'GET:/api/v1/moderation/reviews', NULL, 10, 1, NOW(), NOW()),
(521, 520, 'POST /api/v1/moderation/reviews/:id/approve', 'moderation:review:approve-api', 3, 'POST:/api/v1/moderation/reviews/:id/approve', NULL, 10, 1, NOW(), NOW()),
(522, 520, 'POST /api/v1/moderation/reviews/:id/reject', 'moderation:review:reject-api', 3, 'POST:/api/v1/moderation/reviews/:id/reject', NULL, 20, 1, NOW(), NOW());

-- 3. 系统角色自动关联所有权限
-- 系统角色（is_system=1）天然拥有所有权限
INSERT IGNORE INTO rel_role_permission (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM sys_role r
CROSS JOIN sys_permission p
WHERE r.is_system = 1
  AND p.status = 1;

-- ============================================================================
-- 数据验证查询
-- ============================================================================

-- 验证角色数量（应该至少有 4 个系统角色）
SELECT '角色数量检查' AS check_type, COUNT(*) AS count,
       CASE WHEN COUNT(*) >= 4 THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM sys_role
WHERE is_system = 1;

-- 验证权限数量（应该有多个权限）
SELECT '权限数量检查' AS check_type, COUNT(*) AS count,
       CASE WHEN COUNT(*) > 0 THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM sys_permission;

-- 验证权限树结构（菜单、按钮、API 三层）
SELECT '权限结构检查' AS check_type,
       SUM(CASE WHEN type = 1 THEN 1 ELSE 0 END) AS menu_count,
       SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) AS button_count,
       SUM(CASE WHEN type = 3 THEN 1 ELSE 0 END) AS api_count,
       CASE WHEN SUM(CASE WHEN type = 1 THEN 1 ELSE 0 END) > 0
            AND SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) > 0
            AND SUM(CASE WHEN type = 3 THEN 1 ELSE 0 END) > 0
            THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM sys_permission;

-- 验证角色-权限关联（每个系统角色应该关联所有权限）
SELECT '角色权限关联检查' AS check_type, r.role_code, r.role_name, COUNT(rp.permission_id) AS permission_count,
       CASE WHEN COUNT(rp.permission_id) > 0 THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM sys_role r
LEFT JOIN rel_role_permission rp ON r.id = rp.role_id
WHERE r.is_system = 1
GROUP BY r.id, r.role_code, r.role_name;

-- 初始化完成
SELECT '✅ 权限系统初始化完成' AS message;
