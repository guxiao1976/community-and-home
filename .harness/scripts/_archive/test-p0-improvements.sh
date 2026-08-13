#!/usr/bin/env bash
#
# test-p0-improvements.sh — P0 改进效果验证脚本
#
# 测试目标：
#   1. 验证错误消息格式升级（WHY + FIX + EXAMPLE + REFERENCE）
#   2. 验证模式文档可访问性和完整性
#   3. 模拟 Agent 修复流程

set -eu

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "=========================================="
echo "  P0 改进效果验证"
echo "=========================================="
echo ""

# ─── 测试 1: 创建一个违规文件 ─────────────────────────────────────

echo "测试 1: 创建跨服务 DB 导入违规"
echo "──────────────────────────────────────"

# 备份原文件（如果存在）
TEST_FILE="services/auth-service/api/internal/logic/test_violation_logic.go"
BACKUP_FILE="${TEST_FILE}.backup"

if [[ -f "$TEST_FILE" ]]; then
  mv "$TEST_FILE" "$BACKUP_FILE"
fi

# 创建一个违规文件
mkdir -p "$(dirname "$TEST_FILE")"
cat > "$TEST_FILE" << 'EOF'
package logic

import (
    "context"
    "github.com/guxiao1976/community-user/model"  // ❌ 违规：跨服务导入
)

type TestViolationLogic struct {
    ctx    context.Context
}

func NewTestViolationLogic(ctx context.Context) *TestViolationLogic {
    return &TestViolationLogic{ctx: ctx}
}

func (l *TestViolationLogic) Test() error {
    // 直接访问其他服务的 model
    return nil
}
EOF

echo "✅ 已创建违规文件: $TEST_FILE"
echo ""

# ─── 测试 2: 运行检查并捕获错误消息 ────────────────────────────────

echo "测试 2: 运行 harness-checks.sh 检测违规"
echo "──────────────────────────────────────"

# 运行检查（允许失败）
set +e
CHECK_OUTPUT=$(bash .harness/skills/qa/scripts/harness-checks.sh --service auth-service 2>&1)
CHECK_EXIT_CODE=$?
set -e

echo "检查退出码: $CHECK_EXIT_CODE"

if [[ $CHECK_EXIT_CODE -ne 0 ]]; then
    echo "✅ 预期结果：检查失败（发现违规）"
else
    echo "❌ 意外结果：检查通过（应该失败）"
fi
echo ""

# 提取 cross_service_import 检查的结果
if echo "$CHECK_OUTPUT" | grep -q "cross.service.*import"; then
    echo "✅ 检测到跨服务导入违规"
    echo ""

    # 显示错误消息（人类可读格式）
    echo "人类可读格式:"
    echo "─────────────────────────────"
    echo "$CHECK_OUTPUT" | grep -A 1 "cross.service.*import"
    echo ""
else
    echo "❌ 未检测到跨服务导入违规"
fi

# ─── 测试 3: 验证错误消息包含必需字段 ─────────────────────────────

echo "测试 3: 验证结构化错误消息字段"
echo "──────────────────────────────────────"

# 使用 --json 输出（如果支持）或从脚本内部 RESULTS 数组提取
# 这里我们直接检查输出中是否包含关键信息

REQUIRED_FIELDS=("why" "fix" "example" "reference")
FOUND_COUNT=0

for field in "${REQUIRED_FIELDS[@]}"; do
    # 由于当前是 markdown 输出，我们检查是否包含对应内容
    case "$field" in
        "why")
            if echo "$CHECK_OUTPUT" | grep -q "服务间通信必须通过 gRPC"; then
                echo "✅ WHY 字段: 找到原因说明"
                FOUND_COUNT=$((FOUND_COUNT + 1))
            else
                echo "❌ WHY 字段: 缺失"
            fi
            ;;
        "fix")
            if echo "$CHECK_OUTPUT" | grep -q "移除.*model.*导入"; then
                echo "✅ FIX 字段: 找到修复步骤"
                FOUND_COUNT=$((FOUND_COUNT + 1))
            else
                echo "❌ FIX 字段: 缺失"
            fi
            ;;
        "example")
            if echo "$CHECK_OUTPUT" | grep -q "verify_token_logic.go"; then
                echo "✅ EXAMPLE 字段: 找到参考实现"
                FOUND_COUNT=$((FOUND_COUNT + 1))
            else
                echo "❌ EXAMPLE 字段: 缺失"
            fi
            ;;
        "reference")
            if echo "$CHECK_OUTPUT" | grep -q "cross-service-rpc.md"; then
                echo "✅ REFERENCE 字段: 找到模式文档链接"
                FOUND_COUNT=$((FOUND_COUNT + 1))
            else
                echo "❌ REFERENCE 字段: 缺失"
            fi
            ;;
    esac
done

echo ""
echo "字段完整性: $FOUND_COUNT / 4"
echo ""

# ─── 测试 4: 验证模式文档存在且可读 ──────────────────────────────

echo "测试 4: 验证模式文档完整性"
echo "──────────────────────────────────────"

PATTERN_DOC=".harness/linters/patterns/cross-service-rpc.md"

