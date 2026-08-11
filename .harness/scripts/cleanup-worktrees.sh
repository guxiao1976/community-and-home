#!/usr/bin/env bash
#
# cleanup-worktrees.sh — 清理 Harness pipeline 残留的 git worktree
#
# 背景：pipeline 每次运行用 `isolation: 'worktree'` 创建隔离工作区，
# 但从未清理，导致 .claude/worktrees/ 下积累了 30+ 个指向旧 commit 的
# 残留 worktree（实测 2026-08：36 个，全部指向 2026-07-10 前的提交）。
#
# 安全策略：
#   - 只清理「无未提交改动」的 worktree（有改动的 SKIP 并提示）
#   - 用 `git worktree remove --force`（先尝试规范删除，失败才 rm -rf）
#   - 完成后 `git worktree prune` 清理元数据
#
# 用法（在 WSL 项目根执行，注意不要在 Windows Git Bash 里跑——路径不兼容）：
#   bash .harness/scripts/cleanup-worktrees.sh
#
# 建议：把本脚本加入定期清理（如 cron 每周一次），并在 pipeline 结束后
# 调用 `git worktree prune`。

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "=== Worktree 清理 ==="
echo "当前 worktree 总数: $(git worktree list | wc -l)"

REMOVED=0
SKIPPED=0
while IFS= read -r wt; do
  [[ -z "$wt" ]] && continue
  # 跳过主工作区
  [[ "$wt" == "$PROJECT_ROOT" ]] && continue

  if git -C "$wt" status --porcelain 2>/dev/null | grep -q .; then
    echo "⚠️  SKIP $wt — 存在未提交改动，保留（请人工确认）"
    SKIPPED=$((SKIPPED + 1))
    continue
  fi

  echo "🧹  移除 $wt"
  if git worktree remove --force "$wt" 2>/dev/null; then
    REMOVED=$((REMOVED + 1))
  else
    echo "    （git worktree remove 失败，尝试直接删除目录）"
    rm -rf "$wt"
    REMOVED=$((REMOVED + 1))
  fi
done < <(git worktree list --porcelain | grep '^worktree ' | cut -d' ' -f2-)

git worktree prune

echo ""
echo "=== 清理完成 ==="
echo "已移除: $REMOVED  保留(有改动): $SKIPPED"
echo "剩余 worktree:"
git worktree list
