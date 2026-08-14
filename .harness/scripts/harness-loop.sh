#!/usr/bin/env bash
#
# harness-loop.sh — Harness 自动发现循环
#
# 职责：定期扫描传感器，发现新问题，输出可执行的派发计划。
# 不直接修改代码，只更新 BACKLOG 和输出派发计划。
#
# 用法:
#   bash .harness/scripts/harness-loop.sh                     # 扫描 + 输出计划
#   bash .harness/scripts/harness-loop.sh --auto-create       # 扫描 + 自动建任务 + 输出计划
#   bash .harness/scripts/harness-loop.sh --json              # JSON 输出（供 SessionStart hook 消费）
#   bash .harness/scripts/harness-loop.sh --summary           # 仅摘要
#
# 返回码:
#   0 — 无新问题，无待办需要执行
#   1 — 有新问题或有待办需要执行（通知调用方）

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TASKS_SCRIPT="$PROJECT_ROOT/.harness/scripts/harness-tasks.sh"
TASKS_DIR="$PROJECT_ROOT/.harness/tasks"
BACKLOG="$PROJECT_ROOT/.harness/tasks/BACKLOG.md"
LOOP_RUNS_DIR="$PROJECT_ROOT/.harness/loop-runs"
RUN_ID="run-$(date +%Y-%m-%d-%H%M%S)"
RUN_LOG="$LOOP_RUNS_DIR/${RUN_ID}.md"
NOTIFY_SCRIPT="$PROJECT_ROOT/.harness/scripts/notify.sh"

# Counters for run log summary
NEW_ISSUES=0
DISPATCHED_COUNT=0

AUTO_CREATE=false
OUTPUT_JSON=false
SUMMARY_ONLY=false
AUTO_DISPATCH=false
P1_MAX=2   # P1 自动派发上限（每次 loop 最多派发几个 P1；P0 不设限）

while [[ $# -gt 0 ]]; do
  case "$1" in
    --auto-create) AUTO_CREATE=true; shift ;;
    --auto-dispatch) AUTO_DISPATCH=true; shift ;;
    --p1-max) P1_MAX="${2:-2}"; shift 2 ;;
    --json) OUTPUT_JSON=true; shift ;;
    --summary) SUMMARY_ONLY=true; shift ;;
    *) shift ;;
  esac
done

timestamp() { date -u +"%Y-%m-%dT%H:%M:%SZ"; }

# ─── Run Log Functions ─────────────────────────────────────────────

init_run_log() {
  mkdir -p "$LOOP_RUNS_DIR"
  local trigger="扫描"
  $AUTO_DISPATCH && trigger="自动派发"
  $AUTO_CREATE && trigger="${trigger} + 自动建任务"

  cat > "$RUN_LOG" <<LOGEOF
# Loop Run: ${RUN_ID}

**时间**: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
**触发方式**: ${trigger}

## 传感器扫描结果

| 传感器 | 结果 | 详情 |
|--------|------|------|
LOGEOF
}

log_sensor_result() {
  local sensor="$1" result="$2" detail="${3:-}"
  echo "| ${sensor} | ${result} | ${detail} |" >> "$RUN_LOG"
}

log_dispatch_action() {
  local task_id="$1" action="$2" detail="${3:-}"
  echo "| $(timestamp) | ${task_id} | ${action} | ${detail} |" >> "$RUN_LOG"
}

finalize_run_log() {
  local exit_code="${1:-0}"
  cat >> "$RUN_LOG" <<LOGEOF

## 总结

- 新发现任务: ${NEW_ISSUES:-0}
- 自动派发: ${DISPATCHED_COUNT:-0}
- 退出码: ${exit_code}

---
LOGEOF

  # Auto-cleanup: keep last 50 runs
  ls -t "$LOOP_RUNS_DIR"/run-*.md 2>/dev/null | tail -n +51 | xargs rm -f 2>/dev/null || true
}

# ─── Step 1: 传感器扫描 ────────────────────────────────────────────

scan_sensors() {
  local scan_args=""
  $AUTO_CREATE && scan_args="--auto-create"

  local scan_output

  if $OUTPUT_JSON; then
    scan_output=$(bash "$TASKS_SCRIPT" scan $scan_args 2>&1) || true
    echo "$scan_output"
  elif $SUMMARY_ONLY; then
    scan_output=$(bash "$TASKS_SCRIPT" scan $scan_args 2>&1) || true
    echo "$scan_output" | grep -E "(Found|NEW|STALE|Scan complete|Sensor)" || true
  else
    echo "━━━ Step 1: 传感器扫描 ━━━"
    scan_output=$(bash "$TASKS_SCRIPT" scan $scan_args 2>&1) || true
    echo "$scan_output"
  fi

  # Parse and log sensor results to run log
  _parse_sensor_results "$scan_output"
}

