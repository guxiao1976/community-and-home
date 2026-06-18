#!/usr/bin/env bash
#
# harness-tasks-deps.sh — 任务依赖分析（拓扑排序）
#
# 功能：
#   - 解析 blocked_by 关系
#   - 构建任务 DAG
#   - 拓扑排序（Kahn 算法）
#   - 检测循环依赖
#
# 用法：
#   bash .harness/scripts/harness-tasks-deps.sh --order-by-deps
#   bash .harness/scripts/harness-tasks-deps.sh --check-cycles
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TASKS_DIR="$PROJECT_ROOT/.harness/tasks"

# ─── 解析依赖关系 ───
parse_dependencies() {
    declare -A GRAPH
    declare -A IN_DEGREE

    # 遍历所有 open 任务
    for task_file in "$TASKS_DIR"/task-*.md; do
        [[ ! -f "$task_file" ]] && continue

        STATUS=$(grep "^status:" "$task_file" | awk '{print $2}' || echo "open")
        [[ "$STATUS" != "open" ]] && continue

        TASK_ID=$(basename "$task_file" .md)
        IN_DEGREE[$TASK_ID]=0
        GRAPH[$TASK_ID]=""
    done

    # 解析 blocked_by
    for task_file in "$TASKS_DIR"/task-*.md; do
        [[ ! -f "$task_file" ]] && continue

        STATUS=$(grep "^status:" "$task_file" | awk '{print $2}' || echo "open")
        [[ "$STATUS" != "open" ]] && continue

        TASK_ID=$(basename "$task_file" .md)

        # 提取 blocked_by 列表
        BLOCKED_BY=$(grep "^blocked_by:" "$task_file" | sed 's/blocked_by: \[\(.*\)\]/\1/' | tr ',' '\n' | tr -d ' "' || echo "")

        for dep in $BLOCKED_BY; do
            [[ -z "$dep" ]] && continue

            # dep → TASK_ID（dep 阻塞 TASK_ID）
            if [[ -n "${GRAPH[$dep]:-}" ]]; then
                GRAPH[$dep]="${GRAPH[$dep]} $TASK_ID"
            else
                GRAPH[$dep]="$TASK_ID"
            fi

            # TASK_ID 的入度 +1
            IN_DEGREE[$TASK_ID]=$((IN_DEGREE[$TASK_ID] + 1))
        done
    done

    # 导出为全局变量（Bash 4.3+ 支持 nameref）
    for key in "${!GRAPH[@]}"; do
        echo "GRAPH[$key]=${GRAPH[$key]}"
    done

    for key in "${!IN_DEGREE[@]}"; do
        echo "IN_DEGREE[$key]=${IN_DEGREE[$key]}"
    done
}

# ─── 拓扑排序（Kahn 算法）───
topological_sort() {
    # 重新解析（因为 Bash 关联数组无法直接传递）
    declare -A GRAPH
    declare -A IN_DEGREE
    declare -a QUEUE
    declare -a RESULT

    # 解析依赖
    while IFS= read -r line; do
        if [[ "$line" =~ GRAPH\[(.*)\]=(.*) ]]; then
            GRAPH["${BASH_REMATCH[1]}"]="${BASH_REMATCH[2]}"
        elif [[ "$line" =~ IN_DEGREE\[(.*)\]=(.*) ]]; then
            IN_DEGREE["${BASH_REMATCH[1]}"]="${BASH_REMATCH[2]}"
        fi
    done < <(parse_dependencies)

    # 找到所有入度为 0 的节点
    for task_id in "${!IN_DEGREE[@]}"; do
        if [[ ${IN_DEGREE[$task_id]} -eq 0 ]]; then
            QUEUE+=("$task_id")
        fi
    done

    # Kahn 算法
    while [[ ${#QUEUE[@]} -gt 0 ]]; do
        # 出队
        local current="${QUEUE[0]}"
        QUEUE=("${QUEUE[@]:1}")
        RESULT+=("$current")

        # 减少邻居的入度
        for neighbor in ${GRAPH[$current]:-}; do
            IN_DEGREE[$neighbor]=$((IN_DEGREE[$neighbor] - 1))

            if [[ ${IN_DEGREE[$neighbor]} -eq 0 ]]; then
                QUEUE+=("$neighbor")
            fi
        done
    done

    # 检查是否有环
    if [[ ${#RESULT[@]} -ne ${#IN_DEGREE[@]} ]]; then
        echo "❌ 检测到循环依赖！" >&2
        echo "未处理的任务:" >&2
        for task_id in "${!IN_DEGREE[@]}"; do
            if [[ ${IN_DEGREE[$task_id]} -gt 0 ]]; then
                echo "  - $task_id (入度: ${IN_DEGREE[$task_id]})" >&2
            fi
        done
        return 1
    fi

    # 输出拓扑排序结果
    for task_id in "${RESULT[@]}"; do
        echo "$task_id"
    done
}

# ─── 检测循环依赖 ───
check_cycles() {
    echo "🔍 检查循环依赖..."

    if topological_sort > /dev/null 2>&1; then
        echo "✅ 无循环依赖"
        return 0
    else
        echo "❌ 发现循环依赖"
        return 1
    fi
}

# ─── 主逻辑 ───
case "${1:-help}" in
    --order-by-deps)
        topological_sort
        ;;

    --check-cycles)
        check_cycles
        ;;

    --help|-h)
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  --order-by-deps   按依赖关系拓扑排序"
        echo "  --check-cycles    检测循环依赖"
        echo "  --help            显示帮助"
        exit 0
        ;;

    *)
        echo "未知选项: $1"
        echo "运行 $0 --help 查看用法"
        exit 2
        ;;
esac
