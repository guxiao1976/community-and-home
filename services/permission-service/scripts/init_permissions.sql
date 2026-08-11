-- ============================================================================
-- 权限系统初始化脚本
-- 用途：初始化 4 个系统角色 + 权限树（菜单+按钮+API）+ 角色-权限精确关联
-- 执行：mysql -h<host> -u<user> -p<password> <database> < init_permissions.sql
--
-- is_system 含义变更 (2026-08):
--   改前：is_system=1 → 不可删除 + 自动获得所有权限（代码短路）
--   改后：is_system=1 → 不可删除 + 不可修改（权限由 rel_role_permission 配置决定）
--   如需超级管理员，通过管理后台新建角色并勾选全部权限即可
-- ============================================================================

-- 1. 初始化 8 个系统角色（如果不存在）
-- is_system=1 表示系统角色，仅用于保护不被误删和误改
INSERT IGNORE INTO sys_role (id, role_code, role_name, description, is_system, status, sort_order, created_by, created_at, updated_at)
VALUES
(1, 'owner', '业主', '小区业主，查看自己单元相关信息的权限', 1, 1, 10, 0, NOW(), NOW()),
(2, 'property_admin', '物业管理员', '物业公司管理人员，负责小区日常管理和通知发布', 1, 1, 20, 0, NOW(), NOW()),
(3, 'community_admin', '社区管理员', '社区居委会管理人员，拥有最全面的管理权限', 1, 1, 30, 0, NOW(), NOW()),
(4, 'grid_worker', '网格员', '社区网格员，负责网格内的巡查和信息查看', 1, 1, 40, 0, NOW(), NOW()),
(5, 'tenant', '租户', '小区租户，查看租约房屋相关信息的权限', 1, 1, 50, 0, NOW(), NOW()),
(6, 'committee', '业委会', '小区业委会成员，参与小区公共事务管理', 1, 1, 60, 0, NOW(), NOW()),
(7, 'merchant', '商家', '入驻商家，管理自身商业信息', 1, 1, 70, 0, NOW(), NOW()),
(8, 'sys_admin', '系统管理员', '系统管理员，拥有全部管理权限', 1, 1, 5, 0, NOW(), NOW());

-- 2. 初始化权限树（菜单 → 按钮 → API 三层）
-- type: 1=菜单(menu), 2=按钮(button), 3=API(api)
-- parent_id: 0 表示根节点
-- path: type=3 时存储实际的 REST API 路径前缀（/api/perm/, /api/users/ 等），用于 CheckPermission needle 匹配

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

-- API 层（path 修正为实际 REST API 路径）
(111, 110, 'GET /api/users', 'user:read:list-api', 3, 'GET:/api/users', NULL, 10, 1, NOW(), NOW()),
(112, 110, 'GET /api/users/:id', 'user:read:detail-api', 3, 'GET:/api/users/:id', NULL, 20, 1, NOW(), NOW()),
(121, 120, 'POST /api/users', 'user:create:api', 3, 'POST:/api/users', NULL, 10, 1, NOW(), NOW()),
(131, 130, 'PUT /api/users/:id', 'user:update:api', 3, 'PUT:/api/users/:id', NULL, 10, 1, NOW(), NOW()),
(141, 140, 'DELETE /api/users/:id', 'user:delete:api', 3, 'DELETE:/api/users/:id', NULL, 10, 1, NOW(), NOW()),
(151, 150, 'POST /api/perm/user-roles', 'user:assign-role:api', 3, 'POST:/api/perm/user-roles', NULL, 10, 1, NOW(), NOW());

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

