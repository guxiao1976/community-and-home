#!/usr/bin/env bash
#
# harness-tasks.sh — Task backlog operations for the Harness pipeline.
#
# Usage:
#   harness-tasks.sh list [--priority P0|P1|P2|P3] [--status open|in_progress|closed]
#                         [--service <name>] [--source human|qa|review|sensor]
#   harness-tasks.sh create --title "..." --service <name> --priority P0|P1|P2|P3
#                           --type feature|bug|debt|chore [--source human|qa|review|sensor]
#                           [--detail "..."] [--workload S|M|L]
#   harness-tasks.sh status --id <task-id> --status open|in_progress|review|closed|blocked
#   harness-tasks.sh scan  [--service <name>] [--auto-create]
#   harness-tasks.sh index
#   harness-tasks.sh stats
#
# Examples:
#   harness-tasks.sh list --priority P0
#   harness-tasks.sh list --service moderation-service --status open
#   harness-tasks.sh create --title "修复编译错误" --service moderation-service --priority P0 --type bug
#   harness-tasks.sh status --id task-2026-06-16-001 --status in_progress
#   harness-tasks.sh scan --auto-create
#   harness-tasks.sh index

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TASKS_DIR="$PROJECT_ROOT/.harness/tasks"
TEMPLATE="$TASKS_DIR/TEMPLATE.md"
BACKLOG="$TASKS_DIR/BACKLOG.md"
QA_SCRIPT="$PROJECT_ROOT/.harness/skills/qa/scripts/harness-checks.sh"

# Auto-load .env if GITHUB_TOKEN not already in environment
if [[ -z "${GITHUB_TOKEN:-}" ]] && [[ -f "$PROJECT_ROOT/.env" ]]; then
  export GITHUB_TOKEN=$(grep '^GITHUB_TOKEN=' "$PROJECT_ROOT/.env" | cut -d= -f2-)
fi

# ─── Helpers ──────────────────────────────────────────────────────────

die() { echo "ERROR: $*" >&2; exit 1; }

today() { date +"%Y-%m-%d"; }

# Generate next task ID: task-YYYY-MM-DD-NNN
next_id() {
  local date="$(today)"
  local max_n=0
  for f in "$TASKS_DIR"/task-"${date}"-*.md; do
    [[ -f "$f" ]] || continue
    local n=$(basename "$f" .md | grep -oP '\d{3}$' || echo 0)
    # strip leading zeros to avoid octal interpretation
    n=$((10#${n}))
    [[ $n -gt $max_n ]] && max_n=$n
  done
  printf "task-%s-%03d" "$date" "$((max_n + 1))"
}

# Extract frontmatter field value from a task file
get_field() {
  local file="$1" field="$2"
  grep -oP "(?<=^${field}: ).*" "$file" | head -1 | sed 's/^"//;s/"$//'
}

# ─── list ──────────────────────────────────────────────────────────────

cmd_list() {
  local filter_priority="" filter_status="" filter_service="" filter_source=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --priority) filter_priority="$2"; shift 2 ;;
      --status)   filter_status="$2"; shift 2 ;;
      --service)  filter_service="$2"; shift 2 ;;
      --source)   filter_source="$2"; shift 2 ;;
      *) shift ;;
    esac
  done

  local found=0
  for f in "$TASKS_DIR"/task-*.md; do
    [[ -f "$f" ]] || continue
    local prio=$(get_field "$f" "priority")
    local stat=$(get_field "$f" "status")
    local svc=$(get_field "$f" "service")
    local src=$(get_field "$f" "source")
    local title=$(get_field "$f" "title")
    local id=$(get_field "$f" "id")

    [[ -n "$filter_priority" && "$prio" != "$filter_priority" ]] && continue
    [[ -n "$filter_status"   && "$stat" != "$filter_status" ]] && continue
    [[ -n "$filter_service"  && "$svc" != "$filter_service" ]] && continue
    [[ -n "$filter_source"   && "$src" != "$filter_source" ]] && continue

    printf "%-24s %3s %-6s %-22s %-10s %s\n" "$id" "$prio" "$stat" "$svc" "$src" "$title"
    found=$((found + 1))
  done

  if [[ $found -eq 0 ]]; then
    echo "(no tasks match)"
  fi
}

# ─── create ────────────────────────────────────────────────────────────

cmd_create() {
  local title="" service="" priority="P2" type="feature" source="human" detail="" triage="" workload=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --title)    title="$2"; shift 2 ;;
      --service)  service="$2"; shift 2 ;;
      --priority) priority="$2"; shift 2 ;;
      --type)     type="$2"; shift 2 ;;
      --source)   source="$2"; shift 2 ;;
      --detail)   detail="$2"; shift 2 ;;
      --triage)   triage="$2"; shift 2 ;;
      --workload) workload="$2"; shift 2 ;;
      *) shift ;;
    esac
  done

  [[ -z "$title" ]] && die "--title is required"
  [[ -z "$service" ]] && die "--service is required"

  local id=$(next_id)
  local file="$TASKS_DIR/${id}.md"
  local created=$(today)

  cat > "$file" <<TASKEOF
---
id: ${id}
title: "${title}"
service: ${service}
type: ${type}
priority: ${priority}
status: open
source: ${source}
source_detail: "${detail}"
created: ${created}
blocks: []
blocked_by: []
assigned_run: ""
completed: ""
outcome: ""
triage: "${triage}"
workload: "${workload}"
dispatch_count: 0
---

# ${title}

## 问题

${detail:-（待补充）}

## 完成标准

- [ ] （待定义）

## 执行记录

| 日期 | 事件 | 详情 |
|------|------|------|
| ${created} | 创建 | ${detail:-手动创建} |

## 关联

- 来源: ${detail}
TASKEOF

  echo "Created: $file (id: $id)"

  # Auto-update index
  cmd_index --quiet
}

# ─── status ────────────────────────────────────────────────────────────

cmd_status() {
  local id="" new_status=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --id)     id="$2"; shift 2 ;;
      --status) new_status="$2"; shift 2 ;;
      *) shift ;;
    esac
  done

  [[ -z "$id" ]] && die "--id is required"
  [[ -z "$new_status" ]] && die "--status is required"

  # Validate status against the canonical enum (fix: completed/other typos were silently accepted before)
  case "$new_status" in
    open|in_progress|review|closed|blocked) ;;
    *) die "invalid status: '$new_status' (allowed: open|in_progress|review|closed|blocked)" ;;
  esac

  local file="$TASKS_DIR/${id}.md"
  [[ -f "$file" ]] || die "task file not found: $file"

  local old_status=$(get_field "$file" "status")

  # Update status field
  sed -i "s/^status: ${old_status}$/status: ${new_status}/" "$file"

  # If closing, set completed date and reset dispatch_count
  if [[ "$new_status" == "closed" ]]; then
    local existing_completed=$(get_field "$file" "completed")
    if [[ -z "$existing_completed" ]]; then
      sed -i "s/^completed: \"\"$/completed: \"$(today)\"/" "$file"
    fi
    # Reset dispatch counter so re-opening doesn't inherit stale count
    sed -i "s/^dispatch_count: .*/dispatch_count: 0/" "$file"
  fi

  # Append execution record to the END of the table.
  # FIX: old impl used `sed -i "/^|/a..."` which matched EVERY line starting with
  # "|" (including table header), so each status change multiplied the whole table.
  # Now: locate the LAST table data row (| line that is not the |---| separator)
  # and insert the new record right after it.
  local event="状态变更"
  if [[ "$new_status" == "closed" ]]; then event="关闭"; fi
  if [[ "$new_status" == "in_progress" ]]; then event="开始执行"; fi
  local record="| $(today) | ${event} | ${old_status} → ${new_status} |"
  awk -v rec="$record" '
    { lines[NR] = $0; if ($0 ~ /^\|/ && $0 !~ /^\|---/) last = NR }
    END {
      for (i = 1; i <= NR; i++) {
        print lines[i]
        if (i == last) print rec
      }
    }
  ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"

  echo "Updated: $id status: $old_status → $new_status"

  # Auto-update index
  cmd_index --quiet
}

