#!/usr/bin/env bash
#
# verify-hooks.sh — 验证 Git Hooks 是否正确安装和工作
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 Verifying Git Hooks Installation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 检查的仓库列表
REPOS=(
    "$PROJECT_ROOT"
    "$PROJECT_ROOT/api-proto"
    "$PROJECT_ROOT/common"
    "$PROJECT_ROOT/services/master-data-service"
    "$PROJECT_ROOT/services/moderation-service"
    "$PROJECT_ROOT/services/ai-model-service"
)

all_ok=true
installed=0
missing=0

for repo in "${REPOS[@]}"; do
    repo_name=$(basename "$repo")
    hook_path="$repo/.git/hooks/pre-commit"

    if [[ ! -d "$repo/.git" ]]; then
        echo "⚪ $repo_name — no .git directory (skipped)"
        continue
    fi

    if [[ -L "$hook_path" ]]; then
        target=$(readlink "$hook_path")
        if [[ "$target" == *"pre-commit"* ]]; then
            echo "✅ $repo_name — pre-commit hook installed"
            installed=$((installed + 1))
        else
            echo "⚠️  $repo_name — symlink exists but points to wrong target: $target"
            all_ok=false
        fi
    elif [[ -f "$hook_path" ]]; then
        echo "⚠️  $repo_name — pre-commit exists but is not a symlink"
        all_ok=false
    else
        echo "❌ $repo_name — pre-commit hook NOT installed"
        missing=$((missing + 1))
        all_ok=false
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ $missing -gt 0 ]]; then
    echo "❌ Verification FAILED"
    echo ""
    echo "📊 Summary:"
    echo "  • Installed: $installed"
    echo "  • Missing:   $missing"
    echo ""
    echo "💡 Fix:"
    echo "  bash scripts/install-git-hooks.sh"
    echo ""
    exit 1
fi

if ! $all_ok; then
    echo "⚠️  Verification WARNING - some hooks have issues"
    echo ""
    echo "📊 Summary:"
    echo "  • Installed: $installed"
    echo ""
    echo "💡 Fix:"
    echo "  bash scripts/install-git-hooks.sh"
    echo ""
    exit 1
fi

echo "✅ All Git Hooks Verified Successfully"
echo ""
echo "📊 Summary:"
echo "  • Installed: $installed repositories"
echo ""
echo "🎯 What these hooks do:"
echo "  ✓ Check Logic files have tests"
echo "  ✓ Run tests on changed packages"
echo "  ✓ Verify code formatting"
echo "  ✓ Run static analysis"
echo ""
echo "💡 Test it:"
echo "  1. Make a small change: echo '// test' >> services/user-service/README.md"
echo "  2. Try to commit: git add . && git commit -m 'test'"
echo "  3. The hook will run automatically"
echo ""
