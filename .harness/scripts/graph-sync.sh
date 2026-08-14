#!/usr/bin/env bash
# graph-sync.sh — Sync project knowledge graph to Neo4j
# Usage:
#   ./.harness/scripts/graph-sync.sh              # Incremental (git diff since last sync)
#   ./.harness/scripts/graph-sync.sh --full       # Full rebuild

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
POPULATOR_DIR="$SCRIPT_DIR/graph-populator"
SYNC_STAMP="$PROJECT_ROOT/.harness/.graph_last_sync"

NEO4J_URI="${NEO4J_URI:-bolt://localhost:7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-neo4j123456}"
NEO4J_DB="${NEO4J_DB:-neo4j}"

MODE="${1:-incremental}"

# Check Neo4j is reachable
# FIX: silent `exit 0` on unreachable Neo4j made the QA `graph_freshness` check
# fail forever without any actionable signal (the loop treated it as "success").
# Now fail loudly with a non-zero exit so callers (loop self-heal, humans, CI)
# know the sync did NOT happen and why.
if ! curl -sf -o /dev/null --max-time 5 "http://localhost:7474"; then
    echo "[graph-sync] ❌ ERROR: Neo4j not reachable at http://localhost:7474 (bolt:${NEO4J_URI})" >&2
    echo "   原因: 知识图谱未同步，QA 的 graph_freshness 检查会持续 FAIL。" >&2
    echo "   修复: 启动 Neo4j 后重试:" >&2
    echo "     docker compose up -d neo4j" >&2
    echo "     或设置 NEO4J_URI/NEO4J_PASSWORD 环境变量。" >&2
    echo "   注意: 本脚本退出码为非零（未执行同步），不要将未同步误判为成功。" >&2
    exit 1
fi

cd "$POPULATOR_DIR"

if [[ "$MODE" == "--full" ]]; then
    echo "[graph-sync] Full rebuild..."
    GOWORK=off NEO4J_URI="$NEO4J_URI" NEO4J_USER="$NEO4J_USER" \
        NEO4J_PASSWORD="$NEO4J_PASSWORD" NEO4J_DB="$NEO4J_DB" \
        PROJECT_ROOT="$PROJECT_ROOT" \
        go run main.go --full
else
    echo "[graph-sync] Incremental update..."
    # Find changed files since last sync (or last commit if no stamp)
    if [[ -f "$SYNC_STAMP" ]]; then
        SINCE=$(cat "$SYNC_STAMP")
    else
        SINCE="HEAD~1"
    fi

    # Collect changed files from all nested repos
    CHANGED=""
    for repo in "$PROJECT_ROOT" "$PROJECT_ROOT"/services/*/ "$PROJECT_ROOT"/api-proto/ "$PROJECT_ROOT"/common/ "$PROJECT_ROOT"/web/; do
        if [[ -d "$repo/.git" ]]; then
            if [[ "$SINCE" == "HEAD~1" ]]; then
                # No prior stamp: compare against last commit
                CHANGED+=$(git -C "$repo" diff --name-only "$SINCE" HEAD 2>/dev/null || true)
            else
                # Committed changes since last sync timestamp
                CHANGED+=$(git -C "$repo" log --since=@"$SINCE" --format="" --name-only HEAD 2>/dev/null || true)
                # Also capture uncommitted (working tree) changes
                CHANGED+=$'\n'
                CHANGED+=$(git -C "$repo" diff --name-only HEAD 2>/dev/null || true)
                CHANGED+=$'\n'
                CHANGED+=$(git -C "$repo" diff --cached --name-only 2>/dev/null || true)
            fi
            CHANGED+=$'\n'
        fi
    done
    CHANGED=$(echo "$CHANGED" | grep -E '\.(proto|go|ts|vue|yaml|yml)$' | grep -v '_test.go' | sort -u || true)

    if [[ -z "${CHANGED// }" ]]; then
        echo "[graph-sync] No relevant files changed — skipping"
        exit 0
    fi

    echo "[graph-sync] Changed files:"
    echo "$CHANGED" | head -10
    echo "[graph-sync] ... ($(echo "$CHANGED" | wc -l) files total)"

    GOWORK=off NEO4J_URI="$NEO4J_URI" NEO4J_USER="$NEO4J_USER" \
        NEO4J_PASSWORD="$NEO4J_PASSWORD" NEO4J_DB="$NEO4J_DB" \
        PROJECT_ROOT="$PROJECT_ROOT" \
        go run main.go --incremental --files <(echo "$CHANGED")
fi

# After graph population, generate context files for all services
if command -v python3 &>/dev/null; then
    for svc_dir in "$PROJECT_ROOT"/services/*/; do
        svc_name=$(basename "$svc_dir")
        context_dir="$svc_dir/docs"
        mkdir -p "$context_dir"
        # 幂等写：新内容与现有文件除时间戳行外一致 → 跳过写
        # （否则每次 post-commit sync 更新时间戳 → 需再提交 → 又触发 sync 的循环噪音）
        tmp="$context_dir/.graph-context.$$.tmp"
        bash "$SCRIPT_DIR/graph-query.sh" "$svc_name" > "$tmp" 2>/dev/null || true
        if [ -f "$context_dir/graph-context.md" ] \
           && diff <(grep -v '自动生成于' "$context_dir/graph-context.md" 2>/dev/null) <(grep -v '自动生成于' "$tmp") >/dev/null 2>&1; then
            rm -f "$tmp"
            echo "[graph-sync] graph-context.md 无实质变化，跳过写 $svc_name"
        else
            mv "$tmp" "$context_dir/graph-context.md"
            echo "[graph-sync] Generated graph-context.md for $svc_name"
        fi
    done
fi

# Update stamp
mkdir -p "$(dirname "$SYNC_STAMP")"
date +%s > "$SYNC_STAMP"
echo "[graph-sync] Done. Stamp: $(cat "$SYNC_STAMP")"
