-- Migration: Refactor user-service from old 7-table to new 5-table design
-- Per: docs/specs/user-design.md
-- Run: mysql -h <host> -u <user> -p<pass> user < this_file.sql

-- ========== Phase 1: Clean up old tables ==========

DROP TABLE IF EXISTS user_family_member;
DROP TABLE IF EXISTS user_family;
DROP TABLE IF EXISTS user_uploaded_file;
DROP TABLE IF EXISTS user_homeowner_verification;
DROP TABLE IF EXISTS user_property_binding;
DROP TABLE IF EXISTS user_property_unit;

-- ========== Phase 2: Add new columns to user_base ==========

ALTER TABLE user_base
    ADD COLUMN real_name VARCHAR(50) NULL COMMENT '真实姓名' AFTER avatar_url,
    ADD COLUMN id_card_number VARCHAR(255) NULL COMMENT '身份证号（AES加密）' AFTER real_name,
    ADD COLUMN gender TINYINT NULL COMMENT '性别：1-男 2-女' AFTER id_card_number,
    ADD COLUMN birth_date DATE NULL COMMENT '出生日期' AFTER gender,
    ADD COLUMN preferences JSON NULL COMMENT '用户偏好' AFTER credit_score;

-- ========== Phase 3: Handle status=3 (deleted) records ==========
-- Convert soft-deleted records: set delete_time, change status to 2 (disabled)

UPDATE user_base SET delete_time = NOW() WHERE status = 3 AND delete_time IS NULL;
UPDATE user_base SET status = 2 WHERE status = 3;

-- ========== Phase 4: Drop old columns ==========

ALTER TABLE user_base
    DROP COLUMN user_type,
    DROP COLUMN cert_status,
    DROP COLUMN scope_id;

-- ========== Phase 5: Create new tables ==========

CREATE TABLE user_community_membership (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    user_id             BIGINT NOT NULL COMMENT '用户ID',
    community_id        BIGINT NOT NULL COMMENT '小区ID',
    bind_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1-有效 0-已退出',
    join_time           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    leave_time          DATETIME NULL COMMENT '退出时间',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_user_community (user_id, community_id),
    INDEX idx_community (community_id, bind_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-小区成员关系';

CREATE TABLE user_membership_role (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    user_id             BIGINT NOT NULL COMMENT '用户ID',
    membership_id       BIGINT NULL COMMENT '小区成员关系ID，商家为 NULL',
    community_id        BIGINT NOT NULL DEFAULT 0 COMMENT '小区ID，0=全局角色(商家)',
    role_code           VARCHAR(30) NOT NULL COMMENT '角色编码',
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

CREATE TABLE user_certification (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    role_id             BIGINT NOT NULL COMMENT '角色ID',
    user_id             BIGINT NOT NULL COMMENT '用户ID',
    document_urls       TEXT NULL COMMENT '证明材料URL列表（JSON数组）',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '审核状态：1-待审核 2-已通过 3-已驳回',
    reviewer_id         BIGINT NULL COMMENT '审核人ID',
    review_time         DATETIME NULL COMMENT '审核时间',
    review_notes        VARCHAR(500) NULL COMMENT '审核备注',
    submit_time         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    PRIMARY KEY (id),
    INDEX idx_role (role_id),
    INDEX idx_user (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证记录表';

CREATE TABLE user_residence (
    id                  BIGINT NOT NULL AUTO_INCREMENT,
    membership_id       BIGINT NOT NULL COMMENT '小区成员关系ID',
    house_id            VARCHAR(50) NOT NULL COMMENT '房屋ID，如 1-2-301',
    building            VARCHAR(20) NOT NULL COMMENT '楼号',
    unit                VARCHAR(20) NOT NULL DEFAULT '' COMMENT '单元号',
    room                VARCHAR(20) NOT NULL COMMENT '房号',
    is_primary          TINYINT NOT NULL DEFAULT 0 COMMENT '主房产：1-是 0-否',
    start_date          DATE NULL COMMENT '入住/合同开始日期',
    end_date            DATE NULL COMMENT '搬离/合同结束日期',
    created_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_time        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_house (membership_id, house_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='居民房屋明细表';
