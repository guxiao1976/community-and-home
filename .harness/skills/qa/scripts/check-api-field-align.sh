#!/usr/bin/env bash
# check-api-field-align.sh — 检查后端 json tag 与前端 interface 字段名对齐
# 防止 snake_case 与 camelCase 不匹配的问题
set -eu

ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"

echo "=== API 字段名对齐检查 ==="

errors=0

# 规则：从后端 types.go 提取已知的 snake_case 字段，
# 检查前端是否有人误用了 camelCase 版本

# 已知的 snake_case 字段 → 前端禁止使用的 camelCase 版本
declare -A forbidden=(
  ["created_at"]="createdAt"
  ["updated_at"]="updatedAt"
  ["deleted_at"]="deletedAt"
  ["user_type"]="userType"
  ["real_name"]="realName"
  ["id_card_number"]="idCardNumber"
  ["avatar_url"]="avatarUrl"
  ["birth_date"]="birthDate"
  ["credit_score"]="creditScore"
  ["sort_order"]="sortOrder"
  ["page_size"]="pageSize"
  ["total_pages"]="totalPages"
)

for snake in "${!forbidden[@]}"; do
  camel="${forbidden[$snake]}"
  # 在前端代码中搜索 camelCase 用法（排除 node_modules）
  matches=$(grep -rn "\.${camel}\b" "$ROOT/web/pc/src" --include="*.ts" --include="*.vue" 2>/dev/null | grep -v node_modules | grep -v '/common/types/' || true)
  if [ -n "$matches" ]; then
    while IFS= read -r line; do
      file=$(echo "$line" | cut -d: -f1)
      lineno=$(echo "$line" | cut -d: -f2)
      # 排除定义本身（interface 中的定义允许）
      if echo "$line" | grep -q "^\s*${camel}[?:]\|interface\|type\s"; then
        continue
      fi
      echo "❌ $file:$lineno — 使用了 .$camel，应为 .$snake"
      errors=$((errors + 1))
    done <<< "$matches"
  fi
done

if [ "$errors" -eq 0 ]; then
  echo "✅ 未发现 snake_case/camelCase 字段名不匹配"
else
  echo ""
  echo "共 $errors 处字段名不匹配，请将 camelCase 改为 snake_case"
  exit 1
fi
