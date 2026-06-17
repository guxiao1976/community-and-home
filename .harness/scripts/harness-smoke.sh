#!/usr/bin/env bash
#
# harness-smoke.sh — Runtime smoke test for gRPC service connectivity
#
# 按三层递进验证：
#   L1: 进程存活 — 检查端口是否在监听
#   L2: gRPC 连通性 — grpcurl 调用每个服务的最简单 RPC
#   L3: 依赖链验证 — 验证跨服务 gRPC 调用链
#
# 用法:
#   bash .harness/scripts/harness-smoke.sh                    # 全量 smoke
#   bash .harness/scripts/harness-smoke.sh --service <name>   # 单服务
#   bash .harness/scripts/harness-smoke.sh --json             # JSON 输出
#   bash .harness/scripts/harness-smoke.sh --quick            # 仅 L1+L2（跳过 L3）
#
# 返回码:
#   0 — 全部通过
#   1 — ≥1 项 FAIL
#   2 — 无服务在运行（SKIP）

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
GRPCURL="${GRPCURL:-$(which grpcurl 2>/dev/null || echo '')}"
SERVICE_NAME=""
OUTPUT_JSON=false
QUICK_MODE=false
EXIT_CODE=0

# Counters
L1_PASS=0 L1_FAIL=0 L1_SKIP=0
L2_PASS=0 L2_FAIL=0 L2_SKIP=0
L3_PASS=0 L3_FAIL=0 L3_SKIP=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) SERVICE_NAME="$2"; shift 2 ;;
    --json) OUTPUT_JSON=true; shift ;;
    --quick) QUICK_MODE=true; shift ;;
    *) shift ;;
  esac
done

timestamp() { date -u +"%Y-%m-%dT%H:%M:%SZ"; }

# ─── Service smoke matrix ─────────────────────────────────────────────
# Format: "service_label|port|etcd_key|rpc_service|rpc_method|request_json|depends_on"

SMOKE_MATRIX=(
  "moderation-service|8086|moderation.rpc|moderation.v1.ModerationService|HealthCheck|{}|"
  "ai-model-service|8080|aimodel.rpc|aimodel.v1.AiModelService|HealthCheck|{}|"
  "user-service|8082|user.rpc|user.v1.UserService|ListUsers|{\"page\":{\"page\":1,\"page_size\":1}}|"
  "auth-service|8083|auth.rpc|grpc.reflection.v1.ServerReflection|ServerReflectionInfo|{}|"
  "permission-service|8084|permission.rpc|permission.v1.PermissionService|ListPermissions|{}|"
  "file-service|8085|file.rpc|file.v1.FileService|ListFiles|{\"page\":{\"page\":1,\"page_size\":1}}|"
  "master-data-service|8087|masterdata.rpc|masterdata.v1.MasterdataService|GetConfig|{\"config_key\":\"site_name\"}|"
  "community-hub-service|8088|communityhub.rpc|community.v1.ContactService|ListContacts|{\"community_id\":\"0\"}|"
)

# ─── Helpers ──────────────────────────────────────────────────────────

resolve_svc() {
  for entry in "${SMOKE_MATRIX[@]}"; do
    local svc=$(echo "$entry" | cut -d'|' -f1)
    [[ "$svc" == "$1" ]] && { echo "$entry"; return 0; }
  done
  return 1
}

check_port() {
  local port="$1"
  ss -tlnp 2>/dev/null | grep -q ":${port} " && return 0
  netstat -tlnp 2>/dev/null | grep -q ":${port} " && return 0
  return 1
}