_parse_sensor_results() {
  local output="$1"

  # Sensor 1: QA mechanized checks
  local s1=$(echo "$output" | grep -oP "Found \d+ FAIL" | head -1 | grep -oP '\d+' || echo "0")
  log_sensor_result "Sensor 1: QA checks" "$([[ "$s1" != "0" ]] && echo "${s1} FAIL" || echo "PASS")" ""

  # Sensor 2: TODO stubs
  local s2=$(echo "$output" | grep -oP "Found \d+ TODO stub" | head -1 | grep -oP '\d+' || echo "0")
  log_sensor_result "Sensor 2: TODO stubs" "$([[ "$s2" != "0" ]] && echo "${s2} found" || echo "None")" ""

  # Sensor 3: Review WARNINGs
  local s3=$(echo "$output" | grep -oP "\d+ unfixed WARNING" | head -1 | grep -oP '\d+' || echo "0")
  log_sensor_result "Sensor 3: Review WARNINGs" "$([[ "$s3" != "0" ]] && echo "${s3} unfixed" || echo "None")" ""

  # Sensor 4: Graph freshness
  if echo "$output" | grep -q "STALE"; then
    local age=$(echo "$output" | grep -oP '\d+h' | head -1 || echo "?")
    log_sensor_result "Sensor 4: Graph freshness" "STALE" "${age} since last sync"
  elif echo "$output" | grep -q "Fresh"; then
    log_sensor_result "Sensor 4: Graph freshness" "Fresh" ""
  else
    log_sensor_result "Sensor 4: Graph freshness" "Unknown" ""
  fi

  # Sensor 5: GitHub Issues
  local s5=$(echo "$output" | grep "open issues:" | grep -oP '\d+' | head -1 || echo "0")
  local s5_new=$(echo "$output" | grep -c "NEW: #" || true)
  log_sensor_result "Sensor 5: GitHub Issues" "$([[ ${s5_new:-0} -gt 0 ]] && echo "${s5_new} new" || echo "${s5} open")" ""

  # Sensor 6: PR Reviews
  local s6=$(echo "$output" | grep "Open PRs:" | grep -oP '\d+' | head -1 || echo "0")
  local s6_new=$(echo "$output" | grep -c "NEW: PR #" || true)
  log_sensor_result "Sensor 6: PR Reviews" "$([[ ${s6_new:-0} -gt 0 ]] && echo "${s6_new} changes req" || echo "${s6} open")" ""

  # Capture NEW issues count
  NEW_ISSUES=$(echo "$output" | grep -oP "Scan complete: \K\d+" | head -1 || echo "0")
}

# ─── Service Name Mapping（数据源：.harness/registry/services.json）───────

# 从 registry/services.json 加载 服务名 ↔ 中文名 映射（单一数据源，build-service-registry.sh 生成）。
# 要求 bash 4+（关联数组）；registry 缺失时退化显示原名，不影响主流程。
declare -A SVC_LABEL=()           # name → displayName
declare -A SVC_NAME_BY_LABEL=()   # displayName → name

