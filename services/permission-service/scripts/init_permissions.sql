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
-- 4. 迁移段：数据权限核心 (access-data-permission 阶段① Wave1)
--    -- 目标：能力分层（min_verf_level）+ 注册用户基角色 + 系统审核身份
--    -- 幂等：guard + INSERT IGNORE，可重复执行
-- ============================================================================

-- 4.1 sys_permission.min_verf_level 列（MySQL 8.0 不支持 ADD COLUMN IF NOT EXISTS，用 guard）
-- SEE: [[migration-must-execute]]
SET @min_verf_col := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_permission' AND COLUMN_NAME = 'min_verf_level');
SET @min_verf_ddl := IF(@min_verf_col = 0,
  'ALTER TABLE sys_permission ADD COLUMN min_verf_level TINYINT NOT NULL DEFAULT 0 COMMENT ''能力层级: 0=持角色+数据范围即可, 2=需已认证(默认0)''',
  'SELECT ''min_verf_level column already exists''');
PREPARE min_verf_stmt FROM @min_verf_ddl;
EXECUTE min_verf_stmt;
DEALLOCATE PREPARE min_verf_stmt;

-- 4.2 发布类权限置 min_verf_level=0（默认即为 0，显式标注便于审计）
-- SEE: [[permission-seed-api-path-must-match-routes]] — path 必须与实际 REST 路由一致
UPDATE sys_permission SET min_verf_level = 0
WHERE code IN ('community:notice:create-api', 'community:lostfound:create-api');

-- 4.3 选举类权限（committee:election:vote）min_verf_level=2（需已认证）
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES (600, 0, '业委会选举投票', 'committee:election:vote', 2, NULL, NULL, 80, 1, NOW(), NOW());
UPDATE sys_permission SET min_verf_level = 2 WHERE code = 'committee:election:vote';

-- 4.3.1 敏感权限置 min_verf_level=2（需已认证）——security-arch 评审 CRITICAL
--   能力分层语义下 level-0=「持角色+数据范围即可」，未认证（status∈{0,1}）持角色即可访问；
--   既有敏感权限（user:read=全量用户PII、moderation:read/review=审核数据）必须需已认证，否则未认证用户可 GET /api/users 枚举全部用户。
--   SEE: [[is-system-no-permission-shortcut]] — 权限经 rel_role_permission 配置，认证要求经 min_verf_level 数据驱动
UPDATE sys_permission SET min_verf_level = 2
WHERE code IN ('user:read', 'user:read:list-api', 'user:read:detail-api',
               'moderation:read', 'moderation:read:list-api',
               'moderation:review', 'moderation:review:approve-api', 'moderation:review:reject-api');

-- 4.4 注册用户 browse 权限（registered_user → 读 only）
--    path 与实际 REST 路由一致（[[permission-seed-api-path-must-match-routes]]）
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
(422, 410, 'GET /api/community/notices', 'community:notice:read-list-api', 3, 'GET:/api/community/notices', NULL, 15, 1, NOW(), NOW()),
(433, 430, 'GET /api/community/lostfound', 'community:lostfound:read-list-api', 3, 'GET:/api/community/lostfound', NULL, 15, 1, NOW(), NOW()),
(434, 430, 'GET /api/community/contacts', 'community:contact:read-list-api', 3, 'GET:/api/community/contacts', NULL, 15, 1, NOW(), NOW());

-- 4.5 registered_user 基角色（id=9，is_system=1 仅保护，权限经 rel_role_permission 配置）
-- SEE: [[is-system-no-permission-shortcut]]
INSERT IGNORE INTO sys_role (id, role_code, role_name, description, is_system, status, sort_order, created_by, created_at, updated_at)
VALUES (9, 'registered_user', '注册用户', '注册即自动分配的基角色：browse-only、空数据范围、永久有效', 1, 1, 5, 0, NOW(), NOW());

-- 4.6 registered_user → browse 权限关联（仅读，无发布/选举）
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(9, 422), (9, 433), (9, 434);

