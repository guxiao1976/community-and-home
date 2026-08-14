#!/usr/bin/env bash
#
# harness-self-check.sh — Harness 自身一致性检查（meta-CI）
#
# 查「开发流水线自身」的健康度：流程引用 / 命名口径 / 文档同步 / 配置漂移 / 调用链完整。
# 与 harness-checks.sh（服务代码检查）互补：本脚本只查 harness 基础设施，不查业务代码。
#
# 用法:
#   bash .harness/scripts/harness-self-check.sh            # 全量
#   bash .harness/scripts/harness-self-check.sh --json     # JSON 输出
#
# 返回码: 0 = 全部通过；1 = 有 FAIL；2 = 有 WARN（不阻断）
#
# 接入:
#   - pre-commit：改动 .harness/workflows/ 或 .harness/skills/ 时自动跑
#   - pipeline-review：维度 1 文档新鲜度增强（自动执行）

set -uo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

OUTPUT_JSON=false
[[ "$*" == *--json* ]] && OUTPUT_JSON=true

PASS=0; FAIL=0; WARN=0
declare -a RESULTS

pass() { PASS=$((PASS+1)); RESULTS+=("PASS|$1|$2"); }
fail() { FAIL=$((FAIL+1)); RESULTS+=("FAIL|$1|$2"); }
warn() { WARN=$((WARN+1)); RESULTS+=("WARN|$1|$2"); }

# 检查 1: 流程引用存在（skill/workflow/agent 引用的脚本路径必须存在）
check_refs() {
  local missing=0
  # 关键脚本/文件必须存在（调用链完整性）
  local required=(
    ".harness/workflows/harness-spec-pipeline.js"
    ".harness/workflows/harness-pipeline.js"
    ".harness/workflows/gate-engine.js"
    ".harness/scripts/build-pipeline.sh"
    ".harness/skills/dispatch.md"
    ".harness/skills/review.md"
    ".harness/skills/requirement-analysis.md"
    ".harness/skills/architect-design.md"
    ".harness/agents/owner-agent.md"
    ".harness/agents/subagents/requirement-analyst.md"
    ".harness/agents/subagents/architecture-designer.md"
    ".harness/registry/services.json"
  )
  for f in "${required[@]}"; do
    if [ ! -f "$f" ]; then
      missing=$((missing+1)); fail "refs_exist" "$f 缺失"
    fi
  done
  [ "$missing" -eq 0 ] && pass "refs_exist" "关键文件全部存在"
}

