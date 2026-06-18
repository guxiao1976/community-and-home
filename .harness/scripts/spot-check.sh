#!/usr/bin/env bash
#
# spot-check.sh — 抽查变更文件（用于中置信度审查）
#
# 功能：
#   - 从 git diff 获取变更文件列表
#   - 随机抽样 max(2, total × 30%)
#   - 种子可复现（基于任务 ID hash）
#
# 用法：
#   bash .harness/scripts/spot-check.sh [--seed <task-id>] [--ratio 0.3] [--min 2]
#   bash .harness/scripts/spot-check.sh --service moderation-service --seed task-001
#
# 输出：
#   抽样文件列表（每行一个文件路径）
#

set -euo pipefail

# ─── 配置 ───
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

SEED=""
RATIO=0.3
MIN_SAMPLES=2
SERVICE=""
BASE_REF="HEAD"

# ─── 解析参数 ───
while [[ $# -gt 0 ]]; do
    case $1 in
        --seed)
            SEED="$2"
            shift 2
            ;;
        --ratio)
            RATIO="$2"
            shift 2
            ;;
        --min)
            MIN_SAMPLES="$2"
            shift 2
            ;;
        --service)
            SERVICE="$2"
            shift 2
            ;;
        --base)
            BASE_REF="$2"
            shift 2
            ;;
        --help|-h)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --seed <id>       种子（确保可复现），如任务 ID"
            echo "  --ratio <float>   抽样比例（默认 0.3 = 30%）"
            echo "  --min <int>       最小抽样数（默认 2）"
            echo "  --service <name>  只抽查特定服务目录"
            echo "  --base <ref>      对比基准（默认 HEAD）"
            echo "  --help            显示帮助"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            exit 2
            ;;
    esac
done

# ─── 获取变更文件 ───
cd "$PROJECT_ROOT"

# 获取变更文件列表（排除删除的文件）
if [ -n "$SERVICE" ]; then
    # 只看特定服务
    CHANGED_FILES=$(git diff --name-only --diff-filter=ACM "$BASE_REF" | grep "^services/$SERVICE/" || true)
else
    # 所有服务
    CHANGED_FILES=$(git diff --name-only --diff-filter=ACM "$BASE_REF" | grep -E "^services/.*\.(go|ts|vue)$" || true)
fi

if [ -z "$CHANGED_FILES" ]; then
    echo "# 无变更文件" >&2
    exit 0
fi

# 统计文件数
TOTAL=$(echo "$CHANGED_FILES" | wc -l)

if [ "$TOTAL" -eq 0 ]; then
    echo "# 无变更文件" >&2
    exit 0
fi

# ─── 计算抽样数量 ───
SAMPLE_SIZE=$(echo "$TOTAL * $RATIO" | bc | awk '{print int($1+0.5)}')  # 四舍五入

if [ "$SAMPLE_SIZE" -lt "$MIN_SAMPLES" ]; then
    SAMPLE_SIZE=$MIN_SAMPLES
fi

if [ "$SAMPLE_SIZE" -gt "$TOTAL" ]; then
    SAMPLE_SIZE=$TOTAL
fi

echo "# 变更文件: $TOTAL, 抽样: $SAMPLE_SIZE (ratio=$RATIO, min=$MIN_SAMPLES)" >&2

# ─── 随机抽样 ───
if [ -z "$SEED" ]; then
    # 无种子，真随机
    echo "$CHANGED_FILES" | shuf | head -n "$SAMPLE_SIZE"
else
    # 有种子，可复现（使用种子的 hash 作为 shuf seed）
    SEED_NUM=$(echo -n "$SEED" | md5sum | cut -c1-8 | tr 'a-f' '0-5')  # 转为数字
    SEED_NUM=$((16#$SEED_NUM))  # 16 进制转 10 进制

    echo "# 种子: $SEED (hash=$SEED_NUM)" >&2
    echo "$CHANGED_FILES" | shuf --random-source=<(echo $SEED_NUM) | head -n "$SAMPLE_SIZE"
fi

exit 0
