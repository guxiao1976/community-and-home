#!/usr/bin/env bash
# graph-query.sh — Query Neo4j for a service's context subgraph
# Usage: bash .harness/scripts/graph-query.sh <service-name>
#   e.g.: bash .harness/scripts/graph-query.sh user-service
# Output: Markdown to stdout
#
# Part of the AI Coding Harness knowledge graph integration.
# This script is called by graph-sync.sh to auto-generate
# docs/graph-context.md for each service.

set -euo pipefail

SERVICE_NAME="${1:-}"
if [[ -z "$SERVICE_NAME" ]]; then
    echo "Usage: $0 <service-name>"
    echo "Example: $0 user-service"
    exit 1
fi

NEO4J_URI="${NEO4J_URI:-http://localhost:7474}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-neo4j123456}"
NEO4J_DB="${NEO4J_DB:-neo4j}"

# Check Neo4j is reachable
if ! curl -sf -o /dev/null "$NEO4J_URI"; then
    echo "# 知识图谱上下文 — $SERVICE_NAME"
    echo ""
    echo "> ⚠️ Neo4j 不可用，无法查询知识图谱。请先启动 Neo4j: docker compose up -d neo4j"
    # ── 打点：图谱查询失败（Neo4j 不可达）──
    bash "$(dirname "$0")/log-usage.sh" graph-query service="$SERVICE_NAME" result=unavailable 2>/dev/null || true
    exit 0
fi

