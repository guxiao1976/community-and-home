#!/bin/bash

# 社区管理系统服务启动脚本
# 用法: ./start-services.sh [all|api|rpc|frontend|stop]

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="/tmp/community-logs"

mkdir -p "$LOG_DIR"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查服务是否运行
check_service() {
    local port=$1
    local name=$2
    if ss -tlnp 2>/dev/null | grep -q ":$port "; then
        echo -e "${GREEN}✓${NC} $name (端口 $port) 正在运行"
        return 0
    else
        echo -e "${RED}✗${NC} $name (端口 $port) 未运行"
        return 1
    fi
}

# 启动服务
start_service() {
    local service_dir=$1
    local config_file=$2
    local service_name=$3
    local log_file="$LOG_DIR/$service_name.log"

    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}错误: 目录不存在 $service_dir${NC}"
        return 1
    fi

    cd "$service_dir"

    # 查找可执行文件
    local exe=""
    if [ -f "./api" ]; then
        exe="./api"
    elif [ -f "./$service_name" ]; then
        exe="./$service_name"
    elif [ -f "./${service_name}-api" ]; then
        exe="./${service_name}-api"
    elif [ -f "./${service_name}-rpc" ]; then
        exe="./${service_name}-rpc"
    else
        echo -e "${RED}错误: 找不到可执行文件在 $service_dir${NC}"
        return 1
    fi

    echo -e "${YELLOW}启动${NC} $service_name..."
    nohup $exe -f "$config_file" > "$log_file" 2>&1 &
    sleep 1

    if tail -5 "$log_file" | grep -q "Starting server\|已启动\|ListenOn"; then
        echo -e "${GREEN}✓${NC} $service_name 启动成功，日志: $log_file"
    else
        echo -e "${RED}✗${NC} $service_name 启动可能失败，检查日志: $log_file"
    fi
}

# 停止服务
stop_service() {
    local port=$1
    local name=$2

    local pid=$(ss -tlnp 2>/dev/null | grep ":$port " | grep -oP 'pid=\K[0-9]+' | head -1)
    if [ -n "$pid" ]; then
        echo -e "${YELLOW}停止${NC} $name (PID: $pid)..."
        kill $pid
        sleep 1
        if kill -0 $pid 2>/dev/null; then
            kill -9 $pid
        fi
        echo -e "${GREEN}✓${NC} $name 已停止"
    else
        echo -e "${YELLOW}!${NC} $name 未运行"
    fi
}

# 启动所有 API 服务
start_api_services() {
    echo "=== 启动 API 服务 ==="
    start_service "$BASE_DIR/services/identity/api" "etc/identity-api.yaml" "identity-api"
    start_service "$BASE_DIR/services/masterdata/api" "etc/masterdata-api.yaml" "masterdata-api"
    start_service "$BASE_DIR/services/moderation/api" "etc/moderation-api.yaml" "moderation-api"
}

# 启动所有 RPC 服务
start_rpc_services() {
    echo "=== 启动 RPC 服务 ==="
    start_service "$BASE_DIR/services/identity/rpc" "etc/identity.yaml" "identity-rpc"
    start_service "$BASE_DIR/services/masterdata/rpc" "etc/masterdata.yaml" "masterdata-rpc"
    # ai-model-rpc 通常由 root 启动，跳过
}

# 启动前端
start_frontend() {
    echo "=== 启动前端开发服务器 ==="
    cd "$BASE_DIR/web/pc"

    # 停止现有的 vite 进程
    pkill -f "vite --host" 2>/dev/null
    sleep 1

    # 使用 NO_PROXY 启动，避免代理问题
    NO_PROXY='*' no_proxy='*' nohup npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
    sleep 3

    if tail -10 "$LOG_DIR/frontend.log" | grep -q "Local:"; then
        echo -e "${GREEN}✓${NC} 前端服务启动成功"
        tail -10 "$LOG_DIR/frontend.log" | grep -E "Local:|Network:"
    else
        echo -e "${RED}✗${NC} 前端服务启动失败，检查日志: $LOG_DIR/frontend.log"
    fi
}

# 停止所有服务
stop_all() {
    echo "=== 停止所有服务 ==="
    stop_service 8888 "identity-api"
    stop_service 8889 "masterdata-api"
    stop_service 8890 "moderation-api"
    stop_service 8081 "identity-rpc"
    stop_service 8083 "masterdata-rpc"

    echo "停止前端服务..."
    pkill -f "vite --host" 2>/dev/null && echo -e "${GREEN}✓${NC} 前端服务已停止"
}

# 检查所有服务状态
check_all() {
    echo "=== 服务状态检查 ==="
    echo ""
    echo "API 服务:"
    check_service 8888 "identity-api"
    check_service 8889 "masterdata-api"
    check_service 8890 "moderation-api"
    echo ""
    echo "RPC 服务:"
    check_service 8081 "identity-rpc"
    check_service 8083 "masterdata-rpc"
    check_service 8084 "ai-model-rpc"
    echo ""
    echo "前端服务:"
    check_service 3000 "vite-dev-server"
    echo ""
    echo "依赖服务:"
    check_service 6379 "redis"
    check_service 3306 "mysql"
}

# 主逻辑
case "${1:-all}" in
    all)
        start_rpc_services
        sleep 2
        start_api_services
        sleep 2
        start_frontend
        echo ""
        check_all
        ;;
    api)
        start_api_services
        ;;
    rpc)
        start_rpc_services
        ;;
    frontend)
        start_frontend
        ;;
    stop)
        stop_all
        ;;
    status)
        check_all
        ;;
    *)
        echo "用法: $0 [all|api|rpc|frontend|stop|status]"
        echo ""
        echo "  all      - 启动所有服务（默认）"
        echo "  api      - 只启动 API 服务"
        echo "  rpc      - 只启动 RPC 服务"
        echo "  frontend - 只启动前端服务"
        echo "  stop     - 停止所有服务"
        echo "  status   - 检查服务状态"
        exit 1
        ;;
esac
