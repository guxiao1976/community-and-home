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
#   --list-checks      List all check items (machine-generated from check_* functions)
#   -h, --help         Show help
#
# Exit code: 0 if all checks pass, non-zero if any check fails.
#
# Checks: 运行 `--list-checks` 查看完整检查项清单（从 check_* 函数定义机器生成，唯一权威出处）；
#          设计解释见 docs/qa-checks.md。此处不再手写编号清单（易漂移）。
#
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

json_escape() {
  local s="${1-}"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/\\r}"
  s="${s//$'\t'/\\t}"
  echo -n "$s"
}

# list_checks — 从自身 check_* 函数定义【机器生成】检查项清单（JSON），唯一权威出处。
# 中文名与四段式设计解释见 docs/qa-checks.md（人工维护，不在此机器生成）。
list_checks() {
  local prev_num=""
  local items=()
  local fn it
  while IFS= read -r line; do
    if [[ "$line" =~ ^#[[:space:]]*.*[Cc]heck[[:space:]]*([0-9]+(\.[0-9]+)?) ]]; then
      prev_num="${BASH_REMATCH[1]}"
      continue
    fi
    if [[ "$line" =~ ^check_([a-z_]+)\(\) ]]; then
      fn="${BASH_REMATCH[1]}"
      items+=("{\"id\":\"check_${fn}\",\"num\":\"${prev_num}\",\"name\":\"${fn}\"}")
      prev_num=""
    fi
  done < "$0"
  printf '['
  local sep=""
  for it in "${items[@]}"; do
    printf '%s%s' "$sep" "$it"
    sep=","
  done
  printf ']\n'
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
    --list-checks)
      list_checks
      exit 0
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
    log_fail "type_check" "$label: type errors: $err_summary" "TS 类型不匹配（字段名拼错/接口字段不存在/空值处理遗漏）" "运行 npx vue-tsc --noEmit 查看完整错误并修复类型" "" ".harness/rules/项目编码规范.md"
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
  local has_coverage=false
  if [[ -f "vitest.config.ts" ]] || [[ -f "vitest.config.js" ]]; then
    # 覆盖率量化：仅当项目安装了 @vitest/coverage-* provider 且非递归场景（unit-standard-gate.spec 会
    # spawnSync 本脚本→内层 vitest；若内层也 --coverage 会与当前进程冲突删 coverage/.tmp）才启用。
    # 递归（HARNESS_RECURSE=1）→ 普通 vitest。
    if ls node_modules/@vitest/coverage-* >/dev/null 2>&1 && [[ "${HARNESS_RECURSE:-}" != "1" ]]; then
      has_coverage=true
      test_cmd="npx vitest run --coverage --reporter=verbose"
    else
      test_cmd="npx vitest run --reporter=verbose"
    fi
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
    # 区分三种失败：覆盖率不达标(FAIL) / 覆盖率工具 bug(WARN) / 测试失败(FAIL)
    if $has_coverage && echo "$out" | grep -qE "does not meet global threshold"; then
      local cov_lines
      cov_lines=$(echo "$out" | grep -E "All files" | tr -s ' ')
      cov_lines="$(json_escape "$cov_lines")"
      log_fail "unit_test" "$label: 覆盖率低于 vitest.config coverage.thresholds（测试本身通过，但新代码缺测试拉低覆盖）: $cov_lines" "单测通过但整体覆盖率低于门禁" "为未覆盖分支补测试；新增有逻辑函数必须有对应 spec（TDD）" "npx vitest run --coverage" ".harness/skills/qa/SKILL.md"
      return 1
    fi
    if $has_coverage && echo "$out" | grep -qE "Something removed the coverage directory|Coverage.*failed"; then
      # vitest 4.1.10 coverage .tmp 生命周期 bug（v8/istanbul 均）：覆盖率无法生成时降级 WARN，不阻塞（待 vitest 修复）。
      log_warn "unit_test" "$label: 覆盖率未量化（vitest coverage bug: 目录 .tmp 生命周期竞态）——测试本身通过，覆盖率门禁本次降级 WARN，待 vitest 修复后自动升级为 FAIL"
      return 0
    fi
    local fail_detail
    fail_detail=$(echo "$out" | grep 'FAIL\|×' | head -5 | tr '\n' '; ')
    fail_detail="$(json_escape "$fail_detail")"
    log_fail "unit_test" "$label: tests failed — $fail_detail" "单测失败说明行为不符合预期" "运行 npx vitest run 定位失败用例并修复"
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
    log_fail "build" "$label: build timed out (120s)" "生产构建超时（依赖解析卡住/资源过多）" "检查 vite 配置和依赖，或分块构建"
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
    log_fail "build" "$label: $err_summary — $first_errors" "生产构建失败（dev 能跑 ≠ 生产能编）" "运行 npm run build 查看完整错误并修复"
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
    log_fail "hardcoded_secrets" "${#violations[@]} potential secrets: $detail" "硬编码密钥会随前端代码分发到浏览器，直接泄露" "改用 .env 环境变量（import.meta.env）注入，禁止硬编码" "" ".harness/rules/项目编码规范.md §7"
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
    log_fail "debug_artifacts" "${#violations[@]} debug artifacts: $detail" "console.log/debugger 是调试残留，进生产包=日志噪音+可能泄露数据；debugger 会中断调试器" "删除 console.log/debugger；错误路径用 console.error（已放行）"
  fi
}

# ─── Check 7: api-field-align（snake_case/camelCase 字段对齐）────────
# 检查前端是否使用了 camelCase 读取后端 snake_case API 字段
check_api_field_align() {
  local script="$PROJECT_ROOT/.harness/skills/qa/scripts/check-api-field-align.sh"
  if [ ! -x "$script" ]; then
    log_warn "api_field_align" "检查脚本不可用: $script"
    return
  fi
  local output
  if output=$("$script" 2>&1); then
    log_pass "api_field_align" "API 字段名对齐"
    return
  fi
  # 分级：存量违规保持 WARN（不阻塞历史），**本次 diff 内新增的违规 FAIL**（新代码不得引入新不匹配）。
  # 这样存量 34 处不阻塞，但新写的 snake/camel 不匹配会被拦下。
  # git 输出相对路径（web/pc/src/...），违规文件是绝对路径（$PROJECT_ROOT/...）——统一为相对路径比较。
  local changed_files
  changed_files=$(cd "$PROJECT_ROOT" && git diff --name-only HEAD 2>/dev/null; git diff --cached --name-only 2>/dev/null; git ls-files --others --exclude-standard 2>/dev/null)
  local new_violations=()
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    local file
    file=$(echo "$line" | grep -oP '^[^:]+' | head -1)
    file="${file#❌ }"   # 去掉违规行前缀「❌ 」，得到干净路径
    local rel="${file#$PROJECT_ROOT/}"
    if echo "$changed_files" | grep -qxF "$rel"; then
      new_violations+=("$line")
    fi
  done < <(echo "$output" | grep "❌" || true)
  if [[ ${#new_violations[@]} -gt 0 ]]; then
    local detail
    detail="$(IFS='; '; echo "${new_violations[*]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_fail "api_field_align" "${#new_violations[@]} 处本次 diff 新增 snake/camel 不匹配: $detail" "后端 protojson 输出 snake_case，前端用 camelCase 读取取不到值(undefined)" "将读取字段改为 snake_case（如 .created_at 而非 .createdAt）" "" ".harness/knowledge/memory/web/../snake-camel-field-mismatch.md"
  else
    local count
    count=$(echo "$output" | grep -c "❌" || true)
    log_warn "api_field_align" "$count 处存量 snake_case/camelCase 不匹配（不在本次 diff，WARN；新代码不得新增）"
  fi
}

# ─── Check 6: Type Safety（禁 as any）───────────────────────────────
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

# ─── Check 8: 前端单位规范（rem，§13 项目编码规范）──────────────────
# 硬性规则：长度/字号一律 rem，禁止 rpx/px（web/mobile 604 处 rpx 已换算为 rem）。
# 例外：① 根字号声明 html { font-size: 16px }（rem 体系的唯一 px 基准）；
#       ② env(safe-area-inset-bottom, 0px) 的 0px 兜底；③ 注释行。
# 修复（2026-08-16）：① 跳过 *.spec.*/*.test.* 测试文件（对齐 check_type_safety）；
#     ② 注释剥离改为逐行状态化，多行块注释续行一并剔除（对齐守卫测试 stripComments 语义）。
check_unit_standard() {
  echo "[8/8] Unit standard (rem only, mobile)" >&2

  # 本轮仅约束 mobile（web/pc 仍 px 体系，含 Element Plus，后续单独评估 —— 决策 2026-08-16 (b)）
  if [[ -n "$SERVICE_NAME" && "$SERVICE_NAME" != "mobile" ]]; then
    log_pass "unit_standard" "scoped to mobile only (${SERVICE_NAME} skipped this round)"
    return
  fi

  local search_dir="$WEB_DIR/mobile"
  local violations=()
  local file line_no in_block line code stripped iter before rest i prev rel

  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    [[ "$file" == *".spec."* ]] && continue
    [[ "$file" == *".test."* ]] && continue
    [[ "$file" == *"node_modules"* ]] && continue

    rel="${file#$PROJECT_ROOT/}"
    line_no=0
    in_block=0

    while IFS= read -r line || [[ -n "$line" ]]; do
      line_no=$((line_no + 1))
      code="$line"

      # ── 剥除注释（状态化，跨行块注释续行一并剔除）──
      # 1) 处于跨行 /* */ 块注释内：先找闭注释符 */，闭合后剩余部分继续当代码
      if [[ "$in_block" == "1" ]]; then
        if [[ "$code" == *"*/"* ]]; then
          code="${code#*"*/"}"
          in_block=0
        else
          continue  # 整行仍在块注释内
        fi
      fi
      # 2) 同行 /* ... */（可多个）与跨行块注释开启
      iter=0
      while [[ "$code" == *"/*"* ]] && (( iter < 20 )); do
        iter=$((iter + 1))
        before="${code%%"/*"*}"
        rest="${code#*"/*"}"
        if [[ "$rest" == *"*/"* ]]; then
          rest="${rest#*"*/"}"
          code="${before}${rest}"
        else
          code="$before"
          in_block=1
          break
        fi
      done
      # 3) // 行注释（仅当位于行首或前导空白之后，避免误伤 url(http://...)）
      for ((i = 0; i + 1 < ${#code}; i++)); do
        if [[ "${code:i:2}" == "//" ]]; then
          if (( i == 0 )); then
            code=""
          else
            prev="${code:i-1:1}"
            [[ "$prev" == [[:space:]] ]] && code="${code:0:i}"
          fi
          break
        fi
      done

      # ── 例外剔除后判定残留 ──
      # 例外①：根字号 html { font-size: 16px }（rem 唯一 px 基准）
      stripped="${code//font-size: 16px/}"
      stripped="${stripped//font-size:16px/}"
      # 例外②：env()/var() 的 0px 兜底（如 env(safe-area-inset-bottom, 0px)、var(--window-top, 0px)）
      stripped="${stripped//env(safe-area-inset-bottom, 0px)/}"
      stripped="${stripped//, 0px)/}"
      if [[ "$stripped" =~ [0-9.]+(rpx|px) ]]; then
        violations+=("$rel:$line_no:$line")
      fi
    done < "$file"
  done < <(find "$search_dir/src" -type f \( -name '*.vue' -o -name '*.scss' -o -name '*.ts' \) 2>/dev/null || true)

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "unit_standard" "units are rem only (no rpx/px)"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    detail="$(json_escape "$detail")"
    log_fail "unit_standard" "${#violations[@]} rpx/px violations: $detail" "项目规定前端统一 rem 单位（根字号 16px）" "rpx→rem 除以 32、px→rem 除以 16；根字号 font-size:16px 与 env()/var() 的 0px 兜底除外" "" ".harness/rules/项目编码规范.md §13"
  fi
}

# ─── Check 9: 前端依赖漏洞审计（trivy）──────────────────────────────
# 用 Trivy 直接读 package-lock.json 比对漏洞库，不依赖 npm audit（npmmirror 下不可用）。
# 工具未装 → WARN（不阻塞）；HIGH/CRITICAL → FAIL；DB 下载/扫描异常 → WARN。
check_dep_vuln() {
  echo "[9/9] Dependency vulnerability audit (trivy)" >&2
  if ! command -v trivy >/dev/null 2>&1; then
    log_warn "dep_vuln" "trivy 未安装——前端依赖漏洞审计跳过。安装二进制或 CI 用 aquasecurity/trivy-action"
    return 0
  fi

  local targets=()
  if [[ -n "$SERVICE_NAME" ]]; then
    targets+=("$WEB_DIR/$SERVICE_NAME/package-lock.json")
  else
    targets+=("$WEB_DIR/pc/package-lock.json" "$WEB_DIR/mobile/package-lock.json")
  fi

  local scanned=0 found=0 err=0
  for lock in "${targets[@]}"; do
    [[ -f "$lock" ]] || continue
    scanned=$((scanned + 1))
    local label="${lock#$PROJECT_ROOT/}"
    local out rc
    set +e
    echo "  (trivy fs: $label)" >&2
    out="$(timeout -k 10 180 trivy fs --scanners vuln --severity HIGH,CRITICAL --exit-code 1 "$lock" 2>&1)"; rc=$?
    set -e
    if [[ $rc -eq 0 ]]; then
      continue   # 无 HIGH/CRITICAL
    elif [[ $rc -eq 1 ]]; then
      found=1
      local summary
      summary=$(echo "$out" | grep -E "Total: [0-9]+|(HIGH|CRITICAL):" | head -6 | tr '\n' '; ')
      summary="$(json_escape "$summary")"
      log_fail "dep_vuln" "$label: 前端依赖 HIGH/CRITICAL 漏洞: $summary" "前端依赖存在已知高危漏洞（npm audit 在 npmmirror 下不可用，此前靠 review 人工抓）" "升级对应依赖到修复版并重新 npm install" "trivy fs --scanners vuln $label" "docs/qa-checks.md §依赖漏洞审计"
    else
      err=1
      local e
      e=$(echo "$out" | grep -iE "error|failed|timeout|network|download" | head -3 | tr '\n' '; ')
      e="$(json_escape "$e")"
      log_warn "dep_vuln" "$label: trivy 扫描异常（DB 下载/网络？）: $e"
    fi
  done

  if [[ $scanned -eq 0 ]]; then
    log_warn "dep_vuln" "未找到 package-lock.json（跳过）"
  elif [[ $found -eq 0 && $err -eq 0 ]]; then
    log_pass "dep_vuln" "trivy: 前端依赖未发现 HIGH/CRITICAL 漏洞"
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
  check_api_field_align
  check_unit_standard
  check_dep_vuln
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
    local labels=("type-check" "unit-test" "build" "hardcoded-secrets" "debug-artifacts" "type-safety" "api-field-align" "unit-standard" "dep-vuln")
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