# ─── scan (运行传感器，发现新问题) ────────────────────────────────────

cmd_scan() {
  local svc="" auto_create=false
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --service)     svc="$2"; shift 2 ;;
      --auto-create) auto_create=true; shift ;;
      *) shift ;;
    esac
  done

  echo "=== Harness Task Scanner ==="
  echo "Timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  echo ""

  local new_issues=0

  # ── Sensor 1: Run harness-checks.sh ──
  echo "--- Sensor 1: QA mechanized checks ---"
  local qa_args="--json"
  [[ -n "$svc" ]] && qa_args="$qa_args --service $svc"

  # Checks that are actionable as individual tasks (not just "build failed")
  local ACTIONABLE_CHECKS="proto_jstype json_string cross_service_import hardcoded_secrets graph_freshness claude_structural_data proto_ts_align api_stubs response_wrap"

  if [[ -x "$QA_SCRIPT" ]]; then
    local qa_output
    qa_output=$("$QA_SCRIPT" $qa_args 2>/dev/null) || true

    local fail_count=$(echo "$qa_output" | grep -oP '"status":"FAIL"' | wc -l)
    if [[ $fail_count -gt 0 ]]; then
      echo "  Found $fail_count FAIL item(s)"

      # Parse each FAIL result: extract check name and detail
      while IFS= read -r result_line; do
        [[ -z "$result_line" ]] && continue
        local check=$(echo "$result_line" | grep -oP '"check":"[^"]+"' | cut -d'"' -f4)
        local detail=$(echo "$result_line" | grep -oP '"detail":"[^"]*"' | cut -d'"' -f4 | head -c 150)
        [[ -z "$check" ]] && continue

        # For "all" services, only create tasks for actionable checks
        if [[ -z "$svc" ]]; then
          if ! echo "$ACTIONABLE_CHECKS" | grep -qw "$check"; then
            echo "    SKIP: $check (generic check, need --service to create task)"
            continue
          fi
        fi

        # Dedup: check if task for this check+service already exists.
        # FIX: old key was `qa:${check}` but cmd_create wrote `harness-checks.sh check: ${check}`
        # into source_detail — the two formats never matched, so dedup silently failed and
        # every scan created a NEW duplicate task. Now match BOTH formats (legacy + new).
        local svc_key="${svc:-all}"
        local existing=$(grep -lE "qa:${check}|harness-checks\.sh check: ${check}" "$TASKS_DIR"/task-*.md 2>/dev/null | xargs grep -l "service: ${svc_key}" 2>/dev/null || true)
        if [[ -n "$existing" ]]; then
          local existing_id=$(basename "$existing" .md)
          # Check if the existing task is still open
          local existing_status=$(get_field "$existing" "status")
          if [[ "$existing_status" == "closed" ]]; then
            echo "    REOPEN: $check (was closed in $existing_id, reopening)"
            sed -i "s/^status: closed$/status: open/" "$existing"
          else
            echo "    SKIP: $check (task exists: $existing_id)"
          fi
        else
          local detail_short="${detail:-harness-checks.sh check: $check}"
          echo "    NEW: $check — $detail_short"
          if $auto_create; then
            cmd_create --title "QA FAIL: $check — $detail_short" --service "$svc_key" \
              --priority P1 --type debt --source qa --detail "qa:${check} (harness-checks.sh check: $check)"
          fi
          new_issues=$((new_issues + 1))
        fi
      done < <(echo "$qa_output" | grep -oP '\{[^}]*"status":"FAIL"[^}]*\}')
    else
      echo "  All PASS"
    fi
  else
    echo "  (QA script not found, skipped)"
  fi

  # ── Sensor 2: Check for TODO stubs in API logic ──
  echo "--- Sensor 2: API logic TODO stubs ---"
  local target="$PROJECT_ROOT/services"
  [[ -n "$svc" ]] && target="$PROJECT_ROOT/services/$svc"

  if [[ -d "$target" ]]; then
    local stubs=$(grep -rl "todo: add your logic here" "$target/api/internal/logic/" 2>/dev/null || true)
    if [[ -n "$stubs" ]]; then
      local count=$(echo "$stubs" | wc -l)
      echo "  Found $count TODO stub(s)"
      local existing=$(grep -l "source: qa" "$TASKS_DIR"/task-*.md 2>/dev/null | xargs grep -l "api_stubs" 2>/dev/null || true)
      if [[ -z "$existing" ]]; then
        echo "    NEW: $count TODO stubs (no existing task)"
        if $auto_create; then
          local svc_name="${svc:-all}"
          cmd_create --title "实现 API Logic TODO 桩（$count 个文件）" --service "$svc_name" \
            --priority P1 --type debt --source qa --detail "qa:api_stubs"
        fi
        new_issues=$((new_issues + 1))
      fi
    else
      echo "  No TODO stubs found"
    fi
  fi

  # ── Sensor 3: Check for unfixed review WARNINGs ──
  echo "--- Sensor 3: Unfixed review WARNINGs ---"
  for review_file in "$PROJECT_ROOT"/services/*/_review.md; do
    [[ -f "$review_file" ]] || continue
    local svc_name=$(echo "$review_file" | sed 's|.*/services/||;s|/_review.md||')
    [[ -n "$svc" && "$svc_name" != "$svc" ]] && continue

    # Count unfixed WARNINGs (❌ or ⚠️ markers)
    local unfixed=$(grep -cP '(❌|⚠️).*(未修复|未解决|部分)' "$review_file" 2>/dev/null || true)
    if [[ $unfixed -gt 0 ]]; then
      echo "  $svc_name: $unfixed unfixed WARNING(s) in _review.md"

      local existing=$(grep -l "source: review" "$TASKS_DIR"/task-*.md 2>/dev/null | xargs grep -l "$svc_name" 2>/dev/null || true)
      if [[ -z "$existing" ]]; then
        echo "    NOTE: Review WARNINGs need manual triage (use 'harness-tasks.sh create' to add)"
      fi
    fi
  done

  # ── Sensor 4: Graph freshness ──
  echo "--- Sensor 4: Knowledge graph freshness ---"
  local stamp_file="$PROJECT_ROOT/.harness/.graph_last_sync"
  if [[ -f "$stamp_file" ]]; then
    local stamp=$(cat "$stamp_file")
    local now=$(date +%s)
    local age=$(( (now - stamp) / 3600 ))
    if [[ $age -gt 24 ]]; then
      echo "  STALE: graph last synced ${age}h ago"
      # ── Self-heal: try to auto-sync (with 30min cooldown so we don't hammer Neo4j) ──
      local attempt_file="$PROJECT_ROOT/.harness/.graph_last_attempt"
      local attempt_ts=0
      [[ -f "$attempt_file" ]] && attempt_ts=$(cat "$attempt_file" 2>/dev/null || echo 0)
      if [[ $(( now - attempt_ts )) -gt 1800 ]]; then
        date +%s > "$attempt_file"
        echo "    ↻ 尝试自动同步知识图谱（30 分钟冷却内仅一次）..."
        if bash "$PROJECT_ROOT/.harness/scripts/graph-sync.sh" >/dev/null 2>&1; then
          echo "    ✅ 图谱自动同步成功"
          # stamp updated by graph-sync.sh; re-read for accurate reporting below
          [[ -f "$stamp_file" ]] && age=$(( (now - $(cat "$stamp_file")) / 3600 ))
        else
          echo "    ❌ 自动同步失败（Neo4j 不可达或同步错误）— 请手动检查 graph-sync.sh"
        fi
      fi
      # Re-check freshness after the sync attempt before creating a task
      if [[ $age -gt 24 ]]; then
        local existing=$(grep -l "graph_freshness" "$TASKS_DIR"/task-*.md 2>/dev/null || true)
        if [[ -z "$existing" ]]; then
          if $auto_create; then
            cmd_create --title "同步知识图谱（${age}h 未更新）" --service "global" \
              --priority P2 --type chore --source sensor --detail "sensor:graph_freshness"
          fi
          new_issues=$((new_issues + 1))
        fi
      else
        echo "  ✅ 图谱已通过自动同步恢复 (synced ${age}h ago)"
      fi
    else
      echo "  Fresh (synced ${age}h ago)"
    fi
  else
    echo "  Graph never synced"
  fi

  # ── Sensor 5: GitHub Issues ──
  echo "--- Sensor 5: GitHub Issues ---"
  local github_repo=$(git -C "$PROJECT_ROOT" config --get remote.origin.url 2>/dev/null | grep -oP 'github\.com[:/]\K[^/]+/[^/.]+' || true)
  if [[ -n "$github_repo" ]] && [[ -n "${GITHUB_TOKEN:-}" ]]; then
    local issues_json
    issues_json=$(curl -sf -H "Authorization: Bearer $GITHUB_TOKEN" \
      "https://api.github.com/repos/$github_repo/issues?state=open&per_page=30" 2>/dev/null || true)
    if [[ -n "$issues_json" ]] && echo "$issues_json" | jq -e '. | type == "array"' >/dev/null 2>&1; then
      local issue_count
      issue_count=$(echo "$issues_json" | jq '. | length' 2>/dev/null || echo "0")
      # Safety: issue_count can't exceed API per_page
      if [[ $issue_count -gt 30 ]]; then
        echo "  ⚠️  issue_count=$issue_count exceeds per_page=30, capping"
        issue_count=30
      fi
      echo "  Repo: $github_repo, open issues: $issue_count"
      local new_issue_count=0
      if [[ $issue_count -gt 0 ]]; then
        # Parse each issue: number|title|labels (comma-separated)
        while IFS='|' read -r issue_number issue_title issue_labels; do
          [[ -z "$issue_number" ]] && continue
          local existing=$(grep -l "github:issue:${issue_number}" "$TASKS_DIR"/task-*.md 2>/dev/null || true)
          if [[ -z "$existing" ]]; then
            # Only auto-create for labeled issues (bug, debt, enhancement, fix)
            if echo "$issue_labels" | grep -qP '(bug|debt|enhancement|fix)'; then
              echo "    NEW: #${issue_number} — ${issue_title:-no title}"
              if $auto_create; then
                local label_svc="${svc:-all}"
                cmd_create --title "GitHub #${issue_number}: ${issue_title:-no title}" \
                  --service "$label_svc" --priority P1 --type debt --source github \
                  --detail "github:issue:${issue_number} https://github.com/$github_repo/issues/${issue_number}"
              fi
              new_issue_count=$((new_issue_count + 1))
            fi
          fi
        done < <(echo "$issues_json" | jq -r '.[] | "\(.number)|\(.title // "no title")|\(.labels // [] | map(.name) | join(","))"' 2>/dev/null)
        new_issues=$((new_issues + new_issue_count))
      else
        echo "  No open issues"
      fi
    elif [[ -n "$issues_json" ]]; then
      # jq not available or response not a JSON array — fallback to grep
      local issue_count=$(echo "$issues_json" | grep -oP '"number":' | wc -l)
      if [[ $issue_count -gt 30 ]]; then
        echo "  ⚠️  issue_count=$issue_count exceeds per_page=30, may be inaccurate (jq unavailable)"
      fi
      echo "  Repo: $github_repo, open issues: ~$issue_count (grep fallback, may be imprecise)"
    else
      echo "  (API request failed or no access to $github_repo)"
    fi
  else
    echo "  (no git remote or GITHUB_TOKEN, skipped)"
  fi

  # ── Sensor 6: GitHub PR Reviews ──
  echo "--- Sensor 6: GitHub PR Reviews ---"
  if [[ -n "$github_repo" ]] && [[ -n "${GITHUB_TOKEN:-}" ]]; then
    local prs_json
    prs_json=$(curl -sf -H "Authorization: Bearer $GITHUB_TOKEN" \
      "https://api.github.com/repos/$github_repo/pulls?state=open&per_page=10" 2>/dev/null || true)
    if [[ -n "$prs_json" ]] && echo "$prs_json" | jq -e '. | type == "array"' >/dev/null 2>&1; then
      local pr_count
      pr_count=$(echo "$prs_json" | jq '. | length' 2>/dev/null || echo "0")
      if [[ $pr_count -gt 10 ]]; then
        echo "  ⚠️  pr_count=$pr_count exceeds per_page=10, capping"
        pr_count=10
      fi
      echo "  Open PRs: $pr_count"
      local pr_issues=0
      if [[ $pr_count -gt 0 ]]; then
        # Parse each PR: number|title
        while IFS='|' read -r pr_number pr_title; do
          [[ -z "$pr_number" ]] && continue
          # Check PR review status via reviews API
          local reviews_json
          reviews_json=$(curl -sf -H "Authorization: Bearer $GITHUB_TOKEN" \
            "https://api.github.com/repos/$github_repo/pulls/$pr_number/reviews" 2>/dev/null || true)
          if echo "$reviews_json" | jq -e '.[] | select(.state == "CHANGES_REQUESTED")' >/dev/null 2>&1; then
            local existing=$(grep -l "github:pr:${pr_number}" "$TASKS_DIR"/task-*.md 2>/dev/null || true)
            if [[ -z "$existing" ]]; then
              echo "    NEW: PR #${pr_number} has changes requested — ${pr_title:-no title}"
              if $auto_create; then
                cmd_create --title "修复 PR #${pr_number} review 反馈: ${pr_title:-no title}" \
                  --service "${svc:-all}" --priority P0 --type bug --source github \
                  --detail "github:pr:${pr_number} https://github.com/$github_repo/pull/${pr_number}"
              fi
              pr_issues=$((pr_issues + 1))
            fi
          fi
        done < <(echo "$prs_json" | jq -r '.[] | "\(.number)|\(.title // "no title")"' 2>/dev/null)
        new_issues=$((new_issues + pr_issues))
      else
        echo "  No open PRs"
      fi
    elif [[ -n "$prs_json" ]]; then
      # jq not available or response not a JSON array — fallback
      local pr_count=$(echo "$prs_json" | grep -oP '"number":' | wc -l)
      echo "  Open PRs: ~$pr_count (grep fallback, may be imprecise)"
      local pr_issues=0
      if [[ $pr_count -gt 0 ]]; then
        while IFS= read -r pr_number; do
          [[ -z "$pr_number" ]] && continue
          local reviews_json
          reviews_json=$(curl -sf -H "Authorization: Bearer $GITHUB_TOKEN" \
            "https://api.github.com/repos/$github_repo/pulls/$pr_number/reviews" 2>/dev/null || true)
          if echo "$reviews_json" | grep -q '"state":\s*"CHANGES_REQUESTED"'; then
            local pr_title=$(echo "$prs_json" | grep -oP "\"title\":\s*\"[^\"]*\"" | head -1 | cut -d'"' -f4)
            local existing=$(grep -l "github:pr:${pr_number}" "$TASKS_DIR"/task-*.md 2>/dev/null || true)
            if [[ -z "$existing" ]]; then
              echo "    NEW: PR #${pr_number} has changes requested — $pr_title"
              if $auto_create; then
                cmd_create --title "修复 PR #${pr_number} review 反馈: $pr_title" \
                  --service "${svc:-all}" --priority P0 --type bug --source github \
                  --detail "github:pr:${pr_number} https://github.com/$github_repo/pull/${pr_number}"
              fi
              pr_issues=$((pr_issues + 1))
            fi
          fi
        done < <(echo "$prs_json" | grep -oP '"number":\s*\d+' | grep -oP '\d+')
        new_issues=$((new_issues + pr_issues))
      fi
    else
      echo "  (no PRs or API access limited)"
    fi
  else
    echo "  (no git remote or GITHUB_TOKEN, skipped)"
  fi

  echo ""
  echo "=== Scan complete: $new_issues new issue(s) found ==="
}

