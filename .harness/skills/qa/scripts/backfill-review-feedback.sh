#!/usr/bin/env bash
# backfill-review-feedback.sh — P4.2 评审发现结构化回填
#
# 读取 .harness/review-feedback/*.jsonl（管线评审时写入）：
#   - *.warnings.jsonl  → 按 change 聚合为一条 backlog task（source=review）
#   - *.memory.jsonl    → 追加到 .harness/knowledge/memory/pending-suggestions.md（Owner 审阅后写入记忆）
# 处理完的文件移入 processed/（幂等，不重复建任务）。
#
# 用法: bash .harness/skills/qa/scripts/backfill-review-feedback.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
FB_DIR="$ROOT/.harness/review-feedback"
PROCESSED_DIR="$FB_DIR/processed"
PENDING_MEMORY="$ROOT/.harness/knowledge/memory/pending-suggestions.md"
TASKS_SCRIPT="$ROOT/.harness/scripts/harness-tasks.sh"

mkdir -p "$PROCESSED_DIR" "$ROOT/.harness/knowledge/memory"

echo "=== P4.2 评审发现回填 ==="
processed=0

# ── 1) WARNING → backlog task（按 change 聚合）──
for f in "$FB_DIR"/*.warnings.jsonl; do
  [[ -f "$f" ]] || continue
  change="$(basename "$f" .warnings.jsonl)"
  count=0
  detail=""
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    count=$((count + 1))
    detail="$detail$line"$'\n'
  done < "$f"
  if [[ $count -gt 0 ]]; then
    # 用 python3 解析 JSONL 生成可读 detail（值含引号/转义也能正确处理）
    readable="$(python3 - "$f" <<'PY' 2>/dev/null || echo ''
import json, sys
try:
    with open(sys.argv[1]) as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            d = json.loads(line)
            print(f"- [{d.get('section', '')}] {d.get('issue', '')}（{d.get('lens', '')}）")
except Exception:
    pass
PY
)"
    bash "$TASKS_SCRIPT" create \
      --title "评审发现（review-feedback）: ${change} 共 ${count} 项 WARNING" \
      --service harness --type debt --priority P2 --source review \
      --detail "${change} 评审 WARNING 结构化回填：${readable:-（解析失败，见原文件 $f）}" >/dev/null 2>&1 || echo "  ⚠️ 建任务失败: ${change}"
    mv "$f" "$PROCESSED_DIR/"
    echo "  ✅ ${change}: ${count} 项 WARNING → backlog task"
    processed=$((processed + 1))
  fi
done

# ── 2) Memory 建议 → pending 记忆文件 ──
for f in "$FB_DIR"/*.memory.jsonl; do
  [[ -f "$f" ]] || continue
  service="$(basename "$f" .memory.jsonl)"
  {
    echo ""; echo "--- ${service} $(date +%F) ---"
    cat "$f"
  } >> "$PENDING_MEMORY"
  mv "$f" "$PROCESSED_DIR/"
  echo "  ✅ ${service}: Memory 建议 → pending-suggestions.md"
  processed=$((processed + 1))
done

if [[ $processed -eq 0 ]]; then
  echo "  无待回填反馈（review-feedback 目录为空或已处理）"
fi
echo "=== 完成 ==="
