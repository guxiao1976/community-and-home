#!/usr/bin/env bash
#
# harness-checks-frontend.sh — Mechanized QA checks for frontend services (Vue/React/TS).
#
# Usage:
#   .harness/skills/qa/scripts/harness-checks-frontend.sh --service <name> [--json]
#
# Options:
#   --service <name>   Scope checks to a single frontend app under web/
#   --json             Output results as JSON (default: human-readable text)
#   -h, --help         Show help
#
# Exit code: 0 if all checks pass, non-zero if any check fails.
#
# Checks:
#   1. type-check        — vue-tsc --noEmit / tsc --noEmit
#   2. unit-test         — vitest run / npm test (with 0/0 false-pass detection)
#   3. build             — vite build / npm run build
#   4. hardcoded-secrets — no API keys, tokens, or passwords in source code
#   5. debug-artifacts   — no console.log/debugger in non-test code
#   6. type-safety       — no `as any` escape hatches in production code

set -eu
# pipefail intentionally disabled — grep returning 1 (no match) in pipelines
# would cause premature script exit.

# ─── Config ───────────────────────────────────────────────────────────

PROJECT_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
WEB_DIR="$PROJECT_ROOT/web"
SERVICE_NAME=""
OUTPUT_JSON=false
EXIT_CODE=0

declare -a RESULTS

# Valid frontend service names
VALID_SERVICES=("pc" "mobile")

# ─── Helpers ──────────────────────────────────────────────────────────

timestamp() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

log_pass() {
  local check="$1" detail="$2"
  RESULTS+=("{\"check\":\"$check\",\"status\":\"PASS\",\"detail\":\"$detail\"}")
}

log_fail() {
  local check="$1" detail="$2"
  RESULTS+=("{\"check\":\"$check\",\"status\":\"FAIL\",\"detail\":\"$detail\"}")
  EXIT_CODE=1
}

log_warn() {
  local check="$1" detail="$2"
  RESULTS+=("{\"check\":\"$check\",\"status\":\"WARN\",\"detail\":\"$detail\"}")
}

json_escape() {
  local s="${1-}"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  s="${s//$'\t'/\\t}"
  echo -n "$s"
}

# ─── Parse Args ───────────────────────────────────────────────────────

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service)
      SERVICE_NAME="$2"
      shift 2
      ;;
    --json)
      OUTPUT_JSON=true
      shift
      ;;
    -h|--help)
      sed -n '2,23p' "$0" | sed 's/^# //'
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--service <pc|mobile>] [--json]"
      exit 2
      ;;
  esac
done

# Validate and resolve target directory
if [[ -n "$SERVICE_NAME" ]]; then
  TARGET_DIR="$WEB_DIR/$SERVICE_NAME"
  if [[ ! -d "$TARGET_DIR" ]]; then
    echo "Error: frontend service directory not found: $TARGET_DIR"
    echo "Valid services: ${VALID_SERVICES[*]}"
    exit 2
  fi
else
  TARGET_DIR=""
fi

# Detect package manager (npm/pnpm/yarn)
detect_pm() {
  local dir="$1"
  if [[ -f "$dir/pnpm-lock.yaml" ]]; then echo "pnpm"
  elif [[ -f "$dir/yarn.lock" ]]; then echo "yarn"
  else echo "npm"
  fi
}

# ─── Check 1: Type Check ──────────────────────────────────────────────

check_type_check() {
  echo "[1/6] TypeScript type check" >&2

  if [[ -z "$TARGET_DIR" ]]; then
    local all_pass=true
    for svc in "${VALID_SERVICES[@]}"; do
      local d="$WEB_DIR/$svc"
      [[ ! -d "$d" ]] && continue
      if ! _type_check_one "$d" "$svc"; then
        all_pass=false
      fi
    done
    $all_pass || true # result already logged per-service
    return
  fi

  _type_check_one "$TARGET_DIR" "$SERVICE_NAME"
}

_type_check_one() {
  local dir="$1" label="$2"
  local pm
  pm=$(detect_pm "$dir")

  local out rc
  cd "$dir"

  # Try vue-tsc first (Vue projects), fall back to tsc
  { out="$(npx vue-tsc --noEmit 2>&1)"; rc=$?; } || true
  if [[ $rc -ne 0 ]]; then
    # Try tsc as fallback
    local tsc_out
    { tsc_out="$(npx tsc --noEmit 2>&1)"; rc=$?; } || true
    out="$tsc_out"
  fi
  cd "$PROJECT_ROOT"

  if [[ $rc -eq 0 ]]; then
    log_pass "type_check" "$label: type check passed"
  else
    local err_summary
    err_summary=$(echo "$out" | grep -i 'error' | head -3 | tr '\n' '; ')
    err_summary="$(json_escape "$err_summary")"
    log_fail "type_check" "$label: type errors: $err_summary"
    return 1
  fi
}

