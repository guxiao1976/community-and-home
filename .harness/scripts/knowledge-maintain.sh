#!/usr/bin/env bash
#
# knowledge-maintain.sh — 知识维护脚本
#
# 功能：
#   检测知识系统的健康状态，包括记忆过期、死链接、缺失 frontmatter、
#   CLAUDE.md 与记忆矛盾、索引过期等。支持 --fix 自动修复部分问题。
#
# 用法：
#   bash .harness/scripts/knowledge-maintain.sh --check          # 只读检查
#   bash .harness/scripts/knowledge-maintain.sh --fix --auto     # 自动修复
#   bash .harness/scripts/knowledge-maintain.sh --help
#
# 检查维度：
#   1. [STALE]    记忆超过 90 天未更新 + apply_count=0
#   2. [BROKEN]   记忆中的 [[link]] 引用目标不存在
#   3. [MISSING]  记忆缺少 type/severity/triggers frontmatter
#   4. [INCONSISTENT] CLAUDE.md 中的规则与记忆文件矛盾
#   5. [STALE_INDEX] .memory-index.json 过期（>7 天）
#   6. [ORPHAN]    .claude/memory/ 目录存在但只有 README.md
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MEMORY_DIR="$PROJECT_ROOT/.harness/knowledge/memory"
INDEX_FILE="$MEMORY_DIR/.memory-index.json"
MEMORY_MD="$MEMORY_DIR/MEMORY.md"

MODE="check"
AUTO=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --check) MODE="check"; shift ;;
    --fix) MODE="fix"; shift ;;
    --auto) AUTO=true; shift ;;
    --help|-h) sed -n '2,29p' "$0" | sed 's/^# //'; exit 0 ;;
    *) echo "未知参数: $1"; exit 2 ;;
  esac
done

NOW_EPOCH=$(date +%s)
STALE_DAYS=90
INDEX_STALE_DAYS=7
ISSUES=0
FIXED=0

# ─── 辅助函数 ───
age_days() {
  local file="$1"
  [ ! -f "$file" ] && echo 999 && return
  local mtime
  mtime=$(stat -c %Y "$file" 2>/dev/null || echo 0)
  echo $(( (NOW_EPOCH - mtime) / 86400 ))
}

# ─── 1. [STALE] 过期记忆检查 ───
check_stale() {
  echo "=== [STALE] 过期记忆检查 ==="
  local count=0
  shopt -s globstar nullglob 2>/dev/null || true
for md in "$MEMORY_DIR"/**/*.md; do
    [ ! -f "$md" ] && continue
    bn=$(basename "$md")
    [ "$bn" == "MEMORY.md" ] || [ "$bn" == "MAINTENANCE.md" ] && continue

    age=$(age_days "$md")

    # 检查 apply_count
    apply=$(awk '/^apply_count:/{print $2}' "$md" 2>/dev/null || echo "null")
    [ -z "$apply" ] && apply="missing"

    if [ "$age" -gt "$STALE_DAYS" ] && { [ "$apply" == "0" ] || [ "$apply" == "missing" ] || [ "$apply" == "null" ]; }; then
      slug=$(basename "$md" .md)
      if [ $count -eq 0 ]; then echo ""; fi
      echo "  ⚠️  $slug (${age}d old, apply_count=$apply)"
      count=$((count + 1))
      ISSUES=$((ISSUES + 1))
    fi
  done
  if [ $count -eq 0 ]; then
    echo "  ✅ 无过期记忆"
  else
    echo "  📊 $count 条记忆过期未应用"
  fi
  echo ""
}

