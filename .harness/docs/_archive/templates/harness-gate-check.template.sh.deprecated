#!/usr/bin/env bash
#
# harness-gate-check-v2.sh — 阶段门禁检查（阻断式验证）
#
# Usage:
#   bash .harness/scripts/harness-gate-check-v2.sh --phase <5|6> --change <change-name>
#
# Options:
#   --phase <5|6>       检查哪个阶段的门禁（5=编码完成, 6=集成归档）
#   --change <name>     变更名称（.harness/changes/<name>/）
#   -h, --help          显示帮助
#
# Exit code: 0 if all checks pass, 1 if any check fails (blocks progression)
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PHASE=""
CHANGE_NAME=""

log_pass() { echo "✅ $1"; }
log_fail() { echo "❌ $1"; }
log_warn() { echo "⚠️  $1"; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --phase) PHASE="$2"; shift 2 ;;
    --change) CHANGE_NAME="$2"; shift 2 ;;
    -h|--help) sed -n '2,13p' "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "Error: Unknown option $1"; exit 2 ;;
  esac
done

[[ -z "$PHASE" || ( "$PHASE" != "5" && "$PHASE" != "6" ) ]] && { echo "Error: --phase must be 5 or 6"; exit 2; }
[[ -z "$CHANGE_NAME" ]] && { echo "Error: --change is required"; exit 2; }

CHANGE_DIR="$PROJECT_ROOT/.harness/changes/$CHANGE_NAME"
[[ ! -d "$CHANGE_DIR" ]] && { echo "Error: change directory not found: $CHANGE_DIR"; exit 2; }

echo ""
echo "🚪 Phase $PHASE 门禁检查 — $CHANGE_NAME"
echo ""

FAILED=0

if [[ "$PHASE" == "5" ]]; then
  # Phase 5: 编码阶段 → 集成验证
  
  [[ ! -d "$CHANGE_DIR/impl" ]] && { log_fail "impl/ 目录不存在"; FAILED=1; } || log_pass "impl/ 目录存在"
  
  if [[ -d "$CHANGE_DIR/impl" ]]; then
    QA_PASS=0
    for svc_dir in "$CHANGE_DIR/impl"/*; do
      [[ ! -d "$svc_dir" ]] && continue
      SVC=$(basename "$svc_dir")
      
      if [[ ! -f "$svc_dir/_qa.md" ]]; then
        log_fail "$SVC: 缺少 _qa.md"
        FAILED=1
      elif grep -q "^VERDICT.*PASS" "$svc_dir/_qa.md"; then
        log_pass "$SVC: QA PASS"
        QA_PASS=$((QA_PASS + 1))
      else
        log_fail "$SVC: QA 未通过"
        FAILED=1
      fi
      
      REVIEW_FILES=$(find "$svc_dir" -name "_review*.md" 2>/dev/null | wc -l)
      if [[ $REVIEW_FILES -eq 0 ]]; then
        log_fail "$SVC: 缺少 Review 报告"
        FAILED=1
      else
        PASS=$(grep -l "^VERDICT.*PASS" "$svc_dir"/_review*.md 2>/dev/null | wc -l)
        TOTAL=$REVIEW_FILES
        THRESHOLD=$([[ $TOTAL -eq 1 ]] && echo 1 || echo 2)
        
        if [[ $PASS -ge $THRESHOLD ]]; then
          log_pass "$SVC: Review 通过 ($PASS/$TOTAL)"
        else
          log_fail "$SVC: Review 未通过 ($PASS/$TOTAL，需要 ≥$THRESHOLD)"
          FAILED=1
        fi
      fi
    done
    
    [[ $QA_PASS -eq 0 ]] && { log_fail "没有任何服务通过 QA"; FAILED=1; }
  fi
  
  echo ""
  [[ $FAILED -eq 0 ]] && { echo "✅ Phase 5 门禁通过 — 允许进入集成验证"; exit 0; } || { echo "❌ Phase 5 门禁失败 — 阻塞"; exit 1; }
fi

if [[ "$PHASE" == "6" ]]; then
  # Phase 6: 集成验证 → 交付
  
  [[ ! -f "$CHANGE_DIR/summary.md" ]] && { log_fail "summary.md 不存在"; FAILED=1; } || {
    log_pass "summary.md 存在"
    for section in "阶段 0" "阶段 5" "阶段 6"; do
      grep -q "$section" "$CHANGE_DIR/summary.md" && log_pass "包含: $section" || { log_fail "缺少: $section"; FAILED=1; }
    done
  }
  
  [[ ! -d "$CHANGE_DIR/impl" ]] && { log_fail "impl/ 目录不存在"; FAILED=1; } || {
    QA_CNT=$(find "$CHANGE_DIR/impl" -name "_qa.md" 2>/dev/null | wc -l)
    REV_CNT=$(find "$CHANGE_DIR/impl" -name "_review*.md" 2>/dev/null | wc -l)
    
    [[ $QA_CNT -eq 0 ]] && { log_fail "impl/ 无 QA 报告"; FAILED=1; } || log_pass "impl/ 包含 $QA_CNT 个 QA 报告"
    [[ $REV_CNT -eq 0 ]] && { log_fail "impl/ 无 Review 报告"; FAILED=1; } || log_pass "impl/ 包含 $REV_CNT 个 Review 报告"
  }
  
  INDEX="$PROJECT_ROOT/.harness/changes/INDEX.md"
  [[ -f "$INDEX" ]] && {
    grep -q "$CHANGE_NAME" "$INDEX" && log_pass "INDEX.md 已更新" || { log_fail "INDEX.md 未更新"; FAILED=1; }
  } || { log_fail "INDEX.md 不存在"; FAILED=1; }
  
  echo ""
  cd "$PROJECT_ROOT"
  go build ./... >/dev/null 2>&1 && log_pass "go build ./... 通过" || { log_fail "go build 失败"; FAILED=1; }
  
  echo ""
  [[ $FAILED -eq 0 ]] && { echo "✅ Phase 6 门禁通过 — 允许交付"; exit 0; } || { echo "❌ Phase 6 门禁失败 — 阻塞"; exit 1; }
fi