# ─── Check 2: Unit Tests ──────────────────────────────────────────────

check_unit_test() {
  echo "[2/6] Unit tests (with 0/0 detection)" >&2

  if [[ -z "$TARGET_DIR" ]]; then
    for svc in "${VALID_SERVICES[@]}"; do
      local d="$WEB_DIR/$svc"
      [[ ! -d "$d" ]] && continue
      _unit_test_one "$d" "$svc"
    done
    return
  fi

  _unit_test_one "$TARGET_DIR" "$SERVICE_NAME"
}

_unit_test_one() {
  local dir="$1" label="$2"
  local pm
  pm=$(detect_pm "$dir")

  cd "$dir"

  # Detect test runner
  local test_cmd=""
  if [[ -f "vitest.config.ts" ]] || [[ -f "vitest.config.js" ]]; then
    test_cmd="npx vitest run --reporter=verbose"
  elif grep -q '"jest"' package.json 2>/dev/null; then
    test_cmd="npx jest --verbose"
  elif grep -q '"test"' package.json 2>/dev/null; then
    test_cmd="npm test"
  fi

  if [[ -z "$test_cmd" ]]; then
    log_warn "unit_test" "$label: no test runner detected"
    cd "$PROJECT_ROOT"
    return
  fi

  local out rc
  { out="$($test_cmd 2>&1)"; rc=$?; } || true
  cd "$PROJECT_ROOT"

  # Count test files and tests
  local test_files=0 passed_tests=0
  # vitest 4.x format: "Test Files  N passed (N)"
  test_files=$(echo "$out" | grep -oP 'Test Files\s+\K\d+' | head -1) || test_files=0
  passed_tests=$(echo "$out" | grep -oP 'Tests\s+\K\d+' | head -1) || passed_tests=0
  test_files=$(echo "$test_files" | tr -d '[:space:]')
  passed_tests=$(echo "$passed_tests" | tr -d '[:space:]')
  test_files=${test_files:-0}
  passed_tests=${passed_tests:-0}

  if [[ $rc -ne 0 ]]; then
    local fail_detail
    fail_detail=$(echo "$out" | grep 'FAIL\|×' | head -5 | tr '\n' '; ')
    fail_detail="$(json_escape "$fail_detail")"
    log_fail "unit_test" "$label: tests failed — $fail_detail"
    return 1
  fi

  # 0/0 false-pass detection
  if [[ "$passed_tests" -eq 0 ]] 2>/dev/null; then
    log_warn "unit_test" "$label: 0 tests found — potential 0/0 false-pass"
    return
  fi

  log_pass "unit_test" "$label: $passed_tests tests passed"
}

# ─── Check 3: Build ───────────────────────────────────────────────────

check_build() {
  echo "[3/6] Production build" >&2

  if [[ -z "$TARGET_DIR" ]]; then
    for svc in "${VALID_SERVICES[@]}"; do
      local d="$WEB_DIR/$svc"
      [[ ! -d "$d" ]] && continue
      _build_one "$d" "$svc"
    done
    return
  fi

  _build_one "$TARGET_DIR" "$SERVICE_NAME"
}

