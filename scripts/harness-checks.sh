#!/usr/bin/env bash
#
# harness-checks.sh — Mechanized QA checks for the Community-Home Harness pipeline.
#
# Usage:
#   ./scripts/harness-checks.sh [--service <service-name>] [--json]
#
# Options:
#   --service <name>   Scope checks to a single service directory under services/
#   --json             Output results as JSON (default: human-readable text)
#   -h, --help         Show help
#
# Exit code: 0 if all checks pass, non-zero if any check fails.
#
# Checks:
#   1. go build ./...          — compilation
#   2. go vet ./...            — static analysis
#   3. go test ./...           — unit tests (with 0/0 false-pass detection)
#   4. Proto int64 jstype      — every int64 field must have [jstype = JS_STRING]
#   5. Go json:",string"       — every int64 API field must use json:"...,string"
#   6. Cross-service DB import — no importing another service's model/ package
#   7. Error code format       — use errx constants, not magic numbers
#   8. Hardcoded secrets       — no password/token/secret literals in Go code

set -euo pipefail

# ─── Config ───────────────────────────────────────────────────────────

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SERVICE_NAME=""
OUTPUT_JSON=false
FULL_SCAN=false
EXIT_CODE=0

# Results array: each element is a JSON-like string for internal tracking
declare -a RESULTS

# Service → module path prefix mapping (for cross-service import check)
declare -A SVC_MODULE_MAP
SVC_MODULE_MAP["user-service"]="github.com/guxiao1976/community-user"
SVC_MODULE_MAP["auth-service"]="github.com/guxiao1976/community-auth"
SVC_MODULE_MAP["permission-service"]="github.com/guxiao1976/community-permission"
SVC_MODULE_MAP["file-service"]="github.com/guxiao1976/community-file"
SVC_MODULE_MAP["master-data-service"]="github.com/guxiao1976/community-master-data-service"
SVC_MODULE_MAP["moderation-service"]="github.com/guxiao1976/community-moderation-service"
SVC_MODULE_MAP["monitoring-service"]="github.com/guxiao1976/community-monitoring"
SVC_MODULE_MAP["community-hub-service"]="github.com/guxiao1976/community-hub"
SVC_MODULE_MAP["ai-model-service"]="github.com/guxiao/community-and-home/services/ai-model"

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

# Get changed files from git diff (for diff-only scan mode).
# Returns patterns that can be piped: one file per line, relative to PROJECT_ROOT.
changed_files() {
  local pattern="${1:-*.go}"
  # Unstaged + staged changes, only files matching pattern
  (git diff --name-only HEAD 2>/dev/null; git diff --cached --name-only 2>/dev/null; git ls-files --others --exclude-standard 2>/dev/null) \
    | sort -u | grep -E "\.(${pattern//\*/})\$" 2>/dev/null || true
  # Fallback: if no changes, scan nothing (git diff returns empty)
}

# Escape a string for JSON (basic: backslash + quotes)
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
    --full)
      FULL_SCAN=true
      shift
      ;;
    -h|--help)
      sed -n '2,22p' "$0" | sed 's/^# //'
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--service <name>] [--json]"
      exit 2
      ;;
  esac
done

# Target directory for go commands
if [[ -n "$SERVICE_NAME" ]]; then
  TARGET_DIR="$PROJECT_ROOT/services/$SERVICE_NAME"
  if [[ ! -d "$TARGET_DIR" ]]; then
    echo "Error: service directory not found: $TARGET_DIR"
    exit 2
  fi
else
  TARGET_DIR="$PROJECT_ROOT"
fi

# ─── Check 1: go build ───────────────────────────────────────────────

check_go_build() {
  echo "[1/8] go build ./..." >&2
  local out err rc
  cd "$TARGET_DIR"
  set +e
  out="$(go build ./... 2>&1)"
  rc=$?
  set -e
  cd "$PROJECT_ROOT"

  if [[ $rc -eq 0 ]]; then
    log_pass "go_build" "compilation succeeded"
  else
    # Extract first 3 error lines for concise reporting
    local err_summary
    err_summary="$(echo "$out" | grep -i 'error' | head -3 | tr '\n' '; ')"
    err_summary="$(json_escape "$err_summary")"
    log_fail "go_build" "compilation failed: $err_summary"
  fi
}

# ─── Check 2: go vet ─────────────────────────────────────────────────

check_go_vet() {
  echo "[2/8] go vet ./..." >&2
  local out rc
  cd "$TARGET_DIR"
  set +e
  out="$(go vet ./... 2>&1)"
  rc=$?
  set -e
  cd "$PROJECT_ROOT"

  if [[ $rc -eq 0 ]]; then
    log_pass "go_vet" "no issues"
  else
    # go vet returns non-zero for issues it considers severe
    local warn_summary
    warn_summary="$(echo "$out" | tail -5 | tr '\n' '; ')"
    warn_summary="$(json_escape "$warn_summary")"
    log_fail "go_vet" "vet found issues: $warn_summary"
  fi
}

