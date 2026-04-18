-- Identity Service Seed Data
-- Database: identity_db
-- Description: Initial data including super admin, default roles, and permissions

USE identity_db;

-- ============================================================
-- Seed Data: auth_permission
-- Description: Default permission tree (menu and button level)
-- ============================================================

-- Root Menu: System Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(1, NULL, '系统管理', 'system', 1, '/system', 'setting', 1, 1);

-- User Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(10, 1, '用户管理', 'user', 1, '/system/user', 'user', 10, 1),
(11, 10, '查看用户', 'user:view', 2, NULL, NULL, 1, 1),
(12, 10, '创建用户', 'user:create', 2, NULL, NULL, 2, 1),
(13, 10, '编辑用户', 'user:update', 2, NULL, NULL, 3, 1),
(14, 10, '删除用户', 'user:delete', 2, NULL, NULL, 4, 1);

-- Role Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(20, 1, '角色管理', 'role', 1, '/system/role', 'team', 20, 1),
(21, 20, '查看角色', 'role:view', 2, NULL, NULL, 1, 1),
(22, 20, '创建角色', 'role:create', 2, NULL, NULL, 2, 1),
(23, 20, '编辑角色', 'role:update', 2, NULL, NULL, 3, 1),
(24, 20, '删除角色', 'role:delete', 2, NULL, NULL, 4, 1),
(25, 20, '分配权限', 'role:assign_permission', 2, NULL, NULL, 5, 1);

-- Root Menu: Master Data Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(100, NULL, '主数据管理', 'masterdata', 1, '/masterdata', 'database', 2, 1);

-- Administrative Division Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(110, 100, '行政区划管理', 'division', 1, '/masterdata/division', 'environment', 10, 1),
(111, 110, '查看区划', 'division:view', 2, NULL, NULL, 1, 1),
(112, 110, '创建区划', 'division:create', 2, NULL, NULL, 2, 1),
(113, 110, '编辑区划', 'division:update', 2, NULL, NULL, 3, 1),
(114, 110, '删除区划', 'division:delete', 2, NULL, NULL, 4, 1);

-- Community Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(120, 100, '社区管理', 'community', 1, '/masterdata/community', 'home', 20, 1),
(121, 120, '查看社区', 'community:view', 2, NULL, NULL, 1, 1),
(122, 120, '创建社区', 'community:create', 2, NULL, NULL, 2, 1),
(123, 120, '编辑社区', 'community:update', 2, NULL, NULL, 3, 1),
(124, 120, '提交审核', 'community:submit', 2, NULL, NULL, 4, 1),
(125, 120, '审核社区', 'community:review', 2, NULL, NULL, 5, 1),
(126, 120, '删除社区', 'community:delete', 2, NULL, NULL, 6, 1);

-- Root Menu: Property Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(200, NULL, '房产管理', 'property', 1, '/property', 'shop', 3, 1);

-- Property Unit Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(210, 200, '房产单元管理', 'property_unit', 1, '/property/unit', 'apartment', 10, 1),
(211, 210, '查看房产', 'property_unit:view', 2, NULL, NULL, 1, 1),
(212, 210, '创建房产', 'property_unit:create', 2, NULL, NULL, 2, 1),
(213, 210, '编辑房产', 'property_unit:update', 2, NULL, NULL, 3, 1),
(214, 210, '删除房产', 'property_unit:delete', 2, NULL, NULL, 4, 1);

-- Homeowner Verification
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(220, 200, '业主认证', 'homeowner_verification', 1, '/property/verification', 'safety-certificate', 20, 1),
(221, 220, '查看认证', 'homeowner_verification:view', 2, NULL, NULL, 1, 1),
(222, 220, '提交认证', 'homeowner_verification:submit', 2, NULL, NULL, 2, 1),
(223, 220, '审核认证', 'homeowner_verification:review', 2, NULL, NULL, 3, 1);

-- Property Binding
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(230, 200, '房产绑定', 'property_binding', 1, '/property/binding', 'link', 30, 1),
(231, 230, '查看绑定', 'property_binding:view', 2, NULL, NULL, 1, 1),
(232, 230, '创建绑定', 'property_binding:create', 2, NULL, NULL, 2, 1),
(233, 230, '删除绑定', 'property_binding:delete', 2, NULL, NULL, 3, 1);

-- Root Menu: Family Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(300, NULL, '家庭管理', 'family', 1, '/family', 'team', 4, 1);

