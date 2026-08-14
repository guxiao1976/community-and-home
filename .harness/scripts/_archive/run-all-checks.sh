#!/usr/bin/env bash
#
# run-all-checks.sh — 运行所有检查（集成版本）
#
# 功能：
#   - 运行 harness-checks.sh（Go 后端检查）
#   - 运行 frontend-checks.sh（前端扩展检查）
#   - 汇总所有结果
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PASS=0
FAIL=0
WARN=0

echo "╔════════════════════════════════════════════════════════════╗"
echo "║              🔍 完整检查套件 - 后端+前端                   ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# ═══ 运行后端检查 ═══
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣ 后端检查（harness-checks.sh）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if bash "$PROJECT_ROOT/.harness/skills/qa/scripts/harness-checks.sh" > /tmp/backend-checks.log 2>&1; then
    echo "✅ 后端检查通过"

    # 统计结果
    BACKEND_PASS=$(grep -c "\[PASS\]" /tmp/backend-checks.log || echo "0")
    BACKEND_FAIL=$(grep -c "\[FAIL\]" /tmp/backend-checks.log || echo "0")
    BACKEND_WARN=$(grep -c "\[WARN\]" /tmp/backend-checks.log || echo "0")

    echo "  PASS: $BACKEND_PASS | FAIL: $BACKEND_FAIL | WARN: $BACKEND_WARN"

    PASS=$((PASS + BACKEND_PASS))
    FAIL=$((FAIL + BACKEND_FAIL))
    WARN=$((WARN + BACKEND_WARN))
else
    echo "⚠️  后端检查有失败项"

    # 显示失败项
    grep "\[FAIL\]" /tmp/backend-checks.log | head -5

    BACKEND_FAIL=$(grep -c "\[FAIL\]" /tmp/backend-checks.log || echo "0")
    FAIL=$((FAIL + BACKEND_FAIL))
fi
echo ""

# ═══ 运行前端检查 ═══
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣ 前端检查（frontend-checks.sh）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if bash "$PROJECT_ROOT/.harness/scripts/frontend-checks.sh" > /tmp/frontend-checks.log 2>&1; then
    echo "✅ 前端检查通过"

    FRONTEND_PASS=$(grep -c "\[PASS\]" /tmp/frontend-checks.log || echo "0")
    FRONTEND_FAIL=$(grep -c "\[FAIL\]" /tmp/frontend-checks.log || echo "0")
    FRONTEND_WARN=$(grep -c "\[WARN\]" /tmp/frontend-checks.log || echo "0")

    echo "  PASS: $FRONTEND_PASS | FAIL: $FRONTEND_FAIL | WARN: $FRONTEND_WARN"

    PASS=$((PASS + FRONTEND_PASS))
    FAIL=$((FAIL + FRONTEND_FAIL))
    WARN=$((WARN + FRONTEND_WARN))
else
    echo "⚠️  前端检查有失败项"

    grep "\[FAIL\]" /tmp/frontend-checks.log | head -5

    FRONTEND_FAIL=$(grep -c "\[FAIL\]" /tmp/frontend-checks.log || echo "0")
    FAIL=$((FAIL + FRONTEND_FAIL))
fi
echo ""

# ═══ 汇总结果 ═══
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              📊 检查结果汇总                                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "总计:"
echo "  ✅ PASS: $PASS"
echo "  ❌ FAIL: $FAIL"
echo "  ⚠️  WARN: $WARN"
echo ""

TOTAL=$((PASS + FAIL + WARN))
if [ $TOTAL -gt 0 ]; then
    PASS_RATE=$((PASS * 100 / TOTAL))
    echo "通过率: $PASS_RATE%"
fi
echo ""

if [ $FAIL -eq 0 ]; then
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║          ✅ 所有检查通过！                                 ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    exit 0
else
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║          ⚠️  发现 $FAIL 个失败项                                ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    exit 1
fi
