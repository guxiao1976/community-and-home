#!/usr/bin/env bash
# Load service registry utilities for Shell scripts
# Usage: source .harness/scripts/service-registry-loader.sh

REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
REGISTRY_FILE="$REPO_ROOT/.harness/registry/services.json"

if [ ! -f "$REGISTRY_FILE" ]; then
  echo "❌ Error: Service registry not found at $REGISTRY_FILE" >&2
  echo "   Run: bash .harness/scripts/build-service-registry.sh" >&2
  exit 1
fi

# Load all services into SERVICES array
declare -a SERVICES
while IFS= read -r svc; do
  SERVICES+=("$svc")
done < <(jq -r '.services[].name' "$REGISTRY_FILE")

# Load service module mappings into SVC_MODULE_MAP associative array
declare -A SVC_MODULE_MAP
while IFS='=' read -r svc module; do
  SVC_MODULE_MAP["$svc"]="$module"
done < <(jq -r '.services[] | "\(.name)=\(.module)"' "$REGISTRY_FILE")

# Helper: Get service module path
get_service_module() {
  local svc="$1"
  echo "${SVC_MODULE_MAP[$svc]:-}"
}

# Helper: Check if service exists
service_exists() {
  local svc="$1"
  [[ " ${SERVICES[*]} " =~ " ${svc} " ]]
}

# Helper: List all services
list_services() {
  printf '%s\n' "${SERVICES[@]}"
}
