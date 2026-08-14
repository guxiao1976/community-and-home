#!/usr/bin/env bash
#
# knowledge-load.sh — 知识加载脚本
#
# 功能：
#   根据任务上下文（服务名 + 关键词）查询记忆索引，
#   按优先级排序返回最相关的记忆列表，供 Agent 在任务开始前加载。
#
# 用法：
#   bash .harness/scripts/knowledge-load.sh \
#     --service permission-service \
#     --keywords "CheckPermission, is_system, 权限检查"
#
#   bash .harness/scripts/knowledge-load.sh \
#     --service auth-service \
#     --task "修复短信验证码校验逻辑"    # 自动提取关键词
#
#   bash .harness/scripts/knowledge-load.sh --check-budget  # L1 预算检查
#
# 输出（默认 human-readable，--json 为 JSON）：
#   按优先级排序的记忆列表，含 severity、匹配理由、文件路径
#
# 优先级公式（0-100）：
#   severity: must-follow=40, should-follow=20, info=5
#   service_match: exact=30, all=15, mismatch=-100（排除）
#   keyword_match: 每个触发词匹配 +5 (max 25)
#   recency: updated 90 天内 +5
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
INDEX_FILE="$PROJECT_ROOT/.harness/knowledge/memory/.memory-index.json"
MEMORY_DIR="$PROJECT_ROOT/.harness/knowledge/memory"

# ─── 参数 ───
SERVICE=""
KEYWORDS=()
TASK_DESC=""
JSON_OUT=false
CHECK_BUDGET=false
TOP_N=5
HELP=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --service) SERVICE="$2"; shift 2 ;;
    --keywords) IFS=',' read -ra kws <<< "$2"; KEYWORDS+=("${kws[@]}"); shift 2 ;;
    --task) TASK_DESC="$2"; shift 2 ;;
    --json) JSON_OUT=true; shift ;;
    --check-budget) CHECK_BUDGET=true; shift ;;
    --top) TOP_N="$2"; shift 2 ;;
    --help|-h) HELP=true; shift ;;
    *) echo "未知参数: $1"; exit 2 ;;
  esac
done

if $HELP; then
  sed -n '2,30p' "$0" | sed 's/^# //'
  exit 0
fi

# ─── L1 预算检查 ───
if $CHECK_BUDGET; then
  ROOT_CLAUDE="$PROJECT_ROOT/CLAUDE.md"
  OWNER_AGENT="$PROJECT_ROOT/.harness/agents/owner-agent.md"
  CODING_RULES="$PROJECT_ROOT/.harness/rules/项目编码规范.md"

  L1_LINES=0
  for f in "$ROOT_CLAUDE" "$OWNER_AGENT" "$CODING_RULES"; do
    if [[ -f "$f" ]]; then
      lines=$(wc -l < "$f" 2>/dev/null || echo 0)
      L1_LINES=$((L1_LINES + lines))
    fi
  done

  BUDGET=400
  if $JSON_OUT; then
    echo "{\"l1_lines\": $L1_LINES, \"budget\": $BUDGET, \"over_budget\": $([ $L1_LINES -gt $BUDGET ] && echo true || echo false), \"usage_pct\": $((L1_LINES * 100 / BUDGET))}"
  else
    echo "=== L1 上下文预算 ==="
    echo "  CLAUDE.md:          $(wc -l < "$ROOT_CLAUDE") 行"
    echo "  owner-agent.md:     $(wc -l < "$OWNER_AGENT") 行"
    echo "  项目编码规范:       $(wc -l < "$CODING_RULES") 行"
    echo "  ─────────────────"
    echo "  L1 总计:           $L1_LINES / $BUDGET 行"
    if [ $L1_LINES -gt $BUDGET ]; then
      echo "  🔴 超标 +$((L1_LINES - BUDGET)) 行"
    else
      echo "  🟢 在预算内（余量 $((BUDGET - L1_LINES)) 行）"
    fi
  fi
  exit 0
fi

