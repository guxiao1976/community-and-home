#!/usr/bin/env bash
#
# complete-feature.sh — 一键完成需求的所有后续步骤
#
# 用法:
#   bash .harness/scripts/complete-feature.sh <feature-name> <commit-message>
#
# 示例:
#   bash .harness/scripts/complete-feature.sh "RBAC管理界面" "feat(permission): RBAC管理界面实现"
#
# 自动执行:
#   1. 运行测试
#   2. 提交代码
#   3. 运行 QA 检查
#   4. 提示更新变更索引
#   5. 同步知识图谱
#   6. 最终扫描
#

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 参数检查
FEATURE_NAME="${1:-}"
COMMIT_MSG="${2:-}"

if [[ -z "$FEATURE_NAME" ]]; then
  echo -e "${RED}错误: 缺少功能名称${NC}"
  echo ""
  echo "用法:"
  echo "  bash .harness/scripts/complete-feature.sh <feature-name> <commit-message>"
  echo ""
  echo "示例:"
  echo "  bash .harness/scripts/complete-feature.sh \"RBAC管理界面\" \"feat(permission): RBAC管理界面实现\""
  exit 1
fi

if [[ -z "$COMMIT_MSG" ]]; then
  echo -e "${RED}错误: 缺少提交消息${NC}"
  exit 1
fi

echo "╔════════════════════════════════════════════════════════╗"
echo "║  🚀 完成需求工作流                                      ║"
echo "╠════════════════════════════════════════════════════════╣"
echo "║  功能: $FEATURE_NAME"
echo "║  提交: $COMMIT_MSG"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# ─────────────────────────────────────────────────────────
# 步骤 1: 检查是否有未暂存的修改
# ─────────────────────────────────────────────────────────
echo -e "${BLUE}[1/6]${NC} 检查工作区状态..."

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo -e "${GREEN}  ✓ 发现未提交的修改${NC}"
else
  echo -e "${YELLOW}  ⚠ 工作区没有修改，是否继续？ [y/N]${NC}"
  read -r -p "  " response
  if [[ ! "$response" =~ ^[Yy]$ ]]; then
    echo -e "${RED}  ✗ 已取消${NC}"
    exit 0
  fi
fi

echo ""

# ─────────────────────────────────────────────────────────
# 步骤 2: 运行冒烟测试
# ─────────────────────────────────────────────────────────
echo -e "${BLUE}[2/6]${NC} 运行冒烟测试..."

if bash .harness/scripts/harness-smoke.sh > /tmp/smoke-test.log 2>&1; then
  echo -e "${GREEN}  ✓ 冒烟测试通过${NC}"
else
  echo -e "${RED}  ✗ 冒烟测试失败${NC}"
  echo ""
  echo "  查看日志: cat /tmp/smoke-test.log"
  echo ""
  read -r -p "  是否忽略测试失败继续？ [y/N] " response
  if [[ ! "$response" =~ ^[Yy]$ ]]; then
    echo -e "${RED}  已取消，请先修复测试${NC}"
    exit 1
  fi
fi

echo ""

# ─────────────────────────────────────────────────────────
# 步骤 3: 提交代码
# ─────────────────────────────────────────────────────────
echo -e "${BLUE}[3/6]${NC} 提交代码..."

git add .

# 构造完整的 commit 消息
FULL_COMMIT_MSG="$COMMIT_MSG

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"

git commit -m "$FULL_COMMIT_MSG"

COMMIT_HASH=$(git rev-parse --short HEAD)
echo -e "${GREEN}  ✓ 代码已提交: $COMMIT_HASH${NC}"
echo ""

# ─────────────────────────────────────────────────────────
# 步骤 4: 运行 QA 检查
# ─────────────────────────────────────────────────────────
echo -e "${BLUE}[4/6]${NC} 运行 QA 检查..."

if bash .harness/scripts/harness-tasks.sh scan --auto-create > /tmp/qa-scan.log 2>&1; then
  # 检查是否有新问题
  NEW_ISSUES=$(grep "new issue(s) found" /tmp/qa-scan.log | grep -oP '\d+' || echo "0")
  if [[ "$NEW_ISSUES" -eq 0 ]]; then
    echo -e "${GREEN}  ✓ QA 检查通过，无新问题${NC}"
  else
    echo -e "${YELLOW}  ⚠ 发现 $NEW_ISSUES 个新问题${NC}"
    echo "  查看详情: cat /tmp/qa-scan.log"
  fi
else
  echo -e "${RED}  ✗ QA 检查失败${NC}"
  cat /tmp/qa-scan.log
fi

echo ""

# ─────────────────────────────────────────────────────────
# 步骤 5: 更新变更索引（交互式）
# ─────────────────────────────────────────────────────────
echo -e "${BLUE}[5/6]${NC} 更新变更索引..."
echo ""
echo -e "${YELLOW}  请手动更新 .harness/changes/INDEX.md${NC}"
echo "  添加以下记录："
echo ""
echo "  | 功能名称 | $FEATURE_NAME |"
echo "  | 提交记录 | $COMMIT_HASH |"
echo "  | 完成时间 | $(date +%Y-%m-%d) |"
echo ""
read -r -p "  完成后按 Enter 继续..."

echo ""

# ─────────────────────────────────────────────────────────
# 步骤 6: 同步知识图谱
# ─────────────────────────────────────────────────────────
echo -e "${BLUE}[6/6]${NC} 同步知识图谱..."

if bash .harness/scripts/graph-sync.sh > /tmp/graph-sync.log 2>&1; then
  echo -e "${GREEN}  ✓ 知识图谱同步完成${NC}"
else
  echo -e "${YELLOW}  ⚠ 知识图谱同步失败（可能 Neo4j 未运行）${NC}"
  echo "  查看日志: cat /tmp/graph-sync.log"
fi

echo ""

# ─────────────────────────────────────────────────────────
# 完成总结
# ─────────────────────────────────────────────────────────
echo "╔════════════════════════════════════════════════════════╗"
echo -e "║  ${GREEN}✅ 需求 $FEATURE_NAME 已完成！${NC}"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "📊 完成状态："
echo ""
echo "  最新提交:"
git log --oneline -1
echo ""
echo "  待处理任务:"
bash .harness/scripts/harness-tasks.sh list --status pending | head -5 || echo "    无待处理任务"
echo ""
echo "📝 后续步骤（可选）："
echo "  • 推送到远程: git push"
echo "  • 创建 PR: gh pr create"
echo "  • 部署: bash scripts/deploy.sh"
echo ""