-- 4.7 预留系统审核身份（moderation 回调）：sys_admin + global scope
--    scope_type='global' + scope_id=0 + status=2（verified_at=NULL，走 grant 判定路径）
-- SEE: [[is-system-no-permission-shortcut]] — 无代码级 userId==0 短路
INSERT IGNORE INTO rel_user_role (user_id, role_id, scope_type, scope_id, status) VALUES
(0, 8, 'global', 0, 2);

-- 4.8 业主/租户发布权限（access-data-permission 阶段⑥ 集成验收修复）
--   设计 §3.2/§3.4：未认证业主/租户即可发布（min_verf_level=0 = 持角色+数据范围即可）。
--   Wave 1 种子遗漏 owner/tenant 的发布类权限绑定，导致功能权限层（PermMiddleware）
--   先于数据范围检查拒绝，owner@A 发布/owner@A发B→080006 无法触达数据范围判定。
--   补齐 owner(1)/tenant(5)：发布通知(421) + 发布失物招领(435) + 联系方式维护(436)。
-- SEE: [[permission-seed-api-path-must-match-routes]] — path 与实际 REST 路由一致
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES (436, 430, 'POST /api/community/contacts', 'community:contact:upsert-api', 3, 'POST:/api/community/contacts', NULL, 15, 1, NOW(), NOW());
UPDATE sys_permission SET min_verf_level = 0 WHERE code = 'community:contact:upsert-api';

INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(1, 421), (1, 435), (1, 436),
(5, 421), (5, 435), (5, 436);

-- 4.8.1 业主/租户读列表权限（与 registered_user browse 对齐，保证「读列表按 scope 过滤」可测）
--   owner/tenant 为高于 registered_user 的社区角色，须能读所属小区内容（读列表 + scope 过滤）
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(1, 422), (1, 433), (1, 434),
(5, 422), (5, 433), (5, 434);

-- 4.8.2 选举权限（committee:election:vote, min_verf_level=2）绑定
--   sys_admin 的「全权限」绑定在 permission 600 创建之前执行，缺 600；committee(6) 为业委会角色须可投票。
--   集成验收 T6.1：未认证(status=0)用户选举❌，已认证(status=2)用户选举✅ 依赖此绑定。
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(6, 600), (8, 600);

-- ============================================================================
-- 5. 访问控制（access-control）：内置角色 platforms 种子 + 当前小区/应用状态接口权限码
-- ============================================================================

-- 5.1 内置角色 platforms 种子初值（design §4.1；运行时空值 fail-open）
--   sys_admin=pc、community_admin=pc,mobile、property_admin=pc、
--   owner/tenant/grid_worker/committee/merchant/registered_user=mobile
-- SEE: [[is-system-no-permission-shortcut]] — platforms 为配置属性，不参与权限短路
UPDATE sys_role SET platforms = 'pc'        WHERE role_code = 'sys_admin';
UPDATE sys_role SET platforms = 'pc,mobile' WHERE role_code = 'community_admin';
UPDATE sys_role SET platforms = 'pc'        WHERE role_code = 'property_admin';
UPDATE sys_role SET platforms = 'mobile'    WHERE role_code = 'owner';
UPDATE sys_role SET platforms = 'mobile'    WHERE role_code = 'tenant';
UPDATE sys_role SET platforms = 'mobile'    WHERE role_code = 'grid_worker';
UPDATE sys_role SET platforms = 'mobile'    WHERE role_code = 'committee';
UPDATE sys_role SET platforms = 'mobile'    WHERE role_code = 'merchant';
UPDATE sys_role SET platforms = 'mobile'    WHERE role_code = 'registered_user';

-- 5.2 当前小区 / 应用状态接口权限码（type=3）
--   path 必须与实际 REST 路由一致（user-service Task 3.7 注册的
--   GET /api/users/me/app-state、PUT /api/users/me/current-community）
-- SEE: [[permission-seed-api-path-must-match-routes]]
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
(700, 100, 'GET /api/users/me/app-state', 'user:appstate:read-api', 3, 'GET:/api/users/me/app-state', NULL, 60, 1, NOW(), NOW()),
(701, 100, 'PUT /api/users/me/current-community', 'user:currentcommunity:write-api', 3, 'PUT:/api/users/me/current-community', NULL, 70, 1, NOW(), NOW());

