#!/usr/bin/env bash
#
# pre-commit hook — 提交前检查测试
#
# 安装:
#   cp tools/pre-commit.sh .git/hooks/pre-commit
#   chmod +x .git/hooks/pre-commit
#

set -e

echo "🔍 Pre-commit 检查..."

# 获取所有待提交的 Go 文件
GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' | grep -v '_test\.go$' || true)

if [[ -z "$GO_FILES" ]]; then
    echo "✅ 没有 Go 文件变更"
    exit 0
fi

echo "检查到 Go 文件变更:"
echo "$GO_FILES" | sed 's/^/  - /'
echo ""

# ==================== 检查 1: Logic 文件必须有测试 ====================

echo "[1/4] 检查 Logic 文件是否有对应的测试..."

MISSING_TESTS=()

for file in $GO_FILES; do
    # 只检查 logic 目录下的文件
    if [[ "$file" == *"/logic/"* ]] && [[ "$file" == *"_logic.go" ]]; then
        test_file="${file%_logic.go}_logic_test.go"

        if [[ ! -f "$test_file" ]] && [[ ! $(git diff --cached --name-only | grep "$test_file") ]]; then
            MISSING_TESTS+=("$file")
        fi
    fi
done

if [[ ${#MISSING_TESTS[@]} -gt 0 ]]; then
    echo "❌ 以下 Logic 文件缺少测试:"
    for file in "${MISSING_TESTS[@]}"; do
        echo "  - $file"
        echo "    需要: ${file%_logic.go}_logic_test.go"
    done
    echo ""
    echo "💡 提示:"
    echo "  1. 创建测试文件: vim ${MISSING_TESTS[0]%_logic.go}_logic_test.go"
    echo "  2. 参考测试模板: services/user-service/api/internal/logic/user/user_logic_test.go"
    echo "  3. 或者使用 --no-verify 跳过检查（不推荐）: git commit --no-verify"
    echo ""
    exit 1
fi

echo "  ✅ 所有 Logic 文件都有对应的测试"

# ==================== 检查 2: 运行测试 ====================

echo ""
echo "[2/4] 运行变更文件的测试..."

# 提取所有变更的包路径
PACKAGES=$(echo "$GO_FILES" | xargs -I {} dirname {} | sort -u | sed 's|^|./|' | tr '\n' ' ')

if [[ -n "$PACKAGES" ]]; then
    echo "  测试包: $PACKAGES"

    if go test $PACKAGES -short 2>&1; then
        echo "  ✅ 测试通过"
    else
        echo ""
        echo "❌ 测试失败"
        echo ""
        echo "💡 提示:"
        echo "  1. 修复测试: go test $PACKAGES -v"
        echo "  2. 或者使用 --no-verify 跳过检查（不推荐）"
        echo ""
        exit 1
    fi
else
    echo "  ℹ️  没有需要测试的包"
fi

# ==================== 检查 3: 代码格式 ====================

echo ""
echo "[3/4] 检查代码格式..."

UNFORMATTED=$(echo "$GO_FILES" | xargs gofmt -l || true)

if [[ -n "$UNFORMATTED" ]]; then
    echo "❌ 以下文件格式不正确:"
    echo "$UNFORMATTED" | sed 's/^/  - /'
    echo ""
    echo "💡 提示:"
    echo "  1. 自动格式化: gofmt -w $UNFORMATTED"
    echo "  2. 或者: go fmt ./..."
    echo ""
    exit 1
fi

echo "  ✅ 代码格式正确"

# ==================== 检查 4: 静态分析 ====================

echo ""
echo "[4/4] 运行静态分析..."

if command -v golangci-lint &> /dev/null; then
    if golangci-lint run --new-from-rev=HEAD~ --timeout=2m 2>&1 | grep -E "^(Error|Warning):" || false; then
        echo ""
        echo "⚠️  发现静态分析问题"
        echo ""
        echo "💡 提示:"
        echo "  1. 查看详细信息: golangci-lint run"
        echo "  2. 或者使用 --no-verify 跳过检查（不推荐）"
        echo ""
        # 静态分析问题只警告，不阻止提交
        # exit 1
    else
        echo "  ✅ 静态分析通过"
    fi
else
    echo "  ⚠️  golangci-lint 未安装，跳过静态分析"
    echo "  安装: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# ==================== 完成 ====================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Pre-commit 检查全部通过"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

exit 0
