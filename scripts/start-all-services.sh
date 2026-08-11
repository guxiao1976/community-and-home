#!/bin/bash

# 启动所有微服务
# 使用方法: bash scripts/start-all-services.sh

set -e

PROJECT_ROOT="/home/jiaoxh/my-project/community-and-home"
cd "$PROJECT_ROOT"

# 加载环境变量
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "✓ 环境变量已加载"
else
    echo "✗ 错误：.env 文件不存在"
    exit 1
fi

# 日志目录
LOG_DIR="/tmp/microservices-logs"
mkdir -p "$LOG_DIR"

# 检查基础设施
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  检查基础设施服务"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

docker compose ps | grep -E "mysql|redis|etcd" | grep -E "Up|running" > /dev/null
if [ $? -eq 0 ]; then
    echo "✓ MySQL, Redis, etcd 运行正常"
else
    echo "✗ 错误：基础设施服务未运行"
    echo "  请先执行: docker compose up -d"
    exit 1
fi

# 停止已有服务
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  停止现有服务"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

pkill -f "authservice" 2>/dev/null || true
pkill -f "userservice" 2>/dev/null || true
pkill -f "masterdata" 2>/dev/null || true
pkill -f "fileservice" 2>/dev/null || true
pkill -f "communityhub" 2>/dev/null || true
pkill -f "moderation" 2>/dev/null || true
sleep 2
echo "✓ 已停止所有服务"

# 启动函数
start_service() {
    local service_name=$1
    local service_dir=$2
    local service_bin=$3
    local config_file=$4
    local log_file="$LOG_DIR/${service_name}.log"

    echo ""
    echo "启动 $service_name..."

    cd "$PROJECT_ROOT/$service_dir"
    nohup go run "$service_bin" -f "$config_file" > "$log_file" 2>&1 &
    local pid=$!

    # 等待启动
    sleep 2

    # 检查是否启动成功
    if ps -p $pid > /dev/null 2>&1; then
        echo "✓ $service_name 启动成功 (PID: $pid)"
        echo "  日志: $log_file"
    else
        echo "✗ $service_name 启动失败"
        echo "  查看日志: tail -50 $log_file"
        return 1
    fi

    cd "$PROJECT_ROOT"
}

# 启动所有服务
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  启动微服务"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 1. 基础服务（无依赖）
start_service "auth-service" "services/auth-service/rpc" "authservice.go" "etc/authservice.yaml"
start_service "user-service" "services/user-service/rpc" "userservice.go" "etc/userservice.yaml"
start_service "master-data-service" "services/master-data-service/rpc" "masterdata.go" "etc/masterdata.yaml"
start_service "file-service" "services/file-service/rpc" "fileservice.go" "etc/fileservice.yaml"

# 等待基础服务完全启动并注册到 etcd
echo ""
echo "等待服务注册到 etcd..."
sleep 3

# 2. 依赖其他服务的服务
start_service "community-hub-service" "services/community-hub-service/rpc" "communityhub.go" "etc/communityhub.yaml"
start_service "moderation-service" "services/moderation-service/rpc" "moderation.go" "etc/moderation.yaml"

# 服务状态总结
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  服务状态"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

sleep 2

ps aux | grep -E "authservice|userservice|masterdata|fileservice|communityhub|moderation" | grep -v grep | awk '{print "✓ " $11 " (PID: " $2 ")"}'

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  端口监听"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

netstat -tlnp 2>/dev/null | grep -E ":(8083|8084|8085|8086|8087|8088)" | awk '{print $4 "\t" $7}' || echo "  (需要 root 权限查看进程信息)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  启动完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "查看日志:"
echo "  tail -f $LOG_DIR/<service-name>.log"
echo ""
echo "查看所有服务状态:"
echo "  ps aux | grep -E 'authservice|userservice|masterdata|fileservice|communityhub|moderation' | grep -v grep"
echo ""
