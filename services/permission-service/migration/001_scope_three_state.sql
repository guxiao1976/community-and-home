-- ============================================================================
-- 001_scope_three_state.sql
-- 数据权限核心（access-data-permission 阶段①）：rel_user_role scope 三态语义
--
-- 目标：
--   1. 唯一索引 uk_user_role_scope（幂等基础：registered_user 分配 / Join 自动授权并发只一条）
--   2. scope_type 保留取值 'global' / ''（空串）承载三态（global / limited / empty）
--
-- 幂等：MySQL 8.0 不支持 ADD INDEX IF NOT EXISTS，用 information_schema guard。
-- SEE: [[migration-must-execute]]
-- ============================================================================

-- 唯一索引（已存在则跳过）
SET @uk_exists := (SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'rel_user_role' AND INDEX_NAME = 'uk_user_role_scope');
SET @uk_ddl := IF(@uk_exists = 0,
  'ALTER TABLE rel_user_role ADD UNIQUE KEY uk_user_role_scope (user_id, role_id, scope_type, scope_id)',
  'SELECT ''uk_user_role_scope already exists''');
PREPARE uk_stmt FROM @uk_ddl;
EXECUTE uk_stmt;
DEALLOCATE PREPARE uk_stmt;
