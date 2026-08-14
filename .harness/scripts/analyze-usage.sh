#!/usr/bin/env bash
# analyze-usage.sh — 聚合 .harness/logs/usage/ + pipeline/metrics 输出流水线运行分析报告
# 供 pipeline-review 维度 6 调用（流水线复盘第一手数据：脚本调用/命中率/QA/轮次/图谱）。
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
USAGE="$ROOT/.harness/logs/usage"
METRICS="$ROOT/.harness/logs/pipeline/metrics.jsonl"

echo "=== 流水线运行数据分析 ($(date +%F)) ==="

echo ""
echo "## 脚本调用频率"
total=0
for f in "$USAGE"/*.jsonl; do
  [ -f "$f" ] || continue
  n=$(wc -l < "$f" | xargs)
  total=$((total + n))
  echo "  $n  $(basename "$f")"
done
if [ "$total" -eq 0 ]; then
  echo "  （暂无数据：检查 config/tracking.yml enabled: true，或流水线尚未跑过开发任务）"
fi

echo ""
echo "## 知识检索命中率（knowledge-load）"
kl="$USAGE/knowledge-load.jsonl"
if [ -f "$kl" ]; then
  t=$(wc -l < "$kl" | xargs)
  z=$(jq -s '[.[] | select(.matched == "0")] | length' "$kl" 2>/dev/null || echo "?")
  if [ "$t" -gt 0 ] 2>/dev/null; then rate=$(( (t - z) * 100 / t )); else rate="?"; fi
  echo "  总调用 $t | 命中 0 条记忆 $z 次（命中率 ${rate}%）"
else echo "  无数据"; fi

echo ""
echo "## QA 门禁（harness-checks）"
hc="$USAGE/harness-checks.jsonl"
if [ -f "$hc" ]; then
  t=$(wc -l < "$hc" | xargs)
  f_cnt=$(jq -s '[.[] | select((.fail|tonumber) > 0)] | length' "$hc" 2>/dev/null || echo "?")
  echo "  总运行 $t | 有 FAIL $f_cnt 次"
else echo "  无数据"; fi

echo ""
echo "## pipeline 轮次分布（metrics）"
if [ -f "$METRICS" ]; then
  t=$(wc -l < "$METRICS" | xargs)
  multi=$(jq -s '[.[] | select(.iterations > 1)] | length' "$METRICS" 2>/dev/null || echo "?")
  echo "  总完成 $t | 轮次>1（重试）$multi 次"
else echo "  无数据"; fi

echo ""
echo "## 图谱查询（graph-query）"
gq="$USAGE/graph-query.jsonl"
if [ -f "$gq" ]; then
  t=$(wc -l < "$gq" | xargs)
  ua=$(jq -s '[.[] | select(.result == "unavailable")] | length' "$gq" 2>/dev/null || echo "?")
  echo "  总查询 $t | Neo4j 不可用 $ua 次"
else echo "  无数据"; fi

echo ""
echo "## 建议：数据量 <3 条视为积累期（暂不评价）；有数据时基于数字判断："
echo "  - 命中率低（0 命中占比高）→ 记忆 triggers/关键词匹配需优化"
echo "  - QA 反复 FAIL → 对应检查项或代码习惯问题"
echo "  - 轮次>1 多 → 需求/设计阶段问题（下游返工）"
echo "  - 图谱不可用多 → Neo4j 可用性关注"
