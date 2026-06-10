#!/usr/bin/env bash
# check-proto-ts-align.sh — Verify proto message fields have matching TS interface fields
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
PROTO_DIR="$PROJECT_ROOT/api-proto/api"
TS_DIR="$PROJECT_ROOT/web/common/types"

VIOLATIONS=0

# Convert snake_case to camelCase
snake_to_camel() {
  echo "$1" | sed -E 's/_([a-z])/\U\1/g'
}

# Map proto message names to TS interface files
# Format: proto_file:MessageName -> ts_file:InterfaceName
declare -A MAPPINGS

# user/v1 → identity.ts
MAPPINGS["user/v1/user.proto:CommunityMembership"]="identity.ts:CommunityMembership"
MAPPINGS["user/v1/user.proto:Certification"]="identity.ts:Certification"
MAPPINGS["user/v1/user.proto:Residence"]="identity.ts:Residence"

# auth/v1 → identity.ts
MAPPINGS["auth/v1/auth.proto:LoginResponse"]="identity.ts:LoginResponse"

# permission/v1 → identity.ts
MAPPINGS["permission/v1/permission.proto:Role"]="identity.ts:Role"
MAPPINGS["permission/v1/permission.proto:Permission"]="identity.ts:Permission"

# masterdata/v1 → masterdata.d.ts
MAPPINGS["masterdata/v1/masterdata.proto:Division"]="masterdata.d.ts:AdministrativeDivision"
MAPPINGS["masterdata/v1/masterdata.proto:ResidentialArea"]="masterdata.d.ts:ResidentialArea"
MAPPINGS["masterdata/v1/masterdata.proto:SensitiveWord"]="masterdata.d.ts:SensitiveWord"

# file/v1 → file.d.ts
MAPPINGS["file/v1/file.proto:FileInfo"]="file.d.ts:FileInfo"

for MAPPING in "${!MAPPINGS[@]}"; do
    PROTO_KEY="${MAPPING%%:*}"
    PROTO_MSG="${MAPPING##*:}"
    TS_VAL="${MAPPINGS[$MAPPING]}"

    TS_FILE="${TS_VAL%%:*}"
    TS_IFACE="${TS_VAL##*:}"

    PROTO_FILE="$PROTO_DIR/$PROTO_KEY"
    TS_PATH="$TS_DIR/$TS_FILE"

    if [[ ! -f "$PROTO_FILE" ]]; then
        echo "SKIP: proto file not found: $PROTO_FILE"
        continue
    fi

    # Extract proto fields for this message.
    # Find the message block and extract field names (the identifier before = number).
    PROTO_FIELDS=""
    if grep -q "^message $PROTO_MSG " "$PROTO_FILE" 2>/dev/null; then
        PROTO_FIELDS=$(sed -n '/^message '"$PROTO_MSG"' {/,/^}/p' "$PROTO_FILE" \
          | grep -E '^\s+' \
          | grep -oP '\b[a-z_][a-z_0-9]*(?=\s*=\s*\d+)' \
          | sort -u)
    fi

    if [[ -z "$PROTO_FIELDS" ]]; then
        echo "SKIP: no fields found for proto message $PROTO_MSG (or message not found)"
        continue
    fi

    # Check if TS file has this interface
    if [[ ! -f "$TS_PATH" ]]; then
        echo "MISSING TS FILE: $TS_DIR/$TS_FILE for proto message $PROTO_MSG"
        VIOLATIONS=$((VIOLATIONS + 1))
        continue
    fi

    # Extract TS interface fields (identifier before : or ?:)
    TS_FIELDS=""
    if grep -q "interface $TS_IFACE " "$TS_PATH" 2>/dev/null; then
        TS_FIELDS=$(grep -A60 "interface $TS_IFACE " "$TS_PATH" \
          | grep -E '^\s+\w+\??\s*:' \
          | grep -oP '^\s+\K\w+' \
          | sort -u)
    else
        echo "MISSING TS INTERFACE: $TS_IFACE in $TS_FILE for proto message $PROTO_MSG"
        VIOLATIONS=$((VIOLATIONS + 1))
        continue
    fi

    # Find proto fields missing in TS (check both snake_case and camelCase)
    # Skip known gRPC infrastructure fields that don't map to frontend types:
    #   base    — BaseResp gRPC envelope, handled at transport layer
    #   timestamps — common.v1.Timestamps, represented as flat fields in TS
    for pf in $PROTO_FIELDS; do
        [[ "$pf" == "base" ]] && continue
        [[ "$pf" == "timestamps" ]] && continue
        camel_pf=$(snake_to_camel "$pf")
        if ! echo "$TS_FIELDS" | grep -qx "$pf" && ! echo "$TS_FIELDS" | grep -qx "$camel_pf"; then
            echo "MISMATCH: proto $PROTO_MSG.$pf not found in TS $TS_IFACE ($TS_FILE)"
            VIOLATIONS=$((VIOLATIONS + 1))
        fi
    done
done

if [[ $VIOLATIONS -eq 0 ]]; then
    echo "PASS: All proto fields match TS interfaces"
    exit 0
else
    echo "FAIL: $VIOLATIONS proto→TS alignment violations found"
    exit 1
fi