# Build a JSON payload for a Cypher statement with parameters
cypher_payload() {
    local cypher="$1"
    # Escape double quotes and backslashes for JSON
    local escaped
    escaped=$(printf '%s' "$cypher" | python3 -c "
import sys, json
print(json.dumps(sys.stdin.read()))
")
    printf '{"statements":[{"statement":%s,"parameters":{"svc":"%s"}}]}' \
        "$escaped" "$SERVICE_NAME"
}

neo4j_query() {
    local cypher="$1"
    local payload
    payload=$(cypher_payload "$cypher")
    curl -s -u "$NEO4J_USER:$NEO4J_PASSWORD" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$NEO4J_URI/db/$NEO4J_DB/tx/commit"
}

TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

cat <<EOF
# 知识图谱上下文 — $SERVICE_NAME

> 自动生成于 $TIMESTAMP | 数据源: Neo4j 知识图谱 | 每次 \`graph-sync.sh\` 后刷新

EOF

###############################################################################
# Query 1: Service identity
###############################################################################
RESULT=$(neo4j_query "MATCH (s:Service {id: \$svc}) RETURN s.name AS name, s.language AS language, s.port AS port, s.apiPort AS apiPort")
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    rows = data['results'][0]['data']
    if not rows:
        print('> ⚠️ 服务 \"$SERVICE_NAME\" 未在知识图谱中找到。运行 graph-sync.sh 更新图谱。')
        sys.exit(0)
    r = rows[0]['row']
    port_api = str(r[3]) if r[3] is not None else '-'
    print('## 服务标识')
    print('')
    print('| 属性 | 值 |')
    print('|------|-----|')
    print(f'| 名称 | {r[0]} |')
    print(f'| 语言 | {r[1]} |')
    print(f'| 端口 (gRPC) | {r[2]} |')
    print(f'| 端口 (API)  | {port_api} |')
    print('')
except Exception as e:
    print('> ⚠️ 查询失败:', e)
    print('')
"

# Skip remaining queries if service not found
if echo "$RESULT" | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data['results'][0]['data']
sys.exit(0 if rows else 1)
" 2>/dev/null; then
    :
else
    echo "---"
    echo "*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*"
    exit 0
fi

###############################################################################
# Query 2: Dependencies
###############################################################################
RESULT=$(neo4j_query "MATCH (s:Service {id: \$svc})-[d:DEPENDS_ON]->(dep:Service) RETURN dep.name AS name, d.type AS type ORDER BY dep.name")
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    rows = data['results'][0]['data']
    print('## 服务依赖')
    print('')
    if not rows:
        print('无外部依赖')
    else:
        print('| 依赖服务 | 依赖类型 |')
        print('|---------|---------|')
        for r in rows:
            dep_type = r['row'][1] if r['row'][1] else 'gRPC'
            print(f'| {r[\"row\"][0]} | {dep_type} |')
    print('')
except Exception as e:
    print('> ⚠️ 查询失败:', e)
    print('')
"

###############################################################################
# Query 2b: Reverse dependencies (consumed by)
###############################################################################
RESULT=$(neo4j_query "MATCH (consumer:Service)-[d:DEPENDS_ON]->(s:Service {id: \$svc}) RETURN consumer.name AS name, d.type AS type ORDER BY consumer.name")
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    rows = data['results'][0]['data']
    print('## 被依赖方')
    print('')
    if not rows:
        print('无服务依赖本服务')
    else:
        print('| 消费方 | 依赖类型 |')
        print('|---------|---------|')
        for r in rows:
            print(f'| {r[\"row\"][0]} | {r[\"row\"][1]} |')
    print('')
except Exception as e:
    print('> ⚠️ 查询失败:', e)
    print('')
"

###############################################################################
# Query 3: REST API routes
###############################################################################
RESULT=$(neo4j_query "MATCH (s:Service {id: \$svc})-[:EXPOSES]->(r:ApiRoute) RETURN r.method AS method, r.path AS path ORDER BY r.path")
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    rows = data['results'][0]['data']
    print('## REST API 路由')
    print('')
    if not rows:
        print('无 REST 路由')
    else:
        print('| 方法 | 路径 |')
        print('|------|------|')
        for r in rows:
            print(f'| {r[\"row\"][0]} | {r[\"row\"][1]} |')
    print('')
except Exception as e:
    print('> ⚠️ 查询失败:', e)
    print('')
"

###############################################################################
# Query 4: gRPC RPCs
###############################################################################
RESULT=$(neo4j_query "
MATCH (s:Service {id: \$svc})-[:IMPLEMENTS]->(rpc:ProtoRpc)
RETURN rpc.service AS service, rpc.name AS name,
       rpc.inputType AS inputMsg, rpc.outputType AS outputMsg
ORDER BY rpc.name
")
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    rows = data['results'][0]['data']
    print('## gRPC 接口')
    print('')
    if not rows:
        print('无 gRPC 接口')
    else:
        print('| RPC 方法 | 输入消息 | 输出消息 |')
        print('|---------|---------|---------|')
        for r in rows:
            rpc_name = r['row'][1]
            input_str = r['row'][2] or '-'
            output_str = r['row'][3] or '-'
            print(f'| {rpc_name} | {input_str} | {output_str} |')
    print('')
except Exception as e:
    print('> ⚠️ 查询失败:', e)
    print('')
"

###############################################################################
# Query 5: Database tables
###############################################################################
RESULT=$(neo4j_query "
MATCH (t:DbTable {service: \$svc})
OPTIONAL MATCH (t)-[:HAS_COLUMN]->(c:DbColumn)
RETURN t.name AS table, collect(DISTINCT {name: c.name, type: c.type}) AS columns
ORDER BY t.name
")
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    rows = data['results'][0]['data']
    print('## 数据库表')
    print('')
    if not rows:
        print('无数据库表')
    else:
        print('| 表名 | 列 |')
        print('|------|-----|')
        for r in rows:
            cols = r['row'][1]
            col_strs = [f'{c[\"name\"]} ({c[\"type\"]})' for c in cols if c.get('name')]
            display = ', '.join(col_strs[:15]) if col_strs else '-'
            if len(col_strs) > 15:
                display += ' ...'
            print(f'| {r[\"row\"][0]} | {display} |')
    print('')
except Exception as e:
    print('> ⚠️ 查询失败:', e)
    print('')
"

###############################################################################
# Query 6: Frontend consumers — API calls proxying to this service's routes
###############################################################################
RESULT=$(neo4j_query "
MATCH (s:Service {id: \$svc})-[:EXPOSES]->(r:ApiRoute)
MATCH (ac:ApiCall)-[:PROXIES_TO]->(r)
RETURN DISTINCT ac.method AS method, ac.url AS url, ac.file AS file
LIMIT 30
")

ROW_COUNT=$(echo "$RESULT" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['results'][0]['data']))")

echo '## 前端消费方'
echo ''

if [ "$ROW_COUNT" -gt 0 ]; then
    echo "$RESULT" | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data['results'][0]['data']
print('| 方法 | URL | 文件 |')
print('|------|-----|------|')
for r in rows:
    print(f'| {r[\"row\"][0]} | {r[\"row\"][1]} | {r[\"row\"][2]} |')
"
else
    echo '> ⚠️ 未匹配到服务特定路由。列出所有前端 API 调用：'
    echo ''
    echo '| 方法 | URL | 文件 |'
    echo '|------|-----|------|'
    FALLBACK_RESULT=$(neo4j_query "MATCH (ac:ApiCall) RETURN ac.name AS name, ac.method AS method, ac.url AS url, ac.file AS file ORDER BY ac.name")
    echo "$FALLBACK_RESULT" | python3 -c "
import sys, json
data = json.load(sys.stdin)
rows = data['results'][0]['data']
for r in rows:
    print(f'| {r[\"row\"][1]} | {r[\"row\"][2]} | {r[\"row\"][3]} |')
"
fi
echo ''

###############################################################################
# Query 7: Proto -> Go -> DB lineage
###############################################################################
RESULT=$(neo4j_query "
MATCH (pm:ProtoMessage)-[:GENERATES]->(gs:GoStruct)
WHERE gs.file CONTAINS \"$SERVICE_NAME\"
  AND NOT pm.name ENDS WITH 'Req'
  AND NOT pm.name ENDS WITH 'Resp'
OPTIONAL MATCH (gs)-[:MAPS_TO]->(t:DbTable)
RETURN pm.name AS proto, gs.name AS goStruct, collect(DISTINCT t.name) AS tables
ORDER BY pm.name
LIMIT 20
")
echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    rows = data['results'][0]['data']
    print('## 实体血缘（Proto → Go → DB）')
    print('')
    if not rows:
        print('无实体血缘数据')
    else:
        print('| Proto 消息 | Go 结构体 | 数据库表 |')
        print('|-----------|----------|---------|')
        for r in rows:
            tables = ', '.join(r['row'][2]) if r['row'][2] else '-'
            print(f'| {r[\"row\"][0]} | {r[\"row\"][1]} | {tables} |')
    print('')
except Exception as e:
    print('> ⚠️ 查询失败:', e)
    print('')
"

echo "---"
echo "*此文件由 graph-sync.sh 自动生成，请勿手动编辑。*"

# ── 打点：记录图谱查询调用（服务 + 成功）──
bash "$(dirname "$0")/log-usage.sh" graph-query service="$SERVICE_NAME" result=success 2>/dev/null || true
