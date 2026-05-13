#!/bin/bash

# 社区管理系统 - 停止脚本
# 停止所有前后端服务

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  社区管理系统 - 停止所有服务${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 服务端口配置
declare -A SERVICE_PORTS=(
    ["identity-rpc"]="8081"
    ["masterdata-rpc"]="8083"
    ["identity-api"]="8888"
    ["masterdata-api"]="8889"
    ["moderation-api"]="8890"
    ["frontend"]="3000"
)

# 检查端口是否被占用
check_port() {
    local port=$1
    ss -tlnp 2>/dev/null | grep -q ":$port " && return 0 || return 1
}

# 根据端口停止服务
stop_by_port() {
    local port=$1
    local name=$2

    if ! check_port "$port"; then
        echo -e "${YELLOW}  $name (端口 $port) 未运行${NC}"
        return 0
    fi

    local pids=$(ss -tlnp 2>/dev/null | grep ":$port " | grep -oP 'pid=\K[0-9]+' | sort -u)

    if [ -n "$pids" ]; then
        echo -e "${BLUE}[停止] $name (端口 $port)${NC}"
        for pid in $pids; do
            local cmd=$(ps -p $pid -o comm= 2>/dev/null || echo "unknown")
            echo -e "${YELLOW}  正在停止进程 $pid ($cmd)...${NC}"

            kill $pid 2>/dev/null || true
            sleep 1

            # 如果进程还在，强制杀死
            if kill -0 $pid 2>/dev/null; then
                echo -e "${YELLOW}  进程未响应，强制停止...${NC}"
                kill -9 $pid 2>/dev/null || true
                sleep 0.5
            fi

            # 再次检查
            if kill -0 $pid 2>/dev/null; then
                echo -e "${RED}  ✗ 无法停止进程 $pid${NC}"
                return 1
            else
                echo -e "${GREEN}  ✓ 进程 $pid 已停止${NC}"
            fi
        done

        # 最终检查端口
        sleep 1
        if check_port "$port"; then
            echo -e "${RED}  ✗ 端口 $port 仍被占用${NC}"
            return 1
        else
            echo -e "${GREEN}  ✓ $name 已完全停止${NC}"
            return 0
        fi
    fi

    return 0
}

# 根据进程名停止服务
stop_by_pattern() {
    local pattern=$1
    local name=$2

    local pids=$(pgrep -f "$pattern" 2>/dev/null || true)

    if [ -z "$pids" ]; then
        return 0
    fi

    echo -e "${BLUE}[停止] $name (匹配: $pattern)${NC}"
    for pid in $pids; do
        local cmd=$(ps -p $pid -o cmd= 2>/dev/null | cut -c1-60 || echo "unknown")
        echo -e "${YELLOW}  正在停止进程 $pid: $cmd${NC}"

        kill $pid 2>/dev/null || true
        sleep 1

        if kill -0 $pid 2>/dev/null; then
            echo -e "${YELLOW}  进程未响应，强制停止...${NC}"
            kill -9 $pid 2>/dev/null || true
            sleep 0.5
        fi

        if kill -0 $pid 2>/dev/null; then
            echo -e "${RED}  ✗ 无法停止进程 $pid${NC}"
        else
            echo -e "${GREEN}  ✓ 进程 $pid 已停止${NC}"
        fi
    done

    return 0
}

# 主流程
main() {
    local failed_services=()

    echo -e "${YELLOW}正在停止所有服务...${NC}"
    echo ""

    # 停止前端服务
    echo -e "${BLUE}停止前端服务${NC}"
    echo ""

    # 先按进程名停止
    stop_by_pattern "vite --host" "Vite 开发服务器"
    # 再按端口停止（以防有遗漏）
    if ! stop_by_port "3000" "前端服务"; then
        failed_services+=("frontend")
    fi
    echo ""

    # 停止 API 服务
    echo -e "${BLUE}停止 API 服务${NC}"
    echo ""

    for service in "moderation-api" "masterdata-api" "identity-api"; do
        if ! stop_by_port "${SERVICE_PORTS[$service]}" "$service"; then
            failed_services+=("$service")
        fi
        echo ""
    done

    # 停止 RPC 服务
    echo -e "${BLUE}停止 RPC 服务${NC}"
    echo ""

    for service in "masterdata-rpc" "identity-rpc"; do
        if ! stop_by_port "${SERVICE_PORTS[$service]}" "$service"; then
            failed_services+=("$service")
        fi
        echo ""
    done

    # 显示最终状态
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  服务状态检查${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    local all_stopped=true

    echo "RPC 服务:"
    for service in "identity-rpc" "masterdata-rpc"; do
        local port="${SERVICE_PORTS[$service]}"
        if check_port "$port"; then
            echo -e "${RED}✗${NC} $service (端口 $port) - 仍在运行"
            all_stopped=false
        else
            echo -e "${GREEN}✓${NC} $service (端口 $port) - 已停止"
        fi
    done
    echo ""

    echo "API 服务:"
    for service in "identity-api" "masterdata-api" "moderation-api"; do
        local port="${SERVICE_PORTS[$service]}"
        if check_port "$port"; then
            echo -e "${RED}✗${NC} $service (端口 $port) - 仍在运行"
            all_stopped=false
        else
            echo -e "${GREEN}✓${NC} $service (端口 $port) - 已停止"
        fi
    done
    echo ""

    echo "前端服务:"
    if check_port "3000"; then
        echo -e "${RED}✗${NC} vite-dev-server (端口 3000) - 仍在运行"
        all_stopped=false
    else
        echo -e "${GREEN}✓${NC} vite-dev-server (端口 3000) - 已停止"
    fi
    echo ""

    # 显示结果
    if [ "$all_stopped" = true ]; then
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}  所有服务已成功停止！${NC}"
        echo -e "${GREEN}========================================${NC}"
        echo ""
        exit 0
    else
        echo -e "${RED}========================================${NC}"
        echo -e "${RED}  部分服务停止失败${NC}"
        echo -e "${RED}========================================${NC}"
        echo ""
        echo -e "${YELLOW}提示: 可以手动检查并停止残留进程${NC}"
        echo ""
        exit 1
    fi
}

# 执行主流程
main
