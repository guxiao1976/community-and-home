#!/usr/bin/env bash
#
# notify.sh — Webhook notification sender for Harness pipeline events
#
# 用法:
#   notify.sh --event pipeline_pass --service <name> --detail "<text>"
#   notify.sh --event pipeline_fail --service <name> --detail "<text>"
#   notify.sh --event need_human --service <name> --detail "<text>"
#   notify.sh --event loop_complete --detail "<text>"
#
# 环境变量（opt-in，不配置则不发送）:
#   FEISHU_WEBHOOK_URL — 飞书机器人 webhook URL
#   SLACK_WEBHOOK_URL  — Slack Incoming Webhook URL
#
# 返回码:
#   0 — 发送成功或无 webhook 配置（静默）
#   1 — 发送失败

set -euo pipefail

# ─── Helpers ──────────────────────────────────────────────────────────

EVENT=""
SERVICE=""
DETAIL=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --event)   EVENT="$2"; shift 2 ;;
    --service) SERVICE="$2"; shift 2 ;;
    --detail)  DETAIL="$2"; shift 2 ;;
    --dry-run) DRY_RUN=true; shift ;;
    *) shift ;;
  esac
done

timestamp() { date -u +"%Y-%m-%dT%H:%M:%SZ"; }

# ─── Feishu Card Payload ─────────────────────────────────────────────

build_feishu_payload() {
  local title color
  case "$EVENT" in
    pipeline_pass)
      title="✅ Harness Pipeline PASS — ${SERVICE}"
      color="green"
      ;;
    pipeline_fail)
      title="❌ Harness Pipeline FAIL — ${SERVICE}"
      color="red"
      ;;
    need_human)
      title="🟡 Harness 需要人工决策 — ${SERVICE}"
      color="yellow"
      ;;
    loop_complete)
      title="🔍 Harness Loop 扫描完成"
      color="blue"
      ;;
    *)
      title="Harness: ${EVENT}"
      color="grey"
      ;;
  esac

  cat <<PAYLOAD
{
  "msg_type": "interactive",
  "card": {
    "header": {
      "title": {
        "tag": "plain_text",
        "content": "${title}"
      },
      "template": "${color}"
    },
    "elements": [
      {
        "tag": "div",
        "text": {
          "tag": "lark_md",
          "content": "**服务**: ${SERVICE:-N/A}\n**详情**: ${DETAIL:-无}\n**时间**: $(timestamp)"
        }
      },
      {
        "tag": "hr"
      },
      {
        "tag": "note",
        "elements": [
          {
            "tag": "plain_text",
            "content": "Harness Loop · community-and-home"
          }
        ]
      }
    ]
  }
}
PAYLOAD
}

# ─── Slack Block Payload ─────────────────────────────────────────────

build_slack_payload() {
  local emoji
  case "$EVENT" in
    pipeline_pass)   emoji="✅" ;;
    pipeline_fail)   emoji="❌" ;;
    need_human)      emoji="🟡" ;;
    loop_complete)   emoji="🔍" ;;
    *)               emoji="📢" ;;
  esac

  cat <<PAYLOAD
{
  "blocks": [
    {
      "type": "header",
      "text": {
        "type": "plain_text",
        "text": "${emoji} Harness: ${EVENT}",
        "emoji": true
      }
    },
    {
      "type": "section",
      "fields": [
        {"type": "mrkdwn", "text": "*服务:*\n${SERVICE:-N/A}"},
        {"type": "mrkdwn", "text": "*时间:*\n$(timestamp)"}
      ]
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*详情:*\n${DETAIL:-无}"
      }
    },
    {
      "type": "context",
      "elements": [
        {
          "type": "mrkdwn",
          "text": "Harness Loop · community-and-home"
        }
      ]
    }
  ]
}
PAYLOAD
}

# ─── Senders ─────────────────────────────────────────────────────────

send_feishu() {
  if [[ -z "${FEISHU_WEBHOOK_URL:-}" ]]; then
    return 0
  fi

  local payload
  payload=$(build_feishu_payload)

  if $DRY_RUN; then
    echo "[DRY RUN] Feishu → ${FEISHU_WEBHOOK_URL}"
    echo "$payload"
    return 0
  fi

  local http_code
  http_code=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$FEISHU_WEBHOOK_URL" \
    -H "Content-Type: application/json" \
    -d "$payload" 2>/dev/null || echo "000")

  if [[ "$http_code" == "200" ]]; then
    echo "[notify] Feishu sent OK (event: $EVENT)"
    return 0
  else
    echo "[notify] Feishu send FAILED (HTTP $http_code, event: $EVENT)" >&2
    return 1
  fi
}

send_slack() {
  if [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
    return 0
  fi

  local payload
  payload=$(build_slack_payload)

  if $DRY_RUN; then
    echo "[DRY RUN] Slack → ${SLACK_WEBHOOK_URL}"
    echo "$payload"
    return 0
  fi

  local http_code
  http_code=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$SLACK_WEBHOOK_URL" \
    -H "Content-Type: application/json" \
    -d "$payload" 2>/dev/null || echo "000")

  if [[ "$http_code" == "200" ]]; then
    echo "[notify] Slack sent OK (event: $EVENT)"
    return 0
  else
    echo "[notify] Slack send FAILED (HTTP $http_code, event: $EVENT)" >&2
    return 1
  fi
}

# ─── Main ────────────────────────────────────────────────────────────

if [[ -z "$EVENT" ]]; then
  echo "Usage: notify.sh --event <event> [--service <name>] [--detail <text>] [--dry-run]"
  echo "Events: pipeline_pass | pipeline_fail | need_human | loop_complete"
  exit 1
fi

# Load .env for webhook URLs if not already set
if [[ -z "${FEISHU_WEBHOOK_URL:-}" ]] && [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
  PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
  if [[ -f "$PROJECT_ROOT/.env" ]]; then
    # Only source webhook-related vars, not the whole .env
    FEISHU_WEBHOOK_URL="${FEISHU_WEBHOOK_URL:-$(grep '^FEISHU_WEBHOOK_URL=' "$PROJECT_ROOT/.env" 2>/dev/null | cut -d= -f2- || true)}"
    SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-$(grep '^SLACK_WEBHOOK_URL=' "$PROJECT_ROOT/.env" 2>/dev/null | cut -d= -f2- || true)}"
  fi
fi

# If neither webhook is configured and not dry-run, silently succeed
if [[ -z "${FEISHU_WEBHOOK_URL:-}" ]] && [[ -z "${SLACK_WEBHOOK_URL:-}" ]]; then
  if $DRY_RUN; then
    echo "[DRY RUN] No webhook URLs configured. Set FEISHU_WEBHOOK_URL or SLACK_WEBHOOK_URL in .env to enable."
  fi
  exit 0
fi

send_feishu || true
send_slack || true