# ─── 2. [BROKEN] 死链接检查 ───
check_broken_links() {
  echo "=== [BROKEN] 死链接检查 ==="
  local count=0
  shopt -s globstar nullglob 2>/dev/null || true
for md in "$MEMORY_DIR"/**/*.md; do
    [ ! -f "$md" ] && continue
    bn=$(basename "$md")
    [ "$bn" == "MEMORY.md" ] || [ "$bn" == "MAINTENANCE.md" ] && continue

    # 提取 [[...]] 引用
    links=$(grep -oP '\[\[([^\]]+)\]\]' "$md" 2>/dev/null | sed 's/\[\[//;s/\]\]//' || true)
    [ -z "$links" ] && continue

    slug=$(basename "$md" .md)
    while IFS= read -r link; do
      # 跳过 § 引用（如 [[项目编码规范 §12]]）
      [[ "$link" == *"§"* ]] && continue
      # 跳过带路径的引用
      [[ "$link" == *"/"* ]] && continue

      # 检查目标记忆是否存在（支持子目录搜索）
      target=$(find "$MEMORY_DIR" -name "${link}.md" -type f 2>/dev/null | head -1)
      if [ -z "$target" ]; then
        if [ $count -eq 0 ]; then echo ""; fi
        echo "  ❌ $slug → [[$link]] (目标不存在)"
        count=$((count + 1))
        ISSUES=$((ISSUES + 1))
      fi
    done <<< "$links"
  done
  if [ $count -eq 0 ]; then
    echo "  ✅ 无死链接"
  else
    echo "  📊 $count 个死链接"
  fi
  echo ""
}

# ─── 3. [MISSING] frontmatter 完整性检查 ───
check_frontmatter() {
  echo "=== [MISSING] Frontmatter 完整性检查 ==="
  local count=0
  shopt -s globstar nullglob 2>/dev/null || true
for md in "$MEMORY_DIR"/**/*.md; do
    [ ! -f "$md" ] && continue
    bn=$(basename "$md")
    [ "$bn" == "MEMORY.md" ] || [ "$bn" == "MAINTENANCE.md" ] && continue

    slug=$(basename "$md" .md)
    fm=$(awk '/^---$/{if(++count==2) exit; flag=1; next} flag' "$md" 2>/dev/null || echo "")

    if [ -z "$fm" ]; then
      if [ $count -eq 0 ]; then echo ""; fi
      echo "  ❌ $slug: 无 frontmatter"
      count=$((count + 1))
      ISSUES=$((ISSUES + 1))
      continue
    fi

    missing=""
    for field in triggers type severity; do
      if ! echo "$fm" | grep -q "^${field}:"; then
        missing="$missing $field"
      fi
    done

    if [ -n "$missing" ]; then
      if [ $count -eq 0 ]; then echo ""; fi
      echo "  ⚠️  $slug: 缺失$missing"
      count=$((count + 1))
      ISSUES=$((ISSUES + 1))

      # --fix 模式下尝试从 MEMORY.md 补充 type
      if [ "$MODE" == "fix" ] && $AUTO; then
        mem_type=$(grep "$slug" "$MEMORY_MD" 2>/dev/null | grep -oP '`(pitfall|guideline|process|decision|model)`' | head -1 | tr -d '`' || echo "")
        if [ -n "$mem_type" ] && echo "$missing" | grep -q "type"; then
          sed -i "/^severity:/a type: $mem_type" "$md" 2>/dev/null && \
            echo "    → 已补充 type: $mem_type" && FIXED=$((FIXED + 1)) || true
        fi
      fi
    fi
  done
  if [ $count -eq 0 ]; then
    echo "  ✅ 所有记忆 frontmatter 完整"
  else
    echo "  📊 $count 条记忆 frontmatter 不完整"
  fi
  echo ""
}

# ─── 4. [INCONSISTENT] CLAUDE.md 与记忆矛盾检查 ───
check_claude_consistency() {
  echo "=== [INCONSISTENT] CLAUDE.md 一致性检查 ==="
  local count=0

  # 检查 permission-service: is_system 规则
  perm_claude="$PROJECT_ROOT/services/permission-service/CLAUDE.md"
  if [ -f "$perm_claude" ]; then
    if grep -q "系统角色直接放行\|自动获得所有权限" "$perm_claude" 2>/dev/null; then
      echo "  ❌ permission-service/CLAUDE.md: is_system 规则与记忆矛盾"
      count=$((count + 1))
      ISSUES=$((ISSUES + 1))

      if [ "$MODE" == "fix" ] && $AUTO; then
        sed -i 's/系统角色直接放行.*/is_system 仅防止误删\/误改 — 权限由 rel_role_permission 配置决定/' "$perm_claude" 2>/dev/null && \
          echo "    → 已修正" && FIXED=$((FIXED + 1)) || true
      fi
    fi
  fi

  if [ $count -eq 0 ]; then
    echo "  ✅ 未发现矛盾"
  else
    echo "  📊 $count 处矛盾"
  fi
  echo ""
}

