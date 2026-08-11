#!/usr/bin/env bash
# Generate .service.json metadata files for all services
# This replaces hardcoded service mappings with auto-discovered metadata

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

echo "🔍 Scanning services directory..."

generated=0
skipped=0

for svc_dir in services/*-service; do
  if [ ! -d "$svc_dir" ]; then
    continue
  fi

  svc_name=$(basename "$svc_dir")
  metadata_file="$svc_dir/.service.json"

  # Skip if already exists (but continue loop, don't break)
  if [ -f "$metadata_file" ]; then
    echo "  ⏭️  $svc_name - .service.json already exists, skipping"
    skipped=$((skipped + 1))
    continue
  fi

  # Extract module path from go.mod
  module=""
  language="go"
  if [ -f "$svc_dir/go.mod" ]; then
    module=$(grep "^module" "$svc_dir/go.mod" | awk '{print $2}')
  else
    # No go.mod - likely Python service
    language="python"
    echo "  ⚠️  $svc_name - no go.mod found, assuming Python service"
  fi

  # Detect Chinese display name from CLAUDE.md
  display_name=""
  if [ -f "$svc_dir/CLAUDE.md" ]; then
    # Try to extract from first heading or comments
    display_name=$(grep -m 1 "^# " "$svc_dir/CLAUDE.md" | sed 's/^# //' | sed 's/ — .*//' || echo "")
  fi

  # Fallback to service name if no display name found
  if [ -z "$display_name" ]; then
    display_name="$svc_name"
  fi

  # Check for API and RPC layers
  has_api=false
  has_rpc=false
  [ -d "$svc_dir/api" ] && has_api=true
  [ -d "$svc_dir/rpc" ] && has_rpc=true

  # Generate .service.json
  cat > "$metadata_file" << EOF
{
  "name": "$svc_name",
  "displayName": "$display_name",
  "language": "$language",
  "module": "$module",
  "hasApi": $has_api,
  "hasRpc": $has_rpc,
  "generated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "generatedBy": "generate-service-metadata.sh"
}
EOF

  echo "  ✅ $svc_name → .service.json created"
  ((generated++))
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Summary:"
echo "   Generated: $generated"
echo "   Skipped:   $skipped"
echo "   Total:     $((generated + skipped))"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $generated -gt 0 ]; then
  echo ""
  echo "✅ Service metadata files created successfully"
  echo ""
  echo "💡 Next steps:"
  echo "   1. Review generated .service.json files"
  echo "   2. Run: bash .harness/scripts/build-service-registry.sh"
  echo "   3. Update Pipeline and QA scripts to use registry"
fi
