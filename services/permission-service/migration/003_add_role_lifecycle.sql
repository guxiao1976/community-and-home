-- ============================================================================
-- 003_add_role_lifecycle.sql
-- rel_user_role 生命周期三列补齐（rel-user-role-migration-publish-fix P0）
--
-- 目标：
--   为 rel_user_role 补齐 status / verified_at / expires_at 三列，
--   对齐 model/rel.go RelUserRole db tag，消除从零建库时 MySQL 1054 Unknown column 'status'
--   status INT NOT NULL DEFAULT 2   —— 有 grant 即活跃，DEFAULT 2 不静默失效（0=未认证 1=待审 2=已认证 3=已驳回 4=已过期）
--   verified_at DATETIME NULL       —— NULL = 未认证（未通过）
--   expires_at DATETIME NULL        —— NULL = 永久（与 SQL 谓词 expires_at IS NULL OR expires_at > NOW() 一致）
--
-- 回填：零 guard 外 UPDATE——存量回填由 ADD COLUMN 的列定义在补列当次自动完成
--   （status DEFAULT 2 使存量行置 2；verified_at/expires_at DATETIME NULL 使存量行置 NULL），
--   不改写迁移后新行 / 已存在显式 status=0/4 的存量行。
--
-- 幂等：MySQL 8.0 不支持 ADD COLUMN IF NOT EXISTS，用 information_schema guard（逐列）。
-- SEE: [[migration-must-execute]]
-- ============================================================================

-- status 列（逐列 guard）
SET @status_col := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'rel_user_role' AND COLUMN_NAME = 'status');
SET @status_ddl := IF(@status_col = 0,
  'ALTER TABLE rel_user_role ADD COLUMN status INT NOT NULL DEFAULT 2 COMMENT ''个体角色生命周期: 0=未认证 1=待审 2=已认证 3=已驳回 4=已过期''',
  'SELECT ''status column already exists''');
PREPARE status_stmt FROM @status_ddl;
EXECUTE status_stmt;
DEALLOCATE PREPARE status_stmt;

-- verified_at 列（逐列 guard）
SET @verified_at_col := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'rel_user_role' AND COLUMN_NAME = 'verified_at');
SET @verified_at_ddl := IF(@verified_at_col = 0,
  'ALTER TABLE rel_user_role ADD COLUMN verified_at DATETIME NULL COMMENT ''个体认证通过时间''',
  'SELECT ''verified_at column already exists''');
PREPARE verified_at_stmt FROM @verified_at_ddl;
EXECUTE verified_at_stmt;
DEALLOCATE PREPARE verified_at_stmt;

-- expires_at 列（逐列 guard）
SET @expires_at_col := (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'rel_user_role' AND COLUMN_NAME = 'expires_at');
SET @expires_at_ddl := IF(@expires_at_col = 0,
  'ALTER TABLE rel_user_role ADD COLUMN expires_at DATETIME NULL COMMENT ''个体角色到期时间, NULL=永久''',
  'SELECT ''expires_at column already exists''');
PREPARE expires_at_stmt FROM @expires_at_ddl;
EXECUTE expires_at_stmt;
DEALLOCATE PREPARE expires_at_stmt;

-- 验证三列存在
SELECT 'status/verified_at/expires_at 三列检查' AS check_type, COUNT(*) AS count,
       CASE WHEN COUNT(*) = 3 THEN '✅ PASS' ELSE '❌ FAIL' END AS status
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'rel_user_role'
  AND COLUMN_NAME IN ('status', 'verified_at', 'expires_at');
