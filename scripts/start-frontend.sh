#!/usr/bin/env bash
#
# start-frontend.sh — 一键启动前端（PC 管理后台 + Mobile H5）
#
# 用法：
#   bash scripts/start-frontend.sh               # 启动全部前端
#   bash scripts/start-frontend.sh --pc-only     # 仅 PC 管理后台 (web/pc)
#   bash scripts/start-frontend.sh --mobile-only # 仅移动端 H5 (web/mobile)
#   bash scripts/start-frontend.sh --stop        # 停止前端进程
#
# 依赖：Node.js 18+；web/pc 与 web/mobile 已执行过 npm install
# 日志：PC → /tmp/frontend-pc.log，Mobile → /tmp/frontend-mobile.log
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PC_DIR="$ROOT/web/pc"
MOBILE_DIR="$ROOT/web/mobile"
PC_LOG="/tmp/frontend-pc.log"
MOBILE_LOG="/tmp/frontend-mobile.log"

MODE="all"
STOP=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --pc-only) MODE="pc"; shift ;;
    --mobile-only) MODE="mobile"; shift ;;
    --stop) STOP=true; shift ;;
    -h|--help) sed -n '2,12p' "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "未知参数: $1（支持 --pc-only / --mobile-only / --stop）" >&2; exit 2 ;;
  esac
done

if $STOP; then
  echo "=== 停止前端进程 ==="
  # PC: vite dev server；Mobile: uni (vite) dev server
  pkill -f "vite" 2>/dev/null && echo "  ✅ 已停止 vite 进程" || echo "  ⏭️  无 vite 进程在运行"
  pkill -f "dcloudio" 2>/dev/null || true
  pkill -f "uni -p" 2>/dev/null || true
  echo "✅ 前端已停止"
  exit 0
fi

check_deps() {
  local dir="$1" name="$2"
  if [[ ! -d "$dir/node_modules" ]]; then
    echo "❌ $name 依赖未安装，请先执行: cd $dir && npm install" >&2
    return 1
  fi
  return 0
}

# 真正后台化启动：setsid 创建新会话 + </dev/null 断开 stdin + disown 脱离作业表。
# 原因：在 Claude Code / agent 的 Bash 工具里，普通 `nohup ... &` 会被工具
# 视为子进程持续等待（dev server 永不退出 → 命令卡死、日志为空）。
# 三件套确保命令立即返回、进程由 init 收养、工具不再等待。
launch_bg() {
  local dir="$1" script="$2" log="$3" name="$4"
  cd "$dir"
  setsid nohup npm run "$script" > "$log" 2>&1 < /dev/null &
  local pid=$!
  disown "$pid" 2>/dev/null || true
  cd "$ROOT"
  echo "  ✅ $name 已启动 (PID $pid)，日志: tail -f $log"
  sleep 1
  # 快速自检：1 秒后日志仍为空说明启动失败，提示查看日志
  if [[ ! -s "$log" ]]; then
    echo "  ⚠️  ${name} 启动 1 秒后仍无输出——可能启动失败，请查看: tail -50 $log"
  fi
}

start_pc() {
  echo "=== 启动 PC 管理后台 (web/pc) ==="
  check_deps "$PC_DIR" "PC 前端" || return 1
  launch_bg "$PC_DIR" "dev" "$PC_LOG" "PC"
}

start_mobile() {
  echo "=== 启动 Mobile H5 (web/mobile) ==="
  check_deps "$MOBILE_DIR" "Mobile 前端" || return 1
  launch_bg "$MOBILE_DIR" "dev:h5" "$MOBILE_LOG" "Mobile"
}

case "$MODE" in
  pc) start_pc ;;
  mobile) start_mobile ;;
  *) start_pc; start_mobile ;;
esac

echo ""
echo "══════════════════════════════════════════════"
echo "  前端启动完成"
echo "══════════════════════════════════════════════"
echo "  访问地址（以 vite 实际输出为准）:"
echo "    PC:     http://localhost:5173   (查看 $PC_LOG 确认)"
echo "    Mobile: H5 默认端口见 $MOBILE_LOG"
echo ""
echo "  常用操作:"
echo "    停止:   bash scripts/start-frontend.sh --stop"
echo "    看日志: tail -f $PC_LOG $MOBILE_LOG"
echo ""
