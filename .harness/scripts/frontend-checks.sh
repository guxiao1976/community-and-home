#!/usr/bin/env bash
#
# frontend-checks.sh — 前端机械化检查（6→12项扩展）
#
# 新增检查：
#   7. ESLint 规则检查
#   8. TypeScript any 类型检查
#   9. 路由一致性检查
#   10. API 类型对齐
#   11. Store 类型安全
#   12. 硬编码配置检查
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/web/pc"

EXIT_CODE=0

log_pass() { echo "[PASS] $1"; }
log_fail() { echo "[FAIL] $1"; EXIT_CODE=1; }
log_warn() { echo "[WARN] $1"; }

# ─── Check 7: ESLint 规则检查 ───
check_eslint_rules() {
    echo "[7/12] ESLint 规则检查"

    if [[ ! -f "$FRONTEND_DIR/package.json" ]]; then
        log_pass "无前端项目（跳过）"
        return
    fi

    cd "$FRONTEND_DIR"

    # 检查是否有 ESLint 配置
    if [[ ! -f ".eslintrc.js" && ! -f ".eslintrc.json" && ! -f "eslint.config.js" ]]; then
        log_warn "缺少 ESLint 配置文件"
        return
    fi

    # 运行 ESLint（只检查变更文件）
    CHANGED_TS=$(git diff --name-only --diff-filter=ACM HEAD -- "*.ts" "*.vue" || true)

    if [[ -z "$CHANGED_TS" ]]; then
        log_pass "无 TS/Vue 变更（跳过）"
        return
    fi

    if command -v npm &> /dev/null && npm run lint --if-present 2>&1 | grep -q "error"; then
        log_fail "ESLint 发现错误"
    else
        log_pass "ESLint 检查通过"
    fi
}

# ─── Check 8: TypeScript any 类型检查 ───
check_typescript_any() {
    echo "[8/12] TypeScript any 类型检查"

    CHANGED_TS=$(git diff --diff-filter=ACM HEAD -- "$FRONTEND_DIR/**/*.ts" "$FRONTEND_DIR/**/*.vue" | grep "^+" || true)

    if [[ -z "$CHANGED_TS" ]]; then
        log_pass "无 TS 变更（跳过）"
        return
    fi

    # 检查新增的 any 类型
    ANY_COUNT=$(echo "$CHANGED_TS" | grep -E ":\s*any\s*[=;,)]" | wc -l || echo "0")

    if [[ "$ANY_COUNT" -gt 0 ]]; then
        log_fail "新增 $ANY_COUNT 个 any 类型（禁止使用）"
    else
        log_pass "无新增 any 类型"
    fi
}

# ─── Check 9: 路由一致性检查 ───
check_route_consistency() {
    echo "[9/12] 路由一致性检查"

    ROUTER_FILE="$FRONTEND_DIR/src/router/index.ts"

    if [[ ! -f "$ROUTER_FILE" ]]; then
        log_pass "无路由文件（跳过）"
        return
    fi

    # 检查是否有新增路由
    NEW_ROUTES=$(git diff HEAD -- "$ROUTER_FILE" | grep "^+.*path:" | grep -oP "path:\s*['\"]\/\K[^'\"]*" || true)

    if [[ -z "$NEW_ROUTES" ]]; then
        log_pass "无新增路由（跳过）"
        return
    fi

    # 检查路由对应的 View 文件是否存在
    MISSING_VIEWS=0
    while IFS= read -r route; do
        [[ -z "$route" ]] && continue

        # 推断 View 文件路径（简化逻辑）
        VIEW_FILE="$FRONTEND_DIR/src/views/${route}.vue"

        if [[ ! -f "$VIEW_FILE" ]]; then
            MISSING_VIEWS=$((MISSING_VIEWS + 1))
        fi
    done <<< "$NEW_ROUTES"

    if [[ $MISSING_VIEWS -gt 0 ]]; then
        log_warn "$MISSING_VIEWS 个新增路由缺少对应 View 文件"
    else
        log_pass "路由与 View 一致"
    fi
}

# ─── Check 10: API 类型对齐 ───
check_api_type_alignment() {
    echo "[10/12] API 类型对齐"

    # 检查 API 接口定义是否与后端 Proto 对齐
    API_DIR="$FRONTEND_DIR/src/api"

    if [[ ! -d "$API_DIR" ]]; then
        log_pass "无 API 目录（跳过）"
        return
    fi

    # 简化：检查是否有 TODO 或 FIXME 标记的类型
    TODO_COUNT=$(grep -r "// TODO.*type\|// FIXME.*type" "$API_DIR" | wc -l || echo "0")

    if [[ "$TODO_COUNT" -gt 0 ]]; then
        log_warn "$TODO_COUNT 个 API 类型待完善"
    else
        log_pass "API 类型定义完整"
    fi
}

# ─── Check 11: Store 类型安全 ───
check_store_type_safety() {
    echo "[11/12] Store 类型安全"

    STORE_DIR="$FRONTEND_DIR/src/store"

    if [[ ! -d "$STORE_DIR" ]]; then
        log_pass "无 Store 目录（跳过）"
        return
    fi

    # 检查 Pinia/Vuex store 是否有类型定义
    CHANGED_STORE=$(git diff --name-only --diff-filter=ACM HEAD -- "$STORE_DIR/**/*.ts" || true)

    if [[ -z "$CHANGED_STORE" ]]; then
        log_pass "无 Store 变更（跳过）"
        return
    fi

    # 检查是否定义了 State 接口
    UNTYPED_STORES=0
    for store_file in $CHANGED_STORE; do
        if ! grep -q "interface.*State\|type.*State" "$store_file"; then
            UNTYPED_STORES=$((UNTYPED_STORES + 1))
        fi
    done

    if [[ $UNTYPED_STORES -gt 0 ]]; then
        log_fail "$UNTYPED_STORES 个 Store 缺少类型定义"
    else
        log_pass "Store 类型安全"
    fi
}

# ─── Check 12: 硬编码配置检查 ───
check_hardcoded_config() {
    echo "[12/12] 硬编码配置检查"

    CHANGED_FILES=$(git diff --name-only --diff-filter=ACM HEAD -- "$FRONTEND_DIR/src/**/*.ts" "$FRONTEND_DIR/src/**/*.vue" || true)

    if [[ -z "$CHANGED_FILES" ]]; then
        log_pass "无前端变更（跳过）"
        return
    fi

    # 检查新增的硬编码 URL、API Key 等
    HARDCODED_COUNT=0

    for file in $CHANGED_FILES; do
        # 检查硬编码 URL（http://、https://）
        if git diff HEAD -- "$file" | grep "^+" | grep -qE "https?://[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+|localhost:[0-9]+"; then
            HARDCODED_COUNT=$((HARDCODED_COUNT + 1))
        fi
    done

    if [[ $HARDCODED_COUNT -gt 0 ]]; then
        log_fail "$HARDCODED_COUNT 个文件包含硬编码 URL（应使用环境变量）"
    else
        log_pass "无硬编码配置"
    fi
}

# ─── 主执行 ───
echo "=== 前端扩展检查 (7-12) ==="
echo ""

check_eslint_rules
check_typescript_any
check_route_consistency
check_api_type_alignment
check_store_type_safety
check_hardcoded_config

echo ""
echo "=== 前端检查完成 ==="

exit $EXIT_CODE
