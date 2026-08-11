#!/usr/bin/env bash
#
# tdd-evidence-validator.sh — TDD 证据验证工具
#
# 用途：验证 QA 报告中的 TDD 证据是否包含具体的 RED FAIL 输出
#
# Usage:
#   bash .harness/scripts/tdd-evidence-validator.sh <qa-report.md>
#
# Exit code: 0 if valid, 1 if evidence insufficient
#

set -euo pipefail

QA_FILE="${1:-}"

if [[ -z "$QA_FILE" ]]; then
  echo "Usage: $0 <qa-report.md>"
  exit 2
fi

if [[ ! -f "$QA_FILE" ]]; then
  echo "Error: file not found: $QA_FILE"
  exit 2
fi

echo "🔍 验证 TDD 证据: $QA_FILE"
echo ""

# 检查是否包含 TDD 证据章节
if ! grep -q "## TDD 证据检查" "$QA_FILE"; then
  echo "❌ 缺少 TDD 证据检查章节"
  exit 1
fi

# 提取 TDD 证据表格
TDD_SECTION=$(sed -n '/## TDD 证据检查/,/^##/p' "$QA_FILE" | head -n -1)

# 检查是否有表格
if ! echo "$TDD_SECTION" | grep -q "|.*|.*|.*|"; then
  echo "❌ TDD 证据章节存在但无表格"
  exit 1
fi

# 解析表格行（跳过表头和分隔符）
TABLE_ROWS=$(echo "$TDD_SECTION" | grep "|" | tail -n +3)

if [[ -z "$TABLE_ROWS" ]]; then
  echo "✅ 无新增函数需要 TDD 证据"
  exit 0
fi

INVALID_COUNT=0
VALID_COUNT=0

echo "$TABLE_ROWS" | while IFS='|' read -r _ func_name _ red_confirm _ green_confirm _ status _; do
  # 清理空格
  func_name=$(echo "$func_name" | xargs)
  red_confirm=$(echo "$red_confirm" | xargs)

  [[ -z "$func_name" ]] && continue

  # 检查 RED 确认列是否包含具体 FAIL 输出
  # 有效格式：包含 "FAIL:" 或 "undefined:" 或 "error:" 等实际错误信息
  # 无效格式："看到失败"、"见测试日志"、空白、"✅"

  if [[ "$red_confirm" == *"FAIL"* ]] || [[ "$red_confirm" == *"undefined"* ]] || [[ "$red_confirm" == *"error"* ]] || [[ "$red_confirm" == *"not found"* ]]; then
    echo "✅ $func_name: RED 证据有效 ($red_confirm)"
    VALID_COUNT=$((VALID_COUNT + 1))
  elif [[ "$red_confirm" == "✅"* ]] && [[ "$red_confirm" != *"FAIL"* ]]; then
    # 只有勾号但没有具体错误信息
    echo "❌ $func_name: RED 证据不足 (缺少具体 FAIL 输出)"
    echo "   当前: $red_confirm"
    echo "   需要: 包含实际错误信息，如 'FAIL: undefined: $func_name'"
    INVALID_COUNT=$((INVALID_COUNT + 1))
  elif [[ "$red_confirm" == *"见测试"* ]] || [[ "$red_confirm" == *"看到"* ]]; then
    echo "❌ $func_name: RED 证据不足 (描述性文字，缺少实际输出)"
    echo "   当前: $red_confirm"
    echo "   需要: 粘贴实际的 FAIL 输出行"
    INVALID_COUNT=$((INVALID_COUNT + 1))
  else
    echo "⚠️  $func_name: RED 确认格式不明确"
    echo "   当前: $red_confirm"
  fi
done

# 读取子shell中的计数（使用临时文件）
TMPFILE=$(mktemp)
echo "$TABLE_ROWS" | while IFS='|' read -r _ func_name _ red_confirm _ green_confirm _ status _; do
  func_name=$(echo "$func_name" | xargs)
  red_confirm=$(echo "$red_confirm" | xargs)

  [[ -z "$func_name" ]] && continue

  if [[ "$red_confirm" == *"FAIL"* ]] || [[ "$red_confirm" == *"undefined"* ]] || [[ "$red_confirm" == *"error"* ]] || [[ "$red_confirm" == *"not found"* ]]; then
    echo "VALID" >> "$TMPFILE"
  elif [[ "$red_confirm" == "✅"* ]] && [[ "$red_confirm" != *"FAIL"* ]]; then
    echo "INVALID" >> "$TMPFILE"
  elif [[ "$red_confirm" == *"见测试"* ]] || [[ "$red_confirm" == *"看到"* ]]; then
    echo "INVALID" >> "$TMPFILE"
  fi
done

VALID_COUNT=$(grep -c "VALID" "$TMPFILE" 2>/dev/null || echo 0)
INVALID_COUNT=$(grep -c "INVALID" "$TMPFILE" 2>/dev/null || echo 0)
rm -f "$TMPFILE"

echo ""
echo "总计: $VALID_COUNT 个有效, $INVALID_COUNT 个不足"

if [[ $INVALID_COUNT -gt 0 ]]; then
  echo ""
  echo "❌ TDD 证据验证失败"
  echo ""
  echo "修复建议："
  echo "1. 在 Generator 阶段输出 TDD 证据到 console"
  echo "2. QA 阶段从 console 提取实际的 FAIL 输出"
  echo "3. 在 TDD 证据表格的 RED 确认列中粘贴实际错误信息"
  echo ""
  echo "示例："
  echo "  ❌ 错误: '✅ 见测试日志 PASS'"
  echo "  ✅ 正确: '✅ 看到 FAIL: undefined: CheckTextWithPipeline'"
  exit 1
else
  echo "✅ TDD 证据验证通过"
  exit 0
fi