load_service_registry() {
  local reg="$PROJECT_ROOT/.harness/registry/services.json"
  [[ -f "$reg" ]] || return 0
  while IFS=$'\t' read -r name label; do
    SVC_LABEL["$name"]="$label"
    SVC_NAME_BY_LABEL["$label"]="$name"
  done < <(python3 -c '
import json, sys
d = json.load(open(sys.argv[1]))
for s in d["services"]:
    print(s["name"] + "\t" + s["displayName"])
' "$reg" 2>/dev/null)
}
load_service_registry

service_label() {
  local svc="${1:-}"
  [[ "$svc" == "all" ]] && { echo "全局"; return; }
  [[ -n "${SVC_LABEL[$svc]+x}" ]] && { echo "${SVC_LABEL[$svc]}"; return; }
  echo "${svc:-unknown}"
}

# Resolve service directory from task file's service field
# Handles: "services/xxx" -> "xxx", "xxx" -> "xxx", Chinese displayName
resolve_service_dir() {
  local svc="$1"
  # Strip services/ prefix if present
  svc="${svc#services/}"
  [[ "$svc" == "all" || "$svc" == "全局" ]] && { echo "services/all"; return; }
  if [[ -n "${SVC_LABEL[$svc]+x}" ]]; then
    echo "services/$svc"
  elif [[ -n "${SVC_NAME_BY_LABEL[$svc]+x}" ]]; then
    echo "services/${SVC_NAME_BY_LABEL[$svc]}"
  else
    echo "services/${svc}"
  fi
}

# ─── Auto-Dispatch ────────────────────────────────────────────────────

auto_dispatch_tasks() {
  # Collect P0 auto-tasks (source: qa|review|sensor|github, status: open)
  # FIX(闭环断裂): 旧实现只派发 P0，P1 的 qa/sensor 任务（如 graph_freshness）永远积压。
  # 现在 P0 全量派发 + P1 限量派发（--p1-max，默认 2 个/次），避免 backlog 只进不出。
  local p0_auto p1_auto combined=""
  p0_auto=$(bash "$TASKS_SCRIPT" list --priority P0 --status open 2>/dev/null \
    | grep -E "(review|qa|sensor|github)" || true)
  p1_auto=$(bash "$TASKS_SCRIPT" list --priority P1 --status open 2>/dev/null \
    | grep -E "(review|qa|sensor|github)" || true)

  if [[ -n "${p0_auto// }" ]]; then
    combined+="$p0_auto"$'\n'
  fi
  if [[ -n "${p1_auto// }" ]]; then
    local p1_count=0
    while IFS= read -r line; do
      [[ -z "${line// }" ]] && continue
      p1_count=$((p1_count + 1))
      [[ $p1_count -gt $P1_MAX ]] && break
      combined+="$line"$'\n'
    done <<< "$p1_auto"
  fi

  if [[ -z "${combined// }" ]]; then
    return 0
  fi

  echo ""
  echo "━━━ Auto-Dispatch: P0 全量 + P1 限量(≤${P1_MAX}) 自动派发 ━━━"

  while IFS= read -r line; do
    [[ -z "${line// }" ]] && continue

    local task_id=$(echo "$line" | awk '{print $1}')
    local task_prio=$(echo "$line" | awk '{print $2}')
    local task_svc=$(echo "$line" | awk '{print $4}')
    local task_title=$(echo "$line" | cut -c 65- | sed 's/^[[:space:]]*//')

    # Read triage field from task file
    local task_file="$TASKS_DIR/${task_id}.md"
    local task_triage=""
    if [[ -f "$task_file" ]]; then
      task_triage=$(grep -oP '(?<=^triage: ").*' "$task_file" | head -1 | sed 's/"$//' || echo "")
    fi

    # If triaged as non-auto-fixable, skip
    if [[ -n "$task_triage" ]] && [[ "$task_triage" != "auto-fixable" ]] && [[ "$task_triage" != '""' ]]; then
      echo "  SKIP $task_id ($task_title) — triage: $task_triage (非 auto-fixable)"
      log_dispatch_action "$task_id" "SKIP" "triage: $task_triage"
      continue
    fi

    # ── Convergence guard: skip tasks dispatched too many times ──
    local dispatch_count=0
    if [[ -f "$task_file" ]]; then
      dispatch_count=$(grep -oP '(?<=^dispatch_count: ).*' "$task_file" | head -1 || echo "0")
    fi
    dispatch_count=$(( dispatch_count + 0 ))  # ensure integer, empty → 0
    local MAX_DISPATCH=3
    if [[ $dispatch_count -ge $MAX_DISPATCH ]]; then
      echo "  🛑 BLOCK $task_id — 已派发 ${dispatch_count} 次未完成，升级给人"
      bash "$TASKS_SCRIPT" status --id "$task_id" --status blocked 2>/dev/null || true
      log_dispatch_action "$task_id" "ESCALATE" "派发 ${dispatch_count} 次未完成 → blocked"
      continue
    fi

    # Resolve service name and directory
    local svc_en="${task_svc}"
    # If it's a Chinese name, resolve it
    case "$task_svc" in
      用户服务) svc_en="user-service" ;;
      认证服务) svc_en="auth-service" ;;
      权限服务) svc_en="permission-service" ;;
      文件服务) svc_en="file-service" ;;
      AI模型服务) svc_en="ai-model-service" ;;
      主数据服务) svc_en="master-data-service" ;;
      内容审核服务) svc_en="moderation-service" ;;
      社区枢纽服务) svc_en="community-hub-service" ;;
      监控服务) svc_en="monitoring-service" ;;
      all) svc_en="all" ;;
    esac

    local svc_dir=$(resolve_service_dir "$task_svc")
    local svc_label=$(service_label "${svc_en}")

    # Mark task as in_progress
    bash "$TASKS_SCRIPT" status --id "$task_id" --status in_progress 2>/dev/null || true

    # Increment dispatch_count (convergence tracking)
    if [[ -f "$task_file" ]]; then
      local new_dispatch_count=$(( dispatch_count + 1 ))
      sed -i "s/^dispatch_count: .*/dispatch_count: ${new_dispatch_count}/" "$task_file"
    fi

    # Update assigned_run in task file
    if [[ -f "$task_file" ]]; then
      sed -i "s/^assigned_run: \"\"$/assigned_run: \"${RUN_ID}\"/" "$task_file"
    fi

    # Output dispatch directive for agent consumption
    # Read task type for pipeline routing
    local task_type=$(grep -oP '(?<=^type: ).*' "$task_file" | head -1 || echo "feature")
    # Read workload hint (optional; dispatch uses it if present, else default table)
    local task_workload=$(grep -oP '(?<=^workload: ").*?(?=")' "$task_file" | head -1 || echo "")
    if $OUTPUT_JSON; then
      echo "{\"action\":\"dispatch\",\"id\":\"${task_id}\",\"service\":\"${svc_en}\",\"label\":\"${svc_label}\",\"dir\":\"${svc_dir}\",\"type\":\"${task_type}\",\"workload\":\"${task_workload}\",\"task\":\"${task_title}\"}"
    elif $SUMMARY_ONLY; then
      echo "[DISPATCH] id=${task_id} type=${task_type} service=${svc_en} label=${svc_label} dir=${svc_dir} task=${task_title} workload=${task_workload}"
    else
      echo "  🚀 派发: $task_id → $svc_label"
      echo "  [DISPATCH] id=${task_id} type=${task_type} service=${svc_en} label=${svc_label} dir=${svc_dir} task=${task_title} workload=${task_workload}"
    fi

    log_dispatch_action "$task_id" "DISPATCH" "${svc_label}: ${task_title}"
    DISPATCHED_COUNT=$((DISPATCHED_COUNT + 1))

  done <<< "$p0_auto"

  if [[ $DISPATCHED_COUNT -eq 0 ]]; then
    echo "  （无 P0 可自动派发任务）"
  else
    echo ""
    echo "  ✅ 已派发 $DISPATCHED_COUNT 个 P0 任务"
  fi
}

