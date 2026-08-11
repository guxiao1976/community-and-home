#!/usr/bin/env bash
#
# generate-grpc-mocks.sh — 批量生成 gRPC client mock
#
# Usage:
#   bash tools/generate-grpc-mocks.sh [service-name]
#   bash tools/generate-grpc-mocks.sh all
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PROTO_ROOT="$PROJECT_ROOT/api-proto/gen/go"

generate_grpc_mock() {
    local service_name="$1"
    local proto_dir="$PROTO_ROOT/$service_name/v1"

    if [[ ! -d "$proto_dir" ]]; then
        echo "⚠️  跳过: $service_name (目录不存在)"
        return
    fi

    local grpc_file="$proto_dir/${service_name}_grpc.pb.go"

    if [[ ! -f "$grpc_file" ]]; then
        echo "⚠️  跳过: $service_name (gRPC 文件不存在)"
        return
    fi

    echo "📦 生成 $service_name gRPC mock..."

    mkdir -p "$proto_dir/mocks"

    mockgen -source="$grpc_file" \
        -destination="$proto_dir/mocks/${service_name}_grpc_mock.go" \
        -package=mocks 2>&1 | grep -v "^$" || true

    if [[ -f "$proto_dir/mocks/${service_name}_grpc_mock.go" ]]; then
        local lines=$(wc -l < "$proto_dir/mocks/${service_name}_grpc_mock.go")
        echo "  ✅ 生成完成: ${service_name}_grpc_mock.go ($lines 行)"
    else
        echo "  ❌ 生成失败"
    fi
}

SERVICE="${1:-}"

if [[ -z "$SERVICE" ]]; then
    echo "Usage: $0 <service-name|all>"
    echo ""
    echo "可用服务:"
    ls -1 "$PROTO_ROOT" 2>/dev/null | grep -v "^common$" | sed 's/^/  - /'
    exit 1
fi

if [[ "$SERVICE" == "all" ]]; then
    echo "🚀 批量生成所有 gRPC mock..."
    echo ""

    for service_dir in "$PROTO_ROOT"/*; do
        [[ ! -d "$service_dir" ]] && continue
        service_name=$(basename "$service_dir")
        [[ "$service_name" == "common" ]] && continue

        generate_grpc_mock "$service_name"
    done

    echo ""
    echo "✅ 所有 gRPC mock 生成完成"
else
    generate_grpc_mock "$SERVICE"
fi
