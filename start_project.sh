#!/bin/bash

# 社区管理系统 - 启动脚本
# 启动所有前后端服务

set -e

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="/tmp/community-logs"

mkdir -p "$LOG_DIR"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  社区管理系统 - 启动所有服务${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 服务配置
declare -A SERVICES=(
    # RPC 服务
    ["identity-rpc"]="8081|services/identity/rpc|etc/identity.yaml|identity-rpc"
    ["masterdata-rpc"]="8083|services/masterdata/rpc|etc/masterdata.yaml|masterdata-rpc"

    # API 服务
    ["identity-api"]="8888|services/identity/api|etc/identity-api.yaml|identity-api"
    ["masterdata-api"]="8889|services/masterdata/api|etc/masterdata-api.yaml|api"
    ["moderation-api"]="8890|services/moderation/api|etc/moderation-api.yaml|moderation-api"
)

# 前端服务配置
FRONTEND_PORT=3000
FRONTEND_DIR="web/pc"

# 检查端口是否被占用
check_port() {
    local port=$1
    ss -tlnp 2>/dev/null | grep -q ":$port " && return 0 || return 1
}

# 根据端口查找并停止进程
stop_by_port() {
    local port=$1
    local name=$2

    local pids=$(ss -tlnp 2>/dev/null | grep ":$port " | grep -oP 'pid=\K[0-9]+' | sort -u)

    if [ -n "$pids" ]; then
        echo -e "${YELLOW}  发现已运行的 $name (端口 $port)，正在停止...${NC}"
        for pid in $pids; do
            kill $pid 2>/dev/null || true
            sleep 0.5
            if kill -0 $pid 2>/dev/null; then
                kill -9 $pid 2>/dev/null || true
            fi
        done
        sleep 1
        echo -e "${GREEN}  ✓ 已停止旧的 $name 进程${NC}"
        return 0
    fi
    return 1
}

# 根据进程名停止进程
stop_by_name() {
    local pattern=$1
    local name=$2

    local pids=$(pgrep -f "$pattern" 2>/dev/null || true)

    if [ -n "$pids" ]; then
        echo -e "${YELLOW}  发现已运行的 $name，正在停止...${NC}"
        for pid in $pids; do
            kill $pid 2>/dev/null || true
            sleep 0.5
            if kill -0 $pid 2>/dev/null; then
                kill -9 $pid 2>/dev/null || true
            fi
        done
        sleep 1
        echo -e "${GREEN}  ✓ 已停止旧的 $name 进程${NC}"
        return 0
    fi
    return 1
}

# 启动后端服务
start_backend_service() {
    local name=$1
    local config=$2

    IFS='|' read -r port dir config_file exe_name <<< "$config"

    local service_dir="$BASE_DIR/$dir"
    local log_file="$LOG_DIR/$name.log"

    echo -e "${BLUE}[启动] $name (端口 $port)${NC}"

    # 检查目录
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}  ✗ 错误: 目录不存在 $service_dir${NC}"
        return 1
    fi

    # 停止已存在的服务
    stop_by_port "$port" "$name"

    # 查找可执行文件
    cd "$service_dir"
    local exe=""

    if [ -f "./$exe_name" ]; then
        exe="./$exe_name"
    elif [ -f "./api" ]; then
        exe="./api"
    elif [ -f "./rpc" ]; then
        exe="./rpc"
    else
        echo -e "${RED}  ✗ 错误: 找不到可执行文件 $exe_name${NC}"
        return 1
    fi

    # 启动服务
    nohup $exe -f "$config_file" > "$log_file" 2>&1 &
    local pid=$!

    # 等待启动
    sleep 2

    # 检查进程是否存活
    if ! kill -0 $pid 2>/dev/null; then
        echo -e "${RED}  ✗ 启动失败，进程已退出${NC}"
        echo -e "${YELLOW}  最后 10 行日志:${NC}"
        tail -10 "$log_file" | sed 's/^/    /'
        return 1
    fi

    # 检查端口是否监听
    local retry=0
    while [ $retry -lt 10 ]; do
        if check_port "$port"; then
            echo -e "${GREEN}  ✓ 启动成功 (PID: $pid, 日志: $log_file)${NC}"
            return 0
        fi
        sleep 0.5
        retry=$((retry + 1))
    done

    echo -e "${RED}  ✗ 启动失败，端口未监听${NC}"
    echo -e "${YELLOW}  最后 10 行日志:${NC}"
    tail -10 "$log_file" | sed 's/^/    /'
    return 1
}

