-- Migration: submission_status=4 (待删除) → submission_status=0, submission_type=3 (删除待提交)
-- Execute: mysql -u root -p masterdata_db < migration_submission_status.sql

-- Step 1: Migrate legacy status 4 to new dual-field format
UPDATE md_administrative_division
SET submission_status = 0, submission_type = 3
WHERE submission_status = 4 AND delete_time IS NULL;

-- Step 2: Add missing columns (skip if already exist - check with SHOW COLUMNS first)
-- These columns are defined in masterdata_schema.sql; only run ALTER if needed.
-- Example (MySQL 8.0 does not support ADD COLUMN IF NOT EXISTS):
-- ALTER TABLE md_administrative_division ADD COLUMN submission_type TINYINT NULL COMMENT '1=Create, 2=Update, 3=Delete' AFTER submission_status;