-- API 层（path 修正为 /api/perm/ 前缀）
(211, 210, 'GET /api/perm/roles', 'role:read:list-api', 3, 'GET:/api/perm/roles', NULL, 10, 1, NOW(), NOW()),
(212, 210, 'GET /api/perm/roles/:id', 'role:read:detail-api', 3, 'GET:/api/perm/roles/:id', NULL, 20, 1, NOW(), NOW()),
(221, 220, 'POST /api/perm/roles', 'role:create:api', 3, 'POST:/api/perm/roles', NULL, 10, 1, NOW(), NOW()),
(231, 230, 'PUT /api/perm/roles/:id', 'role:update:api', 3, 'PUT:/api/perm/roles/:id', NULL, 10, 1, NOW(), NOW()),
(241, 240, 'DELETE /api/perm/roles/:id', 'role:delete:api', 3, 'DELETE:/api/perm/roles/:id', NULL, 10, 1, NOW(), NOW()),
(251, 250, 'GET /api/perm/permissions', 'role:permission:list-api', 3, 'GET:/api/perm/permissions', NULL, 10, 1, NOW(), NOW()),
(252, 250, 'POST /api/perm/roles/:id/permissions', 'role:permission:update-api', 3, 'POST:/api/perm/roles/:id/permissions', NULL, 20, 1, NOW(), NOW());

-- ============================================================================
-- 权限管理模块
-- ============================================================================
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
-- 菜单层
(300, 0, '权限管理', 'permission:menu', 1, '/permissions', 'permission', 30, 1, NOW(), NOW()),

-- 按钮层
(310, 300, '查看权限', 'permission:read', 2, NULL, NULL, 10, 1, NOW(), NOW()),

-- API 层（path 修正为 /api/perm/ 前缀）
(311, 310, 'GET /api/perm/permissions', 'permission:read:list-api', 3, 'GET:/api/perm/permissions', NULL, 10, 1, NOW(), NOW());

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

-- API 层（path 修正，前缀根据实际服务路由）
(411, 410, 'GET /api/community/communities', 'community:read:list-api', 3, 'GET:/api/community/communities', NULL, 10, 1, NOW(), NOW()),
(421, 420, 'POST /api/community/notices', 'community:notice:create-api', 3, 'POST:/api/community/notices', NULL, 10, 1, NOW(), NOW()),
(431, 430, 'GET /api/community/activities', 'community:activity:list-api', 3, 'GET:/api/community/activities', NULL, 10, 1, NOW(), NOW()),
(432, 430, 'POST /api/community/activities', 'community:activity:create-api', 3, 'POST:/api/community/activities', NULL, 20, 1, NOW(), NOW());

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

-- API 层（path 修正，前缀根据实际 moderation-service 路由）
(511, 510, 'GET /api/moderation/reviews', 'moderation:read:list-api', 3, 'GET:/api/moderation/reviews', NULL, 10, 1, NOW(), NOW()),
(521, 520, 'POST /api/moderation/reviews/:id/approve', 'moderation:review:approve-api', 3, 'POST:/api/moderation/reviews/:id/approve', NULL, 10, 1, NOW(), NOW()),
(522, 520, 'POST /api/moderation/reviews/:id/reject', 'moderation:review:reject-api', 3, 'POST:/api/moderation/reviews/:id/reject', NULL, 20, 1, NOW(), NOW());

-- 3. 按角色精确赋权（替代原来的 CROSS JOIN 一刀切）
-- 设计原则：
--   - 业主(owner)：只能查看自己相关的信息
--   - 物业管理员(property_admin)：用户查看 + 社区管理 + 审核查看
--   - 社区管理员(community_admin)：除角色管理外的全面管理权限
--   - 网格员(grid_worker)：最小查看权限
--   - 如需超级管理员，通过管理后台新建第 5 个自定义角色并全选权限

-- 业主(role_id=1)：用户查看 + 社区查看
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(1, 100), (1, 110), (1, 111), (1, 112),  -- 用户管理菜单+查看
(1, 400), (1, 410), (1, 411);              -- 社区管理菜单+查看

-- 物业管理员(role_id=2)：用户管理(查看) + 社区管理(查看,通知) + 审核(查看)
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(2, 100), (2, 110), (2, 111), (2, 112),     -- 用户管理
(2, 400), (2, 410), (2, 411),                -- 社区查看
(2, 420), (2, 421),                           -- 发布通知
(2, 500), (2, 510), (2, 511);                -- 审核查看

