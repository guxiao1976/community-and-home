#!/usr/bin/env bash
# Wrapper script for AST-based Go checks
# Usage: bash ast-checks.sh <service-dir> <service-name> [json]

set -euo pipefail

# Calculate project root correctly
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# From .harness/skills/qa/scripts -> go up 3 levels to project root
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

AST_CHECKER="$PROJECT_ROOT/.harness/tools/go-ast-checker/go-ast-checker"
REGISTRY="$PROJECT_ROOT/.harness/registry/services.json"

SERVICE_DIR="${1:-}"
SERVICE_NAME="${2:-}"
JSON_OUTPUT="${3:-false}"

if [ -z "$SERVICE_DIR" ] || [ -z "$SERVICE_NAME" ]; then
  echo "Usage: $0 <service-dir> <service-name> [json]" >&2
  exit 1
fi

# Build AST checker if not exists
if [ ! -f "$AST_CHECKER" ]; then
  echo "⚠️  AST checker not found, building..." >&2
  
  (
    cd "$PROJECT_ROOT/.harness/tools/go-ast-checker" || exit 1
    go build -o go-ast-checker . 2>&1
  ) || {
    echo "❌ Failed to build AST checker" >&2
    exit 1
  }
  
  echo "✅ AST checker built successfully" >&2
fi

# Resolve service dir: harness-checks.sh passes absolute paths, but support relative too.
# (FIX: previously $PROJECT_ROOT/$SERVICE_DIR doubled when SERVICE_DIR was already absolute.)
if [[ "$SERVICE_DIR" = /* ]]; then
  ABS_SERVICE_DIR="$SERVICE_DIR"
else
  ABS_SERVICE_DIR="$PROJECT_ROOT/$SERVICE_DIR"
fi

# Run AST checks with absolute paths
if [ "$JSON_OUTPUT" = "true" ]; then
  "$AST_CHECKER" \
    -service-dir "$ABS_SERVICE_DIR" \
    -service-name "$SERVICE_NAME" \
    -registry "$REGISTRY" \
    -json
else
  "$AST_CHECKER" \
    -service-dir "$ABS_SERVICE_DIR" \
    -service-name "$SERVICE_NAME" \
    -registry "$REGISTRY"
fi
