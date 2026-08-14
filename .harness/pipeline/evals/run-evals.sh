#!/usr/bin/env bash
# run-evals.sh — P4.1 管线 eval 回归运行器
# 发现 .harness/pipeline/evals/*.eval.js 全跑，任一失败 exit 非 0（防管线改动回归）。
#
# 用法: bash .harness/pipeline/evals/run-evals.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
EVAL_DIR="$ROOT/.harness/pipeline/evals"

echo "=== Pipeline Evals ==="
overall=0
for f in "$EVAL_DIR"/*.eval.js; do
  [[ -f "$f" ]] || continue
  name="$(basename "$f")"
  if node "$f" >/dev/null 2>&1; then
    echo "  ✅ $name"
  else
    echo "  ❌ $name"
    overall=1
  fi
done

if [[ $overall -eq 0 ]]; then
  echo "=== 全部 eval 通过 ==="
else
  echo "=== 存在 eval 失败 ==="
fi
exit "$overall"