-- 5.3 挂载到 mobile 社区角色（registered_user/owner/tenant/grid_worker/committee/merchant）
--   sys_admin 的「全权限」绑定在 700/701 创建之前执行（见 3.），此处补挂保持全权限语义
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(9, 700), (9, 701),   -- registered_user（基角色，注册即分配）
(1, 700), (1, 701),   -- owner
(4, 700), (4, 701),   -- grid_worker
(5, 700), (5, 701),   -- tenant
(6, 700), (6, 701),   -- committee
(7, 700), (7, 701),   -- merchant
(8, 700), (8, 701);   -- sys_admin（全权限语义）

-- ============================================================================
-- 6. 通用图文发布权限种子（content-post-generalization，REQ-CPP-3 REVISION）
--    -- 目标：421 min_verf_level 0→2（需已认证）+ grid_worker 授 421 + owner/tenant 撤销 421
--    --       + 读码 422 扩展全部移动端角色 + 新增 423/424/426（读）+ 427/428（写）
--    -- 幂等：UPDATE + INSERT IGNORE + 幂等 DELETE，可重复执行
--    -- 「全部移动端角色」={owner1 / community_admin3 / grid_worker4 / tenant5 / committee6 /
--    --                     merchant7 / sys_admin8 / registered_user9}
--    -- property_admin(2) platforms='pc' 不绑移动端读码与 427/428，仅保留 421（PC 走后续接线，
--    --   创建后编辑/撤回由 080002 作者校验兜底——评审 SHOULD #4 不对称注明）
-- ============================================================================

-- 6.1 421（community:notice:create-api）min_verf_level 0→2（行为变更：需已认证，REVISION）
--     覆盖 4.2 的默认 min_verf_level=0（脚本自上而下，本段恒生效）；写路径角色状态门槛=level-2
--     （status==2 且 verified_at NOT NULL 且未过期，与 GetPublishPermission/community-hub 判定一致）
-- SEE: [[auto-grant-unverified-grant-confers-scope-level0]]
UPDATE sys_permission SET min_verf_level = 2 WHERE code = 'community:notice:create-api';

-- 6.2 撤销 owner(1)/tenant(5) 的 421（保留 435/436——(1,435)/(1,436)/(5,435)/(5,436) 不动）
--     INSERT IGNORE 无法撤销，须显式 DELETE（SEE: [[insert-ignore-swallows-errors]]）
DELETE FROM rel_role_permission WHERE (role_id, permission_id) IN ((1,421),(5,421));

-- 6.3 grid_worker(4) 授 421（本小区发布权，D6；property_admin(2) 保留 421——不做回收，推翻 notice D26）
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES (4, 421);

-- 6.4 新增读码 423/424/426（parent_id=410 community:read）+ 写码 427/428（parent_id=420 community:notice）
--     path 与实际 REST 路由一致（REST 路径保持 /api/community/notices，R2 wire 兼容；勿孤儿节点）
-- SEE: [[permission-seed-api-path-must-match-routes]]
INSERT IGNORE INTO sys_permission (id, parent_id, name, code, type, path, icon, sort_order, status, created_at, updated_at)
VALUES
(423, 410, 'GET /api/community/notices/marquee', 'community:notice:read-marquee-api', 3, 'GET:/api/community/notices/marquee', NULL, 16, 1, NOW(), NOW()),
(424, 410, 'GET /api/community/notices/publish-permission', 'community:notice:publish-permission-api', 3, 'GET:/api/community/notices/publish-permission', NULL, 17, 1, NOW(), NOW()),
(426, 410, 'GET /api/community/notices/:id', 'community:notice:read-detail-api', 3, 'GET:/api/community/notices/:id', NULL, 18, 1, NOW(), NOW()),
(427, 420, 'DELETE /api/community/notices/:id', 'community:notice:delete-api', 3, 'DELETE:/api/community/notices/:id', NULL, 21, 1, NOW(), NOW()),
(428, 420, 'PUT /api/community/notices/:id', 'community:notice:update-api', 3, 'PUT:/api/community/notices/:id', NULL, 22, 1, NOW(), NOW());