_build_one() {
  local dir="$1" label="$2"

  cd "$dir"

  # Detect build command
  local build_cmd=""
  if grep -q '"build"' package.json 2>/dev/null; then
    build_cmd="npm run build"
  else
    build_cmd="npx vite build"
  fi

  local tmp_log
  tmp_log=$(mktemp)
  local rc
  # Use timeout to prevent hanging; redirect to temp file to avoid OOM on large output
  # Run in subshell to avoid set -e pollution
  ( timeout 60 $build_cmd >"$tmp_log" 2>&1 )
  rc=$?
  cd "$PROJECT_ROOT"

  if [[ $rc -eq 0 ]]; then
    rm -f "$tmp_log"
    log_pass "build" "$label: build succeeded"
  elif [[ $rc -eq 124 ]]; then
    rm -f "$tmp_log"
    log_fail "build" "$label: build timed out (120s)"
    return 1
  else
    # Count distinct error types from the log
    local ts_errors vite_errors
    ts_errors=$(grep -c 'error TS' "$tmp_log" 2>/dev/null) || ts_errors=0
    vite_errors=$(grep -c 'Build failed' "$tmp_log" 2>/dev/null) || vite_errors=0
    ts_errors=$(echo "$ts_errors" | tr -d '[:space:]')
    vite_errors=$(echo "$vite_errors" | tr -d '[:space:]')
    ts_errors=${ts_errors:-0}
    vite_errors=${vite_errors:-0}

    local err_summary=""
    if [[ "$ts_errors" -gt 0 ]] 2>/dev/null; then
      err_summary="$ts_errors TS errors"
    fi
    if [[ "$vite_errors" -gt 0 ]] 2>/dev/null; then
      err_summary="${err_summary:+$err_summary, }$vite_errors vite errors"
    fi
    err_summary="${err_summary:-build failed}"

    # Capture first few actual error messages for context
    local first_errors
    first_errors=$(grep 'error TS\|Build failed' "$tmp_log" 2>/dev/null | head -3 | tr '\n' '; ')
    first_errors="$(json_escape "$first_errors")"
    err_summary="$(json_escape "$err_summary")"

    rm -f "$tmp_log"
    log_fail "build" "$label: $err_summary — $first_errors"
    return 1
  fi
}

# ─── Check 4: Hardcoded Secrets ───────────────────────────────────────

check_hardcoded_secrets() {
  echo "[4/6] Hardcoded secrets" >&2

  local search_dir="${TARGET_DIR:-$WEB_DIR}"

  local violations=()
  local patterns=(
    "api_key\s*[:=]\s*['\"][A-Za-z0-9_\-]{8,}['\"]"
    "apikey\s*[:=]\s*['\"][A-Za-z0-9_\-]{8,}['\"]"
    "secret\s*[:=]\s*['\"][A-Za-z0-9_\-]{8,}['\"]"
    "token\s*[:=]\s*['\"][A-Za-z0-9_\-\.]{12,}['\"]"
    "password\s*[:=]\s*['\"][^'\"]{4,}['\"]"
    "sk-[A-Za-z0-9]{20,}"  # OpenAI API key format
    "AKIA[0-9A-Z]{16}"     # AWS access key format
  )

  for pattern in "${patterns[@]}"; do
    while IFS= read -r match; do
      [[ -z "$match" ]] && continue
      local file="${match%%:*}"
      # Skip test files, config files, .env files, node_modules
      [[ "$file" == *".spec."* ]] && continue
      [[ "$file" == *".test."* ]] && continue
      [[ "$file" == *".env"* ]] && continue
      [[ "$file" == *"node_modules"* ]] && continue
      [[ "$file" == *"vite.config"* ]] && continue
      [[ "$file" == *"vitest.config"* ]] && continue
      # Skip if line imports from config/env
      local line_content="${match#*:*:}"
      [[ "$line_content" == *'import.meta.env'* ]] && continue
      [[ "$line_content" == *'process.env'* ]] && continue
      [[ "$line_content" == *'VITE_'* ]] && continue

      local rel="${file#$PROJECT_ROOT/}"
      local line_num
      line_num=$(echo "$match" | cut -d: -f2)
      violations+=("$rel:$line_num:potential secret")
    done < <(grep -rnPE "$pattern" "$search_dir/src" --include='*.ts' --include='*.tsx' --include='*.vue' --include='*.js' 2>/dev/null || true)
  done

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "hardcoded_secrets" "no secrets detected in source"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_fail "hardcoded_secrets" "${#violations[@]} potential secrets: $detail"
  fi
}

# ─── Check 5: Debug Artifacts ─────────────────────────────────────────