grpcurl_call() {
  local port="$1" service="$2" method="$3" data="$4"
  local addr="127.0.0.1:${port}"
  if [[ -z "$GRPCURL" ]]; then
    echo "SKIP:grpcurl not found"
    return 2
  fi
  # auth-service uses reflection, others use direct method call
  if [[ "$service" == "grpc.reflection.v1.ServerReflection" ]]; then
    local ref_out
    ref_out=$("$GRPCURL" -plaintext -max-time 5 "$addr" list 2>&1) || true
    if echo "$ref_out" | grep -q 'grpc\.reflection\|ServerReflection\|user\.\|auth\.\|permission\.\|file\.\|moderation\.\|masterdata\.\|aimodel\.\|community\.' ; then
      echo "PASS:reflection OK ($(echo "$ref_out" | wc -l) services)"
      return 0
    fi
    echo "FAIL:reflection returned unexpected: ${ref_out:0:100}"
    return 1
  fi
  local output
  # Capture both stdout and stderr; grpcurl writes errors to stderr
  output=$("$GRPCURL" -plaintext -max-time 5 -d "$data" "$addr" "$service/$method" 2>&1) || true
  local exit_code=$?

  # grpcurl exits 0 even when RPC returns gRPC error. Check for error indicators.
  if echo "$output" | grep -qiE 'Error invoking method|rpc error|Unavailable|Unimplemented|NotFound|Internal|DeadlineExceeded|connection refused|failed to query'; then
    echo "FAIL:${output:0:120}"
    return 1
  fi

  if [[ $exit_code -eq 0 ]]; then
    # Successfully invoked — extract key info
    local summary=$(echo "$output" | head -1)
    echo "PASS:${summary:0:80}"
    return 0
  else
    echo "FAIL:${output:0:120}"
    return 1
  fi
}

# ─── L1: Process aliveness ────────────────────────────────────────────

smoke_l1() {
  local label="$1" port="$2"
  if check_port "$port"; then
    echo "  ✅ L1 端口 ${port} 在监听"
    L1_PASS=$((L1_PASS + 1))
    return 0
  else
    echo "  ❌ L1 端口 ${port} 未监听 — 服务未启动"
    L1_FAIL=$((L1_FAIL + 1))
    return 1
  fi
}

# ─── L2: gRPC connectivity ────────────────────────────────────────────

smoke_l2() {
  local label="$1" port="$2" svc="$3" method="$4" data="$5"
  local result exit_code
  set +e
  result=$(grpcurl_call "$port" "$svc" "$method" "$data" 2>&1)
  exit_code=$?
  set -e
  local status="${result%%:*}"
  local detail="${result#*:}"

  if [[ $exit_code -eq 0 ]]; then
    echo "  ✅ L2 ${svc}/${method} → ${detail:0:80}"
    L2_PASS=$((L2_PASS + 1))
    return 0
  elif [[ $exit_code -eq 2 ]]; then
    echo "  ⏭️  L2 SKIP — grpcurl 不可用"
    L2_SKIP=$((L2_SKIP + 1))
    return 2
  else
    echo "  ❌ L2 ${svc}/${method} → ${detail:0:120}"
    L2_FAIL=$((L2_FAIL + 1))
    return 1
  fi
}

# ─── L3: Dependency chain verification ────────────────────────────────

smoke_l3() {
  local label="$1" deps="$2"
  [[ -z "$deps" ]] && { L3_SKIP=$((L3_SKIP + 1)); return 0; }

  # Parse dependency list: "svcA:portA,svcB:portB"
  local all_ok=true
  IFS=',' read -ra DEP_ARR <<< "$deps"
  for dep in "${DEP_ARR[@]}"; do
    [[ -z "$dep" ]] && continue
    local dep_svc="${dep%%:*}"
    local dep_port="${dep##*:}"
    echo -n "  L3 ${label} → ${dep_svc} (port ${dep_port}): "
    if check_port "$dep_port"; then
      echo "✅ 可达"
    else
      echo "❌ 不可达"
      all_ok=false
    fi
  done

  if $all_ok; then
    L3_PASS=$((L3_PASS + 1))
    return 0
  else
    L3_FAIL=$((L3_FAIL + 1))
    return 1
  fi
}

# ─── Main smoke loop ──────────────────────────────────────────────────

