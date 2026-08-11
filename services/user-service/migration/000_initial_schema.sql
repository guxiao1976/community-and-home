-- Migration: Initial schema for user-service (全新安装)
-- Database: user
-- Description: 创建用户服务的 5 张核心表
-- Run: docker exec -i mysql mysql -uroot -proot123456 user < 000_initial_schema.sql

-- ========== 1. user_base — 用户基础表 ==========

CREATE TABLE IF NOT EXISTS user_base (
    id                  BIGINT NOT NULL COMMENT '用户ID（雪花算法生成）',
    phone               VARCHAR(255) NOT NULL COMMENT '手机号（AES加密存储）',
    nickname            VARCHAR(100) NULL COMMENT '昵称',
    avatar_url          VARCHAR(500) NULL COMMENT '头像URL',
    real_name           VARCHAR(50) NULL COMMENT '真实姓名（首次认证通过后回填）',
    id_card_number      VARCHAR(255) NULL COMMENT '身份证号（AES加密，首次认证通过后回填）',
    gender              TINYINT NULL COMMENT '性别：1-男 2-女',
    birth_date          DATE NULL COMMENT '出生日期',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-正常 2-禁用',
    credit_score        INT NOT NULL DEFAULT 100 COMMENT '信用分（默认100，最低0）',
    preferences         JSON NULL COMMENT '用户偏好。例: {"default_community_id":123}',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delete_time         DATETIME NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    UNIQUE INDEX idx_phone (phone),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基础表';

-- ========== 2. user_community_membership — 用户-小区成员关系表 ==========

CREATE TABLE IF NOT EXISTS user_community_membership (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    user_id             BIGINT NOT NULL COMMENT '用户ID（FK → user_base.id）',
    community_id        BIGINT NOT NULL COMMENT '小区ID',
    bind_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1-有效 0-已退出',
    join_time           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    leave_time          DATETIME NULL COMMENT '退出时间',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_user_community (user_id, community_id),
    INDEX idx_community (community_id, bind_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-小区成员关系（仅表达加入/退出）';

-- ========== 3. user_membership_role — 角色表 ==========

CREATE TABLE IF NOT EXISTS user_membership_role (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    user_id             BIGINT NOT NULL COMMENT '冗余：用户ID（FK → user_base.id，分片键）',
    membership_id       BIGINT NULL COMMENT '小区成员关系ID（FK → user_community_membership.id），商家为 NULL',
    community_id        BIGINT NOT NULL DEFAULT 0 COMMENT '冗余：小区ID，0=全局角色(商家)',
    role_code           VARCHAR(30) NOT NULL COMMENT '角色编码：owner/tenant/grid_worker/community_admin/property_admin/committee/merchant',
    verf_status         TINYINT NOT NULL DEFAULT 0 COMMENT '认证状态：0-未认证 1-待审 2-已通过 3-已驳回 4-已过期',
    verified_at         DATETIME NULL COMMENT '认证通过时间',
    expires_at          DATETIME NULL COMMENT '过期时间，NULL=永久有效',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_role (membership_id, role_code),
    UNIQUE INDEX uk_user_community_role (user_id, community_id, role_code),
    INDEX idx_community_role (community_id, role_code, verf_status),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色表';

-- ========== 4. user_certification — 认证记录表 ==========

CREATE TABLE IF NOT EXISTS user_certification (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    role_id             BIGINT NOT NULL COMMENT '角色ID（FK → user_membership_role.id）',
    user_id             BIGINT NOT NULL COMMENT '冗余：用户ID（FK → user_base.id）',
    document_urls       TEXT NULL COMMENT '证明材料URL列表（JSON数组，文件存 MinIO/file-service）',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '审核状态：1-待审核 2-已通过 3-已驳回',
    reviewer_id         BIGINT NULL COMMENT '审核人ID',
    review_time         DATETIME NULL COMMENT '审核时间',
    review_notes        VARCHAR(500) NULL COMMENT '审核备注',
    submit_time         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    PRIMARY KEY (id),
    INDEX idx_role (role_id),
    INDEX idx_user (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证记录表（所有角色统一流程）';

-- ========== 5. user_residence — 房屋明细表 ==========

CREATE TABLE IF NOT EXISTS user_residence (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    membership_id       BIGINT NOT NULL COMMENT '小区成员关系ID（FK → user_community_membership.id）',
    user_id             BIGINT NOT NULL COMMENT '冗余：用户ID（分片键，FK → user_base.id）',
    house_id            VARCHAR(50) NOT NULL COMMENT '房屋ID，系统拼接，如 1-2-301',
    building            VARCHAR(20) NOT NULL COMMENT '楼号（用户输入）',
    unit                VARCHAR(20) NOT NULL DEFAULT '' COMMENT '单元号（用户输入）',
    room                VARCHAR(20) NOT NULL COMMENT '房号（用户输入）',
    is_primary          TINYINT NOT NULL DEFAULT 0 COMMENT '多套房时标记主房产：1-是 0-否',
    start_date          DATE NULL COMMENT '入住/合同开始日期',
    end_date            DATE NULL COMMENT '搬离/合同结束日期',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_house (membership_id, house_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='居民房屋明细表';

-- ========== 初始化完成 ==========
-- 所有表均使用雪花算法生成主键（非 AUTO_INCREMENT）
-- 手机号、身份证号使用 AES-256 加密存储
-- 按 user_id 分片策略：所有表均含 user_id，同一用户数据落在同一分片
