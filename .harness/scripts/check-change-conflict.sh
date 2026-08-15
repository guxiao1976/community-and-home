#!/usr/bin/env bash
# ============================================================
# check-change-conflict.sh — 需求冲突检测确定性脚本（非 LLM）
# ============================================================
# 扫描 .harness/changes/*/ 中两个变更的同服务 / 同接口（revises）重叠，
# 输出冲突预警清单，供需求分析（阶段 1）澄清前注入 —— 机械化替代 Step 0 的 LLM 判定。
# 参考 spec-pipeline P3.2 specDeterministicCheck 模式：便宜、客观、确定性。
#
# 用法:
#   bash .harness/scripts/check-change-conflict.sh                      # 扫描全部变更两两比对
#   bash .harness/scripts/check-change-conflict.sh --change <name>      # 只报 <name> 与其它变更的冲突
#   bash .harness/scripts/check-change-conflict.sh --json               # JSON 输出
#
# 判据（同一变更对存在任一即冲突）:
#   C1 同服务重叠  — 两个 .change.yaml 的 services 列表有交集
#   C2 同接口重叠  — 两个 .change.yaml 的 revises 列表有交集（同文件被改）
#
# 退出码: 0（预警非阻塞，供注入参考；与 specDeterministicCheck 一致，发现不 fail）
# ============================================================

set -uo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHANGES_DIR="$PROJECT_ROOT/.harness/changes"
TARGET_CHANGE=""
JSON_OUT=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --change) TARGET_CHANGE="$2"; shift 2 ;;
    --json) JSON_OUT=true; shift ;;
    -h|--help) sed -n '1,25p' "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "未知参数: $1"; exit 2 ;;
  esac
done

# ── 收集变更: 目录名 + 解析 .change.yaml 的 services/revises ──
declare -a change_names=()
declare -A svc_of=()   # change -> "svc1 svc2 ..."
declare -A rev_of=()   # change -> "file1 file2 ..."

collect_changes() {
  local d
  for d in "$CHANGES_DIR"/*/; do
    local name; name="$(basename "$d")"
    # 跳过基础设施目录
    [[ "$name" == "_archive" || "$name" == "TEMPLATE" || "$name" == "README" ]] && continue
    local yaml="$d/.change.yaml"
    [[ -f "$yaml" ]] || continue
    change_names+=("$name")
    # services 段: 缩进的 "- xxx"
    svc_of["$name"]="$(awk '/^services:/{ins=1;next} ins && /^  - /{sub(/^  - /,"");printf "%s ",$0;next} /^[a-zA-Z_]+:/{ins=0}' "$yaml")"
    # revises 段: 同样 "- xxx"；只取行首路径 token（空格前的部分，滤掉 §x.x 章节引用），
    # 且只认含 / 或 . 的真实文件路径
    rev_of["$name"]="$(awk '/^revises:/{ins=1;next} ins && /^  - /{sub(/^  - /,""); f=$1; if (f ~ /[\/.]/) printf "%s ",f; next} /^[a-zA-Z_]+:/{ins=0}' "$yaml")"
  done
}

# ── 冲突判定 ──
declare -a findings=()

find_conflicts() {
  local i j a b svc_a svc_b rev_a rev_b s f
  for ((i=0; i<${#change_names[@]}; i++)); do
    a="${change_names[$i]}"
    [[ -n "$TARGET_CHANGE" && "$a" != "$TARGET_CHANGE" ]] && continue
    for ((j=i+1; j<${#change_names[@]}; j++)); do
      b="${change_names[$j]}"
      svc_a="${svc_of[$a]}"; svc_b="${svc_of[$b]}"
      rev_a="${rev_of[$a]}"; rev_b="${rev_of[$b]}"

      # C1 同服务重叠
      for s in $svc_a; do
        if [[ " $svc_b " == *" $s "* ]]; then
          findings+=("C1|$a|$b|服务重叠: $s")
        fi
      done
      # C2 同接口/同文件重叠
      for f in $rev_a; do
        if [[ " $rev_b " == *" $f "* ]]; then
          findings+=("C2|$a|$b|接口/文件重叠: $f")
        fi
      done
    done
  done
}

# ── 输出 ──
emit() {
  if $JSON_OUT; then
    local first=true
    echo '['
    for f in "${findings[@]}"; do
      $first || echo ','
      first=false
      IFS='|' read -r kind a b msg <<< "$f"
      printf '  {"type":"%s","change_a":"%s","change_b":"%s","detail":"%s"}' "$kind" "$a" "$b" "$msg"
    done
    echo ''
    echo ']'
  else
    if [[ ${#findings[@]} -eq 0 ]]; then
      echo "✅ 无变更冲突（${#change_names[@]} 个变更两两比对）"
    else
      echo "⚠️ 检测到 ${#findings[@]} 个变更冲突："
      for f in "${findings[@]}"; do
        IFS='|' read -r kind a b msg <<< "$f"
        echo "  [$kind] $a ↔ $b — $msg"
      done
    fi
  fi
}

collect_changes
find_conflicts
emit

exit 0
