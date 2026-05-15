-- 为超级管理员角色分配新的用户管理权限
-- 执行前请确保已经运行过 identity_seed.sql 添加了新权限

USE identity_db;

-- 管理员用户管理权限
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(154, 1, '管理员管理', 'identity:admin-user', 1, '/users/admin', 'user-filled', 15, 1),
(155, 154, '查看管理员', 'identity:admin-user:list', 2, NULL, NULL, 1, 1),
(156, 154, '创建管理员', 'identity:admin-user:create', 2, NULL, NULL, 2, 1),
(157, 154, '编辑管理员', 'identity:admin-user:update', 2, NULL, NULL, 3, 1),
(158, 154, '删除管理员', 'identity:admin-user:delete', 2, NULL, NULL, 4, 1),
(159, 154, '分配角色', 'identity:admin-user:assign-role', 2, NULL, NULL, 5, 1);

-- 普通用户管理权限
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(160, 1, '普通用户管理', 'identity:regular-user', 1, '/users/regular', 'user', 16, 1),
(161, 160, '查看普通用户', 'identity:regular-user:list', 2, NULL, NULL, 1, 1),
(162, 160, '创建普通用户', 'identity:regular-user:create', 2, NULL, NULL, 2, 1),
(163, 160, '编辑普通用户', 'identity:regular-user:update', 2, NULL, NULL, 3, 1),
(164, 160, '删除普通用户', 'identity:regular-user:delete', 2, NULL, NULL, 4, 1);

-- 给超级管理员角色分配新权限
INSERT IGNORE INTO auth_role_permission (role_id, permission_id) VALUES
(1, 154), (1, 155), (1, 156), (1, 157), (1, 158), (1, 159),
(1, 160), (1, 161), (1, 162), (1, 163), (1, 164);

-- 验证权限分配
SELECT
    r.name AS role_name,
    p.name AS permission_name,
    p.code AS permission_code
FROM auth_role_permission rp
JOIN auth_role r ON rp.role_id = r.id
JOIN auth_permission p ON rp.permission_id = p.id
WHERE r.id = 1
AND p.id >= 154
ORDER BY p.id;

SELECT '权限分配完成！请重新登录系统以使权限生效。' AS message;
