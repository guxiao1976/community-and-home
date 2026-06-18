#!/usr/bin/env bash
# harness-gate-check.sh — Pipeline gate enforcement
# Usage: bash .harness/scripts/harness-gate-check.sh --phase <0-6> --change <name>
# Exit: 0=PASS, non-zero=FAIL (blocks progress to next phase)
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CHANGES_DIR="$PROJECT_ROOT/.harness/changes"
CHANGE=""; PHASE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --phase) PHASE="$2"; shift 2 ;;
    --change) CHANGE="$2"; shift 2 ;;
    *) echo "Usage: $0 --phase <0-6> --change <name>"; exit 1 ;;
  esac
done
[[ -z "$PHASE" || -z "$CHANGE" ]] && { echo "Usage: $0 --phase <0-6> --change <name>"; exit 1; }

CHANGE_DIR="$CHANGES_DIR/$CHANGE"
EXIT_CODE=0

pass() { echo "[GATE PASS] $1"; }
fail() { echo "[GATE FAIL] $1"; EXIT_CODE=1; }
warn() { echo "[GATE WARN] $1"; }

echo "=== Harness Gate Check: Phase $PHASE — $CHANGE ==="

if [[ "$PHASE" == "0" ]]; then
  [[ -f "$CHANGE_DIR/request.md" ]] || { fail "request.md not found"; exit 1; }
  grep -q "路径" "$CHANGE_DIR/request.md" || fail "missing 路径 in request.md"
  grep -q "理由" "$CHANGE_DIR/request.md" || fail "missing 理由 in request.md"
  grep -q "涉及服务" "$CHANGE_DIR/request.md" || fail "missing 涉及服务 in request.md"
  [[ $EXIT_CODE -eq 0 ]] && pass "request.md with required fields"

elif [[ "$PHASE" == "1" ]]; then
  [[ -f "$CHANGE_DIR/proposal.md" ]] || { fail "proposal.md not found"; exit 1; }
  grep -q "追溯" "$CHANGE_DIR/proposal.md" || warn "proposal.md missing traceability"
  [[ -d "$CHANGE_DIR/specs" ]] && [[ -n "$(ls -A "$CHANGE_DIR/specs" 2>/dev/null)" ]] || warn "specs/ missing or empty"
  pass "proposal.md found"

elif [[ "$PHASE" == "2" ]]; then
  review_dir="$CHANGE_DIR/review"
  [[ -d "$review_dir" ]] || { fail "review/ not found — Phase 2 skipped"; exit 1; }
  found=0; approved=0
  for lens in coverage structure clarity; do
    f="$review_dir/spec_review_${lens}_v1.md"
    [[ -f "$f" ]] && found=$((found + 1))
    grep -q "APPROVED" "$f" 2>/dev/null && approved=$((approved + 1))
  done
  [[ $found -lt 3 ]] && fail "$found/3 review lenses found"
  [[ $approved -lt 2 ]] && fail "$approved/3 approved (need >=2/3)"
  [[ $EXIT_CODE -eq 0 ]] && pass "3/3 reviews, $approved/3 approved"

elif [[ "$PHASE" == "3" ]]; then
  [[ -f "$CHANGE_DIR/design.md" ]] || { fail "design.md not found"; exit 1; }
  [[ -f "$CHANGE_DIR/tasks.md" ]] || { fail "tasks.md not found"; exit 1; }
  grep -qi "TDD\|RED.*GREEN\|测试.*驱动" "$CHANGE_DIR/tasks.md" || warn "tasks.md no TDD reference"
  grep -q "TBD\|TODO\|FIXME" "$CHANGE_DIR/tasks.md" && warn "placeholders in tasks.md"
  pass "design.md + tasks.md found"

elif [[ "$PHASE" == "4" ]]; then
  proto_dir="$PROJECT_ROOT/api-proto"
  [[ -d "$proto_dir" ]] || { pass "no api-proto — skip"; exit 0; }
  has_change=$(git -C "$proto_dir" diff --name-only HEAD~10 2>/dev/null | grep -c '\.proto$' | tr -d ' \n' || echo 0)
  [[ $has_change -gt 0 ]] && pass "proto changes detected" || pass "no proto changes"

elif [[ "$PHASE" == "5" ]]; then
  svc_count=0; qa_fail=0; rev_services=0; rev_pass=0
  for svc_path in "$PROJECT_ROOT/services/"*/; do
    qa="$svc_path/_qa.md"
    if [[ -f "$qa" ]]; then
      svc_count=$((svc_count + 1))
      real_fails=$(grep "FAIL" "$qa" 2>/dev/null | grep -vc "graph freshness" || echo 0)
      [[ $real_fails -gt 0 ]] && qa_fail=$((qa_fail + 1))
    fi
    has_review=0; has_approved=0
    for rf in "$svc_path"/_review*.md; do
      [[ -f "$rf" ]] || continue
      has_review=1
      grep -qE "APPROVED|2/3" "$rf" 2>/dev/null && has_approved=1
    done
    if [[ $has_review -eq 1 ]]; then
      rev_services=$((rev_services + 1))
      [[ $has_approved -eq 1 ]] && rev_pass=$((rev_pass + 1))
    fi
  done
  echo "  Services with QA: $svc_count (QA FAIL: $qa_fail)"
  echo "  Services with Review: $rev_services (passed: $rev_pass)"
  [[ $svc_count -eq 0 ]] && { fail "no services with _qa.md — Phase 5 not executed"; exit 1; }
  [[ $qa_fail -gt 0 ]] && fail "$qa_fail service(s) QA FAIL"
  [[ $rev_services -eq 0 ]] && warn "no services with _review.md"
  [[ $EXIT_CODE -eq 0 ]] && pass "QA+Review complete"

elif [[ "$PHASE" == "6" ]]; then
  [[ -d "$CHANGE_DIR/impl" ]] || { fail "impl/ not found — not archived"; exit 1; }
  [[ -f "$CHANGE_DIR/summary.md" ]] || { fail "summary.md not found"; exit 1; }
  pass "impl/ + summary.md archived"

else
  echo "Invalid phase: $PHASE (0-6)"; exit 2
fi

echo "=== Gate $PHASE: $([ $EXIT_CODE -eq 0 ] && echo PASS || echo FAIL) ==="
exit $EXIT_CODE
