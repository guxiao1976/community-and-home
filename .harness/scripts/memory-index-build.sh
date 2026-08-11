#!/usr/bin/env bash
#
# memory-index-build.sh — 构建记忆倒排索引
#
# 功能：
#   - 遍历 .harness/knowledge/memory/*.md 文件
#   - 解析 frontmatter（triggers, severity, type, service）
#   - 构建倒排索引（trigger → [memory-slug]）
#   - 输出 .memory-index.json
#
# 用法：
#   bash .harness/scripts/memory-index-build.sh
#
# 输出：
#   .harness/knowledge/memory/.memory-index.json
#

set -euo pipefail

# ─── 配置 ───
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MEMORY_DIR="$PROJECT_ROOT/.harness/knowledge/memory"
INDEX_FILE="$MEMORY_DIR/.memory-index.json"

# 检查依赖
if ! command -v jq &> /dev/null; then
    echo "❌ 错误: 需要安装 jq (JSON 处理工具)"
    echo "   安装: sudo apt-get install jq  或  brew install jq"
    exit 1
fi

# 检查目录存在
if [ ! -d "$MEMORY_DIR" ]; then
    echo "❌ 错误: 记忆目录不存在: $MEMORY_DIR"
    exit 1
fi

# ─── 初始化索引 ───
echo "🔨 构建记忆倒排索引..."
echo "   源目录: $MEMORY_DIR"

# 初始化 JSON 结构
cat > "$INDEX_FILE" <<EOF
{
  "version": "1.0",
  "generated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "total_memories": 0,
  "index": {},
  "memories": {}
}
EOF

TOTAL=0
SKIPPED=0

