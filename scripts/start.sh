#!/bin/bash
# ============================================================
# 社区平台 — 启动所有服务（按依赖顺序）
# 使用方式：bash scripts/start.sh
#           bash scripts/start.sh --rpc-only    # 仅 RPC
#           bash scripts/start.sh --api-only    # 仅 API（RPC 已运行）
# ============================================================
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# 加载共享环境变量
set -a
source .env
set +a

RPC_ONLY=false
API_ONLY=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --rpc-only) RPC_ONLY=true; shift ;;
    --api-only) API_ONLY=true; shift ;;
    *) shift ;;
  esac
done

# ─── RPC 服务（按依赖顺序：无依赖 → 有依赖）─────────────────────────

start_rpc() {
  echo "=== 启动 RPC 服务（依赖顺序）==="
  echo ""

  # Tier 0: 无依赖基础服务
  # master-data-service RPC (8087) — sysconfig 基础，所有服务依赖
  echo -n "  [1/8] master-data-service RPC (8087) ... "
  cd "$ROOT/services/master-data-service"
  go run rpc/masterdata.go -f rpc/etc/masterdata.yaml &>/tmp/masterdata-rpc.log &
  echo "pid=$!"

  # user-service RPC (8082) — 用户基础
  echo -n "  [2/8] user-service RPC (8082) ... "
  cd "$ROOT/services/user-service"
  go run rpc/userservice.go -f rpc/etc/userservice.yaml &>/tmp/user-rpc.log &
  echo "pid=$!"

  # permission-service RPC (8084)
  echo -n "  [3/8] permission-service RPC (8084) ... "
  cd "$ROOT/services/permission-service"
  go run rpc/permissionservice.go -f rpc/etc/permissionservice.yaml &>/tmp/perm-rpc.log &
  echo "pid=$!"

  # file-service RPC (8085)
  echo -n "  [4/8] file-service RPC (8085) ... "
  cd "$ROOT/services/file-service"
  go run rpc/fileservice.go -f rpc/etc/fileservice.yaml &>/tmp/file-rpc.log &
  echo "pid=$!"

  # ai-model-service RPC (8080)
  echo -n "  [5/8] ai-model-service RPC (8080) ... "
  cd "$ROOT/services/ai-model-service"
  go run rpc/aimodel.go -f rpc/etc/aimodel.yaml &>/tmp/aimodel-rpc.log &
  echo "pid=$!"

  # 等待基础服务注册到 etcd
  echo ""
  echo "  等待 Tier 0 服务注册..."
  sleep 8

  # Tier 1: 依赖上游服务
  # auth-service RPC (8083) — 依赖 user-service
  echo -n "  [6/8] auth-service RPC (8083) ... "
  cd "$ROOT/services/auth-service/rpc"
  go run authservice.go -f etc/authservice.yaml &>/tmp/auth-rpc.log &
  echo "pid=$!"

  # moderation-service RPC (8086) — 依赖 ai-model + master-data
  echo -n "  [7/8] moderation-service RPC (8086) ... "
  cd "$ROOT/services/moderation-service"
  go run rpc/moderation.go -f rpc/etc/moderation.yaml &>/tmp/moderation-rpc.log &
  echo "pid=$!"

  # community-hub-service RPC (8088) — 依赖 user + moderation
  echo -n "  [8/8] community-hub-service RPC (8088) ... "
  cd "$ROOT/services/community-hub-service"
  go run rpc/communityhub.go -f rpc/etc/communityhub.yaml &>/tmp/communityhub-rpc.log &
  echo "pid=$!"

  echo ""
  echo "  等待 Tier 1 服务注册..."
  sleep 6
  echo ""
  echo "  ✅ 全部 RPC 服务已启动"
}

# ─── API 服务 ─────────────────────────────────────────────────────────

start_api() {
  echo ""
  echo "=== 启动 API 服务 ==="
  echo ""

  echo -n "  user-service API (8882) ... "
  cd "$ROOT/services/user-service"
  go run api/user.go -f api/etc/user-api.yaml &>/tmp/user-api.log &
  echo "pid=$!"

  echo -n "  auth-service API (8881) ... "
  cd "$ROOT/services/auth-service/api"
  go run auth.go -f etc/auth-api.yaml &>/tmp/auth-api.log &
  echo "pid=$!"

  echo -n "  permission-service API (8883) ... "
  cd "$ROOT/services/permission-service"
  go run api/perm.go -f api/etc/perm-api.yaml &>/tmp/perm-api.log &
  echo "pid=$!"

  echo -n "  file-service API (8884) ... "
  cd "$ROOT/services/file-service"
  go run api/file.go -f api/etc/file-api.yaml &>/tmp/file-api.log &
  echo "pid=$!"

  echo -n "  master-data-service API (8889) ... "
  cd "$ROOT/services/master-data-service"
  go run api/masterdata.go -f api/etc/masterdata-api.yaml &>/tmp/masterdata-api.log &
  echo "pid=$!"

  echo -n "  moderation-service API (8890) ... "
  cd "$ROOT/services/moderation-service"
  go run api/moderation.go -f api/etc/moderation-api.yaml &>/tmp/moderation-api.log &
  echo "pid=$!"

  echo -n "  community-hub-service API (8887) ... "
  cd "$ROOT/services/community-hub-service"
  go run api/communityhub.go -f api/etc/communityhub-api.yaml &>/tmp/communityhub-api.log &
  echo "pid=$!"

  echo -n "  ai-model-service API (8891) ... "
  cd "$ROOT/services/ai-model-service"
  go run api/aimodelapi.go -f api/etc/aimodelapi.yaml &>/tmp/aimodel-api.log &
  echo "pid=$!"

  echo -n "  monitoring-service API (8886) ... "
  cd "$ROOT/services/monitoring-service"
  go run api/monitoring.go -f api/etc/monitoring-api.yaml &>/tmp/monitoring-api.log &
  echo "pid=$!"

  sleep 3
  echo ""
  echo "  ✅ 全部 API 服务已启动"
}

# ─── Main ─────────────────────────────────────────────────────────────

cd "$ROOT"

if $API_ONLY; then
  start_api
else
  start_rpc
  if ! $RPC_ONLY; then
    start_api
  fi
fi

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║  全部服务启动完成                                  ║"
echo "╠══════════════════════════════════════════════════╣"
echo "║  RPC:                                             ║"
echo "║    master-data :8087   ai-model    :8080          ║"
echo "║    user        :8082   auth        :8083          ║"
echo "║    permission  :8084   file        :8085          ║"
echo "║    moderation  :8086   community   :8088          ║"
echo "║                                                   ║"
echo "║  API:                                             ║"
echo "║    auth-api    :8881   user-api    :8882          ║"
echo "║    perm-api    :8883   file-api    :8884          ║"
echo "║    monitor-api :8886   hub-api     :8887          ║"
echo "║    master-api  :8889   mod-api     :8890          ║"
echo "║    ai-api      :8891                              ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""
echo "查看状态: bash scripts/status.sh"
echo "冒烟测试: bash .harness/scripts/harness-smoke.sh"
echo "查看日志: tail -f /tmp/*-rpc.log /tmp/*-api.log"
