#!/usr/bin/env bash
# log-usage.sh — 追加一行 JSON 到 .harness/logs/usage/<script>.jsonl
#
# 为流水线复盘提供第一手数据：哪个脚本被调用、参数、结果。
# 记录位置：.harness/logs/usage/<script-name>.jsonl（JSONL：每行一个 JSON，运行时产物，gitignore）
#
# 用法:
#   bash log-usage.sh <script-name> [key=value ...]
#   bash log-usage.sh knowledge-load service=user-service top=5 matched=3 "keywords=角色,权限"
#
# 读取（复盘时）:
#   cat .harness/logs/usage/knowledge-load.jsonl
#   jq -s '.' .harness/logs/usage/harness-checks.jsonl   # 按时间序的 JSON 数组
set -uo pipefail

SCRIPT_NAME="$(basename "${1:-unknown}")"
shift

# JSON 值转义（" \ 换行）
escape_json() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  printf '%s' "$s"
}

json="{\"ts\":\"$(date +%Y-%m-%dT%H:%M:%S)\",\"script\":\"$SCRIPT_NAME\""
for kv in "$@"; do
  [[ "$kv" != *=* ]] && continue
  key="${kv%%=*}"
  val="$(escape_json "${kv#*=}")"
  json="$json,\"$key\":\"$val\""
done
json="$json}"

# 位置：项目根/.harness/logs/usage/<script>.jsonl
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
USAGE_DIR="$ROOT/.harness/logs/usage"
mkdir -p "$USAGE_DIR"
echo "$json" >> "$USAGE_DIR/$SCRIPT_NAME.jsonl"
