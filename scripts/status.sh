#!/bin/bash
# ============================================================
# 社区平台 — 查看服务状态
# 使用方式：bash scripts/status.sh
# ============================================================

echo "=== 服务进程 ==="
ps aux | grep -E "userservice|authservice|permissionservice|fileservice|identity|masterdata|moderation" | grep -v grep | awk '{printf "  %-40s  PID=%-8s\n", substr($0, index($0,$11)), $2}'

echo ""
echo "=== 端口监听 ==="
for port in 8082 8083 8084 8085 8881 8882 8884 8888 8889 8890; do
  name=""
  case $port in
    8082) name="user-service RPC" ;;
    8083) name="auth-service RPC" ;;
    8084) name="permission RPC" ;;
    8085) name="file-service RPC" ;;
    8881) name="auth-service API" ;;
    8882) name="user-service API" ;;
    8884) name="file-service API" ;;
    8888) name="identity API" ;;
    8889) name="master-data API" ;;
    8890) name="moderation API" ;;
  esac
  pid=$(lsof -ti:$port 2>/dev/null)
  if [ -n "$pid" ]; then
    echo "  $port ($name) : OK"
  else
    echo "  $port ($name) : --"
  fi
done

echo ""
echo "=== etcd 注册 ==="
docker exec etcd etcdctl get --prefix "" 2>/dev/null | grep -E "\.rpc/" | while read line; do
  if echo "$line" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+$'; then
    echo "  → $line"
  else
    echo -n "  $line"
  fi
done