run_smoke() {
  local any_running=false

  for entry in "${SMOKE_MATRIX[@]}"; do
    IFS='|' read -r label port etcd_key svc method data deps <<< "$entry"

    # Filter by service if specified
    [[ -n "$SERVICE_NAME" && "$label" != "$SERVICE_NAME" ]] && continue

    echo ""
    echo "━━━ ${label} (port ${port}) ━━━"

    # L1
    if smoke_l1 "$label" "$port"; then
      any_running=true

      # L2
      smoke_l2 "$label" "$port" "$svc" "$method" "$data"

      # L3 (skip in quick mode)
      if ! $QUICK_MODE; then
        smoke_l3 "$label" "$deps"
      else
        L3_SKIP=$((L3_SKIP + 1))
      fi
    else
      L2_SKIP=$((L2_SKIP + 1))
      L3_SKIP=$((L3_SKIP + 1))
    fi
  done

  if ! $any_running; then
    echo ""
    echo "❌ 无服务在运行。请先启动服务:"
    echo "   docker compose up -d          # 中间件"
    echo "   bash scripts/start.sh         # 服务"
    return 2
  fi

  return 0
}

# ─── Output ────────────────────────────────────────────────────────────

print_results() {
  local total_l1=$((L1_PASS + L1_FAIL + L1_SKIP))
  local total_l2=$((L2_PASS + L2_FAIL + L2_SKIP))
  local total_l3=$((L3_PASS + L3_FAIL + L3_SKIP))
  local all_pass=$((L1_PASS + L2_PASS + L3_PASS))
  local all_fail=$((L1_FAIL + L2_FAIL + L3_FAIL))

  echo ""
  echo "╔══════════════════════════════════════╗"
  echo "║   Smoke Test Results                 ║"
  echo "╚══════════════════════════════════════╝"
  echo ""
  echo "   L1 进程存活: ${L1_PASS}✅ ${L1_FAIL}❌ ${L1_SKIP}⏭️"
  echo "   L2 gRPC连通: ${L2_PASS}✅ ${L2_FAIL}❌ ${L2_SKIP}⏭️"
  echo "   L3 依赖链:   ${L3_PASS}✅ ${L3_FAIL}❌ ${L3_SKIP}⏭️"
  echo ""
  echo "   总计: ${all_pass} PASS / ${all_fail} FAIL"

  if [[ $all_fail -gt 0 ]]; then
    echo ""
    echo "⚠️  存在 ${all_fail} 项失败。建议："
    echo "   1. 确认 docker compose up -d 已启动"
    echo "   2. 确认服务已启动 (bash scripts/start.sh)"
    echo "   3. 检查 etcd 服务发现是否正常"
    EXIT_CODE=1
  fi
}

print_json() {
  cat <<JSONEOF
{
  "timestamp": "$(timestamp)",
  "service": "${SERVICE_NAME:-all}",
  "layers": {
    "L1_process": {"pass": $L1_PASS, "fail": $L1_FAIL, "skip": $L1_SKIP},
    "L2_grpc":   {"pass": $L2_PASS, "fail": $L2_FAIL, "skip": $L2_SKIP},
    "L3_deps":   {"pass": $L3_PASS, "fail": $L3_FAIL, "skip": $L3_SKIP}
  },
  "verdict": "$([ $((L1_FAIL + L2_FAIL + L3_FAIL)) -eq 0 ] && echo "PASS" || echo "FAIL")"
}
JSONEOF
}

# ─── Main ──────────────────────────────────────────────────────────────

# Pre-flight: check grpcurl
if [[ -z "$GRPCURL" ]]; then
  echo "⚠️  grpcurl 未找到。L2 gRPC 连通性测试将跳过。"
  echo "   安装: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi

run_smoke || EXIT_CODE=$?

if $OUTPUT_JSON; then
  print_json
else
  print_results
fi

exit $EXIT_CODE
