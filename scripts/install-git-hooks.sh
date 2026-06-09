#!/usr/bin/env bash
# install-git-hooks.sh — Install post-commit hooks for graph auto-sync
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

HOOK_SCRIPT='#!/usr/bin/env bash
# Auto-sync knowledge graph on commit
nohup bash "'$PROJECT_ROOT'/.harness/scripts/graph-sync.sh" > /tmp/graph-sync.log 2>&1 &
'

# Install in each nested git repo
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

for repo in "${REPOS[@]}"; do
    if [[ -d "$repo/.git" ]]; then
        HOOK_FILE="$repo/.git/hooks/post-commit"
        echo "$HOOK_SCRIPT" > "$HOOK_FILE"
        chmod +x "$HOOK_FILE"
        echo "Installed hook: $HOOK_FILE"
    else
        echo "Skipped (no .git): $repo"
    fi
done

echo ""
echo "Git hooks installed. The knowledge graph will auto-sync on every commit."
echo "To test: make a commit and check http://localhost:7474"
