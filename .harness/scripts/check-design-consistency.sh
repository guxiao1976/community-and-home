#!/usr/bin/env bash
# ============================================================
# check-design-consistency.sh — 设计/代码一致性检查（确定性，非 LLM）
# ============================================================
# 第一级：比对 Go model 的 db tag 列 vs 标准迁移源（服务 migration/*.sql +
#         scripts/*.sql + 全局 docs/specs/migration.sql）的列覆盖。
#   报告 WARN：model 引用列但标准迁移源未覆盖（1054 风险 / 建表源不在标准位置）。
# 第二级：--all 全服务扫描，产出「设计一致性体检报告」，--backlog 自动建任务。
#
# 设计原则（克制）：
#   - WARN 而非 FAIL —— model 引用列缺失可能是历史手工迁移、legacy 结构或建表在别处，
#     WARN 提示风险 + 指向"建表源不在标准位置"，不误伤正常提交
#   - 联表 ViewModel 别名黑名单（ur_status/role_status）——JOIN 别名非真实表列
#   - 不强猜建表源 —— 只查"标准位置能定位到的"迁移源
#
# 用法:
#   bash .harness/scripts/check-design-consistency.sh --service <name>     # 单服务（harness-checks 调用）
#   bash .harness/scripts/check-design-consistency.sh --all                # 全服务体检
#   bash .harness/scripts/check-design-consistency.sh --all --backlog     # 全服务 + 自动建 backlog
#   bash .harness/scripts/check-design-consistency.sh --json               # JSON 输出
#
# 退出码: 0 = 无 WARN（或全 WARN 也 0，一致性检查非阻塞）
# ============================================================

set -uo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVICE_NAME=""
MODE="single"   # single | all
JSON_OUT=false
DO_BACKLOG=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) SERVICE_NAME="$2"; MODE="single"; shift 2 ;;
    --all) MODE="all"; shift ;;
    --backlog) DO_BACKLOG=true; shift ;;
    --json) JSON_OUT=true; shift ;;
    -h|--help) sed -n '1,30p' "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "未知参数: $1"; exit 2 ;;
  esac
done

# 联表 ViewModel 别名（JOIN 查询别名，非真实表列）——黑名单消除误报
IGNORE_COLS="ur_status role_status"

# ── 单服务检查：返回缺失列列表（空格分隔），echo 到 stdout ──
check_service() {
  local svc="$1"
  local model_dir="$PROJECT_ROOT/services/$svc/model"
  [[ -d "$model_dir" ]] || return 0   # 无 model 目录则跳过

  # 1. 收集 model 的 db tag 列（唯一，排除黑名单别名）
  local cols missing=""
  cols="$(grep -rhoP 'db:"[^"]+"' "$model_dir"/*.go 2>/dev/null | sed 's/db:"//;s/"//' | sort -u)"
  [[ -z "$cols" ]] && return 0

  # 2. 收集标准迁移源文本
  local sql_text=""
  sql_text="$(cat "$PROJECT_ROOT"/services/$svc/migration/*.sql \
                   "$PROJECT_ROOT"/services/$svc/scripts/*.sql \
                   "$PROJECT_ROOT"/docs/specs/migration.sql 2>/dev/null)"
  [[ -z "$sql_text" ]] && { echo "NO_SQL_SOURCE"; return 0; }

  # 3. 逐列检查覆盖（列名在 SQL 中可能带反引号包裹，如 `type`，需两种形式都匹配）
  #    无管道 + grep -q 直接查变量：管道 + grep -q 在 set -o pipefail 下会因 SIGPIPE 误判
  #    （grep -q 匹配即退出 → echo 写管道被断 → 管道返回非 0 → 误报缺失，实测列集每次不同）。
  #    grep -q 直接吃 heredoc 无管道，稳定。
  local c
  for c in $cols; do
    case " $IGNORE_COLS " in *" $c "*) continue;; esac
    # 匹配裸名 或 反引号包裹名（heredoc 无管道，grep -q 稳定）
    if ! grep -qE "(\`$c\`|$c)" <<< "$sql_text"; then
      missing="$missing $c"
    fi
  done
  echo "$missing"
}

