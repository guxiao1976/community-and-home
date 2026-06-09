-- ==========================================================================
-- Community Home 微服务拆分 - 数据库迁移脚本
-- 数据库凭证：root:root123456@localhost:3306（来自 docker-compose.yml）
-- 日期：2026-05-30
-- ==========================================================================

-- ==========================================================================
-- 1. 用户中心 -- 数据库 user（7 张表）
-- 参考 spec/user.md 数据模型
-- ==========================================================================

USE `user`;

-- 1.1 用户基础表（核心）
CREATE TABLE IF NOT EXISTS `user_base` (
    `id` BIGINT NOT NULL COMMENT '用户ID（雪花算法生成）',
    `phone` VARCHAR(255) NOT NULL COMMENT '手机号（AES加密存储）',
    `nickname` VARCHAR(100) NULL COMMENT '昵称',
    `avatar_url` VARCHAR(500) NULL COMMENT '头像URL',
    `user_type` INT NOT NULL DEFAULT 1 COMMENT '用户类型：1-居民, 2-租户, 3-访客, 4-物业人员, 5-网格员, 6-管理员',
    `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1-正常, 2-禁用, 3-已删除',
    `cert_status` INT NOT NULL DEFAULT 0 COMMENT '实名认证状态：0-未认证, 1-待审核, 2-已通过, 3-已驳回',
    `credit_score` INT NOT NULL DEFAULT 100 COMMENT '信用分（默认100，最低0）',
    `scope_id` BIGINT NULL COMMENT '数据范围ID（如小区ID），NULL=无限制',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `delete_time` DATETIME NULL COMMENT '软删除时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_phone` (`phone`),
    INDEX `idx_status` (`status`),
    INDEX `idx_cert_status` (`cert_status`),
    INDEX `idx_scope_id` (`scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基础表';

-- 1.2 实名认证申请表
CREATE TABLE IF NOT EXISTS `user_homeowner_verification` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `property_unit_id` BIGINT NOT NULL COMMENT '房产单元ID',
    `document_urls` TEXT NULL COMMENT '证明材料URL列表（JSON数组）',
    `real_name` VARCHAR(50) NOT NULL COMMENT '真实姓名',
    `id_card_number` VARCHAR(255) NOT NULL COMMENT '身份证号（AES加密）',
    `verification_status` INT NOT NULL DEFAULT 0 COMMENT '状态：0-待审核, 1-已通过, 2-已驳回',
    `reviewer_id` BIGINT NULL COMMENT '审核人ID',
    `review_time` DATETIME NULL COMMENT '审核时间',
    `review_notes` VARCHAR(500) NULL COMMENT '审核备注',
    `submit_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_status` (`verification_status`),
    INDEX `idx_property` (`property_unit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='实名认证申请表';

-- 1.3 房产单元表
CREATE TABLE IF NOT EXISTS `user_property_unit` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `community_id` BIGINT NOT NULL COMMENT '社区ID',
    `building` VARCHAR(50) NOT NULL COMMENT '楼栋',
    `unit` VARCHAR(50) NOT NULL COMMENT '单元',
    `floor` VARCHAR(10) NULL COMMENT '楼层',
    `area` DOUBLE NULL COMMENT '面积',
    `property_type` INT NOT NULL DEFAULT 1 COMMENT '房产类型',
    `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1-有效, 0-无效',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `delete_time` DATETIME NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_community_building_unit` (`community_id`, `building`, `unit`),
    INDEX `idx_community` (`community_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='房产单元表';

-- 1.4 用户房产绑定表
CREATE TABLE IF NOT EXISTS `user_property_binding` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `property_unit_id` BIGINT NOT NULL COMMENT '房产单元ID',
    `is_primary` INT NOT NULL DEFAULT 0 COMMENT '是否主产权人：1-是, 0-否',
    `bind_status` INT NOT NULL DEFAULT 1 COMMENT '绑定状态：1-有效, 0-已解绑',
    `bind_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    `revoke_time` DATETIME NULL COMMENT '解绑时间',
    `revoked_by` BIGINT NULL COMMENT '解绑操作人ID',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_property` (`property_unit_id`),
    INDEX `idx_user_property` (`user_id`, `property_unit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户房产绑定表';

-- 1.5 家庭分组表
CREATE TABLE IF NOT EXISTS `user_family` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `property_unit_id` BIGINT NOT NULL COMMENT '房产单元ID',
    `family_head_id` BIGINT NOT NULL COMMENT '户主ID（关联 user_base.id）',
    `family_name` VARCHAR(100) NULL COMMENT '家庭名称',
    `status` INT NOT NULL DEFAULT 1 COMMENT '状态',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `delete_time` DATETIME NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_property` (`property_unit_id`),
    INDEX `idx_head` (`family_head_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭分组表';

-- 1.6 家庭成员表
CREATE TABLE IF NOT EXISTS `user_family_member` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `family_id` BIGINT NOT NULL COMMENT '家庭ID',
    `user_id` BIGINT NULL COMMENT '关联用户ID（非注册用户可为NULL）',
    `name` VARCHAR(50) NOT NULL COMMENT '姓名',
    `relationship` VARCHAR(20) NOT NULL COMMENT '关系：spouse/child/parent/other',
    `phone` VARCHAR(255) NULL COMMENT '手机号（AES加密）',
    `id_card_number` VARCHAR(255) NULL COMMENT '身份证号（AES加密）',
    `birth_date` DATETIME NULL COMMENT '出生日期',
    `gender` INT NULL COMMENT '性别：1-男, 2-女',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `delete_time` DATETIME NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    INDEX `idx_family_id` (`family_id`),
    INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='家庭成员表';

-- 1.7 文件附件表（MinIO 存储）
CREATE TABLE IF NOT EXISTS `user_uploaded_file` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '上传用户ID',
    `entity_type` VARCHAR(50) NOT NULL COMMENT '关联实体类型（如 verification）',
    `entity_id` BIGINT NOT NULL COMMENT '关联实体ID',
    `file_name` VARCHAR(255) NOT NULL COMMENT '原始文件名',
    `file_path` VARCHAR(500) NOT NULL COMMENT 'MinIO 路径',
    `file_size` BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小（字节）',
    `file_type` VARCHAR(100) NULL COMMENT 'MIME 类型',
    `bucket_name` VARCHAR(100) NOT NULL COMMENT 'MinIO Bucket 名称',
    `upload_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_entity` (`entity_type`, `entity_id`),
    INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件附件表';


-- ==========================================================================
-- 2. 认证中心 -- 数据库 auth（1 张表）
-- 参考 spec/auth.md 数据模型
-- ==========================================================================

USE `auth`;

-- 2.1 登录凭证表
CREATE TABLE IF NOT EXISTS `auth_credential` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '用户ID（关联 user.user_base.id）',
    `identity_type` VARCHAR(20) NOT NULL COMMENT '身份类型：phone-手机号, sms-短信，wechat-微信',
    `identifier` VARCHAR(512) NOT NULL COMMENT '身份标识（RSA密文存储的手机号或openid）',
    `credential` VARCHAR(512) NOT NULL COMMENT '凭证（bcrypt密文）',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_user_identity` (`user_id`, `identity_type`),
    INDEX `idx_identifier` (`identifier`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录凭证表';


-- ==========================================================================
-- 3. 权限中心 -- 数据库 permission（4 张表）
-- 参考 spec/permission.md 数据模型
-- ==========================================================================

USE `permission`;

-- 3.1 角色定义表
CREATE TABLE IF NOT EXISTS `sys_role` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `role_code` VARCHAR(100) NOT NULL COMMENT '角色编码：owner-房主, property_admin-物业管理员, community_admin-社区管理员, grid_worker-网格员',
    `role_name` VARCHAR(100) NOT NULL COMMENT '角色名称',
    `description` VARCHAR(500) NULL COMMENT '角色描述',
    `is_system` INT NOT NULL DEFAULT 0 COMMENT '是否系统角色（1-是，不可删除且自动获得所有权限）',
    `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
    `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1-启用, 0-禁用',
    `created_by` BIGINT NOT NULL COMMENT '创建者ID',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `delete_time` DATETIME NULL COMMENT '软删除',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_code` (`role_code`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色定义表';

-- 3.2 权限定义表（树形结构）
CREATE TABLE IF NOT EXISTS `sys_permission` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `parent_id` BIGINT NULL COMMENT '父权限ID（NULL=顶级）',
    `name` VARCHAR(100) NOT NULL COMMENT '权限名称',
    `code` VARCHAR(100) NOT NULL COMMENT '权限编码（全局唯一，如 user:read, user:write）',
    `type` INT NOT NULL DEFAULT 3 COMMENT '权限类型：1-菜单(menu), 2-按钮(button), 3-API(api)',
    `path` VARCHAR(500) NULL COMMENT '资源路径（API类型时为 API Path，如 /api/user/v1/GetUser）',
    `icon` VARCHAR(100) NULL COMMENT '图标',
    `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
    `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1-启用, 0-禁用',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_code` (`code`),
    INDEX `idx_parent` (`parent_id`),
    INDEX `idx_type` (`type`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限定义表';

-- 3.3 角色-权限关联表
CREATE TABLE IF NOT EXISTS `rel_role_permission` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `role_id` BIGINT NOT NULL COMMENT '角色ID',
    `permission_id` BIGINT NOT NULL COMMENT '权限ID',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_role_perm` (`role_id`, `permission_id`),
    INDEX `idx_perm` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-权限关联表';

-- 3.4 用户-角色关联表（含数据范围）
CREATE TABLE IF NOT EXISTS `rel_user_role` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `role_id` BIGINT NOT NULL COMMENT '角色ID',
    `scope_type` VARCHAR(50) NOT NULL DEFAULT 'community' COMMENT '作用域类型：community-小区, building-楼栋, unit-单元, grid-网格',
    `scope_id` BIGINT NOT NULL COMMENT '作用域ID（小区ID/楼栋ID等）',
    `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_user_role_scope` (`user_id`, `role_id`, `scope_type`, `scope_id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_role_id` (`role_id`),
    INDEX `idx_scope` (`scope_type`, `scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-角色关联表（含数据范围）';


-- ==========================================================================
-- 4. 插入默认数据
-- ==========================================================================

-- 默认系统角色
INSERT INTO `permission`.`sys_role` (`role_code`, `role_name`, `description`, `is_system`, `status`, `created_by`) VALUES
('super_admin', '超级管理员', '系统超级管理员，自动拥有所有权限', 1, 1, 0),
('community_admin', '社区管理员', '管理单个社区的日常事务', 0, 1, 0),
('property_admin', '物业管理员', '管理小区的房产、住户等信息', 0, 1, 0),
('grid_worker', '网格员', '负责网格内的居民走访和信息采集', 0, 1, 0),
('resident', '居民', '普通社区居民（房主/租户）', 0, 1, 0);

-- 默认权限（API 级别）
INSERT INTO `permission`.`sys_permission` (`parent_id`, `name`, `code`, `type`, `path`, `sort_order`, `status`) VALUES
(NULL, '用户管理', 'user', 1, NULL, 1, 1),
(1, '查看用户', 'user:read', 3, '/api/user/v1/GetUser', 1, 1),
(1, '创建用户', 'user:write', 3, '/api/user/v1/CreateUser', 2, 1),
(1, '更新用户', 'user:update', 3, '/api/user/v1/UpdateUser', 3, 1),
(1, '用户列表', 'user:list', 3, '/api/user/v1/ListUsers', 4, 1),
(NULL, '认证管理', 'auth', 1, NULL, 2, 1),
(6, '登录', 'auth:login', 3, '/api/auth/v1/Login', 1, 1),
(6, '注销', 'auth:logout', 3, '/api/auth/v1/Logout', 2, 1),
(NULL, '权限管理', 'permission', 1, NULL, 3, 1),
(9, '角色管理', 'permission:role', 3, '/api/permission/v1/ListRoles', 1, 1),
(9, '分配角色', 'permission:assign', 3, '/api/permission/v1/AssignRole', 2, 1);

-- 超级管理员拥有所有权限（示例）
-- INSERT INTO `permission`.`rel_role_permission` (`role_id`, `permission_id`)
-- SELECT 1, id FROM `permission`.`sys_permission` WHERE status = 1;