# ─── 遍历记忆文件（支持子目录）───
shopt -s globstar nullglob
for md_file in "$MEMORY_DIR"/**/*.md; do
    basename_file=$(basename "$md_file")

    # 跳过 MEMORY.md 和 MAINTENANCE.md
    if [[ "$basename_file" == "MEMORY.md" ]] || [[ "$basename_file" == "MAINTENANCE.md" ]]; then
        continue
    fi

    if [ ! -f "$md_file" ]; then
        continue
    fi

    SLUG=$(basename "$md_file" .md)

    # 提取 frontmatter（第一个 --- 到第二个 --- 之间）
    FRONTMATTER=$(awk '/^---$/{if(++count==2) exit; flag=1; next} flag' "$md_file" 2>/dev/null || echo "")

    if [ -z "$FRONTMATTER" ]; then
        echo "   ⚠️  跳过 $basename_file (无 frontmatter)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # 提取 triggers 行
    TRIGGERS_LINE=$(echo "$FRONTMATTER" | grep "^triggers:" || echo "")

    if [ -z "$TRIGGERS_LINE" ]; then
        # 如果没有 triggers 字段，尝试从标题提取关键词
        TITLE=$(awk '/^---$/{if(++count==2) flag=1; next} flag && /^# /{print; exit}' "$md_file" | sed 's/^# //' || echo "$SLUG")
        echo "   ⚠️  $basename_file 缺少 triggers，从标题提取: $TITLE"

        # 从标题提取关键词（简单分词）
        TRIGGERS_JSON=$(echo "$TITLE" | tr ' ' '\n' | grep -v -E "^(的|和|与|或|在|对|为|是|了|必须|应该)$" | head -5 | jq -R -s -c 'split("\n") | map(select(length > 0))')
    else
        # 从 triggers 行提取内容部分
        TRIGGERS_VALUE=$(echo "$TRIGGERS_LINE" | sed 's/^triggers: //')

        # 判断是 JSON 数组还是字符串
        if [[ "$TRIGGERS_VALUE" =~ ^\[.*\]$ ]]; then
            # JSON 数组格式: ["a", "b", "c"]
            TRIGGERS_JSON="$TRIGGERS_VALUE"
        elif [[ "$TRIGGERS_VALUE" =~ ^\".*\"$ ]]; then
            # 带引号的字符串: "a b c"
            TRIGGERS_STRING=$(echo "$TRIGGERS_VALUE" | tr -d '"')
            TRIGGERS_JSON=$(echo "$TRIGGERS_STRING" | tr ' ' '\n' | jq -R -s -c 'split("\n") | map(select(length > 0))')
        else
            # 裸字符串: a b c
            TRIGGERS_JSON=$(echo "$TRIGGERS_VALUE" | tr ' ' '\n' | jq -R -s -c 'split("\n") | map(select(length > 0))')
        fi
    fi

    # 验证是否为有效 JSON 数组
    if ! echo "$TRIGGERS_JSON" | jq -e 'type == "array"' > /dev/null 2>&1; then
        echo "   ⚠️  跳过 $basename_file (triggers 解析失败)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    TRIGGER_COUNT=$(echo "$TRIGGERS_JSON" | jq -r 'length')

    if [ "$TRIGGER_COUNT" -eq 0 ]; then
        echo "   ⚠️  跳过 $basename_file (triggers 为空)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # 解析其他字段
    SEVERITY=$(echo "$FRONTMATTER" | grep "^severity:" | awk '{print $2}' || echo "info")
    TYPE=$(echo "$FRONTMATTER" | grep "^type:" | awk '{print $2}' || echo "guideline")
    SERVICE=$(echo "$FRONTMATTER" | grep "^service:" | awk '{print $2}' || echo "all")

    # 提取标题
    TITLE=$(awk '/^---$/{if(++count==2) flag=1; next} flag && /^# /{print; exit}' "$md_file" | sed 's/^# //' || echo "$SLUG")

    # 计算相对 MEMORY_DIR 的路径（如 global/testing-discipline.md）
    REL_PATH="${md_file#$MEMORY_DIR/}"

    # 写入 memories 元数据
    jq --arg slug "$SLUG" \
       --arg title "$TITLE" \
       --arg file "$REL_PATH" \
       --arg severity "$SEVERITY" \
       --arg type "$TYPE" \
       --arg service "$SERVICE" \
       --argjson triggers "$TRIGGERS_JSON" \
       '.memories[$slug] = {title: $title, file: $file, severity: $severity, type: $type, service: $service, triggers: $triggers}' \
       "$INDEX_FILE" > "$INDEX_FILE.tmp" && mv "$INDEX_FILE.tmp" "$INDEX_FILE"

    # 构建倒排索引（每个 trigger → slug）
    for i in $(seq 0 $((TRIGGER_COUNT - 1))); do
        TRIGGER=$(echo "$TRIGGERS_JSON" | jq -r ".[$i]")

        if [ -z "$TRIGGER" ] || [ "$TRIGGER" == "null" ]; then
            continue
        fi

        # 添加到倒排索引
        jq --arg trigger "$TRIGGER" \
           --arg slug "$SLUG" \
           'if .index[$trigger] then .index[$trigger] += [$slug] | .index[$trigger] |= unique else .index[$trigger] = [$slug] end' \
           "$INDEX_FILE" > "$INDEX_FILE.tmp" && mv "$INDEX_FILE.tmp" "$INDEX_FILE"
    done

    TOTAL=$((TOTAL + 1))
    echo "   ✓ $SLUG ($TRIGGER_COUNT triggers)"
done

# ─── 更新总数 ───
jq --arg total "$TOTAL" '.total_memories = ($total | tonumber)' "$INDEX_FILE" > "$INDEX_FILE.tmp" && mv "$INDEX_FILE.tmp" "$INDEX_FILE"

# ─── 输出统计 ───
echo ""
echo "✅ 记忆索引构建完成"
echo "   处理: $TOTAL 条记忆"
if [ $SKIPPED -gt 0 ]; then
    echo "   跳过: $SKIPPED 条"
fi

TRIGGER_COUNT=$(jq '.index | length' "$INDEX_FILE")
echo "   索引: $TRIGGER_COUNT 个唯一 triggers"
echo "   输出: $INDEX_FILE"
echo ""

# 显示前 10 个 triggers
echo "前 10 个 triggers (按字母排序):"
jq -r '.index | to_entries | sort_by(.key) | .[0:10] | .[] | "  \(.key) → \(.value | length) 条记忆"' "$INDEX_FILE"

exit 0
