#!/usr/bin/env bash
#
# harness-checks.sh — Mechanized QA checks for the Community-Home Harness pipeline.
#
# Usage:
#   ./.harness/skills/qa/scripts/harness-checks.sh [--service <service-name>] [--json]
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
#   4. Proto int64 jstype      — every int64 ID field must have [jstype = JS_STRING]
#   5. Go json:",string"       — every int64 ID field must use json:"...,string"
#   6. Cross-service DB import — no importing another service's model/ package
#   7. Error code format       — use errx constants, not magic numbers
#   8. Hardcoded secrets       — no password/token/secret literals in Go code
#   9. Knowledge graph freshness — graph should be synced after latest commit
#  10. CLAUDE.md structural data — warn if structural data (RPC/routes/DB tables) duplicated in CLAUDE.md
#  11. Proto→TypeScript alignment — every proto message field has a matching TS interface field
#  12. API Logic TODO stubs — no api/internal/logic/*.go should contain todo stubs
#  13. Response single-wrap — detect Logic functions returning types with embedded BaseResponse
#  14. Benchmark regression — compare go test -bench against stored baselines
#  15. API smoke test — curl new/modified REST endpoints to verify non-404 (non-blocking)
#  16. Memory index freshness — memory 索引与记忆文件保持同步
#  17. Git hygiene — gitlink/.gitmodules 一致、无孤儿 worktree 分支
#  18. Mutation testing — 有逻辑函数的测试有效性（变异测试，工具未装则 SKIP）
#
# ─── 项目策略 vs 通用引擎 边界 ────────────────────────────────────────
# 以下检查为 Community-Home 项目特有策略（非通用引擎逻辑），迁移本脚本到
# 其他项目时必须替换为对应项目的策略：
#
#   4. Proto int64 jstype — Snowflake ID 精度约束（19 位 > JS 安全整数）
#   5. Go json:",string"  — Snowflake ID 同上
#   6. Cross-service DB   — 服务间仅 gRPC、禁止直连他人 DB 的架构约束
#   7. Error code format  — 5 位 errx 错误码常量（禁用魔法数字）
#
# 其余检查（go build/vet/test、硬编码密钥、图谱新鲜度、TODO 桩、API 冒烟等）
# 为通用引擎逻辑，可跨项目复用。策略归属详见 .harness/config/project-policies.md

set -eu
# NOTE: pipefail intentionally disabled — grep returning 1 (no match) in
# pipelines like `changed_files | grep "$pattern"` would cause premature
# script exit. The script uses explicit EXIT_CODE tracking instead.

# ─── Config ───────────────────────────────────────────────────────────

PROJECT_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
SERVICE_NAME=""
OUTPUT_JSON=false
FULL_SCAN=false
EXIT_CODE=0

# Results array: each element is a JSON-like string for internal tracking
declare -a RESULTS

# ============================================================
# Load service registry (replaces hardcoded SVC_MODULE_MAP)
# ============================================================
REGISTRY_FILE="$PROJECT_ROOT/.harness/registry/services.json"

if [ ! -f "$REGISTRY_FILE" ]; then
  echo "❌ Error: Service registry not found at $REGISTRY_FILE" >&2
  echo "   Run: bash .harness/scripts/build-service-registry.sh" >&2
  exit 1
fi

# Load service module mappings into SVC_MODULE_MAP associative array
declare -A SVC_MODULE_MAP
while IFS='=' read -r svc module; do
  [ -n "$module" ] && SVC_MODULE_MAP["$svc"]="$module"
done < <(jq -r '.services[] | "\(.name)=\(.module)"' "$REGISTRY_FILE")

