#!/usr/bin/env bash
#
# install-hooks.sh — 安装 Git hooks
#
# Usage:
#   bash tools/install-hooks.sh
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"

if [[ ! -d "$HOOKS_DIR" ]]; then
    echo "❌ 不是 Git 仓库: $PROJECT_ROOT"
    exit 1
fi

echo "🔧 安装 Git hooks..."
echo ""

# ==================== Pre-commit hook ====================

echo "[1/1] 安装 pre-commit hook..."

cp "$PROJECT_ROOT/tools/pre-commit.sh" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"

echo "  ✅ pre-commit hook 已安装"

# ==================== 完成 ====================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Git hooks 安装完成"
echo ""
echo "已安装的 hooks:"
echo "  ✅ pre-commit  - 提交前检查测试和代码格式"
echo ""
echo "测试 hooks:"
echo "  git commit -m 'test commit'"
echo ""
echo "跳过 hooks (不推荐):"
echo "  git commit --no-verify -m 'skip hooks'"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
