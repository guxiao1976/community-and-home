#!/usr/bin/env bash
#
# create-file-db.sh - 为 file-service 创建独立数据库
#
# Usage:
#   bash scripts/create-file-db.sh
#

set -euo pipefail

echo "🔧 为 file-service 创建独立数据库..."

docker exec -i mysql mysql -uroot -proot123456 << 'SQL'
-- 创建 file-service 数据库
CREATE DATABASE IF NOT EXISTS file_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 验证
SELECT SCHEMA_NAME FROM information_schema.SCHEMATA
WHERE SCHEMA_NAME NOT IN ('mysql', 'information_schema', 'performance_schema', 'sys', 'default_db')
ORDER BY SCHEMA_NAME;
SQL

echo ""
echo "✅ file_db 数据库创建完成"
echo ""
echo "现在所有微服务都有独立数据库了："
echo "  ✅ user"
echo "  ✅ auth"
echo "  ✅ permission"
echo "  ✅ community_hub_db"
echo "  ✅ moderation_db"
echo "  ✅ ai_model_db"
echo "  ✅ masterdata_db"
echo "  ✅ file_db (新增)"
