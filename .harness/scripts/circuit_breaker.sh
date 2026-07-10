#!/usr/bin/env bash
# Circuit Breaker - Shell implementation
# Replaces circuit_breaker.py to eliminate Python dependency
#
# Usage: source circuit_breaker.sh
#        circuit_breaker_call <operation> <max_failures> <reset_timeout>

set -euo pipefail

# Circuit breaker state file directory
CB_STATE_DIR="${CB_STATE_DIR:-/tmp/circuit-breaker}"
mkdir -p "$CB_STATE_DIR"

# Circuit breaker states
CB_STATE_CLOSED="CLOSED"
CB_STATE_OPEN="OPEN"
CB_STATE_HALF_OPEN="HALF_OPEN"

# Get state file path for an operation
_cb_state_file() {
  local operation="$1"
  echo "$CB_STATE_DIR/$(echo "$operation" | tr '/' '_').state"
}

# Read circuit breaker state
_cb_read_state() {
  local state_file="$1"

  if [ ! -f "$state_file" ]; then
    echo "$CB_STATE_CLOSED:0:0"
    return
  fi

  cat "$state_file"
}

# Write circuit breaker state
_cb_write_state() {
  local state_file="$1"
  local state="$2"
  local failures="$3"
  local opened_at="$4"

  echo "$state:$failures:$opened_at" > "$state_file"
}

# Circuit breaker call
# Returns: 0 if allowed, 1 if circuit open
circuit_breaker_call() {
  local operation="${1:-}"
  local max_failures="${2:-5}"
  local reset_timeout="${3:-60}"

  if [ -z "$operation" ]; then
    echo "Error: operation name required" >&2
    return 2
  fi

  local state_file
  state_file="$(_cb_state_file "$operation")"

  local state_data
  state_data="$(_cb_read_state "$state_file")"

  local state failures opened_at
  IFS=':' read -r state failures opened_at <<< "$state_data"

  local now
  now=$(date +%s)

  case "$state" in
    "$CB_STATE_OPEN")
      # Check if reset timeout has passed
      local elapsed=$((now - opened_at))

      if [ "$elapsed" -ge "$reset_timeout" ]; then
        # Transition to HALF_OPEN
        _cb_write_state "$state_file" "$CB_STATE_HALF_OPEN" "$failures" "$opened_at"
        echo "Circuit breaker: $operation -> HALF_OPEN (reset timeout passed)" >&2
        return 0
      else
        echo "Circuit breaker: $operation is OPEN (${elapsed}s/${reset_timeout}s)" >&2
        return 1
      fi
      ;;

    "$CB_STATE_HALF_OPEN"|"$CB_STATE_CLOSED")
      # Allow the call
      return 0
      ;;

    *)
      # Unknown state, reset to CLOSED
      _cb_write_state "$state_file" "$CB_STATE_CLOSED" "0" "0"
      return 0
      ;;
  esac
}

# Record success
circuit_breaker_success() {
  local operation="${1:-}"

  if [ -z "$operation" ]; then
    echo "Error: operation name required" >&2
    return 2
  fi

  local state_file
  state_file="$(_cb_state_file "$operation")"

  local state_data
  state_data="$(_cb_read_state "$state_file")"

  local state failures opened_at
  IFS=':' read -r state failures opened_at <<< "$state_data"

  if [ "$state" = "$CB_STATE_HALF_OPEN" ]; then
    # Transition back to CLOSED
    _cb_write_state "$state_file" "$CB_STATE_CLOSED" "0" "0"
    echo "Circuit breaker: $operation -> CLOSED (success in HALF_OPEN)" >&2
  elif [ "$state" = "$CB_STATE_CLOSED" ]; then
    # Reset failure count
    _cb_write_state "$state_file" "$CB_STATE_CLOSED" "0" "0"
  fi
}

# Record failure
circuit_breaker_failure() {
  local operation="${1:-}"
  local max_failures="${2:-5}"

  if [ -z "$operation" ]; then
    echo "Error: operation name required" >&2
    return 2
  fi

  local state_file
  state_file="$(_cb_state_file "$operation")"

  local state_data
  state_data="$(_cb_read_state "$state_file")"

  local state failures opened_at
  IFS=':' read -r state failures opened_at <<< "$state_data"

  local new_failures=$((failures + 1))
  local now
  now=$(date +%s)

  if [ "$new_failures" -ge "$max_failures" ]; then
    # Open the circuit
    _cb_write_state "$state_file" "$CB_STATE_OPEN" "$new_failures" "$now"
    echo "Circuit breaker: $operation -> OPEN (failures: $new_failures >= $max_failures)" >&2
  else
    # Stay in current state, increment failure count
    _cb_write_state "$state_file" "$state" "$new_failures" "$opened_at"
    echo "Circuit breaker: $operation failures: $new_failures/$max_failures" >&2
  fi
}

# Reset circuit breaker
circuit_breaker_reset() {
  local operation="${1:-}"

  if [ -z "$operation" ]; then
    echo "Error: operation name required" >&2
    return 2
  fi

  local state_file
  state_file="$(_cb_state_file "$operation")"

  _cb_write_state "$state_file" "$CB_STATE_CLOSED" "0" "0"
  echo "Circuit breaker: $operation -> RESET" >&2
}

# Get circuit breaker status
circuit_breaker_status() {
  local operation="${1:-}"

  if [ -z "$operation" ]; then
    echo "Error: operation name required" >&2
    return 2
  fi

  local state_file
  state_file="$(_cb_state_file "$operation")"

  local state_data
  state_data="$(_cb_read_state "$state_file")"

  local state failures opened_at
  IFS=':' read -r state failures opened_at <<< "$state_data"

  echo "Circuit breaker: $operation"
  echo "  State: $state"
  echo "  Failures: $failures"

  if [ "$opened_at" != "0" ]; then
    local now
    now=$(date +%s)
    local elapsed=$((now - opened_at))
    echo "  Opened: ${elapsed}s ago"
  fi
}

# Example usage function
circuit_breaker_example() {
  echo "Circuit Breaker Example Usage:"
  echo ""
  echo "# Source the library"
  echo "source .harness/scripts/circuit_breaker.sh"
  echo ""
  echo "# Check if operation is allowed"
  echo "if circuit_breaker_call 'api-call' 5 60; then"
  echo "  # Execute operation"
  echo "  if curl -f https://api.example.com; then"
  echo "    circuit_breaker_success 'api-call'"
  echo "  else"
  echo "    circuit_breaker_failure 'api-call' 5"
  echo "  fi"
  echo "else"
  echo "  echo 'Circuit breaker is OPEN'"
  echo "fi"
  echo ""
  echo "# Check status"
  echo "circuit_breaker_status 'api-call'"
  echo ""
  echo "# Reset circuit"
  echo "circuit_breaker_reset 'api-call'"
}
