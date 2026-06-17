#!/bin/bash
# ============================================================
# 社区平台 — 停止所有服务
# 使用方式：bash scripts/stop.sh
# ============================================================

echo "=== 停止所有服务 ==="

# 按进程名匹配终止
PATTERNS=(
  "masterdata"
  "userservice"
  "authservice"
  "permissionservice"
  "fileservice"
  "aimodel"
  "moderation.go"
  "communityhub"
  "monitoring.go"
)

echo "  按进程名终止..."
for pat in "${PATTERNS[@]}"; do
  pids=$(pgrep -f "$pat" 2>/dev/null)
  if [ -n "$pids" ]; then
    kill $pids 2>/dev/null && echo "    ✅ $pat (pid: $pids)"
  fi
done

# 确保端口释放（RPC + API 全部端口）
echo ""
echo "  释放端口..."
ALL_PORTS=(
  8080 8082 8083 8084 8085 8086 8087 8088    # RPC
  8881 8882 8883 8884 8886 8887 8889 8890 8891  # API
)

for port in "${ALL_PORTS[@]}"; do
  pid=$(lsof -ti:$port 2>/dev/null)
  if [ -n "$pid" ]; then
    kill $pid 2>/dev/null && echo "    ✅ port $port released"
  fi
done

sleep 1
echo ""
echo "=== 全部停止 ==="
