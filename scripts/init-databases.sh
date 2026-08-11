#!/usr/bin/env bash
#
# init-databases.sh — 初始化微服务数据库（幂等，绝不删除数据）
#
# 背景：`docker compose up -d` 拉起的 MySQL 是全新空实例，首次部署时
# 没有 user/auth/... 这些库，服务一连就报 Unknown database。本脚本只做
# 一件事：确保 8 个微服务数据库存在。
#
# 数据安全承诺（数据宝贵，绝不误删）：
#   - 只执行 CREATE DATABASE IF NOT EXISTS（库已存在则跳过，零影响）
#   - 脚本内禁止 DROP / TRUNCATE / DELETE FROM（.harness QA 检查会拦截）
#   - 不建表、不插数据（表结构由各服务 migration / model 层负责）
#   - 定期备份请用 scripts/backup-db.sh（另行配置，勿依赖本脚本）
#
# 用法：
#   bash scripts/init-databases.sh              # 幂等建库（缺哪个建哪个）
#   bash scripts/init-databases.sh --check-only # 只检查缺失，不创建
#   bash scripts/init-databases.sh --dry-run    # 只打印将执行的 SQL，不连库
#
# 数据库清单（微服务 → 库名）：
#   user / auth / permission / community_hub_db / moderation_db /
#   ai_model_db / masterdata_db / file_db
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# ─── 参数解析 ──────────────────────────────────────────────────────
CHECK_ONLY=false
DRY_RUN=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=true; shift ;;
    --dry-run)    DRY_RUN=true; shift ;;
    -h|--help) sed -n '2,26p' "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "未知参数: $1（支持 --check-only / --dry-run）" >&2; exit 2 ;;
  esac
done

# ─── 加载数据库凭据（优先 .env，兜底 compose 默认值）─────────────
set -a
[[ -f .env ]] && source .env
set +a
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-${MYSQL_ROOT_PASSWORD:-root123456}}"
MYSQL_CONTAINER="${MYSQL_CONTAINER:-mysql}"

DBS=(user auth permission community_hub_db moderation_db ai_model_db masterdata_db file_db)

# ─── 干跑模式：只打印 SQL，不连库 ────────────────────────────────
if $DRY_RUN; then
  echo "=== DRY-RUN：将执行的 SQL（全部幂等）==="
  for db in "${DBS[@]}"; do
    echo "  CREATE DATABASE IF NOT EXISTS ${db} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
  done
  echo ""
  echo "（dry-run 未连接数据库，未做任何修改）"
  exit 0
fi

# ─── 检查 mysql 容器是否运行 ──────────────────────────────────────
if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -qx "$MYSQL_CONTAINER"; then
  echo "❌ mysql 容器未运行。请先执行: docker compose up -d" >&2
  echo "   （或设置 MYSQL_CONTAINER 指向你的容器名）" >&2
  exit 1
fi

exec_sql() {
  docker exec -i "$MYSQL_CONTAINER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" -N -B
}

# 读取当前已存在的库
existing="$(echo "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA;" | exec_sql 2>/dev/null || echo "")"

# ─── 检查模式 ─────────────────────────────────────────────────────
if $CHECK_ONLY; then
  echo "=== CHECK-ONLY：数据库现状（不创建）==="
  local_missing=0
  for db in "${DBS[@]}"; do
    if echo "$existing" | grep -qx "$db"; then
      echo "  ✅ $db 已存在"
    else
      echo "  ❌ $db 缺失（运行 bash scripts/init-databases.sh 创建）"
      local_missing=$((local_missing + 1))
    fi
  done
  echo ""
  [[ $local_missing -eq 0 ]] && echo "✅ 全部数据库就绪" || echo "⚠️  ${local_missing} 个库缺失"
  exit 0
fi

# ─── 幂等创建（缺哪个建哪个，已有库一律跳过）────────────────────
echo "=== 初始化微服务数据库（幂等）==="
missing=0
for db in "${DBS[@]}"; do
  if echo "$existing" | grep -qx "$db"; then
    echo "  ⏭️  $db 已存在，跳过（不影响已有数据）"
  else
    echo "  ➕ 创建 $db ..."
    echo "CREATE DATABASE IF NOT EXISTS ${db} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" | exec_sql
    missing=$((missing + 1))
  fi
done

echo ""
if [[ $missing -eq 0 ]]; then
  echo "✅ 全部数据库已存在，无需变更（对现有数据零影响）"
else
  echo "✅ 已创建 $missing 个缺失数据库"
fi

# ─── 验证 ─────────────────────────────────────────────────────────
echo ""
echo "=== 数据库清单验证 ==="
echo "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME NOT IN ('mysql','information_schema','performance_schema','sys','default_db') ORDER BY SCHEMA_NAME;" \
  | exec_sql | sed 's/^/  /'
echo ""
echo "💡 如需备份数据: 配置 scripts/backup-db.sh（mysqldump + 保留周期）"