# 检查 2: 命名口径一致（skill 无残留旧流程词）
check_naming() {
  local old_words=(
    "并行 N×Workflow"
    "派发需求分析子 Agent"
    "派发架构设计子 Agent"
    "Dev Agent 路径"
  )
  local hits=0
  for w in "${old_words[@]}"; do
    local found
    found=$(grep -rln "$w" .harness/skills/*.md .harness/agents/owner-agent.md 2>/dev/null | grep -v "_archive" | wc -l)
    if [ "$found" -gt 0 ]; then
      hits=$((hits+1)); warn "naming" "skill/owner-agent 残留旧流程词「$w」($found 处)"
    fi
  done
  [ "$hits" -eq 0 ] && pass "naming" "命名口径一致（无旧流程词）"
}

# 检查 3: 文档同步（总纲 + 核心文档应提到当前 spec-pipeline）
check_docs() {
  local docs=(
    ".harness/docs/harness-architecture.md"
    ".harness/docs/pipeline-architecture.md"
    ".harness/docs/pipeline-flow-complete.md"
    ".harness/docs/pipeline-patterns.md"
    ".harness/docs/pipeline-evolution.md"
  )
  local stale=0
  for f in "${docs[@]}"; do
    if [ ! -f "$f" ] || ! grep -q "spec-pipeline" "$f" 2>/dev/null; then
      stale=$((stale+1)); warn "docs_sync" "$f 未提到 spec-pipeline（可能过时）"
    fi
  done
  [ "$stale" -eq 0 ] && pass "docs_sync" "总纲 + 核心文档均同步 spec-pipeline"
}

# 检查 4: 配置漂移（quality-gates.yml 阶段标注 vs gate-engine 实现）
check_config_drift() {
  local yml="$ROOT/.harness/config/quality-gates.yml"
  local ge="$ROOT/.harness/workflows/gate-engine.js"
  local drift=0
  # yml 标 not-implemented 但 gate-engine 已实现 → 漂移
  if grep -q "status: not-implemented" "$yml" 2>/dev/null; then
    local not_impl
    not_impl=$(grep -c "status: not-implemented" "$yml")
    # gate-engine 里实现了的 phase（简化：requirement_analysis/architecture_design 已实现）
    for phase in requirement_analysis architecture_design integration; do
      if grep -q "not-implemented.*$phase\|$phase.*not-implemented" "$yml" 2>/dev/null; then
        drift=$((drift+1))
      fi
    done
  fi
  [ "$drift" -eq 0 ] && pass "config_drift" "quality-gates.yml 无 not-implemented 残留漂移" || warn "config_drift" "$drift 处阶段标注与 gate-engine 实现不一致"
}

# 检查 5: 调用链完整（dispatch → spec-pipeline → harness-pipeline 契约）
check_chain() {
  local ok=1
  # dispatch 应引用 spec-pipeline（L 级路由）
  grep -q "spec-pipeline" .harness/skills/dispatch.md || { fail "chain" "dispatch.md 未引用 spec-pipeline"; ok=0; }
  # spec-pipeline 应存在并引用 harness-pipeline（阶段 5 委托）
  grep -q "harness-pipeline" .harness/workflows/harness-spec-pipeline.js || { warn "chain" "spec-pipeline 未引用 harness-pipeline（阶段5委托缺失）"; ok=0; }
  # gate-engine 应被 workflow 引用
  grep -q "gate-engine" .harness/workflows/harness-spec-pipeline.js || { warn "chain" "spec-pipeline 未引用 gate-engine"; ok=0; }
  [ "$ok" -eq 1 ] && pass "chain" "dispatch → spec-pipeline → harness-pipeline 调用链完整"
}

# 检查 6: registry ↔ dispatch 服务名同步（新增服务后别名表不得漂移）
check_registry() {
  local reg="$ROOT/.harness/registry/services.json"
  local dispatch="$ROOT/.harness/skills/dispatch.md"
  if [ ! -f "$reg" ]; then fail "registry_sync" "services.json 缺失"; return; fi
  if [ ! -f "$dispatch" ]; then fail "registry_sync" "dispatch.md 缺失"; return; fi
  local missing=0 svc
  while read -r svc; do
    [ -z "$svc" ] && continue
    if ! grep -qE "\`$svc\`" "$dispatch"; then
      missing=$((missing+1)); fail "registry_sync" "dispatch 别名表缺少服务 $svc（registry 已注册）"
    fi
  done < <(python3 -c "import json,sys; print('\n'.join(s['name'] for s in json.load(open(sys.argv[1]))['services']))" "$reg" 2>/dev/null)
  [ "$missing" -eq 0 ] && pass "registry_sync" "registry ↔ dispatch 服务名同步"
}

# 检查 7: changes 追溯链完整性（INDEX 覆盖全部变更目录 + 无悬空链接）
check_traceability() {
  local index="$ROOT/.harness/changes/INDEX.md"
  local cdir="$ROOT/.harness/changes"
  if [ ! -f "$index" ]; then fail "traceability" "changes/INDEX.md 缺失"; return; fi
  local missing=0 dangling=0
  # 实际变更目录（排除 _archive/TEMPLATE）是否都登记到 INDEX
  local indexed
  indexed=$(grep -o '\./[a-z0-9-]*/' "$index" | sed 's|^\./||; s|/$||' | grep -v "change-name" | sort -u)
  local dname
  for d in "$cdir"/*/; do
    dname=$(basename "$d")
    [[ "$dname" == "_archive" || "$dname" == "TEMPLATE" ]] && continue
    if ! echo "$indexed" | grep -qx "$dname"; then
      missing=$((missing+1)); fail "traceability" "变更目录 $dname 未登记到 INDEX"
    fi
  done
  # INDEX 登记的目录是否真实存在（无悬空链接）
  local link
  for link in $indexed; do
    [ -d "$cdir/$link" ] || { dangling=$((dangling+1)); fail "traceability" "INDEX 悬空链接 $link（目录不存在）"; }
  done
  [ "$missing" -eq 0 ] && [ "$dangling" -eq 0 ] && pass "traceability" "changes 追溯链完整（INDEX 全覆盖、无悬空）"
}

# ── 执行 ──
check_refs
check_naming
check_docs
check_config_drift
check_chain
check_registry
check_traceability

# ── 输出 ──
if $OUTPUT_JSON; then
  printf '{"total":%d,"pass":%d,"fail":%d,"warn":%d,"results":[' "$((PASS+FAIL+WARN))" "$PASS" "$FAIL" "$WARN"
  local i=0
  for r in "${RESULTS[@]}"; do
    [ "$i" -gt 0 ] && printf ','
    printf '{"status":"%s","check":"%s","detail":"%s"}' "${r%%|*}" "$(echo "$r" | cut -d'|' -f2)" "$(echo "$r" | cut -d'|' -f3)"
    i=$((i+1))
  done
  printf ']}\n'
else
  echo "=== Harness 自检 ==="
  for r in "${RESULTS[@]}"; do
    echo "  [${r%%|*}] $(echo "$r" | cut -d'|' -f2) — $(echo "$r" | cut -d'|' -f3)"
  done
  echo ""
  echo "=== 摘要: $PASS PASS / $FAIL FAIL / $WARN WARN ==="
fi

# 返回码: FAIL>0 → 1；WARN>0 → 2；全过 → 0
if [ "$FAIL" -gt 0 ]; then exit 1; fi
if [ "$WARN" -gt 0 ]; then exit 2; fi
exit 0
