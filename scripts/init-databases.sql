-- 初始化所有微服务数据库
-- 文件: scripts/init-databases.sql
-- 用途: 创建所有微服务需要的数据库

-- ==================== 创建数据库 ====================

-- 用户服务
CREATE DATABASE IF NOT EXISTS user CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 认证服务
CREATE DATABASE IF NOT EXISTS auth CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 权限服务
CREATE DATABASE IF NOT EXISTS permission CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 社区服务
CREATE DATABASE IF NOT EXISTS community_hub_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 审核服务
CREATE DATABASE IF NOT EXISTS moderation_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- AI 模型服务
CREATE DATABASE IF NOT EXISTS ai_model_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 主数据服务（已存在，但保留创建语句）
CREATE DATABASE IF NOT EXISTS masterdata_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ==================== 验证 ====================

SELECT
    SCHEMA_NAME as '数据库名称',
    DEFAULT_CHARACTER_SET_NAME as '字符集',
    DEFAULT_COLLATION_NAME as '排序规则'
FROM information_schema.SCHEMATA
WHERE SCHEMA_NAME NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys', 'default_db')
ORDER BY SCHEMA_NAME;