-- Family Center
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(310, 300, '家庭中心', 'family_center', 1, '/family/center', 'home', 10, 1),
(311, 310, '查看家庭', 'family:view', 2, NULL, NULL, 1, 1),
(312, 310, '创建家庭', 'family:create', 2, NULL, NULL, 2, 1),
(313, 310, '编辑家庭', 'family:update', 2, NULL, NULL, 3, 1),
(314, 310, '添加成员', 'family:add_member', 2, NULL, NULL, 4, 1),
(315, 310, '编辑成员', 'family:update_member', 2, NULL, NULL, 5, 1),
(316, 310, '删除成员', 'family:delete_member', 2, NULL, NULL, 6, 1);

-- Root Menu: Configuration Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(400, NULL, '配置管理', 'configuration', 1, '/configuration', 'control', 5, 1);

-- Platform Configuration
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(410, 400, '平台配置', 'platform_config', 1, '/configuration/platform', 'setting', 10, 1),
(411, 410, '查看配置', 'config:view', 2, NULL, NULL, 1, 1),
(412, 410, '创建配置', 'config:create', 2, NULL, NULL, 2, 1),
(413, 410, '编辑配置', 'config:update', 2, NULL, NULL, 3, 1),
(414, 410, '审批配置', 'config:approve', 2, NULL, NULL, 4, 1);

-- Sensitive Word Management
INSERT INTO auth_permission (id, parent_id, name, code, type, path, icon, sort_order, status) VALUES
(420, 400, '敏感词管理', 'sensitive_word', 1, '/configuration/sensitive', 'warning', 20, 1),
(421, 420, '查看敏感词', 'sensitive_word:view', 2, NULL, NULL, 1, 1),
(422, 420, '创建敏感词', 'sensitive_word:create', 2, NULL, NULL, 2, 1),
(423, 420, '编辑敏感词', 'sensitive_word:update', 2, NULL, NULL, 3, 1),
(424, 420, '删除敏感词', 'sensitive_word:delete', 2, NULL, NULL, 4, 1);

-- ============================================================
-- Seed Data: auth_role
-- Description: Default system roles
-- ============================================================

INSERT INTO auth_role (id, name, code, description, is_system, sort_order, status, created_by) VALUES
(1, '超级管理员', 'SUPER_ADMIN', '系统超级管理员，拥有所有权限', 1, 1, 1, 1),
(2, '总部管理员', 'HQ_ADMIN', '总部管理员，管理全国数据', 1, 2, 1, 1),
(3, '省级管理员', 'PROVINCE_ADMIN', '省级管理员，管理省级数据', 1, 3, 1, 1),
(4, '市级管理员', 'CITY_ADMIN', '市级管理员，管理市级数据', 1, 4, 1, 1),
(5, '业主', 'HOMEOWNER', '已认证业主', 1, 5, 1, 1);

-- ============================================================
-- Seed Data: auth_role_permission
-- Description: Assign all permissions to SUPER_ADMIN
-- ============================================================

-- Super Admin gets all permissions
INSERT INTO auth_role_permission (role_id, permission_id)
SELECT 1, id FROM auth_permission;

-- HQ Admin gets all permissions except user:delete
INSERT INTO auth_role_permission (role_id, permission_id)
SELECT 2, id FROM auth_permission WHERE code != 'user:delete';

-- Province Admin gets masterdata and community permissions
INSERT INTO auth_role_permission (role_id, permission_id)
SELECT 3, id FROM auth_permission WHERE code IN (
    'masterdata', 'community', 'community:view', 'community:create', 
    'community:update', 'community:submit'
);

-- City Admin gets community view and create permissions
INSERT INTO auth_role_permission (role_id, permission_id)
SELECT 4, id FROM auth_permission WHERE code IN (
    'masterdata', 'community', 'community:view', 'community:create', 
    'community:update', 'community:submit'
);

-- Homeowner gets property and family permissions
INSERT INTO auth_role_permission (role_id, permission_id)
SELECT 5, id FROM auth_permission WHERE code IN (
    'property', 'homeowner_verification', 'homeowner_verification:view', 
    'homeowner_verification:submit', 'property_binding', 'property_binding:view',
    'family', 'family_center', 'family:view', 'family:create', 'family:update',
    'family:add_member', 'family:update_member', 'family:delete_member'
);

-- ============================================================
-- Seed Data: auth_user
-- Description: Super admin user (default password: Admin@123456)
-- Password hash generated with: bcrypt.GenerateFromPassword([]byte("Admin@123456"), 10)
-- ============================================================

INSERT INTO auth_user (id, phone, password_hash, nickname, user_type, status, verification_status, scope_id) VALUES
(1, '13800000000', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '超级管理员', 1, 1, 1, NULL);

-- ============================================================
-- Seed Data: auth_user_role
-- Description: Assign SUPER_ADMIN role to super admin user
-- ============================================================

INSERT INTO auth_user_role (user_id, role_id) VALUES (1, 1);