# Verify registry loaded
if [ ${#SVC_MODULE_MAP[@]} -eq 0 ]; then
  echo "❌ Error: Failed to load service registry" >&2
  exit 1
fi
# ============================================================

# ─── Helpers ──────────────────────────────────────────────────────────

timestamp() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

log_pass() {
  local check="$1" detail="$2"
  RESULTS+=("{\"check\":\"$check\",\"status\":\"PASS\",\"detail\":\"$detail\"}")
}

log_fail() {
  local check="$1" detail="$2" why="${3:-}" fix="${4:-}" example="${5:-}" reference="${6:-}"
  local json_result="{\"check\":\"$check\",\"status\":\"FAIL\",\"detail\":\"$detail\""
  [[ -n "$why" ]] && json_result+=",\"why\":\"$why\""
  [[ -n "$fix" ]] && json_result+=",\"fix\":\"$fix\""
  [[ -n "$example" ]] && json_result+=",\"example\":\"$example\""
  [[ -n "$reference" ]] && json_result+=",\"reference\":\"$reference\""
  json_result+="}"
  RESULTS+=("$json_result")
  EXIT_CODE=1
}

log_warn() {
  local check="$1" detail="$2" why="${3:-}" fix="${4:-}" example="${5:-}" reference="${6:-}"
  local json_result="{\"check\":\"$check\",\"status\":\"WARN\",\"detail\":\"$detail\""
  [[ -n "$why" ]] && json_result+=",\"why\":\"$why\""
  [[ -n "$fix" ]] && json_result+=",\"fix\":\"$fix\""
  [[ -n "$example" ]] && json_result+=",\"example\":\"$example\""
  [[ -n "$reference" ]] && json_result+=",\"reference\":\"$reference\""
  json_result+="}"
  RESULTS+=("$json_result")
}

# Get changed files from git diff (for diff-only scan mode).
# Returns patterns that can be piped: one file per line, relative to PROJECT_ROOT.
changed_files() {
  local pattern="${1:-*.go}"
  # 主仓库 diff（unstaged + staged + untracked）
  local files
  files="$(git diff --name-only HEAD 2>/dev/null; git diff --cached --name-only 2>/dev/null; git ls-files --others --exclude-standard 2>/dev/null)"
  # 嵌套仓库（gitlink）diff：主仓库只见 gitlink 不见内部 .go 改动，需进子仓库扫（盲区修复）
  local gl
  while IFS= read -r gl; do
    [ -z "$gl" ] && continue
    if [ -d "$gl/.git" ]; then
      files="$files
$(git -C "$gl" diff --name-only HEAD 2>/dev/null | sed "s|^|$gl/|")
$(git -C "$gl" diff --cached --name-only 2>/dev/null | sed "s|^|$gl/|")
$(git -C "$gl" ls-files --others --exclude-standard 2>/dev/null | sed "s|^|$gl/|")"
    fi
  done < <(git ls-files -s 2>/dev/null | grep '^160000' | awk '{print $4}')
  echo "$files" | sort -u | grep -E "\.(${pattern//\*/})\$" 2>/dev/null || true
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
  echo "[1/18] go build ./..." >&2
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
  echo "[2/18] go vet ./..." >&2
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
  echo "[3/18] go test ./... (with 0/0 + new-package detection)" >&2
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

  # New-package test gap: find packages with NEW non-test Go files in THIS change that lack tests.
  # FIX: scope = 本次工作树实际改动（generator 直接改主树后 QA 能看见），
  # 不再用 --since="7 days ago"（时间窗口会把历史 commit 的老代码也算进"本次新增"，导致 QA 审错范围）。
  local new_pkgs_no_test=()
  if [[ -n "$SERVICE_NAME" ]] && git -C "$TARGET_DIR" rev-parse --git-dir >/dev/null 2>&1; then
    local new_go_files
    new_go_files=$(cd "$PROJECT_ROOT" && changed_files 'go' | grep "^services/$SERVICE_NAME/" | grep -v '_test\.go$' | sed "s|^services/$SERVICE_NAME/||" || true)
    if [[ -n "$new_go_files" ]]; then
      local new_dirs
      new_dirs=$(echo "$new_go_files" | while read -r f; do [[ -n "$f" ]] && dirname "$f"; done | sort -u || true)
      for dir in $new_dirs; do
        [[ -z "$dir" || "$dir" == "." ]] && continue
        local base
        base=$(basename "$dir")
        # Skip packages where tests are not expected (handler/ subdirs, model, config, etc.)
        [[ "$base" =~ ^(model|config|types|handler|server|svc|middleware|vars)$ ]] && continue
        # Also skip handler subdirectories (e.g. api/internal/handler/review)
        [[ "$dir" =~ /handler/ ]] && continue
        local test_files
        test_files=$(find "$TARGET_DIR/$dir" -maxdepth 1 -name '*_test.go' 2>/dev/null || true)
        if [[ -z "$test_files" ]]; then
          new_pkgs_no_test+=("$dir")
        fi
      done
    fi
  fi

  # Count actual test functions as a cross-check
  local test_funcs=0
  if [[ -n "$SERVICE_NAME" ]]; then
    test_funcs=$(grep -r '^func Test' "$TARGET_DIR" --include="*_test.go" 2>/dev/null | wc -l || echo 0)
  else
    test_funcs=$(grep -r '^func Test' "$PROJECT_ROOT/services" --include="*_test.go" 2>/dev/null | wc -l || echo 0)
  fi

  if [[ ${#new_pkgs_no_test[@]} -gt 0 ]]; then
    local pkg_list
    pkg_list=$(IFS=,; echo "${new_pkgs_no_test[*]}")
    local pkg_list_escaped
    pkg_list_escaped="$(json_escape "$pkg_list")"
    log_warn "go_test" "${passed_packages}P/${failed_packages}F/${no_test_packages}N, ~${test_funcs} tests — NEW packages missing tests: $pkg_list_escaped"
  elif [[ $test_funcs -eq 0 ]]; then
    log_warn "go_test" "${passed_packages} packages passed but 0 TestXxx functions found — verify tests exist"
  else
    log_pass "go_test" "${passed_packages} packages passed, ~${test_funcs} test functions"
  fi
}

# ─── Check 4: Proto int64 jstype ─────────────────────────────────────

check_proto_jstype() {
  echo "[4/18] Proto int64 jstype" >&2
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

      # Only check ID fields (named 'id' or ending with '_id'), not timestamps/counts
      if echo "$line" | grep -qP 'int64\s+(\w*_id|id)\s*=' && ! echo "$line" | grep -q 'jstype'; then
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
    local why="Snowflake ID 是 19 位整数，超过 JavaScript Number.MAX_SAFE_INTEGER（约 16 位），JSON 传输时会精度丢失。"
    local fix="在 proto 文件的 int64 ID 字段后添加 [(gogoproto.jstype) = JS_STRING] 选项，确保 protojson 序列化时以字符串输出。"
    local example="api-proto/api/user/v1/user.proto:15 | api-proto/api/auth/v1/auth.proto:22"
    local reference=".harness/rules/项目编码规范.md §5 | .harness/docs/_archive/linters/patterns/proto-jstype.md"
    log_fail "proto_jstype" "${#violations[@]} violations: $detail" "$why" "$fix" "$example" "$reference"
  fi
}

# ─── Check 5: Go json:",string" (AST-based with regex fallback) ─────

check_json_string() {
  echo "[5/18] Go json:\",string\" (AST)" >&2
  local search_dir
  if [[ -n "$SERVICE_NAME" ]]; then
    search_dir="$TARGET_DIR"
  else
    search_dir="$PROJECT_ROOT/services"
  fi

  # Try AST-based check first
  local ast_checker="$PROJECT_ROOT/.harness/tools/go-ast-checker/go-ast-checker"

  # AST 检查：SERVICE_NAME 存在即调用 ast-checks.sh（其内部负责 go build 二进制）。
  # 修复死锁——此前 [[ -f ast_checker ]] 前置条件挡住构建，二进制永不存在、检查永不执行。
  if [[ -n "$SERVICE_NAME" ]]; then
    echo "  (using AST checker)" >&2

    # Run AST checks and capture JSON output.
    # FIX(管线健壮性): 外层 set -eu 下, ast-checks 返回 FAIL(exit 1)会让本命令替换
    # 直接终止脚本, 连 if 判断都到不了。用 `|| ast_rc=$?` 捕获退出码并继续解析。
    local ast_output ast_rc
    ast_output=$(bash "$PROJECT_ROOT/.harness/skills/qa/scripts/ast-checks.sh" \
      "$search_dir" "$SERVICE_NAME" "true" 2>/dev/null) || ast_rc=$?

    if [[ ${ast_rc:-0} -eq 0 ]]; then
      log_pass "ast_json_string" "all int64 ID fields have json:\",string\" (AST verified)"
      return
    else
      # Parse JSON results and report failures
      local violations=()
      while IFS= read -r item; do
        local check=$(echo "$item" | jq -r '.check // empty' 2>/dev/null)
        if [[ "$check" == "json_string_tag" ]]; then
          local detail=$(echo "$item" | jq -r '.detail // empty' 2>/dev/null)
          local location=$(echo "$item" | jq -r '.location // empty' 2>/dev/null)
          [[ -n "$location" ]] && violations+=("$location: $detail")
        fi
      done < <(echo "$ast_output" | jq -c '.[] // empty' 2>/dev/null)

      if [[ ${#violations[@]} -gt 0 ]]; then
        local detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
        detail="$(json_escape "$detail")"
        local why="Snowflake IDs exceed JavaScript Number.MAX_SAFE_INTEGER, must be transmitted as strings"
        local fix="Add 'string' option to json tag: json:\"field_name,string\""
        local example="UserId int64 \`json:\"user_id,string\"\`"
        local reference=".harness/rules/项目编码规范.md §5"
        log_fail "ast_json_string" "${#violations[@]} violations: $detail" "$why" "$fix" "$example" "$reference"
        return
      else
        log_pass "ast_json_string" "all int64 ID fields have json:\",string\" (AST verified)"
        return
      fi
    fi
  fi

  # Fallback to regex-based check
  echo "  (AST checker not available, using regex fallback)" >&2
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

      # Match: ID fields (ending with 'Id') with int64 and json:"..." tag, missing string option
      if echo "$line" | grep -qP '\w+Id\s+int64.*json:"' && ! echo "$line" | grep -qP 'json:"[^"]*string'; then
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
    local why="Go REST API 的 int64 ID 字段默认 JSON 序列化为数字，前端 JavaScript 解析时精度丢失。"
    local fix="在结构体的 int64 ID 字段的 json tag 中添加 string 选项：json:\\\"user_id,string\\\" 或 json:\\\"id,omitempty,string\\\""
    local example="services/user-service/api/internal/types/types.go:18 | services/auth-service/api/internal/types/types.go:25"
    local reference=".harness/rules/项目编码规范.md §5 | .harness/docs/_archive/linters/patterns/json-string.md"
    log_fail "json_string" "${#violations[@]} violations: $detail" "$why" "$fix" "$example" "$reference"
  fi
}

# ─── Check 6: Cross-service DB import ────────────────────────────────

check_cross_service_import() {
  echo "[6/18] Cross-service DB import" >&2
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

    # Check imports for other services' model/internal packages
    # 文件级子串搜索：覆盖直连/别名导入（import m "path/model"），并补 /internal
    # 盲区——此前只匹配 /model，跨服务 internal 包导入漏检
    for other_svc in "${!SVC_MODULE_MAP[@]}"; do
      [[ "$other_svc" == "$owner_svc" ]] && continue
      local mod_path="${SVC_MODULE_MAP[$other_svc]}"
      if grep -qE "\"${mod_path}/(model|internal)" "$gofile" 2>/dev/null; then
        local rel="${gofile#$PROJECT_ROOT/}"
        violations+=("$rel imports ${other_svc}/(model|internal)")
      fi
    done
  done < <(echo "$go_files")

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "cross_service_import" "no violations"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}")"
    detail="$(json_escape "$detail")"
    local why="服务间通信必须通过 gRPC。直接访问其他服务的数据库破坏服务边界，造成紧耦合。"
    local fix="1. 移除跨服务的 model 包导入\\n2. 在 svcCtx 中添加对应的 RPC 客户端（如 UserRpc）\\n3. 通过 RPC 调用获取数据：svcCtx.UserRpc.GetUserInfo(ctx, req)\\n4. 将 RPC 响应映射到 Logic 返回类型"
    local example="services/auth-service/api/internal/logic/verify_token_logic.go:28-35"
    local reference=".harness/rules/项目编码规范.md §1 | .harness/docs/_archive/linters/patterns/cross-service-rpc.md"
    log_fail "cross_service_import" "${#violations[@]} violations: $detail" "$why" "$fix" "$example" "$reference"
  fi
}

# ─── Check 7: Error code format ──────────────────────────────────────

check_error_codes() {
  echo "[7/18] Error code format" >&2
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
  echo "[8/18] Hardcoded secrets" >&2
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
    'password\s*[:=]{1,2}\s*"[^$"[:space:]]{4,}"'
    'secret\s*[:=]{1,2}\s*"[^$"[:space:]]{4,}"'
    'token\s*[:=]{1,2}\s*"[^$"[:space:]]{8,}"'
    'api_key\s*[:=]{1,2}\s*"[^$"[:space:]]{4,}"'
    'apikey\s*[:=]{1,2}\s*"[^$"[:space:]]{4,}"'
    'passwd\s*[:=]{1,2}\s*"[^$"[:space:]]{4,}"'
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
    done < <(grep -rPni "$pattern" "$search_dir" --include='*.go' 2>/dev/null || true)   # -i: 覆盖驼峰命名（apiKey/clientSecret/dbPassword/accessToken）
  done

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "hardcoded_secrets" "no secrets detected"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    local why="硬编码的密钥会被提交到 Git 历史，造成安全风险。即使后续删除，历史记录中仍然可见。"
    local fix="1. 将密钥移到 .env 文件（已在 .gitignore 中）\\n2. 在配置 YAML 中使用环境变量引用：Password: \\\${DB_PASSWORD}\\n3. 使用 configx.MustLoad 加载配置（自动展开环境变量）"
    local example="services/user-service/api/etc/user-api.yaml:12 | .env.example"
    local reference=".harness/rules/项目编码规范.md §7"
    log_fail "hardcoded_secrets" "${#violations[@]} potential secrets: $detail" "$why" "$fix" "$example" "$reference"
  fi
}

# ─── Check 9: Knowledge graph freshness ───────────────────────────────

check_graph_freshness() {
  echo "[9/18] Knowledge graph freshness" >&2
  local stamp_file="$PROJECT_ROOT/.harness/.graph_last_sync"

  if [[ ! -f "$stamp_file" ]]; then
    local why="知识图谱为 Agent 提供服务依赖、RPC 接口、数据表等上下文。未同步的图谱导致 graph-context.md 缺失。"
    local fix="运行同步脚本创建图谱：bash .harness/scripts/graph-sync.sh --full"
    local reference=".harness/rules/项目编码规范.md §6"
    log_fail "graph_freshness" "graph never synced" "$why" "$fix" "" "$reference"
    return
  fi

  local stamp
  stamp=$(cat "$stamp_file")
  local now
  now=$(date +%s)
  local age=$(( (now - stamp) / 3600 ))

  # Check if any *substantive code* changes since last sync.
  # 只关注会改变 Neo4j 图内容的文件（proto/go/ts/vue/yaml/yml），排除：
  #   - *_test.go（测试，graph-sync 的 CHANGED 检测同样排除）
  #   - **/graph-context.md（图导出产物，其提交是 graph-sync 的结果，非「图过期」信号）
  # 否则 graph-sync 自身产物（graph-context.md）的提交会形成「同步→提交→又 stale」死循环。
  local latest_commit=0
  for repo in "$PROJECT_ROOT" "$PROJECT_ROOT"/services/*/ "$PROJECT_ROOT"/api-proto/; do
    if [[ -d "$repo/.git" ]]; then
      local ts
      ts=$(git -C "$repo" log -1 --format=%ct -- \
        '*.proto' '*.go' '*.ts' '*.vue' '*.yaml' '*.yml' \
        ':(exclude,glob)**/*_test.go' \
        ':(exclude,glob)**/graph-context.md' \
        2>/dev/null || echo 0)
      if [[ $ts -gt $latest_commit ]]; then
        latest_commit=$ts
      fi
    fi
  done

  if [[ $latest_commit -gt $stamp ]]; then
    local why="知识图谱为 Agent 提供服务依赖、RPC 接口、数据表等上下文。过期的图谱会导致 Agent 使用错误的信息。"
    local fix="运行同步脚本更新图谱：bash .harness/scripts/graph-sync.sh"
    local reference=".harness/rules/项目编码规范.md §6 | services/*/docs/graph-context.md"
    log_fail "graph_freshness" "graph is stale (last sync: ${age}h ago, latest commit is newer)" "$why" "$fix" "" "$reference"
  else
    log_pass "graph_freshness" "graph up-to-date (synced ${age}h ago)"
  fi
}

# ─── Check 10: CLAUDE.md structural data ──────────────────────────

check_claude_structural_data() {
  echo "[10/18] CLAUDE.md structural data check" >&2
  local violations=()

  # Determine which CLAUDE.md files to scan
  local claude_files=()
  if [[ -n "$SERVICE_NAME" ]]; then
    local target="$PROJECT_ROOT/services/$SERVICE_NAME/CLAUDE.md"
    if [[ -f "$target" ]]; then
      claude_files+=("$target")
    fi
  else
    for svc_dir in "$PROJECT_ROOT"/services/*/; do
      local cf="${svc_dir}CLAUDE.md"
      [[ -f "$cf" ]] && claude_files+=("$cf")
    done
  fi

  for cf in "${claude_files[@]}"; do
    [[ ! -f "$cf" ]] && continue
    local rel="${cf#$PROJECT_ROOT/}"

    # Check for RPC tables
    if grep -q '|.*RPC.*|' "$cf" 2>/dev/null; then
      violations+=("$rel: contains RPC table (should be in graph-context.md)")
      continue
    fi

    # Check for REST route tables
    if grep -qP '^\| (GET|POST|PUT|DELETE|PATCH) ' "$cf" 2>/dev/null; then
      violations+=("$rel: contains REST route table (should be in graph-context.md)")
      continue
    fi

    # Check for database table listing section
    if grep -q '^## .*数据库表' "$cf" 2>/dev/null; then
      violations+=("$rel: contains database table listing (should be in graph-context.md)")
      continue
    fi

    # Check for model/table file listing section
    if grep -q '^## .*[Mm]odel.*目录' "$cf" 2>/dev/null; then
      violations+=("$rel: contains model/table file listing (should be in graph-context.md)")
      continue
    fi

    # Check for Go dependency list (## 依赖 section with module bullet items)
    local in_deps=0
    while IFS= read -r line; do
      if echo "$line" | grep -q '^## 依赖'; then
        in_deps=1
        continue
      fi
      if [[ $in_deps -eq 1 ]] && echo "$line" | grep -q '^## '; then
        in_deps=0
        continue
      fi
      if [[ $in_deps -eq 1 ]] && echo "$line" | grep -q '^- `github'; then
        violations+=("$rel: contains Go dependency list (should be in graph-context.md)")
        in_deps=0
        continue
      fi
    done < "$cf" || true
  done

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "claude_structural_data" "no structural data duplication in CLAUDE.md"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_warn "claude_structural_data" "${#violations[@]} warnings: $detail"
  fi
}

# ─── Check 11: Proto→TypeScript alignment ─────────────────────────

check_proto_ts_align() {
  echo "[11/18] Proto→TypeScript alignment" >&2
  local check_script="$PROJECT_ROOT/.harness/skills/qa/scripts/check-proto-ts-align.sh"

  if [[ ! -f "$check_script" ]]; then
    log_pass "proto_ts_align" "check script not found (skipped)"
    return
  fi

  local out rc
  set +e
  out="$("$check_script" 2>&1)"
  rc=$?
  set -e

  if [[ $rc -eq 0 ]]; then
    if echo "$out" | grep -q 'AUTO-MISMATCH'; then
      local auto_detail
      auto_detail="$(echo "$out" | grep 'AUTO-MISMATCH' | head -5 | tr '\n' '; ')"
      auto_detail="$(json_escape "$auto_detail")"
      log_warn "proto_ts_align" "自动同名匹配发现 TS 滞后字段（前端类型未同步 proto）: $auto_detail"
    else
      log_pass "proto_ts_align" "all proto fields match TS interfaces"
    fi
  else
    local detail
    detail="$(echo "$out" | grep -E '^(MISMATCH|MISSING)' | head -10 | tr '\n' '; ')"
    detail="$(json_escape "$detail")"
    log_fail "proto_ts_align" "violations: $detail"
  fi
}

# ─── Check 12: API Logic TODO stubs ──────────────────────────────────
# Detects goctl-generated TODO stubs that were never implemented.
# Pattern: "// todo: add your logic here and delete this line"
# These stubs return (nil, nil) causing the Handler to respond
# {code:0, data:null} — a "silent success" that masks missing functionality.

check_api_stubs() {
  echo "[12/18] API logic TODO stubs" >&2
  local target="$PROJECT_ROOT/services"
  [[ -n "$SERVICE_NAME" ]] && target="$PROJECT_ROOT/services/$SERVICE_NAME"

  if [[ ! -d "$target" ]]; then
    log_pass "api_stubs" "no service directory (skipped)"
    return
  fi

  local stubs
  # 覆盖 api 与 rpc 两套 logic（goctl 也会为 gRPC handler 生成 stub）；
  # 无 --service 时遍历所有服务的 logic 目录（原实现只查 services/ 导致空转）
  if [[ -n "$SERVICE_NAME" ]]; then
    stubs=$(grep -rl "todo: add your logic here" "$target/api/internal/logic/" "$target/rpc/internal/logic/" 2>/dev/null || true)
  else
    stubs=$(grep -rl "todo: add your logic here" services/*/api/internal/logic/ services/*/rpc/internal/logic/ 2>/dev/null || true)
  fi

  if [[ -z "$stubs" ]]; then
    log_pass "api_stubs" "no TODO stubs found in API logic"
  else
    local count
    count=$(echo "$stubs" | wc -l)
    local detail
    detail=$(echo "$stubs" | sed "s|$PROJECT_ROOT/||g" | tr '\n' '; ')
    detail="$(json_escape "$detail")"
    local why="goctl 生成的 TODO 桩返回 (nil, nil)，Handler 会响应 {code:0, data:null}，这是'静默成功'，掩盖了未实现的功能。"
    local fix="删除 '// todo: add your logic here and delete this line'，实现真正的业务逻辑。如果是占位接口，至少返回明确的错误：return nil, errx.NewCodeError(50001, \\\"功能未实现\\\")"
    local reference=".harness/rules/项目编码规范.md §9"
    log_fail "api_stubs" "${count} TODO stubs: $detail" "$why" "$fix" "" "$reference"
  fi
}

# ─── Check 13: Response single-wrap ──────────────────────────────────
# Detects Logic functions that return goctl-generated Response types
# (with embedded BaseResponse) when the Handler also wraps via response.Success().
# This causes double-nesting: {code:0, data:{code:0, data:{actual}}}
#
# Detection: find Logic functions whose return type is *types.XxxResponse
# (goctl Response types embed BaseResponse). These should return raw data instead.

check_response_wrap() {
  echo "[13/18] Response single-wrap" >&2
  local target="$PROJECT_ROOT/services"
  [[ -n "$SERVICE_NAME" ]] && target="$PROJECT_ROOT/services/$SERVICE_NAME"

  if [[ ! -d "$target" ]]; then
    log_pass "response_wrap" "no service directory (skipped)"
    return
  fi

  # Find Logic files whose function signature returns *types.XxxResponse
  # 覆盖 api 与 rpc 两套 logic（rpc logic 返回 types.Response 同样会 double-wrap）
  local violations
  if [[ -n "$SERVICE_NAME" ]]; then
    violations=$(grep -rn 'func (l \*.*Logic) .* \*types\.\w*Response' "$target/api/internal/logic/" "$target/rpc/internal/logic/" 2>/dev/null | grep -v test || true)
  else
    violations=$(grep -rn 'func (l \*.*Logic) .* \*types\.\w*Response' services/*/api/internal/logic/ services/*/rpc/internal/logic/ 2>/dev/null | grep -v test || true)
  fi

  if [[ -z "$violations" ]]; then
    log_pass "response_wrap" "no double-wrap risk detected"
  else
    local count
    count=$(echo "$violations" | wc -l)
    local detail
    detail=$(echo "$violations" | sed "s|$PROJECT_ROOT/||g" | tr '\n' '; ')
    detail="$(json_escape "$detail")"
    local why="Logic 返回 *types.XxxResponse（含 BaseResponse），Handler 再用 response.Success() 包装，造成双层嵌套：{code:0, data:{code:0, data:{...}}}"
    local fix="修改 Logic 返回类型为纯业务数据（struct 或 pointer），不使用 goctl 生成的 Response 类型。Handler 中用 response.Success(w, data) 包一层。"
    local example="services/ai-model-service/api/internal/logic/create_model_logic.go:25"
    local reference=".harness/rules/项目编码规范.md §9 | .harness/docs/_archive/linters/patterns/response-wrap.md"
    log_warn "response_wrap" "${count} Logic funcs return Response types (potential double-wrap): $detail" "$why" "$fix" "$example" "$reference"
  fi
}

# Check 14: Benchmark regression — compare against stored baselines (non-blocking)
check_bench_regression() {
  echo "[14/18] Benchmark regression" >&2
  local target="$PROJECT_ROOT/services"
  [[ -n "$SERVICE_NAME" ]] && target="$PROJECT_ROOT/services/$SERVICE_NAME"

  if [[ ! -d "$target" ]]; then
    log_pass "bench_regression" "no Go service directory (skipped)"
    return
  fi

  local baseline="$target/_bench_baseline.txt"
  local has_bench=false

  # Check if any benchmark functions exist
  if grep -rq 'func Benchmark' "$target" 2>/dev/null; then
    has_bench=true
  fi

  if ! $has_bench; then
    log_pass "bench_regression" "no benchmark functions — SKIP (tip: add benchmarks for hot paths)"
    return
  fi

  # Run benchmarks
  local bench_output
  bench_output=$(cd "$target" && go test ./... -bench=. -benchmem -benchtime=1s 2>/dev/null || true)

  if [[ ! -f "$baseline" ]]; then
    # No baseline yet — store current as baseline
    echo "$bench_output" | grep '^Benchmark' > "$baseline" 2>/dev/null || true
    log_pass "bench_regression" "baseline created ($(wc -l < "$baseline" 2>/dev/null || echo 0) benchmarks)"
    return
  fi

  # Compare: extract benchmark name and ns/op from current run
  local regressed_20=0 regressed_50=0 details=""
  while IFS= read -r bline; do
    [[ -z "$bline" ]] && continue
    local bname=$(echo "$bline" | awk '{print $1}')
    local bns=$(echo "$bline" | awk '{print $3}')  # ns/op
    [[ -z "$bname" || -z "$bns" ]] && continue

    local cns
    cns=$(echo "$bench_output" | grep "^${bname}[[:space:]]" | awk '{print $3}' | head -1)
    [[ -z "$cns" ]] && continue  # benchmark removed or renamed

    # Calculate % change
    if [[ $bns -gt 0 ]]; then
      local change=$(( (cns - bns) * 100 / bns ))
      if [[ $change -gt 50 ]]; then
        regressed_50=$((regressed_50 + 1))
        details="${details}${bname}: ${bns}→${cns} ns/op (+${change}%); "
      elif [[ $change -gt 20 ]]; then
        regressed_20=$((regressed_20 + 1))
        details="${details}${bname}: ${bns}→${cns} ns/op (+${change}%); "
      fi
    fi
  done < "$baseline"

  local detail_escaped
  detail_escaped="$(json_escape "$details")"

  if [[ $regressed_50 -gt 0 ]]; then
    log_fail "bench_regression" "${regressed_50} benchmarks >50% slower; ${regressed_20} >20% slower: $detail_escaped"
  elif [[ $regressed_20 -gt 0 ]]; then
    log_warn "bench_regression" "${regressed_20} benchmarks >20% slower: $detail_escaped"
  else
    log_pass "bench_regression" "all benchmarks within 20% of baseline"
  fi
}

# Check 15: API smoke test — curl new/modified REST endpoints to verify non-404 (non-blocking)
check_api_smoke() {
  echo "[15/18] API smoke test" >&2
  local target="$PROJECT_ROOT/services"
  [[ -n "$SERVICE_NAME" ]] && target="$PROJECT_ROOT/services/$SERVICE_NAME"

  if [[ ! -d "$target" ]]; then
    log_pass "api_smoke" "no service directory (skipped)"
    return
  fi

  # Only run for services with an API layer (routes.go)
  local routes_file="$target/api/internal/handler/routes.go"
  if [[ ! -f "$routes_file" ]]; then
    log_pass "api_smoke" "no API routes file (skipped)"
    return
  fi

  # Find newly added route paths in the diff (relative to HEAD~1)
  local new_routes
  new_routes=$(cd "$target" && git diff HEAD~1 -- api/internal/handler/routes.go 2>/dev/null \
    | grep '^\+.*Path:' \
    | grep -oP '"/\K[^"]+' || true)

  if [[ -z "$new_routes" ]]; then
    log_pass "api_smoke" "no new routes detected in diff"
    return
  fi

  # Read the service port from config
  local port=0
  if [[ -f "$target/api/etc/"*.yaml ]]; then
    port=$(grep -oP 'Port:\s*\K\d+' "$target/api/etc/"*.yaml 2>/dev/null | head -1)
  fi
  if [[ "$port" -eq 0 ]]; then
    log_warn "api_smoke" "cannot determine API port — SKIP"
    return
  fi

  # Derive the API URL prefix from the service name (goctl convention:
  # /api/<service-name-without-"-service">). FIX: this used to hardcode
  # "/api/moderation" which sent every other service's smoke test to the
  # moderation-service routes — wrong service, guaranteed false results.
  local svc_bare="${SERVICE_NAME%-service}"
  local api_prefix="/api/${svc_bare:-${SERVICE_NAME}}"

  # Test each new route
  local failed=0 total=0 unreachable=0 details=""
  while IFS= read -r route; do
    [[ -z "$route" ]] && continue
    total=$((total + 1))
    local url="http://127.0.0.1:${port}${api_prefix}${route}"
    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "$url" 2>/dev/null || echo "000")
    # 401/403 = route exists (auth required), 200 = success. 404 = route missing.
    if [[ "$http_code" == "404" ]]; then
      failed=$((failed + 1))
      details="${details}${url}=${http_code}; "
    elif [[ "$http_code" == "000" ]]; then
      # FIX: service not running is NOT a route-missing failure — report as skip
      unreachable=$((unreachable + 1))
    fi
  done <<< "$new_routes"

  local detail_escaped
  detail_escaped="$(json_escape "$details")"

  if [[ $total -eq 0 ]]; then
    log_pass "api_smoke" "no routes to test"
  elif [[ $unreachable -gt 0 && $failed -eq 0 ]]; then
    log_warn "api_smoke" "服务未运行（${unreachable}/${total} 个路由无法连接）— 跳过冒烟验证，非路由缺失"
  elif [[ $failed -gt 0 ]]; then
    log_warn "api_smoke" "${failed}/${total} new routes returned 404 — service may need restart: $detail_escaped"
  else
    log_pass "api_smoke" "${total} new routes verified (non-404)"
  fi
}

# ─── Check 16: Memory Index Freshness ────────────────────────────────

check_memory_index() {
  echo "[16/18] Memory Index Freshness" >&2

  local index_file="$PROJECT_ROOT/.harness/knowledge/memory/.memory-index.json"
  local memory_dir="$PROJECT_ROOT/.harness/knowledge/memory"

  # Check if index file exists
  if [[ ! -f "$index_file" ]]; then
    log_fail "memory_index" "索引文件不存在，运行: bash .harness/scripts/memory-index-build.sh"
    return
  fi

  # Check if jq is available
  if ! command -v jq &> /dev/null; then
    log_warn "memory_index" "jq 不可用，跳过新鲜度检查"
    return
  fi

  # Extract index generation time
  local index_time
  index_time=$(jq -r '.generated_at' "$index_file" 2>/dev/null || echo "")

  if [[ -z "$index_time" ]]; then
    log_fail "memory_index" "索引文件格式错误，无法读取 generated_at"
    return
  fi

  # Convert to epoch (Linux date)
  local index_epoch
  index_epoch=$(date -d "$index_time" +%s 2>/dev/null || echo "0")

  if [[ "$index_epoch" -eq 0 ]]; then
    # Try macOS date format
    index_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$index_time" +%s 2>/dev/null || echo "0")
  fi

  if [[ "$index_epoch" -eq 0 ]]; then
    local fix="重新生成索引：bash .harness/scripts/memory-index-build.sh"
    log_warn "memory_index" "无法解析索引时间: $index_time" "" "$fix"
    return
  fi

  # Find the newest memory file modification time
  local newest_memory
  newest_memory=$(find "$memory_dir" -name "*.md" -not -name "MEMORY.md" -not -name "MAINTENANCE.md" -type f -printf '%T@\n' 2>/dev/null | sort -rn | head -1)

  if [[ -z "$newest_memory" ]]; then
    log_pass "memory_index" "无记忆文件"
    return
  fi

  # Convert to integer (remove decimal)
  local newest_memory_epoch
  newest_memory_epoch=$(printf "%.0f" "$newest_memory")

  # Compare timestamps
  if [[ "$newest_memory_epoch" -gt "$index_epoch" ]]; then
    local newest_file
    newest_file=$(find "$memory_dir" -name "*.md" -not -name "MEMORY.md" -not -name "MAINTENANCE.md" -type f -printf '%T@ %p\n' 2>/dev/null | sort -rn | head -1 | cut -d' ' -f2-)
    local detail
    detail="索引过期 (最新记忆: $(basename "$newest_file"))"
    local why="Memory 索引用于语义搜索和关键词触发。过期索引会导致新记忆无法被检索到。"
    local fix="重新生成索引：bash .harness/scripts/memory-index-build.sh"
    local reference=".harness/knowledge/memory/MAINTENANCE.md"
    log_fail "memory_index" "$detail" "$why" "$fix" "" "$reference"
  else
    log_pass "memory_index" "索引最新 (生成于 $index_time)"
  fi
}

# ─── Check 17: Git hygiene ─────────────────────────────────────────

check_git_hygiene() {
  echo "[17/18] Git hygiene" >&2
  local violations=()

  # 1. gitlink ↔ .gitmodules 一致性
  if [ -f "$PROJECT_ROOT/.gitmodules" ]; then
    while IFS= read -r gl; do
      [ -z "$gl" ] && continue
      if ! grep -qF "path = $gl" "$PROJECT_ROOT/.gitmodules" 2>/dev/null; then
        violations+=("gitlink 无 .gitmodules 条目: $gl")
      fi
    done < <(git -C "$PROJECT_ROOT" ls-files -s 2>/dev/null | grep '^160000' | awk '{print $4}')
  fi

  # 2. 孤儿 worktree 分支
  local orphans
  orphans=$(git -C "$PROJECT_ROOT" branch --format='%(refname:short)' 2>/dev/null | grep -c '^worktree-wf_' || true)
  if [[ "${orphans:-0}" -gt 0 ]]; then
    violations+=("${orphans} 个孤儿 worktree-wf_* 分支未清理")
  fi

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "git_hygiene" "gitlink/.gitmodules 一致，无孤儿 worktree 分支"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}")"
    detail="$(json_escape "$detail")"
    local why="git 治理漂移：子模块未登记或临时分支残留，破坏环境一致性。"
    local fix="补 .gitmodules 条目，删除孤儿 worktree 分支。详见 .harness/rules/Git治理规范.md"
    local reference=".harness/rules/Git治理规范.md"
    log_warn "git_hygiene" "$detail" "$why" "$fix" "" "$reference"
  fi
}

# ─── Check 18: Mutation testing（测试有效性）────────────────────────

check_mutation_testing() {
  echo "[18/18] Mutation testing" >&2
  # 变异测试：对「有逻辑函数」验证测试有效性（替代 RED 证据，见 harness-pipeline-fix/design-tdd-evidence.md T3/T4）
  # 用 gomu（github.com/sivchari/gomu，现代变异测试工具，增量/门禁/并行/JSON）。
  # 注意：gomu --incremental 对 monorepo/go.work 挂起，用 --incremental=false + diff .go 文件范围；
  #      目录级太慢（全包变异），只对 diff 的非测试 .go 文件跑。
  if ! command -v gomu >/dev/null 2>&1; then
    log_pass "mutation_testing" "gomu 未安装，跳过（可选门禁；安装: go install github.com/sivchari/gomu/cmd/gomu@latest）"
    return
  fi
  local go_files
  go_files=$(changed_files 'go' | grep '\.go$' | grep -v '_test\.go$' || true)
  if [[ -z "$go_files" ]]; then
    log_pass "mutation_testing" "无 Go 逻辑变更，跳过"
    return
  fi
  # 限制目标数，避免太慢（最多 3 个 diff 文件；单文件 ~30-60s）
  local targets=()
  while IFS= read -r f; do
    [ -z "$f" ] && continue
    [ -f "$f" ] || continue
    targets+=("$f")
    [[ ${#targets[@]} -ge 3 ]] && break
  done <<< "$go_files"
  if [[ ${#targets[@]} -eq 0 ]]; then
    log_pass "mutation_testing" "diff 无现存 Go 文件，跳过"
    return
  fi
  local out rc
  set +e
  out=$(cd "$PROJECT_ROOT" && timeout 180 gomu run --incremental=false --threshold 80 --ci-mode --fail-on-gate --timeout 15 --workers 4 --output console "${targets[@]}" 2>&1)
  rc=$?
  set -e
  # gomu 的 gate 失败退出码不可靠（pipe 吞码），须解析输出：score 或「below minimum threshold」标志
  local score gate_failed below_threshold
  score=$(echo "$out" | grep -oE 'Mutation Score: [0-9.]+%' | tail -1 | grep -oE '[0-9.]+' | tail -1)
  gate_failed=$(echo "$out" | grep -cE 'below minimum threshold|quality gate failed' || true)
  below_threshold=false
  if [[ -n "$score" ]]; then
    awk -v s="$score" 'BEGIN{exit !(s < 80)}' && below_threshold=true
  fi
  if [[ "${gate_failed:-0}" -gt 0 ]] || $below_threshold; then
    local why="变异存活率高=测试未覆盖关键分支（gomu 实测 permission helpers.go 仅 48-51%）。"
    local fix="为 diff 有逻辑函数补断言覆盖未被变异杀死的分支，或降低阈值（默认 80）。详见 .harness/changes/harness-pipeline-fix/design-tdd-evidence.md"
    local reference=".harness/changes/harness-pipeline-fix/design-tdd-evidence.md"
    log_warn "mutation_testing" "变异分数不足: ${score}%（<80%）" "$why" "$fix" "" "$reference"
  elif [[ $rc -eq 124 ]]; then
    log_warn "mutation_testing" "变异测试超时（180s），跳过"
  else
    log_pass "mutation_testing" "变异分数 ${score:-?}%（≥80% 或未解析到分数）"
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
  check_graph_freshness
  check_claude_structural_data
  check_proto_ts_align
  check_api_stubs
  check_response_wrap
  check_bench_regression
  check_api_smoke
  check_memory_index
  check_git_hygiene
  check_mutation_testing
  # Note: frontend checks use separate script: harness-checks-frontend.sh

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
    local labels=("go build" "go vet" "go test" "proto int64 jstype" "json:\",string\"" "cross-service DB import" "error code format" "hardcoded secrets" "graph freshness" "CLAUDE.md structural data" "proto->TS alignment" "API logic stubs" "response single-wrap" "benchmark regression" "API smoke test" "memory index freshness" "git hygiene" "mutation testing")
    for result in "${RESULTS[@]}"; do
      local status label detail why fix example reference
      status=$(echo "$result" | grep -oP '"status":"\K\w+')
      detail=$(echo "$result" | grep -oP '"detail":"\K[^"]*')
      why=$(echo "$result" | grep -oP '"why":"\K[^"]*' || echo "")
      fix=$(echo "$result" | grep -oP '"fix":"\K[^"]*' || echo "")
      example=$(echo "$result" | grep -oP '"example":"\K[^"]*' || echo "")
      reference=$(echo "$result" | grep -oP '"reference":"\K[^"]*' || echo "")
      label="${labels[$n]}"

      case "$status" in
        PASS)
          echo "[PASS] $((n+1)). $label — $detail"
          ;;
        FAIL)
          echo "[FAIL] $((n+1)). $label — $detail"
          [[ -n "$why" ]] && echo "  WHY: $why"
          [[ -n "$fix" ]] && echo "  FIX: $fix"
          [[ -n "$example" ]] && echo "  EXAMPLE: $example"
          [[ -n "$reference" ]] && echo "  REFERENCE: $reference"
          ;;
        WARN)
          echo "[WARN] $((n+1)). $label — $detail"
          [[ -n "$why" ]] && echo "  WHY: $why"
          [[ -n "$fix" ]] && echo "  FIX: $fix"
          [[ -n "$example" ]] && echo "  EXAMPLE: $example"
          [[ -n "$reference" ]] && echo "  REFERENCE: $reference"
          ;;
      esac
      n=$((n + 1))
    done
    echo ""
    echo "=== Summary: $pass PASS, $fail FAIL, $warn WARN ==="
  fi

  # ── 打点：记录 QA 门禁调用（服务 + PASS/FAIL 数，为流水线复盘提供第一手数据）──
  bash "$PROJECT_ROOT/.harness/scripts/log-usage.sh" harness-checks \
    service="${SERVICE_NAME:-all}" pass="$pass" fail="$fail" warn="$warn" 2>/dev/null || true

  return $EXIT_CODE
}

main "$@"
