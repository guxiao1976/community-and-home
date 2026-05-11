-- =============================================
-- Moderation 微服务 - 数据库迁移脚本
-- =============================================

-- 1. 修改 masterdata_db.md_sensitive_word 表结构
USE masterdata_db;

ALTER TABLE md_sensitive_word
  ADD COLUMN word_type TINYINT NOT NULL DEFAULT 1 COMMENT '1=黑名单(敏感词), 2=白名单(豁免词)' AFTER word,
  ADD COLUMN pinyin_expanded TINYINT(1) DEFAULT 0 COMMENT '是否已生成谐音变体' AFTER action,
  ADD COLUMN submission_status TINYINT DEFAULT 0 COMMENT '提交审核状态: 0=待审, 1=通过, 2=拒绝' AFTER status,
  ADD COLUMN submission_type TINYINT DEFAULT NULL COMMENT '提交类型' AFTER submission_status,
  ADD COLUMN change_snapshot JSON DEFAULT NULL COMMENT '变更快照' AFTER submission_type,
  ADD COLUMN submitter_id BIGINT DEFAULT NULL COMMENT '提交人ID' AFTER change_snapshot,
  ADD COLUMN submit_time DATETIME DEFAULT NULL COMMENT '提交时间' AFTER submitter_id,
  ADD COLUMN reviewer_id BIGINT DEFAULT NULL COMMENT '审核人ID' AFTER submit_time,
  ADD COLUMN review_time DATETIME DEFAULT NULL COMMENT '审核时间' AFTER reviewer_id,
  ADD COLUMN review_notes VARCHAR(500) DEFAULT NULL COMMENT '审核备注' AFTER review_time,
  ADD COLUMN delete_time DATETIME DEFAULT NULL COMMENT '软删除时间' AFTER review_notes,
  ADD INDEX idx_word_type (word_type),
  ADD INDEX idx_status_type (status, word_type);

-- 2. 创建 moderation_db 及审核日志表
CREATE DATABASE IF NOT EXISTS moderation_db
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE moderation_db;

CREATE TABLE IF NOT EXISTS mod_audit_log (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    content_type    VARCHAR(16) NOT NULL COMMENT 'text/image',
    content_summary VARCHAR(500) COMMENT '内容摘要',
    risk_level      VARCHAR(16) NOT NULL COMMENT 'high/medium/low',
    pass            TINYINT(1) NOT NULL COMMENT '是否通过',
    reason          VARCHAR(500) COMMENT '原因',
    check_layer     VARCHAR(32) COMMENT 'ac_engine/small_model/large_model/image_hash',
    matched_items   JSON COMMENT '命中详情',
    user_id         BIGINT COMMENT '提交用户ID',
    source_type     VARCHAR(32) COMMENT '来源: post/comment/nickname',
    source_id       BIGINT COMMENT '来源ID',
    need_review     TINYINT(1) DEFAULT 0 COMMENT '是否需要人工复审',
    review_status   TINYINT(1) DEFAULT 0 COMMENT '复审状态: 0待审/1通过/2拒绝',
    reviewer_id     BIGINT COMMENT '复审人ID',
    review_time     DATETIME COMMENT '复审时间',
    created_time    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_source (source_type, source_id),
    INDEX idx_review (review_status, created_time),
    INDEX idx_created (created_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审核日志表';
