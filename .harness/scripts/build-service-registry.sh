#!/usr/bin/env bash
# Build centralized service registry from individual .service.json files
# Output: .harness/registry/services.json

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

REGISTRY_DIR=".harness/registry"
OUTPUT="$REGISTRY_DIR/services.json"

mkdir -p "$REGISTRY_DIR"

echo "🔨 Building service registry..."

# Start JSON array
echo "{" > "$OUTPUT"
echo "  \"version\": \"1.0.0\"," >> "$OUTPUT"
echo "  \"generated\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"," >> "$OUTPUT"
echo "  \"services\": [" >> "$OUTPUT"

first=true
for svc_dir in services/*-service; do
  if [ ! -f "$svc_dir/.service.json" ]; then
    echo "  ⚠️  $(basename "$svc_dir") - missing .service.json, skipping"
    continue
  fi

  # Add comma separator (except for first item)
  if [ "$first" = false ]; then
    echo "," >> "$OUTPUT"
  fi
  first=false

  # Read and indent the service metadata
  cat "$svc_dir/.service.json" | sed 's/^/    /' | sed '1s/^    /    /' >> "$OUTPUT"

  echo -n "  ✅ $(basename "$svc_dir")"
done

echo "" >> "$OUTPUT"
echo "  ]," >> "$OUTPUT"

# Add web projects
echo "  \"web\": [" >> "$OUTPUT"
echo "    {\"name\": \"pc\", \"displayName\": \"管理后台\", \"type\": \"admin\"}," >> "$OUTPUT"
echo "    {\"name\": \"mobile\", \"displayName\": \"移动端\", \"type\": \"mobile\"}" >> "$OUTPUT"
echo "  ]" >> "$OUTPUT"
echo "}" >> "$OUTPUT"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Service registry built successfully"
echo "   📄 Output: $OUTPUT"
echo "   📊 Services: $(ls -1 services/*-service/.service.json | wc -l)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 Validate with:"
echo "   jq . $OUTPUT"