# ─── Check 3: go test (with 0/0 detection) ───────────────────────────

check_go_test() {
  echo "[3/8] go test ./... (with 0/0 detection)" >&2
  local out rc
  cd "$TARGET_DIR"
  set +e
  out="$(go test ./... -count=1 2>&1)"
  rc=$?
  set -e
  cd "$PROJECT_ROOT"

  # Count results
  local total_packages=0 passed_packages=0 failed_packages=0 no_test_packages=0
  local total_tests=0

  # Parse "ok  pkg  time" lines (passed)
  local ok_lines
  ok_lines="$(echo "$out" | grep -c '^ok ' || true)"
  passed_packages=$ok_lines

  # Parse "FAIL pkg  time" lines (failed)
  local fail_lines
  fail_lines="$(echo "$out" | grep -c '^FAIL\s' || true)"
  failed_packages=$fail_lines

  # Parse "?   pkg [no test files]" lines (no tests)
  local no_test_lines
  no_test_lines="$(echo "$out" | grep -c '^\?\s' || true)"
  no_test_packages=$no_test_lines

  total_packages=$((passed_packages + failed_packages + no_test_packages))

  if [[ $rc -ne 0 ]] || [[ $failed_packages -gt 0 ]]; then
    local fail_detail
    fail_detail="$(echo "$out" | grep '^--- FAIL' | head -5 | tr '\n' '; ')"
    fail_detail="$(json_escape "$fail_detail")"
    log_fail "go_test" "${passed_packages}P/${failed_packages}F/${no_test_packages}N packages; failures: $fail_detail"
    return
  fi

  # 0/0 false-pass detection: if ALL packages have no test files
  if [[ $no_test_packages -gt 0 ]] && [[ $passed_packages -eq 0 ]] && [[ $failed_packages -eq 0 ]]; then
    log_warn "go_test" "${no_test_packages} packages have no test files — potential 0/0 false-pass"
    return
  fi

  # Count actual test functions as a cross-check
  local test_funcs=0
  if [[ -n "$SERVICE_NAME" ]]; then
    test_funcs=$(grep -r '^func Test' "$TARGET_DIR" --include="*_test.go" 2>/dev/null | wc -l || echo 0)
  else
    test_funcs=$(grep -r '^func Test' "$PROJECT_ROOT/services" --include="*_test.go" 2>/dev/null | wc -l || echo 0)
  fi

  if [[ $test_funcs -eq 0 ]]; then
    log_warn "go_test" "${passed_packages} packages passed but 0 TestXxx functions found — verify tests exist"
  else
    log_pass "go_test" "${passed_packages} packages passed, ~${test_funcs} test functions"
  fi
}

# ─── Check 4: Proto int64 jstype ─────────────────────────────────────