check_debug_artifacts() {
  echo "[5/6] Debug artifacts" >&2

  local search_dir="${TARGET_DIR:-$WEB_DIR}"

  local violations=()
  local patterns=(
    "console\.(log|debug|dir|trace)\("
    "debugger[[:space:];]"
  )

  for pattern in "${patterns[@]}"; do
    while IFS= read -r match; do
      [[ -z "$match" ]] && continue
      local file="${match%%:*}"
      # Allow in test files and node_modules
      [[ "$file" == *".spec."* ]] && continue
      [[ "$file" == *".test."* ]] && continue
      [[ "$file" == *"node_modules"* ]] && continue
      # Allow console.error (intentional error logging)
      local line_content="${match#*:*:}"
      [[ "$line_content" == *"console.error"* ]] && continue
      # Allow in logger/utils files
      [[ "$file" == *"logger"* ]] && continue

      local rel="${file#$PROJECT_ROOT/}"
      local line_num
      line_num=$(echo "$match" | cut -d: -f2)
      violations+=("$rel:$line_num")
    done < <(grep -rnPE "$pattern" "$search_dir/src" --include='*.ts' --include='*.tsx' --include='*.vue' --include='*.js' 2>/dev/null || true)
  done

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "debug_artifacts" "no debug artifacts in production code"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_warn "debug_artifacts" "${#violations[@]} debug artifacts: $detail"
  fi
}

# ─── Check 6: Type Safety ─────────────────────────────────────────────

check_type_safety() {
  echo "[6/6] TypeScript type safety" >&2

  local search_dir="${TARGET_DIR:-$WEB_DIR}"

  # Count `as any` usages in non-test production code
  local violations=()
  while IFS= read -r match; do
    [[ -z "$match" ]] && continue
    local file="${match%%:*}"
    [[ "$file" == *".spec."* ]] && continue
    [[ "$file" == *".test."* ]] && continue
    [[ "$file" == *"node_modules"* ]] && continue

    local rel="${file#$PROJECT_ROOT/}"
    local line_num
    line_num=$(echo "$match" | cut -d: -f2)
    violations+=("$rel:$line_num")
  done < <(grep -rn 'as any' "$search_dir/src" --include='*.ts' --include='*.tsx' --include='*.vue' 2>/dev/null || true)

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "type_safety" "no 'as any' escape hatches"
  else
    local count=${#violations[@]}
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 500)"
    detail="$(json_escape "$detail")"
    log_warn "type_safety" "$count 'as any' usages (aspirational target ≤10): $detail"
  fi
}

# ─── Main ─────────────────────────────────────────────────────────────

main() {
  if ! $OUTPUT_JSON; then
    echo "=== Harness Frontend Mechanized Checks ==="
    echo "Service: ${SERVICE_NAME:-all}"
    echo "Timestamp: $(timestamp)"
    echo ""
  fi

  set +e  # Allow checks to fail without exiting the script
  check_type_check
  check_unit_test
  check_build
  check_hardcoded_secrets
  check_debug_artifacts
  check_type_safety
  set -e

  # Count results
  local pass=0 fail=0 warn=0
  for result in "${RESULTS[@]}"; do
    local status
    status=$(echo "$result" | grep -oP '"status":"\K\w+')
    case "$status" in
      PASS) pass=$((pass + 1)) ;;
      FAIL) fail=$((fail + 1)) ;;
      WARN) warn=$((warn + 1)) ;;
    esac
  done

  if $OUTPUT_JSON; then
    printf '{\n'
    printf '  "timestamp": "%s",\n' "$(timestamp)"
    printf '  "service": "%s",\n' "${SERVICE_NAME:-all}"
    printf '  "results": [\n'
    local first=true
    for result in "${RESULTS[@]}"; do
      if $first; then first=false; else printf ',\n'; fi
      printf '    %s' "$result"
    done
    printf '\n  ],\n'
    printf '  "summary": {"pass": %d, "fail": %d, "warn": %d},\n' "$pass" "$fail" "$warn"
    printf '  "exit_code": %d\n' "$([ $fail -gt 0 ] && echo 1 || echo 0)"
    printf '}\n'
  else
    local n=0
    local labels=("type-check" "unit-test" "build" "hardcoded-secrets" "debug-artifacts" "type-safety")
    for result in "${RESULTS[@]}"; do
      local status label detail
      status=$(echo "$result" | grep -oP '"status":"\K\w+')
      detail=$(echo "$result" | grep -oP '"detail":"\K[^"]*')
      label="${labels[$n]}"
      case "$status" in
        PASS) echo "[PASS] $((n+1)). $label — $detail" ;;
        FAIL) echo "[FAIL] $((n+1)). $label — $detail" ;;
        WARN) echo "[WARN] $((n+1)). $label — $detail" ;;
      esac
      n=$((n + 1))
    done
    echo ""
    echo "=== Summary: $pass PASS, $fail FAIL, $warn WARN ==="
  fi

  return $EXIT_CODE
}

main "$@"