# ─── 5. [STALE_INDEX] 索引过期检查 ───
check_index_freshness() {
  echo "=== [STALE_INDEX] 索引时效检查 ==="
  if [ ! -f "$INDEX_FILE" ]; then
    echo "  ❌ 索引不存在"
    ISSUES=$((ISSUES + 1))

    if [ "$MODE" == "fix" ] && $AUTO; then
      bash "$SCRIPT_DIR/memory-index-build.sh" > /dev/null 2>&1 && \
        echo "    → 索引已重建" && FIXED=$((FIXED + 1)) || true
    fi
    return
  fi

  age=$(age_days "$INDEX_FILE")
  generated=$(jq -r '.generated_at // "unknown"' "$INDEX_FILE" 2>/dev/null || echo "unknown")
  count=$(jq -r '.total_memories // 0' "$INDEX_FILE" 2>/dev/null || echo 0)

  # 统计实际记忆文件数
  actual=0
  shopt -s globstar nullglob 2>/dev/null || true
for md in "$MEMORY_DIR"/**/*.md; do
    [ ! -f "$md" ] && continue
    bn=$(basename "$md")
    [ "$bn" == "MEMORY.md" ] || [ "$bn" == "MAINTENANCE.md" ] && continue
    actual=$((actual + 1))
  done

  if [ "$age" -gt "$INDEX_STALE_DAYS" ]; then
    echo "  ⚠️  索引 ${age} 天前生成 ($generated)，索引 $count 条，实际 $actual 条"
    [ "$count" -ne "$actual" ] && echo "  ❌ 索引与实际数量不符: $count vs $actual"
    ISSUES=$((ISSUES + 1))

    if [ "$MODE" == "fix" ] && $AUTO; then
      bash "$SCRIPT_DIR/memory-index-build.sh" > /dev/null 2>&1 && \
        echo "    → 索引已重建" && FIXED=$((FIXED + 1)) || true
    fi
  else
    echo "  ✅ 索引 ${age} 天前生成，$count 条记忆"
  fi
  echo ""
}

# ─── 6. [ORPHAN] 空记忆目录检查 ───
check_orphan_dirs() {
  echo "=== [ORPHAN] 空服务记忆目录检查 ==="
  local count=0
  for svc_dir in "$PROJECT_ROOT"/services/*/; do
    mem_dir="${svc_dir}.claude/memory"
    if [ -d "$mem_dir" ]; then
      actual=$(find "$mem_dir" -name "*.md" ! -name "README.md" -type f 2>/dev/null | wc -l)
      if [ "$actual" -eq 0 ]; then
        svc=$(basename "$svc_dir")
        if [ $count -eq 0 ]; then echo ""; fi
        echo "  ⚠️  $svc: .claude/memory/ 为空（仅 README.md）"
        count=$((count + 1))
        ISSUES=$((ISSUES + 1))
      fi
    fi
  done
  if [ $count -eq 0 ]; then
    echo "  ✅ 所有服务记忆目录有内容"
  else
    echo "  📊 $count 个服务记忆目录为空"
  fi
  echo ""
}

# ─── Main ───
echo "🔍 知识维护检查 ($(date +%Y-%m-%d))"
echo ""

check_stale
check_broken_links
check_frontmatter
check_claude_consistency
check_index_freshness
check_orphan_dirs

# ─── 总结 ───
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ISSUES -eq 0 ]; then
  echo "✅ 知识系统健康，0 问题"
else
  echo "⚠️  发现 $ISSUES 个问题"
  if [ "$MODE" == "fix" ]; then
    echo "   已修复: $FIXED 个"
  else
    echo "   运行 --fix --auto 自动修复可修复的问题"
  fi
fi
echo ""

# 返回码
[ $ISSUES -gt 0 ] && exit 1 || exit 0
