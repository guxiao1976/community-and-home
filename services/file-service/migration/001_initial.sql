-- Migration: Initial schema for file-service
-- Database: file_db
-- Description: 创建文件上传记录表
-- Run: docker exec -i mysql mysql -uroot -proot123456 file_db < 001_initial.sql

CREATE DATABASE IF NOT EXISTS file_db
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE file_db;

-- ========== uploaded_file — 文件上传记录表 ==========

CREATE TABLE IF NOT EXISTS uploaded_file (
    id              BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID（JSON输出为string，Snowflake兼容）',
    user_id         BIGINT NOT NULL COMMENT '上传用户ID',
    entity_type     VARCHAR(64) NOT NULL DEFAULT '' COMMENT '业务关联类型：verification/avatar/article',
    entity_id       BIGINT NOT NULL DEFAULT 0 COMMENT '关联的业务实体ID，0=未关联',
    file_name       VARCHAR(255) NOT NULL COMMENT '原始文件名',
    file_path       VARCHAR(512) NOT NULL COMMENT 'MinIO object key',
    file_size       BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小（字节）',
    mime_type       VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'MIME类型',
    bucket_name     VARCHAR(64) NOT NULL DEFAULT 'community-home' COMMENT 'MinIO bucket名称',
    upload_time     DATETIME NOT NULL COMMENT '上传完成时间',
    is_deleted      TINYINT NOT NULL DEFAULT 0 COMMENT '软删除标记：0-正常 1-已删除',
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    INDEX idx_user (user_id),
    INDEX idx_entity (entity_type, entity_id),
    INDEX idx_deleted (is_deleted, updated_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件上传记录表';

-- ========== 初始化说明 ==========
-- Object Key 命名规则: uploads/{user_id}/{timestamp_nano}_{filename}
-- 上传流程：
--   1. 客户端请求预签名 URL
--   2. 客户端直传 MinIO（15分钟有效期）
--   3. 客户端确认上传 → 写入此表
-- 删除策略：软删除（is_deleted=1），DB优先于MinIO一致性
