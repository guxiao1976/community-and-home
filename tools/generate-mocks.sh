#!/usr/bin/env bash
#
# generate-mocks.sh — 批量生成 Mock 接口
#
# Usage:
#   bash tools/generate-mocks.sh [service-name]
#
# Examples:
#   bash tools/generate-mocks.sh user-service
#   bash tools/generate-mocks.sh all
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SERVICE="${1:-}"

generate_mock_for_service() {
    local svc_name="$1"
    local svc_dir="$PROJECT_ROOT/services/$svc_name"

    if [[ ! -d "$svc_dir" ]]; then
        echo "❌ 服务目录不存在: $svc_dir"
        return 1
    fi

    echo "📦 生成 $svc_name 的 Mock..."

    # 创建 mocks 目录
    mkdir -p "$svc_dir/model/mocks"
    mkdir -p "$svc_dir/rpc/mocks"

    # 查找所有 *model_gen.go 文件
    local model_files
    model_files=$(find "$svc_dir/model" -name "*model_gen.go" 2>/dev/null || true)

    if [[ -z "$model_files" ]]; then
        echo "  ⚠️  未找到 *model_gen.go 文件"
    else
        for model_file in $model_files; do
            local base_name=$(basename "$model_file" .go)
            local mock_file="$svc_dir/model/mocks/${base_name}_mock.go"

            echo "  生成: $(basename "$mock_file")"
            mockgen -source="$model_file" \
                -destination="$mock_file" \
                -package=mocks 2>/dev/null || {
                echo "  ⚠️  跳过: $model_file (可能没有接口)"
            }
        done
    fi

    echo "  ✅ $svc_name Mock 生成完成"
    echo ""
}

if [[ -z "$SERVICE" ]]; then
    echo "Usage: $0 <service-name|all>"
    echo ""
    echo "Available services:"
    ls -1 "$PROJECT_ROOT/services" | grep -v "^_" | sed 's/^/  - /'
    exit 1
fi

if [[ "$SERVICE" == "all" ]]; then
    echo "🚀 批量生成所有服务的 Mock..."
    echo ""

    for svc_dir in "$PROJECT_ROOT/services"/*; do
        [[ ! -d "$svc_dir" ]] && continue
        svc_name=$(basename "$svc_dir")
        [[ "$svc_name" == "_"* ]] && continue

        generate_mock_for_service "$svc_name"
    done

    echo "✅ 所有服务 Mock 生成完成"
else
    generate_mock_for_service "$SERVICE"
fi
