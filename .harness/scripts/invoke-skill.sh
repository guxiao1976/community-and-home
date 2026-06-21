#!/bin/bash
# 技能集成包装脚本
# 用于在流水线中安全地调用技能，失败不阻塞主流程

set -e

SKILL_NAME="$1"
SKILL_ARGS="$2"
OUTPUT_FILE="${3:-.harness/tmp/skill-output.txt}"

# 创建临时目录
mkdir -p .harness/tmp

echo "=== 调用技能: $SKILL_NAME ===" | tee "$OUTPUT_FILE"
echo "参数: $SKILL_ARGS" | tee -a "$OUTPUT_FILE"
echo "" | tee -a "$OUTPUT_FILE"

# 尝试调用技能
if npx skills run "$SKILL_NAME" "$SKILL_ARGS" >> "$OUTPUT_FILE" 2>&1; then
    echo "✅ 技能 $SKILL_NAME 执行成功" | tee -a "$OUTPUT_FILE"
    exit 0
else
    echo "⚠️ 技能 $SKILL_NAME 执行失败，但不阻塞流水线" | tee -a "$OUTPUT_FILE"
    exit 0  # 返回 0 以不阻塞流水线
fi
