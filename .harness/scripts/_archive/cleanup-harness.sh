#!/bin/bash
# .harness 目录清理脚本
# 基于审计结果，安全删除未使用/过期的文件
#
# 执行前请审查此脚本！
# 用法: bash .harness/scripts/cleanup-harness.sh [--dry-run]

set -e

DRY_RUN=false
if [ "$1" = "--dry-run" ]; then
    DRY_RUN=true
    echo "🔍 DRY RUN 模式（不会实际删除文件）"
    echo ""
fi

PROJECT_ROOT="/home/jiaoxh/my-project/community-and-home"
cd "$PROJECT_ROOT"

# 辅助函数
do_rm() {
    if $DRY_RUN; then
        echo "  [DRY] rm $1"
    else
        rm -v "$1"
    fi
}

do_mv() {
    if $DRY_RUN; then
        echo "  [DRY] mv $1 → $2"
    else
        mv -v "$1" "$2"
    fi
}

# ============================================================
# 阶段 1: 创建归档目录
# ============================================================
echo "=== 创建归档目录 ==="
mkdir -p docs/archived/skills
mkdir -p .harness/loop-runs/_archive/{2026-06-16,2026-06-17,2026-06-18}
mkdir -p .harness/changes/_archive
echo "✅ 目录创建完成"
echo ""

# ============================================================
# 阶段 2: 删除未使用的 Skills（0 次引用）
# ============================================================
echo "=== 删除未使用的 Skills ==="
do_rm ".harness/skills/adaptive-review.md"
do_rm ".harness/skills/triage.md"
echo ""

# ============================================================
# 阶段 3: 归档降级方案（备用）
# ============================================================
echo "=== 归档降级方案 Skills（备用）==="
do_mv ".harness/skills/agent-serial-mode.md" "docs/archived/skills/"
do_mv ".harness/skills/unit-test-write.md" "docs/archived/skills/"
echo ""

# ============================================================
# 阶段 4: 归档旧 loop-runs（保留最近 3 天）
# ============================================================
echo "=== 归档旧 loop-runs ==="

# 6月16日（9个文件）
for file in .harness/loop-runs/run-2026-06-16-*.md; do
    [ -f "$file" ] || continue
    do_mv "$file" ".harness/loop-runs/_archive/2026-06-16/"
done

# 6月17日（4个文件）
for file in .harness/loop-runs/run-2026-06-17-*.md; do
    [ -f "$file" ] || continue
    do_mv "$file" ".harness/loop-runs/_archive/2026-06-17/"
done

# 6月18日（保留最后1个，移动其他）
files_18=$(ls -t .harness/loop-runs/run-2026-06-18-*.md 2>/dev/null || true)
count_18=$(echo "$files_18" | grep -c "run-" || echo 0)
if [ "$count_18" -gt 1 ]; then
    echo "$files_18" | tail -n +2 | while read file; do
        do_mv "$file" ".harness/loop-runs/_archive/2026-06-18/"
    done
else
    echo "  6-18 只有 $count_18 个文件，全部保留"
fi
echo ""

# ============================================================
# 阶段 5: 整理 changes
# ============================================================
echo "=== 整理 changes 目录 ==="
do_mv ".harness/changes/dry-run-2026-06-09.md" ".harness/changes/_archive/"
do_mv ".harness/changes/moderation-integration-retro.md" ".harness/changes/moderation-integration/retro.md"
echo ""

# ============================================================
# 阶段 6: 删除重复的 improvement-plans
# ============================================================
echo "=== 删除重复的 improvement-plans ==="
do_rm ".harness/improvement-plans/P0-memory-index.md"
echo ""

# ============================================================
# 统计
# ============================================================
echo "=========================================="
echo "✅ 清理完成统计"
echo "=========================================="
echo "剩余 Skills: $(ls .harness/skills/*.md 2>/dev/null | wc -l) 个"
echo "剩余 Loop Runs: $(ls .harness/loop-runs/run-*.md 2>/dev/null | wc -l) 个"
echo "归档 Loop Runs: $(find .harness/loop-runs/_archive -name 'run-*.md' 2>/dev/null | wc -l) 个"
echo "归档 Skills: $(ls docs/archived/skills/*.md 2>/dev/null | wc -l) 个"
echo ""

if $DRY_RUN; then
    echo "🔍 这是 DRY RUN 模式的输出"
    echo "如果确认无误，执行: bash .harness/scripts/cleanup-harness.sh"
else
    echo "✅ 实际清理已完成"
    echo ""
    echo "建议提交变更:"
    echo "  git add .harness/ docs/archived/"
    echo "  git commit -m 'chore(harness): 清理未使用的 Skills 和过期的 loop-runs'"
fi
