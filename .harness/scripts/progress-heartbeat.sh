#!/bin/bash
# ============================================================
# progress-heartbeat.sh — 每 N 分钟进度心跳（长任务防"假死"）
# ============================================================
# 用法：
#   1) 启动心跳（前台阻塞；在后台以 Monitor 运行即可自动送达通知）：
#        bash .harness/scripts/progress-heartbeat.sh [interval_seconds]
#      默认 interval=300（5 分钟）。
#   2) 执行 Agent 在每个步骤边界更新状态文件（单行）：
#       echo "当前: <正在做什么>" > $CLAUDE_JOB_DIR/tmp/progress.txt
#     心跳每 interval 秒发一条通知：`[HH:MM:SS] <状态文件内容>`。
#
# 依赖：$CLAUDE_JOB_DIR/tmp/progress.txt（由执行 Agent 维护）
# 设计依据：.harness/docs/context-management.md §4.1
# ============================================================

INTERVAL="${1:-300}"

# 状态文件默认路径（可用 PROGRESS_FILE 覆盖）
STATUS_FILE="${PROGRESS_FILE:-${CLAUDE_JOB_DIR:-/tmp}/tmp/progress.txt}"

mkdir -p "$(dirname "$STATUS_FILE")"
[ -f "$STATUS_FILE" ] || echo "心跳已启动" > "$STATUS_FILE"

while true; do
  printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$(cat "$STATUS_FILE" 2>/dev/null || echo '无进度标记')"
  sleep "$INTERVAL"
done