-- 社区管理员(role_id=3)：用户+角色(查看)+社区(完整)+审核(完整)+活动管理
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(3, 100), (3, 110), (3, 111), (3, 112),      -- 用户查看
(3, 200), (3, 210), (3, 211), (3, 212),      -- 角色查看
(3, 400), (3, 410), (3, 411),                 -- 社区查看
(3, 420), (3, 421),                            -- 发布通知
(3, 430), (3, 431), (3, 432),                 -- 活动管理
(3, 500), (3, 510), (3, 511),                 -- 审核查看
(3, 520), (3, 521), (3, 522);                 -- 审核操作

-- 网格员(role_id=4)：最小权限 — 用户查看 + 社区查看
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(4, 100), (4, 110), (4, 111), (4, 112),      -- 用户查看
(4, 400), (4, 410), (4, 411);                 -- 社区查看

-- 租户(role_id=5)：用户查看 + 社区查看（与业主类似，无角色/权限管理）
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(5, 100), (5, 110), (5, 111), (5, 112),      -- 用户查看
(5, 400), (5, 410), (5, 411);                 -- 社区查看

-- 业委会(role_id=6)：用户查看 + 社区查看 + 发布通知
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(6, 100), (6, 110), (6, 111), (6, 112),      -- 用户查看
(6, 400), (6, 410), (6, 411),                 -- 社区查看
(6, 420), (6, 421);                           -- 发布通知

-- 商家(role_id=7)：用户查看 + 社区查看
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(7, 100), (7, 110), (7, 111), (7, 112),      -- 用户查看
(7, 400), (7, 410), (7, 411);                 -- 社区查看

-- 系统管理员(role_id=8)：全部权限（选择所有 permission_id）
INSERT IGNORE INTO rel_role_permission (role_id, permission_id)
SELECT 8, id FROM sys_permission WHERE status = 1;

-- ============================================================================
-- 数据验证查询
-- ============================================================================

-- 验证角色数量（应该至少有 8 个系统角色）
SELECT '角色数量检查' AS check_type, COUNT(*) AS count,
       CASE WHEN COUNT(*) >= 8 THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM sys_role
WHERE is_system = 1;

-- 验证每个系统角色的权限数量（精确赋权后，不再是全权限）
SELECT '角色权限分布' AS check_type, r.role_code, r.role_name, COUNT(rp.permission_id) AS perm_count,
       CASE
         WHEN r.role_code = 'owner' AND COUNT(rp.permission_id) >= 5 THEN '✅ PASS'
         WHEN r.role_code = 'property_admin' AND COUNT(rp.permission_id) >= 8 THEN '✅ PASS'
         WHEN r.role_code = 'community_admin' AND COUNT(rp.permission_id) >= 15 THEN '✅ PASS'
         WHEN r.role_code = 'grid_worker' AND COUNT(rp.permission_id) >= 5 THEN '✅ PASS'
         WHEN r.role_code = 'sys_admin' AND COUNT(rp.permission_id) >= 20 THEN '✅ PASS'
         ELSE '⚠️ REVIEW'
       END AS status
FROM sys_role r
LEFT JOIN rel_role_permission rp ON r.id = rp.role_id
WHERE r.is_system = 1
GROUP BY r.id, r.role_code, r.role_name
ORDER BY r.sort_order;

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

-- 验证 API path 使用正确前缀（不应再有 /api/v1/ 遗留）
SELECT 'API Path 检查' AS check_type, COUNT(*) AS legacy_count,
       CASE WHEN COUNT(*) = 0 THEN '✅ PASS' ELSE '❌ FAIL (残留 /api/v1/ 前缀)' END AS status
FROM sys_permission
WHERE type = 3 AND path LIKE '%/api/v1/%';

-- 初始化完成
SELECT '✅ 权限系统初始化完成 (v2: 配置驱动, is_system 无特权)' AS message;