# ─── index (重新生成 BACKLOG.md) ──────────────────────────────────────

cmd_index() {
  local quiet=false
  [[ "${1:-}" == "--quiet" ]] && quiet=true

  local p0="" p1="" p2="" p3="" in_prog="" blocked=""
  local count_p0=0 count_p1=0 count_p2=0 count_p3=0 count_ip=0 count_bl=0
  local -A svc_count

  for f in "$TASKS_DIR"/task-*.md; do
    [[ -f "$f" ]] || continue
    local id=$(get_field "$f" "id")
    local prio=$(get_field "$f" "priority")
    local stat=$(get_field "$f" "status")
    local svc=$(get_field "$f" "service")
    local src=$(get_field "$f" "source")
    local title=$(get_field "$f" "title")

    local line="- [$title](${id}.md) — $svc, $prio, $stat, $src"

    case "$stat" in
      in_progress) in_prog+="$line"$'\n'; count_ip=$((count_ip + 1)) ;;
      blocked)     blocked+="$line"$'\n'; count_bl=$((count_bl + 1)) ;;
      *)
        case "$prio" in
          P0) p0+="$line"$'\n'; count_p0=$((count_p0 + 1)) ;;
          P1) p1+="$line"$'\n'; count_p1=$((count_p1 + 1)) ;;
          P2) p2+="$line"$'\n'; count_p2=$((count_p2 + 1)) ;;
          P3) p3+="$line"$'\n'; count_p3=$((count_p3 + 1)) ;;
        esac
        ;;
    esac

    # Per-service count (guard: some legacy task files may lack a service field)
    local svc_key2="${svc:-unknown}"
    local cur=${svc_count[$svc_key2]:-0}
    svc_count[$svc_key2]=$((cur + 1))
  done

  cat > "$BACKLOG" <<BACKLOGEOF
