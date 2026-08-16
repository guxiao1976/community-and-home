-- Migration: file guard - file_type/confirmed columns
-- Database: file_db
-- Description: 附件安全重构（content-post-generalization）——magic-bytes 嗅探类型与确认标志
-- Run: docker exec -i mysql mysql -uroot -proot123456 file_db < 002_file_guard.sql

USE file_db;

-- ========== uploaded_file — 附件安全守卫扩展 ==========
-- file_type：白名单规范类型（magic-bytes 嗅探产出，如 png/pdf/doc），存量行 NULL 免回填
-- confirmed：上传流程完成标志（ConfirmUpload 成功置 1）；存量行默认 1 免嗅探即视为已确认
ALTER TABLE uploaded_file
    ADD COLUMN file_type VARCHAR(20) DEFAULT NULL COMMENT '白名单规范类型（magic-bytes 嗅探产出，如 png/pdf/doc）',
    ADD COLUMN confirmed TINYINT NOT NULL DEFAULT 1 COMMENT '上传流程完成标志（ConfirmUpload 成功置 1；存量行默认 1 免嗅探）';
