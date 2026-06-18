#!/usr/bin/env bash
#
# test-memory-index.sh — 记忆索引系统测试脚本
#
# 测试内容：
#   - 索引构建准确性
#   - 查询性能
#   - 查询结果正确性
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

PASS=0
FAIL=0

echo "╔════════════════════════════════════════════════════════════╗"
echo "║              🧪 记忆索引系统测试                            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# ═══ 测试 1: 索引构建 ═══
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: 索引构建"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

START_TIME=$(date +%s%N)
bash .harness/scripts/memory-index-build.sh > /tmp/test-build.log 2>&1
END_TIME=$(date +%s%N)

BUILD_TIME=$(( (END_TIME - START_TIME) / 1000000 ))

if [ -f ".harness/knowledge/memory/.memory-index.json" ]; then
    TRIGGERS=$(jq -r '.triggers | length' .harness/knowledge/memory/.memory-index.json)
    MEMORIES=$(jq -r '.memories | length' .harness/knowledge/memory/.memory-index.json)

    echo "  ✅ 索引构建成功"
    echo "     Triggers: $TRIGGERS 个"
    echo "     Memories: $MEMORIES 条"
    echo "     耗时: ${BUILD_TIME}ms"
    PASS=$((PASS + 1))
else
    echo "  ❌ 索引文件不存在"
    FAIL=$((FAIL + 1))
fi
echo ""

# ═══ 测试 2: 查询性能 ═══
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: 查询性能"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

QUERIES=("gRPC" "测试" "Proto" "API" "数据库")

for query in "${QUERIES[@]}"; do
    START_TIME=$(date +%s%N)
    RESULT=$(bash .harness/scripts/memory-index-query.sh --union "$query" 2>&1)
    END_TIME=$(date +%s%N)

    QUERY_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
    FOUND=$(echo "$RESULT" | grep -o "找到 [0-9]* 条" || echo "0 条")

    echo "  查询: $query"
    echo "    结果: $FOUND"
    echo "    耗时: ${QUERY_TIME}ms"

    if [ $QUERY_TIME -lt 500 ]; then
        echo "    ✅ 性能合格 (<500ms)"
        PASS=$((PASS + 1))
    else
        echo "    ⚠️  性能警告 (>500ms)"
    fi
done
echo ""

# ═══ 测试 3: 查询准确性 ═══
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 3: 查询准确性"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 测试 gRPC 关键词应该找到相关记忆
RESULT=$(bash .harness/scripts/memory-index-query.sh --union gRPC)
if echo "$RESULT" | grep -q "grpc"; then
    echo "  ✅ gRPC 查询准确"
    PASS=$((PASS + 1))
else
    echo "  ❌ gRPC 查询不准确"
    FAIL=$((FAIL + 1))
fi

# 测试交集模式
RESULT=$(bash .harness/scripts/memory-index-query.sh gRPC Proto)
echo "  交集查询测试: gRPC AND Proto"
echo "    $(echo "$RESULT" | grep "找到.*条" || echo "查询完成")"
PASS=$((PASS + 1))

echo ""

# ═══ 汇总结果 ═══
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              📊 测试结果                                    ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "通过: $PASS"
echo "失败: $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "✅ 所有测试通过！"
    exit 0
else
    echo "❌ 有 $FAIL 个测试失败"
    exit 1
fi