# 启动前端服务
start_frontend() {
    echo -e "${BLUE}[启动] 前端开发服务器 (端口 $FRONTEND_PORT)${NC}"

    local frontend_dir="$BASE_DIR/$FRONTEND_DIR"
    local log_file="$LOG_DIR/frontend.log"

    # 检查目录
    if [ ! -d "$frontend_dir" ]; then
        echo -e "${RED}  ✗ 错误: 目录不存在 $frontend_dir${NC}"
        return 1
    fi

    # 停止已存在的 vite 进程
    stop_by_name "vite --host" "前端服务"
    stop_by_port "$FRONTEND_PORT" "前端服务"

    # 启动前端
    cd "$frontend_dir"
    NO_PROXY='*' no_proxy='*' nohup npm run dev > "$log_file" 2>&1 &
    local pid=$!

    # 等待启动
    echo -e "${YELLOW}  等待前端服务启动...${NC}"
    sleep 5

    # 检查进程是否存活
    if ! kill -0 $pid 2>/dev/null; then
        echo -e "${RED}  ✗ 启动失败，进程已退出${NC}"
        echo -e "${YELLOW}  最后 20 行日志:${NC}"
        tail -20 "$log_file" | sed 's/^/    /'
        return 1
    fi

    # 检查端口是否监听
    local retry=0
    while [ $retry -lt 20 ]; do
        if check_port "$FRONTEND_PORT"; then
            echo -e "${GREEN}  ✓ 启动成功 (PID: $pid, 日志: $log_file)${NC}"

            # 显示访问地址
            local ip=$(hostname -I | awk '{print $1}')
            echo -e "${GREEN}  访问地址:${NC}"
            echo -e "    http://localhost:$FRONTEND_PORT/"
            echo -e "    http://$ip:$FRONTEND_PORT/"
            return 0
        fi
        sleep 0.5
        retry=$((retry + 1))
    done

    echo -e "${RED}  ✗ 启动失败，端口未监听${NC}"
    echo -e "${YELLOW}  最后 20 行日志:${NC}"
    tail -20 "$log_file" | sed 's/^/    /'
    return 1
}

# 检查服务状态
check_service_status() {
    local name=$1
    local port=$2

    if check_port "$port"; then
        echo -e "${GREEN}✓${NC} $name (端口 $port)"
        return 0
    else
        echo -e "${RED}✗${NC} $name (端口 $port)"
        return 1
    fi
}

# 主流程
main() {
    local failed_services=()

    echo -e "${YELLOW}步骤 1/3: 启动 RPC 服务${NC}"
    echo ""

    # 启动 RPC 服务
    for service in "identity-rpc" "masterdata-rpc"; do
        if ! start_backend_service "$service" "${SERVICES[$service]}"; then
            failed_services+=("$service")
        fi
        echo ""
    done

    # RPC 服务启动后等待一下
    sleep 2

    echo -e "${YELLOW}步骤 2/3: 启动 API 服务${NC}"
    echo ""

    # 启动 API 服务
    for service in "identity-api" "masterdata-api" "moderation-api"; do
        if ! start_backend_service "$service" "${SERVICES[$service]}"; then
            failed_services+=("$service")
        fi
        echo ""
    done

    # API 服务启动后等待一下
    sleep 2

    echo -e "${YELLOW}步骤 3/3: 启动前端服务${NC}"
    echo ""

    # 启动前端
    if ! start_frontend; then
        failed_services+=("frontend")
    fi
    echo ""

    # 显示最终状态
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  服务状态总览${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    echo "RPC 服务:"
    check_service_status "identity-rpc" "8081"
    check_service_status "masterdata-rpc" "8083"
    check_service_status "ai-model-rpc" "8084"
    echo ""

    echo "API 服务:"
    check_service_status "identity-api" "8888"
    check_service_status "masterdata-api" "8889"
    check_service_status "moderation-api" "8890"
    echo ""

    echo "前端服务:"
    check_service_status "vite-dev-server" "3000"
    echo ""

    echo "依赖服务:"
    check_service_status "redis" "6379"
    check_service_status "mysql" "3306"
    echo ""

    # 显示失败的服务
    if [ ${#failed_services[@]} -gt 0 ]; then
        echo -e "${RED}以下服务启动失败:${NC}"
        for service in "${failed_services[@]}"; do
            echo -e "  - $service"
        done
        echo ""
        echo -e "${YELLOW}请检查日志目录: $LOG_DIR${NC}"
        echo ""
        exit 1
    else
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}  所有服务启动成功！${NC}"
        echo -e "${GREEN}========================================${NC}"
        echo ""
    fi
}

# 执行主流程
main
