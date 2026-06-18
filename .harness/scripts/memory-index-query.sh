#!/usr/bin/env bash
#
# memory-index-query.sh — 查询记忆倒排索引
#
# 功能：
#   - 根据关键词查询匹配的记忆
#   - 支持多关键词查询（交集或并集）
#   - 按 severity 排序结果
#
# 用法：
#   bash .harness/scripts/memory-index-query.sh gRPC Proto 测试
#   bash .harness/scripts/memory-index-query.sh --union gRPC Proto  # 并集模式
#   bash .harness/scripts/memory-index-query.sh --service moderation-service gRPC
#
# 输出：
#   匹配的记忆 slug 列表（按 severity 排序）
#

set -euo pipefail

# ─── 配置 ───
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
INDEX_FILE="$PROJECT_ROOT/.harness/knowledge/memory/.memory-index.json"

MODE="intersect"  # intersect 或 union
SERVICE_FILTER=""
KEYWORDS=()

# ─── 解析参数 ───
while [[ $# -gt 0 ]]; do
    case $1 in
        --union)
            MODE="union"
            shift
            ;;
        --service)
            SERVICE_FILTER="$2"
            shift 2
            ;;
        --help|-h)
            echo "用法: $0 [--union] [--service <name>] <keyword1> [keyword2...]"
            echo ""
            echo "选项:"
            echo "  --union          并集模式（默认：交集）"
            echo "  --service <name> 只查询特定服务的记忆"
            echo "  --help           显示帮助"
            echo ""
            echo "示例:"
            echo "  $0 gRPC timeout"
            echo "  $0 --union gRPC Proto"
            echo "  $0 --service moderation-service 审核"
            exit 0
            ;;
        *)
            KEYWORDS+=("$1")
            shift
            ;;
    esac
done

# ─── 检查依赖 ───
if ! command -v jq &> /dev/null; then
    echo "❌ 错误: 需要安装 jq"
    exit 1
fi

if [ ! -f "$INDEX_FILE" ]; then
    echo "❌ 错误: 索引文件不存在: $INDEX_FILE"
    echo "   运行: bash .harness/scripts/memory-index-build.sh"
    exit 1
fi

if [ ${#KEYWORDS[@]} -eq 0 ]; then
    echo "❌ 错误: 至少提供一个关键词"
    echo "   运行: $0 --help"
    exit 1
fi

# ─── 查询索引 ───
echo "🔍 查询记忆索引..."
echo "   关键词: ${KEYWORDS[*]}"
echo "   模式: $MODE"
if [ -n "$SERVICE_FILTER" ]; then
    echo "   服务过滤: $SERVICE_FILTER"
fi
echo ""

# 收集所有匹配的 slugs
ALL_SLUGS=()

for keyword in "${KEYWORDS[@]}"; do
    # 从索引查询 trigger
    MATCHED=$(jq -r --arg kw "$keyword" '.index[$kw] // []' "$INDEX_FILE")

    if [ "$MATCHED" != "[]" ]; then
        # 转换为数组
        while IFS= read -r slug; do
            ALL_SLUGS+=("$slug")
        done < <(echo "$MATCHED" | jq -r '.[]')
    fi
done

# 根据模式处理
if [ "$MODE" == "intersect" ]; then
    # 交集：只保留出现次数 = 关键词数量的 slugs
    UNIQUE_SLUGS=($(printf '%s\n' "${ALL_SLUGS[@]}" | sort | uniq -c | awk -v count="${#KEYWORDS[@]}" '$1 == count {print $2}'))
else
    # 并集：去重
    UNIQUE_SLUGS=($(printf '%s\n' "${ALL_SLUGS[@]}" | sort | uniq))
fi

if [ ${#UNIQUE_SLUGS[@]} -eq 0 ]; then
    echo "❌ 未找到匹配的记忆"
    exit 0
fi

# ─── 按 severity 排序 ───
echo "✅ 找到 ${#UNIQUE_SLUGS[@]} 条匹配记忆"
echo ""

# 定义 severity 权重
declare -A SEVERITY_WEIGHT
SEVERITY_WEIGHT["must-follow"]=3
SEVERITY_WEIGHT["should-follow"]=2
SEVERITY_WEIGHT["info"]=1

# 输出结果（按 severity 排序）
for severity in "must-follow" "should-follow" "info"; do
    for slug in "${UNIQUE_SLUGS[@]}"; do
        # 读取该记忆的元数据
        MEM_DATA=$(jq -r --arg slug "$slug" '.memories[$slug]' "$INDEX_FILE")

        if [ "$MEM_DATA" == "null" ]; then
            continue
        fi

        MEM_SEVERITY=$(echo "$MEM_DATA" | jq -r '.severity')
        MEM_SERVICE=$(echo "$MEM_DATA" | jq -r '.service')
        MEM_TITLE=$(echo "$MEM_DATA" | jq -r '.title')
        MEM_TYPE=$(echo "$MEM_DATA" | jq -r '.type')

        # 服务过滤
        if [ -n "$SERVICE_FILTER" ] && [ "$MEM_SERVICE" != "$SERVICE_FILTER" ] && [ "$MEM_SERVICE" != "all" ]; then
            continue
        fi

        # 按 severity 分组输出
        if [ "$MEM_SEVERITY" == "$severity" ]; then
            echo "[$MEM_SEVERITY] $slug"
            echo "  标题: $MEM_TITLE"
            echo "  服务: $MEM_SERVICE | 类型: $MEM_TYPE"
            echo ""
        fi
    done
done

exit 0