# ─── Step 2: 生成派发计划 ──────────────────────────────────────────

dispatch_plan() {
  local candidates=""

  # 收集 P0 自动任务（source: qa|review|sensor）
  local p0_auto
  p0_auto=$(bash "$TASKS_SCRIPT" list --priority P0 --status open 2>/dev/null \
    | grep -E "(review|qa|sensor)" || true)

  # 收集 P1 自动任务
  local p1_auto
  p1_auto=$(bash "$TASKS_SCRIPT" list --priority P1 --status open 2>/dev/null \
    | grep -E "(review|qa|sensor)" || true)

  # 收集 P0/P1 人工任务（需要确认）
  local human_tasks
  human_tasks=$(bash "$TASKS_SCRIPT" list --priority P0 --status open 2>/dev/null \
    | grep "human" || true)
  human_tasks+=$'\n'
  human_tasks+=$(bash "$TASKS_SCRIPT" list --priority P1 --status open 2>/dev/null \
    | grep "human" || true)

  if $OUTPUT_JSON; then
    # Build JSON output
    echo "{"
    echo "  \"timestamp\": \"$(timestamp)\","
    echo "  \"auto_dispatch\": ["
    local first=true
    while IFS= read -r line; do
      [[ -z "${line// }" ]] && continue
      local id=$(echo "$line" | awk '{print $1}')
      local prio=$(echo "$line" | awk '{print $2}')
      local svc=$(echo "$line" | awk '{print $4}')
      local title=$(echo "$line" | cut -c 65-)
      if $first; then first=false; else echo "    ,"; fi
      printf '    {"id":"%s","priority":"%s","service":"%s","title":"%s"}' "$id" "$prio" "$svc" "$title"
    done < <(echo "$p0_auto"; echo "$p1_auto")
    echo ""
    echo "  ],"
    echo "  \"human_confirm\": ["
    first=true
    while IFS= read -r line; do
      [[ -z "${line// }" ]] && continue
      local id=$(echo "$line" | awk '{print $1}')
      local prio=$(echo "$line" | awk '{print $2}')
      local svc=$(echo "$line" | awk '{print $4}')
      local title=$(echo "$line" | cut -c 65-)
      if $first; then first=false; else echo "    ,"; fi
      printf '    {"id":"%s","priority":"%s","service":"%s","title":"%s"}' "$id" "$prio" "$svc" "$title"
    done < <(echo "$human_tasks")
    echo ""
    echo "  ]"
    echo "}"
  else
    echo ""
    echo "━━━ Step 2: 派发计划 ━━━"

    local auto_count=0 human_count=0

    # 自动派发候选人
    echo ""
    echo "🔧 可自动派发（source: qa|review|sensor）："
    echo ""
    if [[ -n "${p0_auto// }" ]]; then
      echo "  P0:"
      echo "$p0_auto" | while IFS= read -r line; do
        [[ -z "${line// }" ]] && continue
        echo "    $line"
      done
      auto_count=$((auto_count + $(echo "$p0_auto" | grep -c . || true)))
    fi
    if [[ -n "${p1_auto// }" ]]; then
      echo "  P1:"
      echo "$p1_auto" | while IFS= read -r line; do
        [[ -z "${line// }" ]] && continue
        echo "    $line"
      done
      auto_count=$((auto_count + $(echo "$p1_auto" | grep -c . || true)))
    fi
    [[ $auto_count -eq 0 ]] && echo "  （暂无）"

    # 需要人工确认的任务
    echo ""
    echo "👤 需要人工确认优先级（source: human）："
    echo ""
    local non_empty_human=$(echo "$human_tasks" | grep -v '^$' || true)
    if [[ -n "${non_empty_human// }" ]]; then
      echo "$non_empty_human" | while IFS= read -r line; do
        [[ -z "${line// }" ]] && continue
        echo "    $line"
      done
      human_count=$(echo "$non_empty_human" | grep -c . || true)
    else
      echo "  （暂无）"
    fi

    echo ""
    echo "━━━ 摘要 ━━━"
    echo "  可自动派发: $auto_count  需人工确认: $human_count"
    echo ""

    if [[ $auto_count -gt 0 || $human_count -gt 0 ]]; then
      echo "💡 说「执行派发计划」启动管线，或说「派发 <任务ID>」执行单个"
    fi
  fi

  # Return code: non-zero if work exists
  local total=$(( $(echo "$p0_auto" | grep -c . || true) + $(echo "$p1_auto" | grep -c . || true) ))
  return $(( total > 0 ? 1 : 0 ))
}

