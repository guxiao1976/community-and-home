-- Migration: Initial schema for auth-service
-- Database: auth
-- Description: 创建登录凭证表
-- Run: docker exec mysql mysql -uroot -proot123456 auth < 001_initial.sql

-- ========== 创建 auth_credential 表 ==========

CREATE TABLE IF NOT EXISTS auth_credential (
    id              BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    user_id         BIGINT NOT NULL COMMENT '用户ID（FK → user_base.id）',
    identity_type   VARCHAR(20) NOT NULL COMMENT '登录方式：phone / sms / wechat',
    identifier      VARCHAR(255) NOT NULL COMMENT '标识符（AES 加密的手机号）',
    credential      VARCHAR(255) NOT NULL COMMENT '凭证（bcrypt 密文）',
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE INDEX uk_identity (identity_type, identifier),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录凭证表';

-- ========== 初始化说明 ==========
-- 表设计说明：
-- 1. identifier: AES-256 加密存储的手机号（不存明文）
-- 2. credential: bcrypt 加盐哈希的密码
-- 3. 支持多种登录方式（phone/sms/wechat），同一 identity_type 下 identifier 唯一
-- 4. 一个用户可有多条凭证（账密登录、短信登录分开记录）
