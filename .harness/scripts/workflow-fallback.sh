#!/usr/bin/env bash
#
# workflow-fallback.sh — Workflow 降级决策引擎
#
# 功能：
#   - 检测 Workflow 工具可用性
#   - 根据场景选择执行模式（workflow/agent/inline）
#   - 实现熔断器机制
#
# 用法：
#   mode=$(bash .harness/scripts/workflow-fallback.sh select --urgency normal --services 3)
#   echo "Selected mode: $mode"
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# 熔断器状态文件
CIRCUIT_BREAKER_FILE="$PROJECT_ROOT/.harness/runtime/workflow-circuit-breaker.state"

# ─── 健康检查 ───
check_workflow_health() {
    # 简单测试：尝试执行最小 Workflow
    local test_script='export const meta = {name:"test"}; return {status:"ok"}'

    # TODO: 调用 Workflow 工具测试
    # 这里简化为检查最近的失败记录

    if [[ -f "$CIRCUIT_BREAKER_FILE" ]]; then
        local last_check=$(cat "$CIRCUIT_BREAKER_FILE" | jq -r '.last_check')
        local fail_count=$(cat "$CIRCUIT_BREAKER_FILE" | jq -r '.fail_count')
        local circuit_open=$(cat "$CIRCUIT_BREAKER_FILE" | jq -r '.circuit_open')

        # 如果熔断器打开且未超过 30 分钟，返回不可用
        if [[ "$circuit_open" == "true" ]]; then
            local now=$(date +%s)
            local last=$(date -d "$last_check" +%s 2>/dev/null || echo "0")
            local elapsed=$((now - last))

            if [[ $elapsed -lt 1800 ]]; then  # 30 分钟
                echo "unhealthy"
                return 1
            fi
        fi
    fi

    echo "healthy"
    return 0
}

# ─── 记录失败 ───
record_failure() {
    mkdir -p "$(dirname "$CIRCUIT_BREAKER_FILE")"

    local now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local fail_count=1
    local circuit_open=false

    if [[ -f "$CIRCUIT_BREAKER_FILE" ]]; then
        fail_count=$(cat "$CIRCUIT_BREAKER_FILE" | jq -r '.fail_count')
        fail_count=$((fail_count + 1))
    fi

    # 连续 3 次失败 → 打开熔断器
    if [[ $fail_count -ge 3 ]]; then
        circuit_open=true
    fi

    cat > "$CIRCUIT_BREAKER_FILE" <<EOF
{
  "last_check": "$now",
  "fail_count": $fail_count,
  "circuit_open": $circuit_open
}
EOF

    if $circuit_open; then
        echo "⚠️  Workflow 熔断器已打开（连续 $fail_count 次失败），30 分钟内降级到 Agent 模式" >&2
    fi
}

# ─── 重置熔断器 ───
reset_circuit_breaker() {
    if [[ -f "$CIRCUIT_BREAKER_FILE" ]]; then
        rm "$CIRCUIT_BREAKER_FILE"
    fi
}

# ─── 模式选择 ───
select_mode() {
    local urgency="${1:-normal}"  # normal/emergency
    local service_count="${2:-1}"
    local file_count="${3:-10}"

    # 紧急情况 + 单文件小改动 → inline
    if [[ "$urgency" == "emergency" && $file_count -le 10 ]]; then
        echo "inline"
        return 0
    fi

    # 检查 Workflow 健康度
    local health=$(check_workflow_health)

    if [[ "$health" == "healthy" ]]; then
        # Workflow 可用，正常模式
        echo "workflow"
    else
        # Workflow 不可用，降级到 Agent 串行
        echo "agent"
    fi
}

# ─── 主逻辑 ───
case "${1:-help}" in
    select)
        shift
        URGENCY="normal"
        SERVICE_COUNT=1
        FILE_COUNT=10

        while [[ $# -gt 0 ]]; do
            case $1 in
                --urgency)
                    URGENCY="$2"
                    shift 2
                    ;;
                --services)
                    SERVICE_COUNT="$2"
                    shift 2
                    ;;
                --files)
                    FILE_COUNT="$2"
                    shift 2
                    ;;
                *)
                    shift
                    ;;
            esac
        done

        select_mode "$URGENCY" "$SERVICE_COUNT" "$FILE_COUNT"
        ;;

    check)
        check_workflow_health
        ;;

    fail)
        record_failure
        ;;

    reset)
        reset_circuit_breaker
        echo "✅ 熔断器已重置"
        ;;

    *)
        echo "用法: $0 {select|check|fail|reset} [选项]"
        echo ""
        echo "命令:"
        echo "  select  选择执行模式（workflow/agent/inline）"
        echo "  check   检查 Workflow 健康度"
        echo "  fail    记录一次失败"
        echo "  reset   重置熔断器"
        exit 1
        ;;
esac
