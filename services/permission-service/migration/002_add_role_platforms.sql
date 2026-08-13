-- ============================================================================
-- 002_add_role_platforms.sql
-- 访问控制（access-control）：sys_role 增 platforms 列
--
-- 目标：
--   角色「允许登录的端」配置（pc/mobile，逗号分隔；空=未声明，运行时 fail-open 允许所有端）
--   由 auth-service 端准入判定消费（GetUserRoles → role.platforms）
--
-- 幂等：MySQL 8.0 不支持 ADD COLUMN IF NOT EXISTS，用 information_schema guard。
-- SEE: [[migration-must-execute]]
-- ============================================================================

SET @platforms_col := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_role' AND COLUMN_NAME = 'platforms');
SET @platforms_ddl := IF(@platforms_col = 0,
  'ALTER TABLE sys_role ADD COLUMN platforms VARCHAR(32) NOT NULL DEFAULT '''' COMMENT ''允许登录的端，逗号分隔：pc,mobile；空=未声明（fail-open）''',
  'SELECT ''platforms column already exists''');
PREPARE platforms_stmt FROM @platforms_ddl;
EXECUTE platforms_stmt;
DEALLOCATE PREPARE platforms_stmt;

-- 验证列存在
SELECT 'platforms 列检查' AS check_type, COUNT(*) AS count,
       CASE WHEN COUNT(*) = 1 THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_role' AND COLUMN_NAME = 'platforms';
