#!/usr/bin/env bash
#
# harness-tasks-score.sh — 任务评分系统（加权优先级）
#
# 功能：
#   - 计算任务优先级评分
#   - 考虑 priority、age、urgency、dependency
#   - 更新 BACKLOG.md 排序
#
# 用法：
#   bash .harness/scripts/harness-tasks-score.sh [--update-index]
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TASKS_DIR="$PROJECT_ROOT/.harness/tasks"
BACKLOG="$TASKS_DIR/BACKLOG.md"

UPDATE_INDEX=false

# ─── 解析参数 ───
while [[ $# -gt 0 ]]; do
    case $1 in
        --update-index)
            UPDATE_INDEX=true
            shift
            ;;
        --help|-h)
            echo "用法: $0 [--update-index]"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            exit 2
            ;;
    esac
done

# ─── 评分公式 ───
# score = priority_weight × (1 + age_factor) × urgency_multiplier × dependency_boost

calculate_score() {
    local task_file=$1

    # 提取字段
    local priority=$(grep "^priority:" "$task_file" | awk '{print $2}' || echo "P2")
    local created=$(grep "^created:" "$task_file" | awk '{print $2}' || echo "$(date +%Y-%m-%d)")
    local source=$(grep "^source:" "$task_file" | awk '{print $2}' || echo "human")
    local severity=$(grep "^severity:" "$task_file" | awk '{print $2}' || echo "")

    # Priority weight
    case "$priority" in
        P0) priority_weight=100 ;;
        P1) priority_weight=50 ;;
        P2) priority_weight=20 ;;
        P3) priority_weight=10 ;;
        *) priority_weight=20 ;;
    esac

    # Age factor (min(age_days × 0.05, 2.0))
    local age_days=$(( ($(date +%s) - $(date -d "$created" +%s)) / 86400 ))
    local age_factor=$(echo "scale=2; a = $age_days * 0.05; if (a > 2.0) 2.0 else a" | bc)

    # Urgency multiplier
    local urgency_multiplier=1.0
    if [[ "$source" == "qa" || "$source" == "review" ]] && [[ "$severity" == "CRITICAL" ]]; then
        urgency_multiplier=1.5
    elif [[ "$source" == "sensor" && "$severity" == "security" ]]; then
        urgency_multiplier=2.0
    fi

    # Dependency boost (统计有多少任务 blocked_by 该任务)
    local task_id=$(basename "$task_file" .md)
    local blocked_count=$(grep -r "blocked_by:.*$task_id" "$TASKS_DIR" 2>/dev/null | wc -l || echo "0")
    local dependency_boost=1.0
    if [ "$blocked_count" -gt 0 ]; then
        dependency_boost=1.2
    fi

    # 计算最终 score
    local score=$(echo "scale=1; $priority_weight * (1 + $age_factor) * $urgency_multiplier * $dependency_boost" | bc)

    echo "$score"
}

# ─── 处理所有 open 任务 ───
echo "🔢 计算任务优先级评分..."

TOTAL=0

for task_file in "$TASKS_DIR"/task-*.md; do
    [[ ! -f "$task_file" ]] && continue

    # 只处理 open 状态的任务
    STATUS=$(grep "^status:" "$task_file" | awk '{print $2}' || echo "open")
    [[ "$STATUS" != "open" ]] && continue

    SCORE=$(calculate_score "$task_file")

    # 写入 computed_score 到任务文件
    if grep -q "^computed_score:" "$task_file"; then
        # 更新现有值
        sed -i "s/^computed_score:.*/computed_score: $SCORE/" "$task_file"
    else
        # 插入新字段（在 status 之后）
        sed -i "/^status:/a computed_score: $SCORE" "$task_file"
    fi

    TOTAL=$((TOTAL + 1))

    TASK_ID=$(basename "$task_file" .md)
    PRIORITY=$(grep "^priority:" "$task_file" | awk '{print $2}')
    echo "  ✓ $TASK_ID ($PRIORITY) → score=$SCORE"
done

echo ""
echo "✅ 已计算 $TOTAL 个 open 任务的评分"

# ─── 更新 BACKLOG.md ───
if $UPDATE_INDEX; then
    echo ""
    echo "📝 更新 BACKLOG.md 排序..."

    # 重建 BACKLOG（按 score 降序）
    {
        echo "# Task Backlog"
        echo ""
        echo "按优先级评分排序（score = priority × age × urgency × dependency）"
        echo ""
        echo "| Task ID | Priority | Score | Status | Title |"
        echo "|---------|----------|-------|--------|-------|"

        # 遍历所有 open 任务，按 score 排序
        for task_file in "$TASKS_DIR"/task-*.md; do
            [[ ! -f "$task_file" ]] && continue

            STATUS=$(grep "^status:" "$task_file" | awk '{print $2}' || echo "open")
            [[ "$STATUS" != "open" ]] && continue

            TASK_ID=$(basename "$task_file" .md)
            PRIORITY=$(grep "^priority:" "$task_file" | awk '{print $2}')
            SCORE=$(grep "^computed_score:" "$task_file" | awk '{print $2}' || echo "0")
            TITLE=$(grep "^# " "$task_file" | head -1 | sed 's/^# //')

            echo "$SCORE|$TASK_ID|$PRIORITY|$STATUS|$TITLE"
        done | sort -t'|' -k1 -rn | awk -F'|' '{printf "| %s | %s | %s | %s | %s |\n", $2, $3, $1, $4, $5}'

    } > "$BACKLOG"

    echo "✅ BACKLOG.md 已更新"
fi

exit 0
