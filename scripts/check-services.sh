#!/bin/bash

# 微服务状态检查脚本

PROJECT_ROOT="/home/jiaoxh/my-project/community-and-home"
cd "$PROJECT_ROOT"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  微服务状态检查"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 定义服务列表
declare -A services
services=(
    ["auth-service"]="8083"
    ["user-service"]="8084"
    ["file-service"]="8085"
    ["moderation-service"]="8086"
    ["master-data-service"]="8087"
    ["community-hub-service"]="8088"
)

# 检查进程
echo "📋 进程状态："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
for service in "${!services[@]}"; do
    if ps aux | grep -E "${service%%-*}" | grep -v grep > /dev/null 2>&1; then
        pid=$(ps aux | grep -E "${service%%-*}" | grep -v grep | head -1 | awk '{print $2}')
        echo "✅ $service (PID: $pid)"
    else
        echo "❌ $service (未运行)"
    fi
done

echo ""
echo "🔌 端口监听："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
for service in "${!services[@]}"; do
    port="${services[$service]}"
    if netstat -tln 2>/dev/null | grep ":$port " > /dev/null 2>&1; then
        echo "✅ $service :$port"
    else
        echo "❌ $service :$port (未监听)"
    fi
done

echo ""
echo "📊 总结："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
running=$(ps aux | grep -E "authservice|userservice|fileservice|masterdata|communityhub|moderation" | grep -v grep | wc -l)
echo "运行中的服务: $running / ${#services[@]}"

echo ""
echo "📝 日志目录: /tmp/microservices-logs/"
echo ""
