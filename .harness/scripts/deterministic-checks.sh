#!/usr/bin/env bash
# Deterministic Checks - Run before AI judgment
# Usage: bash deterministic-checks.sh <service-dir> [--json]

set -euo pipefail

SERVICE_DIR="${1:-}"
JSON_OUTPUT="${2:-false}"

if [ -z "$SERVICE_DIR" ]; then
  echo "Usage: $0 <service-dir> [--json]" >&2
  exit 1
fi

if [ ! -d "$SERVICE_DIR" ]; then
  echo "Error: Service directory not found: $SERVICE_DIR" >&2
  exit 1
fi

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT/$SERVICE_DIR"

# Results storage
declare -A RESULTS
OVERALL_STATUS="PASS"

# Helper: Record result
record_result() {
  local check="$1"
  local status="$2"
  local detail="${3:-}"

  RESULTS["$check"]="$status|$detail"

  if [ "$status" = "FAIL" ]; then
    OVERALL_STATUS="FAIL"
  fi
}

# Check 1: Compilation
check_compile() {
  echo "  [1/6] Checking compilation..." >&2

  if go build ./... > /tmp/compile.log 2>&1; then
    record_result "compile" "PASS" "All packages compile successfully"
  else
    local error=$(head -5 /tmp/compile.log | tr '\n' ' ')
    record_result "compile" "FAIL" "Compilation failed: $error"
  fi
}

# Check 2: Tests
check_tests() {
  echo "  [2/6] Running tests..." >&2

  # Run tests and capture results
  if go test ./... -v > /tmp/test.log 2>&1; then
    local test_count=$(grep -c "^=== RUN" /tmp/test.log || echo 0)
    record_result "tests" "PASS" "All tests passed ($test_count tests)"
  else
    # Check for 0/0 false pass
    if grep -q "testing: warning: no tests to run" /tmp/test.log; then
      record_result "tests" "FAIL" "No tests found (0/0 false pass)"
    else
      local failed=$(grep -c "^--- FAIL:" /tmp/test.log || echo 0)
      record_result "tests" "FAIL" "$failed test(s) failed"
    fi
  fi
}

# Check 3: Test Coverage
check_coverage() {
  echo "  [3/6] Checking test coverage..." >&2

  # Run tests with coverage
  if go test ./... -coverprofile=/tmp/coverage.out > /dev/null 2>&1; then
    # Calculate total coverage
    local coverage=$(go tool cover -func=/tmp/coverage.out 2>/dev/null | grep total | awk '{print $3}' | sed 's/%//')

    if [ -n "$coverage" ]; then
      # Check against threshold (80%)
      if (( $(echo "$coverage >= 80" | bc -l) )); then
        record_result "coverage" "PASS" "Coverage: ${coverage}% (≥80%)"
      else
        record_result "coverage" "WARN" "Coverage: ${coverage}% (<80%)"
      fi
    else
      record_result "coverage" "WARN" "Could not determine coverage"
    fi
  else
    record_result "coverage" "SKIP" "Tests failed, coverage not calculated"
  fi
}

# Check 4: Static Analysis
check_static_analysis() {
  echo "  [4/6] Running static analysis..." >&2

  if go vet ./... > /tmp/vet.log 2>&1; then
    record_result "static_analysis" "PASS" "No issues found"
  else
    local issues=$(wc -l < /tmp/vet.log)
    local sample=$(head -2 /tmp/vet.log | tr '\n' '; ')
    record_result "static_analysis" "WARN" "$issues issue(s) found: $sample"
  fi
}

# Check 5: Code Format
check_format() {
  echo "  [5/6] Checking code format..." >&2

  local unformatted=$(gofmt -l . 2>/dev/null | grep -v vendor | wc -l)

  if [ "$unformatted" -eq 0 ]; then
    record_result "format" "PASS" "All files properly formatted"
  else
    local files=$(gofmt -l . 2>/dev/null | grep -v vendor | head -3 | tr '\n' ', ')
    record_result "format" "WARN" "$unformatted file(s) need formatting: $files"
  fi
}

# Check 6: Dependencies
check_dependencies() {
  echo "  [6/6] Verifying dependencies..." >&2

  if [ -f "go.mod" ]; then
    if go mod verify > /tmp/mod.log 2>&1; then
      record_result "dependencies" "PASS" "All dependencies verified"
    else
      local error=$(cat /tmp/mod.log | tr '\n' ' ')
      record_result "dependencies" "FAIL" "Dependency verification failed: $error"
    fi
  else
    record_result "dependencies" "SKIP" "No go.mod found"
  fi
}

# Run all checks
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
echo "Running Deterministic Checks" >&2
echo "Service: $SERVICE_DIR" >&2
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2

check_compile
check_tests
check_coverage
check_static_analysis
check_format
check_dependencies

echo "" >&2

# Output results
if [ "$JSON_OUTPUT" = "--json" ]; then
  # JSON output
  echo "{"
  echo "  \"service\": \"$SERVICE_DIR\","
  echo "  \"timestamp\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\","
  echo "  \"overall_status\": \"$OVERALL_STATUS\","
  echo "  \"checks\": ["

  first=true
  for check in "${!RESULTS[@]}"; do
    IFS='|' read -r status detail <<< "${RESULTS[$check]}"

    if [ "$first" = true ]; then
      first=false
    else
      echo ","
    fi

    echo -n "    {\"check\": \"$check\", \"status\": \"$status\", \"detail\": \"$detail\"}"
  done

  echo ""
  echo "  ]"
  echo "}"
else
  # Human-readable output
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
  echo "Results" >&2
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2

  for check in compile tests coverage static_analysis format dependencies; do
    if [ -n "${RESULTS[$check]:-}" ]; then
      IFS='|' read -r status detail <<< "${RESULTS[$check]}"

      case "$status" in
        PASS)
          echo "  ✅ $check: $detail" >&2
          ;;
        FAIL)
          echo "  ❌ $check: $detail" >&2
          ;;
        WARN)
          echo "  ⚠️  $check: $detail" >&2
          ;;
        SKIP)
          echo "  ⏭️  $check: $detail" >&2
          ;;
      esac
    fi
  done

  echo "" >&2
  echo "Overall: $OVERALL_STATUS" >&2
fi

# Exit code
if [ "$OVERALL_STATUS" = "PASS" ]; then
  exit 0
else
  exit 1
fi
