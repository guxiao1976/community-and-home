#!/bin/bash
# ============================================================
# 社区平台 — 重启所有服务
# 使用方式：bash scripts/restart.sh
# ============================================================

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
bash "$ROOT/scripts/stop.sh"
sleep 2
bash "$ROOT/scripts/start.sh"
