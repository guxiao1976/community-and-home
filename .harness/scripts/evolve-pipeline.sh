#!/bin/bash
# ============================================================
# evolve-pipeline.sh — 从 Incident 记录分析管道弱点并提议改进
# ============================================================
# 用法: bash .harness/scripts/evolve-pipeline.sh
#
# 扫描 .harness/logs/incidents/*.yml → 按 pattern 聚合 →
# 检查当前 gate-engine.js 是否已覆盖 → 输出改进建议

set -e

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
INCIDENT_DIR="$ROOT/.harness/logs/incidents"
GATE_FILE="$ROOT/.harness/workflows/gate-engine.js"

echo "╔══════════════════════════════════════════════════════════╗"
echo "║  Pipeline Evolution Analyzer                             ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

# ── 统计 incident 数量 ────────────────────────────────────
count=$(ls "$INCIDENT_DIR"/*.yml 2>/dev/null | grep -v _template | wc -l)
echo "📊 Incident 记录数: $count"
echo ""

if [ "$count" -eq 0 ]; then
  echo "✅ 暂无 Incident 记录，管道无需进化。"
  exit 0
fi

# ── 统计 pattern 出现次数（YAML 行格式: "  - run_old_code  # 注释"）────
echo "🔍 问题模式分布:"
echo ""

declare -A pattern_count
for f in "$INCIDENT_DIR"/*.yml; do
  [ "$(basename "$f")" = "_template.yml" ] && continue
  while IFS= read -r line; do
    # 提取 "  - pattern_name" 中的 pattern_name
    pattern=$(echo "$line" | sed -n 's/^[[:space:]]*-[[:space:]]*\([a-z_]*\).*/\1/p')
    [ -z "$pattern" ] && continue
    # 跳过 YAML 关键字
    [[ "$pattern" == "date" || "$pattern" == "task" || "$pattern" == "pain_rounds" ]] && continue
    pattern_count["$pattern"]=$((${pattern_count["$pattern"]:-0} + 1))
  done < <(grep "^\s*- " "$f")
done

for p in "${!pattern_count[@]}"; do
  printf "  %-35s %d 次\n" "$p" "${pattern_count[$p]}"
done

echo ""

# ── 检查 gate-engine.js 是否已有对应门禁 ───────────────────
echo "🔬 门禁覆盖分析:"
echo ""

declare -A gate_map
gate_map["run_old_code"]="verify_process_fresh"
gate_map["process_not_restarted"]="verify_process_fresh"
gate_map["no_self_verify"]="Skill: verify-before-deliver"
gate_map["compile_pass_but_runtime_fail"]="verify_compile"
gate_map["db_cache_stale"]="verify_db_no_duplicates"
gate_map["config_missing"]="verify_compile (间接)"
gate_map["type_mismatch"]="verify_compile (间接)"
gate_map["duplicated_effort"]="verify_no_panic_in_logs"
gate_map["missing_gate"]="gate-engine 扩展机制"

has_suggestion=false
for p in "${!pattern_count[@]}"; do
  count=${pattern_count[$p]}
  gate="${gate_map[$p]:-未覆盖}"

  if [ "$gate" = "未覆盖" ]; then
    echo "  ⚠️  $p ($count 次) → 无对应门禁 — 建议新增"
    has_suggestion=true
  else
    echo "  ✅ $p ($count 次) → $gate"
  fi
done

echo ""

# ── 汇总建议 ───────────────────────────────────────────────
echo "💡 改进建议:"
echo ""

suggestion_count=0

for p in "${!pattern_count[@]}"; do
  count=${pattern_count[$p]}
  gate="${gate_map[$p]:-}"

  case "$gate" in
    "verify_process_fresh")
      if [ "$count" -ge 3 ]; then
        echo "  [$((++suggestion_count))] verify_process_fresh 累计 $count 次 → 建议从 WARN 升级为 BLOCK"
      fi
      ;;
    "未覆盖")
      echo "  [$((++suggestion_count))] $p ($count 次) → 在 gate-engine.js 新增对应门禁"
      ;;
  esac
done

if [ "$suggestion_count" -eq 0 ]; then
  echo "  ✅ 当前门禁已覆盖所有已知问题模式，无需改进。"
fi

echo ""
echo "══════════════════════════════════════════════════════════"
echo "  执行: 手动修改 $GATE_FILE"
echo "  记录: $INCIDENT_DIR/"
echo "══════════════════════════════════════════════════════════"