# ─── 从任务描述提取关键词 ───
if [ -n "$TASK_DESC" ] && [ ${#KEYWORDS[@]} -eq 0 ]; then
  # 简单分词：中英文混合提取
  # 提取英文单词（2+ 字符）
  eng_words=$(echo "$TASK_DESC" | grep -oE '[a-zA-Z_]{2,}' | tr '\n' ' ' || true)
  # 中文滑动窗口提取（2-4 字子串，覆盖完整词组；grep 无法重叠匹配，用 python3）
  chn_words=$(echo "$TASK_DESC" | python3 -c "
import sys
text = sys.stdin.read().strip()
words = set()
for i in range(len(text)):
    for L in (2,3,4):
        w = text[i:i+L]
        if len(w) == L and all('\u4e00' <= c <= '\u9fff' for c in w):
            words.add(w)
print(' '.join(sorted(words)))
" 2>/dev/null || echo "")
  combined="$eng_words $chn_words"
  # 去重
  read -ra KEYWORDS <<< "$(echo "$combined" | tr ' ' '\n' | sort -u | tr '\n' ' ')"

  if $JSON_OUT; then
    echo "{\"extracted_keywords\": [$(printf '"%s",' "${KEYWORDS[@]}" | sed 's/,$//')]}" >&2
  else
    echo "📝 从任务描述自动提取关键词: ${KEYWORDS[*]:-无}" >&2
  fi
fi

if [ ${#KEYWORDS[@]} -eq 0 ]; then
  echo "❌ 需要 --keywords 或 --task 参数" >&2
  exit 1
fi

# ─── 依赖检查 ───
if ! command -v jq &> /dev/null; then
  echo "❌ 需要安装 jq" >&2
  exit 1
fi

if [ ! -f "$INDEX_FILE" ]; then
  echo "❌ 索引不存在: $INDEX_FILE" >&2
  echo "   运行: bash .harness/scripts/memory-index-build.sh" >&2
  exit 1
fi

# ─── 查询匹配记忆 ───
declare -A MEM_SCORES=()
declare -A MEM_KW_COUNT=()
declare -A MEM_REASONS=()
declare -A MEM_SEVERITIES=()
declare -A MEM_SERVICES=()
declare -A MEM_FILES=()
declare -A MEM_TITLES=()

NOW_EPOCH=$(date +%s)

for kw in "${KEYWORDS[@]}"; do
  # 忽略过短的关键词
  [ ${#kw} -lt 2 ] && continue

  matched=$(jq -r --arg kw "$kw" '.index[$kw] // [] | .[]' "$INDEX_FILE" 2>/dev/null || echo "")
  [ -z "$matched" ] && continue

  while IFS= read -r slug; do
    [ -z "$slug" ] && continue

    # 获取记忆元数据
    mem=$(jq -r --arg slug "$slug" '.memories[$slug]' "$INDEX_FILE")
    [ "$mem" == "null" ] && continue

    mem_svc=$(echo "$mem" | jq -r '.service // "all"')
    mem_sev=$(echo "$mem" | jq -r '.severity // "info"')
    mem_file=$(echo "$mem" | jq -r '.file // ""')
    mem_title=$(echo "$mem" | jq -r '.title // ""')

    # 服务过滤：只匹配 service==SERVICE 或 service=="all"
    if [ -n "$SERVICE" ] && [ "$mem_svc" != "all" ] && [ "$mem_svc" != "$SERVICE" ]; then
      continue
    fi

    # 计分：首次命中加一次性基础分（severity/service/时效），每次命中累加 keyword 计数。
    # 修复：此前 MEM_SCORES 每次被覆盖（非累加），kw_count 恒 1/2——多关键词加分永不达标
    if [ -z "${MEM_SCORES[$slug]:-}" ]; then
      # 首次命中：基础分（severity + service + 时效，只算一次）
      score=0
      reasons=""
      case "$mem_sev" in
        must-follow) score=$((score + 40)); reasons="severity=must-follow" ;;
        should-follow) score=$((score + 20)); reasons="severity=should-follow" ;;
        *) score=$((score + 5)); reasons="severity=info" ;;
      esac
      if [ -n "$SERVICE" ] && [ "$mem_svc" == "$SERVICE" ]; then
        score=$((score + 30)); reasons="$reasons, service=exact"
      elif [ "$mem_svc" == "all" ]; then
        score=$((score + 15)); reasons="$reasons, service=global"
      fi
      mem_path="$MEMORY_DIR/$mem_file"
      [ ! -f "$mem_path" ] && mem_path=$(find "$MEMORY_DIR" -name "$mem_file" -type f 2>/dev/null | head -1)
      if [ -f "$mem_path" ]; then
        mtime=$(stat -c %Y "$mem_path" 2>/dev/null || echo 0)
        age_days=$(( (NOW_EPOCH - mtime) / 86400 ))
        if [ "$age_days" -le 90 ]; then
          score=$((score + 5)); reasons="$reasons, updated=${age_days}d ago"
        fi
      fi
      kw_count=1
    else
      # 后续命中：保留基础分，累加 keyword 计数
      score=${MEM_SCORES[$slug]}
      reasons="${MEM_REASONS[$slug]}"
      kw_count=$(( ${MEM_KW_COUNT[$slug]:-1} + 1 ))
    fi

    # keyword bonus（每命中一个关键词 +5，上限 25）
    bonus=$((kw_count * 5))
    [ $bonus -gt 25 ] && bonus=25
    score=$((score + bonus))
    reasons="$reasons, keywords=$kw_count"

    MEM_SCORES[$slug]=$score
    MEM_REASONS[$slug]="$reasons"
    MEM_KW_COUNT[$slug]=$kw_count
    MEM_SEVERITIES[$slug]="$mem_sev"
    MEM_SERVICES[$slug]="$mem_svc"
    MEM_FILES[$slug]="$mem_file"
    MEM_TITLES[$slug]="$mem_title"

  done <<< "$matched"
done

# ─── 排序输出 ───
# set -u 下空关联数组 ${#arr[@]} 报 unbound，用 + 兼容
if [ ${#MEM_SCORES[@]} -eq 0 ]; then
  if $JSON_OUT; then
    echo '{"memories": [], "summary": "未找到匹配的记忆"}'
  else
    echo "🔍 未找到匹配的记忆"
    echo "   关键词: ${KEYWORDS[*]}"
    [ -n "$SERVICE" ] && echo "   服务: $SERVICE"
    echo ""
    echo "💡 建议：运行 bash .harness/scripts/memory-index-build.sh 重建索引"
  fi
  # ── 打点：无匹配记忆也记录（matched=0，保证打点完整；否则复盘丢"零命中"样本）──
  bash "$SCRIPT_DIR/log-usage.sh" knowledge-load \
    service="${SERVICE:-all}" top="$TOP_N" matched="0" \
    "keywords=$(IFS=,; echo "${KEYWORDS[*]}")" 2>/dev/null || true
  exit 0
fi

# 按分数降序排列
sorted=$(for slug in "${!MEM_SCORES[@]}"; do
  echo "${MEM_SCORES[$slug]} $slug"
done | sort -rn | head -n "$TOP_N")

must_follow_count=0
should_follow_count=0

if $JSON_OUT; then
  echo -n '{"memories": ['
  first=true
  while IFS=' ' read -r score slug; do
    [ -z "$slug" ] && continue
    sev="${MEM_SEVERITIES[$slug]}"
    [ "$sev" == "must-follow" ] && must_follow_count=$((must_follow_count + 1))
    [ "$sev" == "should-follow" ] && should_follow_count=$((should_follow_count + 1))

    $first || echo -n ','
    first=false
    echo -n "{\"slug\":\"$slug\",\"file\":\"${MEM_FILES[$slug]}\",\"title\":\"${MEM_TITLES[$slug]}\",\"severity\":\"$sev\",\"service\":\"${MEM_SERVICES[$slug]}\",\"priority\":$score,\"reason\":\"${MEM_REASONS[$slug]}\"}"
  done <<< "$sorted"
  echo "],\"summary\":\"加载 $(echo "$sorted" | wc -l) 条记忆，其中 must-follow: $must_follow_count 条\"}"
else
  echo "🔍 知识加载结果"
  echo "   关键词: ${KEYWORDS[*]}"
  [ -n "$SERVICE" ] && echo "   服务: $SERVICE"
  echo ""

  while IFS=' ' read -r score slug; do
    [ -z "$slug" ] && continue
    sev="${MEM_SEVERITIES[$slug]}"
    [ "$sev" == "must-follow" ] && must_follow_count=$((must_follow_count + 1))
    [ "$sev" == "should-follow" ] && should_follow_count=$((should_follow_count + 1))

    icon="📌"
    [ "$sev" == "must-follow" ] && icon="🔴"
    [ "$sev" == "should-follow" ] && icon="🟡"
    [ "$sev" == "info" ] && icon="ℹ️ "

    echo "$icon [$score] ${MEM_TITLES[$slug]}"
    echo "   slug: $slug | ${MEM_SERVICES[$slug]} | ${MEM_REASONS[$slug]}"
    echo "   file: ${MEM_FILES[$slug]}"
    echo ""
  done <<< "$sorted"

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "加载 $(echo "$sorted" | wc -l | xargs) 条记忆 | must-follow: $must_follow_count | should-follow: $should_follow_count"
fi

# ── 打点：记录知识检索调用（为流水线复盘提供第一手数据：命中率/关键词/服务）──
matched_count=$(echo "$sorted" | wc -l | xargs)
bash "$SCRIPT_DIR/log-usage.sh" knowledge-load \
  service="${SERVICE:-all}" top="$TOP_N" matched="$matched_count" \
  "keywords=$(IFS=,; echo "${KEYWORDS[*]}")" 2>/dev/null || true
