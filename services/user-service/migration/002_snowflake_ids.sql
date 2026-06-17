-- Migration: 统一 ID 策略 + user_residence 补充 user_id
-- 1. AUTO_INCREMENT → 雪花 ID（分库分表兼容）
-- 2. user_residence 新增 user_id（分片键冗余）

-- ========== Phase 1: 去 AUTO_INCREMENT，改为雪花 ID ==========
-- 注意：已有数据的 id 保持不变，新插入使用雪花算法生成

ALTER TABLE user_community_membership MODIFY id BIGINT NOT NULL;
ALTER TABLE user_membership_role      MODIFY id BIGINT NOT NULL;
ALTER TABLE user_certification        MODIFY id BIGINT NOT NULL;
ALTER TABLE user_residence            MODIFY id BIGINT NOT NULL;

-- ========== Phase 2: user_residence 补充 user_id ==========

ALTER TABLE user_residence
    ADD COLUMN user_id BIGINT NOT NULL DEFAULT 0 COMMENT '冗余：用户ID（分片键）' AFTER membership_id;

-- 回填现有数据
UPDATE user_residence r
JOIN user_community_membership m ON m.id = r.membership_id
SET r.user_id = m.user_id
WHERE r.user_id = 0;

-- 添加索引
ALTER TABLE user_residence ADD INDEX idx_user_id (user_id);
