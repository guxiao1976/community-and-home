#!/usr/bin/env bash
# install-git-hooks.sh — Install Git hooks (pre-commit + post-commit)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# ==================== 配置 ====================

# pre-commit hook source
PRE_COMMIT_SOURCE="$PROJECT_ROOT/.harness/scripts/git-hooks/pre-commit"

# post-commit hook script for graph auto-sync
POST_COMMIT_SCRIPT='#!/usr/bin/env bash
# Auto-sync knowledge graph on commit
nohup bash "'$PROJECT_ROOT'/.harness/scripts/graph-sync.sh" > /tmp/graph-sync.log 2>&1 &
'

# 所有需要安装 hook 的 Git 仓库
REPOS=(
    "$PROJECT_ROOT"
    "$PROJECT_ROOT/api-proto"
    "$PROJECT_ROOT/common"
    "$PROJECT_ROOT/services/user-service"
    "$PROJECT_ROOT/services/auth-service"
    "$PROJECT_ROOT/services/permission-service"
    "$PROJECT_ROOT/services/master-data-service"
    "$PROJECT_ROOT/services/moderation-service"
    "$PROJECT_ROOT/services/ai-model-service"
    "$PROJECT_ROOT/services/file-service"
    "$PROJECT_ROOT/services/community-hub-service"
)

# ==================== 函数 ====================

install_pre_commit() {
    local repo=$1
    local hooks_dir="$repo/.git/hooks"
    local target="$hooks_dir/pre-commit"

    # 检查源文件是否存在
    if [[ ! -f "$PRE_COMMIT_SOURCE" ]]; then
        echo "  ⚠️  pre-commit source not found: $PRE_COMMIT_SOURCE"
        return 1
    fi

    # 创建 hooks 目录（如果不存在）
    mkdir -p "$hooks_dir"

    # 检查是否已经是正确的符号链接
    if [[ -L "$target" && "$(readlink "$target")" == "$PRE_COMMIT_SOURCE" ]]; then
        echo "  ✅ pre-commit already installed (symlink)"
        return 0
    fi

    # 备份现有 hook（如果存在且不是符号链接）
    if [[ -f "$target" && ! -L "$target" ]]; then
        backup="$target.backup.$(date +%Y%m%d_%H%M%S)"
        mv "$target" "$backup"
        echo "  📦 Backed up existing hook to: $backup"
    fi

    # 创建符号链接
    ln -sf "$PRE_COMMIT_SOURCE" "$target"
    chmod +x "$target"
    echo "  ✅ pre-commit installed (symlink)"
}

install_post_commit() {
    local repo=$1
    local hooks_dir="$repo/.git/hooks"
    local target="$hooks_dir/post-commit"

    mkdir -p "$hooks_dir"

    # 检查是否已安装相同内容
    if [[ -f "$target" ]] && grep -q "graph-sync.sh" "$target" 2>/dev/null; then
        echo "  ✅ post-commit already installed"
        return 0
    fi

    # 写入 post-commit hook
    echo "$POST_COMMIT_SCRIPT" > "$target"
    chmod +x "$target"
    echo "  ✅ post-commit installed"
}

# ==================== 主流程 ====================

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 Installing Git Hooks"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

installed_count=0
skipped_count=0

for repo in "${REPOS[@]}"; do
    repo_name=$(basename "$repo")

    if [[ ! -d "$repo/.git" ]]; then
        echo "❌ $repo_name — no .git directory (skipped)"
        skipped_count=$((skipped_count + 1))
        continue
    fi

    echo "📁 $repo_name"

    # 安装 pre-commit hook
    install_pre_commit "$repo"

    # 仅主仓库安装 post-commit hook（graph-sync）
    if [[ "$repo" == "$PROJECT_ROOT" ]]; then
        install_post_commit "$repo"
    fi

    installed_count=$((installed_count + 1))
    echo ""
done

# ==================== 总结 ====================

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Git Hooks Installation Complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Summary:"
echo "  • Installed: $installed_count repositories"
echo "  • Skipped:   $skipped_count repositories"
echo ""
echo "🔍 Pre-commit Hook (质量门禁):"
echo "  ✓ Logic 文件必须有测试"
echo "  ✓ 运行变更包的测试"
echo "  ✓ 代码格式检查"
echo "  ✓ 静态分析"
echo ""
echo "📡 Post-commit Hook (主仓库):"
echo "  ✓ Knowledge graph auto-sync"
echo ""
echo "💡 Usage:"
echo "  • Hooks run automatically on 'git commit'"
echo "  • To bypass: git commit --no-verify (not recommended)"
echo "  • To test: make a small change and commit"
echo ""