# ── 输出 ──
declare -a findings=()    # "svc|col1 col2"

collect_all() {
  local svc
  for svc in $(ls "$PROJECT_ROOT"/services/ | grep -- '-service$'); do
    local miss
    miss="$(check_service "$svc")"
    case "$miss" in
      ""|NO_SQL_SOURCE) continue;;   # 全覆盖或无迁移源
      *) findings+=("$svc|$miss");;
    esac
  done
}

if [[ "$MODE" == "single" ]]; then
  # harness-checks 调用：单服务，WARN 输出
  miss=""
  miss="$(check_service "$SERVICE_NAME")" 2>/dev/null || true
  miss="${miss:-}"
  if [[ -n "$miss" && "$miss" != "NO_SQL_SOURCE" ]]; then
    if $JSON_OUT; then
      printf '{"check":"design_consistency","status":"WARN","service":"%s","detail":"model 列未覆盖标准迁移源","missing":"%s"}\n' "$SERVICE_NAME" "$(echo "$miss" | xargs)"
    else
      echo "WARN: $SERVICE_NAME model 引用列未覆盖标准迁移源:$miss"
    fi
  elif [[ "$miss" == "NO_SQL_SOURCE" ]]; then
    if $JSON_OUT; then
      printf '{"check":"design_consistency","status":"WARN","service":"%s","detail":"服务无标准迁移源（migration/scripts/docs-specs）——建表源不在标准位置"}\n' "$SERVICE_NAME"
    else
      echo "WARN: $SERVICE_NAME 无标准迁移源，建表源可能不在仓内权威文档"
    fi
  else
    if $JSON_OUT; then
      printf '{"check":"design_consistency","status":"PASS","service":"%s","detail":"model 列全部覆盖标准迁移源"}\n' "$SERVICE_NAME"
    else
      echo "PASS: $SERVICE_NAME model 列全部覆盖标准迁移源"
    fi
  fi
  exit 0
fi

# ── --all 全服务体检 ──
collect_all

if $JSON_OUT; then
  echo '['
  first=true
  for f in "${findings[@]}"; do
    $first || echo ','
    first=false
    IFS='|' read -r svc miss <<< "$f"
    printf '  {"service":"%s","missing_columns":"%s"}' "$svc" "$(echo "$miss" | xargs)"
  done
  echo ''
  echo ']'
  exit 0
fi

echo "=== 设计一致性体检报告（$PROJECT_ROOT/services/*） ==="
if [[ ${#findings[@]} -eq 0 ]]; then
  echo "✅ 全部服务 model 列均覆盖标准迁移源"
  exit 0
fi
echo "⚠️ 发现 ${#findings[@]} 个服务存在 model 列未覆盖标准迁移源："
for f in "${findings[@]}"; do
  IFS='|' read -r svc miss <<< "$f"
  echo "  [$svc] 缺失列: $(echo $miss | xargs)"
  echo "    → 可能: 历史手工迁移未入库 / legacy 结构 / 建表源不在标准位置。人工确认后补迁移或登记待办。"
done

# ── --backlog 自动建任务 ──
if $DO_BACKLOG && [[ ${#findings[@]} -gt 0 ]]; then
  echo ""
  echo "→ 生成 backlog 任务..."
  ts="$(date +%Y-%m-%d-%H%M%S 2>/dev/null || echo manual)"
  for f in "${findings[@]}"; do
    IFS='|' read -r svc miss <<< "$f"
    bash "$PROJECT_ROOT/.harness/scripts/harness-tasks.sh" create \
      --title "设计一致性: $svc model 列未覆盖标准迁移源" \
      --service "$svc" \
      --priority P2 \
      --type debt \
      --source human \
      --workload S \
      --detail "check-design-consistency.sh 体检发现 $svc 的 model db tag 列未覆盖标准迁移源（migration/scripts/docs-specs/migration.sql）：$(echo $miss | xargs)。可能为历史手工迁移未入库/legacy 结构/建表源不在标准位置，需人工确认后补迁移或登记豁免。" >/dev/null 2>&1 && echo "  已建任务: $svc"
  done
fi

exit 0
