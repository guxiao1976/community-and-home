#!/bin/bash

# Worktree 同步脚本
# 用法: ./sync-worktree.sh identity

set -e

WORKTREE_NAME=$1
REPO_ROOT="/home/jiaoxh/my-code/community-and-home"
WORKTREE_PATH="$REPO_ROOT/.claude/worktrees/$WORKTREE_NAME"

if [ -z "$WORKTREE_NAME" ]; then
    echo "用法: $0 <worktree-name>"
    echo "示例: $0 identity"
    echo ""
    echo "可用的 worktree:"
    git worktree list | grep -v "^$REPO_ROOT " | awk '{print "  -", $1}' | sed "s|$REPO_ROOT/.claude/worktrees/||"
    exit 1
fi

if [ ! -d "$WORKTREE_PATH" ]; then
    echo "❌ Worktree '$WORKTREE_NAME' 不存在"
    exit 1
fi

echo "🔄 开始同步 worktree: $WORKTREE_NAME"
echo ""

# 1. 更新主仓库 main 分支
echo "📥 步骤 1/4: 更新主仓库 main 分支..."
cd "$REPO_ROOT"
git checkout main -q
BEFORE_PULL=$(git rev-parse HEAD)
git pull origin main
AFTER_PULL=$(git rev-parse HEAD)

if [ "$BEFORE_PULL" = "$AFTER_PULL" ]; then
    echo "   ✅ main 分支已是最新"
else
    echo "   ✅ main 分支已更新: $BEFORE_PULL -> $AFTER_PULL"
fi
echo ""

# 2. 进入 worktree
echo "📂 步骤 2/4: 进入 worktree..."
cd "$WORKTREE_PATH"
BRANCH=$(git branch --show-current)
echo "   当前分支: $BRANCH"
echo ""

# 3. 检查是否有未提交的修改
echo "🔍 步骤 3/4: 检查工作区状态..."
if ! git diff-index --quiet HEAD --; then
    echo "   ⚠️  有未提交的修改，请先提交或暂存"
    git status --short
    echo ""
    read -p "是否暂存这些修改？(y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git stash push -m "自动暂存: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "   ✅ 已暂存修改"
        STASHED=true
    else
        echo "   ❌ 请先处理未提交的修改"
        exit 1
    fi
else
    echo "   ✅ 工作区干净"
    STASHED=false
fi
echo ""

# 4. 合并 main 分支
echo "🔀 步骤 4/4: 合并 main 分支..."
BEFORE_MERGE=$(git rev-parse HEAD)
if git merge main --no-edit; then
    AFTER_MERGE=$(git rev-parse HEAD)
    if [ "$BEFORE_MERGE" = "$AFTER_MERGE" ]; then
        echo "   ✅ 已是最新，无需合并"
    else
        echo "   ✅ 合并成功: $BEFORE_MERGE -> $AFTER_MERGE"
    fi

    # 恢复暂存的修改
    if [ "$STASHED" = true ]; then
        echo ""
        echo "🔄 恢复暂存的修改..."
        if git stash pop; then
            echo "   ✅ 已恢复暂存的修改"
        else
            echo "   ⚠️  恢复时有冲突，请手动解决"
        fi
    fi
else
    echo "   ❌ 合并失败，有冲突需要解决"
    echo ""
    echo "冲突文件:"
    git diff --name-only --diff-filter=U
    echo ""
    echo "请手动解决冲突后执行:"
    echo "  git add ."
    echo "  git commit"
    exit 1
fi

echo ""
echo "✅ 同步完成！可以开始工作了"
echo ""
echo "当前状态:"
git log --oneline -3
