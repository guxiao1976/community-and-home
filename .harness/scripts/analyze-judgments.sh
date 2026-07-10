#!/usr/bin/env bash
# Analyze AI Judgment Logs
# Usage: bash analyze-judgments.sh [days]

set -euo pipefail

DAYS="${1:-7}"
LOG_DIR=".harness/logs/judgments"

if [ ! -d "$LOG_DIR" ]; then
  echo "No judgment logs found in $LOG_DIR"
  exit 0
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "AI Judgment Analysis (Last $DAYS days)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Find log files from last N days
LOG_FILES=$(find "$LOG_DIR" -name "*.json" -mtime -"$DAYS" | sort)

if [ -z "$LOG_FILES" ]; then
  echo "No logs found in the last $DAYS days"
  exit 0
fi

# Merge all logs
ALL_LOGS=$(cat $LOG_FILES | jq -s 'add')

# Total judgments
TOTAL=$(echo "$ALL_LOGS" | jq 'length')
echo "📊 Total Judgments: $TOTAL"
echo ""

# AI Status Distribution
echo "🤖 AI Judgment Distribution:"
AI_PASS=$(echo "$ALL_LOGS" | jq '[.[] | select(.original_ai_status == "PASS")] | length')
AI_FAIL=$(echo "$ALL_LOGS" | jq '[.[] | select(.original_ai_status == "FAIL")] | length')
echo "  PASS: $AI_PASS ($((AI_PASS * 100 / TOTAL))%)"
echo "  FAIL: $AI_FAIL ($((AI_FAIL * 100 / TOTAL))%)"
echo ""

# Override Statistics
echo "🔄 Override Statistics:"
OVERRIDDEN=$(echo "$ALL_LOGS" | jq '[.[] | select(.overridden == true)] | length')
echo "  Total Overrides: $OVERRIDDEN ($((OVERRIDDEN * 100 / TOTAL))%)"

if [ "$OVERRIDDEN" -gt 0 ]; then
  OVERRIDE_TO_FAIL=$(echo "$ALL_LOGS" | jq '[.[] | select(.overridden == true and .final_status == "FAIL")] | length')
  OVERRIDE_TO_PASS=$(echo "$ALL_LOGS" | jq '[.[] | select(.overridden == true and .final_status == "PASS")] | length')
  echo "    → To FAIL: $OVERRIDE_TO_FAIL"
  echo "    → To PASS: $OVERRIDE_TO_PASS"
fi
echo ""

# Conflict Type Distribution
echo "⚠️  Conflict Type Distribution:"
CONFLICTS=$(echo "$ALL_LOGS" | jq '[.[] | .conflicts[]] | group_by(.type) | map({type: .[0].type, count: length}) | sort_by(.count) | reverse')

if [ "$(echo "$CONFLICTS" | jq 'length')" -gt 0 ]; then
  echo "$CONFLICTS" | jq -r '.[] | "  \(.type): \(.count)"'
else
  echo "  No conflicts detected"
fi
echo ""

# Human Review Required
echo "👤 Human Review:"
HUMAN_REVIEW=$(echo "$ALL_LOGS" | jq '[.[] | select(.human_review_required == true)] | length')
echo "  Required: $HUMAN_REVIEW ($((HUMAN_REVIEW * 100 / TOTAL))%)"
echo ""

# Accuracy Calculation
echo "🎯 AI Accuracy:"
# Accurate = No override + Override due to false positive warnings (not real errors)
ACCURATE=$(echo "$ALL_LOGS" | jq '[.[] | select(.overridden == false)] | length')
ACCURACY=$((ACCURATE * 100 / TOTAL))
echo "  Without Override: $ACCURATE/$TOTAL ($ACCURACY%)"
echo ""

# Deterministic Check Pass Rate
echo "✅ Deterministic Check Pass Rate:"
COMPILE_PASS=$(echo "$ALL_LOGS" | jq '[.[] | select(.deterministic_summary.compile == true)] | length')
TESTS_PASS=$(echo "$ALL_LOGS" | jq '[.[] | select(.deterministic_summary.tests == true)] | length')
DEPS_PASS=$(echo "$ALL_LOGS" | jq '[.[] | select(.deterministic_summary.dependencies == true)] | length')

echo "  Compile: $COMPILE_PASS/$TOTAL ($((COMPILE_PASS * 100 / TOTAL))%)"
echo "  Tests: $TESTS_PASS/$TOTAL ($((TESTS_PASS * 100 / TOTAL))%)"
echo "  Dependencies: $DEPS_PASS/$TOTAL ($((DEPS_PASS * 100 / TOTAL))%)"
echo ""

# Service Breakdown
echo "📦 By Service:"
SERVICES=$(echo "$ALL_LOGS" | jq -r '.[].service' | sort | uniq -c | sort -rn)
echo "$SERVICES" | awk '{printf "  %s: %d judgments\n", $2, $1}'
echo ""

# Recent Issues
echo "🔍 Recent Issues (Last 5):"
RECENT_ISSUES=$(echo "$ALL_LOGS" | jq -r '
  [.[] | select(.conflicts | length > 0)] |
  sort_by(.timestamp) |
  reverse |
  .[0:5] |
  .[] |
  "  [\(.timestamp | split("T")[0])] \(.service): \(.conflicts[0].message)"
')

if [ -n "$RECENT_ISSUES" ]; then
  echo "$RECENT_ISSUES"
else
  echo "  No recent issues"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Analysis complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
