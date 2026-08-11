#!/usr/bin/env bash
#
# setup-cron.sh — 配置 Harness 定时任务
#
# 用法:
#   bash .harness/scripts/setup-cron.sh          # 交互式配置
#   bash .harness/scripts/setup-cron.sh --auto   # 自动配置（使用默认值）
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CRON_BACKUP="/tmp/crontab.backup.$(date +%Y%m%d-%H%M%S)"

AUTO_MODE=false
REMOVE_MODE=false

case "${1:-}" in
  --auto)
    AUTO_MODE=true
    ;;
  --remove)
    REMOVE_MODE=true
    ;;
esac

# 移除模式
if $REMOVE_MODE; then
  echo "════════════════════════════════════════════════════════"
  echo "  移除 Harness 定时任务"
  echo "════════════════════════════════════════════════════════"
  echo ""

  if crontab -l > "$CRON_BACKUP" 2>/dev/null; then
    echo "✓ 已备份当前 crontab 到: $CRON_BACKUP"

    # 移除 Harness 相关任务
    grep -v "harness\|graph-sync\|Harness 自动化任务\|传感器扫描\|知识图谱同步\|完整质量检查" "$CRON_BACKUP" > /tmp/crontab.new 2>/dev/null || true

    if crontab /tmp/crontab.new; then
      echo -e "${GREEN}✓ Harness 定时任务已移除${NC}"
      rm -f /tmp/crontab.new
    else
      echo -e "${RED}✗ 移除失败${NC}"
      exit 1
    fi
  else
    echo "ℹ 当前用户没有 crontab"
  fi
  exit 0
fi

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "════════════════════════════════════════════════════════"
echo "  Harness 定时任务配置工具"
echo "════════════════════════════════════════════════════════"
echo ""

# 检查当前 crontab
echo -e "${YELLOW}[1/4] 检查当前 crontab...${NC}"
if crontab -l > "$CRON_BACKUP" 2>/dev/null; then
  echo "  ✓ 已备份当前 crontab 到: $CRON_BACKUP"
  EXISTING_JOBS=$(grep -c "harness\|graph-sync" "$CRON_BACKUP" 2>/dev/null || echo "0")
  if [[ $EXISTING_JOBS -gt 0 ]]; then
    echo -e "  ${YELLOW}⚠ 发现 $EXISTING_JOBS 个已存在的 Harness 相关任务${NC}"
  fi
else
  echo "  ℹ 当前用户没有 crontab，将创建新的"
  touch "$CRON_BACKUP"
fi
echo ""

# 准备新的 cron 任务
echo -e "${YELLOW}[2/4] 准备定时任务配置...${NC}"

CRON_TASKS="
# ════════════════════════════════════════════════════════════
# Harness 自动化任务
# 项目: community-and-home
# 配置时间: $(date '+%Y-%m-%d %H:%M:%S')
# ════════════════════════════════════════════════════════════

# 1. 传感器扫描 - 每4小时运行一次
# 检测：QA问题、TODO桩、Review警告、图谱新鲜度、GitHub Issues/PRs
0 */4 * * * cd $PROJECT_ROOT && bash .harness/scripts/harness-tasks.sh scan --auto-create >> /tmp/harness-scan.log 2>&1

# 2. 知识图谱同步 - 每天凌晨2点
# 同步 Proto、Go、TypeScript 代码到 Neo4j 知识图谱
# 注意: Neo4j 未启动时本任务会 exit 1 并写入日志（显式失败，非静默跳过）
0 2 * * * cd $PROJECT_ROOT && bash .harness/scripts/graph-sync.sh >> /tmp/graph-sync.log 2>&1

# 3. 完整质量检查 - 每周一上午9点
# 运行完整的后端+前端质量检查套件
0 9 * * 1 cd $PROJECT_ROOT && bash .harness/scripts/run-all-checks.sh >> /tmp/harness-weekly-check.log 2>&1

# 4. 残留 worktree 清理 - 每周日凌晨3点
# 清理 pipeline 隔离用 worktree 的残留（只删无未提交改动的）
0 3 * * 0 cd $PROJECT_ROOT && bash .harness/scripts/cleanup-worktrees.sh >> /tmp/harness-worktree-cleanup.log 2>&1

# 注意：
# - 日志文件位于 /tmp/，定期清理
# - 如需调整频率，使用 'crontab -e' 编辑
# - 移除这些任务：bash $PROJECT_ROOT/.harness/scripts/setup-cron.sh --remove
# ════════════════════════════════════════════════════════════
"

echo "  准备添加以下任务："
echo "    • 传感器扫描: 每4小时 (0 */4 * * *)"
echo "    • 知识图谱同步: 每天2:00 AM (0 2 * * *)"
echo "    • 完整质量检查: 每周一9:00 AM (0 9 * * 1)"
echo "    • worktree 清理: 每周日3:00 AM (0 3 * * 0)"
echo ""

# 确认
if ! $AUTO_MODE; then
  echo -e "${YELLOW}[3/4] 确认配置...${NC}"
  read -p "  是否继续安装这些定时任务？[y/N] " -r
  echo ""
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}✗ 已取消${NC}"
    exit 0
  fi
else
  echo -e "${GREEN}[3/4] 自动模式，跳过确认${NC}"
  echo ""
fi

# 安装 cron 任务
echo -e "${YELLOW}[4/4] 安装定时任务...${NC}"

# 移除旧的 Harness 任务
grep -v "harness\|graph-sync\|Harness 自动化任务" "$CRON_BACKUP" > /tmp/crontab.new 2>/dev/null || true

# 添加新任务
echo "$CRON_TASKS" >> /tmp/crontab.new

# 安装新的 crontab
if crontab /tmp/crontab.new; then
  echo -e "${GREEN}  ✓ 定时任务已安装${NC}"
  echo ""
  echo "════════════════════════════════════════════════════════"
  echo -e "${GREEN}✓ 配置完成！${NC}"
  echo "════════════════════════════════════════════════════════"
  echo ""
  echo "已添加的定时任务："
  echo ""
  crontab -l | grep -A 12 "Harness 自动化任务"
  echo ""
  echo "日志文件位置："
  echo "  • 传感器扫描: /tmp/harness-scan.log"
  echo "  • 知识图谱同步: /tmp/graph-sync.log"
  echo "  • 完整检查: /tmp/harness-weekly-check.log"
  echo ""
  echo "管理命令："
  echo "  • 查看任务: crontab -l"
  echo "  • 编辑任务: crontab -e"
  echo "  • 移除任务: bash $PROJECT_ROOT/.harness/scripts/setup-cron.sh --remove"
  echo ""
else
  echo -e "${RED}✗ 安装失败${NC}"
  echo "  备份文件: $CRON_BACKUP"
  exit 1
fi

# 清理临时文件
rm -f /tmp/crontab.new