-- 6.5 422 扩展绑定全部移动端角色（现仅 (9,1,5)；补 grid_worker4 / committee6 / merchant7 /
--     community_admin3 / sys_admin8——sys_admin 的「全权限」绑定在 422 创建前执行，须补挂）
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(3, 422), (4, 422), (6, 422), (7, 422), (8, 422);

-- 6.6 新增 423/424/426/427/428 绑定全部移动端角色 {1,3,4,5,6,7,8,9}（各 8 角色 × 5 码 = 40 条）
--     property_admin(2) platforms='pc' 不绑（platforms 端准入；427/428 绑全移动端，
--     property_admin 创建后操作由 080002 作者校验兜底——评审 SHOULD #4 不对称注明）
INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES
(1, 423), (3, 423), (4, 423), (5, 423), (6, 423), (7, 423), (8, 423), (9, 423),
(1, 424), (3, 424), (4, 424), (5, 424), (6, 424), (7, 424), (8, 424), (9, 424),
(1, 426), (3, 426), (4, 426), (5, 426), (6, 426), (7, 426), (8, 426), (9, 426),
(1, 427), (3, 427), (4, 427), (5, 427), (6, 427), (7, 427), (8, 427), (9, 427),
(1, 428), (3, 428), (4, 428), (5, 428), (6, 428), (7, 428), (8, 428), (9, 428);

-- 6.7 content-post 权限种子验证（REQ-CPP-3 断言精确到具体码）
--     ① owner/tenant 撤销 421 生效（(1,421)/(5,421) 删除），② 发布角色保留（property_admin/grid_worker/community_admin/committee 持 421）
SELECT 'content-post 写权限 421' AS check_type,
       (SELECT COUNT(*) FROM rel_role_permission WHERE (role_id, permission_id) IN ((1,421),(5,421))) AS owner_tenant_421,
       (SELECT COUNT(*) FROM rel_role_permission WHERE (role_id, permission_id) IN ((2,421),(3,421),(4,421),(6,421))) AS publish_roles_421,
       CASE WHEN (SELECT COUNT(*) FROM rel_role_permission WHERE (role_id, permission_id) IN ((1,421),(5,421))) = 0
             AND (SELECT COUNT(*) FROM rel_role_permission WHERE (role_id, permission_id) IN ((2,421),(3,421),(4,421),(6,421))) = 4
             THEN '✅ PASS' ELSE '❌ FAIL' END AS status;

SELECT '421 min_verf_level' AS check_type, min_verf_level,
       CASE WHEN min_verf_level = 2 THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM sys_permission WHERE code = 'community:notice:create-api';

SELECT 'content-post 读码 422 扩展' AS check_type,
       (SELECT COUNT(*) FROM rel_role_permission WHERE permission_id = 422 AND role_id IN (1,3,4,5,6,7,8,9)) AS mobile_bindings,
       CASE WHEN (SELECT COUNT(*) FROM rel_role_permission WHERE permission_id = 422 AND role_id IN (1,3,4,5,6,7,8,9)) = 8
             THEN '✅ PASS' ELSE '❌ FAIL' END AS status;

SELECT 'content-post 新增读/写码绑定' AS check_type,
       (SELECT COUNT(*) FROM rel_role_permission WHERE permission_id IN (423,424,426,427,428) AND role_id IN (1,3,4,5,6,7,8,9)) AS bindings,
       CASE WHEN (SELECT COUNT(*) FROM rel_role_permission WHERE permission_id IN (423,424,426,427,428) AND role_id IN (1,3,4,5,6,7,8,9)) = 40
             THEN '✅ PASS' ELSE '❌ FAIL' END AS status;

SELECT 'content-post 新增码 parent_id（防孤儿节点）' AS check_type,
       (SELECT COUNT(*) FROM sys_permission WHERE id IN (423,424,426) AND parent_id = 410) AS read_parent_ok,
       (SELECT COUNT(*) FROM sys_permission WHERE id IN (427,428) AND parent_id = 420) AS write_parent_ok,
       CASE WHEN (SELECT COUNT(*) FROM sys_permission WHERE id IN (423,424,426) AND parent_id = 410) = 3
             AND (SELECT COUNT(*) FROM sys_permission WHERE id IN (427,428) AND parent_id = 420) = 2
             THEN '✅ PASS' ELSE '❌ FAIL' END AS status;

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
