#!/bin/bash

# 启动所有 API 网关服务（REST API）
# 使用方法: bash scripts/start-all-apis.sh

set -e

PROJECT_ROOT="/home/jiaoxh/my-project/community-and-home"
cd "$PROJECT_ROOT"

# 加载环境变量
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  加载环境变量"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f .env ]; then
    set -a
    source .env
    set +a
    echo "✓ 环境变量已加载"
else
    echo "✗ 错误：.env 文件不存在"
    exit 1
fi

# 日志目录
LOG_DIR="/tmp/microservices-logs"
mkdir -p "$LOG_DIR"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  停止现有 API 服务"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

pkill -f "auth.*api" 2>/dev/null || true
pkill -f "user.*api" 2>/dev/null || true
pkill -f "file.*api" 2>/dev/null || true
pkill -f "communityhub.*api" 2>/dev/null || true
pkill -f "moderation.*api" 2>/dev/null || true
sleep 2
echo "✓ 已停止所有 API 服务"

# 启动函数
start_api() {
    local service_name=$1
    local service_dir=$2
    local api_bin=$3
    local config_file=$4
    local log_file="$LOG_DIR/${service_name}-api.log"

    echo ""
    echo "启动 $service_name API..."

    cd "$PROJECT_ROOT/$service_dir"

    # 确保环境变量传递给子进程
    nohup env MYSQL_USER="$MYSQL_USER" \
              MYSQL_PASSWORD="$MYSQL_PASSWORD" \
              REDIS_PASSWORD="$REDIS_PASSWORD" \
              go run "$api_bin" -f "$config_file" > "$log_file" 2>&1 &

    local pid=$!
    sleep 3

    # 检查是否启动成功
    if ps -p $pid > /dev/null 2>&1; then
        echo "✓ $service_name API 启动成功 (PID: $pid)"
        echo "  日志: $log_file"
    else
        echo "✗ $service_name API 启动失败"
        echo "  查看日志: tail -50 $log_file"
        return 1
    fi

    cd "$PROJECT_ROOT"
}

# 启动所有 API 服务
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  启动 API 网关服务"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 启动各服务的 API 层
start_api "auth-service" "services/auth-service/api" "auth.go" "etc/auth-api.yaml"
start_api "community-hub-service" "services/community-hub-service/api" "communityhub.go" "etc/communityhub-api.yaml"
start_api "file-service" "services/file-service/api" "file.go" "etc/file-api.yaml"
start_api "moderation-service" "services/moderation-service/api" "moderation.go" "etc/moderation-api.yaml"
start_api "master-data-service" "services/master-data-service/api" "masterdata.go" "etc/masterdata-api.yaml"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  API 服务端口"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

sleep 2

echo "等待端口监听..."
for port in 8881 8884 8886 8887 8889; do
    if netstat -tln 2>/dev/null | grep ":$port " > /dev/null; then
        echo "✓ 端口 $port 监听中"
    else
        echo "? 端口 $port 待确认"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  启动完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