# ─── Main ───────────────────────────────────────────────────────────

init_run_log

if ! $OUTPUT_JSON && ! $SUMMARY_ONLY; then
  echo "╔══════════════════════════════════════╗"
  echo "║   Harness Loop — $(date "+%Y-%m-%d %H:%M")"
  echo "╚══════════════════════════════════════╝"
  echo ""
fi

scan_sensors

# dispatch_plan may return non-zero (signals "work exists")
# Use || pattern to prevent set -e from killing the script
EXIT_CODE=0
dispatch_plan || EXIT_CODE=$?

# Auto-dispatch P0 tasks if --auto-dispatch flag is set
if $AUTO_DISPATCH; then
  auto_dispatch_tasks
  # Re-evaluate exit code: signal if we dispatched something
  if [[ $DISPATCHED_COUNT -gt 0 ]]; then
    EXIT_CODE=1
  fi
fi

finalize_run_log "$EXIT_CODE"

# Notify on loop complete if there were issues or dispatches
if [[ $EXIT_CODE -ne 0 ]] && [[ -x "$NOTIFY_SCRIPT" ]]; then
  "$NOTIFY_SCRIPT" --event loop_complete --detail "发现 ${NEW_ISSUES} 新问题, 派发 ${DISPATCHED_COUNT} 任务" 2>/dev/null || true
fi

if ! $OUTPUT_JSON && ! $SUMMARY_ONLY; then
  if [[ $EXIT_CODE -eq 0 ]]; then
    echo "✅ 无待办，一切正常。"
  fi
fi

exit $EXIT_CODE
