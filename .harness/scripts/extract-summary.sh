#!/usr/bin/env bash
#
# extract-summary.sh — 从 QA/Review 报告提取摘要
#
# 功能：
#   - 从 _qa.md 提取摘要章节（前 N 行或特定章节）
#   - 从 _review_*.md 提取摘要和 CRITICAL 表格
#
# 用法：
#   bash .harness/scripts/extract-summary.sh --file _qa.md --lines 50
#   bash .harness/scripts/extract-summary.sh --file _review_security-arch.md --section "摘要"
#
# 输出：
#   提取的摘要文本（markdown 格式）
#

set -euo pipefail

# ─── 配置 ───
FILE=""
LINES=50
SECTION=""

# ─── 解析参数 ───
while [[ $# -gt 0 ]]; do
    case $1 in
        --file)
            FILE="$2"
            shift 2
            ;;
        --lines)
            LINES="$2"
            shift 2
            ;;
        --section)
            SECTION="$2"
            shift 2
            ;;
        --help|-h)
            echo "用法: $0 --file <path> [--lines N | --section <name>]"
            echo ""
            echo "选项:"
            echo "  --file <path>       输入文件路径"
            echo "  --lines <N>         提取前 N 行（默认 50）"
            echo "  --section <name>    提取特定章节（如 '摘要'）"
            echo "  --help              显示帮助"
            echo ""
            echo "示例:"
            echo "  $0 --file services/xxx/_qa.md --lines 50"
            echo "  $0 --file services/xxx/_review_*.md --section 摘要"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            exit 2
            ;;
    esac
done

# ─── 参数检查 ───
if [ -z "$FILE" ]; then
    echo "错误: 必须指定 --file" >&2
    exit 2
fi

if [ ! -f "$FILE" ]; then
    echo "错误: 文件不存在: $FILE" >&2
    exit 2
fi

# ─── 提取摘要 ───
if [ -n "$SECTION" ]; then
    # 提取特定章节（从 ## 章节名 到下一个 ## 或文件末尾）
    awk -v section="$SECTION" '
        /^## / {
            if ($0 ~ section) {
                in_section = 1
                print
                next
            } else if (in_section) {
                exit
            }
        }
        in_section { print }
    ' "$FILE"
else
    # 提取前 N 行
    head -n "$LINES" "$FILE"
fi

exit 0