# 待办索引

> Loop 启动时读取本文件，按优先级调度任务。
> 格式：\`- [标题](文件.md) — 服务, 优先级, 状态, 来源\`
>
> 来源类型：\`human\`=人安排的战略任务 | \`qa\`=QA 传感器检测 | \`review\`=Review 发现 | \`sensor\`=自动传感器
>
> 最后更新：$(today)

## P0 — 立即处理（阻塞性问题）

${p0:-（暂无）}

## P1 — 本周

${p1:-（暂无）}

## P2 — 本月

${p2:-（暂无）}

## P3 — 以后

${p3:-（暂无）}

## 进行中

${in_prog:-（暂无）}

## 已阻塞

${blocked:-（暂无）}

---

## 统计

| 优先级 | 数量 |
|:------:|:----:|
| P0 | $count_p0 |
| P1 | $count_p1 |
| P2 | $count_p2 |
| P3 | $count_p3 |
| 进行中 | $count_ip |
| 已阻塞 | $count_bl |
| **合计** | **$((count_p0 + count_p1 + count_p2 + count_p3 + count_ip + count_bl))** |

| 服务 | 数量 |
|------|:----:|
BACKLOGEOF

  for svc in "${!svc_count[@]}"; do
    echo "| $svc | ${svc_count[$svc]} |" >> "$BACKLOG"
  done
  echo "" >> "$BACKLOG"

  $quiet || echo "Index regenerated: $BACKLOG"
}

# ─── stats ──────────────────────────────────────────────────────────────

cmd_stats() {
  echo "=== Task Backlog Statistics ==="
  echo ""

  local total=0 p0=0 p1=0 p2=0 p3=0 open=0 ip=0 closed=0 block=0 review=0
  declare -A svc_count src_count

  for f in "$TASKS_DIR"/task-*.md; do
    [[ -f "$f" ]] || continue
    total=$((total + 1))

    local prio=$(get_field "$f" "priority")
    local stat=$(get_field "$f" "status")
    local svc=$(get_field "$f" "service")
    local src=$(get_field "$f" "source")

    case "$prio" in P0) p0=$((p0+1)) ;; P1) p1=$((p1+1)) ;; P2) p2=$((p2+1)) ;; P3) p3=$((p3+1)) ;; esac
    case "$stat" in
      open) open=$((open+1)) ;;
      in_progress) ip=$((ip+1)) ;;
      closed) closed=$((closed+1)) ;;
      review) review=$((review+1)) ;;
      blocked) block=$((block+1)) ;;
    esac

    local cs=${svc_count[${svc:-unknown}]:-0}; svc_count[${svc:-unknown}]=$((cs+1))
    local ss=${src_count[$src]:-0}; src_count[$src]=$((ss+1))
  done

  echo "总计: $total tasks"
  echo ""
  echo "按优先级:"
  echo "  P0: $p0  P1: $p1  P2: $p2  P3: $p3"
  echo ""
  echo "按状态:"
  echo "  open: $open  in_progress: $ip  review: $review  blocked: $block  closed: $closed"
  echo ""
  echo "按服务:"
  for svc in "${!svc_count[@]}"; do
    printf "  %-30s %d\n" "$svc" "${svc_count[$svc]}"
  done
  echo ""
  echo "按来源:"
  for src in "${!src_count[@]}"; do
    printf "  %-10s %d\n" "$src" "${src_count[$src]}"
  done
}

# ─── metrics (管线健康指标) ─────────────────────────────────────────

cmd_metrics() {
  local output_json=false
  local days=7
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --json) output_json=true; shift ;;
      --days) days="$2"; shift 2 ;;
      *) shift ;;
    esac
  done

  local RUNS_DIR="$PROJECT_ROOT/.harness/loop-runs"
  local cutoff_date=$(date -d "$days days ago" +%Y%m%d%H%M%S 2>/dev/null || date -v-${days}d +%Y%m%d%H%M%S 2>/dev/null || echo "0")

  # Collect recent runs
  local run_files=()
  for f in "$RUNS_DIR"/run-*.md; do
    [[ -f "$f" ]] || continue
    local run_id=$(basename "$f" .md | sed 's/^run-//' | tr -d '-')
    [[ "$run_id" -ge "$cutoff_date" ]] || continue
    run_files+=("$f")
  done

  local total_runs=${#run_files[@]}
  local resolved_tasks=0 failed_runs=0
  local qa_fail_sum=0 qa_fail_first="?" qa_fail_last="?"
  local review_warn_sum=0 review_warn_first="?" review_warn_last="?"
  local new_issues_sum=0 dispatch_sum=0
  local first_ts="" last_ts=""

  for f in "${run_files[@]}"; do
    local ts=$(grep -oP '\*\*时间\*\*: \K.*' "$f" | head -1)
    [[ -n "$ts" ]] && { [[ -z "$first_ts" ]] && first_ts="$ts"; last_ts="$ts"; }

    local qa_fail=$(grep -oP 'Sensor 1.*?\|\s*\K\d+(?= FAIL)' "$f" | head -1 || echo "0")
    qa_fail_sum=$((qa_fail_sum + qa_fail))
    [[ "$qa_fail_first" == "?" ]] && qa_fail_first="$qa_fail"
    qa_fail_last="$qa_fail"

    local rw=$(grep -oP 'Sensor 3.*?\|\s*\K\d+(?= unfixed)' "$f" | head -1 || echo "0")
    review_warn_sum=$((review_warn_sum + rw))
    [[ "$review_warn_first" == "?" ]] && review_warn_first="$rw"
    review_warn_last="$rw"

    local exit_code=$(grep -oP '\- 退出码: \K\d+' "$f" | head -1 || echo "0")
    [[ "$exit_code" -eq 0 ]] && resolved_tasks=$((resolved_tasks + 1)) || failed_runs=$((failed_runs + 1))

    local ni=$(grep -oP '\- 新发现任务: \K\d+' "$f" | head -1 || echo "0")
    new_issues_sum=$((new_issues_sum + ni))
    local dp=$(grep -oP '\- 自动派发: \K\d+' "$f" | head -1 || echo "0")
    dispatch_sum=$((dispatch_sum + dp))
  done

  # Task health + closure metrics (闭环率/生命周期/backlog 年龄)
  local total_tasks=0 zombie_tasks=0 stale_tasks=0 blocked_tasks=0
  local backlog_gt7=0 backlog_gt30=0
  local lifecycle_sum=0 lifecycle_n=0
  declare -A src_total src_closed
  local stale_cutoff=$(date -d "7 days ago" +%Y-%m-%d 2>/dev/null || date -v-7d +%Y-%m-%d 2>/dev/null || echo "")
  local cutoff7=$(date -d "7 days ago" +%Y-%m-%d 2>/dev/null || date -v-7d +%Y-%m-%d 2>/dev/null || echo "")
  local cutoff30=$(date -d "30 days ago" +%Y-%m-%d 2>/dev/null || date -v-30d +%Y-%m-%d 2>/dev/null || echo "")
  for tf in "$TASKS_DIR"/task-*.md; do
    [[ -f "$tf" ]] || continue
    local stat=$(get_field "$tf" "status")
    local created=$(get_field "$tf" "created")
    local src=$(get_field "$tf" "source")
    local dc=$(grep -oP '(?<=^dispatch_count: ).*' "$tf" | head -1 || echo "0")
    total_tasks=$((total_tasks + 1))
    [[ "$stat" == "blocked" ]] && blocked_tasks=$((blocked_tasks + 1))
    [[ "$dc" -ge 3 && "$stat" != "closed" && "$stat" != "blocked" ]] && zombie_tasks=$((zombie_tasks + 1))
    [[ -n "$stale_cutoff" && "$created" < "$stale_cutoff" && "$stat" != "closed" ]] && stale_tasks=$((stale_tasks + 1))
    # 闭环率：按来源统计 closed/total（qa/review/sensor/github 自动来源尤其重要）
    src_total[$src]=$((${src_total[$src]:-0} + 1))
    if [[ "$stat" == "closed" ]]; then
      src_closed[$src]=$((${src_closed[$src]:-0} + 1))
      # 生命周期天数（created → completed）
      local c_ts=$(date -d "$created" +%s 2>/dev/null || echo "0")
      local d_ts=$(date -d "$(get_field "$tf" "completed" | tr -d '"')" +%s 2>/dev/null || echo "0")
      [[ $c_ts -gt 0 && $d_ts -gt 0 ]] && { lifecycle_sum=$((lifecycle_sum + (d_ts - c_ts) / 86400)); lifecycle_n=$((lifecycle_n + 1)); }
    else
      [[ -n "$cutoff7" && "$created" < "$cutoff7" ]] && backlog_gt7=$((backlog_gt7 + 1))
      [[ -n "$cutoff30" && "$created" < "$cutoff30" ]] && backlog_gt30=$((backlog_gt30 + 1))
    fi
  done
  local avg_lifecycle=0
  [[ $lifecycle_n -gt 0 ]] && avg_lifecycle=$(( lifecycle_sum / lifecycle_n ))

  # Sensor trend arrows
  _trend() {
    local first="$1" last="$2"
    [[ "$first" == "?" || "$last" == "?" ]] && { echo "—"; return; }
    if [[ "$last" -lt "$first" ]]; then echo "↓ improving"; elif [[ "$last" -gt "$first" ]]; then echo "↑ worsening"; else echo "→ stable"; fi
  }

  if $output_json; then
    # 预构建 closureBySource JSON（函数内构建，避免子 shell 丢失关联数组类型）
    local closure_json=""
    for s in qa review sensor github human; do
      local st=${src_total[$s]:-0} sc=${src_closed[$s]:-0}
      local sr=$(( st > 0 ? sc * 100 / st : 0 ))
      closure_json+="\"$s\": {\"total\": $st, \"closed\": $sc, \"rate\": $sr}"
      [[ "$s" != "human" ]] && closure_json+=", "
    done
    cat <<JSONEOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "period": "${days}d",
  "firstRun": "${first_ts:-N/A}",
  "lastRun": "${last_ts:-N/A}",
  "totalRuns": $total_runs,
  "resolvedRuns": $resolved_tasks,
  "failedRuns": $failed_runs,
  "avgNewIssuesPerRun": $(( total_runs > 0 ? new_issues_sum / total_runs : 0 )),
  "avgDispatchPerRun": $(( total_runs > 0 ? dispatch_sum / total_runs : 0 )),
  "sensors": {
    "qaChecks": { "first": "$qa_fail_first", "last": "$qa_fail_last", "trend": "$(_trend "$qa_fail_first" "$qa_fail_last")" },
    "reviewWarnings": { "first": "$review_warn_first", "last": "$review_warn_last", "trend": "$(_trend "$review_warn_first" "$review_warn_last")" }
  },
  "taskHealth": {
    "total": $total_tasks,
    "zombieTasks": $zombie_tasks,
    "staleTasks": $stale_tasks,
    "blockedTasks": $blocked_tasks,
    "avgLifecycleDays": $avg_lifecycle,
    "backlogAge": { "gt7d": $backlog_gt7, "gt30d": $backlog_gt30 }
  },
  "closureBySource": {
    ${closure_json}
  }
}
JSONEOF
  else
    echo "╔══════════════════════════════════════╗"
    echo "║   Pipeline Health Metrics (${days}d)"
    echo "╚══════════════════════════════════════╝"
    echo ""
    echo "━━━ Loop Activity ━━━"
    echo "  时间范围: ${first_ts:-N/A} → ${last_ts:-N/A}"
    echo "  总运行: $total_runs   成功: $resolved_tasks   失败: $failed_runs"
    echo "  平均新问题/次: $(( total_runs > 0 ? new_issues_sum / total_runs : 0 ))   平均派发/次: $(( total_runs > 0 ? dispatch_sum / total_runs : 0 ))"
    echo ""
    echo "━━━ Sensor Trends (最早 → 最新) ━━━"
    echo "  Sensor 1 (QA FAILs):  ${qa_fail_first} → ${qa_fail_last}  $(_trend "$qa_fail_first" "$qa_fail_last")"
    echo "  Sensor 3 (Review WARNINGs): ${review_warn_first} → ${review_warn_last}  $(_trend "$review_warn_first" "$review_warn_last")"
    echo ""
    echo "━━━ Task Health ━━━"
    echo "  活跃任务: $total_tasks   僵尸(≥3次派发): $zombie_tasks   >7天未关: $stale_tasks   已阻塞: $blocked_tasks"
    echo ""
    echo "━━━ 闭环率与 Backlog 健康 ━━━"
    echo "  平均任务生命周期: ${avg_lifecycle} 天   backlog 年龄: >7天 ${backlog_gt7} 个, >30天 ${backlog_gt30} 个"
    local closure_line=""
    for s in qa review sensor github human; do
      local st=${src_total[$s]:-0} sc=${src_closed[$s]:-0}
      local sr=$(( st > 0 ? sc * 100 / st : 0 ))
      closure_line+="${s}=${sc}/${st}(${sr}%)  "
    done
    echo "  按来源闭环率: ${closure_line}"
    if [[ ${src_total[qa]:-0} -gt 0 ]] && [[ ${src_closed[qa]:-0} -lt ${src_total[qa]} ]]; then
      echo "  ⚠️  QA 传感器任务闭环率 < 100%（${src_closed[qa]:-0}/${src_total[qa]:-0}）— 传感器发现问题未被解决"
    fi
    echo ""
    if [[ $zombie_tasks -gt 0 ]]; then
      echo "⚠️  $zombie_tasks 个僵尸任务需要人工介入:"
      for tf in "$TASKS_DIR"/task-*.md; do
        local dc=$(grep -oP '(?<=^dispatch_count: ).*' "$tf" | head -1 || echo "0")
        [[ "$dc" -ge 3 ]] || continue
        local id=$(get_field "$tf" "id")
        local title=$(get_field "$tf" "title")
        local stat=$(get_field "$tf" "status")
        [[ "$stat" == "closed" || "$stat" == "blocked" ]] && continue
        echo "    - $id ($title) — 已派发 ${dc} 次, 状态: $stat"
      done
      echo ""
    fi
    echo "💡 说 'harness-tasks.sh metrics --json' 获取机器可读格式"
  fi
}

# ─── memory-health (记忆系统体检) ─────────────────────────────────────

cmd_memory_health() {
  local MEMORY_DIR="$PROJECT_ROOT/.harness/knowledge/memory"
  local SERVICES_DIR="$PROJECT_ROOT/services"

  echo "=== Memory System Health ==="
  echo ""

  # Count standard vs non-standard memories
  local standard=0 nonstandard=0 standard_triggers=0
  local -a nonstandard_slugs
  local -A trigger_map  # trigger → list of slugs

  for mf in "$MEMORY_DIR"/*.md; do
    [[ -f "$mf" ]] || continue
    [[ "$(basename "$mf")" == "MEMORY.md" || "$(basename "$mf")" == "MAINTENANCE.md" ]] && continue
    local slug=$(basename "$mf" .md)

    if grep -q '^triggers:' "$mf" 2>/dev/null; then
      standard=$((standard + 1))
      local trig=$(grep -oP '(?<=^triggers: ").*' "$mf" | head -1 | sed 's/"$//' || echo "")
      [[ -n "$trig" ]] && standard_triggers=$((standard_triggers + 1))
    else
      nonstandard=$((nonstandard + 1))
      nonstandard_slugs+=("$slug")
    fi
  done

  echo "━━━ Schema Compliance ━━━"
  echo "  ✅ $standard/$((standard + nonstandard)) memories use standard frontmatter"
  if [[ $nonstandard -gt 0 ]]; then
    echo "  ⚠️  $nonstandard non-standard (invisible to trigger matching):"
    for s in "${nonstandard_slugs[@]}"; do
      echo "     - $s"
    done
  fi
  echo ""

  # Trigger overlap detection
  echo "━━━ Trigger Overlaps ━━━"
  local overlap_found=false
  declare -A seen_trig
  for mf in "$MEMORY_DIR"/*.md; do
    [[ -f "$mf" ]] || continue
    local slug=$(basename "$mf" .md)
    local trig=$(grep -oP '(?<=^triggers: ").*' "$mf" | head -1 | sed 's/"$//' || echo "")
    [[ -z "$trig" ]] && continue
    for word in $trig; do
      word=$(echo "$word" | tr -d ',')
      [[ ${#word} -lt 3 ]] && continue  # skip short/noise words
      local existing="${seen_trig[$word]:-}"
      if [[ -n "$existing" ]]; then
        if ! echo "$existing" | grep -q "$slug"; then
          echo "  ⚠️  \"$word\" → $existing + $slug"
          overlap_found=true
        fi
        seen_trig[$word]="$existing, $slug"
      else
        seen_trig[$word]="$slug"
      fi
    done
  done
  $overlap_found || echo "  ✅ No significant overlaps detected"
  echo ""

  # Lifecycle: count by status
  echo "━━━ Lifecycle ━━━"
  local active=0 draft=0 superseded=0 other=0
  for mf in "$MEMORY_DIR"/*.md; do
    [[ -f "$mf" ]] || continue
    [[ "$(basename "$mf")" == "MEMORY.md" || "$(basename "$mf")" == "MAINTENANCE.md" ]] && continue
    local st=$(grep -oP '(?<=^status: ).*' "$mf" | head -1 || echo "no-status")
    case "$st" in
      active) active=$((active + 1)) ;;
      draft) draft=$((draft + 1)) ;;
      superseded) superseded=$((superseded + 1)) ;;
      *) other=$((other + 1)) ;;
    esac
  done
  echo "  Active: $active | Draft: $draft | Superseded: $superseded | No-status: $other"
  if [[ $superseded -eq 0 ]] && [[ $((active + other)) -gt 10 ]]; then
    echo "  ⚠️  0 memories retired — memory base growing monotonically"
  fi
  echo ""

  # Usage tracking
  echo "━━━ Usage Tracking ━━━"
  local have_counter=0 ever_applied=0
  for mf in "$MEMORY_DIR"/*.md; do
    [[ -f "$mf" ]] || continue
    if grep -q '^apply_count:' "$mf" 2>/dev/null; then
      have_counter=$((have_counter + 1))
      local ac=$(grep -oP '(?<=^apply_count: ).*' "$mf" | head -1 || echo "0")
      [[ "$ac" != "0" && "$ac" != "null" && -n "$ac" ]] && ever_applied=$((ever_applied + 1))
    fi
  done
  echo "  $have_counter/$((standard + nonstandard)) have apply_count fields — $ever_applied ever applied"
  if [[ $ever_applied -eq 0 ]] && [[ $have_counter -gt 0 ]]; then
    echo "  💡 Generator memory report is not parsed to update counters"
  fi
  echo ""

  # Service-local memories
  echo "━━━ Service Scope ━━━"
  local svc_mem_count=0
  for svc_dir in "$SERVICES_DIR"/*/; do
    local svc_mem="$svc_dir.harness/knowledge/memory/"
    if [[ -d "$svc_mem" ]]; then
      local count=$(ls "$svc_mem"/*.md 2>/dev/null | wc -l)
      svc_mem_count=$((svc_mem_count + count))
    fi
  done
  if [[ $svc_mem_count -eq 0 ]]; then
    echo "  ⚠️  0 service-local memories (services/*/.harness/knowledge/memory/ is empty)"
  else
    echo "  $svc_mem_count service-local memories"
  fi
  echo ""
  echo "💡 运行 'harness-tasks.sh memory-health' 随时体检记忆系统"
}

# ─── retrospective (月度复盘报告) ─────────────────────────────────────

cmd_retrospective() {
  local days=30
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --days) days="$2"; shift 2 ;;
      *) shift ;;
    esac
  done

  local today=$(date +%Y-%m-%d)
  local cutoff=$(date -d "$days days ago" +%Y-%m-%d 2>/dev/null || date -v-${days}d +%Y-%m-%d 2>/dev/null || echo "")

  echo "╔══════════════════════════════════════════════════╗"
  echo "║   Pipeline Retrospective (${days}d: ${cutoff:-?} → ${today})"
  echo "╚══════════════════════════════════════════════════╝"
  echo ""

  # ── 1. Task lifecycle analysis ──
  local total=0 closed=0 blocked=0 open=0 in_progress=0
  local total_dispatch=0 max_dispatch=0
  local -A type_count type_closed
  local -A source_count source_closed

  for tf in "$TASKS_DIR"/task-*.md; do
    [[ -f "$tf" ]] || continue
    local created=$(get_field "$tf" "created")
    [[ -n "$cutoff" && "$created" < "$cutoff" ]] && continue

    total=$((total + 1))
    local stat=$(get_field "$tf" "status")
    local ttype=$(get_field "$tf" "type")
    local src=$(get_field "$tf" "source")
    local dc=$(grep -oP '(?<=^dispatch_count: ).*' "$tf" | head -1 || echo "0")
    dc=$(( dc + 0 ))

    case "$stat" in
      closed) closed=$((closed + 1)); type_closed[$ttype]=$((${type_closed[$ttype]:-0} + 1)); source_closed[$src]=$((${source_closed[$src]:-0} + 1)) ;;
      blocked) blocked=$((blocked + 1)) ;;
      in_progress) in_progress=$((in_progress + 1)) ;;
      open) open=$((open + 1)) ;;
    esac
    type_count[$ttype]=$((${type_count[$ttype]:-0} + 1))
    source_count[$src]=$((${source_count[$src]:-0} + 1))
    total_dispatch=$((total_dispatch + dc))
    [[ $dc -gt $max_dispatch ]] && max_dispatch=$dc
  done

  echo "━━━ 1. 任务生命周期 ━━━"
  echo "  总任务: $total   已关闭: $closed   进行中: $in_progress   阻塞: $blocked   待处理: $open"
  if [[ $total -gt 0 ]]; then
    echo "  关闭率: $(( closed * 100 / total ))%   平均派发次数: $(( total_dispatch / (total > 0 ? total : 1) ))   最大派发: $max_dispatch"
  fi
  echo ""

  # ── 2. Type effectiveness ──
  echo "━━━ 2. 类型效率 ━━━"
  printf "  %-10s %6s %6s %8s\n" "类型" "总数" "已关闭" "关闭率"
  for t in feature bug debt chore; do
    local tc=${type_count[$t]:-0}
    local tcl=${type_closed[$t]:-0}
    local rate=$(( tc > 0 ? tcl * 100 / tc : 0 ))
    printf "  %-10s %6s %6s %7s%%\n" "$t" "$tc" "$tcl" "$rate"
  done
  echo ""

  # ── 3. Source effectiveness ──
  echo "━━━ 3. 来源效率 ━━━"
  printf "  %-10s %6s %6s %8s\n" "来源" "总数" "已关闭" "关闭率"
  for s in human qa review sensor github; do
    local sc=${source_count[$s]:-0}
    local scl=${source_closed[$s]:-0}
    local rate=$(( sc > 0 ? scl * 100 / sc : 0 ))
    printf "  %-10s %6s %6s %7s%%\n" "$s" "$sc" "$scl" "$rate"
  done
  echo ""

  # ── 4. Memory reference analysis ──
  echo "━━━ 4. 记忆引用热度 ━━━"
  local -A mem_hits
  local total_refs=0
  for mf in "$PROJECT_ROOT/.harness/knowledge/memory"/*.md; do
    [[ -f "$mf" ]] || continue
    [[ "$(basename "$mf")" == "MEMORY.md" || "$(basename "$mf")" == "MAINTENANCE.md" ]] && continue
    local ac=$(grep -oP '(?<=^apply_count: ).*' "$mf" | head -1 || echo "0")
    ac=$(( ac + 0 ))
    [[ $ac -gt 0 ]] || continue
    local slug=$(basename "$mf" .md)
    mem_hits["$slug"]=$ac
    total_refs=$((total_refs + ac))
  done
  if [[ $total_refs -gt 0 ]]; then
    for slug in "${!mem_hits[@]}"; do
      printf "  %-45s %s\n" "$slug" "${mem_hits[$slug]} 次"
    done | sort -t' ' -k2 -rn | head -10
    local mem_count=0
    for s in "${!mem_hits[@]}"; do mem_count=$((mem_count + 1)); done
    echo "  总计: ${mem_count} 条被引用, ${total_refs} 次引用"
  else
    echo "  💡 尚无记忆被 apply_count 追踪（Pipeline Generator 引用未解析到计数器）"
  fi
  echo ""

  # ── 5. Sensor contribution ──
  echo "━━━ 5. Sensor 贡献 ━━━"
  echo "  （每个 sensor 发现并转化为任务的数量）"
  printf "  %-25s %6s %6s\n" "Sensor" "创建任务" "已关闭"
  local s1_c=0 s1_cl=0 s2_c=0 s2_cl=0 s4_c=0 s4_cl=0 s5_c=0 s5_cl=0 s6_c=0 s6_cl=0
  for tf in "$TASKS_DIR"/task-*.md; do
    [[ -f "$tf" ]] || continue
    local sd=$(get_field "$tf" "source_detail")
    local stat=$(get_field "$tf" "status")
    if echo "$sd" | grep -q "harness-checks.sh check:"; then s1_c=$((s1_c+1)); [[ "$stat" == "closed" ]] && s1_cl=$((s1_cl+1)); fi
    if echo "$sd" | grep -q "api_stubs"; then s2_c=$((s2_c+1)); [[ "$stat" == "closed" ]] && s2_cl=$((s2_cl+1)); fi
    if echo "$sd" | grep -q "graph_freshness"; then s4_c=$((s4_c+1)); [[ "$stat" == "closed" ]] && s4_cl=$((s4_cl+1)); fi
    if echo "$sd" | grep -q "github:issue:"; then s5_c=$((s5_c+1)); [[ "$stat" == "closed" ]] && s5_cl=$((s5_cl+1)); fi
    if echo "$sd" | grep -q "github:pr:"; then s6_c=$((s6_c+1)); [[ "$stat" == "closed" ]] && s6_cl=$((s6_cl+1)); fi
  done
  printf "  %-25s %6s %6s\n" "Sensor 1 (QA checks)" "$s1_c" "$s1_cl"
  printf "  %-25s %6s %6s\n" "Sensor 2 (TODO stubs)" "$s2_c" "$s2_cl"
  printf "  %-25s %6s %6s\n" "Sensor 4 (Graph freshness)" "$s4_c" "$s4_cl"
  printf "  %-25s %6s %6s\n" "Sensor 5 (GitHub Issues)" "$s5_c" "$s5_cl"
  printf "  %-25s %6s %6s\n" "Sensor 6 (PR Reviews)" "$s6_c" "$s6_cl"
  echo ""

  # ── 6. Top bottlenecks ──
  echo "━━━ 6. 瓶颈 Top 5 ━━━"
  echo "  （派发次数最多的未完成任务）"
  local bottleneck_count=0
  for tf in "$TASKS_DIR"/task-*.md; do
    [[ -f "$tf" ]] || continue
    local stat=$(get_field "$tf" "status")
    [[ "$stat" == "closed" ]] && continue
    local dc=$(grep -oP '(?<=^dispatch_count: ).*' "$tf" | head -1 || echo "0")
    dc=$(( dc + 0 ))
    [[ $dc -le 1 ]] && continue
    local id=$(get_field "$tf" "id")
    local title=$(get_field "$tf" "title")
    local ttype=$(get_field "$tf" "type")
    printf "  %s (%s) — %s 次派发, 状态: %s\n" "$id" "$ttype" "$dc" "$stat"
    echo "    $title"
    bottleneck_count=$((bottleneck_count + 1))
    [[ $bottleneck_count -ge 5 ]] && break
  done
  [[ $bottleneck_count -eq 0 ]] && echo "  ✅ 无重复派发任务"
  echo ""

  # ── 7. Recommendations ──
  echo "━━━ 7. 建议 ━━━"
  if [[ $max_dispatch -ge 5 ]]; then
    echo "  ⚠️  存在被派发 $max_dispatch 次的任务 → 检查收敛检测是否生效"
  fi
  if [[ $total -gt 0 ]] && [[ $(( closed * 100 / total )) -lt 50 ]]; then
    echo "  ⚠️  任务关闭率 < 50% → 管线产出低于流入，backlog 在增长"
  fi
  local debt_count=${type_count[debt]:-0} debt_cl=${type_closed[debt]:-0}
  if [[ $debt_count -gt $((total * 60 / 100)) ]]; then
    echo "  💡 debt 类任务占比 > 60% → 考虑将高频 debt 模式升级为 harness-checks 检查项或 memory"
  fi
  if [[ $total_refs -eq 0 ]]; then
    echo "  💡 记忆引用计数未激活 → Generator 的 [[memory-slug]] 引用未被解析到计数器"
  fi
  if [[ $total -gt 0 ]] && [[ $(( closed * 100 / total )) -ge 80 ]]; then
    echo "  ✅ 管线运行良好，关闭率 ≥ 80%"
  fi
  echo ""
  echo "💡 每月运行一次 'harness-tasks.sh retrospective' 跟踪管线健康趋势"
}

# ─── ingest-memories: Review → Memory 反馈闭环 ─────────────────────────
# 从 Pipeline Review 阶段收集的 memorySuggestions（JSON 数组），
# 自动检查是否已有对应记忆文件，未覆盖的创建 BACKLOG 任务供人工审核。
# 用法：echo '[...]' | harness-tasks.sh ingest-memories --service <name>
cmd_ingest_memories() {
  local service_name=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --service) service_name="$2"; shift 2 ;;
      *) shift ;;
    esac
  done

  # 从 stdin 读取 JSON
  local input
  input=$(cat)

  if [ -z "$input" ] || [ "$input" == "[]" ] || [ "$input" == "null" ]; then
    echo "ℹ️  无 Memory Suggestions，跳过"
    return 0
  fi

  # 解析 JSON 数组
  if ! echo "$input" | jq -e 'type == "array"' > /dev/null 2>&1; then
    echo "⚠️  输入不是有效的 JSON 数组，跳过"
    return 0
  fi

  local count=$(echo "$input" | jq 'length')
  echo "🔍 收到 $count 条 Memory Suggestions (Review → Memory 反馈闭环)"
  echo ""

  local created=0 skipped=0 failed=0

  for i in $(seq 0 $((count - 1))); do
    local suggestion=$(echo "$input" | jq -r ".[$i]")
    local slug=$(echo "$suggestion" | jq -r '.slug // ""' 2>/dev/null || true)
    local title=$(echo "$suggestion" | jq -r '.title // ""' 2>/dev/null || true)
    local category=$(echo "$suggestion" | jq -r '.category // "guideline"' 2>/dev/null || true)
    local severity=$(echo "$suggestion" | jq -r '.severity // "should-follow"' 2>/dev/null || true)
    local triggers=$(echo "$suggestion" | jq -r '.triggers // ""' 2>/dev/null || true)
    local content=$(echo "$suggestion" | jq -r '.content // ""' 2>/dev/null || true)

    [ -z "$slug" ] && failed=$((failed + 1)) && continue

    # 检查记忆是否已存在
    local existing=""
    existing=$(find "$PROJECT_ROOT/.harness/knowledge/memory" -name "${slug}.md" -type f 2>/dev/null | head -1) || true
    if [ -n "$existing" ]; then
      echo "  ⏭️  $slug — 记忆已存在，跳过"
      skipped=$((skipped + 1))
      continue
    fi

    # 检查是否已有同名 BACKLOG 任务
    local dup_task=""
    dup_task=$(grep -rl "slug:.*${slug}" "$TASKS_DIR"/task-*.md 2>/dev/null | head -1) || true
    if [ -n "$dup_task" ]; then
      echo "  ⏭️  $slug — BACKLOG 任务已存在"
      skipped=$((skipped + 1))
      continue
    fi

    # 创建 BACKLOG 任务（P1, triage=pending — 需要人工审核后才能 dispatch）
    local task_title="[Memory] ${title:-$slug}"
    local task_detail="## Memory Suggestion (Review → Memory 反馈闭环)

**slug**: \`${slug}\`
**category**: ${category}
**severity**: ${severity}
**triggers**: ${triggers}
**来源**: Pipeline Review 阶段自动检测

### 建议内容

${content}

---
> 🤖 此任务由 Pipeline Review 自动生成。Dispatch 前请人工审核建议内容是否准确、是否与已有记忆重复。
"

    # 直接创建任务文件（避免 cmd_create 内部 shift 参数冲突）
    local id=$(next_id)
    local file="$TASKS_DIR/${id}.md"
    local task_created=$(today)

    cat > "$file" <<TASKEOF
---
id: ${id}
title: "${task_title}"
service: ${service_name:-all}
type: memory
priority: P1
status: open
source: review
source_detail: "Pipeline Review → Memory 反馈闭环"
created: ${task_created}
blocks: []
blocked_by: []
assigned_run: ""
completed: ""
outcome: ""
triage: "pending"
dispatch_count: 0
slug: "${slug}"
---

# ${task_title}

${task_detail}

## 执行记录

| 日期 | 事件 | 详情 |
|------|------|------|
| ${task_created} | 创建 | Pipeline Review 自动生成 → 人工审核后 dispatch |

## 关联

- 来源: Review Memory Suggestion
- slug: \`${slug}\`
TASKEOF

    echo "  ✅ $slug → $id (P1, triage=pending)"
    created=$((created + 1))
    # 静默更新索引
    cmd_index --quiet 2>/dev/null || true
  done

  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Memory Suggestions 处理完成: $created 新建, $skipped 跳过, $failed 失败"
}

# ─── Main ──────────────────────────────────────────────────────────────

case "${1:-}" in
  list)           shift; cmd_list "$@" ;;
  create)         shift; cmd_create "$@" ;;
  status)         shift; cmd_status "$@" ;;
  scan)           shift; cmd_scan "$@" ;;
  index)          shift; cmd_index "$@" ;;
  stats)          shift; cmd_stats "$@" ;;
  metrics)        shift; cmd_metrics "$@" ;;
  memory-health)  shift; cmd_memory_health "$@" ;;
  retrospective)  shift; cmd_retrospective "$@" ;;
  ingest-memories) shift; cmd_ingest_memories "$@" ;;
  -h|--help|help)
    sed -n '3,17p' "$0" | sed 's/^# //'
    ;;
  *)
    echo "Usage: harness-tasks.sh <command> [options]"
    echo "Commands: list | create | status | scan | index | stats | metrics | memory-health"
    echo "Try 'harness-tasks.sh help' for details."
    exit 1
    ;;
esac
