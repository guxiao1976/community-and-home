# Check 5: Snowflake ID json:",string" tag (AST-based)
check_ast_json_string() {
  echo "  5. Checking Snowflake ID json:\",string\" tags (AST)..." >&2

  local search_dir="${SERVICE_DIR:-services}"

  # Use AST checker if available
  local ast_checker="$PROJECT_ROOT/.harness/tools/go-ast-checker/go-ast-checker"

  if [ -f "$ast_checker" ] && [ -n "$SERVICE_NAME" ]; then
    # Run AST-based check
    local json_output
    json_output=$(bash "$PROJECT_ROOT/.harness/skills/qa/scripts/ast-checks.sh" \
      "$search_dir" "$SERVICE_NAME" "true" 2>&1)

    if [ $? -eq 0 ]; then
      log_pass "ast_json_string" "all int64 ID fields have json:\",string\" (AST verified)"
    else
      # Parse JSON results
      local violations=()
      while IFS= read -r item; do
        local detail=$(echo "$item" | jq -r '.detail' 2>/dev/null || echo "")
        local location=$(echo "$item" | jq -r '.location' 2>/dev/null || echo "")
        [ -n "$detail" ] && violations+=("$location: $detail")
      done < <(echo "$json_output" | jq -c '.[]' 2>/dev/null || echo "")

      if [ ${#violations[@]} -gt 0 ]; then
        local detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
        local why="Snowflake IDs exceed JavaScript Number.MAX_SAFE_INTEGER, must be transmitted as strings"
        local fix="Add 'string' option to json tag: json:\"field_name,string\""
        local example="UserId int64 \`json:\"user_id,string\"\`"
        local reference=".harness/rules/项目编码规范.md §5"
        log_fail "ast_json_string" "$detail" "$why" "$fix" "$example" "$reference"
      else
        log_pass "ast_json_string" "all int64 ID fields have json:\",string\" (AST verified)"
      fi
    fi
  else
    # Fallback to regex-based check (legacy)
    echo "    ⚠️  AST checker not found, using regex fallback" >&2
    check_json_string_regex
  fi
}

# Legacy regex-based check (kept as fallback)
check_json_string_regex() {
  local search_dir="${SERVICE_DIR:-services}"
  local violations=()

  local go_files
  if $FULL_SCAN; then
    go_files=$(find "$search_dir" -name '*.go' -not -name '*_test.go' -not -path '*/vendor/*' | sort)
  else
    go_files=$(cd "$PROJECT_ROOT" && changed_files 'go' | { if [[ -n "$SERVICE_NAME" ]]; then grep "^services/$SERVICE_NAME/"; else grep "^services/"; fi; } | sed "s|^|$PROJECT_ROOT/|" | sort)
    if [[ -z "$go_files" ]]; then
      log_pass "json_string" "no Go changes in diff (skipped)"
      return
    fi
  fi

  while IFS= read -r gofile; do
    [[ -z "$gofile" ]] && continue
    [[ ! -f "$gofile" ]] && continue
    local in_struct=0 line_num=0
    while IFS= read -r line; do
      line_num=$((line_num + 1))
      if echo "$line" | grep -qP '^\s*type\s+\w+\s+struct\s*\{'; then
        in_struct=1
        continue
      fi
      [[ $in_struct -eq 1 ]] && echo "$line" | grep -q '^\s*\}' && { in_struct=0; continue; }
      [[ $in_struct -eq 0 ]] && continue

      if echo "$line" | grep -qP '\w+Id\s+int64.*json:"' && ! echo "$line" | grep -qP 'json:"[^"]*string'; then
        if echo "$line" | grep -qP '(path|form|header|db):"'; then
          continue
        fi
        local field
        field=$(echo "$line" | grep -oP '^\s*\K\w+')
        local rel="${gofile#$PROJECT_ROOT/}"
        violations+=("$rel:$line_num:$field")
      fi
    done < "$gofile"
  done < <(echo "$go_files")

  if [[ ${#violations[@]} -eq 0 ]]; then
    log_pass "json_string_regex" "all int64 API fields have json:\",string\" (regex)"
  else
    local detail
    detail="$(printf '%s; ' "${violations[@]}" | head -c 2000)"
    local why="Snowflake IDs exceed JavaScript Number.MAX_SAFE_INTEGER"
    local fix="Add string option: json:\"field,string\""
    log_fail "json_string_regex" "$detail" "$why" "$fix"
  fi
}
