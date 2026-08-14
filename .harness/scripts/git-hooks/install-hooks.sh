#!/usr/bin/env bash
# 安装 git hooks 到 .git/hooks/（pre-commit + post-commit）
# 用法: bash .harness/scripts/git-hooks/install-hooks.sh
set -e

HOOKS_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HOOKS_DIR/../../.." && pwd)"

for hook in pre-commit post-commit; do
  if [[ -f "$HOOKS_DIR/$hook" ]]; then
    cp "$HOOKS_DIR/$hook" "$ROOT/.git/hooks/$hook"
    chmod +x "$ROOT/.git/hooks/$hook"
    echo "✓ 安装 $hook → .git/hooks/$hook"
  fi
done
echo "完成。pre-commit 做提交前门禁，post-commit 后台增量同步知识图谱。"