if [[ -f "$PATTERN_DOC" ]]; then
    echo "✅ 模式文档存在: $PATTERN_DOC"

    # 检查文档关键章节
    REQUIRED_SECTIONS=("核心原则" "为什么" "错误模式" "正确模式" "完整示例" "检查清单")
    SECTION_FOUND=0

    for section in "${REQUIRED_SECTIONS[@]}"; do
        if grep -q "$section" "$PATTERN_DOC"; then
            echo "  ✅ 包含章节: $section"
            SECTION_FOUND=$((SECTION_FOUND + 1))
        else
            echo "  ❌ 缺失章节: $section"
        fi
    done

    echo ""
    echo "章节完整性: $SECTION_FOUND / ${#REQUIRED_SECTIONS[@]}"

    # 检查示例代码引用
    if grep -q "verify_token_logic.go:28-35" "$PATTERN_DOC"; then
        echo "  ✅ 包含参考实现路径"
    else
        echo "  ❌ 缺失参考实现路径"
    fi

else
    echo "❌ 模式文档不存在: $PATTERN_DOC"
fi

echo ""

# ─── 测试 5: 模拟 Agent 修复流程 ──────────────────────────────────

echo "测试 5: 模拟 Agent 修复流程"
echo "──────────────────────────────────────"

echo "步骤 1: Agent 读取错误消息中的 reference 字段"
echo "  → .harness/linters/patterns/cross-service-rpc.md"
echo ""

echo "步骤 2: Agent 读取模式文档（前 20 行预览）"
echo "─────────────────────────────"
head -20 "$PATTERN_DOC" | sed 's/^/  /'
echo "  ..."
echo ""

echo "步骤 3: Agent 看到修复清单"
echo "─────────────────────────────"
grep -A 10 "## 检查清单" "$PATTERN_DOC" | sed 's/^/  /'
echo ""

echo "步骤 4: Agent 按文档修复代码"
echo "  → 移除跨服务 model 导入"
echo "  → 添加 RPC 客户端配置"
echo "  → 调用 RPC 方法"
echo ""

echo "步骤 5: Agent 重新运行检查 → 通过 ✓"
echo ""

# ─── 测试 6: 验证其他模式文档 ──────────────────────────────────────

echo "测试 6: 验证所有模式文档"
echo "──────────────────────────────────────"

PATTERN_DIR=".harness/linters/patterns"
PATTERN_DOCS=(
    "proto-jstype.md"
    "json-string.md"
    "cross-service-rpc.md"
    "response-wrap.md"
    "README.md"
)

DOC_FOUND=0
for doc in "${PATTERN_DOCS[@]}"; do
    if [[ -f "$PATTERN_DIR/$doc" ]]; then
        LINES=$(wc -l < "$PATTERN_DIR/$doc")
        echo "✅ $doc — $LINES 行"
        DOC_FOUND=$((DOC_FOUND + 1))
    else
        echo "❌ $doc — 缺失"
    fi
done

echo ""
echo "文档完整性: $DOC_FOUND / ${#PATTERN_DOCS[@]}"
echo ""

# ─── 清理测试文件 ───────────────────────────────────────────────────

echo "清理: 移除测试违规文件"
echo "──────────────────────────────────────"

if [[ -f "$TEST_FILE" ]]; then
    rm "$TEST_FILE"
    echo "✅ 已删除: $TEST_FILE"
fi

if [[ -f "$BACKUP_FILE" ]]; then
    mv "$BACKUP_FILE" "$TEST_FILE"
    echo "✅ 已恢复原文件"
fi

echo ""

# ─── 最终评分 ──────────────────────────────────────────────────────

echo "=========================================="
echo "  测试结果总结"
echo "=========================================="
echo ""

TOTAL_SCORE=0
MAX_SCORE=0

# 测试 1: 违规文件创建
MAX_SCORE=$((MAX_SCORE + 1))
TOTAL_SCORE=$((TOTAL_SCORE + 1))

# 测试 2: 检查失败检测
MAX_SCORE=$((MAX_SCORE + 1))
if [[ $CHECK_EXIT_CODE -ne 0 ]]; then
    TOTAL_SCORE=$((TOTAL_SCORE + 1))
fi

# 测试 3: 错误消息字段完整性
MAX_SCORE=$((MAX_SCORE + 4))
TOTAL_SCORE=$((TOTAL_SCORE + FOUND_COUNT))

# 测试 4: 模式文档章节完整性
MAX_SCORE=$((MAX_SCORE + 6))
TOTAL_SCORE=$((TOTAL_SCORE + SECTION_FOUND))

# 测试 6: 所有模式文档存在
MAX_SCORE=$((MAX_SCORE + 5))
TOTAL_SCORE=$((TOTAL_SCORE + DOC_FOUND))

echo "总分: $TOTAL_SCORE / $MAX_SCORE"
echo ""

if [[ $TOTAL_SCORE -eq $MAX_SCORE ]]; then
    echo "✅ 所有测试通过！P0 改进验证成功。"
    exit 0
elif [[ $TOTAL_SCORE -ge $((MAX_SCORE * 80 / 100)) ]]; then
    echo "⚠️  大部分测试通过（$((TOTAL_SCORE * 100 / MAX_SCORE))%），但仍有改进空间。"
    exit 0
else
    echo "❌ 测试失败率较高（$((TOTAL_SCORE * 100 / MAX_SCORE))%），需要检查改进实施。"
    exit 1
fi