check_proto_jstype() {
  echo "[4/8] Proto int64 jstype" >&2
  local proto_dir="$PROJECT_ROOT/api-proto/api"
  local violations=()

  if [[ ! -d "$proto_dir" ]]; then
    log_pass "proto_jstype" "no proto directory found (skipped)"
    return
  fi

  # Determine which proto files to scan
  local proto_files
  if $FULL_SCAN; then
    proto_files=$(find "$proto_dir" -name '*.proto' -type f | sort)
    echo "  (full scan)" >&2
  else
    proto_files=$(cd "$PROJECT_ROOT" && changed_files 'proto' | sed "s|^|$PROJECT_ROOT/|" | sort)
    if [[ -z "$proto_files" ]]; then
      log_pass "proto_jstype" "no proto changes in diff (skipped)"
      return
    fi
    echo "  (diff scan: $(echo "$proto_files" | wc -l) proto files)" >&2
  fi

  while IFS= read -r proto_file; do
    [[ -z "$proto_file" ]] && continue
    [[ ! -f "$proto_file" ]] && continue
    local line_num=0
    while IFS= read -r line; do
      line_num=$((line_num + 1))
      [[ -z "${line// }" ]] && continue
      [[ "$line" =~ ^[[:space:]]*// ]] && continue
      [[ "$line" =~ ^[[:space:]]*\* ]] && continue

      if echo "$line" | grep -qP 'int64\s+\w+\s*=' && ! echo "$line" | grep -q 'jstype'; then
        local field
        field=$(echo "$line" | grep -oP 'int64\s+\K\w+')
        local rel="${proto_file#$proto_dir/}"
        violations+=("$rel:$line_num:$field")
      fi
    done < "$proto_file"
  done < <(echo "$proto_files")

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "proto_jstype" "all int64 fields have jstype"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_fail "proto_jstype" "${#violations[@]} violations: $detail"
  fi
}

# ─── Check 5: Go json:",string" ──────────────────────────────────────

check_json_string() {
  echo "[5/8] Go json:\",string\"" >&2
  local search_dir
  if [[ -n "$SERVICE_NAME" ]]; then
    search_dir="$TARGET_DIR"
  else
    search_dir="$PROJECT_ROOT/services"
  fi

  local violations=()

  # Determine which Go files to scan
  local go_files
  if $FULL_SCAN; then
    go_files=$(find "$search_dir" -name '*.go' -not -name '*_test.go' -not -path '*/vendor/*' | sort)
    echo "  (full scan)" >&2
  else
    go_files=$(cd "$PROJECT_ROOT" && changed_files 'go' | { if [[ -n "$SERVICE_NAME" ]]; then grep "^services/$SERVICE_NAME/"; else grep "^services/"; fi; } | sed "s|^|$PROJECT_ROOT/|" | sort)
    if [[ -z "$go_files" ]]; then
      log_pass "json_string" "no Go changes in diff (skipped)"
      return
    fi
    echo "  (diff scan: $(echo "$go_files" | wc -l) Go files)" >&2
  fi

  while IFS= read -r gofile; do
    [[ -z "$gofile" ]] && continue
    [[ ! -f "$gofile" ]] && continue
    local in_struct=0 line_num=0
    while IFS= read -r line; do
      line_num=$((line_num + 1))
      # Track struct boundaries (simplistic but effective)
      if echo "$line" | grep -qP '^\s*type\s+\w+\s+struct\s*\{'; then
        in_struct=1
        continue
      fi
      [[ $in_struct -eq 1 ]] && echo "$line" | grep -q '^\s*\}' && { in_struct=0; continue; }
      [[ $in_struct -eq 0 ]] && continue

      # Match: field line with int64 and json:"..." tag
      if echo "$line" | grep -qP 'int64.*json:"' && ! echo "$line" | grep -qP 'json:"[^"]*string'; then
        # Skip path:"" / form:"" / header:"" / db:"" tags (not JSON API fields)
        if echo "$line" | grep -qP '(path|form|header|db):"'; then
          continue
        fi
        local field
        field=$(echo "$line" | grep -oP '^\s*\K\w+')
        local rel="${gofile#$PROJECT_ROOT/}"
        violations+=("$rel:$line_num:$field")
      fi
    done < "$gofile"
  done < <(echo "$go_files")

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "json_string" "all int64 API fields have json:\",string\""
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_fail "json_string" "${#violations[@]} violations: $detail"
  fi
}

# ─── Check 6: Cross-service DB import ────────────────────────────────

check_cross_service_import() {
  echo "[6/8] Cross-service DB import" >&2
  local search_dir
  if [[ -n "$SERVICE_NAME" ]]; then
    search_dir="$TARGET_DIR"
  else
    search_dir="$PROJECT_ROOT/services"
  fi

  local violations=()
  local current_svc="${SERVICE_NAME}"

  # Determine which Go files to scan
  local go_files
  if $FULL_SCAN; then
    go_files=$(find "$search_dir" -name '*.go' -not -name '*_test.go' -not -path '*/vendor/*' | sort)
    echo "  (full scan)" >&2
  else
    go_files=$(cd "$PROJECT_ROOT" && changed_files 'go' | { if [[ -n "$SERVICE_NAME" ]]; then grep "^services/$SERVICE_NAME/"; else grep "^services/"; fi; } | sed "s|^|$PROJECT_ROOT/|" | sort)
    if [[ -z "$go_files" ]]; then
      log_pass "cross_service_import" "no Go changes in diff (skipped)"
      return
    fi
    echo "  (diff scan: $(echo "$go_files" | wc -l) Go files)" >&2
  fi

  while IFS= read -r gofile; do
    [[ -z "$gofile" ]] && continue
    [[ ! -f "$gofile" ]] && continue
    # Determine which service this file belongs to
    local owner_svc=""
    for svc in "${!SVC_MODULE_MAP[@]}"; do
      if echo "$gofile" | grep -q "services/$svc/"; then
        owner_svc="$svc"
        break
      fi
    done
    [[ -z "$owner_svc" ]] && continue

    # Check imports for other services' model packages
    while IFS= read -r import_line; do
      for other_svc in "${!SVC_MODULE_MAP[@]}"; do
        [[ "$other_svc" == "$owner_svc" ]] && continue
        local mod_path="${SVC_MODULE_MAP[$other_svc]}"
        if echo "$import_line" | grep -q "\"${mod_path}/model\""; then
          local rel="${gofile#$PROJECT_ROOT/}"
          violations+=("$rel imports ${other_svc}/model")
        fi
      done
    done < <(grep -P '^\s*"[^"]+' "$gofile" 2>/dev/null || true)
  done < <(echo "$go_files")

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "cross_service_import" "no violations"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}")"
    detail="$(json_escape "$detail")"
    log_fail "cross_service_import" "${#violations[@]} violations: $detail"
  fi
}

# ─── Check 7: Error code format ──────────────────────────────────────

check_error_codes() {
  echo "[7/8] Error code format" >&2
  local search_dir
  if [[ -n "$SERVICE_NAME" ]]; then
    search_dir="$TARGET_DIR"
  else
    search_dir="$PROJECT_ROOT/services"
  fi

  local violations=()

  # Find calls to errx.NewCodeError / errx.Wrap / errx.Wrapf with raw integers
  # Pattern: errx.NewCodeError(70001, ...) — raw integer as first arg
  while IFS= read -r match; do
    local file="${match%%:*}"
    local rest="${match#*:}"
    local line_num="${rest%%:*}"
    local content="${rest#*:}"

    # Extract the integer argument
    local code_val
    code_val=$(echo "$content" | grep -oP 'errx\.\w+\(\s*\K\d+')
    if [[ -n "$code_val" ]]; then
      # Skip 0 (success code)
      [[ "$code_val" == "0" ]] && continue
      local rel="${file#$PROJECT_ROOT/}"
      violations+=("$rel:$line_num:raw=$code_val")
    fi
  done < <(grep -rnP 'errx\.(NewCodeError|Wrap|Wrapf)\(\s*\d+' "$search_dir" --include='*.go' 2>/dev/null || true)

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "error_codes" "no magic numbers found (all use named constants or 0)"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_warn "error_codes" "${#violations[@]} magic numbers: $detail"
  fi
}

# ─── Check 8: Hardcoded secrets ──────────────────────────────────────

check_hardcoded_secrets() {
  echo "[8/8] Hardcoded secrets" >&2
  local search_dir
  if [[ -n "$SERVICE_NAME" ]]; then
    search_dir="$TARGET_DIR"
  else
    search_dir="$PROJECT_ROOT/services"
  fi

  local violations=()

  # Search for literal assignment patterns that look like secrets
  # Exclude: test files, yaml configs, env var references (${...})
  local patterns=(
    'password\s*[:=]\s*"[^$"[:space:]]{4,}"'
    'secret\s*[:=]\s*"[^$"[:space:]]{4,}"'
    'token\s*[:=]\s*"[^$"[:space:]]{8,}"'
    'api_key\s*[:=]\s*"[^$"[:space:]]{4,}"'
    'apikey\s*[:=]\s*"[^$"[:space:]]{4,}"'
    'passwd\s*[:=]\s*"[^$"[:space:]]{4,}"'
  )

  for pattern in "${patterns[@]}"; do
    while IFS= read -r match; do
      [[ -z "$match" ]] && continue
      local file="${match%%:*}"
      # Skip test files and configs
      [[ "$file" == *_test.go ]] && continue
      [[ "$file" == *.yaml ]] && continue
      [[ "$file" == *.yml ]] && continue
      [[ "$file" == *.env* ]] && continue
      # Skip if line contains ${ (env var reference)
      local line_content="${match#*:*:}"
      [[ "$line_content" == *'${'* ]] && continue

      local rel="${file#$PROJECT_ROOT/}"
      local line_num
      line_num=$(echo "$match" | cut -d: -f2)
      violations+=("$rel:$line_num:potential secret")
    done < <(grep -rnPE "$pattern" "$search_dir" --include='*.go' 2>/dev/null || true)
  done

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "hardcoded_secrets" "no secrets detected"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_fail "hardcoded_secrets" "${#violations[@]} potential secrets: $detail"
  fi
}

# ─── Main ─────────────────────────────────────────────────────────────

main() {
  if ! $OUTPUT_JSON; then
    echo "=== Harness Mechanized Checks ==="
    echo "Service: ${SERVICE_NAME:-all}"
    echo "Timestamp: $(timestamp)"
    echo ""
  fi

  check_go_build
  check_go_vet
  check_go_test
  check_proto_jstype
  check_json_string
  check_cross_service_import
  check_error_codes
  check_hardcoded_secrets

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
    # Output full JSON — use printf to avoid bash interpreting inner quotes
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
    # Human-readable output
    local n=0
    local labels=("go build" "go vet" "go test" "proto int64 jstype" "json:\",string\"" "cross-service DB import" "error code format" "hardcoded secrets")
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
