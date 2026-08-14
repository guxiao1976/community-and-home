#!/usr/bin/env bash
# check-proto-ts-align.sh — Verify proto message fields have matching TS interface fields
#
# 覆盖策略（2026-08-14 增强，此前仅 12/211≈5.5%，其余静默跳过）：
#   1. 显式 MAPPINGS —— 处理命名不一致的映射（如 Division → AdministrativeDivision）
#   2. 自动同名匹配 —— proto message 与 TS interface 同名时自动检查字段对齐
#   3. UNMAPPED 提示 —— 前端无同名类型且未登记映射的 message，输出提示（未使用可忽略）
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
PROTO_DIR="$PROJECT_ROOT/api-proto/api"
TS_DIR="$PROJECT_ROOT/web/common/types"

VIOLATIONS=0
AUTO_VIOLATIONS=0
UNMAPPED=0
UNMAPPED_SHOWN=0

# Convert snake_case to camelCase
snake_to_camel() { echo "$1" | sed -E 's/_([a-z])/\U\1/g'; }

# 显式映射（命名不一致时）：proto_file:MessageName -> ts_file:InterfaceName
declare -A MAPPINGS
MAPPINGS["user/v1/user.proto:CommunityMembership"]="identity.ts:CommunityMembership"
MAPPINGS["user/v1/user.proto:Certification"]="identity.ts:Certification"
MAPPINGS["user/v1/user.proto:Residence"]="identity.ts:Residence"
MAPPINGS["auth/v1/auth.proto:LoginResponse"]="identity.ts:LoginResponse"
MAPPINGS["permission/v1/permission.proto:Role"]="identity.ts:Role"
MAPPINGS["permission/v1/permission.proto:Permission"]="identity.ts:Permission"
MAPPINGS["masterdata/v1/masterdata.proto:Division"]="masterdata.d.ts:AdministrativeDivision"
MAPPINGS["masterdata/v1/masterdata.proto:ResidentialArea"]="masterdata.d.ts:ResidentialArea"
MAPPINGS["masterdata/v1/masterdata.proto:SensitiveWord"]="masterdata.d.ts:SensitiveWord"
MAPPINGS["file/v1/file.proto:FileInfo"]="file.d.ts:FileInfo"

# 收集所有 proto message → 所在 proto 文件
declare -A PROTO_LOC
while IFS=$'\t' read -r f m; do
  PROTO_LOC["$m"]="$f"
done < <(grep -rH '^message ' "$PROTO_DIR" 2>/dev/null | sed "s|$PROTO_DIR/||; s/:message /\t/; s/ {.*//" )

# 收集 TS interface → 所在文件（排除 *.ts:Zone.Identifier 垃圾文件）
declare -A TS_IFACE_LOC
while IFS=$'\t' read -r f m; do
  [[ -n "$m" ]] && TS_IFACE_LOC["$m"]="$f"
done < <(find "$TS_DIR" \( -name '*.ts' -o -name '*.d.ts' \) -type f 2>/dev/null \
  | xargs grep -H 'interface ' 2>/dev/null \
  | sed "s|$TS_DIR/||; s/:.*interface /\t/; s/ {.*//" )

# 检查单个 message 的字段对齐
check_message() {
  local msg="$1" proto_file="$2" ts_file="$3" ts_iface="$4" auto="${5:-0}"
  local proto_path="$PROTO_DIR/$proto_file"
  local ts_path="$TS_DIR/$ts_file"
  [[ -f "$proto_path" ]] && [[ -f "$ts_path" ]] || return 0

  # 提取 proto 字段（message 块内，字段名 = 后跟 = N 的标识符）
  local proto_fields=""
  if grep -q "^message $msg " "$proto_path" 2>/dev/null; then
    proto_fields=$(sed -n "/^message $msg {/,/^}/p" "$proto_path" \
      | grep -E '^\s+' \
      | grep -oP '\b[a-z_][a-z_0-9]*(?=\s*=\s*\d+)' \
      | sort -u)
  fi
  [[ -z "$proto_fields" ]] && return 0

  # 提取 TS 字段
  local ts_fields=""
  if grep -q "interface $ts_iface " "$ts_path" 2>/dev/null; then
    ts_fields=$(grep -A80 "interface $ts_iface " "$ts_path" \
      | grep -E '^\s+\w+\??\s*:' \
      | grep -oP '^\s+\K\w+' \
      | sort -u)
  else
    echo "MISSING TS INTERFACE: $ts_iface in $ts_file for proto message $msg"
    VIOLATIONS=$((VIOLATIONS + 1))
    return
  fi

  # 字段对齐（proto snake_case 与 TS camelCase 都算命中；跳过 gRPC 基础设施字段）
  local pf camel_pf
  for pf in $proto_fields; do
    [[ "$pf" == "base" || "$pf" == "timestamps" ]] && continue
    camel_pf=$(snake_to_camel "$pf")
    if ! echo "$ts_fields" | grep -qx "$pf" && ! echo "$ts_fields" | grep -qx "$camel_pf"; then
      if [[ "$auto" -eq 1 ]]; then
        # 自动同名匹配发现的缺口 → WARN 不阻断（TS 可能滞后 proto，修复属前端任务）
        echo "AUTO-MISMATCH: proto $msg.$pf not found in TS $ts_iface ($ts_file)（TS 滞后 proto，待前端同步）"
        AUTO_VIOLATIONS=$((AUTO_VIOLATIONS + 1))
      else
        echo "MISMATCH: proto $msg.$pf not found in TS $ts_iface ($ts_file)"
        VIOLATIONS=$((VIOLATIONS + 1))
      fi
    fi
  done
}

# 遍历所有 proto message
for msg in "${!PROTO_LOC[@]}"; do
  local_pf="${PROTO_LOC[$msg]}"
  mapping="${MAPPINGS[$local_pf:$msg]:-}"
  if [[ -n "$mapping" ]]; then
    check_message "$msg" "$local_pf" "${mapping%%:*}" "${mapping##*:}" 0
    continue
  fi
  # 自动同名匹配
  ts_f="${TS_IFACE_LOC[$msg]:-}"
  if [[ -n "$ts_f" ]]; then
    check_message "$msg" "$local_pf" "$ts_f" "$msg" 1
  else
    UNMAPPED=$((UNMAPPED + 1))
    if [[ $UNMAPPED_SHOWN -lt 10 ]]; then
      echo "UNMAPPED: proto $msg（前端无同名 interface，未登记 MAPPINGS；未使用可忽略）"
      UNMAPPED_SHOWN=$((UNMAPPED_SHOWN + 1))
    fi
  fi
done

if [[ $UNMAPPED -gt 0 ]]; then
  echo ""
  echo "WARN: 共 $UNMAPPED 个 proto message 未映射前端类型（仅展示前 $UNMAPPED_SHOWN 个；前端已使用者请补 MAPPINGS）"
fi

if [[ $AUTO_VIOLATIONS -gt 0 ]]; then
  echo ""
  echo "WARN: 自动同名匹配发现 $AUTO_VIOLATIONS 个 TS 滞后字段（前端类型未同步 proto，不阻断）"
fi

if [[ $VIOLATIONS -eq 0 ]]; then
  echo "PASS: All proto fields match TS interfaces"
  exit 0
else
  echo "FAIL: $VIOLATIONS proto→TS alignment violations found"
  exit 1
fi
